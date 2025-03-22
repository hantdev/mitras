package grpc_test

import (
	"os"
	"testing"

	"github.com/hantdev/mitras/domains/private/mocks"
)

var svc *mocks.Service

func TestMain(m *testing.M) {
	svc = new(mocks.Service)
	server := startGRPCServer(svc, port)

	code := m.Run()

	server.GracefulStop()

	os.Exit(code)
}
