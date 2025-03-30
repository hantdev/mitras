// Package main contains http-adapter main function to start the http-adapter service.
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/hantdev/hermina"
	herminahttp "github.com/hantdev/hermina/pkg/http"
	"github.com/hantdev/hermina/pkg/session"
	grpcChannelsV1 "github.com/hantdev/mitras/api/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/api/grpc/clients/v1"
	adapter "github.com/hantdev/mitras/http"
	httpapi "github.com/hantdev/mitras/http/api"
	mitraslog "github.com/hantdev/mitras/logger"
	mitrasauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/authn/authsvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	jaegerclient "github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/messaging"
	"github.com/hantdev/mitras/pkg/messaging/brokers"
	brokerstracing "github.com/hantdev/mitras/pkg/messaging/brokers/tracing"
	msgevents "github.com/hantdev/mitras/pkg/messaging/events"
	"github.com/hantdev/mitras/pkg/messaging/handler"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName           = "http_adapter"
	envPrefix         = "MITRAS_HTTP_ADAPTER_"
	envPrefixClients  = "MITRAS_CLIENTS_GRPC_"
	envPrefixChannels = "MITRAS_CHANNELS_GRPC_"
	envPrefixAuth     = "MITRAS_AUTH_GRPC_"
	defSvcHTTPPort    = "80"
	targetHTTPPort    = "81"
	targetHTTPHost    = "http://localhost"
)

type config struct {
	LogLevel      string  `env:"MITRAS_HTTP_ADAPTER_LOG_LEVEL"   envDefault:"info"`
	BrokerURL     string  `env:"MITRAS_MESSAGE_BROKER_URL"       envDefault:"nats://localhost:4222"`
	JaegerURL     url.URL `env:"MITRAS_JAEGER_URL"               envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry bool    `env:"MITRAS_SEND_TELEMETRY"           envDefault:"false"`
	InstanceID    string  `env:"MITRAS_HTTP_ADAPTER_INSTANCE_ID" envDefault:""`
	TraceRatio    float64 `env:"MITRAS_JAEGER_TRACE_RATIO"       envDefault:"1.0"`
	ESURL         string  `env:"MITRAS_ES_URL"                   envDefault:"nats://localhost:4222"`
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

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefix}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

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

	pub, err := brokers.NewPublisher(ctx, cfg.BrokerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer pub.Close()
	pub = brokerstracing.NewPublisher(httpServerConfig, tracer, pub)

	pub, err = msgevents.NewPublisherMiddleware(ctx, pub, cfg.ESURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create event store middleware: %s", err))
		exitCode = 1
		return
	}

	svc := newService(pub, authn, clientsClient, channelsClient, logger, tracer)
	targetServerCfg := server.Config{Port: targetHTTPPort}

	hs := httpserver.NewServer(ctx, cancel, svcName, targetServerCfg, httpapi.MakeHandler(logger, cfg.InstanceID), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return proxyHTTP(ctx, httpServerConfig, logger, svc)
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("HTTP adapter service terminated: %s", err))
	}
}

func newService(pub messaging.Publisher, authn mitrasauthn.Authentication, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, logger *slog.Logger, tracer trace.Tracer) session.Handler {
	svc := adapter.NewHandler(pub, authn, clients, channels, logger)
	svc = handler.NewTracing(tracer, svc)
	svc = handler.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics(svcName, "api")
	svc = handler.MetricsMiddleware(svc, counter, latency)
	return svc
}

func proxyHTTP(ctx context.Context, cfg server.Config, logger *slog.Logger, sessionHandler session.Handler) error {
	config := hermina.Config{
		Address:    fmt.Sprintf("%s:%s", "", cfg.Port),
		Target:     fmt.Sprintf("%s:%s", targetHTTPHost, targetHTTPPort),
		PathPrefix: "/",
	}
	if cfg.CertFile != "" || cfg.KeyFile != "" {
		tlsCert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return err
		}
		config.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		}
	}
	mp, err := herminahttp.NewProxy(config, sessionHandler, logger)
	if err != nil {
		return err
	}
	http.HandleFunc("/", mp.ServeHTTP)

	errCh := make(chan error)
	switch {
	case cfg.CertFile != "" || cfg.KeyFile != "":
		go func() {
			errCh <- mp.Listen(ctx)
		}()
		logger.Info(fmt.Sprintf("%s service HTTPS server listening at %s:%s with TLS cert %s and key %s", svcName, cfg.Host, cfg.Port, cfg.CertFile, cfg.KeyFile))
	default:
		go func() {
			errCh <- mp.Listen(ctx)
		}()
		logger.Info(fmt.Sprintf("%s service HTTP server listening at %s:%s without TLS", svcName, cfg.Host, cfg.Port))
	}

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy HTTP shutdown at %s", config.Target))
		return nil
	case err := <-errCh:
		return err
	}
}
