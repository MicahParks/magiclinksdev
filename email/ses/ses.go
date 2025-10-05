package ses

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	netMail "net/mail"
	textTemplate "text/template"

	jt "github.com/MicahParks/jsontype"
	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/email"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

const (
	charSet = "UTF-8"
)

// InitializedConfig is the configuration for the AWS SES provider after it has been initialized.
type InitializedConfig struct {
	FromEmail *netMail.Address
}

// Config is the configuration for the AWS SES provider. It includes assets to create an AWS session.
type Config struct {
	AWSRegion   string                         `json:"awsRegion"`
	AccessKeyID string                         `json:"accessKeyID"`
	FromEmail   *jt.JSONType[*netMail.Address] `json:"fromEmail"`
	SecretKey   string                         `json:"secretKey"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (c Config) DefaultsAndValidate() (Config, error) {
	if c.AWSRegion == "" {
		return Config{}, fmt.Errorf("AWS region not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	if c.AccessKeyID == "" {
		return Config{}, fmt.Errorf("AWS access key ID not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	if c.FromEmail.Get() == nil {
		return Config{}, fmt.Errorf("AWS SES from email not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	if c.SecretKey == "" {
		return Config{}, fmt.Errorf("AWS secret key not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	return c, nil
}

// SES is an email provider that uses AWS SES.
type SES struct {
	from              *netMail.Address
	magicLinkHTMLTmpl *template.Template
	magicLinkTxtTmpl  *textTemplate.Template
	oTPHTMLTmpl       *template.Template
	oTPTxtTmpl        *textTemplate.Template
	ses               *sesv2.Client
}

// NewProvider creates a new SES provider. It will create an AWS session using the provided configuration.
func NewProvider(ctx context.Context, conf Config) (SES, error) {
	appCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(conf.AccessKeyID, conf.SecretKey, ""))
	cfg, err := config.LoadDefaultConfig(ctx, func(o *config.LoadOptions) error {
		o.Credentials = appCreds
		o.Region = conf.AWSRegion
		return nil
	})
	if err != nil {
		return SES{}, fmt.Errorf("failed to load default AWS config: %w", err)
	}
	svc := sesv2.NewFromConfig(cfg)
	requiredConf := InitializedConfig{
		FromEmail: conf.FromEmail.Get(),
	}
	s, err := NewProviderInitialized(requiredConf, svc)
	if err != nil {
		return SES{}, fmt.Errorf("failed to create SES provider: %w", err)
	}
	return s, nil
}

// NewProviderInitialized creates a new SES provider with an initialized configuration.
func NewProviderInitialized(conf InitializedConfig, svc *sesv2.Client) (SES, error) {
	magicLinkHTMLTmpl := template.Must(template.New("").Parse(email.MagicLinkHTMLTemplate))
	magicLinkTxtTML := textTemplate.Must(textTemplate.New("").Parse(email.MagicLinkTextTemplate))
	otpHTMLTmpl := template.Must(template.New("").Parse(email.OTPHTMLTemplate))
	otpTxtTmpl := textTemplate.Must(textTemplate.New("").Parse(email.OTPTextTemplate))
	s := SES{
		from:              conf.FromEmail,
		magicLinkHTMLTmpl: magicLinkHTMLTmpl,
		magicLinkTxtTmpl:  magicLinkTxtTML,
		oTPHTMLTmpl:       otpHTMLTmpl,
		oTPTxtTmpl:        otpTxtTmpl,
		ses:               svc,
	}
	return s, nil
}

func (s SES) SendMagicLink(ctx context.Context, e email.Email) error {
	return s.sendEmail(ctx, e, s.magicLinkHTMLTmpl, s.magicLinkTxtTmpl)
}
func (s SES) SendOTP(ctx context.Context, e email.Email) error {
	return s.sendEmail(ctx, e, s.oTPHTMLTmpl, s.oTPTxtTmpl)
}

func (s SES) sendEmail(ctx context.Context, e email.Email, htmlTmpl *template.Template, txtTmpl *textTemplate.Template) error {
	htmlBuf := bytes.NewBuffer(nil)
	err := htmlTmpl.Execute(htmlBuf, e.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to execute template for HTML email: %w", err)
	}
	textBuf := bytes.NewBuffer(nil)
	err = txtTmpl.Execute(textBuf, e.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to execute template for text email: %w", err)
	}

	input := &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{
				e.To.String(),
			},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Body: &types.Body{
					Html: &types.Content{
						Charset: aws.String(charSet),
						Data:    aws.String(htmlBuf.String()),
					},
					Text: &types.Content{
						Charset: aws.String(charSet),
						Data:    aws.String(textBuf.String()),
					},
				},
				Subject: &types.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(e.Subject),
				},
			},
		},
		FromEmailAddress: mld.Ptr(s.from.String()),
	}

	_, err = s.ses.SendEmail(ctx, input)
	if err != nil {
		var mr *types.MessageRejected
		var dnve *types.MailFromDomainNotVerifiedException
		switch {
		case errors.As(err, &mr):
			return fmt.Errorf("failed to send email due to AWS message rejected error: %w", err)
		case errors.As(err, &dnve):
			return fmt.Errorf("failed to send email due to AWS sending from unverfied domain error: %w", err)
		}
		return fmt.Errorf("failed to send email via AWS SES: %w", err)
	}

	return nil
}
