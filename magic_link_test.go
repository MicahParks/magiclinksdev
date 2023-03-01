package magiclinksdev_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

type testClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

func TestMagicLink(t *testing.T) {
	const customRedirectQueryKey = "customRedirectQueryKey"

	reqBody := model.LinkCreateRequest{
		LinkArgs: model.LinkCreateArgs{
			JWTCreateArgs: model.JWTCreateArgs{
				JWTClaims:          map[string]string{"foo": "bar"},
				JWTLifespanSeconds: 0,
			},
			LinkLifespan:     0,
			RedirectQueryKey: customRedirectQueryKey,
			RedirectURL:      "https://github.com/MicahParks/magiclinksdev",
		},
	}
	marshaled, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	recorder := httptest.NewRecorder()
	u, err := assets.conf.Server.BaseURL.Get().Parse(network.PathLinkCreate)
	if err != nil {
		t.Fatalf("Failed to parse URL: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, u.Path, bytes.NewReader(marshaled))
	req.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)
	req.Header.Set(middleware.APIKeyHeader, assets.sa.APIKey.String())
	assets.mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Received non-200 status code: %d\n%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Header().Get(mld.HeaderContentType) != mld.ContentTypeJSON {
		t.Fatalf("Received non-JSON content type: %s", recorder.Header().Get(mld.HeaderContentType))
	}

	var linkCreateResponse model.LinkCreateResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &linkCreateResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, linkCreateResponse.LinkCreateResults.MagicLink, nil)
	reqSent := time.Now()
	assets.mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("Expected status code %d, got %d", http.StatusSeeOther, recorder.Code)
	}

	redirectURL, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("Failed to parse redirect URL in header: %v", err)
	}

	jwtB64 := redirectURL.Query().Get(customRedirectQueryKey)
	if jwtB64 == "" {
		t.Fatalf("Expected JWT in redirect URL query, got none")
	}

	claims := &testClaims{}
	token, err := jwt.ParseWithClaims(jwtB64, claims, func(token *jwt.Token) (interface{}, error) {
		for _, key := range assets.keys {
			if key.Custom.SigningDefault {
				switch k := key.Key.(type) {
				case ed25519.PrivateKey:
					return k.Public(), nil
				default:
					panic("unexpected default signing key")
				}
			}
		}
		panic("no default signing key")
	})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}
	if !token.Valid {
		t.Fatalf("JWT is not valid")
	}

	if claims.Issuer != assets.conf.Server.Iss {
		t.Fatalf("Expected issuer %q, got %q", assets.conf.Server.Iss, claims.Issuer)
	}
	if claims.Subject != "" {
		t.Fatalf("Expected subject %q, got %q", "", claims.Subject)
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != assets.sa.Aud.String() {
		t.Fatalf("Expected audience %q, got %q", assets.sa.Aud.String(), claims.Audience)
	}
	checkTime(t, claims.ExpiresAt.Time, reqSent.Add(assets.conf.Server.Validation.JWTLifespanDefault.Get()))
	checkTime(t, claims.NotBefore.Time, reqSent)
	checkTime(t, claims.IssuedAt.Time, reqSent)
	confirmUUID(t, claims.ID)

	if claims.Foo != "bar" {
		t.Fatalf("Expected foo to be bar, got %q", claims.Foo)
	}

	redirectURL.RawQuery = ""
	if redirectURL.String() != reqBody.LinkArgs.RedirectURL {
		t.Fatalf("Expected redirect URL %q, got %q", reqBody.LinkArgs.RedirectURL, redirectURL.String())
	}

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, linkCreateResponse.LinkCreateResults.MagicLink, nil)
	assets.mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("Expected status code %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func checkTime(t *testing.T, actual, expected time.Time) {
	const leeway = time.Millisecond
	if actual.Sub(expected) > leeway {
		t.Fatalf("Expected time %q, got %q", expected, actual)
	}
}

func confirmUUID(t *testing.T, u string) {
	_, err := uuid.Parse(u)
	if err != nil {
		t.Fatalf("Failed to parse UUID: %v", err)
	}
}
