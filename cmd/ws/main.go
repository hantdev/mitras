// Package main contains websocket-adapter main function to start the websocket-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/hantdev/hermina/pkg/session"
	"github.com/hantdev/hermina/pkg/websockets"
	grpcChannelsV1 "github.com/hantdev/mitras/api/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/api/grpc/clients/v1"
	mitraslog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/authn/authsvc"
	"github.com/hantdev/mitras/pkg/grpcclient"
	jaegerclient "github.com/hantdev/mitras/pkg/jaeger"
	"github.com/hantdev/mitras/pkg/messaging"
	"github.com/hantdev/mitras/pkg/messaging/brokers"
	brokerstracing "github.com/hantdev/mitras/pkg/messaging/brokers/tracing"
	msgevents "github.com/hantdev/mitras/pkg/messaging/events"
	"github.com/hantdev/mitras/pkg/prometheus"
	"github.com/hantdev/mitras/pkg/server"
	httpserver "github.com/hantdev/mitras/pkg/server/http"
	"github.com/hantdev/mitras/pkg/uuid"
	"github.com/hantdev/mitras/ws"
	httpapi "github.com/hantdev/mitras/ws/api"
	"github.com/hantdev/mitras/ws/tracing"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName           = "ws-adapter"
	envPrefixHTTP     = "MITRAS_WS_ADAPTER_HTTP_"
	envPrefixClients  = "MITRAS_CLIENTS_GRPC_"
	envPrefixChannels = "MITRAS_CHANNELS_GRPC_"
	envPrefixAuth     = "MITRAS_AUTH_GRPC_"
	defSvcHTTPPort    = "8190"
	targetWSPort      = "8191"
	targetWSHost      = "localhost"
)

type config struct {
	LogLevel      string  `env:"MITRAS_WS_ADAPTER_LOG_LEVEL"    envDefault:"info"`
	BrokerURL     string  `env:"MITRAS_MESSAGE_BROKER_URL"      envDefault:"nats://localhost:4222"`
	JaegerURL     url.URL `env:"MITRAS_JAEGER_URL"              envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry bool    `env:"MITRAS_SEND_TELEMETRY"          envDefault:"false"`
	InstanceID    string  `env:"MITRAS_WS_ADAPTER_INSTANCE_ID"  envDefault:""`
	TraceRatio    float64 `env:"MITRAS_JAEGER_TRACE_RATIO"      envDefault:"1.0"`
	ESURL         string  `env:"MITRAS_ES_URL"                  envDefault:"nats://localhost:4222"`
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
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	targetServerConfig := server.Config{
		Port: targetWSPort,
		Host: targetWSHost,
	}

	clientsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&clientsClientCfg, env.Options{Prefix: envPrefixClients}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
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
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	nps, err := brokers.NewPubSub(ctx, cfg.BrokerURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer nps.Close()
	nps = brokerstracing.NewPubSub(targetServerConfig, tracer, nps)

	nps, err = msgevents.NewPubSubMiddleware(ctx, nps, cfg.ESURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create event store middleware: %s", err))
		exitCode = 1
		return
	}

	svc := newService(clientsClient, channelsClient, nps, logger, tracer)

	hs := httpserver.NewServer(ctx, cancel, svcName, targetServerConfig, httpapi.MakeHandler(ctx, svc, logger, cfg.InstanceID), logger)

	g.Go(func() error {
		g.Go(func() error {
			return hs.Start()
		})
		handler := ws.NewHandler(nps, logger, authn, clientsClient, channelsClient)
		return proxyWS(ctx, httpServerConfig, targetServerConfig, logger, handler)
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("WS adapter service terminated: %s", err))
	}
}

func newService(clientsClient grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, nps messaging.PubSub, logger *slog.Logger, tracer trace.Tracer) ws.Service {
	svc := ws.New(clientsClient, channels, nps)
	svc = tracing.New(tracer, svc)
	svc = httpapi.LoggingMiddleware(svc, logger)
	counter, latency := prometheus.MakeMetrics("ws_adapter", "api")
	svc = httpapi.MetricsMiddleware(svc, counter, latency)
	return svc
}

func proxyWS(ctx context.Context, hostConfig, targetConfig server.Config, logger *slog.Logger, handler session.Handler) error {
	target := fmt.Sprintf("ws://%s:%s", targetConfig.Host, targetConfig.Port)
	address := fmt.Sprintf("%s:%s", hostConfig.Host, hostConfig.Port)
	wp, err := websockets.NewProxy(address, target, logger, handler)
	if err != nil {
		return err
	}

	errCh := make(chan error)

	go func() {
		if hostConfig.CertFile != "" && hostConfig.KeyFile != "" {
			logger.Info(fmt.Sprintf("ws-adapter service HTTP server listening at %s:%s with TLS", hostConfig.Host, hostConfig.Port))
			errCh <- wp.ListenTLS(hostConfig.CertFile, hostConfig.KeyFile)
		} else {
			logger.Info(fmt.Sprintf("ws-adapter service HTTP server listening at %s:%s without TLS", hostConfig.Host, hostConfig.Port))
			errCh <- wp.Listen()
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT WS shutdown at %s", target))
		return nil
	case err := <-errCh:
		return err
	}
}
