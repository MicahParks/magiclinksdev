package magiclink

import (
	"fmt"
	"html/template"
	"net/http"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/recaptcha"
)

const (
	// ReCAPTCHAV3QueryButtonBypassKey is the URL query parameter key for bypassing the reCAPTCHA v3 test.
	ReCAPTCHAV3QueryButtonBypassKey = "button-bypass"
	// ReCAPTCHAV3QueryButtonBypassValue is the URL query parameter value for bypassing the reCAPTCHA v3 test.
	ReCAPTCHAV3QueryButtonBypassValue = "true"
)

// ReCAPTCHAV3Config is the configuration for Google's reCAPTCHA v3.
type ReCAPTCHAV3Config struct {
	APKPackageName []string                `json:"apkPackageName"`
	Action         []string                `json:"action"`
	Hostname       []string                `json:"hostname"`
	MinScore       float64                 `json:"minScore"`
	SecretKey      string                  `json:"secretKey"`
	TemplateData   ReCAPTCHAV3TemplateData `json:"templateData"`

	Verifier recaptcha.VerifierV3 `json:"-"`
}

func (r ReCAPTCHAV3Config) DefaultsAndValidate() (ReCAPTCHAV3Config, error) {
	if r.MinScore == 0 {
		r.MinScore = 0.5
	}
	if r.SecretKey == "" {
		return r, fmt.Errorf("%w: ReCAPTCHA v3 secret key is required", jt.ErrDefaultsAndValidate)
	}
	var err error
	r.TemplateData, err = r.TemplateData.DefaultsAndValidate()
	if err != nil {
		return r, fmt.Errorf("failed to validate ReCAPTCHA v3 template data: %w", err)
	}
	return r, nil
}

// ReCAPTCHAV3Redirector is a Redirector that uses Google's reCAPTCHA v3 to verify the user.
type ReCAPTCHAV3Redirector struct {
	checkOpts recaptcha.V3ResponseCheckOptions
	tmpl      *template.Template
	tmplData  ReCAPTCHAV3TemplateData
	verifier  recaptcha.VerifierV3
}

// NewReCAPTCHAV3Redirector creates a new ReCAPTCHAV3Redirector with the given config.
func NewReCAPTCHAV3Redirector(config ReCAPTCHAV3Config) Redirector {
	tmpl := template.Must(template.New("").Parse(recaptchav3Template))
	checkOpts := recaptcha.V3ResponseCheckOptions{
		APKPackageName: config.APKPackageName,
		Action:         config.Action,
		Hostname:       config.Hostname,
		Score:          config.MinScore,
	}
	verifier := config.Verifier
	if verifier == nil {
		verifier = recaptcha.NewVerifierV3(config.SecretKey, recaptcha.VerifierV3Options{})
	}
	r := ReCAPTCHAV3Redirector{
		checkOpts: checkOpts,
		tmpl:      tmpl,
		tmplData:  config.TemplateData,
		verifier:  verifier,
	}
	return r
}

func (r ReCAPTCHAV3Redirector) Redirect(args RedirectorParams) {
	ctx := args.Request.Context()

	token := args.Request.URL.Query().Get("token")
	if args.Request.Method == http.MethodPost {
		if token != "" {
			resp, err := r.verifier.Verify(args.Request.Context(), token, "") // remoteIP left blank because reverse-proxies are a common use case. Could be configurable.
			if err != nil {
				args.Writer.WriteHeader(http.StatusBadRequest)
				return
			}
			err = resp.Check(r.checkOpts)
			if err != nil {
				args.Writer.WriteHeader(http.StatusBadRequest)
				return
			}
			jwtB64, response, err := args.ReadAndExpireLink(ctx, args.Secret)
			if err != nil {
				args.Writer.WriteHeader(http.StatusNotFound)
				return
			}
			u := redirectURLFromResponse(response, jwtB64)
			args.Writer.WriteHeader(http.StatusOK)
			_, _ = args.Writer.Write([]byte(u.String()))
			return
		}
		if r.tmplData.ButtonBypass && args.Request.URL.Query().Get(ReCAPTCHAV3QueryButtonBypassKey) == ReCAPTCHAV3QueryButtonBypassValue {
			jwtB64, response, err := args.ReadAndExpireLink(ctx, args.Secret)
			if err != nil {
				args.Writer.WriteHeader(http.StatusNotFound)
				return
			}
			u := redirectURLFromResponse(response, jwtB64)
			http.Redirect(args.Writer, args.Request, u.String(), http.StatusSeeOther)
			return
		}
	}

	tData := r.tmplData
	if tData.ButtonBypass {
		u := copyURL(args.Request.URL)
		query := u.Query()
		query.Set(ReCAPTCHAV3QueryButtonBypassKey, ReCAPTCHAV3QueryButtonBypassValue)
		u.RawQuery = query.Encode()
		tData.FormAction = u.String()
	}
	_ = r.tmpl.Execute(args.Writer, tData)
}

// ReCAPTCHAV3TemplateData is the configuration for the HTML template for Google's reCAPTCHA v3.
type ReCAPTCHAV3TemplateData struct {
	ButtonBypass bool          `json:"buttonBypass"`
	ButtonText   string        `json:"buttonText"`
	CSS          template.CSS  `json:"css"`
	Code         string        `json:"code"`
	HTMLTitle    string        `json:"htmlTitle"`
	Instruction  string        `json:"instruction"`
	SiteKey      template.HTML `json:"siteKey"`
	Title        string        `json:"title"`

	FormAction string `json:"-"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (r ReCAPTCHAV3TemplateData) DefaultsAndValidate() (ReCAPTCHAV3TemplateData, error) {
	if r.ButtonText == "" {
		r.ButtonText = "Continue"
	}
	if r.CSS == "" {
		r.CSS = template.CSS(defaultCSS)
	}
	if r.Instruction == "" {
		r.Instruction += "This page helps prevent robots from using magic links. You should be redirected automatically."
	}
	if r.HTMLTitle == "" {
		r.HTMLTitle = "Magic Link - Browser Check"
	}
	if r.SiteKey == "" {
		return r, fmt.Errorf("%w: SiteKey is required", jt.ErrDefaultsAndValidate)
	}
	if r.Code == "" {
		r.Code = "BROWSER CHECK"
	}
	if r.Title == "" {
		r.Title = "Checking your browser..."
	}
	return r, nil
}
