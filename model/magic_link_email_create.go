package model

import (
	"fmt"
	"net/mail"
	"unicode/utf8"
)

// MagicLinkEmailCreateArgs are the unvalidated arguments for creating a magic link email.
type MagicLinkEmailCreateArgs struct {
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

// Validate implements the Validatable interface.
func (p MagicLinkEmailCreateArgs) Validate(config Validation) (ValidMagicLinkEmailCreateArgs, error) {
	if p.ButtonText == "" {
		p.ButtonText = "Magic link"
	}
	if p.LogoImageURL != "" {
		u, err := httpURL(config, p.LogoClickURL)
		if err != nil {
			return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("failed to parse logo click URL: %w", err)
		}
		p.LogoClickURL = u.String()
		u, err = httpURL(config, p.LogoImageURL)
		if err != nil {
			return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("failed to parse logo image URL: %w", err)
		}
		p.LogoImageURL = u.String()
	} else {
		p.LogoClickURL = ""
	}
	runeCount := uint(utf8.RuneCountInString(p.ServiceName))
	if runeCount < config.ServiceNameMinUTF8 || runeCount > config.ServiceNameMaxUTF8 {
		return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("%w: service name must be between %d and %d UTF8 runes", ErrInvalidModel, config.ServiceNameMinUTF8, config.ServiceNameMaxUTF8)
	}
	if len(p.Subject) < 5 || len(p.Subject) > 100 {
		return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("%w: subject must be between 5 and 100 characters", ErrInvalidModel)
	}
	if len(p.Title) < 5 || len(p.Title) > 256 {
		return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("%w: title must be between 5 and 256 characters", ErrInvalidModel)
	}
	address, err := mail.ParseAddress(p.ToEmail)
	if err != nil {
		return ValidMagicLinkEmailCreateArgs{}, fmt.Errorf("failed to parse email address: %w", err)
	}
	address.Name = p.ToName
	valid := ValidMagicLinkEmailCreateArgs{
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

// ValidMagicLinkEmailCreateArgs are the validated arguments for creating a magic link email.
type ValidMagicLinkEmailCreateArgs struct {
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

// MagicLinkEmailCreateRequest is the unvalidated request to create a magic link email.
type MagicLinkEmailCreateRequest struct {
	MagicLinkEmailCreateArgs MagicLinkEmailCreateArgs `json:"magicLinkEmailCreateArgs"`
	MagicLinkCreateArgs      MagicLinkCreateArgs      `json:"magicLinkCreateArgs"`
}

// Validate implements the Validatable interface.
func (b MagicLinkEmailCreateRequest) Validate(config Validation) (ValidMagicLinkEmailCreateRequest, error) {
	magicLinkEmailCreateArgs, err := b.MagicLinkEmailCreateArgs.Validate(config)
	if err != nil {
		return ValidMagicLinkEmailCreateRequest{}, fmt.Errorf("failed to validate email args: %w", err)
	}
	magicLinkCreateArgs, err := b.MagicLinkCreateArgs.Validate(config)
	if err != nil {
		return ValidMagicLinkEmailCreateRequest{}, fmt.Errorf("failed to validate link args: %w", err)
	}
	valid := ValidMagicLinkEmailCreateRequest{
		MagicLinkEmailCreateArgs: magicLinkEmailCreateArgs,
		MagicLinkCreateArgs:      magicLinkCreateArgs,
	}
	return valid, nil
}

// ValidMagicLinkEmailCreateRequest is the validated request to create an email link.
type ValidMagicLinkEmailCreateRequest struct {
	MagicLinkEmailCreateArgs ValidMagicLinkEmailCreateArgs
	MagicLinkCreateArgs      ValidMagicLinkCreateArgs
}

// MagicLinkEmailCreateResults are the results of creating an email link.
type MagicLinkEmailCreateResults struct {
	MagicLinkCreateResults MagicLinkCreateResults `json:"magicLinkCreateResults"`
}

// MagicLinkEmailCreateResponse is the response to creating an email link.
type MagicLinkEmailCreateResponse struct {
	MagicLinkEmailCreateResults MagicLinkEmailCreateResults `json:"magicLinkEmailCreateResults"`
	RequestMetadata             RequestMetadata             `json:"requestMetadata"`
}
