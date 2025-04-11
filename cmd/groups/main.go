// Package main contains groups main function to start the groups service.
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
	"github.com/go-chi/chi/v5"
	"github.com/hantdev/mitras/groups"
	gpsvc "github.com/hantdev/mitras/groups"
	grpcapi "github.com/hantdev/mitras/groups/api/grpc"
	httpapi "github.com/hantdev/mitras/groups/api/http"
	"github.com/hantdev/mitras/groups/events"
	"github.com/hantdev/mitras/groups/middleware"
	"github.com/hantdev/mitras/groups/postgres"
	pgroups "github.com/hantdev/mitras/groups/private"
	"github.com/hantdev/mitras/groups/tracing"
	grpcChannelsV1 "github.com/hantdev/mitras/internal/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/internal/grpc/clients/v1"
	grpcGroupsV1 "github.com/hantdev/mitras/internal/grpc/groups/v1"
	smqlog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	smqauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
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
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

const (
	svcName           = "groups"
	envPrefixDB       = "MITRAS_GROUPS_DB_"
	envPrefixHTTP     = "MITRAS_GROUPS_HTTP_"
	envPrefixgRPC     = "MITRAS_GROUPS_GRPC_"
	envPrefixAuth     = "MITRAS_AUTH_GRPC_"
	envPrefixDomains  = "MITRAS_DOMAINS_GRPC_"
	envPrefixChannels = "MITRAS_CHANNELS_GRPC_"
	envPrefixClients  = "MITRAS_CLIENTS_AUTH_GRPC_"
	defDB             = "groups"
	defSvcHTTPPort    = "9004"
	defSvcgRPCPort    = "7004"
)

type config struct {
	LogLevel            string  `env:"MITRAS_GROUPS_LOG_LEVEL"          envDefault:"info"`
	InstanceID          string  `env:"MITRAS_GROUPS_INSTANCE_ID"        envDefault:""`
	JaegerURL           url.URL `env:"MITRAS_JAEGER_URL"                envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool    `env:"MITRAS_SEND_TELEMETRY"            envDefault:"true"`
	ESURL               string  `env:"MITRAS_ES_URL"                    envDefault:"nats://localhost:4222"`
	TraceRatio          float64 `env:"MITRAS_JAEGER_TRACE_RATIO"        envDefault:"1.0"`
	SpicedbHost         string  `env:"MITRAS_SPICEDB_HOST"              envDefault:"localhost"`
	SpicedbPort         string  `env:"MITRAS_SPICEDB_PORT"              envDefault:"50051"`
	SpicedbSchemaFile   string  `env:"MITRAS_SPICEDB_SCHEMA_FILE"       envDefault:"schema.zed"`
	SpicedbPreSharedKey string  `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"    envDefault:"12345678"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
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

	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	gm, err := postgres.Migration()
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	db, err := pgclient.Setup(dbConfig, *gm)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer db.Close()

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
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

	authClientConfig := grpcclient.Config{}
	if err := env.ParseWithOptions(&authClientConfig, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	authn, authnHandler, err := authsvcAuthn.NewAuthentication(ctx, authClientConfig)
	if err != nil {
		logger.Error("failed to create authn " + err.Error())
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("Authn successfully connected to auth gRPC server " + authnHandler.Secure())

	authz, authzHandler, err := authsvcAuthz.NewAuthorization(ctx, authClientConfig)
	if err != nil {
		logger.Error("failed to create authz " + err.Error())
		exitCode = 1
		return
	}
	defer authzHandler.Close()
	logger.Info("Authz successfully connected to auth gRPC server " + authzHandler.Secure())

	policyService, err := newPolicyService(cfg, logger)
	if err != nil {
		logger.Error("failed to create new policies service " + err.Error())
		exitCode = 1
		return
	}
	logger.Info("Policy client successfully connected to spicedb gRPC server")

	chgrpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&chgrpcCfg, env.Options{Prefix: envPrefixChannels}); err != nil {
		logger.Error(fmt.Sprintf("failed to load channels gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	channelsClient, channelsHandler, err := grpcclient.SetupChannelsClient(ctx, chgrpcCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to channels gRPC server: %s", err))
		exitCode = 1
		return
	}
	defer channelsHandler.Close()
	logger.Info("Groups gRPC client successfully connected to channels gRPC server " + channelsHandler.Secure())

	thgrpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&thgrpcCfg, env.Options{Prefix: envPrefixClients}); err != nil {
		logger.Error(fmt.Sprintf("failed to load clients gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	clientsClient, clientsHandler, err := grpcclient.SetupClientsClient(ctx, thgrpcCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to clients gRPC server: %s", err))
		exitCode = 1
		return
	}
	defer clientsHandler.Close()
	logger.Info("Clients gRPC client successfully connected to clients gRPC server " + clientsHandler.Secure())

	svc, psvc, err := newService(ctx, authz, policyService, db, dbConfig, channelsClient, clientsClient, tracer, logger, cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to setup service: %s", err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}

	mux := chi.NewRouter()
	httpSrv := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, authn, mux, logger, cfg.InstanceID), logger)

	grpcServerConfig := server.Config{}
	if err := env.ParseWithOptions(&grpcServerConfig, env.Options{Prefix: envPrefixgRPC}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	registerGroupsServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		grpcGroupsV1.RegisterGroupsServiceServer(srv, grpcapi.NewServer(psvc))
	}
	gs := grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerGroupsServer, logger)

	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return httpSrv.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSrv)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("groups service terminated: %s", err))
	}
}

func newService(ctx context.Context, authz smqauthz.Authorization, policy policies.Service, db *sqlx.DB, dbConfig pgclient.Config, channels grpcChannelsV1.ChannelsServiceClient, clients grpcClientsV1.ClientsServiceClient, tracer trace.Tracer, logger *slog.Logger, c config) (groups.Service, pgroups.Service, error) {
	database := pg.NewDatabase(db, dbConfig, tracer)
	idp := uuid.New()
	sid, err := sid.New()
	if err != nil {
		return nil, nil, err
	}

	availableActions, builtInRoles, err := availableActionsAndBuiltInRoles(c.SpicedbSchemaFile)
	if err != nil {
		return nil, nil, err
	}

	// Creating groups service
	repo := postgres.New(database)
	svc, err := gpsvc.NewService(repo, policy, idp, channels, clients, sid, availableActions, builtInRoles)
	if err != nil {
		return nil, nil, err
	}
	svc, err = events.New(ctx, svc, c.ESURL)
	if err != nil {
		return nil, nil, err
	}

	svc, err = middleware.AuthorizationMiddleware(policies.GroupType, svc, repo, authz, groups.NewOperationPermissionMap(), groups.NewRolesOperationPermissionMap(), groups.NewExternalOperationPermissionMap())
	if err != nil {
		return nil, nil, err
	}

	svc = tracing.New(svc, tracer)
	svc = middleware.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("groups", "api")
	svc = middleware.MetricsMiddleware(svc, counter, latency)

	psvc := pgroups.New(repo)
	return svc, psvc, err
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
	availableActions, err := spicedbdecoder.GetActionsFromSchema(spicedbSchemaFile, policies.GroupType)
	if err != nil {
		return []roles.Action{}, map[roles.BuiltInRoleName][]roles.Action{}, err
	}

	builtInRoles := map[roles.BuiltInRoleName][]roles.Action{
		groups.BuiltInRoleAdmin: availableActions,
	}

	return availableActions, builtInRoles, err
}
