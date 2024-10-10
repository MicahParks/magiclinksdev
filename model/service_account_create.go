package model

import (
	"fmt"
)

// ServiceAccountCreateParams are the unvalidated parameters to create a service account.
type ServiceAccountCreateParams struct{}

// Validate implements the Validatable interface.
func (s ServiceAccountCreateParams) Validate(_ Validation) (ValidServiceAccountCreateParams, error) {
	return ValidServiceAccountCreateParams(s), nil
}

// ValidServiceAccountCreateParams are the validated parameters to create a service account.
type ValidServiceAccountCreateParams struct{}

// ServiceAccountCreateRequest is the unvalidated request to create a service account.
type ServiceAccountCreateRequest struct {
	ServiceAccountCreateParams ServiceAccountCreateParams `json:"serviceAccountCreateParams"`
}

// Validate implements the Validatable interface.
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

// ValidServiceAccountCreateRequest is the validated request to create a service account.
type ValidServiceAccountCreateRequest struct {
	ServiceAccountCreateParams ValidServiceAccountCreateParams
}

// ServiceAccountCreateResults are the results of creating a service account.
type ServiceAccountCreateResults struct {
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

// ServiceAccountCreateResponse is the response to creating a service account.
type ServiceAccountCreateResponse struct {
	ServiceAccountCreateResults ServiceAccountCreateResults `json:"serviceAccountCreateResults"`
	RequestMetadata             RequestMetadata             `json:"requestMetadata"`
}
