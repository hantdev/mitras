package api

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/hantdev/mitras"
	grpcTokenV1 "github.com/hantdev/mitras/api/grpc/token/v1"
	mitrasauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/oauth2"
	"github.com/hantdev/mitras/users"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for Users and Groups API endpoints.
func MakeHandler(cls users.Service, authn mitrasauthn.Authentication, tokensvc grpcTokenV1.TokenServiceClient, selfRegister bool, mux *chi.Mux, logger *slog.Logger, instanceID string, pr *regexp.Regexp, idp mitras.IDProvider, providers ...oauth2.Provider) http.Handler {
	mux = usersHandler(cls, authn, tokensvc, selfRegister, mux, logger, pr, idp, providers...)

	mux.Get("/health", mitras.Health("users", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}