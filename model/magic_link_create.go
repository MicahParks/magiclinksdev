package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

type MagicLinkCreateParams struct {
	JWTCreateParams  JWTCreateParams `json:"jwtCreateParams"`
	LifespanSeconds  int             `json:"lifespanSeconds"`
	RedirectQueryKey string          `json:"redirectQueryKey"`
	RedirectURL      string          `json:"redirectURL"`
}

func (p MagicLinkCreateParams) Validate(config Validation) (ValidMagicLinkCreateParams, error) {
	validJWTCreateParams, err := p.JWTCreateParams.Validate(config)
	if err != nil {
		return ValidMagicLinkCreateParams{}, fmt.Errorf("failed to validate JWT create args: %w", err)
	}
	lifespan := time.Duration(p.LifespanSeconds) * time.Second
	if lifespan == 0 {
		lifespan = time.Hour
	} else if lifespan < 5*time.Second || lifespan > config.LifeSpanSeconds.Get() {
		return ValidMagicLinkCreateParams{}, fmt.Errorf("%w: link lifespan must be between 5 and %d", ErrInvalidModel, int(config.LifeSpanSeconds.Get().Seconds()))
	}
	if p.RedirectQueryKey == "" {
		p.RedirectQueryKey = magiclink.DefaultRedirectQueryKey
	}
	u, err := httpURL(config, p.RedirectURL)
	if err != nil {
		return ValidMagicLinkCreateParams{}, fmt.Errorf("failed to validate URL: %w", err)
	}
	valid := ValidMagicLinkCreateParams{
		Lifespan:         lifespan,
		JWTCreateParams:  validJWTCreateParams,
		RedirectQueryKey: p.RedirectQueryKey,
		RedirectURL:      u,
	}
	return valid, nil
}

type ValidMagicLinkCreateParams struct {
	Lifespan         time.Duration
	JWTCreateParams  ValidJWTCreateParams
	RedirectQueryKey string
	RedirectURL      *url.URL
}

type MagicLinkCreateRequest struct {
	MagicLinkCreateParams MagicLinkCreateParams `json:"magicLinkCreateParams"`
}

func (b MagicLinkCreateRequest) Validate(config Validation) (ValidMagicLinkCreateRequest, error) {
	magicLinkParams, err := b.MagicLinkCreateParams.Validate(config)
	if err != nil {
		return ValidMagicLinkCreateRequest{}, fmt.Errorf("failed to validate magic link args: %w", err)
	}
	valid := ValidMagicLinkCreateRequest{
		MagicLinkParams: magicLinkParams,
	}
	return valid, nil
}

type ValidMagicLinkCreateRequest struct {
	MagicLinkParams ValidMagicLinkCreateParams
}

type MagicLinkCreateResults struct {
	MagicLink string `json:"magicLink"`
	Secret    string `json:"secret"`
}

type MagicLinkCreateResponse struct {
	MagicLinkCreateResults MagicLinkCreateResults `json:"magicLinkCreateResults"`
	RequestMetadata        RequestMetadata        `json:"requestMetadata"`
}
