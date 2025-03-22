package grpc

import "github.com/hantdev/mitras/pkg/connections"

type entity struct {
	id          string
	domain      string
	parentGroup string
	status      uint8
}

type authenticateRes struct {
	id            string
	authenticated bool
}

type retrieveEntitiesRes struct {
	total   uint64
	limit   uint64
	offset  uint64
	clients []entity
}

type retrieveEntityRes entity

type connectionsReq struct {
	connections []connection
}

type connection struct {
	clientID  string
	channelID string
	domainID  string
	connType  connections.ConnType
}
type connectionsRes struct {
	ok bool
}

type removeChannelConnectionsRes struct{}

type UnsetParentGroupFromClientRes struct{}