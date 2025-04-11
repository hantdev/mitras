package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/hantdev/mitras/domains"
)

func deleteUserFromDomainsEndpoint(svc domains.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteUserPoliciesReq)
		if err := req.validate(); err != nil {
			return deleteUserRes{}, err
		}

		if err := svc.DeleteUserFromDomains(ctx, req.ID); err != nil {
			return deleteUserRes{}, err
		}

		return deleteUserRes{deleted: true}, nil
	}
}
