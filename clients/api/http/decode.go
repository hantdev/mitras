package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	api "github.com/hantdev/mitras/api/http"
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/clients"
	"github.com/hantdev/mitras/pkg/errors"
)

const clientID = "clientID"

func decodeViewClient(_ context.Context, r *http.Request) (interface{}, error) {
	roles, err := apiutil.ReadBoolQuery(r, api.RolesKey, false)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	req := viewClientReq{
		id:    chi.URLParam(r, clientID),
		roles: roles,
	}

	return req, nil
}

func decodeListClients(_ context.Context, r *http.Request) (interface{}, error) {
	name, err := apiutil.ReadStringQuery(r, api.NameKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	tag, err := apiutil.ReadStringQuery(r, api.TagKey, "")
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefGroupStatus)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	status, err := clients.ToStatus(s)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	meta, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	offset, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}
	limit, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	dir, err := apiutil.ReadStringQuery(r, api.DirKey, api.DefDir)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	order, err := apiutil.ReadStringQuery(r, api.OrderKey, api.DefOrder)
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	allActions, err := apiutil.ReadStringQuery(r, api.ActionsKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	actions := []string{}

	allActions = strings.TrimSpace(allActions)
	if allActions != "" {
		actions = strings.Split(allActions, ",")
	}
	roleID, err := apiutil.ReadStringQuery(r, api.RoleIDKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	roleName, err := apiutil.ReadStringQuery(r, api.RoleNameKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	accessType, err := apiutil.ReadStringQuery(r, api.AccessTypeKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	userID, err := apiutil.ReadStringQuery(r, api.UserKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	groupID, err := apiutil.ReadStringQuery(r, api.GroupKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	channelID, err := apiutil.ReadStringQuery(r, api.ChannelKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	connType, err := apiutil.ReadStringQuery(r, api.ConnTypeKey, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	id, err := apiutil.ReadStringQuery(r, api.IDOrder, "")
	if err != nil {
		return listClientsReq{}, errors.Wrap(apiutil.ErrValidation, err)
	}

	req := listClientsReq{
		Page: clients.Page{
			Name:           name,
			Tag:            tag,
			Status:         status,
			Metadata:       meta,
			RoleName:       roleName,
			RoleID:         roleID,
			Actions:        actions,
			AccessType:     accessType,
			Order:          order,
			Dir:            dir,
			Offset:         offset,
			Limit:          limit,
			Group:          groupID,
			Channel:        channelID,
			ConnectionType: connType,
			ID:             id,
		},
		userID: userID,
	}
	return req, nil
}

func decodeUpdateClient(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := updateClientReq{
		id: chi.URLParam(r, clientID),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return req, nil
}

func decodeUpdateClientTags(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := updateClientTagsReq{
		id: chi.URLParam(r, clientID),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return req, nil
}

func decodeUpdateClientCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := updateClientCredentialsReq{
		id: chi.URLParam(r, clientID),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return req, nil
}

func decodeCreateClientReq(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	var c clients.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	req := createClientReq{
		client: c,
	}

	return req, nil
}

func decodeCreateClientsReq(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	c := createClientsReq{}
	if err := json.NewDecoder(r.Body).Decode(&c.Clients); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return c, nil
}

func decodeChangeClientStatus(_ context.Context, r *http.Request) (interface{}, error) {
	req := changeClientStatusReq{
		id: chi.URLParam(r, clientID),
	}

	return req, nil
}

func decodeSetClientParentGroupStatus(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := setClientParentGroupReq{
		id: chi.URLParam(r, clientID),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func decodeRemoveClientParentGroupStatus(_ context.Context, r *http.Request) (interface{}, error) {
	req := removeClientParentGroupReq{
		id: chi.URLParam(r, clientID),
	}

	return req, nil
}

func decodeDeleteClientReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := deleteClientReq{
		id: chi.URLParam(r, clientID),
	}

	return req, nil
}
