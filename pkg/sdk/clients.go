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
	permissionsEndpoint = "permissions"
	clientsEndpoint     = "clients"
	connectEndpoint     = "connect"
	disconnectEndpoint  = "disconnect"
	identifyEndpoint    = "identify"
	rolesEndpoint       = "roles"
	actionsEndpoint     = "actions"
)

// Client represents mitras client.
type Client struct {
	ID          string                    `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	DomainID    string                    `json:"domain_id,omitempty"`
	ParentGroup string                    `json:"parent_group_id,omitempty"`
	Credentials ClientCredentials         `json:"credentials"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
	CreatedAt   time.Time                 `json:"created_at,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at,omitempty"`
	UpdatedBy   string                    `json:"updated_by,omitempty"`
	Status      string                    `json:"status,omitempty"`
	Permissions []string                  `json:"permissions,omitempty"`
	Roles       []roles.MemberRoleActions `json:"roles,omitempty"`
}

type ClientCredentials struct {
	Identity string `json:"identity,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

func (sdk mitrasSDK) CreateClient(client Client, domainID, token string) (Client, errors.SDKError) {
	data, err := json.Marshal(client)
	if err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusCreated)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	client = Client{}
	if err := json.Unmarshal(body, &client); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return client, nil
}

func (sdk mitrasSDK) CreateClients(clients []Client, domainID, token string) ([]Client, errors.SDKError) {
	data, err := json.Marshal(clients)
	if err != nil {
		return []Client{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, "bulk")

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return []Client{}, sdkerr
	}

	var ctr createClientsRes
	if err := json.Unmarshal(body, &ctr); err != nil {
		return []Client{}, errors.NewSDKError(err)
	}

	return ctr.Clients, nil
}

func (sdk mitrasSDK) Clients(pm PageMetadata, domainID, token string) (ClientsPage, errors.SDKError) {
	endpoint := fmt.Sprintf("%s/%s", domainID, clientsEndpoint)
	url, err := sdk.withQueryParams(sdk.clientsURL, endpoint, pm)
	if err != nil {
		return ClientsPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return ClientsPage{}, sdkerr
	}

	var cp ClientsPage
	if err := json.Unmarshal(body, &cp); err != nil {
		return ClientsPage{}, errors.NewSDKError(err)
	}

	return cp, nil
}

func (sdk mitrasSDK) Client(id, domainID, token string) (Client, errors.SDKError) {
	if id == "" {
		return Client{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, id)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	var t Client
	if err := json.Unmarshal(body, &t); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mitrasSDK) UpdateClient(t Client, domainID, token string) (Client, errors.SDKError) {
	if t.ID == "" {
		return Client{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, t.ID)

	data, err := json.Marshal(t)
	if err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	t = Client{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mitrasSDK) UpdateClientTags(t Client, domainID, token string) (Client, errors.SDKError) {
	data, err := json.Marshal(t)
	if err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/tags", sdk.clientsURL, domainID, clientsEndpoint, t.ID)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	t = Client{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mitrasSDK) UpdateClientSecret(id, secret, domainID, token string) (Client, errors.SDKError) {
	ucsr := updateClientSecretReq{Secret: secret}

	data, err := json.Marshal(ucsr)
	if err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/secret", sdk.clientsURL, domainID, clientsEndpoint, id)

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	var t Client
	if err = json.Unmarshal(body, &t); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mitrasSDK) EnableClient(id, domainID, token string) (Client, errors.SDKError) {
	return sdk.changeClientStatus(id, enableEndpoint, domainID, token)
}

func (sdk mitrasSDK) DisableClient(id, domainID, token string) (Client, errors.SDKError) {
	return sdk.changeClientStatus(id, disableEndpoint, domainID, token)
}

func (sdk mitrasSDK) changeClientStatus(id, status, domainID, token string) (Client, errors.SDKError) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, id, status)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return Client{}, sdkerr
	}

	t := Client{}
	if err := json.Unmarshal(body, &t); err != nil {
		return Client{}, errors.NewSDKError(err)
	}

	return t, nil
}

func (sdk mitrasSDK) SetClientParent(id, domainID, groupID, token string) errors.SDKError {
	scpg := parentGroupReq{ParentGroupID: groupID}
	data, err := json.Marshal(scpg)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, id, parentEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusOK)

	return sdkerr
}

func (sdk mitrasSDK) RemoveClientParent(id, domainID, groupID, token string) errors.SDKError {
	rcpg := parentGroupReq{ParentGroupID: groupID}
	data, err := json.Marshal(rcpg)
	if err != nil {
		return errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, id, parentEndpoint)
	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, data, nil, http.StatusNoContent)

	return sdkerr
}

func (sdk mitrasSDK) DeleteClient(id, domainID, token string) errors.SDKError {
	if id == "" {
		return errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.clientsURL, domainID, clientsEndpoint, id)
	_, _, sdkerr := sdk.processRequest(http.MethodDelete, url, token, nil, nil, http.StatusNoContent)
	return sdkerr
}

func (sdk mitrasSDK) CreateClientRole(id, domainID string, rq RoleReq, token string) (Role, errors.SDKError) {
	return sdk.createRole(sdk.clientsURL, clientsEndpoint, id, domainID, rq, token)
}

func (sdk mitrasSDK) ClientRoles(id, domainID string, pm PageMetadata, token string) (RolesPage, errors.SDKError) {
	return sdk.listRoles(sdk.clientsURL, clientsEndpoint, id, domainID, pm, token)
}

func (sdk mitrasSDK) ClientRole(id, roleID, domainID, token string) (Role, errors.SDKError) {
	return sdk.viewRole(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, token)
}

func (sdk mitrasSDK) UpdateClientRole(id, roleID, newName, domainID string, token string) (Role, errors.SDKError) {
	return sdk.updateRole(sdk.clientsURL, clientsEndpoint, id, roleID, newName, domainID, token)
}

func (sdk mitrasSDK) DeleteClientRole(id, roleID, domainID, token string) errors.SDKError {
	return sdk.deleteRole(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, token)
}

func (sdk mitrasSDK) AddClientRoleActions(id, roleID, domainID string, actions []string, token string) ([]string, errors.SDKError) {
	return sdk.addRoleActions(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, actions, token)
}

func (sdk mitrasSDK) ClientRoleActions(id, roleID, domainID string, token string) ([]string, errors.SDKError) {
	return sdk.listRoleActions(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, token)
}

func (sdk mitrasSDK) RemoveClientRoleActions(id, roleID, domainID string, actions []string, token string) errors.SDKError {
	return sdk.removeRoleActions(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, actions, token)
}

func (sdk mitrasSDK) RemoveAllClientRoleActions(id, roleID, domainID, token string) errors.SDKError {
	return sdk.removeAllRoleActions(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, token)
}

func (sdk mitrasSDK) AddClientRoleMembers(id, roleID, domainID string, members []string, token string) ([]string, errors.SDKError) {
	return sdk.addRoleMembers(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, members, token)
}

func (sdk mitrasSDK) ClientRoleMembers(id, roleID, domainID string, pm PageMetadata, token string) (RoleMembersPage, errors.SDKError) {
	return sdk.listRoleMembers(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, pm, token)
}

func (sdk mitrasSDK) RemoveClientRoleMembers(id, roleID, domainID string, members []string, token string) errors.SDKError {
	return sdk.removeRoleMembers(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, members, token)
}

func (sdk mitrasSDK) RemoveAllClientRoleMembers(id, roleID, domainID, token string) errors.SDKError {
	return sdk.removeAllRoleMembers(sdk.clientsURL, clientsEndpoint, id, roleID, domainID, token)
}

func (sdk mitrasSDK) AvailableClientRoleActions(domainID, token string) ([]string, errors.SDKError) {
	return sdk.listAvailableRoleActions(sdk.clientsURL, clientsEndpoint, domainID, token)
}

func (sdk mitrasSDK) ListClientMembers(clientID, domainID string, pm PageMetadata, token string) (EntityMembersPage, errors.SDKError) {
	return sdk.listEntityMembers(sdk.clientsURL, domainID, clientsEndpoint, clientID, token, pm)
}
