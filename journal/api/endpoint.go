package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/hantdev/mitras/internal/api"
	"github.com/hantdev/mitras/journal"
	"github.com/hantdev/mitras/pkg/apiutil"
	"github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
)

func retrieveJournalsEndpoint(svc journal.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveJournalsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}

		page, err := svc.RetrieveAll(ctx, session, req.page)
		if err != nil {
			return nil, err
		}

		return pageRes{
			JournalsPage: page,
		}, nil
	}
}
