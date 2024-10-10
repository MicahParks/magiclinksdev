package model

import (
	"encoding/json"
	"fmt"
)

// JWTValidateParams are the unvalidated parameters for validating a JWT.
type JWTValidateParams struct {
	JWT string `json:"jwt"`
}

// Validate implements the Validatable interface.
func (j JWTValidateParams) Validate(_ Validation) (ValidJWTValidateParams, error) {
	valid := ValidJWTValidateParams(j)
	return valid, nil
}

// ValidJWTValidateParams are the validated parameters for validating a JWT.
type ValidJWTValidateParams struct {
	JWT string
}

// JWTValidateRequest is the unvalidated request to validate a JWT.
type JWTValidateRequest struct {
	JWTValidateParams JWTValidateParams `json:"jwtValidateParams"`
}

// Validate implements the Validatable interface.
func (j JWTValidateRequest) Validate(config Validation) (ValidJWTValidateRequest, error) {
	valid, err := j.JWTValidateParams.Validate(config)
	if err != nil {
		return ValidJWTValidateRequest{}, fmt.Errorf("failed to validate JWT validate args: %w", err)
	}
	return ValidJWTValidateRequest{
		JWTValidateParams: valid,
	}, nil
}

// ValidJWTValidateRequest is the validated request to validate a JWT.
type ValidJWTValidateRequest struct {
	JWTValidateParams ValidJWTValidateParams
}

// JWTValidateResults are the results of validating a JWT.
type JWTValidateResults struct {
	JWTClaims json.RawMessage `json:"claims"`
}

// JWTValidateResponse is the response to validating a JWT.
type JWTValidateResponse struct {
	JWTValidateResults JWTValidateResults `json:"jwtValidateResults"`
	RequestMetadata    RequestMetadata    `json:"requestMetadata"`
}
