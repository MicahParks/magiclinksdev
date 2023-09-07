package handle

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/MicahParks/jwkset"

	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/email"
	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/rlimit"
	"github.com/MicahParks/magiclinksdev/storage"
)

// Server is the magiclinksdev server.
type Server struct {
	Config         config.Config
	Ctx            context.Context
	EmailProvider  email.Provider
	HTTPMux        *http.ServeMux
	JWKS           jwkset.JWKSet[storage.JWKSetCustomKeyMeta]
	Limiter        rlimit.RateLimiter
	MagicLink      magiclink.MagicLink[storage.MagicLinkCustomCreateArgs, storage.MagicLinkCustomReadResponse, storage.JWKSetCustomKeyMeta]
	Store          storage.Storage
	Logger         *slog.Logger
	MiddlewareHook MiddlewareHook
}

// MiddlewareToggle contains fields to turn middleware on and off.
type MiddlewareToggle struct {
	Admin     bool
	Authn     bool
	CommitTx  bool
	RateLimit bool
}

// MiddlewareOptions contains options for applying middleware.
type MiddlewareOptions struct {
	Handler http.Handler
	Path    string
	Toggle  MiddlewareToggle
}

// MiddlewareHook is a function that can be used to modify the middleware options.
type MiddlewareHook interface {
	Hook(options MiddlewareOptions) MiddlewareOptions
}

// MiddlewareHookFunc is a function that can be used to modify the middleware options.
type MiddlewareHookFunc func(options MiddlewareOptions) MiddlewareOptions

// Hook implements the MiddlewareHook interface.
func (h MiddlewareHookFunc) Hook(options MiddlewareOptions) MiddlewareOptions {
	return h(options)
}
