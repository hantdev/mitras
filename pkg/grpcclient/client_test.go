package grpcclient_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	grpcClientsV1 "github.com/hantdev/mitras/api/grpc/clients/v1"
	grpcDomainsV1 "github.com/hantdev/mitras/api/grpc/domains/v1"
	grpcTokenV1 "github.com/hantdev/mitras/api/grpc/token/v1"
	tokengrpcapi "github.com/hantdev/mitras/auth/api/grpc/token"
	"github.com/hantdev/mitras/auth/mocks"
	clientsgrpcapi "github.com/hantdev/mitras/clients/api/grpc"
	climocks "github.com/hantdev/mitras/clients/private/mocks"
	domainsgrpcapi "github.com/hantdev/mitras/domains/api/grpc"
	domainsMocks "github.com/hantdev/mitras/domains/private/mocks"
	mitraslog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/hantdev/mitras/pkg/grpcclient"
	"github.com/hantdev/mitras/pkg/server"
	grpcserver "github.com/hantdev/mitras/pkg/server/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestSetupToken(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerAuthServiceServer := func(srv *grpc.Server) {
		grpcTokenV1.RegisterTokenServiceServer(srv, tokengrpcapi.NewTokenServer(new(mocks.Service)))
	}
	gs := grpcserver.NewServer(ctx, cancel, "auth", server.Config{Port: "12345"}, registerAuthServiceServer, mitraslog.NewMock())
	go func() {
		err := gs.Start()
		assert.Nil(t, err, fmt.Sprintf(`"Unexpected error creating server %s"`, err))
	}()
	defer func() {
		err := gs.Stop()
		assert.Nil(t, err, fmt.Sprintf(`"Unexpected error stopping server %s"`, err))
	}()

	cases := []struct {
		desc   string
		config grpcclient.Config
		err    error
	}{
		{
			desc: "successful",
			config: grpcclient.Config{
				URL:     "localhost:12345",
				Timeout: time.Second,
			},
			err: nil,
		},
		{
			desc: "failed with empty URL",
			config: grpcclient.Config{
				URL:     "",
				Timeout: time.Second,
			},
			err: errors.New("service is not serving"),
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			client, handler, err := grpcclient.SetupTokenClient(context.Background(), c.config)
			assert.True(t, errors.Contains(err, c.err), fmt.Sprintf("expected %s to contain %s", err, c.err))
			if err == nil {
				assert.NotNil(t, client)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestSetupClientsClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registerClientsServiceServer := func(srv *grpc.Server) {
		grpcClientsV1.RegisterClientsServiceServer(srv, clientsgrpcapi.NewServer(new(climocks.Service)))
	}
	gs := grpcserver.NewServer(ctx, cancel, "clients", server.Config{Port: "12345"}, registerClientsServiceServer, mitraslog.NewMock())
	go func() {
		err := gs.Start()
		assert.Nil(t, err, fmt.Sprintf(`"Unexpected error creating server %s"`, err))
	}()
	time.Sleep(time.Second)
	defer func() {
		err := gs.Stop()
		assert.Nil(t, err, fmt.Sprintf(`"Unexpected error stopping server %s"`, err))
	}()

	cases := []struct {
		desc   string
		config grpcclient.Config
		err    error
	}{
		{
			desc: "successful",
			config: grpcclient.Config{
				URL:     "localhost:12345",
				Timeout: time.Second,
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			client, handler, err := grpcclient.SetupClientsClient(context.Background(), c.config)
			assert.True(t, errors.Contains(err, c.err), fmt.Sprintf("expected %s to contain %s", err, c.err))
			if err == nil {
				assert.NotNil(t, client)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestSetupDomainsClient(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	registerDomainsServiceServer := func(srv *grpc.Server) {
		grpcDomainsV1.RegisterDomainsServiceServer(srv, domainsgrpcapi.NewDomainsServer(new(domainsMocks.Service)))
	}
	gs := grpcserver.NewServer(ctx, cancel, "domains", server.Config{Port: "12345"}, registerDomainsServiceServer, mitraslog.NewMock())
	go func() {
		err := gs.Start()
		assert.Nil(t, err, fmt.Sprintf("Unexpected error creating server %s", err))
	}()
	time.Sleep(time.Second)
	defer func() {
		err := gs.Stop()
		assert.Nil(t, err, fmt.Sprintf("Unexpected error stopping server %s", err))
	}()

	cases := []struct {
		desc   string
		config grpcclient.Config
		err    error
	}{
		{
			desc: "successfully",
			config: grpcclient.Config{
				URL:     "localhost:12345",
				Timeout: time.Second,
			},
			err: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			client, handler, err := grpcclient.SetupDomainsClient(context.Background(), c.config)
			assert.True(t, errors.Contains(err, c.err), fmt.Sprintf("expected %s to contain %s", err, c.err))
			if err == nil {
				assert.NotNil(t, client)
				assert.NotNil(t, handler)
			}
		})
	}
}
