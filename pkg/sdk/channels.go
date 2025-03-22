package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/pkg/errors"
	"github.com/hantdev/mitras/pkg/roles"
)

const (
	channelsEndpoint = "channels"
	parentEndpoint   = "parent"
)

// Channel represents supermq channel.
type Channel struct {
	ID          string                    `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	ParentGroup string                    `json:"parent_group_id,omitempty"`
	DomainID    string                    `json:"domain_id,omitempty"`
	Metadata    Metadata                  `json:"metadata,omitempty"`
	CreatedAt   time.Time                 `json:"created_at,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at,omitempty"`
	UpdatedBy   string                    `json:"updated_by,omitempty"`
	Status      string                    `json:"status,omitempty"`
	Permissions []string                  `json:"permissions,omitempty"`
	Roles       []roles.MemberRoleActions `json:"roles,omitempty"`
}

func (sdk mitrasSDK) CreateChannel(c Channel, domainID, token string) (Channel, errors.SDKError) {
	data, err := json.Marshal(c)
	if err != nil {
		return Channel{}, errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusCreated)
	if sdkerr != nil {
		return Channel{}, sdkerr
	}

	c = Channel{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	return c, nil
}

func (sdk mitrasSDK) CreateChannels(channels []Channel, domainID, token string) ([]Channel, errors.SDKError) {
	data, err := json.Marshal(channels)
	if err != nil {
		return []Channel{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, "bulk")

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return []Channel{}, sdkerr
	}

	res := createChannelsRes{}
	if err := json.Unmarshal(body, &res); err != nil {
		return []Channel{}, errors.NewSDKError(err)
	}

	return res.Channels, nil
}

func (sdk mitrasSDK) Channels(pm PageMetadata, domainID, token string) (ChannelsPage, errors.SDKError) {
	endpoint := fmt.Sprintf("%s/%s", domainID, channelsEndpoint)
	url, err := sdk.withQueryParams(sdk.channelsURL, endpoint, pm)
	if err != nil {
		return ChannelsPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return ChannelsPage{}, sdkerr
	}

	var cp ChannelsPage
	if err = json.Unmarshal(body, &cp); err != nil {
		return ChannelsPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

func (sdk mitrasSDK) Channel(id, domainID, token string) (Channel, errors.SDKError) {
	if id == "" {
		return Channel{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, id)

	_, body, err := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if err != nil {
		return Channel{}, err
	}

	var c Channel
	if err := json.Unmarshal(body, &c); err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	return c, nil
}

func (sdk mitrasSDK) UpdateChannel(c Channel, domainID, token string) (Channel, errors.SDKError) {
	if c.ID == "" {
		return Channel{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, c.ID)

	data, err := json.Marshal(c)
	if err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Channel{}, sdkerr
	}

	c = Channel{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	return c, nil
}

func (sdk mitrasSDK) UpdateChannelTags(c Channel, domainID, token string) (Channel, errors.SDKError) {
	if c.ID == "" {
		return Channel{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s/tags", sdk.channelsURL, domainID, channelsEndpoint, c.ID)

	data, err := json.Marshal(c)
	if err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Channel{}, sdkerr
	}

	c = Channel{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	return c, nil
}

func (sdk mitrasSDK) Connect(conn Connection, domainID, token string) errors.SDKError {
	data, err := json.Marshal(conn)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, connectEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusCreated)

	return sdkerr
}

func (sdk mitrasSDK) Disconnect(conn Connection, domainID, token string) errors.SDKError {
	data, err := json.Marshal(conn)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, disconnectEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusNoContent)

	return sdkerr
}

func (sdk mitrasSDK) ConnectClients(channelID string, clientIDs, connTypes []string, domainID, token string) errors.SDKError {
	conn := Connection{
		ClientIDs: clientIDs,
		Types:     connTypes,
	}
	data, err := json.Marshal(conn)
	if err != nil {
		return errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, channelID, connectEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusCreated)

	return sdkerr
}

func (sdk mitrasSDK) DisconnectClients(channelID string, clientIDs, connTypes []string, domainID, token string) errors.SDKError {
	conn := Connection{
		ClientIDs: clientIDs,
		Types:     connTypes,
	}
	data, err := json.Marshal(conn)
	if err != nil {
		return errors.NewSDKError(err)
	}
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, channelID, disconnectEndpoint)

	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusNoContent)

	return sdkerr
}

func (sdk mitrasSDK) EnableChannel(id, domainID, token string) (Channel, errors.SDKError) {
	return sdk.changeChannelStatus(id, enableEndpoint, domainID, token)
}

func (sdk mitrasSDK) DisableChannel(id, domainID, token string) (Channel, errors.SDKError) {
	return sdk.changeChannelStatus(id, disableEndpoint, domainID, token)
}

func (sdk mitrasSDK) DeleteChannel(id, domainID, token string) errors.SDKError {
	if id == "" {
		return errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, id)
	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, nil, nil, http.StatusNoContent)
	return sdkerr
}

func (sdk mitrasSDK) changeChannelStatus(id, status, domainID, token string) (Channel, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, id, status)

	_, body, err := sdk.processRequest(http.MethodPost, url, token, nil, nil, http.StatusOK)
	if err != nil {
		return Channel{}, err
	}
	c := Channel{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Channel{}, errors.NewSDKError(err)
	}

	return c, nil
}

func (sdk mitrasSDK) SetChannelParent(id, domainID, groupID, token string) errors.SDKError {
	scpg := parentGroupReq{ParentGroupID: groupID}
	data, err := json.Marshal(scpg)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, id, parentEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusOK)

	return sdkerr
}

func (sdk mitrasSDK) RemoveChannelParent(id, domainID, groupID, token string) errors.SDKError {
	rcpg := parentGroupReq{ParentGroupID: groupID}
	data, err := json.Marshal(rcpg)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.channelsURL, domainID, channelsEndpoint, id, parentEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, data, nil, http.StatusNoContent)

	return sdkerr
}

func (sdk mitrasSDK) ListChannelMembers(channelID, domainID string, pm PageMetadata, token string) (EntityMembersPage, errors.SDKError) {
	return sdk.listEntityMembers(sdk.channelsURL, domainID, channelsEndpoint, channelID, token, pm)
}
