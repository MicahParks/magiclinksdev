package email

import (
	_ "embed"
	"html/template"
)

const (
	// MSOButtonStop is the HTML to stop MSO button spacing.
	MSOButtonStop = `<!--[if mso]>
      <i hidden style="mso-font-width: 150%;">&emsp;&#8203;</i>
    <![endif]-->`
	// MSOButtonStart is the HTML to start MSO button spacing.
	MSOButtonStart = `<!--[if mso]>
      <i style="mso-font-width: 150%; mso-text-raise: 30px" hidden>&emsp;</i>
    <![endif]-->`
	// MSOHead is the HTML to start MSO head.
	MSOHead = `<!--[if mso]>
  <noscript>
    <xml>
      <o:OfficeDocumentSettings xmlns:o="urn:schemas-microsoft-com:office:office">
        <o:PixelsPerInch>96</o:PixelsPerInch>
      </o:OfficeDocumentSettings>
    </xml>
  </noscript>
  <style>
    td,th,div,p,a,h1,h2,h3,h4,h5,h6 {font-family: "Segoe UI", sans-serif; mso-line-height-rule: exactly;}
  </style>
  <![endif]-->`
)

// MagicLinkHTMLTemplate is the HTML template for the magic link email.
//
//go:embed html.gohtml
var MagicLinkHTMLTemplate string

// MagicLinkTextTemplate is the text template for the magic link email.
//
//go:embed text.gotxt
var MagicLinkTextTemplate string

// MagicLinkTemplateData is the data for the magic link email template.
type MagicLinkTemplateData struct {
	ButtonText   string
	Expiration   string
	Greeting     string
	MagicLink    string
	Meta         TemplateMetadata
	Subtitle     string
	Title        string
	LogoImageURL string
	LogoClickURL string
	LogoAltText  string
	ReCATPTCHA   bool
}

// OTPHTMLTemplate is the HTML template for the OTP email.
//
//go:embed otp_html.gohtml
var OTPHTMLTemplate string

// OTPTextTemplate is the text template for the OTP email.
//
//go:embed otp_text.gotxt
var OTPTextTemplate string

// OTPTemplateData is the data for the OTP email template.
type OTPTemplateData struct {
	Expiration   string
	Greeting     string
	MagicLink    string
	Meta         TemplateMetadata
	OTP          string
	Subtitle     string
	Title        string
	LogoImageURL string
	LogoClickURL string
	LogoAltText  string
}

// TemplateMetadata contains non-configurable metadata for the email templates.
type TemplateMetadata struct {
	HTMLInstruction string
	HTMLTitle       string
	MSOButtonStop   template.HTML
	MSOButtonStart  template.HTML
	MSOHead         template.HTML
}
