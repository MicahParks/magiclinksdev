package model

import (
	"fmt"
)

// ServiceAccountCreateArgs are the unvalidated arguments to create a service account.
type ServiceAccountCreateArgs struct{}

// Validate implements the Validatable interface.
func (n ServiceAccountCreateArgs) Validate(config Validation) (ValidServiceAccountCreateArgs, error) {
	return ValidServiceAccountCreateArgs(n), nil
}

// ValidServiceAccountCreateArgs are the validated arguments to create a service account.
type ValidServiceAccountCreateArgs struct{}

// ServiceAccountCreateRequest is the unvalidated request to create a service account.
type ServiceAccountCreateRequest struct {
	CreateServiceAccountArgs ServiceAccountCreateArgs `json:"createServiceAccountArgs"`
}

// Validate implements the Validatable interface.
func (b ServiceAccountCreateRequest) Validate(config Validation) (ValidServiceAccountCreateRequest, error) {
	createServiceAccountArgs, err := b.CreateServiceAccountArgs.Validate(config)
	if err != nil {
		return ValidServiceAccountCreateRequest{}, fmt.Errorf("failed to validate create service account args: %w", err)
	}
	valid := ValidServiceAccountCreateRequest{
		CreateServiceAccountArgs: createServiceAccountArgs,
	}
	return valid, nil
}

// ValidServiceAccountCreateRequest is the validated request to create a service account.
type ValidServiceAccountCreateRequest struct {
	CreateServiceAccountArgs ValidServiceAccountCreateArgs
}

// ServiceAccountCreateResults are the results of creating a service account.
type ServiceAccountCreateResults struct {
	ServiceAccount ServiceAccount `json:"serviceAccount"`
}

// ServiceAccountCreateResponse is the response to creating a service account.
type ServiceAccountCreateResponse struct {
	CreateServiceAccountResults ServiceAccountCreateResults `json:"createServiceAccountResults"`
	RequestMetadata             RequestMetadata             `json:"requestMetadata"`
}
