package api

import (
	"context"

	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/go-kit/kit/endpoint"
)

func sendMessageEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(publishReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		return publishMessageRes{}, nil
	}
}