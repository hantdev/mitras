package grpcclient

import (
	"context"

	grpcCommonV1 "github.com/hantdev/mitras/api/grpc/common/v1"
	grpcDomainsV1 "github.com/hantdev/mitras/api/grpc/domains/v1"
	"github.com/hantdev/mitras/domains"
	pkgDomains "github.com/hantdev/mitras/pkg/domains"
	"github.com/hantdev/mitras/pkg/grpcclient"
)

type authorization struct {
	domainsSvcClient grpcDomainsV1.DomainsServiceClient
}

var _ pkgDomains.Authorization = (*authorization)(nil)

func NewAuthorization(ctx context.Context, cfg grpcclient.Config) (pkgDomains.Authorization, grpcDomainsV1.DomainsServiceClient, grpcclient.Handler, error) {
	domainsClient, domainsHandler, err := grpcclient.SetupDomainsClient(ctx, cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	return authorization{domainsSvcClient: domainsClient}, domainsClient, domainsHandler, nil
}

func (a authorization) RetrieveEntity(ctx context.Context, id string) (domains.Domain, error) {
	req := grpcCommonV1.RetrieveEntityReq{
		Id: id,
	}
	res, err := a.domainsSvcClient.RetrieveEntity(ctx, &req)
	if err != nil {
		return domains.Domain{}, err
	}

	return domains.Domain{
		ID:     res.Entity.GetId(),
		Status: domains.Status(res.Entity.GetStatus()),
	}, nil
}
