package magiclinksdev_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

func TestOTP(t *testing.T) {

	for _, tc := range []struct {
		name    string
		reqBody model.OTPCreateRequest
	}{
		{
			name: "NumericDefault",
			reqBody: model.OTPCreateRequest{
				OTPCreateParams: model.OTPCreateParams{
					CharSetAlphaLower: false,
					CharSetAlphaUpper: false,
					CharSetNumeric:    true,
					Length:            0,
					LifespanSeconds:   0,
				},
			},
		},
		{
			name: "AllLong",
			reqBody: model.OTPCreateRequest{
				OTPCreateParams: model.OTPCreateParams{
					CharSetAlphaLower: true,
					CharSetAlphaUpper: true,
					CharSetNumeric:    true,
					Length:            12,
					LifespanSeconds:   0,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			marshaled, err := json.Marshal(tc.reqBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			recorder := httptest.NewRecorder()
			u, err := assets.conf.Server.BaseURL.Get().Parse(network.PathOTPCreate)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, u.Path, bytes.NewReader(marshaled))
			req.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)
			req.Header.Set(middleware.APIKeyHeader, assets.sa.APIKey.String())
			assets.mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusCreated {
				t.Fatalf("Received non-200 status code: %d\n%s", recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get(mld.HeaderContentType) != mld.ContentTypeJSON {
				t.Fatalf("Received non-JSON content type: %s", recorder.Header().Get(mld.HeaderContentType))
			}

			var optCreateResponse model.OTPCreateResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &optCreateResponse)
			if err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			recorder = httptest.NewRecorder()
			u, err = assets.conf.Server.BaseURL.Get().Parse(network.PathOTPValidate)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			body := model.OTPValidateRequest{
				OTPValidateParams: model.OTPValidateParams{
					ID:  optCreateResponse.OTPCreateResults.ID,
					OTP: optCreateResponse.OTPCreateResults.OTP,
				},
			}
			marshaled, err = json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			req = httptest.NewRequest(http.MethodPost, u.String(), bytes.NewReader(marshaled))
			req.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)
			req.Header.Set(middleware.APIKeyHeader, assets.sa.APIKey.String())
			assets.mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
			}

			recorder = httptest.NewRecorder()
			u, err = assets.conf.Server.BaseURL.Get().Parse(network.PathOTPValidate)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			body = model.OTPValidateRequest{
				OTPValidateParams: model.OTPValidateParams{
					ID:  optCreateResponse.OTPCreateResults.ID,
					OTP: optCreateResponse.OTPCreateResults.OTP,
				},
			}
			marshaled, err = json.Marshal(body)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}
			req = httptest.NewRequest(http.MethodPost, u.String(), bytes.NewReader(marshaled))
			req.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)
			req.Header.Set(middleware.APIKeyHeader, assets.sa.APIKey.String())
			assets.mux.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}
