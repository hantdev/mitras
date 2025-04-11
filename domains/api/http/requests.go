package http

import (
	"github.com/hantdev/mitras/domains"
	"github.com/hantdev/mitras/pkg/apiutil"
)

type page struct {
	offset   uint64
	limit    uint64
	order    string
	dir      string
	name     string
	metadata map[string]interface{}
	tag      string
	roleID   string
	roleName string
	actions  []string
	status   domains.Status
}

type createDomainReq struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Alias    string                 `json:"alias"`
}

func (req createDomainReq) validate() error {
	if req.Name == "" {
		return apiutil.ErrMissingName
	}

	if req.Alias == "" {
		return apiutil.ErrMissingAlias
	}

	return nil
}

type retrieveDomainRequest struct {
	domainID string
}

func (req retrieveDomainRequest) validate() error {
	if req.domainID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type updateDomainReq struct {
	domainID string
	Name     *string                 `json:"name,omitempty"`
	Metadata *map[string]interface{} `json:"metadata,omitempty"`
	Tags     *[]string               `json:"tags,omitempty"`
	Alias    *string                 `json:"alias,omitempty"`
}

func (req updateDomainReq) validate() error {
	if req.domainID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type listDomainsReq struct {
	page
}

func (req listDomainsReq) validate() error {
	return nil
}

type enableDomainReq struct {
	domainID string
}

func (req enableDomainReq) validate() error {
	if req.domainID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type disableDomainReq struct {
	domainID string
}

func (req disableDomainReq) validate() error {
	if req.domainID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type freezeDomainReq struct {
	domainID string
}

func (req freezeDomainReq) validate() error {
	if req.domainID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}
