package network

import (
	"fmt"
	"net/http"

	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

const (
	// PathEmailLinkCreate is the path to the email link creation endpoint.
	PathEmailLinkCreate = "email-link/create"
	// PathJWKS is the path to the JWKS endpoint.
	PathJWKS = "jwks.json"
	// PathJWTCreate is the path to the JWT creation endpoint.
	PathJWTCreate = "jwt/create"
	// PathJWTValidate is the path to the JWT validation endpoint.
	PathJWTValidate = "jwt/validate"
	// PathLinkCreate is the path to the link creation endpoint.
	PathLinkCreate = "link/create"
	// PathReady is the path to the ready endpoint.
	PathReady = "ready"
	// PathServiceAccountCreate is the path to the service account creation endpoint.
	PathServiceAccountCreate = "admin/service-account/create"
)

// CreateHTTPHandlers creates the HTTP handlers for the server.
func CreateHTTPHandlers(server *handle.Server) (*http.ServeMux, error) {
	pathMagicLinkHandler := server.Config.RelativeRedirectURL.Get().EscapedPath()
	options := []handle.MiddlewareOptions{
		{
			Handler: server.MagicLink.MagicLinkHandler(),
			Path:    pathMagicLinkHandler,
			Toggle: handle.MiddlewareToggle{
				CommitTx: true,
			},
		},
		{
			Handler: HTTPEmailLinkCreate(server),
			Path:    PathEmailLinkCreate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
			},
		},
		{
			Handler: server.MagicLink.JWKSHandler(),
			Path:    PathJWKS,
			Toggle: handle.MiddlewareToggle{
				CommitTx: true,
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
			Handler: HTTPLinkCreate(server),
			Path:    PathLinkCreate,
			Toggle: handle.MiddlewareToggle{
				Authn:     true,
				RateLimit: true,
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
