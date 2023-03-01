package mldtest

import (
	"context"

	"github.com/MicahParks/magiclinksdev/email"
)

// ErrorProvider is an email provider that always returns an error.
type ErrorProvider struct{}

// Send implements the email.Provider interface.
func (e ErrorProvider) Send(_ context.Context, _ email.Email) error {
	return ErrMLDTest
}

// NopProvider is an email provider that does nothing.
type NopProvider struct{}

// Send implements the email.Provider interface.
func (n NopProvider) Send(_ context.Context, _ email.Email) error {
	return nil
}
