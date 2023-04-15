package email

import (
	_ "embed"
	"html/template"
)

const (
	// MSOButtonStop is the HTML to stop MSO button spacing.
	MSOButtonStop = `<!--[if mso]>
                          <i style="letter-spacing: 24px">&#8202;</i>
                        <![endif]-->`
	// MSOButtonStart is the HTML to start MSO button spacing.
	MSOButtonStart = `<!--[if mso]>
                        <i style="mso-text-raise: 30px; letter-spacing: 24px">&#8202;</i>
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

// HTMLTemplate is the HTML template for the email.
//
//go:embed html.gohtml
var HTMLTemplate string

// TextTemplate is the text template for the email.
//
//go:embed text.gotxt
var TextTemplate string

// TemplateData is the data for the email templates.
type TemplateData struct {
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

type TemplateMetadata struct {
	HTMLInstruction string
	HTMLTitle       string
	MSOButtonStop   template.HTML
	MSOButtonStart  template.HTML
	MSOHead         template.HTML
}
