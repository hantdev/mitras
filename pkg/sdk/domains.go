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
	domainsEndpoint = "domains"
	freezeEndpoint  = "freeze"
)

// Domain represents mitras domain.
type Domain struct {
	ID          string                    `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Metadata    Metadata                  `json:"metadata,omitempty"`
	Tags        []string                  `json:"tags,omitempty"`
	Alias       string                    `json:"alias,omitempty"`
	Status      string                    `json:"status,omitempty"`
	Permission  string                    `json:"permission,omitempty"`
	CreatedBy   string                    `json:"created_by,omitempty"`
	CreatedAt   time.Time                 `json:"created_at,omitempty"`
	UpdatedBy   string                    `json:"updated_by,omitempty"`
	UpdatedAt   time.Time                 `json:"updated_at,omitempty"`
	Permissions []string                  `json:"permissions,omitempty"`
	Roles       []roles.MemberRoleActions `json:"roles,omitempty"`
}

func (sdk mitrasSDK) CreateDomain(domain Domain, token string) (Domain, errors.SDKError) {
	data, err := json.Marshal(domain)
	if err != nil {
		return Domain{}, errors.NewSDKError(err)
	}

	url := fmt.Sprintf("%s/%s", sdk.domainsURL, domainsEndpoint)

	_, body, sdkerr := sdk.processRequest(http.MethodPost, url, token, data, nil, http.StatusCreated)
	if sdkerr != nil {
		return Domain{}, sdkerr
	}

	var d Domain
	if err := json.Unmarshal(body, &d); err != nil {
		return Domain{}, errors.NewSDKError(err)
	}
	return d, nil
}

func (sdk mitrasSDK) Domains(pm PageMetadata, token string) (DomainsPage, errors.SDKError) {
	url, err := sdk.withQueryParams(sdk.domainsURL, domainsEndpoint, pm)
	if err != nil {
		return DomainsPage{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return DomainsPage{}, sdkerr
	}

	var dp DomainsPage
	if err := json.Unmarshal(body, &dp); err != nil {
		return DomainsPage{}, errors.NewSDKError(err)
	}

	return dp, nil
}

func (sdk mitrasSDK) Domain(domainID, token string) (Domain, errors.SDKError) {
	if domainID == "" {
		return Domain{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.domainsURL, domainsEndpoint, domainID)

	_, body, sdkerr := sdk.processRequest(http.MethodGet, url, token, nil, nil, http.StatusOK)
	if sdkerr != nil {
		return Domain{}, sdkerr
	}

	var domain Domain
	if err := json.Unmarshal(body, &domain); err != nil {
		return Domain{}, errors.NewSDKError(err)
	}

	return domain, nil
}

func (sdk mitrasSDK) UpdateDomain(domain Domain, token string) (Domain, errors.SDKError) {
	if domain.ID == "" {
		return Domain{}, errors.NewSDKError(apiutil.ErrMissingID)
	}
	url := fmt.Sprintf("%s/%s/%s", sdk.domainsURL, domainsEndpoint, domain.ID)

	data, err := json.Marshal(domain)
	if err != nil {
		return Domain{}, errors.NewSDKError(err)
	}

	_, body, sdkerr := sdk.processRequest(http.MethodPatch, url, token, data, nil, http.StatusOK)
	if sdkerr != nil {
		return Domain{}, sdkerr
	}

	var d Domain
	if err := json.Unmarshal(body, &d); err != nil {
		return Domain{}, errors.NewSDKError(err)
	}
	return d, nil
}

func (sdk mitrasSDK) EnableDomain(domainID, token string) errors.SDKError {
	return sdk.changeDomainStatus(token, domainID, enableEndpoint)
}

func (sdk mitrasSDK) DisableDomain(domainID, token string) errors.SDKError {
	return sdk.changeDomainStatus(token, domainID, disableEndpoint)
}

func (sdk mitrasSDK) FreezeDomain(domainID, token string) errors.SDKError {
	return sdk.changeDomainStatus(token, domainID, freezeEndpoint)
}

func (sdk mitrasSDK) changeDomainStatus(token, id, status string) errors.SDKError {
	url := fmt.Sprintf("%s/%s/%s/%s", sdk.domainsURL, domainsEndpoint, id, status)
	_, _, sdkerr := sdk.processRequest(http.MethodPost, url, token, nil, nil, http.StatusOK)
	return sdkerr
}

func (sdk mitrasSDK) CreateDomainRole(id string, rq RoleReq, token string) (Role, errors.SDKError) {
	return sdk.createRole(sdk.domainsURL, domainsEndpoint, id, "", rq, token)
}

func (sdk mitrasSDK) DomainRoles(id string, pm PageMetadata, token string) (RolesPage, errors.SDKError) {
	return sdk.listRoles(sdk.domainsURL, domainsEndpoint, id, "", pm, token)
}

func (sdk mitrasSDK) DomainRole(id, roleID, token string) (Role, errors.SDKError) {
	return sdk.viewRole(sdk.domainsURL, domainsEndpoint, id, roleID, "", token)
}

func (sdk mitrasSDK) UpdateDomainRole(id, roleID, newName string, token string) (Role, errors.SDKError) {
	return sdk.updateRole(sdk.domainsURL, domainsEndpoint, id, roleID, newName, "", token)
}

func (sdk mitrasSDK) DeleteDomainRole(id, roleID, token string) errors.SDKError {
	return sdk.deleteRole(sdk.domainsURL, domainsEndpoint, id, roleID, "", token)
}

func (sdk mitrasSDK) AddDomainRoleActions(id, roleID string, actions []string, token string) ([]string, errors.SDKError) {
	return sdk.addRoleActions(sdk.domainsURL, domainsEndpoint, id, roleID, "", actions, token)
}

func (sdk mitrasSDK) DomainRoleActions(id, roleID string, token string) ([]string, errors.SDKError) {
	return sdk.listRoleActions(sdk.domainsURL, domainsEndpoint, id, roleID, "", token)
}

func (sdk mitrasSDK) RemoveDomainRoleActions(id, roleID string, actions []string, token string) errors.SDKError {
	return sdk.removeRoleActions(sdk.domainsURL, domainsEndpoint, id, roleID, "", actions, token)
}

func (sdk mitrasSDK) RemoveAllDomainRoleActions(id, roleID, token string) errors.SDKError {
	return sdk.removeAllRoleActions(sdk.domainsURL, domainsEndpoint, id, roleID, "", token)
}

func (sdk mitrasSDK) AddDomainRoleMembers(id, roleID string, members []string, token string) ([]string, errors.SDKError) {
	return sdk.addRoleMembers(sdk.domainsURL, domainsEndpoint, id, roleID, "", members, token)
}

func (sdk mitrasSDK) DomainRoleMembers(id, roleID string, pm PageMetadata, token string) (RoleMembersPage, errors.SDKError) {
	return sdk.listRoleMembers(sdk.domainsURL, domainsEndpoint, id, roleID, "", pm, token)
}

func (sdk mitrasSDK) RemoveDomainRoleMembers(id, roleID string, members []string, token string) errors.SDKError {
	return sdk.removeRoleMembers(sdk.domainsURL, domainsEndpoint, id, roleID, "", members, token)
}

func (sdk mitrasSDK) RemoveAllDomainRoleMembers(id, roleID, token string) errors.SDKError {
	return sdk.removeAllRoleMembers(sdk.domainsURL, domainsEndpoint, id, roleID, "", token)
}

func (sdk mitrasSDK) AvailableDomainRoleActions(token string) ([]string, errors.SDKError) {
	return sdk.listAvailableRoleActions(sdk.domainsURL, domainsEndpoint, "", token)
}

func (sdk mitrasSDK) ListDomainMembers(domainID string, pm PageMetadata, token string) (EntityMembersPage, errors.SDKError) {
	return sdk.listEntityMembers(sdk.domainsURL, domainID, domainsEndpoint, domainID, token, pm)
}
