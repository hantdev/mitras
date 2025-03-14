package auth_test

import (
	"os"
	"testing"

	"github.com/hantdev/mitras/auth/mocks"
)

var svc *mocks.Service

func TestMain(m *testing.M) {
	svc = new(mocks.Service)
	server := startGRPCServer(svc, port)

	code := m.Run()

	server.GracefulStop()

	os.Exit(code)
}