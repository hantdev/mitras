package grpc

type groupBasic struct {
	id          string
	domain      string
	parentGroup string
	status      uint8
}

type retrieveEntityRes groupBasic