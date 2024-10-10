package magiclink

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"

	mld "github.com/MicahParks/magiclinksdev"
)

const (
	// DefaultRedirectQueryKey is the default URL query parameter to contain the JWT in after a magic link has been
	// clicked.
	DefaultRedirectQueryKey = "jwt"
	// DefaultSecretQueryKey is the default URL query parameter to contain the secret for a magic link.
	DefaultSecretQueryKey = "secret"
)

// MagicLink holds the necessary assets for the magic link service.
type MagicLink struct {
	Store             Storage
	customRedirector  Redirector
	errorHandler      ErrorHandler
	jwks              *jwksCache
	reCAPTCHAV3Config ReCAPTCHAV3Config
	secretQueryKey    string
	serviceURL        *url.URL
}

// NewMagicLink creates a new MagicLink. The given setupCtx is only used during the creation of the MagicLink.
func NewMagicLink(setupCtx context.Context, config Config) (MagicLink, error) {
	var m MagicLink

	err := config.Valid()
	if err != nil {
		return m, fmt.Errorf("failed to validate config: %w", err)
	}

	jCache, err := newJWKSCache(setupCtx, config.JWKS)
	if err != nil {
		return m, fmt.Errorf("failed to create JWK Set cache: %w", err)
	}

	secretQueryKey := config.SecretQueryKey
	if secretQueryKey == "" {
		secretQueryKey = DefaultSecretQueryKey
	}

	var store Storage
	store = config.Store
	if store == nil {
		store = NewMemoryStorage()
	}

	m = MagicLink{
		Store:             store,
		customRedirector:  config.CustomRedirector,
		errorHandler:      config.ErrorHandler,
		jwks:              jCache,
		reCAPTCHAV3Config: ReCAPTCHAV3Config{},
		secretQueryKey:    secretQueryKey,
		serviceURL:        config.ServiceURL,
	}

	return m, nil
}

// JWKSHandler is an HTTP handler that responds to requests with the JWK Set as JSON.
func (m MagicLink) JWKSHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, err := m.jwks.get(request.Context())
		if err != nil {
			m.handleError(fmt.Errorf("%w: %s", ErrJWKSJSON, err), http.StatusInternalServerError, request, writer)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set(mld.HeaderContentType, mld.ContentTypeJSON)
		_, _ = writer.Write(body)
	})
}

// JWKSet is a getter method to return the underlying JWK Set.
func (m MagicLink) JWKSet() jwkset.Storage {
	return m.jwks.storage
}

// NewLink creates a magic link with the given parameters.
func (m MagicLink) NewLink(ctx context.Context, args CreateArgs) (CreateResponse, error) {
	err := args.Valid()
	if err != nil {
		return CreateResponse{}, fmt.Errorf("failed to validate args: %w", err)
	}

	secret, err := m.Store.Create(ctx, args)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("failed to create link: %w", err)
	}

	serviceURL := copyURL(m.serviceURL)
	queryResult := serviceURL.Query()
	queryResult.Set(m.secretQueryKey, secret) // This overwrites any existing values.
	serviceURL.RawQuery = queryResult.Encode()

	resp := CreateResponse{
		MagicLink: serviceURL,
		Secret:    secret,
	}

	return resp, nil
}

// MagicLinkHandler is an HTTP handler that accepts HTTP requests with magic link secrets, then redirects to the given
// URL with the JWT as a query parameter.
func (m MagicLink) MagicLinkHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		secret := r.URL.Query().Get(m.secretQueryKey)
		if secret == "" {
			m.handleError(ErrMagicLinkMissingSecret, http.StatusBadRequest, r, w)
			return
		}

		if m.customRedirector != nil {
			args := RedirectorArgs{
				ReadAndExpireLink: m.HandleMagicLink,
				Request:           r,
				Secret:            secret,
				Writer:            w,
			}
			m.customRedirector.Redirect(args)
			return
		}

		jwtB64, response, err := m.HandleMagicLink(ctx, secret)
		if err != nil {
			if errors.Is(err, ErrLinkNotFound) {
				m.handleError(err, http.StatusNotFound, r, w)
				return
			}
			m.handleError(err, http.StatusInternalServerError, r, w)
			return
		}
		u := redirectURLFromResponse(response, jwtB64)
		http.Redirect(w, r, u.String(), http.StatusSeeOther)
	})
}

// HandleMagicLink is a method that accepts a magic link secret, then returns the signed JWT.
func (m MagicLink) HandleMagicLink(ctx context.Context, secret string) (jwtB64 string, response ReadResponse, err error) {
	response, err = m.Store.Read(ctx, secret)
	if err != nil {
		if errors.Is(err, ErrLinkNotFound) {
			return "", response, err
		}
		return "", response, ErrMagicLinkRead
	}

	var jwk jwkset.JWK
	if response.CreateArgs.JWTKeyID != nil {
		jwk, err = m.jwks.storage.KeyRead(ctx, *response.CreateArgs.JWTKeyID)
		if err != nil {
			return "", response, fmt.Errorf("%w: %s", ErrJWKSReadGivenKID, err)
		}
	} else {
		allKeys, err := m.jwks.storage.KeyReadAll(ctx)
		if err != nil {
			return "", response, fmt.Errorf("%w: %s", ErrJWKSSnapshot, err)
		}
		if len(allKeys) == 0 {
			return "", response, ErrJWKSEmpty
		}
		jwk = allKeys[0] // TODO Why the first key? Should this be the signing default? If so, update docs and implementations.
	}

	signingMethod := jwt.GetSigningMethod(response.CreateArgs.JWTSigningMethod)
	if signingMethod == nil {
		signingMethod = BestSigningMethod(jwk.Key())
	}

	token := jwt.NewWithClaims(signingMethod, response.CreateArgs.JWTClaims)
	token.Header[jwkset.HeaderKID] = jwk.Marshal().KID
	jwtB64, err = token.SignedString(jwk.Key())
	if err != nil {
		return "", response, fmt.Errorf("%w: %s", ErrJWTSign, err)
	}

	return jwtB64, response, nil
}

func (m MagicLink) handleError(err error, suggestedResponseCode int, request *http.Request, writer http.ResponseWriter) {
	args := ErrorHandlerArgs{
		Err:                   err,
		Request:               request,
		SuggestedResponseCode: suggestedResponseCode,
		Writer:                writer,
	}
	if m.errorHandler != nil {
		m.errorHandler.Handle(args)
		return
	}
	writer.WriteHeader(suggestedResponseCode)
}

// BestSigningMethod returns the best signing method for the given key.
func BestSigningMethod(key any) jwt.SigningMethod {
	var signingMethod jwt.SigningMethod
	switch key := key.(type) {
	case *ecdsa.PrivateKey:
		curve := key.Curve
		signingMethod = signingMethodECDSACurve(curve, signingMethod)
	case *ecdsa.PublicKey:
		curve := key.Curve
		signingMethod = signingMethodECDSACurve(curve, signingMethod)
	case ed25519.PrivateKey, ed25519.PublicKey:
		signingMethod = jwt.SigningMethodEdDSA
	case *rsa.PrivateKey:
		size := key.Size()
		signingMethod = signingMethodRSASize(size)
	case *rsa.PublicKey:
		size := key.Size()
		signingMethod = signingMethodRSASize(size)
	default:
		signingMethod = jwt.SigningMethodHS512
	}
	return signingMethod
}

func redirectURLFromResponse(response ReadResponse, jwtB64 string) *url.URL {
	u := copyURL(response.CreateArgs.RedirectURL)
	query := u.Query()
	queryKey := response.CreateArgs.RedirectQueryKey
	if queryKey == "" {
		queryKey = DefaultRedirectQueryKey
	}
	query.Add(queryKey, jwtB64)
	u.RawQuery = query.Encode()
	return u
}

func signingMethodECDSACurve(curve elliptic.Curve, signingMethod jwt.SigningMethod) jwt.SigningMethod {
	switch curve {
	case elliptic.P256():
		signingMethod = jwt.SigningMethodES256
	case elliptic.P384():
		signingMethod = jwt.SigningMethodES384
	case elliptic.P521():
		signingMethod = jwt.SigningMethodES512
	}
	return signingMethod
}

func signingMethodRSASize(size int) jwt.SigningMethod {
	var signingMethod jwt.SigningMethod
	switch size {
	case 256:
		signingMethod = jwt.SigningMethodRS256
	case 384:
		signingMethod = jwt.SigningMethodRS384
	case 512:
		signingMethod = jwt.SigningMethodRS512
	}
	return signingMethod
}

func copyURL(u *url.URL) *url.URL {
	c, _ := url.Parse(u.String())
	return c
}
