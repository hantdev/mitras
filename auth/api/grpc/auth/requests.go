package auth

import (
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/auth"
)

type authenticateReq struct {
	token string
}

func (req authenticateReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	return nil
}

// authReq represents authorization request. It contains:
// 1. subject - an action invoker
// 2. object - an entity over which action will be executed
// 3. action - type of action that will be executed (read/write).
type authReq struct {
	Domain      string
	SubjectType string
	SubjectKind string
	Subject     string
	Relation    string
	Permission  string
	ObjectType  string
	Object      string
}

func (req authReq) validate() error {
	if req.Subject == "" || req.SubjectType == "" {
		return apiutil.ErrMissingPolicySub
	}

	if req.Object == "" || req.ObjectType == "" {
		return apiutil.ErrMissingPolicyObj
	}

	if req.Permission == "" {
		return apiutil.ErrMalformedPolicyPer
	}

	return nil
}

type authPATReq struct {
	userID           string
	patID            string
	entityType       auth.EntityType
	optionalDomainID string
	operation        auth.Operation
	entityID         string
}

func (req authPATReq) validate() error {
	if req.userID == "" {
		return apiutil.ErrMissingUserID
	}
	if req.patID == "" {
		return apiutil.ErrMissingPATID
	}
	return nil
}
