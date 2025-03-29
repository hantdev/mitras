package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/caarlos0/env/v11"
	grpcAuthV1 "github.com/hantdev/mitras/api/grpc/auth/v1"
	grpcTokenV1 "github.com/hantdev/mitras/api/grpc/token/v1"
	"github.com/hantdev/mitras/auth"
	api "github.com/hantdev/mitras/auth/api"
	authgrpcapi "github.com/hantdev/mitras/auth/api/grpc/auth"
	tokengrpcapi "github.com/hantdev/mitras/auth/api/grpc/token"
	httpapi "github.com/hantdev/mitras/auth/api/http"
	"github.com/hantdev/mitras/auth/cache"
	"github.com/hantdev/mitras/auth/hasher"
	"github.com/hantdev/mitras/auth/jwt"
	apostgres "github.com/hantdev/mitras/auth/postgres"
	"github.com/hantdev/mitras/auth/tracing"
	redisclient "github.com/hantdev/mitras/internal/clients/redis"
	mitraslog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/policies/spicedb"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	grpcserver "github.com/hantdev/mitras/pkg/server/grpc"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
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
	svcName        = "auth"
	envPrefixHTTP  = "MITRAS_AUTH_HTTP_"
	envPrefixGrpc  = "MITRAS_AUTH_GRPC_"
	envPrefixDB    = "MITRAS_AUTH_DB_"
	defDB          = "auth"
	defSvcHTTPPort = "8189"
	defSvcGRPCPort = "8181"
)

type config struct {
	LogLevel                   string        `env:"MITRAS_AUTH_LOG_LEVEL"                envDefault:"info"`
	SecretKey                  string        `env:"MITRAS_AUTH_SECRET_KEY"               envDefault:"secret"`
	JaegerURL                  url.URL       `env:"MITRAS_JAEGER_URL"                    envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry              bool          `env:"MITRAS_SEND_TELEMETRY"                envDefault:"false"`
	InstanceID                 string        `env:"MITRAS_AUTH_ADAPTER_INSTANCE_ID"      envDefault:""`
	AccessDuration             time.Duration `env:"MITRAS_AUTH_ACCESS_TOKEN_DURATION"    envDefault:"1h"`
	RefreshDuration            time.Duration `env:"MITRAS_AUTH_REFRESH_TOKEN_DURATION"   envDefault:"24h"`
	InvitationDuration         time.Duration `env:"MITRAS_AUTH_INVITATION_DURATION"      envDefault:"168h"`
	SpicedbHost                string        `env:"MITRAS_SPICEDB_HOST"                  envDefault:"localhost"`
	SpicedbPort                string        `env:"MITRAS_SPICEDB_PORT"                  envDefault:"50051"`
	SpicedbSchemaFile          string        `env:"MITRAS_SPICEDB_SCHEMA_FILE"           envDefault:"./docker/spicedb/schema.zed"`
	SpicedbPreSharedKey        string        `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"        envDefault:"12345678"`
	TraceRatio                 float64       `env:"MITRAS_JAEGER_TRACE_RATIO"            envDefault:"1.0"`
	ESURL                      string        `env:"MITRAS_ES_URL"                        envDefault:"nats://localhost:4222"`
	CacheURL                   string        `env:"MITRAS_AUTH_CACHE_URL"                envDefault:"redis://localhost:6379/0"`
	CacheKeyDuration           time.Duration `env:"MITRAS_AUTH_CACHE_KEY_DURATION"       envDefault:"10m"`
	AuthCalloutURLs            []string      `env:"MITRAS_AUTH_CALLOUT_URLS"             envDefault:"" envSeparator:","`
	AuthCalloutMethod          string        `env:"MITRAS_AUTH_CALLOUT_METHOD"           envDefault:"POST"`
	AuthCalloutTLSVerification bool          `env:"MITRAS_AUTH_CALLOUT_TLS_VERIFICATION" envDefault:"true"`
	AuthCalloutTimeout         time.Duration `env:"MITRAS_AUTH_CALLOUT_TIMEOUT"          envDefault:"10s"`
	AuthCalloutCACert          string        `env:"MITRAS_AUTH_CALLOUT_CA_CERT"          envDefault:""`
	AuthCalloutCert            string        `env:"MITRAS_AUTH_CALLOUT_CERT"             envDefault:""`
	AuthCalloutKey             string        `env:"MITRAS_AUTH_CALLOUT_KEY"              envDefault:""`
	AuthCalloutPermissions     []string      `env:"MITRAS_AUTH_CALLOUT_INVOKE_PERMISSIONS" envDefault:"" envSeparator:","`
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

	cacheclient, err := redisclient.Connect(cfg.CacheURL)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer cacheclient.Close()

	am := apostgres.Migration()
	db, err := pgclient.Setup(dbConfig, *am)
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

	spicedbclient, err := initSpiceDB(ctx, cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init spicedb grpc client : %s\n", err.Error()))
		exitCode = 1
		return
	}

	svc, err := newService(db, tracer, cfg, dbConfig, logger, spicedbclient, cacheclient, cfg.CacheKeyDuration)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create service : %s\n", err.Error()))
		exitCode = 1
		return
	}

	grpcServerConfig := server.Config{Port: defSvcGRPCPort}
	if err := env.ParseWithOptions(&grpcServerConfig, env.Options{Prefix: envPrefixGrpc}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	registerAuthServiceServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		grpcTokenV1.RegisterTokenServiceServer(srv, tokengrpcapi.NewTokenServer(svc))
		grpcAuthV1.RegisterAuthServiceServer(srv, authgrpcapi.NewAuthServer(svc))
	}

	gs := grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerAuthServiceServer, logger)

	g.Go(func() error {
		return gs.Start()
	})

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, logger, cfg.InstanceID), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("users service terminated: %s", err))
	}
}

func initSpiceDB(ctx context.Context, cfg config) (*authzed.ClientWithExperimental, error) {
	client, err := authzed.NewClientWithExperimentalAPIs(
		fmt.Sprintf("%s:%s", cfg.SpicedbHost, cfg.SpicedbPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(cfg.SpicedbPreSharedKey),
	)
	if err != nil {
		return client, err
	}

	if err := initSchema(ctx, client, cfg.SpicedbSchemaFile); err != nil {
		return client, err
	}

	return client, nil
}

func initSchema(ctx context.Context, client *authzed.ClientWithExperimental, schemaFilePath string) error {
	schemaContent, err := os.ReadFile(schemaFilePath)
	if err != nil {
		return fmt.Errorf("failed to read spice db schema file : %w", err)
	}

	if _, err = client.SchemaServiceClient.WriteSchema(ctx, &v1.WriteSchemaRequest{Schema: string(schemaContent)}); err != nil {
		return fmt.Errorf("failed to create schema in spicedb : %w", err)
	}

	return nil
}

func newService(db *sqlx.DB, tracer trace.Tracer, cfg config, dbConfig pgclient.Config, logger *slog.Logger, spicedbClient *authzed.ClientWithExperimental, cacheClient *redis.Client, keyDuration time.Duration) (auth.Service, error) {
	cache := cache.NewPatsCache(cacheClient, keyDuration)

	database := pgclient.NewDatabase(db, dbConfig, tracer)
	keysRepo := apostgres.New(database)
	patsRepo := apostgres.NewPatRepo(database, cache)
	hasher := hasher.New()
	idProvider := uuid.New()

	pEvaluator := spicedb.NewPolicyEvaluator(spicedbClient, logger)
	pService := spicedb.NewPolicyService(spicedbClient, logger)

	t := jwt.New([]byte(cfg.SecretKey))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: !cfg.AuthCalloutTLSVerification,
	}
	if cfg.AuthCalloutCert != "" || cfg.AuthCalloutKey != "" {
		clientTLSCert, err := tls.LoadX509KeyPair(cfg.AuthCalloutCert, cfg.AuthCalloutKey)
		if err != nil {
			return nil, err
		}
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		caCert, err := os.ReadFile(cfg.AuthCalloutCACert)
		if err != nil {
			return nil, err
		}
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate")
		}
		tlsConfig.RootCAs = certPool
		tlsConfig.Certificates = []tls.Certificate{clientTLSCert}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: cfg.AuthCalloutTimeout,
	}
	callback, err := auth.NewCallback(httpClient, cfg.AuthCalloutMethod, cfg.AuthCalloutURLs, cfg.AuthCalloutPermissions)
	if err != nil {
		return nil, err
	}

	svc := auth.New(keysRepo, patsRepo, nil, hasher, idProvider, t, pEvaluator, pService, cfg.AccessDuration, cfg.RefreshDuration, cfg.InvitationDuration, callback)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("auth", "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	svc = tracing.New(svc, tracer)

	return svc, nil
}
