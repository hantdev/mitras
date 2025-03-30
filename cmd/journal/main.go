// Package main contains journal main function to start the journal service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/hantdev/mitras/journal"
	httpapi "github.com/hantdev/mitras/journal/api"
	"github.com/hantdev/mitras/journal/events"
	"github.com/hantdev/mitras/journal/middleware"
	journalpg "github.com/hantdev/mitras/journal/postgres"
	mitraslog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	mitrasauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	domainsAuthz "github.com/hantdev/mitras/pkg/domains/grpcclient"
	"github.com/hantdev/mitras/pkg/events/store"
	"github.com/hantdev/mitras/pkg/grpcclient"
	jaegerclient "github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/postgres"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	"github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName          = "journal"
	envPrefixDB      = "MITRAS_JOURNAL_DB_"
	envPrefixHTTP    = "MITRAS_JOURNAL_HTTP_"
	envPrefixAuth    = "MITRAS_AUTH_GRPC_"
	envPrefixDomains = "MITRAS_DOMAINS_GRPC_"
	defDB            = "journal"
	defSvcHTTPPort   = "9021"
)

type config struct {
	LogLevel      string  `env:"MITRAS_JOURNAL_LOG_LEVEL"   envDefault:"info"`
	ESURL         string  `env:"MITRAS_ES_URL"              envDefault:"nats://localhost:4222"`
	JaegerURL     url.URL `env:"MITRAS_JAEGER_URL"          envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry bool    `env:"MITRAS_SEND_TELEMETRY"      envDefault:"false"`
	InstanceID    string  `env:"MITRAS_JOURNAL_INSTANCE_ID" envDefault:""`
	TraceRatio    float64 `env:"MITRAS_JAEGER_TRACE_RATIO"  envDefault:"1.0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mitraslog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	var exitCode int
	defer mitraslog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	db, err := pgclient.Setup(dbConfig, *journalpg.Migration())
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	authClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&authClientCfg, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	authn, authnHandler, err := authsvcAuthn.NewAuthentication(ctx, authClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("AuthN successfully connected to auth gRPC server " + authnHandler.Secure())

	domsGrpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&domsGrpcCfg, env.Options{Prefix: envPrefixDomains}); err != nil {
		logger.Error(fmt.Sprintf("failed to load domains gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	domAuthz, _, domainsHandler, err := domainsAuthz.NewAuthorization(ctx, domsGrpcCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer domainsHandler.Close()

	authz, authzHandler, err := authsvcAuthz.NewAuthorization(ctx, authClientCfg, domAuthz)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authzHandler.Close()
	logger.Info("AuthZ successfully connected to auth gRPC server " + authzHandler.Secure())

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("error shutting down tracer provider: %s", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	svc := newService(db, dbConfig, authz, logger, tracer)

	subscriber, err := store.NewSubscriber(ctx, cfg.ESURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create subscriber: %s", err))
		exitCode = 1
		return
	}

	logger.Info("Subscribed to Event Store")

	if err := events.Start(ctx, svcName, subscriber, svc); err != nil {
		logger.Error("failed to start %s service: %s", svcName, err)
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}

	hs := http.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, authn, logger, svcName, cfg.InstanceID), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(db *sqlx.DB, dbConfig pgclient.Config, authz mitrasauthz.Authorization, logger *slog.Logger, tracer trace.Tracer) journal.Service {
	database := postgres.NewDatabase(db, dbConfig, tracer)
	repo := journalpg.NewRepository(database)
	idp := uuid.New()

	svc := journal.NewService(idp, repo)
	svc = middleware.AuthorizationMiddleware(svc, authz)
	svc = middleware.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("journal", "journal_writer")
	svc = middleware.MetricsMiddleware(svc, counter, latency)
	svc = middleware.Tracing(svc, tracer)

	return svc
}
