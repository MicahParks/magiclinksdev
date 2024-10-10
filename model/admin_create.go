package model

import (
	"fmt"

	"github.com/google/uuid"
)

// AdminCreateParams are the unvalidated parameters for creating an admin.
type AdminCreateParams struct {
	APIKey                     uuid.UUID                  `json:"apiKey"`
	Aud                        uuid.UUID                  `json:"aud"`
	UUID                       uuid.UUID                  `json:"uuid"`
	ServiceAccountCreateParams ServiceAccountCreateParams `json:"serviceAccountCreateParams"`
}

// Validate validates the admin create parameters.
func (a AdminCreateParams) Validate(config Validation) (ValidAdminCreateParams, error) {
	saParams, err := a.ServiceAccountCreateParams.Validate(config)
	if err != nil {
		return ValidAdminCreateParams{}, fmt.Errorf("failed to validate service account args: %w", err)
	}
	valid := ValidAdminCreateParams{
		APIKey:                          a.APIKey,
		Aud:                             a.Aud,
		UUID:                            a.UUID,
		ValidServiceAccountCreateParams: saParams,
	}
	return valid, nil
}

// ValidAdminCreateParams are the validated parameters for creating an admin.
type ValidAdminCreateParams struct {
	APIKey                          uuid.UUID
	Aud                             uuid.UUID
	UUID                            uuid.UUID
	ValidServiceAccountCreateParams ValidServiceAccountCreateParams
}
