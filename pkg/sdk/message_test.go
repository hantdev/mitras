package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hantdev/hermina"
	proxy "github.com/hantdev/hermina/pkg/http"
	chmocks "github.com/hantdev/mitras/channels/mocks"
	climocks "github.com/hantdev/mitras/clients/mocks"
	adapter "github.com/hantdev/mitras/http"
	"github.com/hantdev/mitras/http/api"
	grpcChannelsV1 "github.com/hantdev/mitras/internal/grpc/channels/v1"
	grpcClientsV1 "github.com/hantdev/mitras/internal/grpc/clients/v1"
	smqlog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/apiutil"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	authnmocks "github.com/hantdev/mitras/pkg/authn/mocks"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	pubsub "github.com/hantdev/mitras/pkg/messaging/mocks"
	sdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/hantdev/mitras/pkg/transformers/senml"
	"github.com/hantdev/mitras/readers"
	readersapi "github.com/hantdev/mitras/readers/api"
	readersmocks "github.com/hantdev/mitras/readers/mocks"
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

func setupReaders() (*httptest.Server, *authnmocks.Authentication, *readersmocks.MessageRepository) {
	repo := new(readersmocks.MessageRepository)
	authn := new(authnmocks.Authentication)
	clientsGRPCClient = new(climocks.ClientsServiceClient)
	channelsGRPCClient = new(chmocks.ChannelsServiceClient)

	mux := readersapi.MakeHandler(repo, authn, clientsGRPCClient, channelsGRPCClient, "test", "")
	return httptest.NewServer(mux), authn, repo
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
			err := mgsdk.SendMessage(tc.chanName, tc.msg, tc.clientKey)
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

func TestReadMessages(t *testing.T) {
	ts, authn, repo := setupReaders()
	defer ts.Close()

	channelID := "channelID"
	msgValue := 1.6
	boolVal := true
	msg := senml.Message{
		Name:      "current",
		Time:      1720000000,
		Value:     &msgValue,
		Publisher: validID,
	}
	invalidMsg := "[{\"n\":\"current\",\"t\":-1,\"v\":1.6}]"

	sdkConf := sdk.Config{
		ReaderURL: ts.URL,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc            string
		token           string
		chanName        string
		domainID        string
		messagePageMeta sdk.MessagePageMetadata
		authzErr        error
		authnErr        error
		repoRes         readers.MessagesPage
		repoErr         error
		response        sdk.MessagesPage
		err             errors.SDKError
	}{
		{
			desc:     "read messages successfully",
			token:    validToken,
			chanName: channelID,
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
					Level:  0,
				},
				Publisher: validID,
				BoolValue: &boolVal,
			},
			repoRes: readers.MessagesPage{
				Total:    1,
				Messages: []readers.Message{msg},
			},
			repoErr: nil,
			response: sdk.MessagesPage{
				PageRes: sdk.PageRes{
					Total: 1,
				},
				Messages: []senml.Message{msg},
			},
			err: nil,
		},
		{
			desc:     "read messages successfully with subtopic",
			token:    validToken,
			chanName: channelID + ".subtopic",
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
				},
				Publisher: validID,
			},
			repoRes: readers.MessagesPage{
				Total:    1,
				Messages: []readers.Message{msg},
			},
			repoErr: nil,
			response: sdk.MessagesPage{
				PageRes: sdk.PageRes{
					Total: 1,
				},
				Messages: []senml.Message{msg},
			},
			err: nil,
		},
		{
			desc:     "read messages with invalid token",
			token:    invalidToken,
			chanName: channelID,
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
				},
				Subtopic:  "subtopic",
				Publisher: validID,
			},
			authzErr: svcerr.ErrAuthorization,
			repoRes:  readers.MessagesPage{},
			response: sdk.MessagesPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrAuthorization, svcerr.ErrAuthorization), http.StatusUnauthorized),
		},
		{
			desc:     "read messages with empty token",
			token:    "",
			chanName: channelID,
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
				},
				Subtopic:  "subtopic",
				Publisher: validID,
			},
			authnErr: svcerr.ErrAuthentication,
			repoRes:  readers.MessagesPage{},
			response: sdk.MessagesPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerToken), http.StatusUnauthorized),
		},
		{
			desc:     "read messages with empty channel ID",
			token:    validToken,
			chanName: "",
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
				},
				Subtopic:  "subtopic",
				Publisher: validID,
			},
			repoRes:  readers.MessagesPage{},
			repoErr:  nil,
			response: sdk.MessagesPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "read messages with invalid message page metadata",
			token:    validToken,
			chanName: channelID,
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
					Metadata: map[string]interface{}{
						"key": make(chan int),
					},
				},
				Subtopic:  "subtopic",
				Publisher: validID,
			},
			repoRes:  readers.MessagesPage{},
			repoErr:  nil,
			response: sdk.MessagesPage{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:     "read messages with response that cannot be unmarshalled",
			token:    validToken,
			chanName: channelID,
			domainID: validID,
			messagePageMeta: sdk.MessagePageMetadata{
				PageMetadata: sdk.PageMetadata{
					Offset: 0,
					Limit:  10,
				},
				Subtopic:  "subtopic",
				Publisher: validID,
			},
			repoRes: readers.MessagesPage{
				Total:    1,
				Messages: []readers.Message{invalidMsg},
			},
			repoErr:  nil,
			response: sdk.MessagesPage{},
			err:      errors.NewSDKError(errors.New("json: cannot unmarshal string into Go struct field MessagesPage.messages of type senml.Message")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authCall1 := authn.On("Authenticate", mock.Anything, tc.token).Return(smqauthn.Session{UserID: validID}, tc.authnErr)
			authzCall := channelsGRPCClient.On("Authorize", mock.Anything, mock.Anything).Return(&grpcChannelsV1.AuthzRes{Authorized: true}, tc.authzErr)
			repoCall := repo.On("ReadAll", channelID, mock.Anything).Return(tc.repoRes, tc.repoErr)
			response, err := mgsdk.ReadMessages(tc.messagePageMeta, tc.chanName, tc.domainID, tc.token)
			fmt.Println(err)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, response)
			if tc.err == nil {
				ok := repoCall.Parent.AssertCalled(t, "ReadAll", channelID, mock.Anything)
				assert.True(t, ok)
			}
			authCall1.Unset()
			authzCall.Unset()
			repoCall.Unset()
		})
	}
}
