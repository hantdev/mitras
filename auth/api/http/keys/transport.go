package keys

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	api "github.com/hantdev/mitras/api/http"
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/auth"
	"github.com/hantdev/mitras/pkg/errors"
)

const contentType = "application/json"

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, mux *chi.Mux, logger *slog.Logger) *chi.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Route("/keys", func(r chi.Router) {
		r.Post("/", kithttp.NewServer(
			issueEndpoint(svc),
			decodeIssue,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Get("/{id}", kithttp.NewServer(
			(retrieveEndpoint(svc)),
			decodeKeyReq,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Delete("/{id}", kithttp.NewServer(
			(revokeEndpoint(svc)),
			decodeKeyReq,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)
	})
	return mux
}

func decodeIssue(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := issueKeyReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeKeyReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := keyReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	return req, nil
}
