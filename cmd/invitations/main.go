// Package main contains invitations main function to start the invitations service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	grpcTokenV1 "github.com/hantdev/mitras/internal/grpc/token/v1"
	"github.com/hantdev/mitras/invitations"
	"github.com/hantdev/mitras/invitations/api"
	"github.com/hantdev/mitras/invitations/middleware"
	invitationspg "github.com/hantdev/mitras/invitations/postgres"
	smqlog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	smqauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	"github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/postgres"
	clientspg "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	mgsdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/hantdev/mitras/pkg/server"
	"github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "invitations"
	envPrefixDB    = "MITRAS_INVITATIONS_DB_"
	envPrefixHTTP  = "MITRAS_INVITATIONS_HTTP_"
	envPrefixAuth  = "MITRAS_AUTH_GRPC_"
	defDB          = "invitations"
	defSvcHTTPPort = "9020"
)

type config struct {
	LogLevel      string  `env:"MITRAS_INVITATIONS_LOG_LEVEL"      envDefault:"info"`
	UsersURL      string  `env:"MITRAS_USERS_URL"                  envDefault:"http://localhost:9002"`
	DomainsURL    string  `env:"MITRAS_DOMAINS_URL"                envDefault:"http://localhost:8189"`
	InstanceID    string  `env:"MITRAS_INVITATIONS_INSTANCE_ID"    envDefault:""`
	JaegerURL     url.URL `env:"MITRAS_JAEGER_URL"                 envDefault:"http://localhost:4318/v1/traces"`
	TraceRatio    float64 `env:"MITRAS_JAEGER_TRACE_RATIO"         envDefault:"1.0"`
	SendTelemetry bool    `env:"MITRAS_SEND_TELEMETRY"             envDefault:"true"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	dbConfig := clientspg.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s database configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	db, err := clientspg.Setup(dbConfig, *invitationspg.Migration())
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	authClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&authClientCfg, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC client configuration : %s", err.Error()))
		exitCode = 1
		return
	}
	tokenClient, tokenHandler, err := grpcclient.SetupTokenClient(ctx, authClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer tokenHandler.Close()
	logger.Info("Token service client successfully connected to auth gRPC server " + tokenHandler.Secure())

	authn, authnHandler, err := authsvcAuthn.NewAuthentication(ctx, authClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("AuthN successfully connected to auth gRPC server " + authnHandler.Secure())

	authz, authzHandler, err := authsvcAuthz.NewAuthorization(ctx, authClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authzHandler.Close()
	logger.Info("Authz successfully connected to auth gRPC server " + authzHandler.Secure())

	tp, err := jaeger.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	svc, err := newService(db, dbConfig, authz, tokenClient, tracer, cfg, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create %s service: %s", svcName, err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	httpSvr := http.NewServer(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, logger, authn, cfg.InstanceID), logger)

	g.Go(func() error {
		return httpSvr.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSvr)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(db *sqlx.DB, dbConfig clientspg.Config, authz smqauthz.Authorization, token grpcTokenV1.TokenServiceClient, tracer trace.Tracer, conf config, logger *slog.Logger) (invitations.Service, error) {
	database := postgres.NewDatabase(db, dbConfig, tracer)
	repo := invitationspg.NewRepository(database)

	config := mgsdk.Config{
		UsersURL:   conf.UsersURL,
		DomainsURL: conf.DomainsURL,
	}
	sdk := mgsdk.NewSDK(config)

	svc := invitations.NewService(token, repo, sdk)
	svc = middleware.AuthorizationMiddleware(authz, svc)
	svc = middleware.Tracing(svc, tracer)
	svc = middleware.Logging(logger, svc)
	counter, latency := prometheus.MakeMetrics(svcName, "api")
	svc = middleware.Metrics(counter, latency, svc)

	return svc, nil
}
