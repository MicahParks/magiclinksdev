package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

const (
	// SaaSBaseURL is the base URL for the SaaS offering. The SaaS offering is optional and the magiclinksdev project
	// can be self-hosted.
	SaaSBaseURL = "https://magiclinks.dev"
	// SaaSIss is the iss claim for JWTs in the SaaS offering.
	SaaSIss = SaaSBaseURL
)

var (
	// ErrClientConfig indicates there is an error with the client configuration.
	ErrClientConfig = errors.New("client config invalid")
	// ErrNotReady indicates the magiclinksdev server deployment is not ready.
	ErrNotReady = errors.New("magiclinksdev server deployment is not ready")
)

// Options are used to configure the Client.
type Options struct {
	DisableKeyfunc bool
	KeyfuncOptions *keyfunc.Options
	HTTP           *http.Client
}

// Client is a client for the magiclinksdev project.
type Client struct {
	apiKey  uuid.UUID
	aud     uuid.UUID
	baseURL *url.URL
	http    *http.Client
	jwtIss  string
	keyf    keyfunc.Keyfunc
}

// New creates a new magiclinksdev client. The apiKey and aud are tied to the service account being used. The baseURL is
// the HTTP(S) location of the magiclinksdev deployment. Only use HTTPS in production. For the SaaS offering, use the
// SaaSBaseURL constant. The iss is the issuer of the JWTs, which is in the configuration of the magiclinksdev
// deployment. For the SaaS offering, use the SaaSIss constant. Providing an empty string for the iss will disable
// issuer validation.
func New(apiKey, aud uuid.UUID, baseURL, iss string, options Options) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return Client{}, fmt.Errorf("failed to parse base URL: %w", err)
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return Client{}, fmt.Errorf("invalid base URL scheme: %w", ErrClientConfig)
	}

	c := Client{
		apiKey:  apiKey,
		aud:     aud,
		baseURL: u,
		jwtIss:  iss,
	}

	c.http = http.DefaultClient
	if options.HTTP != nil {
		c.http = options.HTTP
	}

	if !options.DisableKeyfunc {
		jwksURL, err := u.Parse(network.PathJWKS)
		if err != nil {
			return Client{}, fmt.Errorf("failed to parse JWKS URL: %w", err)
		}
		c.keyf, err = keyfunc.NewDefault([]string{jwksURL.String()})
		if err != nil {
			return Client{}, fmt.Errorf("failed to get JWKS: %w", err)
		}
	}

	return c, nil
}

// LocalJWTValidate validates a JWT locally. If the claims argument is not nil, its value will be passed directly to
// jwt.ParseWithClaims. The claims should be unmarshalled into the provided non-nil pointer after the function call. See
// the documentation for jwt.ParseWithClaims for more information. Registered JWT claims will be validated regardless if
// claims are specified or not.
func (c Client) LocalJWTValidate(token string, claims jwt.Claims) (*jwt.Token, error) {
	if c.keyf == nil {
		return nil, fmt.Errorf("%w: client configuration disabled JWK Set client, keyfunc, please enable keyfunc in magiclinksdev client creation options", ErrClientConfig)
	}

	var registered jwt.RegisteredClaims
	t, err := jwt.ParseWithClaims(token, &registered, c.keyf.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if !t.Valid {
		return nil, fmt.Errorf("%w: invalid JWT", handle.ErrToken)
	}

	if !slices.Contains(registered.Audience, c.aud.String()) {
		return nil, fmt.Errorf("%w: invalid JWT audience, this token is likely signed for another service account under the same mld instance", handle.ErrToken)
	}
	if c.jwtIss != "" {
		if registered.Issuer != c.jwtIss {
			return nil, fmt.Errorf("%w: invalid JWT issuer", handle.ErrToken)
		}
	}

	if claims != nil {
		t, err = jwt.ParseWithClaims(token, claims, c.keyf.Keyfunc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JWT: %w", err)
		}
		if !t.Valid {
			return nil, fmt.Errorf("%w: invalid JWT", handle.ErrToken)
		}
	}

	return t, nil
}

// EmailLinkCreate calls the /email-magic-link/create endpoint and returns the appropriate response.
func (c Client) EmailLinkCreate(ctx context.Context, req model.EmailLinkCreateRequest) (model.EmailLinkCreateResponse, model.Error, error) {
	resp, errResp, err := request[model.EmailLinkCreateRequest, model.EmailLinkCreateResponse](ctx, c, http.StatusCreated, network.PathEmailMagicLinkCreate, req)
	if err != nil {
		return model.EmailLinkCreateResponse{}, errResp, fmt.Errorf("failed to create email link: %w", err)
	}
	return resp, errResp, nil
}

// JWTCreate calls the /jwt/create endpoint and returns the appropriate response.
func (c Client) JWTCreate(ctx context.Context, req model.JWTCreateRequest) (model.JWTCreateResponse, model.Error, error) {
	resp, errResp, err := request[model.JWTCreateRequest, model.JWTCreateResponse](ctx, c, http.StatusCreated, network.PathJWTCreate, req)
	if err != nil {
		return model.JWTCreateResponse{}, errResp, fmt.Errorf("failed to create JWT: %w", err)
	}
	return resp, errResp, nil
}

// JWTValidate calls the /jwt/validate endpoint and returns the appropriate response. In most cases, it would be best to
// use the LocalJWTValidate method instead. The LocalJWTValidate method will use a cached version of the JWK Set, which
// saves a network call.
func (c Client) JWTValidate(ctx context.Context, req model.JWTValidateRequest) (model.JWTValidateResponse, model.Error, error) {
	resp, errResp, err := request[model.JWTValidateRequest, model.JWTValidateResponse](ctx, c, http.StatusOK, network.PathJWTValidate, req)
	if err != nil {
		return model.JWTValidateResponse{}, errResp, fmt.Errorf("failed to validate JWT: %w", err)
	}
	return resp, errResp, nil
}

// LinkCreate calls the /magic-link/create endpoint and returns the appropriate response.
func (c Client) LinkCreate(ctx context.Context, req model.LinkCreateRequest) (model.LinkCreateResponse, model.Error, error) {
	resp, errResp, err := request[model.LinkCreateRequest, model.LinkCreateResponse](ctx, c, http.StatusCreated, network.PathMagicLinkCreate, req)
	if err != nil {
		return model.LinkCreateResponse{}, errResp, fmt.Errorf("failed to create link: %w", err)
	}
	return resp, errResp, nil
}

// Ready calls the /ready endpoint. An error is returned if the service is not ready.
func (c Client) Ready(ctx context.Context) error {
	u, err := c.baseURL.Parse(network.PathReady)
	if err != nil {
		return fmt.Errorf("failed to parse ready URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create ready request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get HTTP ready response: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status code %d", ErrNotReady, resp.StatusCode)
	}

	return nil
}

// ServiceAccountCreate calls the /admin/service-account/create endpoint and returns the appropriate response.
func (c Client) ServiceAccountCreate(ctx context.Context, req model.ServiceAccountCreateRequest) (model.ServiceAccountCreateResponse, model.Error, error) {
	resp, errResp, err := request[model.ServiceAccountCreateRequest, model.ServiceAccountCreateResponse](ctx, c, http.StatusCreated, network.PathServiceAccountCreate, req)
	if err != nil {
		return model.ServiceAccountCreateResponse{}, errResp, fmt.Errorf("failed to create service account: %w", err)
	}
	return resp, errResp, nil
}

func request[Req, Resp any](ctx context.Context, c Client, goodStatus int, relPath string, req Req) (Resp, model.Error, error) {
	var resp Resp

	data, err := json.Marshal(req)
	if err != nil {
		return resp, model.Error{}, fmt.Errorf("failed to JSON marshal body: %w", err)
	}
	b := bytes.NewBuffer(data)

	u, err := c.baseURL.Parse(relPath)
	if err != nil {
		return resp, model.Error{}, fmt.Errorf("failed to parse relative path: %w", err)
	}

	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), b)
	if err != nil {
		return resp, model.Error{}, fmt.Errorf("failed to create new request: %w", err)
	}
	hReq.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)
	hReq.Header.Set(middleware.APIKeyHeader, c.apiKey.String())

	hResp, err := c.http.Do(hReq)
	if err != nil {
		return resp, model.Error{}, fmt.Errorf("failed to do request: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer hResp.Body.Close()

	if hResp.StatusCode != goodStatus {
		var errResp model.Error
		err = json.NewDecoder(hResp.Body).Decode(&errResp)
		if err != nil {
			return resp, model.Error{}, fmt.Errorf("failed to decode error response: %w", err)
		}
		return resp, errResp, fmt.Errorf("unexpected response code: %d: message: %q", hResp.StatusCode, errResp.Message)
	}

	err = json.NewDecoder(hResp.Body).Decode(&resp)
	if err != nil {
		return resp, model.Error{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return resp, model.Error{}, nil
}
