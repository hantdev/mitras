package grpc

import (
	"github.com/hantdev/mitras/pkg/apiutil"
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
