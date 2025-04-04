package http

import (
	"log/slog"

	"github.com/hantdev/mitras"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	api "github.com/hantdev/mitras/api/http"
	apiutil "github.com/hantdev/mitras/api/http/util"
	"github.com/hantdev/mitras/channels"
	smqauthn "github.com/hantdev/mitras/pkg/authn"
	roleManagerHttp "github.com/hantdev/mitras/pkg/roles/rolemanager/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// MakeHandler returns a HTTP handler for Channels API endpoints.
func MakeHandler(svc channels.Service, authn smqauthn.Authentication, mux *chi.Mux, logger *slog.Logger, instanceID string, idp mitras.IDProvider) *chi.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	d := roleManagerHttp.NewDecoder("channelID")

	mux.Route("/{domainID}/channels", func(r chi.Router) {
		r.Use(api.AuthenticateMiddleware(authn, true))
		r.Use(api.RequestIDMiddleware(idp))

		r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
			createChannelEndpoint(svc),
			decodeCreateChannelReq,
			api.EncodeResponse,
			opts...,
		), "create_channel").ServeHTTP)

		r.Post("/bulk", otelhttp.NewHandler(kithttp.NewServer(
			createChannelsEndpoint(svc),
			decodeCreateChannelsReq,
			api.EncodeResponse,
			opts...,
		), "create_channels").ServeHTTP)

		r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
			listChannelsEndpoint(svc),
			decodeListChannels,
			api.EncodeResponse,
			opts...,
		), "list_channels").ServeHTTP)

		r.Post("/connect", otelhttp.NewHandler(kithttp.NewServer(
			connectEndpoint(svc),
			decodeConnectRequest,
			api.EncodeResponse,
			opts...,
		), "connect").ServeHTTP)

		r.Post("/disconnect", otelhttp.NewHandler(kithttp.NewServer(
			disconnectEndpoint(svc),
			decodeDisconnectRequest,
			api.EncodeResponse,
			opts...,
		), "disconnect").ServeHTTP)

		r = roleManagerHttp.EntityAvailableActionsRouter(svc, d, r, opts)

		r.Route("/{channelID}", func(r chi.Router) {
			r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
				viewChannelEndpoint(svc),
				decodeViewChannel,
				api.EncodeResponse,
				opts...,
			), "view_channel").ServeHTTP)

			r.Patch("/", otelhttp.NewHandler(kithttp.NewServer(
				updateChannelEndpoint(svc),
				decodeUpdateChannel,
				api.EncodeResponse,
				opts...,
			), "update_channel_name_and_metadata").ServeHTTP)

			r.Patch("/tags", otelhttp.NewHandler(kithttp.NewServer(
				updateChannelTagsEndpoint(svc),
				decodeUpdateChannelTags,
				api.EncodeResponse,
				opts...,
			), "update_channel_tag").ServeHTTP)

			r.Delete("/", otelhttp.NewHandler(kithttp.NewServer(
				deleteChannelEndpoint(svc),
				decodeDeleteChannelReq,
				api.EncodeResponse,
				opts...,
			), "delete_channel").ServeHTTP)

			r.Post("/enable", otelhttp.NewHandler(kithttp.NewServer(
				enableChannelEndpoint(svc),
				decodeChangeChannelStatus,
				api.EncodeResponse,
				opts...,
			), "enable_channel").ServeHTTP)

			r.Post("/disable", otelhttp.NewHandler(kithttp.NewServer(
				disableChannelEndpoint(svc),
				decodeChangeChannelStatus,
				api.EncodeResponse,
				opts...,
			), "disable_channel").ServeHTTP)

			r.Post("/parent", otelhttp.NewHandler(kithttp.NewServer(
				setChannelParentGroupEndpoint(svc),
				decodeSetChannelParentGroupStatus,
				api.EncodeResponse,
				opts...,
			), "set_channel_parent_group").ServeHTTP)

			r.Delete("/parent", otelhttp.NewHandler(kithttp.NewServer(
				removeChannelParentGroupEndpoint(svc),
				decodeRemoveChannelParentGroupStatus,
				api.EncodeResponse,
				opts...,
			), "remove_channel_parent_group").ServeHTTP)

			r.Post("/connect", otelhttp.NewHandler(kithttp.NewServer(
				connectChannelClientEndpoint(svc),
				decodeConnectChannelClientRequest,
				api.EncodeResponse,
				opts...,
			), "connect_channel_client").ServeHTTP)

			r.Post("/disconnect", otelhttp.NewHandler(kithttp.NewServer(
				disconnectChannelClientsEndpoint(svc),
				decodeDisconnectChannelClientsRequest,
				api.EncodeResponse,
				opts...,
			), "disconnect_channel_client").ServeHTTP)

			roleManagerHttp.EntityRoleMangerRouter(svc, d, r, opts)
		})
	})

	mux.Get("/health", mitras.Health("channels", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
