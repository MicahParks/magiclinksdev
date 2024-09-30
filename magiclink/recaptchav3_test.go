package magiclink_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MicahParks/recaptcha"

	"github.com/MicahParks/magiclinksdev/magiclink"
)

func TestReCAPTCHAV3Redirector_Redirect(t *testing.T) {
	const (
		jwtB64FromBackend = "jwtB64"
		magicLinkTarget   = "https://magiclinks.dev/"
	)

	magicLinkTargetWithJWT := fmt.Sprintf("%s?%s=%s", magicLinkTarget, magiclink.DefaultRedirectQueryKey, jwtB64FromBackend)

	tc := []struct {
		name         string
		body         string
		buttonBypass bool
		haveBody     bool
		htmlResp     bool
		httpRedirect bool
		method       string
		respCode     int
		url          string
		verifier     recaptcha.VerifierV3
	}{
		{
			name:     "NoButtonBypassFrontend",
			haveBody: true,
			htmlResp: true,
			method:   http.MethodGet,
			respCode: http.StatusOK,
			url:      "",
		},
		{
			name:     "NoButtonBypassBackendError",
			method:   http.MethodPost,
			respCode: http.StatusBadRequest,
			url:      "?token=error",
			verifier: recaptchav3Error{},
		},
		{
			name:     "NoButtonBypassBackendFailure",
			method:   http.MethodPost,
			respCode: http.StatusBadRequest,
			url:      "?token=bad",
			verifier: recaptchav3Failure{},
		},
		{
			name:     "NoButtonBypassBackendSuccess",
			body:     magicLinkTargetWithJWT,
			haveBody: true,
			method:   http.MethodPost,
			respCode: http.StatusOK,
			url:      "?token=good",
			verifier: recaptchav3Success{},
		},
		{
			name:         "ButtonBypassFrontend",
			buttonBypass: true,
			haveBody:     true,
			htmlResp:     true,
			method:       http.MethodGet,
			respCode:     http.StatusOK,
			url:          "",
		},
		{
			name:         "ButtonBypassBackend",
			buttonBypass: true,
			httpRedirect: true,
			method:       http.MethodPost,
			respCode:     http.StatusSeeOther,
			url:          fmt.Sprintf("?%s=%s", magiclink.ReCAPTCHAV3QueryButtonBypassKey, magiclink.ReCAPTCHAV3QueryButtonBypassValue),
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			conf := magiclink.ReCAPTCHAV3Config{
				TemplateData: magiclink.ReCAPTCHAV3TemplateData{
					ButtonBypass: tt.buttonBypass,
				},
				Verifier: tt.verifier,
			}
			redirector := magiclink.NewReCAPTCHAV3Redirector[any](conf)

			r, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v.", err)
			}
			recorder := httptest.NewRecorder()
			args := magiclink.RedirectorArgs[any]{
				ReadAndExpireLink: func(ctx context.Context, secret string) (jwtB64 string, response magiclink.ReadResponse[any], err error) {
					return jwtB64FromBackend, magiclink.ReadResponse[any]{
						CreateArgs: magiclink.CreateArgs{
							RedirectQueryKey: magiclink.DefaultRedirectQueryKey,
							RedirectURL:      must(url.Parse(magicLinkTarget)),
						},
					}, nil
				},
				Request: r,
				Writer:  recorder,
			}

			redirector.Redirect(args)

			if recorder.Code != tt.respCode {
				t.Fatalf("Expected status code %v, but got %v.", tt.respCode, recorder.Code)
			}

			if tt.httpRedirect {
				redirectTo := recorder.Header().Get("Location")
				if redirectTo != magicLinkTargetWithJWT {
					t.Fatalf("Expected redirect to %v, but got %v.", magicLinkTargetWithJWT, redirectTo)
				}
				return
			}

			if tt.haveBody {
				if recorder.Body.Len() == 0 {
					t.Fatalf("Expected a body, but got none.")
				}
				body := recorder.Body.String()
				if tt.htmlResp {
					if tt.buttonBypass != strings.Contains(body, "</button>") {
						t.Fatalf("Bypass button precense was incorrect.")
					}
				} else {
					if tt.body != "" && tt.body != body {
						t.Fatalf("Expected body %v, but got %v.", tt.body, body)
					}
				}
				return
			}

			if recorder.Body.Len() != 0 {
				t.Fatalf("Expected no body, but got %v.", recorder.Body.String())
			}
		})
	}
}

type recaptchav3Error struct{}

func (r recaptchav3Error) Verify(_ context.Context, _ string, _ string) (recaptcha.V3Response, error) {
	return recaptcha.V3Response{}, errors.New("testing error")
}

type recaptchav3Failure struct {
	args magiclink.ReCAPTCHAV3Config
}

func (r recaptchav3Failure) Verify(_ context.Context, _ string, _ string) (recaptcha.V3Response, error) {
	return recaptcha.V3Response{
		APKPackageName: firstOrEmpty(r.args.APKPackageName),
		Action:         firstOrEmpty(r.args.Action),
		Hostname:       firstOrEmpty(r.args.Hostname),
		Score:          0,
		Success:        false,
	}, nil
}

type recaptchav3Success struct {
	args magiclink.ReCAPTCHAV3Config
}

func (r recaptchav3Success) Verify(_ context.Context, _ string, _ string) (recaptcha.V3Response, error) {
	return recaptcha.V3Response{
		APKPackageName: firstOrEmpty(r.args.APKPackageName),
		Action:         firstOrEmpty(r.args.Action),
		Hostname:       firstOrEmpty(r.args.Hostname),
		Score:          1,
		Success:        true,
	}, nil
}

func firstOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
