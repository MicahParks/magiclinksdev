package model

import (
	"fmt"
	"net/mail"
	"unicode/utf8"
)

type OTPEmailCreateParams struct {
	Greeting     string `json:"greeting"`
	LogoClickURL string `json:"logoClickURL"`
	LogoImageURL string `json:"logoImageURL"`
	ServiceName  string `json:"serviceName"`
	Subject      string `json:"subject"`
	SubTitle     string `json:"subTitle"`
	Title        string `json:"title"`
	ToEmail      string `json:"toEmail"`
	ToName       string `json:"toName"`
}

func (p OTPEmailCreateParams) Validate(config Validation) (ValidOTPEmailCreateParams, error) {
	if p.LogoImageURL != "" {
		u, err := httpURL(config, p.LogoClickURL)
		if err != nil {
			return ValidOTPEmailCreateParams{}, fmt.Errorf("failed to parse logo click URL: %w", err)
		}
		p.LogoClickURL = u.String()
		u, err = httpURL(config, p.LogoImageURL)
		if err != nil {
			return ValidOTPEmailCreateParams{}, fmt.Errorf("failed to parse logo image URL: %w", err)
		}
		p.LogoImageURL = u.String()
	} else {
		p.LogoClickURL = ""
	}
	runeCount := uint(utf8.RuneCountInString(p.ServiceName))
	if runeCount < config.ServiceNameMinUTF8 || runeCount > config.ServiceNameMaxUTF8 {
		return ValidOTPEmailCreateParams{}, fmt.Errorf("%w: service name must be between %d and %d UTF8 runes", ErrInvalidModel, config.ServiceNameMinUTF8, config.ServiceNameMaxUTF8)
	}
	if len(p.Subject) < 5 || len(p.Subject) > 100 {
		return ValidOTPEmailCreateParams{}, fmt.Errorf("%w: subject must be between 5 and 100 characters", ErrInvalidModel)
	}
	if len(p.Title) < 5 || len(p.Title) > 256 {
		return ValidOTPEmailCreateParams{}, fmt.Errorf("%w: title must be between 5 and 256 characters", ErrInvalidModel)
	}
	address, err := mail.ParseAddress(p.ToEmail)
	if err != nil {
		return ValidOTPEmailCreateParams{}, fmt.Errorf("failed to parse email address: %w", err)
	}
	address.Name = p.ToName
	valid := ValidOTPEmailCreateParams{
		Greeting:     p.Greeting,
		LogoClickURL: p.LogoClickURL,
		LogoImageURL: p.LogoImageURL,
		ServiceName:  p.ServiceName,
		Subject:      p.Subject,
		SubTitle:     p.SubTitle,
		Title:        p.Title,
		ToEmail:      address,
	}
	return valid, nil
}

type ValidOTPEmailCreateParams struct {
	Greeting     string
	LogoClickURL string
	LogoImageURL string
	ServiceName  string
	Subject      string
	SubTitle     string
	Title        string
	ToEmail      *mail.Address
}

type OTPEmailCreateRequest struct {
	OTPCreateParams      OTPCreateParams      `json:"otpCreateParams"`
	OTPEmailCreateParams OTPEmailCreateParams `json:"otpEmailCreateParams"`
}

func (b OTPEmailCreateRequest) Validate(config Validation) (ValidOTPEmailCreateRequest, error) {
	otpEmailCreateParams, err := b.OTPEmailCreateParams.Validate(config)
	if err != nil {
		return ValidOTPEmailCreateRequest{}, fmt.Errorf("failed to validate email params: %w", err)
	}
	otpCreateParams, err := b.OTPCreateParams.Validate(config)
	if err != nil {
		return ValidOTPEmailCreateRequest{}, fmt.Errorf("failed to validate link params: %w", err)
	}
	valid := ValidOTPEmailCreateRequest{
		OTPCreateParams:      otpCreateParams,
		OTPEmailCreateParams: otpEmailCreateParams,
	}
	return valid, nil
}

type ValidOTPEmailCreateRequest struct {
	OTPCreateParams      ValidOTPCreateParams
	OTPEmailCreateParams ValidOTPEmailCreateParams
}

type OTPEmailCreateResults struct {
	OTPCreateResults OTPCreateResults `json:"otpCreateResults"`
}

type OTPEmailCreateResponse struct {
	OTPEmailCreateResults OTPEmailCreateResults `json:"otpEmailCreateResults"`
	RequestMetadata       RequestMetadata       `json:"requestMetadata"`
}
