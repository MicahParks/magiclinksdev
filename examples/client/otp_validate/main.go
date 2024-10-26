package main

import (
	"context"
	"encoding/json"
	"flag"
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

	var (
		id  string
		otp string
	)
	flag.StringVar(&id, "id", "", "The ID of the OTP to validate.")
	flag.StringVar(&otp, "otp", "", "The OTP to validate.")
	flag.Parse()
	if id == "" || otp == "" {
		flag.Usage()
		os.Exit(1)
	}

	c, err := client.New(mldtest.APIKey, mldtest.Aud, mldtest.BaseURL, mldtest.Iss, client.Options{})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create client.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	req := model.OTPValidateRequest{
		OTPValidateParams: model.OTPValidateParams{
			ID:  id,
			OTP: otp,
		},
	}
	resp, mldErr, err := c.OTPValidate(ctx, req)
	if err != nil {
		if mldErr.Code != 0 {
			logger = logger.With(
				"code", mldErr.Code,
				"message", mldErr.Message,
				"requestUUID", mldErr.RequestMetadata.UUID,
			)
		}
		logger.ErrorContext(ctx, "Failed to validate OTP.",
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
	println("OTP is valid.")
}
