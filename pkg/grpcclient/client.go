package grpcclient

import (
	"context"

	grpcTokenV1 "github.com/hantdev/mitras/api/grpc/token/v1"
	tokengrpc "github.com/hantdev/mitras/auth/api/grpc/token"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

// SetupTokenClient loads auth services token gRPC configuration and creates new Token services gRPC client.
//
// For example:
//
// tokenClient, tokenHandler, err := grpcclient.SetupTokenClient(ctx, grpcclient.Config{}).
func SetupTokenClient(ctx context.Context, cfg Config) (grpcTokenV1.TokenServiceClient, Handler, error) {
	client, err := NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	health := grpchealth.NewHealthClient(client.Connection())
	resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
		// Health Service name is the svcName provided during gRPC server creation `grpcserver.NewServer(ctx, cancel, svcName, grpcServerConfig, registerAuthServiceServer, logger)`
		Service: "auth",
	})
	if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		return nil, nil, ErrSvcNotServing
	}

	return tokengrpc.NewTokenClient(client.Connection(), cfg.Timeout), client, nil
}
