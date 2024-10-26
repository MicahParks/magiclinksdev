package otp

import (
	"context"
	"errors"
	"time"
)

var ErrOTPInvalid = errors.New("OTP invalid")

type CreateParams struct {
	CharSetAlphaLower bool
	CharSetAlphaUpper bool
	CharSetNumeric    bool
	Expires           time.Time
	Length            uint
}

type CreateResult struct {
	CreateParams CreateParams
	ID           string
	OTP          string
}

type Storage interface {
	OTPCreate(ctx context.Context, params CreateParams) (CreateResult, error)
	OTPValidate(ctx context.Context, id, o string) error
}
