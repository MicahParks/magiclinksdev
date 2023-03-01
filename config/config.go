package config

import (
	"fmt"
	"net/url"
	"time"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
)

// Config is the configuration for the magiclinksdev server.
type Config struct {
	AdminConfig         []model.AdminCreateArgs     `json:"adminConfig"`
	BaseURL             *jt.JSONType[*url.URL]      `json:"baseURL"`
	Iss                 string                      `json:"iss"`
	JWKS                JWKS                        `json:"jwks"`
	RelativeRedirectURL *jt.JSONType[*url.URL]      `json:"relativeRedirectURL"`
	RequestTimeout      *jt.JSONType[time.Duration] `json:"requestTimeout"`
	RequestMaxBodyBytes int64                       `json:"requestMaxBodyBytes"`
	SecretQueryKey      string                      `json:"secretQueryKey"`
	ShutdownTimeout     *jt.JSONType[time.Duration] `json:"shutdownTimeout"`
	Validation          model.Validation            `json:"validation"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (c Config) DefaultsAndValidate() (Config, error) {
	baseURL := c.BaseURL.Get()
	if baseURL == nil {
		return Config{}, fmt.Errorf("base URL is required: %w", jt.ErrDefaultsAndValidate)
	}
	switch baseURL.Scheme {
	case "http", "https":
	default:
		return Config{}, fmt.Errorf("base URL scheme must be http or https: %w", jt.ErrDefaultsAndValidate)
	}
	if baseURL.Host == "" {
		return Config{}, fmt.Errorf("base URL host is required: %w", jt.ErrDefaultsAndValidate)
	}
	if c.RelativeRedirectURL.Get() == nil {
		u, err := url.Parse(mld.DefaultRelativePathRedirect)
		if err != nil {
			return Config{}, fmt.Errorf("failed to parse default relative URL redirect: %w", err)
		}
		c.RelativeRedirectURL = jt.New(u)
	} else {
		u := c.RelativeRedirectURL.Get()
		if u.Scheme != "" || u.Host != "" || u.Path == "" {
			return Config{}, fmt.Errorf("relative URL redirect must be relative: %w", jt.ErrDefaultsAndValidate)
		}
	}
	_, err := baseURL.Parse(c.RelativeRedirectURL.Get().String())
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse base relative URL path for magic links: %w", err)
	}
	if c.Iss == "" {
		return Config{}, fmt.Errorf("iss is required: %w", jt.ErrDefaultsAndValidate)
	}
	if c.RequestTimeout.Get() == 0 {
		c.RequestTimeout = jt.New(5 * time.Second)
	}
	if c.RequestMaxBodyBytes == 0 {
		c.RequestMaxBodyBytes = 1 << 20 // 1 MB.
	}
	if c.SecretQueryKey == "" {
		c.SecretQueryKey = magiclink.DefaultSecretQueryKey
	}
	if c.ShutdownTimeout.Get() == 0 {
		c.ShutdownTimeout = jt.New(time.Second)
	}
	c.Validation, err = c.Validation.DefaultsAndValidate()
	if err != nil {
		return Config{}, fmt.Errorf("failed to validate and apply defaults for validation: %w", err)
	}
	return c, nil
}

// JWKS is the JSON Web Key Set configuration.
type JWKS struct {
	IgnoreDefault bool `json:"ignoreDefault"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (j JWKS) DefaultsAndValidate() (JWKS, error) {
	return j, nil
}
