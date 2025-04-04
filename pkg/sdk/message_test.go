package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hantdev/hermina"
	proxy "github.com/hantdev/hermina/pkg/http"
	grpcChannelsV1 "github.com/hantdev/mitras/api/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/api/grpc/clients/v1"
	apiutil "github.com/hantdev/mitras/api/http/util"
	chmocks "github.com/hantdev/mitras/channels/mocks"
	climocks "github.com/hantdev/mitras/clients/mocks"
	adapter "github.com/hantdev/mitras/http"
	"github.com/hantdev/mitras/http/api"
	smqlog "github.com/hantdev/mitras/logger"
	authnmocks "github.com/hantdev/mitras/pkg/authn/mocks"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	pubsub "github.com/hantdev/mitras/pkg/messaging/mocks"
	sdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	channelsGRPCClient *chmocks.ChannelsServiceClient
	clientsGRPCClient  *climocks.ClientsServiceClient
)

func setupMessages() (*httptest.Server, *pubsub.PubSub) {
	clientsGRPCClient = new(climocks.ClientsServiceClient)
	channelsGRPCClient = new(chmocks.ChannelsServiceClient)
	pub := new(pubsub.PubSub)
	authn := new(authnmocks.Authentication)
	handler := adapter.NewHandler(pub, authn, clientsGRPCClient, channelsGRPCClient, smqlog.NewMock())

	mux := api.MakeHandler(smqlog.NewMock(), "")
	target := httptest.NewServer(mux)

	config := hermina.Config{
		Address: "",
		Target:  target.URL,
	}
	mp, err := proxy.NewProxy(config, handler, smqlog.NewMock())
	if err != nil {
		return nil, nil
	}

	return httptest.NewServer(http.HandlerFunc(mp.ServeHTTP)), pub
}

func TestSendMessage(t *testing.T) {
	ts, pub := setupMessages()
	defer ts.Close()

	msg := `[{"n":"current","t":-1,"v":1.6}]`
	clientKey := "clientKey"
	channelID := "channelID"

	sdkConf := sdk.Config{
		HTTPAdapterURL:  ts.URL,
		MsgContentType:  "application/senml+json",
		TLSVerification: false,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc      string
		chanName  string
		msg       string
		clientKey string
		authRes   *grpcClientsV1.AuthnRes
		authErr   error
		svcErr    error
		err       errors.SDKError
	}{
		{
			desc:      "publish message successfully",
			chanName:  channelID,
			msg:       msg,
			clientKey: clientKey,
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: true, Id: ""},
			authErr:   nil,
			svcErr:    nil,
			err:       nil,
		},
		{
			desc:      "publish message with empty client key",
			chanName:  channelID,
			msg:       msg,
			clientKey: "",
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: false, Id: ""},
			authErr:   svcerr.ErrAuthentication,
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "publish message with invalid client key",
			chanName:  channelID,
			msg:       msg,
			clientKey: "invalid",
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: false, Id: ""},
			authErr:   svcerr.ErrAuthentication,
			svcErr:    svcerr.ErrAuthentication,
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "publish message with invalid channel ID",
			chanName:  wrongID,
			msg:       msg,
			clientKey: clientKey,
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: false, Id: ""},
			authErr:   svcerr.ErrAuthentication,
			svcErr:    svcerr.ErrAuthentication,
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "publish message with empty message body",
			chanName:  channelID,
			msg:       "",
			clientKey: clientKey,
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: true, Id: ""},
			authErr:   nil,
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrEmptyMessage), http.StatusBadRequest),
		},
		{
			desc:      "publish message with channel subtopic",
			chanName:  channelID + ".subtopic",
			msg:       msg,
			clientKey: clientKey,
			authRes:   &grpcClientsV1.AuthnRes{Authenticated: true, Id: ""},
			authErr:   nil,
			svcErr:    nil,
			err:       nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authzCall := clientsGRPCClient.On("Authenticate", mock.Anything, mock.Anything).Return(tc.authRes, tc.authErr)
			authnCall := channelsGRPCClient.On("Authorize", mock.Anything, mock.Anything).Return(&grpcChannelsV1.AuthzRes{Authorized: true}, nil)
			svcCall := pub.On("Publish", mock.Anything, channelID, mock.Anything).Return(tc.svcErr)
			err := mgsdk.SendMessage(context.Background(), tc.chanName, tc.msg, tc.clientKey)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Publish", mock.Anything, channelID, mock.Anything)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authzCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestSetContentType(t *testing.T) {
	ts, _ := setupMessages()
	defer ts.Close()

	sdkConf := sdk.Config{
		HTTPAdapterURL:  ts.URL,
		MsgContentType:  "application/senml+json",
		TLSVerification: false,
	}
	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc  string
		cType sdk.ContentType
		err   errors.SDKError
	}{
		{
			desc:  "set senml+json content type",
			cType: "application/senml+json",
			err:   nil,
		},
		{
			desc:  "set invalid content type",
			cType: "invalid",
			err:   errors.NewSDKError(apiutil.ErrUnsupportedContentType),
		},
	}
	for _, tc := range cases {
		err := mgsdk.SetContentType(tc.cType)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}
