package model

import (
	"fmt"
	"time"
)

type OTPCreateParams struct {
	CharSetAlphaLower bool  `json:"charSetAlphaLower"`
	CharSetAlphaUpper bool  `json:"charSetAlphaUpper"`
	CharSetNumeric    bool  `json:"charSetNumeric"`
	Length            uint  `json:"length"`
	LifespanSeconds   int64 `json:"lifespanSeconds"`
}

func (o OTPCreateParams) Validate(config Validation) (ValidOTPCreateParams, error) {
	if !o.CharSetAlphaLower && !o.CharSetAlphaUpper && !o.CharSetNumeric {
		return ValidOTPCreateParams{}, fmt.Errorf("%w: at least one character set must be selected", ErrInvalidModel)
	}
	length := o.Length
	if length == 0 {
		length = 6
	} else if length < 1 || length > 12 { // Limited by email template.
		return ValidOTPCreateParams{}, fmt.Errorf("%w: link length must be between 1 and 12", ErrInvalidModel)
	}
	lifespan := time.Duration(o.LifespanSeconds) * time.Second
	if lifespan == 0 {
		lifespan = time.Hour
	} else if lifespan < 5*time.Second || lifespan > config.LifeSpanSeconds.Get() {
		return ValidOTPCreateParams{}, fmt.Errorf("%w: link lifespan must be between 5 and %d", ErrInvalidModel, int(config.LifeSpanSeconds.Get().Seconds()))
	}
	valid := ValidOTPCreateParams{
		CharSetAlphaLower: o.CharSetAlphaLower,
		CharSetAlphaUpper: o.CharSetAlphaUpper,
		CharSetNumeric:    o.CharSetNumeric,
		Length:            length,
		Lifespan:          lifespan,
	}
	return valid, nil
}

type ValidOTPCreateParams struct {
	CharSetAlphaLower bool
	CharSetAlphaUpper bool
	CharSetNumeric    bool
	Length            uint
	Lifespan          time.Duration
}

type OTPCreateRequest struct {
	OTPCreateParams OTPCreateParams `json:"otpCreateParams"`
}

func (o OTPCreateRequest) Validate(config Validation) (ValidOTPCreateRequest, error) {
	validOTPCreateParams, err := o.OTPCreateParams.Validate(config)
	if err != nil {
		return ValidOTPCreateRequest{}, fmt.Errorf("failed to validate OTP create args: %w", err)
	}
	valid := ValidOTPCreateRequest{
		OTPCreateParams: validOTPCreateParams,
	}
	return valid, nil
}

type ValidOTPCreateRequest struct {
	OTPCreateParams ValidOTPCreateParams
}

type OTPCreateResults struct {
	ID  string `json:"id"`
	OTP string `json:"otp"`
}

type OTPCreateResponse struct {
	OTPCreateResults OTPCreateResults `json:"otpCreateResults"`
	RequestMetadata  RequestMetadata  `json:"requestMetadata"`
}
