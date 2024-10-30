package mldtest

import (
	"context"

	"github.com/MicahParks/magiclinksdev/email"
)

// NopProvider is an email provider that does nothing.
type NopProvider struct{}

func (n NopProvider) SendMagicLink(_ context.Context, _ email.Email) error {
	return nil
}
func (n NopProvider) SendOTP(_ context.Context, _ email.Email) error {
	return nil
}
