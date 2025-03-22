package grpc

type authorizeRes struct {
	authorized bool
}

type removeClientConnectionsRes struct{}

type unsetParentGroupFromChannelsRes struct{}

type channelBasic struct {
	id          string
	domain      string
	parentGroup string
	status      uint8
}

type retrieveEntityRes channelBasic