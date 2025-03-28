package token

import (
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/auth"
)

type issueReq struct {
	userID  string
	keyType auth.KeyType
}

func (req issueReq) validate() error {
	if req.keyType != auth.AccessKey &&
		req.keyType != auth.APIKey &&
		req.keyType != auth.RecoveryKey &&
		req.keyType != auth.InvitationKey {
		return apiutil.ErrInvalidAuthKey
	}

	return nil
}

type refreshReq struct {
	refreshToken string
}

func (req refreshReq) validate() error {
	if req.refreshToken == "" {
		return apiutil.ErrMissingSecret
	}

	return nil
}
