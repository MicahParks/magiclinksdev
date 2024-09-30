package magiclink_test

import (
	"context"
	"crypto"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

const (
	jwksPath      = "/jwks.json"
	magicLinkPath = "/magic-link"
)

type dynamicHandler struct {
	handler http.Handler
}

func (d *dynamicHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	d.handler.ServeHTTP(writer, request)
}

type jwtClaims struct {
	CustomValue1 string `json:"customValue1"`
	CustomValue2 string `json:"customValue2"`
	jwt.RegisteredClaims
}

type setupArgs struct {
	errorHandler     magiclink.ErrorHandler
	jwksGet          bool
	jwksGetDelay     time.Duration
	jwksCacheRefresh time.Duration
	jwksStore        jwkset.Storage
	secretQueryKey   string
}

type createArg struct {
	JWTKeyID         *string
	JWTSigningMethod string
	RedirectQueryKey string
}

type testCase struct {
	createArgs []createArg
	setupParam setupArgs
	name       string
}

func TestTable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	redirectChan := make(chan url.Values, 1)
	appServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		redirectChan <- request.URL.Query()
		writer.WriteHeader(http.StatusOK)
	}))
	defer appServer.Close()

	for _, tc := range makeCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			testCreateCases(ctx, t, appServer, tc.createArgs, redirectChan, tc.setupParam)
		})
	}
}

func testCreateCases(ctx context.Context, t *testing.T, appServer *httptest.Server, createArgs []createArg, redirectChan <-chan url.Values, sParam setupArgs) {
	m, magicServer := magiclinkSetup[any](ctx, t, sParam)
	defer magicServer.Close()

	redirectURL, err := url.Parse(appServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse httptest server URL: %s", err)
	}
	claims := jwtClaims{
		CustomValue1: "value1",
		CustomValue2: "value2",
	}

	if sParam.jwksGet {
		time.Sleep(sParam.jwksGetDelay)
		resp, err := http.Get(magicServer.URL + jwksPath)
		if err != nil {
			t.Fatalf("Failed to get JWKS: %s", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Failed to get JWKS: %d", resp.StatusCode)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read JWKS response body: %s", err)
		}
		if len(body) == 0 {
			t.Fatalf("Failed to get JWKS: empty body")
		}
		_ = resp.Body.Close()
	}

	for _, cParam := range createArgs {
		if cParam.RedirectQueryKey == "" {
			cParam.RedirectQueryKey = magiclink.DefaultRedirectQueryKey
		}
		cP := magiclink.CreateArgs{
			JWTClaims:        claims,
			JWTKeyID:         cParam.JWTKeyID,
			JWTSigningMethod: cParam.JWTSigningMethod,
			RedirectQueryKey: cParam.RedirectQueryKey,
			RedirectURL:      redirectURL,
		}
		createResp, err := m.NewLink(ctx, cP)
		if err != nil {
			t.Fatalf("Failed to create magic link: %s", err)
		}

		resp, err := http.Get(createResp.MagicLink.String())
		if err != nil {
			t.Fatalf("Failed to GET magic link: %s", err)
		}
		if resp.StatusCode != http.StatusOK { // The default HTTP client will follow redirects.
			t.Fatalf("Magic link GET did not return 200 OK: %d", resp.StatusCode)
		}

		redirectQuery := <-redirectChan
		jwtB64 := redirectQuery.Get(cParam.RedirectQueryKey)
		if jwtB64 == "" {
			t.Fatalf("Magic link did not contain JWT")
		}

		resultClaims := jwtClaims{}
		token, err := jwt.ParseWithClaims(jwtB64, &resultClaims, keyfunc(ctx, m.JWKSet()))
		if err != nil {
			t.Fatalf("Failed to parse JWT: %s", err)
		}
		if !token.Valid {
			t.Fatalf("JWT was not valid")
		}
		if resultClaims.CustomValue1 != claims.CustomValue1 || resultClaims.CustomValue2 != claims.CustomValue2 {
			t.Fatalf("JWT claims did not match expected claims")
		}
	}
}

func magiclinkSetup[CustomReadResponse any](ctx context.Context, t *testing.T, args setupArgs) (magiclink.MagicLink[CustomReadResponse], *httptest.Server) {
	dH := &dynamicHandler{}
	server := httptest.NewServer(dH)
	serviceURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse httptest server URL: %s", err)
	}
	serviceURL, err = serviceURL.Parse(magicLinkPath)
	if err != nil {
		t.Fatalf("Failed to parse magic link path: %s", err)
	}

	config := magiclink.Config[CustomReadResponse]{
		ErrorHandler:   args.errorHandler,
		ServiceURL:     serviceURL,
		SecretQueryKey: args.secretQueryKey,
		Store:          nil,
		JWKS: magiclink.JWKSArgs{
			CacheRefresh: args.jwksCacheRefresh,
			Store:        args.jwksStore,
		},
	}
	m, err := magiclink.NewMagicLink(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create MagicLink service: %s", err)
	}
	jwksHandler := m.JWKSHandler()
	magicLinkHandler := m.MagicLinkHandler()
	dH.handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == jwksPath {
			jwksHandler.ServeHTTP(writer, request)
			return
		} else if request.URL.Path == magicLinkPath {
			magicLinkHandler.ServeHTTP(writer, request)
			return
		}
		writer.WriteHeader(http.StatusNotFound)
	})

	return m, server
}

func keyfunc(ctx context.Context, store jwkset.Storage) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		kid, ok := token.Header[jwkset.HeaderKID].(string)
		if !ok {
			return nil, errors.New("failed to parse kid from token header")
		}
		jwk, err := store.KeyRead(ctx, kid)
		if err != nil {
			return nil, err
		}
		type publicKeyer interface {
			Public() crypto.PublicKey
		}
		key := jwk.Key()
		pk, ok := key.(publicKeyer)
		if ok {
			return pk.Public(), nil
		}
		return key, nil
	}
}
