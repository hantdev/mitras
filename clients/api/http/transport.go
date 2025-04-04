package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hantdev/mitras"
	"github.com/hantdev/mitras/clients"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for clients and Groups API endpoints.
func MakeHandler(tsvc clients.Service, authn smqauthn.Authentication, mux *chi.Mux, logger *slog.Logger, instanceID string, idp mitras.IDProvider) http.Handler {
	mux = clientsHandler(tsvc, authn, mux, logger, idp)

	mux.Get("/health", mitras.Health("clients", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
