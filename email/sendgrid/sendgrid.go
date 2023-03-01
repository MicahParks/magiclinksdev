package sendgrid

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	netMail "net/mail"
	textTemplate "text/template"

	jt "github.com/MicahParks/jsontype"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/MicahParks/magiclinksdev/email"
)

// Config is the configuration for the SendGrid provider.
type Config struct {
	APIKey    string                         `json:"apiKey"`
	FromEmail *jt.JSONType[*netMail.Address] `json:"fromEmail"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (s Config) DefaultsAndValidate() (Config, error) {
	if s.APIKey == "" {
		return Config{}, fmt.Errorf("SendGrid API key not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	if s.FromEmail.Get() == nil {
		return Config{}, fmt.Errorf("SendGrid from email not provided in configuration: %w", jt.ErrDefaultsAndValidate)
	}
	return s, nil
}

type sendGrid struct {
	client   *sendgrid.Client
	from     *netMail.Address
	htmlTmpl *template.Template
	textTmpl *textTemplate.Template
}

// NewProvider creates a new SendGrid email provider.
func NewProvider(conf Config) (email.Provider, error) {
	client := sendgrid.NewSendClient(conf.APIKey)
	htmlTmpl := template.Must(template.New("").Parse(email.HTMLTemplate))
	textTmpl := textTemplate.Must(textTemplate.New("").Parse(email.TextTemplate))
	s := sendGrid{
		client:   client,
		from:     conf.FromEmail.Get(),
		htmlTmpl: htmlTmpl,
		textTmpl: textTmpl,
	}
	return s, nil
}

// Send implements the email.Provider interface.
func (s sendGrid) Send(ctx context.Context, e email.Email) error {
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

	from := mail.NewEmail(s.from.Name, s.from.Address)
	to := mail.NewEmail(e.To.Name, e.To.Address)
	message := mail.NewSingleEmail(from, e.Subject, to, textBuf.String(), htmlBuf.String())

	trackingSettings := mail.TrackingSettings{
		ClickTracking:        mail.NewClickTrackingSetting().SetEnable(false).SetEnableText(false),
		OpenTracking:         mail.NewOpenTrackingSetting().SetEnable(false),
		SubscriptionTracking: mail.NewSubscriptionTrackingSetting().SetEnable(false),
		GoogleAnalytics:      mail.NewGaSetting().SetEnable(false),
		BCC:                  mail.NewBCCSetting().SetEnable(false),
		BypassListManagement: mail.NewSetting(false),
		Footer:               mail.NewFooterSetting().SetEnable(false),
		SandboxMode:          nil,
	}
	message.SetTrackingSettings(&trackingSettings)

	resp, err := s.client.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email from SendGrid API client package: %w: %w", email.ErrProvider, err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("SendGrid API response status code not %d: got %d", resp.StatusCode, http.StatusAccepted)
	}

	return nil
}
