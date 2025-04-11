package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hantdev/mitras/domains"
	"github.com/hantdev/mitras/internal/api"
	"github.com/hantdev/mitras/pkg/apiutil"
	"github.com/hantdev/mitras/pkg/errors"
)

func decodeCreateDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := createDomainReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeRetrieveDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := retrieveDomainRequest{
		domainID: chi.URLParam(r, "domainID"),
	}
	return req, nil
}

func decodeUpdateDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := updateDomainReq{
		domainID: chi.URLParam(r, "domainID"),
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeListDomainRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	page, err := decodePageRequest(ctx, r)
	if err != nil {
		return nil, err
	}
	req := listDomainsReq{
		page: page,
	}

	return req, nil
}

func decodeEnableDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := enableDomainReq{
		domainID: chi.URLParam(r, "domainID"),
	}
	return req, nil
}

func decodeDisableDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := disableDomainReq{
		domainID: chi.URLParam(r, "domainID"),
	}
	return req, nil
}

func decodeFreezeDomainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := freezeDomainReq{
		domainID: chi.URLParam(r, "domainID"),
	}
	return req, nil
}

func decodePageRequest(_ context.Context, r *http.Request) (page, error) {
	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefClientStatus)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	st, err := domains.ToStatus(s)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	or, err := apiutil.ReadStringQuery(r, api.OrderKey, api.DefOrder)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	dir, err := apiutil.ReadStringQuery(r, api.DirKey, api.DefDir)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	m, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	n, err := apiutil.ReadStringQuery(r, api.NameKey, "")
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	t, err := apiutil.ReadStringQuery(r, api.TagKey, "")
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	allActions, err := apiutil.ReadStringQuery(r, api.ActionsKey, "")
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	var actions []string

	allActions = strings.TrimSpace(allActions)
	if allActions != "" {
		actions = strings.Split(allActions, ",")
	}
	roleID, err := apiutil.ReadStringQuery(r, api.RoleIDKey, "")
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	roleName, err := apiutil.ReadStringQuery(r, api.RoleNameKey, "")
	if err != nil {
		return page{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	return page{
		offset:   o,
		order:    or,
		dir:      dir,
		limit:    l,
		name:     n,
		metadata: m,
		tag:      t,
		roleID:   roleID,
		roleName: roleName,
		actions:  actions,
		status:   st,
	}, nil
}
