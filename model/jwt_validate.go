package model

import (
	"encoding/json"
	"fmt"
)

type JWTValidateParams struct {
	JWT string `json:"jwt"`
}

func (j JWTValidateParams) Validate(_ Validation) (ValidJWTValidateParams, error) {
	valid := ValidJWTValidateParams(j)
	return valid, nil
}

type ValidJWTValidateParams struct {
	JWT string
}

type JWTValidateRequest struct {
	JWTValidateParams JWTValidateParams `json:"jwtValidateParams"`
}

func (j JWTValidateRequest) Validate(config Validation) (ValidJWTValidateRequest, error) {
	valid, err := j.JWTValidateParams.Validate(config)
	if err != nil {
		return ValidJWTValidateRequest{}, fmt.Errorf("failed to validate JWT validate args: %w", err)
	}
	return ValidJWTValidateRequest{
		JWTValidateParams: valid,
	}, nil
}

type ValidJWTValidateRequest struct {
	JWTValidateParams ValidJWTValidateParams
}

type JWTValidateResults struct {
	JWTClaims json.RawMessage `json:"claims"`
}

type JWTValidateResponse struct {
	JWTValidateResults JWTValidateResults `json:"jwtValidateResults"`
	RequestMetadata    RequestMetadata    `json:"requestMetadata"`
}
