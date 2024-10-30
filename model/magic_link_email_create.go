package model

import (
	"fmt"
	"net/mail"
	"unicode/utf8"
)

type MagicLinkEmailCreateParams struct {
	ButtonText   string `json:"buttonText"`
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

func (p MagicLinkEmailCreateParams) Validate(config Validation) (ValidMagicLinkEmailCreateParams, error) {
	if p.ButtonText == "" {
		p.ButtonText = "Magic link"
	}
	if p.LogoImageURL != "" {
		u, err := httpURL(config, p.LogoClickURL)
		if err != nil {
			return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("failed to parse logo click URL: %w", err)
		}
		p.LogoClickURL = u.String()
		u, err = httpURL(config, p.LogoImageURL)
		if err != nil {
			return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("failed to parse logo image URL: %w", err)
		}
		p.LogoImageURL = u.String()
	} else {
		p.LogoClickURL = ""
	}
	runeCount := uint(utf8.RuneCountInString(p.ServiceName))
	if runeCount < config.ServiceNameMinUTF8 || runeCount > config.ServiceNameMaxUTF8 {
		return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("%w: service name must be between %d and %d UTF8 runes", ErrInvalidModel, config.ServiceNameMinUTF8, config.ServiceNameMaxUTF8)
	}
	if len(p.Subject) < 5 || len(p.Subject) > 100 {
		return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("%w: subject must be between 5 and 100 characters", ErrInvalidModel)
	}
	if len(p.Title) < 5 || len(p.Title) > 256 {
		return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("%w: title must be between 5 and 256 characters", ErrInvalidModel)
	}
	address, err := mail.ParseAddress(p.ToEmail)
	if err != nil {
		return ValidMagicLinkEmailCreateParams{}, fmt.Errorf("failed to parse email address: %w", err)
	}
	address.Name = p.ToName
	valid := ValidMagicLinkEmailCreateParams{
		ButtonText:   p.ButtonText,
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

type ValidMagicLinkEmailCreateParams struct {
	ButtonText   string
	Greeting     string
	LogoClickURL string
	LogoImageURL string
	ServiceName  string
	Subject      string
	SubTitle     string
	Title        string
	ToEmail      *mail.Address
}

type MagicLinkEmailCreateRequest struct {
	MagicLinkCreateParams      MagicLinkCreateParams      `json:"magicLinkCreateParams"`
	MagicLinkEmailCreateParams MagicLinkEmailCreateParams `json:"magicLinkEmailCreateParams"`
}

func (b MagicLinkEmailCreateRequest) Validate(config Validation) (ValidMagicLinkEmailCreateRequest, error) {
	magicLinkEmailCreateParams, err := b.MagicLinkEmailCreateParams.Validate(config)
	if err != nil {
		return ValidMagicLinkEmailCreateRequest{}, fmt.Errorf("failed to validate email params: %w", err)
	}
	magicLinkCreateParams, err := b.MagicLinkCreateParams.Validate(config)
	if err != nil {
		return ValidMagicLinkEmailCreateRequest{}, fmt.Errorf("failed to validate link params: %w", err)
	}
	valid := ValidMagicLinkEmailCreateRequest{
		MagicLinkCreateParams:      magicLinkCreateParams,
		MagicLinkEmailCreateParams: magicLinkEmailCreateParams,
	}
	return valid, nil
}

type ValidMagicLinkEmailCreateRequest struct {
	MagicLinkCreateParams      ValidMagicLinkCreateParams
	MagicLinkEmailCreateParams ValidMagicLinkEmailCreateParams
}

type MagicLinkEmailCreateResults struct {
	MagicLinkCreateResults MagicLinkCreateResults `json:"magicLinkCreateResults"`
}

type MagicLinkEmailCreateResponse struct {
	MagicLinkEmailCreateResults MagicLinkEmailCreateResults `json:"magicLinkEmailCreateResults"`
	RequestMetadata             RequestMetadata             `json:"requestMetadata"`
}
