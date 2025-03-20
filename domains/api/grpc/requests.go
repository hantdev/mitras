package grpc

import (
	apiutil "github.com/hantdev/mitras/api/http/util"
)

type deleteUserPoliciesReq struct {
	ID string
}

func (req deleteUserPoliciesReq) validate() error {
	if req.ID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type retrieveEntityReq struct {
	ID string
}

func (req retrieveEntityReq) validate() error {
	if req.ID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}