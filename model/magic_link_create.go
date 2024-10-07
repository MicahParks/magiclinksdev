package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

// MagicLinkCreateArgs are the unvalidated arguments for creating a magic link.
type MagicLinkCreateArgs struct {
	JWTCreateArgs    JWTCreateArgs `json:"jwtCreateArgs"`
	LifespanSeconds  int           `json:"lifespanSeconds"` // TODO Move into link limits.
	RedirectQueryKey string        `json:"redirectQueryKey"`
	RedirectURL      string        `json:"redirectURL"`
}

// Validate validates the link create arguments.
func (p MagicLinkCreateArgs) Validate(config Validation) (ValidMagicLinkCreateArgs, error) {
	validJWTCreateArgs, err := p.JWTCreateArgs.Validate(config)
	if err != nil {
		return ValidMagicLinkCreateArgs{}, fmt.Errorf("failed to validate JWT create args: %w", err)
	}
	lifespan := time.Duration(p.LifespanSeconds) * time.Second
	if lifespan == 0 {
		lifespan = time.Hour
	} else if lifespan < 5*time.Second || lifespan > config.LifeSpanSeconds.Get() {
		return ValidMagicLinkCreateArgs{}, fmt.Errorf("%w: link lifespan must be between 5 and %d", ErrInvalidModel, int(config.LifeSpanSeconds.Get().Seconds()))
	}

	if p.RedirectQueryKey == "" {
		p.RedirectQueryKey = magiclink.DefaultRedirectQueryKey
	}
	u, err := httpURL(config, p.RedirectURL)
	if err != nil {
		return ValidMagicLinkCreateArgs{}, fmt.Errorf("failed to validate URL: %w", err)
	}
	valid := ValidMagicLinkCreateArgs{
		LinkLifespan:     lifespan,
		JWTCreateArgs:    validJWTCreateArgs,
		RedirectQueryKey: p.RedirectQueryKey,
		RedirectURL:      u,
	}
	return valid, nil
}

// ValidMagicLinkCreateArgs are the validated arguments for creating a magic link.
type ValidMagicLinkCreateArgs struct {
	LinkLifespan     time.Duration
	JWTCreateArgs    ValidJWTCreateArgs
	RedirectQueryKey string
	RedirectURL      *url.URL
}

// MagicLinkCreateRequest is the request to create a magic link.
type MagicLinkCreateRequest struct {
	MagicLinkCreateArgs MagicLinkCreateArgs `json:"magicLinkCreateArgs"`
}

// Validate validates the link create request.
func (b MagicLinkCreateRequest) Validate(config Validation) (ValidMagicLinkCreateRequest, error) {
	magicLinkArgs, err := b.MagicLinkCreateArgs.Validate(config)
	if err != nil {
		return ValidMagicLinkCreateRequest{}, fmt.Errorf("failed to validate magic link args: %w", err)
	}
	valid := ValidMagicLinkCreateRequest{
		MagicLinkArgs: magicLinkArgs,
	}
	return valid, nil
}

// ValidMagicLinkCreateRequest is the validated request to create a magic link.
type ValidMagicLinkCreateRequest struct {
	MagicLinkArgs ValidMagicLinkCreateArgs
}

// MagicLinkCreateResults are the results of creating a magic link.
type MagicLinkCreateResults struct {
	MagicLink string `json:"magicLink"`
	Secret    string `json:"secret"`
}

// MagicLinkCreateResponse is the response to creating a magic link.
type MagicLinkCreateResponse struct {
	MagicLinkCreateResults MagicLinkCreateResults `json:"magicLinkCreateResults"`
	RequestMetadata        RequestMetadata        `json:"requestMetadata"`
}
