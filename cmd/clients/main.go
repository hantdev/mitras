// Package main contains clients main function to start the clients service.
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
	grpcChannelsV1 "github.com/hantdev/mitras/api/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/api/grpc/clients/v1"
	grpcGroupsV1 "github.com/hantdev/mitras/api/grpc/groups/v1"
	"github.com/hantdev/mitras/clients"
	grpcapi "github.com/hantdev/mitras/clients/api/grpc"
	httpapi "github.com/hantdev/mitras/clients/api/http"
	"github.com/hantdev/mitras/clients/cache"
	"github.com/hantdev/mitras/clients/events"
	"github.com/hantdev/mitras/clients/middleware"
	"github.com/hantdev/mitras/clients/postgres"
	pClients "github.com/hantdev/mitras/clients/private"
	"github.com/hantdev/mitras/clients/tracing"
	dpostgres "github.com/hantdev/mitras/domains/postgres"
	gpostgres "github.com/hantdev/mitras/groups/postgres"
	redisclient "github.com/hantdev/mitras/internal/clients/redis"
	mitraslog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	mitrasauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	dconsumer "github.com/hantdev/mitras/pkg/domains/events/consumer"
	domainsAuthz "github.com/hantdev/mitras/pkg/domains/grpcclient"
	gconsumer "github.com/hantdev/mitras/pkg/groups/events/consumer"
	"github.com/hantdev/mitras/pkg/grpcclient"
	jaegerclient "github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/hantdev/mitras/pkg/policies/spicedb"
	pg "github.com/hantdev/mitras/pkg/postgres"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/roles"
	"github.com/hantdev/mitras/pkg/server"
	grpcserver "github.com/hantdev/mitras/pkg/server/grpc"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/sid"
	spicedbdecoder "github.com/hantdev/mitras/pkg/spicedb"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	svcName            = "clients"
	envPrefixDB        = "MITRAS_CLIENTS_DB_"
	envPrefixHTTP      = "MITRAS_CLIENTS_HTTP_"
	envPrefixGRPC      = "MITRAS_CLIENTS_GRPC_"
	envPrefixAuth      = "MITRAS_AUTH_GRPC_"
	envPrefixChannels  = "MITRAS_CHANNELS_GRPC_"
	envPrefixGroups    = "MITRAS_GROUPS_GRPC_"
	envPrefixDomains   = "MITRAS_DOMAINS_GRPC_"
	defDB              = "clients"
	defSvcHTTPPort     = "9000"
	defSvcAuthGRPCPort = "7000"
)

type config struct {
	InstanceID          string        `env:"MITRAS_CLIENTS_INSTANCE_ID"        envDefault:""`
	LogLevel            string        `env:"MITRAS_CLIENTS_LOG_LEVEL"          envDefault:"info"`
	StandaloneID        string        `env:"MITRAS_CLIENTS_STANDALONE_ID"      envDefault:""`
	StandaloneToken     string        `env:"MITRAS_CLIENTS_STANDALONE_TOKEN"   envDefault:""`
	CacheURL            string        `env:"MITRAS_CLIENTS_CACHE_URL"          envDefault:"redis://localhost:6379/0"`
	CacheKeyDuration    time.Duration `env:"MITRAS_CLIENTS_CACHE_KEY_DURATION" envDefault:"10m"`
	JaegerURL           url.URL       `env:"MITRAS_JAEGER_URL"                 envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool          `env:"MITRAS_SEND_TELEMETRY"             envDefault:"fasle"`
	ESURL               string        `env:"MITRAS_ES_URL"                     envDefault:"nats://localhost:4222"`
	ESConsumerName      string        `env:"MITRAS_CLIENTS_EVENT_CONSUMER"     envDefault:"clients"`
	TraceRatio          float64       `env:"MITRAS_JAEGER_TRACE_RATIO"         envDefault:"1.0"`
	SpicedbHost         string        `env:"MITRAS_SPICEDB_HOST"               envDefault:"localhost"`
	SpicedbPort         string        `env:"MITRAS_SPICEDB_PORT"               envDefault:"50051"`
	SpicedbPreSharedKey string        `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"     envDefault:"12345678"`
	SpicedbSchemaFile   string        `env:"MITRAS_SPICEDB_SCHEMA_FILE"        envDefault:"schema.zed"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// Create new clients configuration
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	var logger *slog.Logger
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

	// Create new database for clients
	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	tm, err := postgres.Migration()
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	db, err := pgclient.Setup(dbConfig, *tm)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	// Setup new redis cache client
	cacheclient, err := redisclient.Connect(cfg.CacheURL)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer cacheclient.Close()

	policyEvaluator, policyService, err := newSpiceDBPolicyServiceEvaluator(cfg, logger)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	logger.Info("Policy evaluator and Policy manager are successfully connected to SpiceDB gRPC server")

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
	defer authnClient.Close()
	logger.Info("AuthN  successfully connected to auth gRPC server " + authnClient.Secure())

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

	authz, authzClient, err := authsvcAuthz.NewAuthorization(ctx, grpcCfg, domAuthz)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authzClient.Close()
	logger.Info("AuthZ  successfully connected to auth gRPC server " + authnClient.Secure())

	chgrpccfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&chgrpccfg, env.Options{Prefix: envPrefixChannels}); err != nil {
		logger.Error(fmt.Sprintf("failed to load channels gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	channelsgRPC, channelsClient, err := grpcclient.SetupChannelsClient(ctx, chgrpccfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	logger.Info("Channels gRPC client successfully connected to channels gRPC server " + channelsClient.Secure())
	defer channelsClient.Close()

	groupsgRPCCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&groupsgRPCCfg, env.Options{Prefix: envPrefixGroups}); err != nil {
		logger.Error(fmt.Sprintf("failed to load groups gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	groupsClient, groupsHandler, err := grpcclient.SetupGroupsClient(ctx, groupsgRPCCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to groups gRPC server: %s", err))
		exitCode = 1
		return
	}
	defer groupsHandler.Close()
	logger.Info("Groups gRPC client successfully connected to groups gRPC server " + groupsHandler.Secure())

	svc, psvc, err := newService(ctx, db, dbConfig, authz, policyEvaluator, policyService, cacheclient, cfg, channelsgRPC, groupsClient, tracer, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create services: %s", err))
		exitCode = 1
		return
	}

	ddatabase := pg.NewDatabase(db, dbConfig, tracer)
	drepo := dpostgres.NewRepository(ddatabase)

	if err := dconsumer.DomainsEventsSubscribe(ctx, drepo, cfg.ESURL, cfg.ESConsumerName, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to create domains event store : %s", err))
		exitCode = 1
		return
	}

	gdatabase := pg.NewDatabase(db, dbConfig, tracer)
	grepo := gpostgres.New(gdatabase)

	if err := gconsumer.GroupsEventsSubscribe(ctx, grepo, cfg.ESURL, cfg.ESConsumerName, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to create groups event store : %s", err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	mux := chi.NewRouter()
	idp := uuid.New()
	httpSvc := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, authn, mux, logger, cfg.InstanceID, idp), logger)

	grpcServerConfig := server.Config{Port: defSvcAuthGRPCPort}
	if err := env.ParseWithOptions(&grpcServerConfig, env.Options{Prefix: envPrefixGRPC}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	registerClientsServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		grpcClientsV1.RegisterClientsServiceServer(srv, grpcapi.NewServer(psvc))
	}
	gs := grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerClientsServer, logger)

	// Start all servers
	g.Go(func() error {
		return httpSvc.Start()
	})

	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSvc)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(ctx context.Context, db *sqlx.DB, dbConfig pgclient.Config, authz mitrasauthz.Authorization, pe policies.Evaluator, ps policies.Service, cacheClient *redis.Client, cfg config, channels grpcChannelsV1.ChannelsServiceClient, groups grpcGroupsV1.GroupsServiceClient, tracer trace.Tracer, logger *slog.Logger) (clients.Service, pClients.Service, error) {
	database := pg.NewDatabase(db, dbConfig, tracer)
	repo := postgres.NewRepository(database)

	idp := uuid.New()
	sidp, err := sid.New()
	if err != nil {
		return nil, nil, err
	}

	// Clients service
	cache := cache.NewCache(cacheClient, cfg.CacheKeyDuration)

	availableActions, builtInRoles, err := availableActionsAndBuiltInRoles(cfg.SpicedbSchemaFile)
	if err != nil {
		return nil, nil, err
	}

	csvc, err := clients.NewService(repo, ps, cache, channels, groups, idp, sidp, availableActions, builtInRoles)
	if err != nil {
		return nil, nil, err
	}

	csvc, err = events.NewEventStoreMiddleware(ctx, csvc, cfg.ESURL)
	if err != nil {
		return nil, nil, err
	}

	csvc = tracing.New(csvc, tracer)

	counter, latency := prometheus.MakeMetrics(svcName, "api")
	csvc = middleware.MetricsMiddleware(csvc, counter, latency)
	csvc = middleware.MetricsMiddleware(csvc, counter, latency)

	csvc, err = middleware.AuthorizationMiddleware(policies.ClientType, csvc, authz, repo, clients.NewOperationPermissionMap(), clients.NewRolesOperationPermissionMap(), clients.NewExternalOperationPermissionMap())
	if err != nil {
		return nil, nil, err
	}
	csvc = middleware.LoggingMiddleware(csvc, logger)

	isvc := pClients.New(repo, cache, pe, ps)

	return csvc, isvc, err
}

func newSpiceDBPolicyServiceEvaluator(cfg config, logger *slog.Logger) (policies.Evaluator, policies.Service, error) {
	client, err := authzed.NewClientWithExperimentalAPIs(
		fmt.Sprintf("%s:%s", cfg.SpicedbHost, cfg.SpicedbPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(cfg.SpicedbPreSharedKey),
	)
	if err != nil {
		return nil, nil, err
	}
	pe := spicedb.NewPolicyEvaluator(client, logger)
	ps := spicedb.NewPolicyService(client, logger)

	return pe, ps, nil
}

func availableActionsAndBuiltInRoles(spicedbSchemaFile string) ([]roles.Action, map[roles.BuiltInRoleName][]roles.Action, error) {
	availableActions, err := spicedbdecoder.GetActionsFromSchema(spicedbSchemaFile, policies.ClientType)
	if err != nil {
		return []roles.Action{}, map[roles.BuiltInRoleName][]roles.Action{}, err
	}

	builtInRoles := map[roles.BuiltInRoleName][]roles.Action{
		clients.BuiltInRoleAdmin: availableActions,
	}

	return availableActions, builtInRoles, err
}
