package sdk_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/bootstrap/api"
	bmocks "github.com/hantdev/mitras/bootstrap/mocks"
	chmocks "github.com/hantdev/mitras/channels/mocks"
	climocks "github.com/hantdev/mitras/clients/mocks"
	smqlog "github.com/hantdev/mitras/logger"
	authnmocks "github.com/hantdev/mitras/pkg/authn/mocks"
	"github.com/hantdev/mitras/pkg/errors"
	sdk "github.com/hantdev/mitras/pkg/sdk"
	readersapi "github.com/hantdev/mitras/readers/api"
	readersmocks "github.com/hantdev/mitras/readers/mocks"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	clientsTs, _, _ := setupClients()
	defer clientsTs.Close()

	usersTs, _, _ := setupUsers()
	defer usersTs.Close()

	certsTs, _, _ := setupCerts()
	defer certsTs.Close()

	bootstrapTs := setupMinimalBootstrap()
	defer bootstrapTs.Close()

	readerTs := setupMinimalReader()
	defer readerTs.Close()

	httpAdapterTs, _ := setupMessages()
	defer httpAdapterTs.Close()

	sdkConf := sdk.Config{
		ClientsURL:      clientsTs.URL,
		UsersURL:        usersTs.URL,
		CertsURL:        certsTs.URL,
		BootstrapURL:    bootstrapTs.URL,
		ReaderURL:       readerTs.URL,
		HTTPAdapterURL:  httpAdapterTs.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mgsdk := sdk.NewSDK(sdkConf)
	cases := []struct {
		desc        string
		service     string
		empty       bool
		description string
		status      string
		err         errors.SDKError
	}{
		{
			desc:        "get clients service health check",
			service:     "clients",
			empty:       false,
			err:         nil,
			description: "clients service",
			status:      "pass",
		},
		{
			desc:        "get users service health check",
			service:     "users",
			empty:       false,
			err:         nil,
			description: "users service",
			status:      "pass",
		},
		{
			desc:        "get certs service health check",
			service:     "certs",
			empty:       false,
			err:         nil,
			description: "certs service",
			status:      "pass",
		},
		{
			desc:        "get bootstrap service health check",
			service:     "bootstrap",
			empty:       false,
			err:         nil,
			description: "bootstrap service",
			status:      "pass",
		},
		{
			desc:        "get reader service health check",
			service:     "reader",
			empty:       false,
			err:         nil,
			description: "test service",
			status:      "pass",
		},
		{
			desc:        "get http-adapter service health check",
			service:     "http-adapter",
			empty:       false,
			err:         nil,
			description: "http service",
			status:      "pass",
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			h, err := mgsdk.Health(tc.service)
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, h.Status, fmt.Sprintf("%s: expected %s status, got %s", tc.desc, tc.status, h.Status))
			assert.Equal(t, tc.empty, h.Version == "", fmt.Sprintf("%s: expected non-empty version", tc.desc))
			assert.Equal(t, mitras.Commit, h.Commit, fmt.Sprintf("%s: expected non-empty commit", tc.desc))
			assert.Equal(t, tc.description, h.Description, fmt.Sprintf("%s: expected proper description, got %s", tc.desc, h.Description))
			assert.Equal(t, mitras.BuildTime, h.BuildTime, fmt.Sprintf("%s: expected default epoch date, got %s", tc.desc, h.BuildTime))
		})
	}
}

func setupMinimalBootstrap() *httptest.Server {
	bsvc := new(bmocks.Service)
	reader := new(bmocks.ConfigReader)
	logger := smqlog.NewMock()
	authn := new(authnmocks.Authentication)
	mux := api.MakeHandler(bsvc, authn, reader, logger, "")

	return httptest.NewServer(mux)
}

func setupMinimalReader() *httptest.Server {
	repo := new(readersmocks.MessageRepository)
	channels := new(chmocks.ChannelsServiceClient)
	authn := new(authnmocks.Authentication)
	clients := new(climocks.ClientsServiceClient)

	mux := readersapi.MakeHandler(repo, authn, clients, channels, "test", "")
	return httptest.NewServer(mux)
}
