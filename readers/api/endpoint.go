package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	grpcChannelsV1 "github.com/hantdev/mitras/internal/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/internal/grpc/clients/v1"
	"github.com/hantdev/mitras/pkg/apiutil"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	"github.com/hantdev/mitras/readers"
)

func listMessagesEndpoint(svc readers.MessageRepository, authn smqauthn.Authentication, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMessagesReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		if err := authnAuthz(ctx, req, authn, clients, channels); err != nil {
			return nil, errors.Wrap(svcerr.ErrAuthorization, err)
		}

		page, err := svc.ReadAll(req.chanID, req.pageMeta)
		if err != nil {
			return nil, err
		}

		return pageRes{
			PageMetadata: page.PageMetadata,
			Total:        page.Total,
			Messages:     page.Messages,
		}, nil
	}
}
