// Package recaptcha implements server side validation of reCAPTCHA V3 responses.
//
// See the below linked documentation for more information:
// https://developers.google.com/recaptcha/docs/v3
// https://developers.google.com/recaptcha/docs/verify
package recaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrCheck indicates that the reCAPTCHA response failed a check.
	ErrCheck = errors.New("reCAPTCHA check failed")

	// ErrInvalidVerifyInputs indicates that the reCAPTCHA V3 verify inputs are invalid.
	ErrInvalidVerifyInputs = errors.New("reCAPTCHA V3 verify inputs are invalid")
)

// VerifierV3Options are the options for creating a new reCAPTCHA V3 verifier. All fields can safely be left blank.
type VerifierV3Options struct {
	HTTPClient *http.Client
	VerifyURL  string
}

// V3ResponseCheckOptions are the options for checking a reCAPTCHA V3 response.
type V3ResponseCheckOptions struct {
	// APKPackageName confirms the response's APK package name is in the given slice.
	APKPackageName []string
	// Action confirms the response's action is in the given slice.
	Action []string
	// Hostname confirms the response's hostname is in the given slice.
	Hostname []string
	// Score confirms the response's score is at least this value.
	Score float64
}

// V3Response is the response from Google's reCAPTCHA V3 API.
type V3Response struct {
	APKPackageName string    `json:"apk_package_name"`
	Action         string    `json:"action"`
	ChallengeTS    time.Time `json:"challenge_ts"`
	ErrorCodes     []string  `json:"error-codes"`
	Hostname       string    `json:"hostname"`
	Score          float64   `json:"score"`
	Success        bool      `json:"success"`
}

// Check confirms the reCAPATCHA V3 response is valid and meets the given requirements.
func (resp V3Response) Check(options V3ResponseCheckOptions) error {
	if len(resp.ErrorCodes) > 0 {
		return fmt.Errorf("%w: error codes: %v", ErrCheck, resp.ErrorCodes)
	}
	if !resp.Success {
		return fmt.Errorf("%w: reCAPTCHA response success is false", ErrCheck)
	}
	if len(options.APKPackageName) != 0 {
		ok := inSlice(resp.APKPackageName, options.APKPackageName)
		if !ok {
			return fmt.Errorf("%w: APK package name %q not in set", ErrCheck, resp.APKPackageName)
		}
	}
	if len(options.Action) != 0 {
		ok := inSlice(resp.Action, options.Action)
		if !ok {
			return fmt.Errorf("%w: action %q not in set", ErrCheck, resp.Action)
		}
	}
	if len(options.Hostname) != 0 {
		ok := inSlice(resp.Hostname, options.Hostname)
		if !ok {
			return fmt.Errorf("%w: hostname %q not in set", ErrCheck, resp.Hostname)
		}
	}
	if options.Score > 0 {
		if resp.Score < options.Score {
			return fmt.Errorf("%w: score %f less than %f", ErrCheck, resp.Score, options.Score)
		}
	}
	return nil
}

// VerifierV3 is the interface for verifying reCAPTCHA V3 responses.
type VerifierV3 interface {
	Verify(ctx context.Context, response string, remoteIP string) (V3Response, error)
}

// recaptchaVerifierV3 is the default implementation of VerifierV3.
type recaptchaVerifierV3 struct {
	httpClient *http.Client
	secret     string
	verifyURL  string
}

// NewVerifierV3 creates a new reCAPTCHA V3 verifier.
func NewVerifierV3(secret string, options VerifierV3Options) VerifierV3 {
	httpClient := http.DefaultClient
	if options.HTTPClient != nil {
		httpClient = options.HTTPClient
	}
	verifyURL := "https://www.google.com/recaptcha/api/siteverify"
	if options.VerifyURL != "" {
		verifyURL = options.VerifyURL
	}
	return &recaptchaVerifierV3{
		httpClient: httpClient,
		secret:     secret,
		verifyURL:  verifyURL,
	}
}

// Verify helps implement the VerifierV3 interface.
func (verifier recaptchaVerifierV3) Verify(ctx context.Context, response string, remoteIP string) (V3Response, error) {
	if response == "" {
		return V3Response{}, fmt.Errorf("%w: response (g-recaptcha-response from client) is an empty string", ErrInvalidVerifyInputs)
	}

	form := url.Values{
		"secret":   {verifier.secret},
		"response": {response},
	}

	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifier.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return V3Response{}, fmt.Errorf("failed to create reCAPTCHA V3 request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := verifier.httpClient.Do(req)
	if err != nil {
		return V3Response{}, fmt.Errorf("failed to perform reCAPTCHA V3 request: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var r V3Response
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return V3Response{}, fmt.Errorf("failed to parse reCAPTCHA V3 JSON response: %w", err)
	}

	return r, nil
}

type testVerifierV3 struct {
	response V3Response
	err      error
}

// NewTestVerifierV3 creates a new reCAPTCHA V3 verifier that returns the given response and error. The intended purpose
// is for testing.
func NewTestVerifierV3(response V3Response, err error) VerifierV3 {
	return &testVerifierV3{
		response: response,
		err:      err,
	}
}

// Verify helps implement the VerifierV3 interface.
func (t testVerifierV3) Verify(_ context.Context, _ string, _ string) (V3Response, error) {
	return t.response, t.err
}

func inSlice(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
