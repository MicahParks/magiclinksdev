package magiclink

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v4"

	mld "github.com/MicahParks/magiclinksdev"
)

const (
	// DefaultRedirectQueryKey is the default URL query parameter to contain the JWT in after a magic link has been
	// clicked.
	DefaultRedirectQueryKey = "jwt"
	// DefaultSecretQueryKey is the default URL query parameter to contain the secret for a magic link.
	DefaultSecretQueryKey = "secret"
)

type ReCAPTCHAV3TemplateConfig struct {
	CSS              template.CSS  `json:"css"`
	Code             string        `json:"code"`
	HTMLTitle        string        `json:"htmlTitle"`
	Instruction      string        `json:"instruction"`
	ReCAPTCHASiteKey template.HTML `json:"reCAPTCHASiteKey"`
	Title            string        `json:"title"`
}

type recaptchav3TemplateData struct {
	ButtonSkipsVerification bool
	Config                  ReCAPTCHAV3TemplateConfig
	Redirect                string
}

func (f ReCAPTCHAV3TemplateConfig) DefaultsAndValidate() (ReCAPTCHAV3TemplateConfig, error) {
	if f.CSS == "" {
		f.CSS = template.CSS(defaultCSS)
	}
	if f.Instruction == "" {
		f.Instruction = "Click the below button if this page does not automatically redirect. This page is meant to stop robots from using magic links."
	}
	if f.HTMLTitle == "" {
		f.HTMLTitle = "Magic Link - Browser Check"
	}
	if f.ReCAPTCHASiteKey == "" {
		return f, fmt.Errorf("%w: ReCAPTCHASiteKey is required", jt.ErrDefaultsAndValidate)
	}
	if f.Code == "" {
		f.Code = "BROWSER CHECK"
	}
	if f.Title == "" {
		f.Title = "Checking your browser..."
	}
	return f, nil
}

// MagicLink holds the necessary assets for the magic link service.
type MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta any] struct {
	Store             Storage[CustomCreateArgs, CustomReadResults, CustomKeyMeta]
	errorHandler      ErrorHandler
	tmpl              *template.Template
	jwks              *jwksCache[CustomKeyMeta]
	preventRobotsEnum PreventRobotsEnum
	reCAPTCHAV3Config ReCAPTCHAV3Config
	secretQueryKey    string
	serviceURL        *url.URL
}

// NewMagicLink creates a new MagicLink. The given setupCtx is only used during the creation of the MagicLink.
func NewMagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta any](setupCtx context.Context, config Config[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) (MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta], error) {
	var m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]

	err := config.Valid()
	if err != nil {
		return m, fmt.Errorf("failed to validate config: %w", err)
	}

	htmlTemplate := config.HTMLTemplate
	if htmlTemplate == "" {
		htmlTemplate = recaptchav3Template
	}
	tmpl, err := template.New("").Parse(htmlTemplate)
	if err != nil {
		return m, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	jCache, err := newJWKSCache(setupCtx, config.JWKS)
	if err != nil {
		return m, fmt.Errorf("failed to create JWK Set cache: %w", err)
	}

	secretQueryKey := config.SecretQueryKey
	if secretQueryKey == "" {
		secretQueryKey = DefaultSecretQueryKey
	}

	var store Storage[CustomCreateArgs, CustomReadResults, CustomKeyMeta]
	store = config.Store
	if store == nil {
		store = NewMemoryStorage[CustomCreateArgs, CustomReadResults, CustomKeyMeta]()
	}

	m = MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]{
		Store:             store,
		errorHandler:      config.ErrorHandler,
		tmpl:              tmpl,
		jwks:              jCache,
		preventRobotsEnum: config.PreventRobotsDefault,
		reCAPTCHAV3Config: ReCAPTCHAV3Config{},
		secretQueryKey:    secretQueryKey,
		serviceURL:        config.ServiceURL,
	}

	return m, nil
}

// JWKSHandler is an HTTP handler that responds to requests with the JWK Set as JSON.
func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) JWKSHandler() http.Handler {
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
func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) JWKSet() jwkset.JWKSet[CustomKeyMeta] {
	return m.jwks.jwks
}

// NewLink creates a magic link with the given parameters.
func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) NewLink(ctx context.Context, args CreateArgs[CustomCreateArgs]) (CreateResponse, error) {
	err := args.Valid()
	if err != nil {
		return CreateResponse{}, fmt.Errorf("failed to validate args: %w", err)
	}

	secret, err := m.Store.CreateLink(ctx, args)
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
func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) MagicLinkHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()

		secret := request.URL.Query().Get(m.secretQueryKey)
		if secret == "" {
			m.handleError(ErrMagicLinkMissingSecret, http.StatusBadRequest, request, writer)
			return
		}

		jwtB64, response, err := m.HandleMagicLink(ctx, secret)
		if err != nil {
			if errors.Is(err, ErrLinkNotFound) {
				m.handleError(err, http.StatusNotFound, request, writer)
				return
			}
			m.handleError(err, http.StatusInternalServerError, request, writer)
			return
		}

		u := copyURL(response.CreateArgs.RedirectURL)
		query := u.Query()
		queryKey := response.CreateArgs.RedirectQueryKey
		if queryKey == "" {
			queryKey = DefaultRedirectQueryKey
		}
		query.Add(queryKey, jwtB64)
		u.RawQuery = query.Encode()

		preventRobotsEnum := PreventRobotsEnum("") // TODO Determine if request specified no robot prevention.
		if preventRobotsEnum == PreventRobotsDefault {
			preventRobotsEnum = m.preventRobotsEnum
		}
		switch preventRobotsEnum {
		case PreventRobotsReCAPTCHAV3:
			// Proceed.
		default:
			// TODO Make and handle custom error. Use http.StatusSeeOther as recommended response code? See what other cases do.
			fallthrough
		case PreventRobotsNone:
			http.Redirect(writer, request, u.String(), http.StatusSeeOther)
			return
		}

		tData := recaptchav3TemplateData{
			ButtonSkipsVerification: false, // TODO Get from request.
			Config:                  m.reCAPTCHAV3Config.TemplateConfig,
			Redirect:                u.String(),
		}

		err = m.tmpl.Execute(writer, tData)
		if err != nil {
			m.handleError(fmt.Errorf("failed to execute HTML template: %w", err), http.StatusInternalServerError, request, writer)
			return
		}
	})
}

// HandleMagicLink is a method that accepts a magic link secret, then returns the signed JWT.
func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) HandleMagicLink(ctx context.Context, secret string) (jwtB64 string, response ReadResponse[CustomCreateArgs, CustomReadResults], err error) {
	response, err = m.Store.ReadLink(ctx, secret)
	if err != nil {
		if errors.Is(err, ErrLinkNotFound) {
			return "", response, err
		}
		return "", response, ErrMagicLinkRead
	}

	var meta jwkset.KeyWithMeta[CustomKeyMeta]
	if response.CreateArgs.JWTKeyID != nil {
		meta, err = m.jwks.jwks.Store.ReadKey(ctx, *response.CreateArgs.JWTKeyID)
		if err != nil {
			return "", response, fmt.Errorf("%w: %s", ErrJWKSReadGivenKID, err)
		}
	} else {
		allKeys, err := m.jwks.jwks.Store.SnapshotKeys(ctx)
		if err != nil {
			return "", response, fmt.Errorf("%w: %s", ErrJWKSSnapshot, err)
		}
		if len(allKeys) == 0 {
			return "", response, ErrJWKSEmpty
		}
		meta = allKeys[0]
	}

	signingMethod := jwt.GetSigningMethod(response.CreateArgs.JWTSigningMethod)
	if signingMethod == nil {
		signingMethod = BestSigningMethod(meta.Key)
	}

	token := jwt.NewWithClaims(signingMethod, response.CreateArgs.JWTClaims)
	token.Header[jwkset.HeaderKID] = meta.KeyID
	jwtB64, err = token.SignedString(meta.Key)
	if err != nil {
		return "", response, fmt.Errorf("%w: %s", ErrJWTSign, err)
	}

	return jwtB64, response, nil
}

func (m MagicLink[CustomCreateArgs, CustomReadResults, CustomKeyMeta]) handleError(err error, suggestedResponseCode int, request *http.Request, writer http.ResponseWriter) {
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
func BestSigningMethod(key interface{}) jwt.SigningMethod {
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
