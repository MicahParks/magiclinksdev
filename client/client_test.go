package client

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/mldtest"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/setup"
)

const (
	keyID = "keyID"
)

var (
	baseURL, _        = url.Parse(mldtest.BaseURL)
	_, signingKey, _  = ed25519.GenerateKey(nil)
	_, attackerKey, _ = ed25519.GenerateKey(nil)
	redirectPath, _   = url.Parse(mld.DefaultRelativePathRedirect)

	jwtCreateArgs = model.JWTCreateArgs{
		Claims:          mldtest.TClaims,
		LifespanSeconds: 5,
	}
	linkArgs = model.MagicLinkCreateArgs{
		JWTCreateArgs:    jwtCreateArgs,
		LifespanSeconds:  5,
		RedirectQueryKey: magiclink.DefaultRedirectQueryKey,
		RedirectURL:      "http://example.com",
	}
)

func TestNew(t *testing.T) {
	tc := []struct {
		name            string
		baseURL         string
		disabledKeyfunc bool
		errIs           error
		errAs           any
		keyfuncOpts     *keyfunc.Options
	}{
		{
			name:            "Valid",
			baseURL:         SaaSBaseURL,
			disabledKeyfunc: true,
			errIs:           nil,
		},
		{
			name:            "Invalid URL",
			baseURL:         "://",
			disabledKeyfunc: true,
			errAs:           &url.Error{},
		},
		{
			name:            "Invalid URL Scheme",
			baseURL:         "",
			disabledKeyfunc: true,
			errIs:           ErrClientConfig,
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			opt := Options{
				DisableKeyfunc: c.disabledKeyfunc,
				KeyfuncOptions: c.keyfuncOpts,
				HTTP:           http.DefaultClient,
			}
			_, err := New(mldtest.APIKey, mldtest.Aud, c.baseURL, mldtest.Iss, opt)
			if c.errIs != nil && !errors.Is(err, c.errIs) {
				t.Fatalf("Should have failed to create client with given error: %v.", err)
			}
			if c.errAs != nil && !errors.As(err, &c.errAs) {
				t.Fatalf("Should have failed to create client with given error: %v.", err)
			}
		})
	}
}

func TestEmailLinkCreate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)

	req := model.MagicLinkEmailCreateRequest{
		MagicLinkEmailCreateArgs: model.MagicLinkEmailCreateArgs{
			ButtonText:   "Test button text",
			Greeting:     "Test greeting",
			LogoClickURL: mldtest.LogoClickURL,
			LogoImageURL: mldtest.LogoImageURL,
			ServiceName:  mldtest.ServiceName,
			Subject:      "Test subject",
			SubTitle:     "Test subtitle",
			Title:        "Test title",
			ToEmail:      "customer@example.com",
			ToName:       "Test name",
		},
		MagicLinkCreateArgs: linkArgs,
	}
	resp, mldErr, err := c.EmailLinkCreate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create email link: %v.", err)
	}
	if mldErr.Code != 0 {
		t.Fatalf("Failed to create email link: %v.", mldErr)
	}

	validateMetadata(t, resp.RequestMetadata)
	validateLinkResults(t, resp.MagicLinkEmailCreateResults.MagicLinkCreateResults)
}

func TestJWTCreate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)
	jwtCreateHelper(ctx, t, c)
}

func TestLocalJWTValidate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)
	raw := jwtCreateHelper(ctx, t, c)

	var claims mldtest.TestClaims
	token, err := c.LocalJWTValidate(raw, &claims)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v.", err)
	}

	if !token.Valid {
		t.Fatalf("JWT should be valid.")
	}

	if !claims.Equal(mldtest.TClaims) {
		t.Fatalf("JWT claims should be equal.")
	}
}

func TestLocalJWTValidateForged(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, mldtest.TClaims)
	token.Header[jwkset.HeaderKID] = keyID
	raw, err := token.SignedString(attackerKey)
	if err != nil {
		t.Fatalf("Failed to sign attacker JWT: %v.", err)
	}

	var claims mldtest.TestClaims
	_, err = c.LocalJWTValidate(raw, &claims)
	if err == nil {
		t.Fatalf("Should return error when JWT is invalid: %v.", err)
	}
}

func TestJWTValidate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)
	raw := jwtCreateHelper(ctx, t, c)

	req := model.JWTValidateRequest{
		JWTValidateArgs: model.JWTValidateArgs{
			JWT: raw,
		},
	}
	resp, mldErr, err := c.JWTValidate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v.", err)
	}

	if mldErr.Code != 0 {
		t.Fatalf("Failed to validate JWT. API error: %#v.", mldErr)
	}

	validateMetadata(t, resp.RequestMetadata)

	var claims mldtest.TestClaims
	err = json.Unmarshal(resp.JWTValidateResults.JWTClaims, &claims)
	if err != nil {
		t.Fatalf("Failed to unmarshal JWT claims: %v.", err)
	}

	if !claims.Equal(mldtest.TClaims) {
		t.Fatalf("JWT claims should be equal.")
	}
}

func TestJWTValidateForged(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, mldtest.TClaims)
	token.Header[jwkset.HeaderKID] = keyID
	raw, err := token.SignedString(attackerKey)
	if err != nil {
		t.Fatalf("Failed to sign attacker JWT: %v.", err)
	}

	req := model.JWTValidateRequest{
		JWTValidateArgs: model.JWTValidateArgs{
			JWT: raw,
		},
	}
	_, mldErr, err := c.JWTValidate(ctx, req)
	if err == nil {
		t.Fatalf("Should return error when JWT is invalid: %v.", err)
	}

	validateMetadata(t, mldErr.RequestMetadata)

	if mldErr.Code != 422 {
		t.Fatalf("JWT signed by attacker key, response should have 422 status: %#v.", mldErr)
	}
}

func TestLinkCreate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)

	req := model.MagicLinkCreateRequest{
		MagicLinkCreateArgs: linkArgs,
	}
	resp, mldErr, err := c.LinkCreate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create magic link: %v.", err)
	}
	if mldErr.Code != 0 {
		t.Fatalf("Failed to create magic link. API error: %#v.", mldErr)
	}

	validateMetadata(t, resp.RequestMetadata)
	validateLinkResults(t, resp.MagicLinkCreateResults)
}

func TestServiceAccountCreate(t *testing.T) {
	ctx := createCtx(t)
	c := newClient(ctx, t)

	req := model.ServiceAccountCreateRequest{
		ServiceAccountCreateArgs: model.ServiceAccountCreateArgs{},
	}
	resp, mldErr, err := c.ServiceAccountCreate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create service account: %v.", err)
	}
	if mldErr.Code != 0 {
		t.Fatalf("Failed to create service account. API error: %#v.", mldErr)
	}

	validateMetadata(t, resp.RequestMetadata)

	if resp.ServiceAccountCreateResults.ServiceAccount.Admin {
		t.Fatalf("Created service account should not be an admin.")
	}
	if resp.ServiceAccountCreateResults.ServiceAccount.UUID == uuid.Nil {
		t.Fatalf("Created service account should have non-nil UUID.")
	}
	if resp.ServiceAccountCreateResults.ServiceAccount.Aud == uuid.Nil {
		t.Fatalf("Created service account should have non-nil audience UUID.")
	}
	if resp.ServiceAccountCreateResults.ServiceAccount.APIKey == uuid.Nil {
		t.Fatalf("Created service account should have non-nil API key.")
	}
}

func createCtx(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	return ctx
}

func jwtCreateHelper(ctx context.Context, t *testing.T, c Client) string {
	req := model.JWTCreateRequest{
		JWTCreateArgs: jwtCreateArgs,
	}
	resp, mldErr, err := c.JWTCreate(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create JWT: %v.", err)
	}
	if mldErr.Code != 0 {
		t.Fatalf("Failed to create JWT. API error: %#v.", mldErr)
	}

	validateMetadata(t, resp.RequestMetadata)

	if resp.JWTCreateResults.JWT == "" {
		t.Fatalf("JWT should not be empty.")
	}

	claims := mldtest.TestClaims{}
	token, err := jwt.ParseWithClaims(resp.JWTCreateResults.JWT, &claims, c.keyf.Keyfunc)
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v.", err)
	}
	if !token.Valid {
		t.Fatalf("JWT should be valid.")
	}
	if !claims.Equal(mldtest.TClaims) {
		t.Fatalf("JWT should have correct claims.")
	}

	return resp.JWTCreateResults.JWT
}

func newClient(ctx context.Context, t *testing.T) Client {
	conf := config.Config{
		AdminConfig: []model.AdminCreateArgs{
			{
				APIKey:                   mldtest.APIKey,
				Aud:                      mldtest.Aud,
				UUID:                     mldtest.SAUUID,
				ServiceAccountCreateArgs: model.ServiceAccountCreateArgs{},
			},
		},
		BaseURL: jt.New(baseURL),
		Iss:     mldtest.Iss,
		JWKS: config.JWKS{
			IgnoreDefault: true,
		},
		RelativeRedirectURL: jt.New(redirectPath),
		RequestTimeout:      jt.New(time.Second),
		RequestMaxBodyBytes: 0,
		SecretQueryKey:      magiclink.DefaultSecretQueryKey,
		ShutdownTimeout:     jt.New(time.Second),
		Validation:          model.Validation{},
	}

	sa := model.ServiceAccount{
		UUID:   mldtest.SAUUID,
		APIKey: mldtest.APIKey,
		Aud:    mldtest.Aud,
		Admin:  true,
	}
	testStorageOptions := mldtest.TestStorageOptions{
		Key:   signingKey,
		KeyID: keyID,
		SA:    sa,
	}
	interfaces := setup.ServerInterfaces{
		EmailProvider: mldtest.NopProvider{},
		RateLimiter:   mldtest.NopLimiter{},
		Store:         mldtest.NewTestStorage(testStorageOptions),
	}
	opt := setup.ServerOptions{
		HTTPMux:               nil,
		MagicLinkErrorHandler: nil,
		MiddlewareHook:        nil,
		Logger:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	conf, err := conf.DefaultsAndValidate()
	if err != nil {
		t.Fatalf("Failed to validate config: %v.", err)
	}
	s, err := setup.CreateServer(ctx, conf, opt, interfaces)
	if err != nil {
		t.Fatalf("Failed to create server: %v.", err)
	}
	mux, err := network.CreateHTTPHandlers(s)
	if err != nil {
		t.Fatalf("Failed to create HTTP handlers: %v.", err)
	}
	server := httptest.NewServer(mux)

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v.", err)
	}
	serverURL, err = serverURL.Parse(conf.BaseURL.Get().Path)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v.", err)
	}

	c, err := New(mldtest.APIKey, mldtest.Aud, serverURL.String(), mldtest.Iss, Options{
		DisableKeyfunc: false,
		KeyfuncOptions: nil,
		HTTP:           server.Client(),
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v.", err)
	}

	err = c.Ready(ctx)
	if err != nil {
		t.Fatalf("Failed to wait for client to be ready: %v.", err)
	}

	return c
}

func validateLinkResults(t *testing.T, results model.MagicLinkCreateResults) {
	ml, err := url.Parse(results.MagicLink)
	if err != nil {
		t.Fatalf("Failed to parse magic link: %v.", err)
	}

	switch ml.Scheme {
	case "http", "https":
	default:
		t.Fatalf("Magic link should have http or https scheme.")
	}
	if !strings.HasPrefix(ml.String(), baseURL.String()) {
		t.Fatalf("Magic link should have base URL prefix.")
	}

	secret := ml.Query().Get(magiclink.DefaultSecretQueryKey)
	u, err := uuid.Parse(secret)
	if err != nil {
		t.Fatalf("Failed to parse secret in URL query as UUID: %v.", err)
	}
	if u == uuid.Nil {
		t.Fatalf("Secret should not be nil.")
	}
}

func validateMetadata(t *testing.T, meta model.RequestMetadata) {
	if meta.UUID == uuid.Nil {
		t.Fatalf("Request metadata UUID should not be empty.")
	}
}
