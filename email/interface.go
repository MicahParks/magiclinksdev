package email

import (
	"context"
	"errors"
	"net/mail"
)

// ErrProvider is returned when there is an error with the email provider.
var ErrProvider = errors.New("error with email provider")

// Provider is the interface for an email provider.
type Provider interface {
	SendMagicLink(ctx context.Context, e Email) error
	SendOTP(ctx context.Context, e Email) error
}

// Email is the model for an email.
type Email struct {
	Subject      string
	TemplateData any
	To           *mail.Address
}
