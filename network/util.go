package network

import (
	"fmt"
	"net/http"

	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

const (
	// PathJWKS is the path to the JWKS endpoint.
	PathJWKS = "jwks.json"
	// PathReady is the path to the ready endpoint.
	PathReady = "ready"
	// PathServiceAccountCreate is the path to the service account creation endpoint.
	PathServiceAccountCreate = "admin/service-account/create"
	// PathJWTCreate is the path to the JWT creation endpoint.
	PathJWTCreate = "jwt/create"
	// PathJWTValidate is the path to the JWT validation endpoint.
	PathJWTValidate = "jwt/validate"
	// PathMagicLinkCreate is the path to the link creation endpoint.
	PathMagicLinkCreate = "magic-link/create"
	// PathMagicLinkEmailCreate is the path to the magic link email creation endpoint.
	PathMagicLinkEmailCreate = "magic-link-email/create"
)

// CreateHTTPHandlers creates the HTTP handlers for the server.
func CreateHTTPHandlers(server *handle.Server) (*http.ServeMux, error) {
	pathMagicLinkHandler := server.Config.RelativeRedirectURL.Get().EscapedPath()
	options := []handle.MiddlewareOptions{
		{
			Handler: server.MagicLink.JWKSHandler(),
			Path:    PathJWKS,
			Toggle: handle.MiddlewareToggle{
				CommitTx: true,
			},
		},
		{
			Handler: server.MagicLink.MagicLinkHandler(),
			Path:    pathMagicLinkHandler,
			Toggle: handle.MiddlewareToggle{
				CommitTx: true,
			},
		},
		{
			Handler: HTTPReady(server),
			Path:    PathReady,
			Toggle:  handle.MiddlewareToggle{},
		},
		{
			Handler: HTTPServiceAccountCreate(server),
			Path:    PathServiceAccountCreate,
			Toggle: handle.MiddlewareToggle{
				Admin: true,
				Authn: true,
			},
		},
		{
			Handler: HTTPJWTCreate(server),
			Path:    PathJWTCreate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
			},
		},
		{
			Handler: HTTPJWTValidate(server),
			Path:    PathJWTValidate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
			},
		},
		{
			Handler: HTTPMagicLinkCreate(server),
			Path:    PathMagicLinkCreate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
			},
		},
		{
			Handler: HTTPMagicLinkEmailCreate(server),
			Path:    PathMagicLinkEmailCreate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
			},
		},
	}

	for _, opt := range options {
		u, err := server.Config.BaseURL.Get().Parse(opt.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse path %q: %w", opt.Path, err)
		}
		server.HTTPMux.Handle(u.Path, middleware.Apply(server, opt))
	}

	return server.HTTPMux, nil
}
