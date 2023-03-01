package model

import (
	"fmt"
	"net/url"
	"time"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

// LinkCreateArgs are the unvalidated arguments for creating a link.
type LinkCreateArgs struct {
	JWTCreateArgs    JWTCreateArgs `json:"jwtCreateArgs"`
	LinkLifespan     int           `json:"linkLifespan"`
	RedirectQueryKey string        `json:"redirectQueryKey"`
	RedirectURL      string        `json:"redirectUrl"`
}

// Validate validates the link create arguments.
func (p LinkCreateArgs) Validate(config Validation) (ValidLinkCreateArgs, error) {
	validJWTCreateArgs, err := p.JWTCreateArgs.Validate(config)
	if err != nil {
		return ValidLinkCreateArgs{}, fmt.Errorf("failed to validate JWT create args: %w", err)
	}
	lifespan := time.Duration(p.LinkLifespan) * time.Second
	if lifespan == 0 {
		lifespan = time.Hour
	} else if lifespan < 5*time.Second || lifespan > config.LinkLifespanMax.Get() {
		return ValidLinkCreateArgs{}, fmt.Errorf("%w: link lifespan must be between 5 and %d", ErrInvalidModel, int(config.LinkLifespanMax.Get().Seconds()))
	}

	if p.RedirectQueryKey == "" {
		p.RedirectQueryKey = magiclink.DefaultRedirectQueryKey
	}
	u, err := httpURL(config, p.RedirectURL)
	if err != nil {
		return ValidLinkCreateArgs{}, fmt.Errorf("failed to validate URL: %w", err)
	}
	valid := ValidLinkCreateArgs{
		LinkLifespan:     lifespan,
		JWTCreateArgs:    validJWTCreateArgs,
		RedirectQueryKey: p.RedirectQueryKey,
		RedirectURL:      u,
	}
	return valid, nil
}

// ValidLinkCreateArgs are the validated arguments for creating a link.
type ValidLinkCreateArgs struct {
	LinkLifespan     time.Duration
	JWTCreateArgs    ValidJWTCreateArgs
	RedirectQueryKey string
	RedirectURL      *url.URL
}

// LinkCreateRequest is the request to create a link.
type LinkCreateRequest struct {
	LinkArgs LinkCreateArgs `json:"linkArgs"`
}

// Validate validates the link create request.
func (b LinkCreateRequest) Validate(config Validation) (ValidLinkCreateRequest, error) {
	linkArgs, err := b.LinkArgs.Validate(config)
	if err != nil {
		return ValidLinkCreateRequest{}, fmt.Errorf("failed to validate link args: %w", err)
	}
	valid := ValidLinkCreateRequest{
		LinkArgs: linkArgs,
	}
	return valid, nil
}

// ValidLinkCreateRequest is the validated request to create a link.
type ValidLinkCreateRequest struct {
	LinkArgs ValidLinkCreateArgs
}

// LinkCreateResults are the results of creating a link.
type LinkCreateResults struct {
	MagicLink string `json:"magicLink"`
	Secret    string `json:"secret"`
}

// LinkCreateResponse is the response to creating a link.
type LinkCreateResponse struct {
	LinkCreateResults LinkCreateResults `json:"linkCreateResults"`
	RequestMetadata   RequestMetadata   `json:"requestMetadata"`
}
