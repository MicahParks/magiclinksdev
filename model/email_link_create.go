package model

import (
	"fmt"
	"net/mail"
	"unicode/utf8"
)

// EmailLinkCreateArgs are the unvalidated arguments for creating an email link.
type EmailLinkCreateArgs struct {
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
func (p EmailLinkCreateArgs) Validate(config Validation) (ValidEmailLinkCreateArgs, error) {
	if p.ButtonText == "" {
		p.ButtonText = "Magic link"
	}
	if p.LogoImageURL != "" {
		u, err := httpURL(config, p.LogoClickURL)
		if err != nil {
			return ValidEmailLinkCreateArgs{}, fmt.Errorf("failed to parse logo click URL: %w", err)
		}
		p.LogoClickURL = u.String()
		u, err = httpURL(config, p.LogoImageURL)
		if err != nil {
			return ValidEmailLinkCreateArgs{}, fmt.Errorf("failed to parse logo image URL: %w", err)
		}
		p.LogoImageURL = u.String()
	} else {
		p.LogoClickURL = ""
	}
	runeCount := uint(utf8.RuneCountInString(p.ServiceName))
	if runeCount < config.ServiceNameMinUTF8 || runeCount > config.ServiceNameMaxUTF8 {
		return ValidEmailLinkCreateArgs{}, fmt.Errorf("%w: service name must be between %d and %d UTF8 runes", ErrInvalidModel, config.ServiceNameMinUTF8, config.ServiceNameMaxUTF8)
	}
	if len(p.Subject) < 5 || len(p.Subject) > 100 {
		return ValidEmailLinkCreateArgs{}, fmt.Errorf("%w: subject must be between 5 and 100 characters", ErrInvalidModel)
	}
	if len(p.Title) < 5 || len(p.Title) > 256 {
		return ValidEmailLinkCreateArgs{}, fmt.Errorf("%w: title must be between 5 and 256 characters", ErrInvalidModel)
	}
	address, err := mail.ParseAddress(p.ToEmail)
	if err != nil {
		return ValidEmailLinkCreateArgs{}, fmt.Errorf("failed to parse email address: %w", err)
	}
	address.Name = p.ToName
	valid := ValidEmailLinkCreateArgs{
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

// ValidEmailLinkCreateArgs are the validated arguments for creating an email link.
type ValidEmailLinkCreateArgs struct {
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

// EmailLinkCreateRequest is the unvalidated request to create an email link.
type EmailLinkCreateRequest struct {
	EmailArgs EmailLinkCreateArgs `json:"emailArgs"`
	LinkArgs  LinkCreateArgs      `json:"linkArgs"`
}

// Validate implements the Validatable interface.
func (b EmailLinkCreateRequest) Validate(config Validation) (ValidEmailLinkCreateRequest, error) {
	emailArgs, err := b.EmailArgs.Validate(config)
	if err != nil {
		return ValidEmailLinkCreateRequest{}, fmt.Errorf("failed to validate email args: %w", err)
	}
	linkArgs, err := b.LinkArgs.Validate(config)
	if err != nil {
		return ValidEmailLinkCreateRequest{}, fmt.Errorf("failed to validate link args: %w", err)
	}
	valid := ValidEmailLinkCreateRequest{
		EmailArgs: emailArgs,
		LinkArgs:  linkArgs,
	}
	return valid, nil
}

// ValidEmailLinkCreateRequest is the validated request to create an email link.
type ValidEmailLinkCreateRequest struct {
	EmailArgs ValidEmailLinkCreateArgs
	LinkArgs  ValidLinkCreateArgs
}

// EmailLinkCreateResults are the results of creating an email link.
type EmailLinkCreateResults struct {
	LinkCreateResults LinkCreateResults `json:"linkCreateResults"`
}

// EmailLinkCreateResponse is the response to creating an email link.
type EmailLinkCreateResponse struct {
	EmailLinkCreateResults EmailLinkCreateResults `json:"emailLinkCreateResults"`
	RequestMetadata        RequestMetadata        `json:"requestMetadata"`
}
