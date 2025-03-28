package keys

import (
	"time"

	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/auth"
)

type issueKeyReq struct {
	token    string
	Type     auth.KeyType  `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.Type != auth.AccessKey &&
		req.Type != auth.RecoveryKey &&
		req.Type != auth.APIKey {
		return apiutil.ErrInvalidAPIKey
	}

	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}
