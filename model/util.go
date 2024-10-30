package model

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"
	"unicode/utf8"

	jt "github.com/MicahParks/jsontype"
	"github.com/google/uuid"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

// ErrInvalidModel is returned when a model is invalid.
var ErrInvalidModel = errors.New("invalid model")

// Error is the model for an error.
type Error struct {
	Code            int             `json:"code"`
	Message         string          `json:"message"`
	RequestMetadata RequestMetadata `json:"requestMetadata"`
}

// NewError creates a new error.
func NewError(ctx context.Context, code int, message string) Error {
	return Error{
		Code:    code,
		Message: message,
		RequestMetadata: RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}
}

// RequestMetadata is the model for request metadata.
type RequestMetadata struct {
	UUID uuid.UUID `json:"uuid"`
}

// ServiceAccount is the model for a service account and its metadata.
type ServiceAccount struct {
	UUID   uuid.UUID `json:"uuid"`
	APIKey uuid.UUID `json:"apiKey"`
	Aud    uuid.UUID `json:"aud"`
	Admin  bool      `json:"admin"`
}

// Validation contains information on how to validate models.
type Validation struct {
	LinkLifespanDefault *jt.JSONType[time.Duration] `json:"linkLifespanDefault"`
	LifeSpanSeconds     *jt.JSONType[time.Duration] `json:"maxLinkLifespan"`
	JWTClaimsMaxBytes   uint                        `json:"maxJWTClaimsBytes"`
	JWTLifespanDefault  *jt.JSONType[time.Duration] `json:"jwtLifespanDefault"`
	JWTLifespanMax      *jt.JSONType[time.Duration] `json:"maxJWTLifespan"`
	ServiceNameMinUTF8  uint                        `json:"serviceNameMinUTF8"`
	ServiceNameMaxUTF8  uint                        `json:"serviceNameMaxUTF8"`
	URLMaxLength        uint                        `json:"urlMaxLength"`
}

func (v Validation) DefaultsAndValidate() (Validation, error) {
	if v.LinkLifespanDefault.Get() == 0 {
		v.LinkLifespanDefault = jt.New(time.Hour)
	}
	if v.LifeSpanSeconds.Get() == 0 {
		v.LifeSpanSeconds = jt.New(mld.Over250Years)
	}
	if v.JWTClaimsMaxBytes == 0 {
		v.JWTClaimsMaxBytes = 4096
	}
	if v.JWTLifespanDefault.Get() == 0 {
		v.JWTLifespanDefault = jt.New(5 * time.Minute)
	}
	if v.JWTLifespanMax.Get() == 0 {
		v.JWTLifespanMax = jt.New(mld.Over250Years)
	}
	if v.ServiceNameMinUTF8 == 0 {
		v.ServiceNameMinUTF8 = 5
	}
	if v.ServiceNameMaxUTF8 == 0 {
		v.ServiceNameMaxUTF8 = 256
	}
	if v.URLMaxLength == 0 {
		v.URLMaxLength = 2048
	}
	return v, nil
}

func httpURL(config Validation, raw string) (*url.URL, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service URL: %w", err)
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return nil, fmt.Errorf("service logo URL scheme must be http or https: %w", jt.ErrDefaultsAndValidate)
	}
	runeCount := uint(utf8.RuneCountInString(u.String()))
	if runeCount > config.URLMaxLength {
		return nil, fmt.Errorf("service logo URL must be less than or equal to %d runes: %w", config.URLMaxLength, jt.ErrDefaultsAndValidate)
	}
	return u, nil
}
