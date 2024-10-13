package magiclink

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"

	mld "github.com/MicahParks/magiclinksdev"
)

// CreateParams are the parameters for creating a magic link.
type CreateParams struct {
	// Expires is the time the magic link will expire. Use of this field is REQUIRED for all use cases.
	Expires time.Time

	// JWTClaims is a data structure that can marshal to JSON as the JWT Claims. Make sure to embed AND populate
	// jwt.RegisteredClaims if your use case supports standard claims. If you shadow the .Valid() method of the
	// interface, be sure to call .RegisteredClaims.Valid() in your implementation. If you are using a non-memory
	// Storage implementation, be sure the data structure is the proper Go type after being read. Use of this field is
	// OPTIONAL and RECOMMENDED for most use cases.
	JWTClaims jwt.Claims

	// JWTKeyID is the key ID of the key used to sign the JWT. If this is empty, the first key in the JWK Set will be
	// used. If no key with the key ID is present when the magic link is visited, the service will produce a HTTP 500
	// error. Use of this field is OPTIONAL.
	JWTKeyID *string

	// JWTSigningMethod is a string that can be passed to jwt.GetSigningMethod to get the appropriate jwt.SigningMethod
	// to sign the JWT with. Do not populate this field unless you know what you are doing. By default, the most secure
	// method will be used. If a key type other than ECDSA, EdDSA, or RSA is used, then the most secure HMAC signing
	// method will be chosen by default. Use of this field is OPTIONAL and NOT RECOMMENDED for most use cases.
	JWTSigningMethod string

	// RedirectQueryKey is the URL query key used in the redirect. It will contain the JWT. If this is empty, "jwt" will
	// be used. Use of this field is OPTIONAL.
	RedirectQueryKey string

	// RedirectURL is the URL to redirect to after the JWT is verified. Use of this field is REQUIRED for all use cases.
	RedirectURL *url.URL
}

// Valid confirms the CreateParams are valid.
func (p CreateParams) Valid() error {
	if p.Expires.IsZero() {
		return fmt.Errorf("%w: Expires is required", mld.ErrParams)
	}
	if p.RedirectURL == nil {
		return fmt.Errorf("%w: RedirectURL is required", mld.ErrParams)
	}
	return nil
}

// ReadResult is the result after a magic link has been read.
type ReadResult struct {
	// CreateParams are the parameters used to create the magic link.
	CreateParams CreateParams
	// Visited is the first time the magic link was visited. This is nil if the magic link has not been visited.
	Visited *time.Time
}

// CreateResponse is the response after a magic link has been created.
type CreateResponse struct {
	MagicLink *url.URL
	Secret    string
}

// Errors that ErrorHandler needs to handle.
var (
	// ErrJWKSEmpty is a possible error for an ErrorHandler implementation to handle.
	ErrJWKSEmpty = errors.New("JWK Set is empty")
	// ErrJWKSJSON is a possible error for an ErrorHandler implementation to handle.
	ErrJWKSJSON = errors.New("failed to get JWK Set as JSON")
	// ErrJWKSReadGivenKID is a possible error for an ErrorHandler implementation to handle.
	ErrJWKSReadGivenKID = errors.New("failed to read JWK with given key ID")
	// ErrJWKSSnapshot is a possible error for an ErrorHandler implementation to handle.
	ErrJWKSSnapshot = errors.New("failed to snapshot JWK Set")
	// ErrJWTSign is a possible error for an ErrorHandler implementation to handle.
	ErrJWTSign = errors.New("failed to sign JWT")
	// ErrLinkNotFound is a possible error for an ErrorHandler implementation to handle.
	ErrLinkNotFound = errors.New("link not found")
	// ErrMagicLinkMissingSecret is a possible error for an ErrorHandler implementation to handle.
	ErrMagicLinkMissingSecret = errors.New("visited magic link endpoint without a secret")
	// ErrMagicLinkRead is a possible error for an ErrorHandler implementation to handle.
	ErrMagicLinkRead = errors.New("failed to read the magic link from storage")
)

// ErrorHandlerParams are the parameters passed to an ErrorHandler when an error occurs.
type ErrorHandlerParams struct {
	Err                   error
	Request               *http.Request
	SuggestedResponseCode int
	Writer                http.ResponseWriter
}

// ErrorHandler handles errors that occur in MagicLink's HTTP handlers.
type ErrorHandler interface {
	// Handle consumes an error and writes a response to the given writer. The set of possible errors to check by
	// unwrapping with errors.Is is documented above the interface's source code.
	Handle(args ErrorHandlerParams)
}

// ErrorHandlerFunc is a function that implements the ErrorHandler interface.
type ErrorHandlerFunc func(args ErrorHandlerParams)

// Handle implements the ErrorHandler interface.
func (f ErrorHandlerFunc) Handle(args ErrorHandlerParams) {
	f(args)
}

// Config contains the required assets to create a MagicLink service.
type Config struct {
	ErrorHandler     ErrorHandler
	JWKS             JWKSParams
	CustomRedirector Redirector
	ServiceURL       *url.URL
	SecretQueryKey   string
	Store            Storage
}

// Valid confirms the Config is valid.
func (c Config) Valid() error {
	if c.ServiceURL == nil {
		return fmt.Errorf("%w: include a service URL, this is used to build magic links", mld.ErrParams)
	}
	return nil
}

// JWKSParams are the parameters for the MagicLink service's JWK Set.
type JWKSParams struct {
	CacheRefresh time.Duration
	Store        jwkset.Storage
}
