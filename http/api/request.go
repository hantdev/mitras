package api

import (
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/pkg/messaging"
)

type publishReq struct {
	msg   *messaging.Message
	token string
}

func (req publishReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerKey
	}
	if len(req.msg.Payload) == 0 {
		return apiutil.ErrEmptyMessage
	}

	return nil
}