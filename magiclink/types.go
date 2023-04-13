package magiclink

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v4"
)

var (
	// ErrArgs indicates that the given parameters are invalid.
	ErrArgs = errors.New("invalid arguments")
)

// CreateArgs are the arguments for creating a magic link.
type CreateArgs[CustomCreateArgs any] struct {
	// Custom is additional data or metadata for your use case.
	Custom CustomCreateArgs

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

// Valid confirms the CreateArgs are valid.
func (p CreateArgs[CustomCreateArgs]) Valid() error {
	if p.RedirectURL == nil {
		return fmt.Errorf("%w: RedirectURL is required", ErrArgs)
	}
	return nil
}

// ReadResponse is the response after a magic link has been read.
type ReadResponse[CustomCreateArgs, CustomReadResponse any] struct {
	// Custom is additional data or metadata for your use case.
	Custom CustomReadResponse

	// CreateArgs are the parameters used to create the magic link.
	CreateArgs CreateArgs[CustomCreateArgs]
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

// ErrorHandlerArgs are the arguments passed to an ErrorHandler when an error occurs.
type ErrorHandlerArgs struct {
	Err                   error
	Request               *http.Request
	SuggestedResponseCode int
	Writer                http.ResponseWriter
}

// ErrorHandler handles errors that occur in MagicLink's HTTP handlers.
type ErrorHandler interface {
	// Handle consumes an error and writes a response to the given writer. The set of possible errors to check by
	// unwrapping with errors.Is is documented above the interface's source code.
	Handle(args ErrorHandlerArgs)
}

// ErrorHandlerFunc is a function that implements the ErrorHandler interface.
type ErrorHandlerFunc func(args ErrorHandlerArgs)

// Handle implements the ErrorHandler interface.
func (f ErrorHandlerFunc) Handle(args ErrorHandlerArgs) {
	f(args)
}

const (
	// PreventRobotsDefault indicates that the default robot prevention should be used.
	PreventRobotsDefault PreventRobotsEnum = ""
	// PreventRobotsNone indicates that no robot prevention should be used.
	PreventRobotsNone PreventRobotsEnum = "none"
	// PreventRobotsReCAPTCHAV3 indicates that Google's reCAPTCHA v3 should be used to prevent robots.
	PreventRobotsReCAPTCHAV3 PreventRobotsEnum = "recaptchav3"
)

// PreventRobotsEnum is an enum that determines how robots will be prevented from following magic links.
type PreventRobotsEnum string

// ReCAPTCHAV3Config is the configuration for Google's reCAPTCHA v3.
type ReCAPTCHAV3Config struct {
	APKPackageName []string `json:"apkPackageName"`
	Action         []string `json:"action"`
	Hostname       []string `json:"hostname"`
	MinScore       float64  `json:"minScore"`
	SecretKey      string   `json:"secretKey"`
	TemplateConfig ReCAPTCHAV3TemplateConfig
}

func (r ReCAPTCHAV3Config) DefaultsAndValidate() (ReCAPTCHAV3Config, error) {
	if r.MinScore == 0 {
		r.MinScore = 0.5
	}
	if r.SecretKey == "" {
		return r, fmt.Errorf("%w: ReCAPTCHA v3 secret key is required", jt.ErrDefaultsAndValidate)
	}
	var err error
	r.TemplateConfig, err = r.TemplateConfig.DefaultsAndValidate()
	if err != nil {
		return r, fmt.Errorf("failed to validate ReCAPTCHA v3 template data: %w", err)
	}
	return r, nil
}

// Config contains the required assets to create a MagicLink service.
type Config[CustomCreateArgs, CustomReadResults, CustomKeyMeta any] struct {
	ErrorHandler         ErrorHandler
	HTMLTemplate         string
	JWKS                 JWKSArgs[CustomKeyMeta]
	PreventRobotsDefault PreventRobotsEnum
	ReCAPTCHAV3Config    ReCAPTCHAV3Config
	ServiceURL           *url.URL
	SecretQueryKey       string
	Store                Storage[CustomCreateArgs, CustomReadResults, CustomKeyMeta]
}

// Valid confirms the Config is valid.
func (c Config[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) Valid() error {
	// TODO If c.PreventRobotsDefault is an empty string, default to PreventRobotsNone. Do this in the mld code.
	switch c.PreventRobotsDefault {
	case PreventRobotsNone, PreventRobotsReCAPTCHAV3:
		// Proceed.
	default:
		return fmt.Errorf("%w: unknown PreventRobotsDefault value: %q", ErrArgs, c.PreventRobotsDefault)
	}
	if c.ServiceURL == nil {
		return fmt.Errorf("%w: include a service URL, this is used to build magic links", ErrArgs)
	}
	return nil
}

// JWKSArgs are the parameters for the MagicLink service's JWK Set.
type JWKSArgs[CustomKeyMeta any] struct {
	CacheRefresh time.Duration
	Store        jwkset.Storage[CustomKeyMeta]
}
