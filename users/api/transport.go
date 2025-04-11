package api

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/hantdev/mitras"
	"github.com/go-chi/chi/v5"
	grpcTokenV1 "github.com/hantdev/mitras/internal/grpc/token/v1"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/hantdev/mitras/pkg/oauth2"
	"github.com/hantdev/mitras/users"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for Users and Groups API endpoints.
func MakeHandler(cls users.Service, authn smqauthn.Authentication, tokensvc grpcTokenV1.TokenServiceClient, selfRegister bool, mux *chi.Mux, logger *slog.Logger, instanceID string, pr *regexp.Regexp, providers ...oauth2.Provider) http.Handler {
	mux = usersHandler(cls, authn, tokensvc, selfRegister, mux, logger, pr, providers...)

	mux.Get("/health", mitras.Health("users", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
