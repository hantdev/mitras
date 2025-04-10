package authsvc

import (
	"context"

	"github.com/hantdev/mitras/auth/api/grpc/auth"
	grpcAuthV1 "github.com/hantdev/mitras/internal/grpc/auth/v1"
	"github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/hantdev/mitras/pkg/grpcclient"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

type authentication struct {
	authSvcClient grpcAuthV1.AuthServiceClient
}

var _ authn.Authentication = (*authentication)(nil)

func NewAuthentication(ctx context.Context, cfg grpcclient.Config) (authn.Authentication, grpcclient.Handler, error) {
	client, err := grpcclient.NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	health := grpchealth.NewHealthClient(client.Connection())
	resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
		Service: "auth",
	})
	if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		return nil, nil, grpcclient.ErrSvcNotServing
	}
	authSvcClient := auth.NewAuthClient(client.Connection(), cfg.Timeout)
	return authentication{authSvcClient}, client, nil
}

func (a authentication) Authenticate(ctx context.Context, token string) (authn.Session, error) {
	res, err := a.authSvcClient.Authenticate(ctx, &grpcAuthV1.AuthNReq{Token: token})
	if err != nil {
		return authn.Session{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	return authn.Session{DomainUserID: res.GetId(), UserID: res.GetUserId(), DomainID: res.GetDomainId()}, nil
}
