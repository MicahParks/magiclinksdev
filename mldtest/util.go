package mldtest

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// BaseURL is the test base URL for the service.
	BaseURL = "http://localhost:8080/api/v2/"
	// Iss is the test issuer for the service.
	Iss = BaseURL
	// LogoImageURL is the test service logo for the service.
	LogoImageURL = "http://example.com/logo.png"
	// ServiceName is the test service name for the service.
	ServiceName = "Example service"
	// LogoClickURL is the test service URL for the service.
	LogoClickURL = "http://example.com"
	// LinksExpireAfter is the amount of time links expire after for tests.
	LinksExpireAfter = 24 * 30 * time.Hour
)

var (
	// APIKey is the test API key for the service.
	APIKey = uuid.MustParse("40084740-0bc3-455d-b298-e23a31561580")
	// Aud is the test audience for the service.
	Aud = uuid.MustParse("ad9e9d84-92ea-4f07-bac9-5d898d59c83b")
	// ErrMLDTest is the test error for the service.
	ErrMLDTest = errors.New("mldtest")
	// TClaims is the test claims for the service.
	TClaims = TestClaims{Foo: "foo"}
	// SAUUID is the test service account UUID for the service.
	SAUUID = uuid.MustParse("1e079d6d-a8b9-4065-aa8d-86906accd211")
)

// TestClaims is the test claims for the service.
type TestClaims struct {
	Foo string `json:"foo"`
}

func (t TestClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return nil, nil
}
func (t TestClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}
func (t TestClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}
func (t TestClaims) GetIssuer() (string, error) {
	return "", nil
}
func (t TestClaims) GetSubject() (string, error) {
	return "", nil
}
func (t TestClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// Equal returns true if the two claims are equal.
func (t TestClaims) Equal(c TestClaims) bool {
	return t.Foo == c.Foo
}
