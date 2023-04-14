package magiclink

import (
	"fmt"
	"html/template"
	"net/http"

	jt "github.com/MicahParks/jsontype"
	"github.com/MicahParks/recaptcha"
)

// ReCAPTCHAV3Config is the configuration for Google's reCAPTCHA v3.
type ReCAPTCHAV3Config struct {
	APKPackageName []string `json:"apkPackageName"`
	Action         []string `json:"action"`
	Hostname       []string `json:"hostname"`
	MinScore       float64  `json:"minScore"`
	SecretKey      string   `json:"secretKey"`
	TemplateConfig ReCAPTCHAV3TemplateConfig
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (r ReCAPTCHAV3Config) DefaultsAndValidate() (ReCAPTCHAV3Config, error) {
	if r.MinScore == 0 {
		r.MinScore = 0.5
	}
	if r.SecretKey == "" {
		return r, fmt.Errorf("%w: ReCAPTCHA v3 secret key is required", jt.ErrDefaultsAndValidate)
	}
	var err error
	r.TemplateConfig, err = r.TemplateConfig.DefaultsAndValidate()
	if err != nil {
		return r, fmt.Errorf("failed to validate ReCAPTCHA v3 template data: %w", err)
	}
	return r, nil
}

type ReCAPTCHAV3Redirector[CustomCreateArgs, CustomReadResponse any] struct {
	checkOpts  recaptcha.V3ResponseCheckOptions
	tmpl       *template.Template
	tmplConfig ReCAPTCHAV3TemplateConfig
	verifier   recaptcha.VerifierV3
}

func NewReCAPTCHAV3Redirector[CustomCreateArgs, CustomReadResponse any](config ReCAPTCHAV3Config) Redirector[CustomCreateArgs, CustomReadResponse] {
	tmpl := template.Must(template.New("").Parse(recaptchav3Template))
	checkOpts := recaptcha.V3ResponseCheckOptions{
		APKPackageName: config.APKPackageName,
		Action:         config.Action,
		Hostname:       config.Hostname,
		Score:          config.MinScore,
	}
	r := ReCAPTCHAV3Redirector[CustomCreateArgs, CustomReadResponse]{
		checkOpts:  checkOpts,
		tmpl:       tmpl,
		tmplConfig: config.TemplateConfig,
		verifier:   recaptcha.NewVerifierV3(config.SecretKey, recaptcha.VerifierV3Options{}),
	}
	return r
}

func (r ReCAPTCHAV3Redirector[CustomCreateArgs, CustomReadResponse]) Redirect(args RedirectArgs[CustomCreateArgs, CustomReadResponse]) {
	token := args.Request.URL.Query().Get("token") // TODO Make this a constant.
	if token != "" && args.Request.Method == http.MethodPost {
		resp, err := r.verifier.Verify(args.Request.Context(), token, "") // remoteIP left blank because reverse-proxies are a common use case. Could be configurable.
		if err != nil {
			args.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = resp.Check(r.checkOpts)
		if err != nil {
			args.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		args.Writer.WriteHeader(http.StatusOK)
		// TODO Write actual magic link to response.
	}

	tData := recaptchav3TemplateData{
		ButtonSkipsVerification: false, // TODO Get from create args.
		Config:                  r.tmplConfig,
	}

	err := r.tmpl.Execute(args.Writer, tData)
	if err != nil {
		args.Writer.WriteHeader(http.StatusInternalServerError)
	}
}

type recaptchav3TemplateData struct {
	ButtonSkipsVerification bool
	Config                  ReCAPTCHAV3TemplateConfig
}

// ReCAPTCHAV3TemplateConfig is the configuration for the HTML template for Google's reCAPTCHA v3.
type ReCAPTCHAV3TemplateConfig struct {
	CSS              template.CSS  `json:"css"`
	Code             string        `json:"code"`
	HTMLTitle        string        `json:"htmlTitle"`
	Instruction      string        `json:"instruction"`
	ReCAPTCHASiteKey template.HTML `json:"reCAPTCHASiteKey"`
	Title            string        `json:"title"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (f ReCAPTCHAV3TemplateConfig) DefaultsAndValidate() (ReCAPTCHAV3TemplateConfig, error) {
	if f.CSS == "" {
		f.CSS = template.CSS(defaultCSS)
	}
	if f.Instruction == "" {
		f.Instruction = "Click the below button if this page does not automatically redirect. This page is meant to stop robots from using magic links."
	}
	if f.HTMLTitle == "" {
		f.HTMLTitle = "Magic Link - Browser Check"
	}
	if f.ReCAPTCHASiteKey == "" {
		return f, fmt.Errorf("%w: ReCAPTCHASiteKey is required", jt.ErrDefaultsAndValidate)
	}
	if f.Code == "" {
		f.Code = "BROWSER CHECK"
	}
	if f.Title == "" {
		f.Title = "Checking your browser..."
	}
	return f, nil
}
