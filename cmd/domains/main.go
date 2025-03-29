package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	grpcDomainsV1 "github.com/hantdev/mitras/api/grpc/domains/v1"
	"github.com/hantdev/mitras/domains"
	domainsSvc "github.com/hantdev/mitras/domains"
	domainsgrpcapi "github.com/hantdev/mitras/domains/api/grpc"
	httpapi "github.com/hantdev/mitras/domains/api/http"
	cache "github.com/hantdev/mitras/domains/cache"
	"github.com/hantdev/mitras/domains/events"
	dmw "github.com/hantdev/mitras/domains/middleware"
	dpostgres "github.com/hantdev/mitras/domains/postgres"
	"github.com/hantdev/mitras/domains/private"
	dtracing "github.com/hantdev/mitras/domains/tracing"
	redisclient "github.com/hantdev/mitras/internal/clients/redis"
	mitraslog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	"github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	domainsAuthz "github.com/hantdev/mitras/pkg/domains/psvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	"github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/hantdev/mitras/pkg/policies/spicedb"
	"github.com/hantdev/mitras/pkg/postgres"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/roles"
	"github.com/hantdev/mitras/pkg/server"
	grpcserver "github.com/hantdev/mitras/pkg/server/grpc"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/sid"
	spicedbdecoder "github.com/hantdev/mitras/pkg/spicedb"
	"github.com/hantdev/mitras/pkg/uuid"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	svcName        = "domains"
	envPrefixHTTP  = "MITRAS_DOMAINS_HTTP_"
	envPrefixGrpc  = "MITRAS_DOMAINS_GRPC_"
	envPrefixDB    = "MITRAS_DOMAINS_DB_"
	envPrefixAuth  = "MITRAS_AUTH_GRPC_"
	defDB          = "domains"
	defSvcHTTPPort = "9004"
	defSvcGRPCPort = "7004"
)

type config struct {
	LogLevel            string        `env:"MITRAS_DOMAINS_LOG_LEVEL"            envDefault:"info"`
	JaegerURL           url.URL       `env:"MITRAS_JAEGER_URL"                   envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool          `env:"MITRAS_SEND_TELEMETRY"               envDefault:"false"`
	CacheURL            string        `env:"MITRAS_DOMAINS_CACHE_URL"            envDefault:"redis://localhost:6379/0"`
	CacheKeyDuration    time.Duration `env:"MITRAS_DOMAINS_CACHE_KEY_DURATION"   envDefault:"10m"`
	InstanceID          string        `env:"MITRAS_DOMAINS_INSTANCE_ID"          envDefault:""`
	SpicedbHost         string        `env:"MITRAS_SPICEDB_HOST"                 envDefault:"localhost"`
	SpicedbPort         string        `env:"MITRAS_SPICEDB_PORT"                 envDefault:"50051"`
	SpicedbSchemaFile   string        `env:"MITRAS_SPICEDB_SCHEMA_FILE"          envDefault:"schema.zed"`
	SpicedbPreSharedKey string        `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"       envDefault:"12345678"`
	TraceRatio          float64       `env:"MITRAS_JAEGER_TRACE_RATIO"           envDefault:"1.0"`
	ESURL               string        `env:"MITRAS_ES_URL"                       envDefault:"nats://localhost:4222"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := mitraslog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
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
	}

	dm, err := dpostgres.Migration()
	if err != nil {
		logger.Error(fmt.Sprintf("failed create migrations for domain: %s", err.Error()))
		exitCode = 1
		return
	}

	db, err := pgclient.Setup(dbConfig, *dm)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

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

	time.Sleep(1 * time.Second)

	clientConfig := grpcclient.Config{}
	if err := env.ParseWithOptions(&clientConfig, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC server configuration : %s", err))
		exitCode = 1
		return
	}

	authn, authnHandler, err := authsvcAuthn.NewAuthentication(ctx, clientConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("authn failed to connect to auth gRPC server : %s", err.Error()))
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("Authn successfully connected to auth gRPC server " + authnHandler.Secure())

	database := postgres.NewDatabase(db, dbConfig, tracer)
	domainsRepo := dpostgres.NewRepository(database)

	cacheclient, err := redisclient.Connect(cfg.CacheURL)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer cacheclient.Close()
	cache := cache.NewDomainsCache(cacheclient, cfg.CacheKeyDuration)

	psvc := private.New(domainsRepo, cache)

	domAuthz := domainsAuthz.NewAuthorization(psvc)

	authz, authzHandler, err := authsvcAuthz.NewAuthorization(ctx, clientConfig, domAuthz)
	if err != nil {
		logger.Error(fmt.Sprintf("authz failed to connect to auth gRPC server : %s", err.Error()))
		exitCode = 1
		return
	}
	defer authzHandler.Close()
	logger.Info("Authz successfully connected to auth gRPC server " + authzHandler.Secure())

	policyService, err := newPolicyService(cfg, logger)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	logger.Info("Policy client successfully connected to spicedb gRPC server")

	svc, err := newDomainService(ctx, domainsRepo, cache, tracer, cfg, authz, policyService, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create %s service: %s", svcName, err.Error()))
		exitCode = 1
		return
	}

	grpcServerConfig := server.Config{Port: defSvcGRPCPort}
	if err := env.ParseWithOptions(&grpcServerConfig, env.Options{Prefix: envPrefixGrpc}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	registerDomainsServiceServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		grpcDomainsV1.RegisterDomainsServiceServer(srv, domainsgrpcapi.NewDomainsServer(psvc))
	}

	gs := grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerDomainsServiceServer, logger)

	g.Go(func() error {
		return gs.Start()
	})

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	mux := chi.NewMux()
	idp := uuid.New()
	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, authn, mux, logger, cfg.InstanceID, idp), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("domains service terminated: %s", err))
	}
}

func newDomainService(ctx context.Context, domainsRepo domainsSvc.Repository, cache domainsSvc.Cache, tracer trace.Tracer, cfg config, authz authz.Authorization, policiessvc policies.Service, logger *slog.Logger) (domains.Service, error) {
	idProvider := uuid.New()
	sidProvider, err := sid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to init short id provider : %w", err)
	}

	availableActions, builtInRoles, err := availableActionsAndBuiltInRoles(cfg.SpicedbSchemaFile)
	if err != nil {
		return nil, err
	}

	svc, err := domainsSvc.New(domainsRepo, cache, policiessvc, idProvider, sidProvider, availableActions, builtInRoles)
	if err != nil {
		return nil, fmt.Errorf("failed to init domain service: %w", err)
	}
	svc, err = events.NewEventStoreMiddleware(ctx, svc, cfg.ESURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init domain event store middleware: %w", err)
	}

	svc, err = dmw.AuthorizationMiddleware(policies.DomainType, svc, authz, domains.NewOperationPermissionMap(), domains.NewRolesOperationPermissionMap())
	if err != nil {
		return nil, err
	}

	counter, latency := prometheus.MakeMetrics("domains", "api")
	svc = dmw.MetricsMiddleware(svc, counter, latency)

	svc = dmw.LoggingMiddleware(svc, logger)

	svc = dtracing.New(svc, tracer)
	return svc, nil
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

func availableActionsAndBuiltInRoles(spicedbSchemaFile string) ([]roles.Action, map[roles.BuiltInRoleName][]roles.Action, error) {
	availableActions, err := spicedbdecoder.GetActionsFromSchema(spicedbSchemaFile, policies.DomainType)
	if err != nil {
		return []roles.Action{}, map[roles.BuiltInRoleName][]roles.Action{}, err
	}

	builtInRoles := map[roles.BuiltInRoleName][]roles.Action{
		domains.BuiltInRoleAdmin: availableActions,
	}

	return availableActions, builtInRoles, err
}
