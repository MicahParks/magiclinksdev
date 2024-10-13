package mldtest

import (
	"context"

	"github.com/MicahParks/magiclinksdev/email"
)

// ErrorProvider is an email provider that always returns an error.
type ErrorProvider struct{}

func (e ErrorProvider) SendMagicLink(_ context.Context, _ email.Email) error {
	return ErrMLDTest
}
func (e ErrorProvider) SendOTP(_ context.Context, _ email.Email) error {
	return ErrMLDTest
}

// NopProvider is an email provider that does nothing.
type NopProvider struct{}

func (n NopProvider) SendMagicLink(_ context.Context, _ email.Email) error {
	return nil
}
func (n NopProvider) SendOTP(_ context.Context, _ email.Email) error {
	return nil
}
