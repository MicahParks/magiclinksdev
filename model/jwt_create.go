package model

import (
	"encoding/json"
	"fmt"
	"time"
)

type JWTCreateParams struct {
	Alg             string `json:"alg"`
	Claims          any    `json:"claims"`
	LifespanSeconds int    `json:"lifespanSeconds"`
}

func (j JWTCreateParams) Validate(config Validation) (ValidJWTCreateParams, error) {
	marshaled, err := json.Marshal(j.Claims)
	if err != nil {
		return ValidJWTCreateParams{}, fmt.Errorf("failed to JSON marshal claims: %w", err)
	}
	lifespan := time.Duration(j.LifespanSeconds) * time.Second
	if lifespan == 0 {
		lifespan = 5 * time.Minute
	} else if lifespan < 5*time.Second || lifespan > config.JWTLifespanMax.Get() {
		return ValidJWTCreateParams{}, fmt.Errorf("%w: JWT lifespan seconds must be between 5 seconds and %d", ErrInvalidModel, int(config.JWTLifespanMax.Get().Seconds()))
	}
	if uint(len(marshaled)) > config.JWTClaimsMaxBytes {
		return ValidJWTCreateParams{}, fmt.Errorf("%w: JWT claims must be less than %d bytes", ErrInvalidModel, config.JWTClaimsMaxBytes)
	}
	valid := ValidJWTCreateParams{
		Alg:      j.Alg,
		Claims:   marshaled,
		Lifespan: lifespan,
	}
	return valid, nil
}

type ValidJWTCreateParams struct {
	Alg      string
	Claims   json.RawMessage
	Lifespan time.Duration
}

type JWTCreateRequest struct {
	JWTCreateParams JWTCreateParams `json:"jwtCreateParams"`
}

func (j JWTCreateRequest) Validate(config Validation) (ValidJWTCreateRequest, error) {
	valid, err := j.JWTCreateParams.Validate(config)
	if err != nil {
		return ValidJWTCreateRequest{}, fmt.Errorf("failed to validate JWT create args: %w", err)
	}
	return ValidJWTCreateRequest{
		JWTCreateParams: valid,
	}, nil
}

type ValidJWTCreateRequest struct {
	JWTCreateParams ValidJWTCreateParams
}

type JWTCreateResults struct {
	JWT string `json:"jwt"`
}

type JWTCreateResponse struct {
	JWTCreateResults JWTCreateResults `json:"jwtCreateResults"`
	RequestMetadata  RequestMetadata  `json:"requestMetadata"`
}
