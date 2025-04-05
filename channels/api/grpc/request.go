package grpc

import (
	"github.com/hantdev/mitras/pkg/connections"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/hantdev/mitras/pkg/policies"
)

var errDomainID = errors.New("domain id required for users")

type authorizeReq struct {
	domainID   string
	channelID  string
	clientID   string
	clientType string
	connType   connections.ConnType
}

func (req authorizeReq) validate() error {
	if req.clientType == policies.UserType && req.domainID == "" {
		return errDomainID
	}
	return nil
}

type removeClientConnectionsReq struct {
	clientID string
}

type unsetParentGroupFromChannelsReq struct {
	parentGroupID string
}

type retrieveEntityReq struct {
	Id string
}
