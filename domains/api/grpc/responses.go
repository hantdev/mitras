package grpc

type deleteUserRes struct {
	deleted bool
}

type retrieveEntityRes struct {
	id     string
	status uint8
}
