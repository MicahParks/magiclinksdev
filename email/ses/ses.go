package ses

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	netMail "net/mail"
	textTemplate "text/template"

	jt "github.com/MicahParks/jsontype"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"

	"github.com/MicahParks/magiclinksdev/email"
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
	from     *netMail.Address
	htmlTmpl *template.Template
	ses      *ses.SES
	textTmpl *textTemplate.Template
}

// NewProvider creates a new SES provider. It will create an AWS session using the provided configuration.
func NewProvider(conf Config) (SES, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(conf.AccessKeyID, conf.SecretKey, ""),
		Region:      aws.String(conf.AWSRegion),
	})
	if err != nil {
		return SES{}, fmt.Errorf("failed to create AWS session: %w", err)
	}
	svc := ses.New(sess)
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
func NewProviderInitialized(conf InitializedConfig, svc *ses.SES) (SES, error) {
	htmlTmpl := template.Must(template.New("").Parse(email.HTMLTemplate))
	textTmpl := textTemplate.Must(textTemplate.New("").Parse(email.TextTemplate))
	s := SES{
		from:     conf.FromEmail,
		htmlTmpl: htmlTmpl,
		ses:      svc,
		textTmpl: textTmpl,
	}
	return s, nil
}

// Send implements the email.Provider interface.
func (s SES) Send(ctx context.Context, e email.Email) error {
	htmlBuf := bytes.NewBuffer(nil)
	err := s.htmlTmpl.Execute(htmlBuf, e.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to execute template for HTML email: %w", err)
	}
	textBuf := bytes.NewBuffer(nil)
	err = s.textTmpl.Execute(textBuf, e.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to execute template for text email: %w", err)
	}

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(e.To.String()),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBuf.String()),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBuf.String()),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(e.Subject),
			},
		},
		Source: aws.String(s.from.String()),
	}

	_, err = s.ses.SendEmailWithContext(ctx, input)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				return fmt.Errorf("failed to send email due to AWS message rejected error: %w", err)
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				return fmt.Errorf("failed to send email due to AWS sending from unverfied domain error: %w", err)
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				return fmt.Errorf("failed to send email due to AWS configuration does not exist error: %w", err)
			default:
				return fmt.Errorf("failed to send email due to AWS error: %w", err)
			}
		}
		return fmt.Errorf("failed to send email via AWS SES: %w", err)
	}

	return nil
}
