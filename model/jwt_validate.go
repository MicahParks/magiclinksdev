package model

import (
	"encoding/json"
	"fmt"
)

// JWTValidateArgs are the unvalidated arguments for validating a JWT.
type JWTValidateArgs struct {
	JWT string `json:"jwt"`
}

// Validate implements the Validatable interface.
func (j JWTValidateArgs) Validate(_ Validation) (ValidJWTValidateArgs, error) {
	valid := ValidJWTValidateArgs(j)
	return valid, nil
}

// ValidJWTValidateArgs are the validated arguments for validating a JWT.
type ValidJWTValidateArgs struct {
	JWT string
}

// JWTValidateRequest is the unvalidated request to validate a JWT.
type JWTValidateRequest struct {
	JWTValidateArgs JWTValidateArgs `json:"jwtValidateArgs"`
}

// Validate implements the Validatable interface.
func (j JWTValidateRequest) Validate(config Validation) (ValidJWTValidateRequest, error) {
	valid, err := j.JWTValidateArgs.Validate(config)
	if err != nil {
		return ValidJWTValidateRequest{}, fmt.Errorf("failed to validate JWT validate args: %w", err)
	}
	return ValidJWTValidateRequest{
		JWTValidateArgs: valid,
	}, nil
}

// ValidJWTValidateRequest is the validated request to validate a JWT.
type ValidJWTValidateRequest struct {
	JWTValidateArgs ValidJWTValidateArgs
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
