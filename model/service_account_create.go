package model

import (
	"fmt"
)

// ServiceAccountCreateArgs are the unvalidated arguments to create a service account.
type ServiceAccountCreateArgs struct{}

// Validate implements the Validatable interface.
func (s ServiceAccountCreateArgs) Validate(_ Validation) (ValidServiceAccountCreateArgs, error) {
	return ValidServiceAccountCreateArgs(s), nil
}

// ValidServiceAccountCreateArgs are the validated arguments to create a service account.
type ValidServiceAccountCreateArgs struct{}

// ServiceAccountCreateRequest is the unvalidated request to create a service account.
type ServiceAccountCreateRequest struct {
	ServiceAccountCreateArgs ServiceAccountCreateArgs `json:"serviceAccountCreateArgs"`
}

// Validate implements the Validatable interface.
func (s ServiceAccountCreateRequest) Validate(config Validation) (ValidServiceAccountCreateRequest, error) {
	serviceAccountCreateArgs, err := s.ServiceAccountCreateArgs.Validate(config)
	if err != nil {
		return ValidServiceAccountCreateRequest{}, fmt.Errorf("failed to validate service account create args: %w", err)
	}
	valid := ValidServiceAccountCreateRequest{
		ServiceAccountCreateArgs: serviceAccountCreateArgs,
	}
	return valid, nil
}

// ValidServiceAccountCreateRequest is the validated request to create a service account.
type ValidServiceAccountCreateRequest struct {
	ServiceAccountCreateArgs ValidServiceAccountCreateArgs
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
