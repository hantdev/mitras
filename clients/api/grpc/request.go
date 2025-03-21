package grpc

type authenticateReq struct {
	ClientID     string
	ClientSecret string
}

type retrieveEntitiesReq struct {
	Ids []string
}

type retrieveEntityReq struct {
	Id string
}

type removeChannelConnectionsReq struct {
	channelID string
}

type UnsetParentGroupFromClientReq struct {
	parentGroupID string
}