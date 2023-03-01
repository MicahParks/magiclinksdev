package model

import (
	"fmt"

	"github.com/google/uuid"
)

// AdminCreateArgs are the unvalidated arguments for creating an admin.
type AdminCreateArgs struct {
	APIKey                   uuid.UUID                `json:"apiKey"`
	Aud                      uuid.UUID                `json:"aud"`
	UUID                     uuid.UUID                `json:"uuid"`
	ServiceAccountCreateArgs ServiceAccountCreateArgs `json:"serviceAccountCreateArgs"`
}

// Validate validates the admin create arguments.
func (a AdminCreateArgs) Validate(config Validation) (ValidAdminCreateArgs, error) {
	saArgs, err := a.ServiceAccountCreateArgs.Validate(config)
	if err != nil {
		return ValidAdminCreateArgs{}, fmt.Errorf("failed to validate service account args: %w", err)
	}
	valid := ValidAdminCreateArgs{
		APIKey:                        a.APIKey,
		Aud:                           a.Aud,
		UUID:                          a.UUID,
		ValidServiceAccountCreateArgs: saArgs,
	}
	return valid, nil
}

// ValidAdminCreateArgs are the validated arguments for creating an admin.
type ValidAdminCreateArgs struct {
	APIKey                        uuid.UUID
	Aud                           uuid.UUID
	UUID                          uuid.UUID
	ValidServiceAccountCreateArgs ValidServiceAccountCreateArgs
}
