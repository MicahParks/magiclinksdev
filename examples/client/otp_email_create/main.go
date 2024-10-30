package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/client"
	"github.com/MicahParks/magiclinksdev/mldtest"
	"github.com/MicahParks/magiclinksdev/model"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := slog.Default()

	c, err := client.New(mldtest.APIKey, mldtest.Aud, mldtest.BaseURL, mldtest.Iss, client.Options{})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create client.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	req := model.OTPEmailCreateRequest{
		OTPCreateParams: model.OTPCreateParams{
			CharSetAlphaLower: false,
			CharSetAlphaUpper: false,
			CharSetNumeric:    true,
			Length:            0,
			LifespanSeconds:   0,
		},
		OTPEmailCreateParams: model.OTPEmailCreateParams{
			Greeting:     "Hello John Doe,",
			LogoClickURL: "https://magiclinks.dev",
			LogoImageURL: "https://magiclinks.dev/typeface-gray.png",
			ServiceName:  "magiclinks.dev",
			Subject:      "Verify your email - magiclinks.dev",
			SubTitle:     "Use this One-Time Password (OTP) to verify your email address.",
			Title:        "Your OTP is below.",
			ToEmail:      "johndoe@example.com",
			ToName:       "John Doe",
		},
	}
	resp, mldErr, err := c.OTPEmailCreate(ctx, req)
	if err != nil {
		if mldErr.Code != 0 {
			logger = logger.With(
				"code", mldErr.Code,
				"message", mldErr.Message,
				"requestUUID", mldErr.RequestMetadata.UUID,
			)
		}
		logger.ErrorContext(ctx, "Failed to create OTP.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		logger.ErrorContext(ctx, "Failed to marshal response.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	println(string(data))
}
