// Package main contains users main function to start the users service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/hantdev/mitras/internal/email"
	grpcDomainsV1 "github.com/hantdev/mitras/internal/grpc/domains/v1"
	grpcTokenV1 "github.com/hantdev/mitras/internal/grpc/token/v1"
	smqlog "github.com/hantdev/mitras/logger"
	authsvcAuthn "github.com/hantdev/mitras/pkg/authn/authsvc"
	smqauthz "github.com/hantdev/mitras/pkg/authz"
	authsvcAuthz "github.com/hantdev/mitras/pkg/authz/authsvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	jaegerclient "github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/oauth2"
	googleoauth "github.com/hantdev/mitras/pkg/oauth2/google"
	"github.com/hantdev/mitras/pkg/policies"
	"github.com/hantdev/mitras/pkg/policies/spicedb"
	pg "github.com/hantdev/mitras/pkg/postgres"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/hantdev/mitras/users"
	"github.com/hantdev/mitras/users/api"
	"github.com/hantdev/mitras/users/emailer"
	"github.com/hantdev/mitras/users/events"
	"github.com/hantdev/mitras/users/hasher"
	"github.com/hantdev/mitras/users/middleware"
	"github.com/hantdev/mitras/users/postgres"
	"github.com/hantdev/mitras/users/tracing"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	svcName          = "users"
	envPrefixDB      = "MITRAS_USERS_DB_"
	envPrefixHTTP    = "MITRAS_USERS_HTTP_"
	envPrefixAuth    = "MITRAS_AUTH_GRPC_"
	envPrefixDomains = "MITRAS_DOMAINS_GRPC_"
	envPrefixGoogle  = "MITRAS_GOOGLE_"
	defDB            = "users"
	defSvcHTTPPort   = "9002"
)

type config struct {
	LogLevel            string        `env:"MITRAS_USERS_LOG_LEVEL"           envDefault:"info"`
	AdminEmail          string        `env:"MITRAS_USERS_ADMIN_EMAIL"         envDefault:"admin@example.com"`
	AdminPassword       string        `env:"MITRAS_USERS_ADMIN_PASSWORD"      envDefault:"12345678"`
	AdminUsername       string        `env:"MITRAS_USERS_ADMIN_USERNAME"      envDefault:"admin"`
	AdminFirstName      string        `env:"MITRAS_USERS_ADMIN_FIRST_NAME"    envDefault:"super"`
	AdminLastName       string        `env:"MITRAS_USERS_ADMIN_LAST_NAME"     envDefault:"admin"`
	PassRegexText       string        `env:"MITRAS_USERS_PASS_REGEX"          envDefault:"^.{8,}$"`
	ResetURL            string        `env:"MITRAS_TOKEN_RESET_ENDPOINT"      envDefault:"/reset-request"`
	JaegerURL           url.URL       `env:"MITRAS_JAEGER_URL"                envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry       bool          `env:"MITRAS_SEND_TELEMETRY"            envDefault:"true"`
	InstanceID          string        `env:"MITRAS_USERS_INSTANCE_ID"         envDefault:""`
	ESURL               string        `env:"MITRAS_ES_URL"                    envDefault:"nats://localhost:4222"`
	TraceRatio          float64       `env:"MITRAS_JAEGER_TRACE_RATIO"        envDefault:"1.0"`
	SelfRegister        bool          `env:"MITRAS_USERS_ALLOW_SELF_REGISTER" envDefault:"false"`
	OAuthUIRedirectURL  string        `env:"MITRAS_OAUTH_UI_REDIRECT_URL"     envDefault:"http://localhost:9095/domains"`
	OAuthUIErrorURL     string        `env:"MITRAS_OAUTH_UI_ERROR_URL"        envDefault:"http://localhost:9095/error"`
	DeleteInterval      time.Duration `env:"MITRAS_USERS_DELETE_INTERVAL"     envDefault:"24h"`
	DeleteAfter         time.Duration `env:"MITRAS_USERS_DELETE_AFTER"        envDefault:"720h"`
	SpicedbHost         string        `env:"MITRAS_SPICEDB_HOST"              envDefault:"localhost"`
	SpicedbPort         string        `env:"MITRAS_SPICEDB_PORT"              envDefault:"50051"`
	SpicedbPreSharedKey string        `env:"MITRAS_SPICEDB_PRE_SHARED_KEY"    envDefault:"12345678"`
	PassRegex           *regexp.Regexp
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}
	passRegex, err := regexp.Compile(cfg.PassRegexText)
	if err != nil {
		log.Fatalf("invalid password validation rules %s\n", cfg.PassRegexText)
	}
	cfg.PassRegex = passRegex

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

	ec := email.Config{}
	if err := env.Parse(&ec); err != nil {
		logger.Error(fmt.Sprintf("failed to load email configuration : %s", err.Error()))
		exitCode = 1
		return
	}

	dbConfig := pgclient.Config{Name: defDB}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}

	migration := postgres.Migration()
	db, err := pgclient.Setup(dbConfig, *migration)
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

	tokenClient, tokenHandler, err := grpcclient.SetupTokenClient(ctx, authClientConfig)
	if err != nil {
		logger.Error("failed to create token gRPC client " + err.Error())
		exitCode = 1
		return
	}
	defer tokenHandler.Close()
	logger.Info("Token service client successfully connected to auth gRPC server " + tokenHandler.Secure())

	authn, authnHandler, err := authsvcAuthn.NewAuthentication(ctx, authClientConfig)
	if err != nil {
		logger.Error("failed to create authn " + err.Error())
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("AuthN successfully connected to auth gRPC server " + authnHandler.Secure())

	authz, authzHandler, err := authsvcAuthz.NewAuthorization(ctx, authClientConfig)
	if err != nil {
		logger.Error("failed to create authz " + err.Error())
		exitCode = 1
		return
	}
	defer authzHandler.Close()
	logger.Info("AuthZ successfully connected to auth gRPC server " + authzHandler.Secure())

	domainsClientConfig := grpcclient.Config{}
	if err := env.ParseWithOptions(&domainsClientConfig, env.Options{Prefix: envPrefixDomains}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	domainsClient, domainsHandler, err := grpcclient.SetupDomainsClient(ctx, domainsClientConfig)
	if err != nil {
		logger.Error("failed to setup domain gRPC clients " + err.Error())
		exitCode = 1
		return
	}
	defer domainsHandler.Close()
	logger.Info("DomainsService gRPC client successfully connected to domains gRPC server " + domainsHandler.Secure())

	policyService, err := newPolicyService(cfg, logger)
	if err != nil {
		logger.Error("failed to create new policies service " + err.Error())
		exitCode = 1
		return
	}
	logger.Info("Policy client successfully connected to spicedb gRPC server")

	csvc, err := newService(ctx, authz, tokenClient, policyService, domainsClient, db, dbConfig, tracer, cfg, ec, logger)
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

	oauthConfig := oauth2.Config{}
	if err := env.ParseWithOptions(&oauthConfig, env.Options{Prefix: envPrefixGoogle}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s Google configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	oauthProvider := googleoauth.NewProvider(oauthConfig, cfg.OAuthUIRedirectURL, cfg.OAuthUIErrorURL)

	mux := chi.NewRouter()
	httpSrv := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(csvc, authn, tokenClient, cfg.SelfRegister, mux, logger, cfg.InstanceID, cfg.PassRegex, oauthProvider), logger)

	g.Go(func() error {
		return httpSrv.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSrv)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("users service terminated: %s", err))
	}
}

func newService(ctx context.Context, authz smqauthz.Authorization, token grpcTokenV1.TokenServiceClient, policyService policies.Service, domainsClient grpcDomainsV1.DomainsServiceClient, db *sqlx.DB, dbConfig pgclient.Config, tracer trace.Tracer, c config, ec email.Config, logger *slog.Logger) (users.Service, error) {
	database := pg.NewDatabase(db, dbConfig, tracer)
	idp := uuid.New()
	hsr := hasher.New()

	// Creating users service
	repo := postgres.NewRepository(database)
	emailerClient, err := emailer.New(c.ResetURL, &ec)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to configure e-mailing util: %s", err.Error()))
	}

	svc := users.NewService(token, repo, policyService, emailerClient, hsr, idp)

	svc, err = events.NewEventStoreMiddleware(ctx, svc, c.ESURL)
	if err != nil {
		return nil, err
	}
	svc = middleware.AuthorizationMiddleware(svc, authz, c.SelfRegister)

	svc = tracing.New(svc, tracer)
	svc = middleware.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics(svcName, "api")
	svc = middleware.MetricsMiddleware(svc, counter, latency)

	userID, err := createAdmin(ctx, c, repo, hsr, svc)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create admin client: %s", err))
	}
	if err := createAdminPolicy(ctx, userID, authz, policyService); err != nil {
		return nil, err
	}

	users.NewDeleteHandler(ctx, repo, policyService, domainsClient, c.DeleteInterval, c.DeleteAfter, logger)

	return svc, err
}

func createAdmin(ctx context.Context, c config, repo users.Repository, hsr users.Hasher, svc users.Service) (string, error) {
	id, err := uuid.New().ID()
	if err != nil {
		return "", err
	}
	hash, err := hsr.Hash(c.AdminPassword)
	if err != nil {
		return "", err
	}

	user := users.User{
		ID:        id,
		Email:     c.AdminEmail,
		FirstName: c.AdminFirstName,
		LastName:  c.AdminLastName,
		Credentials: users.Credentials{
			Username: "admin",
			Secret:   hash,
		},
		Metadata: users.Metadata{
			"role": "admin",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      users.AdminRole,
		Status:    users.EnabledStatus,
	}

	if u, err := repo.RetrieveByEmail(ctx, user.Email); err == nil {
		return u.ID, nil
	}

	// Create an admin
	if _, err = repo.Save(ctx, user); err != nil {
		return "", err
	}
	if _, err = svc.IssueToken(ctx, c.AdminUsername, c.AdminPassword); err != nil {
		return "", err
	}
	return user.ID, nil
}

func createAdminPolicy(ctx context.Context, userID string, authz smqauthz.Authorization, policyService policies.Service) error {
	if err := authz.Authorize(ctx, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     userID,
		Permission:  policies.AdministratorRelation,
		Object:      policies.MitrasObject,
		ObjectType:  policies.PlatformType,
	}); err != nil {
		err := policyService.AddPolicy(ctx, policies.Policy{
			SubjectType: policies.UserType,
			Subject:     userID,
			Relation:    policies.AdministratorRelation,
			Object:      policies.MitrasObject,
			ObjectType:  policies.PlatformType,
		})
		if err != nil {
			return err
		}
	}
	return nil
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
