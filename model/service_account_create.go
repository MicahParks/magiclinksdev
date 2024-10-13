package model

import (
	"fmt"
)

type ServiceAccountCreateParams struct{}

func (s ServiceAccountCreateParams) Validate(_ Validation) (ValidServiceAccountCreateParams, error) {
	return ValidServiceAccountCreateParams(s), nil
}

type ValidServiceAccountCreateParams struct{}

type ServiceAccountCreateRequest struct {
	ServiceAccountCreateParams ServiceAccountCreateParams `json:"serviceAccountCreateParams"`
}

func (s ServiceAccountCreateRequest) Validate(config Validation) (ValidServiceAccountCreateRequest, error) {
	serviceAccountCreateParams, err := s.ServiceAccountCreateParams.Validate(config)
	if err != nil {
		return ValidServiceAccountCreateRequest{}, fmt.Errorf("failed to validate service account create args: %w", err)
	}
	valid := ValidServiceAccountCreateRequest{
		ServiceAccountCreateParams: serviceAccountCreateParams,
	}
	return valid, nil
}

type ValidServiceAccountCreateRequest struct {
	ServiceAccountCreateParams ValidServiceAccountCreateParams
}

type ServiceAccountCreateResults struct {
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

type ServiceAccountCreateResponse struct {
	ServiceAccountCreateResults ServiceAccountCreateResults `json:"serviceAccountCreateResults"`
	RequestMetadata             RequestMetadata             `json:"requestMetadata"`
}
