// Package main contains bootstrap main function to start the bootstrap service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/caarlos0/env/v11"
	"github.com/hantdev/mitras/bootstrap"
	"github.com/hantdev/mitras/bootstrap/api"
	"github.com/hantdev/mitras/bootstrap/events/consumer"
	"github.com/hantdev/mitras/bootstrap/events/producer"
	"github.com/hantdev/mitras/bootstrap/middleware"
	bootstrappg "github.com/hantdev/mitras/bootstrap/postgres"
	"github.com/hantdev/mitras/bootstrap/tracing"
	smqlog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	smqauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	"github.com/hantdev/mitras/pkg/events"
	"github.com/hantdev/mitras/pkg/events/store"
	"github.com/hantdev/mitras/pkg/grpcclient"
	"github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/hantdev/mitras/pkg/policies/spicedb"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	mgsdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/hantdev/mitras/pkg/server"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	svcName        = "bootstrap"
	envPrefixDB    = "MITRAS_BOOTSTRAP_DB_"
	envPrefixHTTP  = "MITRAS_BOOTSTRAP_HTTP_"
	envPrefixAuth  = "MITRAS_AUTH_GRPC_"
	defDB          = "bootstrap"
	defSvcHTTPPort = "9013"

	stream   = "events.mitras.clients"
	streamID = "mitras.bootstrap"
)

type config struct {
	LogLevel            string  `env:"MITRAS_BOOTSTRAP_LOG_LEVEL"        envDefault:"info"`
	EncKey              string  `env:"MITRAS_BOOTSTRAP_ENCRYPT_KEY"      envDefault:"12345678910111213141516171819202"`
	ESConsumerName      string  `env:"MITRAS_BOOTSTRAP_EVENT_CONSUMER"   envDefault:"bootstrap"`
	ClientsURL          string  `env:"MITRAS_CLIENTS_URL"                envDefault:"http://localhost:9000"`
	JaegerURL           url.URL `env:"MITRAS_JAEGER_URL"                 envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool    `env:"MITRAS_SEND_TELEMETRY"             envDefault:"true"`
	InstanceID          string  `env:"MITRAS_BOOTSTRAP_INSTANCE_ID"      envDefault:""`
	ESURL               string  `env:"MITRAS_ES_URL"                     envDefault:"nats://localhost:4222"`
	TraceRatio          float64 `env:"MITRAS_JAEGER_TRACE_RATIO"         envDefault:"1.0"`
	SpicedbHost         string  `env:"MITRAS_SPICEDB_HOST"               envDefault:"localhost"`
	SpicedbPort         string  `env:"MITRAS_SPICEDB_PORT"               envDefault:"50051"`
	SpicedbPreSharedKey string  `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"     envDefault:"12345678"`
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

	// Create new postgres client
	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
	}
	db, err := pgclient.Setup(dbConfig, *bootstrappg.Migration())
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	policySvc, err := newPolicyService(cfg, logger)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	logger.Info("Policy client successfully connected to spicedb gRPC server")

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

	grpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&grpcCfg, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	authn, authnClient, err := authsvcAuthn.NewAuthentication(ctx, grpcCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	logger.Info("AuthN successfully connected to auth gRPC server " + authnClient.Secure())
	defer authnClient.Close()

	authz, authzClient, err := authsvcAuthz.NewAuthorization(ctx, grpcCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authzClient.Close()
	logger.Info("AuthZ successfully connected to auth gRPC server " + authzClient.Secure())

	// Create new service
	svc, err := newService(ctx, authz, policySvc, db, tracer, logger, cfg, dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create %s service: %s", svcName, err))
		exitCode = 1
		return
	}

	if err = subscribeToClientsES(ctx, svc, cfg, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to subscribe to clients event store: %s", err))
		exitCode = 1
		return
	}

	logger.Info("Subscribed to Event Store")

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, authn, bootstrap.NewConfigReader([]byte(cfg.EncKey)), logger, cfg.InstanceID), logger)

	// Start servers
	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Bootstrap service terminated: %s", err))
	}
}

func newService(ctx context.Context, authz smqauthz.Authorization, policySvc policies.Service, db *sqlx.DB, tracer trace.Tracer, logger *slog.Logger, cfg config, dbConfig pgclient.Config) (bootstrap.Service, error) {
	database := pgclient.NewDatabase(db, dbConfig, tracer)

	repoConfig := bootstrappg.NewConfigRepository(database, logger)

	config := mgsdk.Config{
		ClientsURL: cfg.ClientsURL,
	}

	sdk := mgsdk.NewSDK(config)
	idp := uuid.New()

	svc := bootstrap.New(policySvc, repoConfig, sdk, []byte(cfg.EncKey), idp)

	publisher, err := store.NewPublisher(ctx, cfg.ESURL, streamID)
	if err != nil {
		return nil, err
	}

	svc = middleware.AuthorizationMiddleware(svc, authz)
	svc = producer.NewEventStoreMiddleware(svc, publisher)
	svc = middleware.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics(svcName, "api")
	svc = middleware.MetricsMiddleware(svc, counter, latency)
	svc = tracing.New(svc, tracer)

	return svc, nil
}

func subscribeToClientsES(ctx context.Context, svc bootstrap.Service, cfg config, logger *slog.Logger) error {
	subscriber, err := store.NewSubscriber(ctx, cfg.ESURL, logger)
	if err != nil {
		return err
	}

	subConfig := events.SubscriberConfig{
		Stream:   stream,
		Consumer: cfg.ESConsumerName,
		Handler:  consumer.NewEventHandler(svc),
	}
	return subscriber.Subscribe(ctx, subConfig)
}

func newPolicyService(cfg config, logger *slog.Logger) (policies.Service, error) {
	client, err := authzed.NewClientWithExperimentalAPIs(
		fmt.Sprintf("%s:%s", cfg.SpicedbHost, cfg.SpicedbPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(cfg.SpicedbPreSharedKey),
	)
	if err != nil {
		return nil, err
	}
	policySvc := spicedb.NewPolicyService(client, logger)

	return policySvc, nil
}
