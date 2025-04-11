// Package main contains postgres-reader main function to start the postgres-reader service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	smqlog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/authn/authsvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	pgclient "github.com/hantdev/mitras/pkg/postgres"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/hantdev/mitras/readers"
	"github.com/hantdev/mitras/readers/api"
	"github.com/hantdev/mitras/readers/postgres"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

const (
	svcName           = "postgres-reader"
	envPrefixDB       = "MITRAS_POSTGRES_"
	envPrefixHTTP     = "MITRAS_POSTGRES_READER_HTTP_"
	envPrefixAuth     = "MITRAS_AUTH_GRPC_"
	envPrefixClients  = "MITRAS_CLIENTS_AUTH_GRPC_"
	envPrefixChannels = "MITRAS_CHANNELS_GRPC_"
	defDB             = "mitras"
	defSvcHTTPPort    = "9009"
)

type config struct {
	LogLevel      string `env:"MITRAS_POSTGRES_READER_LOG_LEVEL"     envDefault:"info"`
	SendTelemetry bool   `env:"MITRAS_SEND_TELEMETRY"                envDefault:"true"`
	InstanceID    string `env:"MITRAS_POSTGRES_READER_INSTANCE_ID"   envDefault:""`
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

	dbConfig := pgclient.Config{}
	if err := env.ParseWithOptions(&dbConfig, env.Options{Prefix: envPrefixDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	db, err := pgclient.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to setup postgres database : %s", err))
		exitCode = 1
		return
	}
	defer db.Close()

	clientsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&clientsClientCfg, env.Options{Prefix: envPrefixClients}); err != nil {
		logger.Error(fmt.Sprintf("failed to load clients gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	clientsClient, clientsHandler, err := grpcclient.SetupClientsClient(ctx, clientsClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer clientsHandler.Close()
	logger.Info("Clients service gRPC client successfully connected to clients gRPC server " + clientsHandler.Secure())

	channelsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&channelsClientCfg, env.Options{Prefix: envPrefixChannels}); err != nil {
		logger.Error(fmt.Sprintf("failed to load channels gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	channelsClient, channelsHandler, err := grpcclient.SetupChannelsClient(ctx, channelsClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer channelsHandler.Close()
	logger.Info("Channels service gRPC client successfully connected to channels gRPC server " + channelsHandler.Secure())

	authnCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&authnCfg, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	authn, authnHandler, err := authsvc.NewAuthentication(ctx, authnCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authnHandler.Close()
	logger.Info("authn successfully connected to auth gRPC server " + authnHandler.Secure())

	repo := newService(db, logger)

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(repo, authn, clientsClient, channelsClient, svcName, cfg.InstanceID), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Postgres reader service terminated: %s", err))
	}
}

func newService(db *sqlx.DB, logger *slog.Logger) readers.MessageRepository {
	svc := postgres.New(db)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("postgres", "message_reader")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
