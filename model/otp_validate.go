package model

import (
	"fmt"

	"github.com/google/uuid"
)

type OTPValidateParams struct {
	ID  string `json:"id"`
	OTP string `json:"otp"`
}

func (o OTPValidateParams) Validate(config Validation) (ValidOTPValidateParams, error) {
	_, err := uuid.Parse(o.ID)
	if err != nil {
		return ValidOTPValidateParams{}, fmt.Errorf("currently all OTP IDs must be UUIDs: %w", ErrInvalidModel)
	}
	if len(o.OTP) == 0 {
		return ValidOTPValidateParams{}, fmt.Errorf("%w: OTP cannot be empty", ErrInvalidModel)
	}
	valid := ValidOTPValidateParams{
		ID:  o.ID,
		OTP: o.OTP,
	}
	return valid, nil
}

type ValidOTPValidateParams struct {
	ID  string
	OTP string
}

type OTPValidateRequest struct {
	OTPValidateParams OTPValidateParams `json:"otpValidateParams"`
}

func (o OTPValidateRequest) Validate(config Validation) (ValidOTPValidateRequest, error) {
	validParams, err := o.OTPValidateParams.Validate(config)
	if err != nil {
		return ValidOTPValidateRequest{}, fmt.Errorf("failed to validate OTP validate args: %w", err)
	}
	valid := ValidOTPValidateRequest{
		OTPValidateParams: validParams,
	}
	return valid, nil
}

type ValidOTPValidateRequest struct {
	OTPValidateParams ValidOTPValidateParams
}

type OTPValidateResponse struct{}
