package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hantdev/mitras/channels"
	chapi "github.com/hantdev/mitras/channels/api/http"
	chmocks "github.com/hantdev/mitras/channels/mocks"
	"github.com/hantdev/mitras/clients"
	"github.com/hantdev/mitras/internal/testsutil"
	smqlog "github.com/hantdev/mitras/logger"
	"github.com/hantdev/mitras/pkg/apiutil"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	authnmocks "github.com/hantdev/mitras/pkg/authn/mocks"
	"github.com/hantdev/mitras/pkg/connections"
	"github.com/hantdev/mitras/pkg/errors"
	svcerr "github.com/hantdev/mitras/pkg/errors/service"
	sdk "github.com/hantdev/mitras/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	channelName = "channelName"
	newName     = "newName"
	channel     = generateTestChannel(&testing.T{})
)

func setupChannels() (*httptest.Server, *chmocks.Service, *authnmocks.Authentication) {
	svc := new(chmocks.Service)
	logger := smqlog.NewMock()
	authn := new(authnmocks.Authentication)
	mux := chi.NewRouter()
	chapi.MakeHandler(svc, authn, mux, logger, "")

	return httptest.NewServer(mux), svc, authn
}

func TestCreateChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	createChannelReq := channels.Channel{
		Name:     channel.Name,
		Metadata: clients.Metadata{"role": "client"},
		Status:   clients.EnabledStatus,
	}

	channelReq := sdk.Channel{
		Name:     channel.Name,
		Metadata: validMetadata,
		Status:   clients.EnabledStatus.String(),
	}

	parentID := testsutil.GenerateUUID(&testing.T{})
	pChannel := channel
	pChannel.ParentGroup = parentID

	iChannel := convertChannel(channel)
	iChannel.Metadata = clients.Metadata{
		"test": make(chan int),
	}

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc             string
		channelReq       sdk.Channel
		domainID         string
		token            string
		session          smqauthn.Session
		createChannelReq channels.Channel
		svcRes           []channels.Channel
		svcErr           error
		authenticateRes  smqauthn.Session
		authenticateErr  error
		response         sdk.Channel
		err              errors.SDKError
	}{
		{
			desc:             "create channel successfully",
			channelReq:       channelReq,
			domainID:         domainID,
			token:            validToken,
			createChannelReq: createChannelReq,
			svcRes:           []channels.Channel{convertChannel(channel)},
			svcErr:           nil,
			response:         channel,
			err:              nil,
		},
		{
			desc:             "create channel with existing name",
			channelReq:       channelReq,
			domainID:         domainID,
			token:            validToken,
			createChannelReq: createChannelReq,
			svcRes:           []channels.Channel{},
			svcErr:           svcerr.ErrCreateEntity,
			response:         sdk.Channel{},
			err:              errors.NewSDKErrorWithStatus(svcerr.ErrCreateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc: "create channel that can't be marshalled",
			channelReq: sdk.Channel{
				Name: "test",
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			domainID:         domainID,
			token:            validToken,
			createChannelReq: channels.Channel{},
			svcRes:           []channels.Channel{},
			svcErr:           nil,
			response:         sdk.Channel{},
			err:              errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc: "create channel with parent group",
			channelReq: sdk.Channel{
				Name:        channel.Name,
				ParentGroup: parentID,
				Status:      clients.EnabledStatus.String(),
			},
			domainID: domainID,
			token:    validToken,
			createChannelReq: channels.Channel{
				Name:        channel.Name,
				ParentGroup: parentID,
				Status:      clients.EnabledStatus,
			},
			svcRes:   []channels.Channel{convertChannel(pChannel)},
			svcErr:   nil,
			response: pChannel,
			err:      nil,
		},
		{
			desc: "create channel with invalid parent",
			channelReq: sdk.Channel{
				Name:        channel.Name,
				ParentGroup: wrongID,
				Status:      clients.EnabledStatus.String(),
			},
			domainID: domainID,
			token:    validToken,
			createChannelReq: channels.Channel{
				Name:        channel.Name,
				ParentGroup: wrongID,
				Status:      clients.EnabledStatus,
			},
			svcRes:   []channels.Channel{},
			svcErr:   svcerr.ErrCreateEntity,
			response: sdk.Channel{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrCreateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc: "create a channel with every field defined",
			channelReq: sdk.Channel{
				ID:          channel.ID,
				ParentGroup: parentID,
				Name:        channel.Name,
				Metadata:    validMetadata,
				CreatedAt:   channel.CreatedAt,
				UpdatedAt:   channel.UpdatedAt,
				Status:      clients.EnabledStatus.String(),
			},
			domainID: domainID,
			token:    validToken,
			createChannelReq: channels.Channel{
				ID:          channel.ID,
				ParentGroup: parentID,
				Name:        channel.Name,
				Metadata:    clients.Metadata{"role": "client"},
				CreatedAt:   channel.CreatedAt,
				UpdatedAt:   channel.UpdatedAt,
				Status:      clients.EnabledStatus,
			},
			svcRes:   []channels.Channel{convertChannel(pChannel)},
			svcErr:   nil,
			response: pChannel,
			err:      nil,
		},
		{
			desc:             "create channel with response that can't be unmarshalled",
			channelReq:       channelReq,
			domainID:         domainID,
			token:            validToken,
			createChannelReq: createChannelReq,
			svcRes:           []channels.Channel{iChannel},
			svcErr:           nil,
			response:         sdk.Channel{},
			err:              errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("CreateChannels", mock.Anything, tc.session, tc.createChannelReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.CreateChannel(tc.channelReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "CreateChannels", mock.Anything, tc.session, tc.createChannelReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListChannels(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	var chs []sdk.Channel
	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	for i := 10; i < 100; i++ {
		gr := sdk.Channel{
			ID:       generateUUID(t),
			Name:     fmt.Sprintf("channel_%d", i),
			Metadata: sdk.Metadata{"name": fmt.Sprintf("client_%d", i)},
		}
		chs = append(chs, gr)
	}

	cases := []struct {
		desc             string
		domainID         string
		token            string
		session          smqauthn.Session
		status           clients.Status
		total            uint64
		offset           uint64
		limit            uint64
		level            int
		name             string
		metadata         sdk.Metadata
		channelsPageMeta channels.PageMetadata
		svcRes           channels.Page
		svcErr           error
		authenticateRes  smqauthn.Session
		authenticateErr  error
		response         sdk.ChannelsPage
		err              errors.SDKError
	}{
		{
			desc:     "list channels successfully",
			token:    validToken,
			domainID: domainID,
			limit:    limit,
			offset:   offset,
			total:    total,
			channelsPageMeta: channels.PageMetadata{
				Offset:     offset,
				Limit:      limit,
				Permission: defPermission,
			},
			svcRes: channels.Page{
				PageMetadata: channels.PageMetadata{
					Total: uint64(len(chs[offset:limit])),
				},
				Channels: convertChannels(chs[offset:limit]),
			},
			response: sdk.ChannelsPage{
				PageRes: sdk.PageRes{
					Total: uint64(len(chs[offset:limit])),
				},
				Channels: chs[offset:limit],
			},
			err: nil,
		},
		{
			desc:     "list channels with invalid token",
			token:    invalidToken,
			domainID: domainID,
			offset:   offset,
			limit:    limit,
			channelsPageMeta: channels.PageMetadata{
				Offset: offset,
				Limit:  limit,
			},
			svcRes:          channels.Page{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.ChannelsPage{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:             "list channels with empty token",
			token:            "",
			domainID:         validID,
			offset:           offset,
			limit:            limit,
			channelsPageMeta: channels.PageMetadata{},
			svcRes:           channels.Page{},
			svcErr:           nil,
			response:         sdk.ChannelsPage{},
			err:              errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "list channels with zero limit",
			token:    validToken,
			domainID: domainID,
			offset:   offset,
			limit:    0,
			channelsPageMeta: channels.PageMetadata{
				Offset:     offset,
				Limit:      10,
				Permission: defPermission,
			},
			svcRes: channels.Page{
				PageMetadata: channels.PageMetadata{
					Total: uint64(len(chs[offset:])),
				},
				Channels: convertChannels(chs[offset:limit]),
			},
			svcErr: nil,
			response: sdk.ChannelsPage{
				PageRes: sdk.PageRes{
					Total: uint64(len(chs[offset:])),
				},
				Channels: chs[offset:limit],
			},
			err: nil,
		},
		{
			desc:             "list channels with limit greater than max",
			token:            validToken,
			domainID:         domainID,
			offset:           offset,
			limit:            110,
			channelsPageMeta: channels.PageMetadata{},
			svcRes:           channels.Page{},
			svcErr:           nil,
			response:         sdk.ChannelsPage{},
			err:              errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrLimitSize), http.StatusBadRequest),
		},
		{
			desc:     "list channels with level",
			token:    validToken,
			domainID: domainID,
			offset:   0,
			limit:    1,
			level:    1,
			channelsPageMeta: channels.PageMetadata{
				Offset:     offset,
				Limit:      1,
				Permission: defPermission,
			},
			svcRes: channels.Page{
				PageMetadata: channels.PageMetadata{
					Total: 1,
				},
				Channels: convertChannels(chs[0:1]),
			},
			svcErr: nil,
			response: sdk.ChannelsPage{
				PageRes: sdk.PageRes{
					Total: 1,
				},
				Channels: chs[0:1],
			},
			err: nil,
		},
		{
			desc:     "list channels with metadata",
			token:    validToken,
			domainID: domainID,
			offset:   0,
			limit:    10,
			metadata: sdk.Metadata{"name": "client_89"},
			channelsPageMeta: channels.PageMetadata{
				Offset:     offset,
				Limit:      10,
				Permission: defPermission,
				Metadata:   clients.Metadata{"name": "client_89"},
			},
			svcRes: channels.Page{
				PageMetadata: channels.PageMetadata{
					Total: 1,
				},
				Channels: convertChannels([]sdk.Channel{chs[89]}),
			},
			svcErr: nil,
			response: sdk.ChannelsPage{
				PageRes: sdk.PageRes{
					Total: 1,
				},
				Channels: []sdk.Channel{chs[89]},
			},
			err: nil,
		},
		{
			desc:     "list channels with invalid metadata",
			token:    validToken,
			domainID: domainID,
			offset:   0,
			limit:    10,
			metadata: sdk.Metadata{
				"test": make(chan int),
			},
			channelsPageMeta: channels.PageMetadata{},
			svcRes:           channels.Page{},
			svcErr:           nil,
			response:         sdk.ChannelsPage{},
			err:              errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:     "list channels with service response that can't be unmarshalled",
			token:    validToken,
			domainID: domainID,
			offset:   0,
			limit:    10,
			channelsPageMeta: channels.PageMetadata{
				Offset:     0,
				Limit:      10,
				Permission: defPermission,
			},
			svcRes: channels.Page{
				PageMetadata: channels.PageMetadata{
					Total: 1,
				},
				Channels: []channels.Channel{{
					ID: generateUUID(t),
					Metadata: clients.Metadata{
						"test": make(chan int),
					},
				}},
			},
			svcErr:   nil,
			response: sdk.ChannelsPage{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			pm := sdk.PageMetadata{
				Offset:   tc.offset,
				Limit:    tc.limit,
				Level:    uint64(tc.level),
				Metadata: tc.metadata,
			}
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("ListChannels", mock.Anything, tc.session, tc.channelsPageMeta).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Channels(pm, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ListChannels", mock.Anything, tc.session, tc.channelsPageMeta)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestViewChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	channelRes := convertChannel(channel)
	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		svcRes          channels.Channel
		svcErr          error
		authenticateErr error
		response        sdk.Channel
		err             errors.SDKError
	}{
		{
			desc:      "view channel successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channelRes.ID,
			svcRes:    channelRes,
			svcErr:    nil,
			response:  channel,
			err:       nil,
		},
		{
			desc:            "view channel with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channelRes.ID,
			svcRes:          channels.Channel{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.Channel{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "view channel with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channelRes.ID,
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "view channel for wrong id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			svcRes:    channels.Channel{},
			svcErr:    svcerr.ErrViewEntity,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusBadRequest),
		},
		{
			desc:      "view channel with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKError(apiutil.ErrMissingID),
		},
		{
			desc:      "view channel with service response that can't be unmarshalled",
			domainID:  domainID,
			token:     validToken,
			channelID: channelRes.ID,
			svcRes: channels.Channel{
				ID: generateUUID(t),
				Metadata: clients.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Channel{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("ViewChannel", mock.Anything, tc.session, tc.channelID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Channel(tc.channelID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ViewChannel", mock.Anything, tc.session, tc.channelID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUpdateChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	mChannel := convertChannel(channel)
	mChannel.Metadata = clients.Metadata{
		"field": "value2",
	}
	msdkChannel := channel
	msdkChannel.Metadata = sdk.Metadata{
		"field": "value2",
	}

	nChannel := convertChannel(channel)
	nChannel.Name = newName
	nsdkChannel := channel
	nsdkChannel.Name = newName

	aChannel := convertChannel(channel)
	aChannel.Name = newName
	aChannel.Metadata = clients.Metadata{"field": "value2"}
	asdkChannel := channel
	asdkChannel.Name = newName
	asdkChannel.Metadata = sdk.Metadata{"field": "value2"}

	cases := []struct {
		desc             string
		domainID         string
		token            string
		session          smqauthn.Session
		channelReq       sdk.Channel
		updateChannelReq channels.Channel
		svcRes           channels.Channel
		svcErr           error
		authenticateErr  error
		response         sdk.Channel
		err              errors.SDKError
	}{
		{
			desc:     "update channel name",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			updateChannelReq: channels.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			svcRes:   nChannel,
			svcErr:   nil,
			response: nsdkChannel,
			err:      nil,
		},
		{
			desc:     "update channel metadata",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID: channel.ID,
				Metadata: sdk.Metadata{
					"field": "value2",
				},
			},
			updateChannelReq: channels.Channel{
				ID:       channel.ID,
				Metadata: clients.Metadata{"field": "value2"},
			},
			svcRes:   mChannel,
			svcErr:   nil,
			response: msdkChannel,
			err:      nil,
		},
		{
			desc:     "update channel with every field defined",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:       channel.ID,
				Name:     newName,
				Metadata: sdk.Metadata{"field": "value2"},
			},
			updateChannelReq: channels.Channel{
				ID:   channel.ID,
				Name: newName,

				Metadata: clients.Metadata{"field": "value2"},
			},
			svcRes:   aChannel,
			svcErr:   nil,
			response: asdkChannel,
			err:      nil,
		},
		{
			desc:     "update channel name with invalid channel id",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:   wrongID,
				Name: newName,
			},
			updateChannelReq: channels.Channel{
				ID:   wrongID,
				Name: newName,
			},
			svcRes:   channels.Channel{},
			svcErr:   svcerr.ErrNotFound,
			response: sdk.Channel{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrNotFound, http.StatusNotFound),
		},
		{
			desc:     "update channel description with invalid channel id",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID: wrongID,
			},
			updateChannelReq: channels.Channel{
				ID: wrongID,
			},
			svcRes:   channels.Channel{},
			svcErr:   svcerr.ErrNotFound,
			response: sdk.Channel{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrNotFound, http.StatusNotFound),
		},
		{
			desc:     "update channel metadata with invalid channel id",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID: wrongID,
				Metadata: sdk.Metadata{
					"field": "value2",
				},
			},
			updateChannelReq: channels.Channel{
				ID:       wrongID,
				Metadata: clients.Metadata{"field": "value2"},
			},
			svcRes:   channels.Channel{},
			svcErr:   svcerr.ErrNotFound,
			response: sdk.Channel{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrNotFound, http.StatusNotFound),
		},
		{
			desc:     "update channel with invalid token",
			domainID: domainID,
			token:    invalidToken,
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			updateChannelReq: channels.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			svcRes:          channels.Channel{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.Channel{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "update channel with empty token",
			domainID: domainID,
			token:    "",
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			updateChannelReq: channels.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			svcRes:   channels.Channel{},
			svcErr:   nil,
			response: sdk.Channel{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "update channel with name that is too long",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: strings.Repeat("a", 1025),
			},
			updateChannelReq: channels.Channel{},
			svcRes:           channels.Channel{},
			svcErr:           nil,
			response:         sdk.Channel{},
			err:              errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrNameSize), http.StatusBadRequest),
		},
		{
			desc:     "update channel that can't be marshalled",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: "test",
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			updateChannelReq: channels.Channel{},
			svcRes:           channels.Channel{},
			svcErr:           nil,
			response:         sdk.Channel{},
			err:              errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:     "update channel with service response that can't be unmarshalled",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			updateChannelReq: channels.Channel{
				ID:   channel.ID,
				Name: newName,
			},
			svcRes: channels.Channel{
				ID: generateUUID(t),
				Metadata: clients.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Channel{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
		{
			desc:     "update channel with empty channel id",
			domainID: domainID,
			token:    validToken,
			channelReq: sdk.Channel{
				Name: newName,
			},
			updateChannelReq: channels.Channel{},
			svcRes:           channels.Channel{},
			svcErr:           nil,
			response:         sdk.Channel{},
			err:              errors.NewSDKError(apiutil.ErrMissingID),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("UpdateChannel", mock.Anything, tc.session, tc.updateChannelReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.UpdateChannel(tc.channelReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "UpdateChannel", mock.Anything, tc.session, tc.updateChannelReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestEnableChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		svcRes          channels.Channel
		svcErr          error
		authenticateErr error
		response        sdk.Channel
		err             errors.SDKError
	}{
		{
			desc:      "enable channel successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			svcRes:    convertChannel(channel),
			svcErr:    nil,
			response:  channel,
			err:       nil,
		},
		{
			desc:            "enable channel with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channel.ID,
			svcRes:          channels.Channel{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.Channel{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "enable channel with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channel.ID,
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "enable channel with invalid channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			svcRes:    channels.Channel{},
			svcErr:    svcerr.ErrNotFound,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrNotFound, http.StatusNotFound),
		},
		{
			desc:      "enable channel with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "enable channel with service response that can't be unmarshalled",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			svcRes: channels.Channel{
				ID: generateUUID(t),
				Metadata: clients.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Channel{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("EnableChannel", mock.Anything, tc.session, tc.channelID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.EnableChannel(tc.channelID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "EnableChannel", mock.Anything, tc.session, tc.channelID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDisableChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	dChannel := channel
	dChannel.Status = clients.DisabledStatus.String()

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		svcRes          channels.Channel
		svcErr          error
		authenticateErr error
		response        sdk.Channel
		err             errors.SDKError
	}{
		{
			desc:      "disable channel successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			svcRes:    convertChannel(dChannel),
			svcErr:    nil,
			response:  dChannel,
			err:       nil,
		},
		{
			desc:            "disable channel with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channel.ID,
			svcRes:          channels.Channel{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.Channel{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "disable channel with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channel.ID,
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "disable channel with invalid channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			svcRes:    channels.Channel{},
			svcErr:    svcerr.ErrNotFound,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrNotFound, http.StatusNotFound),
		},
		{
			desc:      "disable channel with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			svcRes:    channels.Channel{},
			svcErr:    nil,
			response:  sdk.Channel{},
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "disable channel with service response that can't be unmarshalled",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			svcRes: channels.Channel{
				ID: generateUUID(t),
				Metadata: clients.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Channel{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("DisableChannel", mock.Anything, tc.session, tc.channelID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.DisableChannel(tc.channelID, tc.domainID, tc.token)
			fmt.Println(resp)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "DisableChannel", mock.Anything, tc.session, tc.channelID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDeleteChannel(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		svcErr          error
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:      "delete channel successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			svcErr:    nil,
			err:       nil,
		},
		{
			desc:            "delete channel with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channel.ID,
			authenticateErr: svcerr.ErrAuthentication,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "delete channel with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channel.ID,
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "delete channel with invalid channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			svcErr:    svcerr.ErrRemoveEntity,
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrRemoveEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:      "delete channel with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			svcErr:    svcerr.ErrRemoveEntity,
			err:       errors.NewSDKError(apiutil.ErrMissingID),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("RemoveChannel", mock.Anything, tc.session, tc.channelID).Return(tc.svcErr)
			err := mgsdk.DeleteChannel(tc.channelID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "RemoveChannel", mock.Anything, tc.session, tc.channelID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestConnect(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	clientID := generateUUID(t)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		connection      sdk.Connection
		svcErr          error
		authenticateRes smqauthn.Session
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:     "connect successfully",
			domainID: domainID,
			token:    validToken,
			connection: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "connect with invalid token",
			domainID: domainID,
			token:    invalidToken,
			connection: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			authenticateErr: svcerr.ErrAuthentication,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "connect with empty token",
			domainID: domainID,
			token:    "",
			connection: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			err: errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "connect with invalid channel id",
			domainID: domainID,
			token:    validToken,
			connection: sdk.Connection{
				ChannelIDs: []string{wrongID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: svcerr.ErrAuthorization,
			err:    errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "connect with empty channel id",
			domainID: domainID,
			token:    validToken,
			connection: sdk.Connection{
				ChannelIDs: []string{},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "connect with empty client id",
			domainID: domainID,
			token:    validToken,
			connection: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			connTypes := []connections.ConnType{}
			for _, ct := range tc.connection.Types {
				connType, err := connections.ParseConnType(ct)
				assert.Nil(t, err, fmt.Sprintf("error parsing connection type %s", ct))
				connTypes = append(connTypes, connType)
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("Connect", mock.Anything, tc.session, tc.connection.ChannelIDs, tc.connection.ClientIDs, connTypes).Return(tc.svcErr)
			err := mgsdk.Connect(tc.connection, tc.domainID, tc.token)
			fmt.Println(err)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Connect", mock.Anything, tc.session, tc.connection.ChannelIDs, tc.connection.ClientIDs, connTypes)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDisconnect(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	clientID := generateUUID(t)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		disconnect      sdk.Connection
		svcErr          error
		authenticateRes smqauthn.Session
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:     "disconnect successfully",
			domainID: domainID,
			token:    validToken,
			disconnect: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "disconnect with invalid token",
			domainID: domainID,
			token:    invalidToken,
			disconnect: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			authenticateErr: svcerr.ErrAuthentication,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "disconnect with empty token",
			domainID: domainID,
			token:    "",
			disconnect: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			err: errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "disconnect with invalid channel id",
			domainID: domainID,
			token:    validToken,
			disconnect: sdk.Connection{
				ChannelIDs: []string{wrongID},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: svcerr.ErrAuthorization,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidIDFormat), http.StatusBadRequest),
		},
		{
			desc:     "disconnect with empty channel id",
			domainID: domainID,
			token:    validToken,
			disconnect: sdk.Connection{
				ChannelIDs: []string{},
				ClientIDs:  []string{clientID},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "disconnect with empty client id",
			domainID: domainID,
			token:    validToken,
			disconnect: sdk.Connection{
				ChannelIDs: []string{channel.ID},
				ClientIDs:  []string{},
				Types:      []string{"Publish", "Subscribe"},
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			connTypes := []connections.ConnType{}
			for _, ct := range tc.disconnect.Types {
				connType, err := connections.ParseConnType(ct)
				assert.Nil(t, err, fmt.Sprintf("error parsing connection type %s", ct))
				connTypes = append(connTypes, connType)
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("Disconnect", mock.Anything, tc.session, tc.disconnect.ChannelIDs, tc.disconnect.ClientIDs, connTypes).Return(tc.svcErr)
			err := mgsdk.Disconnect(tc.disconnect, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Disconnect", mock.Anything, tc.session, tc.disconnect.ChannelIDs, tc.disconnect.ClientIDs, connTypes)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestConnectClient(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	clientID := generateUUID(t)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		clientID        string
		connType        string
		svcErr          error
		authenticateRes smqauthn.Session
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:      "connect successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			clientID:  clientID,
			connType:  "Publish",
			svcErr:    nil,
			err:       nil,
		},
		{
			desc:            "connect with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channel.ID,
			clientID:        clientID,
			connType:        "Publish",
			authenticateErr: svcerr.ErrAuthentication,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "connect with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channel.ID,
			clientID:  clientID,
			connType:  "Publish",
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "connect with invalid channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			clientID:  clientID,
			connType:  "Publish",
			svcErr:    svcerr.ErrAuthorization,
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:      "connect with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			clientID:  clientID,
			connType:  "Publish",
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "connect with empty client id",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			clientID:  "",
			connType:  "Publish",
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidIDFormat), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			connType, err := connections.ParseConnType(tc.connType)
			assert.Nil(t, err, fmt.Sprintf("error parsing connection type %s", tc.connType))
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("Connect", mock.Anything, tc.session, []string{tc.channelID}, []string{tc.clientID}, []connections.ConnType{connType}).Return(tc.svcErr)
			err = mgsdk.ConnectClient(tc.clientID, tc.channelID, []string{tc.connType}, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Connect", mock.Anything, tc.session, []string{tc.channelID}, []string{tc.clientID}, []connections.ConnType{connType})
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDisconnectClient(t *testing.T) {
	ts, gsvc, auth := setupChannels()
	defer ts.Close()

	conf := sdk.Config{
		ChannelsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	clientID := generateUUID(t)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		channelID       string
		clientID        string
		connType        string
		svcErr          error
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:      "disconnect successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			clientID:  clientID,
			connType:  "Publish",
			svcErr:    nil,
			err:       nil,
		},
		{
			desc:            "disconnect with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			channelID:       channel.ID,
			clientID:        clientID,
			connType:        "Publish",
			authenticateErr: svcerr.ErrAuthentication,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "disconnect with empty token",
			domainID:  domainID,
			token:     "",
			channelID: channel.ID,
			clientID:  clientID,
			connType:  "Publish",
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "disconnect with invalid channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: wrongID,
			clientID:  clientID,
			connType:  "Publish",
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidIDFormat), http.StatusBadRequest),
		},
		{
			desc:      "disconnect with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			clientID:  clientID,
			connType:  "Publish",
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "disconnect with empty client id",
			domainID:  domainID,
			token:     validToken,
			channelID: channel.ID,
			clientID:  "",
			connType:  "Publish",
			svcErr:    nil,
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidIDFormat), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = smqauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			connType, err := connections.ParseConnType(tc.connType)
			assert.Nil(t, err, fmt.Sprintf("error parsing connection type %s", tc.connType))
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := gsvc.On("Disconnect", mock.Anything, tc.session, []string{tc.channelID}, []string{tc.clientID}, []connections.ConnType{connType}).Return(tc.svcErr)
			err = mgsdk.DisconnectClient(tc.clientID, tc.channelID, []string{tc.connType}, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Disconnect", mock.Anything, tc.session, []string{tc.channelID}, []string{tc.clientID}, []connections.ConnType{connType})
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func generateTestChannel(t *testing.T) sdk.Channel {
	createdAt, err := time.Parse(time.RFC3339, "2023-03-03T00:00:00Z")
	assert.Nil(t, err, fmt.Sprintf("unexpected error %s", err))
	updatedAt := createdAt
	ch := sdk.Channel{
		ID:        testsutil.GenerateUUID(&testing.T{}),
		DomainID:  testsutil.GenerateUUID(&testing.T{}),
		Name:      channelName,
		Metadata:  sdk.Metadata{"role": "client"},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return ch
}
