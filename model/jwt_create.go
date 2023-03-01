package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// JWTCreateArgs are the unvalidated arguments for creating a JWT.
type JWTCreateArgs struct {
	JWTClaims          any `json:"jwtClaims"`
	JWTLifespanSeconds int `json:"jwtLifespanSeconds"`
}

// Validate implements the Validatable interface.
func (j JWTCreateArgs) Validate(config Validation) (ValidJWTCreateArgs, error) {
	marshaled, err := json.Marshal(j.JWTClaims)
	if err != nil {
		return ValidJWTCreateArgs{}, fmt.Errorf("failed to JSON marshal claims: %w", err)
	}
	lifespan := time.Duration(j.JWTLifespanSeconds) * time.Second
	if lifespan == 0 {
		lifespan = 5 * time.Minute
	} else if lifespan < 5*time.Second || lifespan > config.JWTLifespanMax.Get() {
		return ValidJWTCreateArgs{}, fmt.Errorf("%w: JWT lifespan seconds must be between 5 seconds and %d", ErrInvalidModel, int(config.JWTLifespanMax.Get().Seconds()))
	}
	if uint(len(marshaled)) > config.JWTClaimsMaxBytes {
		return ValidJWTCreateArgs{}, fmt.Errorf("%w: JWT claims must be less than %d bytes", ErrInvalidModel, config.JWTClaimsMaxBytes)
	}
	valid := ValidJWTCreateArgs{
		JWTClaims:   marshaled,
		JWTLifespan: lifespan,
	}
	return valid, nil
}

// ValidJWTCreateArgs are the validated arguments for creating a JWT.
type ValidJWTCreateArgs struct {
	JWTClaims   json.RawMessage
	JWTLifespan time.Duration
}

// JWTCreateRequest is the unvalidated request to create a JWT.
type JWTCreateRequest struct {
	JWTCreateArgs JWTCreateArgs `json:"jwtCreateArgs"`
}

// Validate implements the Validatable interface.
func (j JWTCreateRequest) Validate(config Validation) (ValidJWTCreateRequest, error) {
	valid, err := j.JWTCreateArgs.Validate(config)
	if err != nil {
		return ValidJWTCreateRequest{}, fmt.Errorf("failed to validate JWT create args: %w", err)
	}
	return ValidJWTCreateRequest{
		JWTCreateArgs: valid,
	}, nil
}

// ValidJWTCreateRequest is the validated request to create a JWT.
type ValidJWTCreateRequest struct {
	JWTCreateArgs ValidJWTCreateArgs
}

// JWTCreateResults are the results of creating a JWT.
type JWTCreateResults struct {
	JWT string `json:"jwt"`
}

// JWTCreateResponse is the response to creating a JWT.
type JWTCreateResponse struct {
	JWTCreateResults JWTCreateResults `json:"jwtCreateResults"`
	RequestMetadata  RequestMetadata  `json:"requestMetadata"`
}
