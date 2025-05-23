package main

import (
	"context"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.Default()

	c, err := client.New(mldtest.APIKey, mldtest.Aud, mldtest.BaseURL, mldtest.Iss, client.Options{})
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create client.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	req := model.ServiceAccountCreateRequest{
		ServiceAccountCreateParams: model.ServiceAccountCreateParams{},
	}
	resp, mldErr, err := c.ServiceAccountCreate(ctx, req)
	if err != nil {
		if mldErr.Code != 0 {
			logger = logger.With(
				"code", mldErr.Code,
				"message", mldErr.Message,
				"requestUUID", mldErr.RequestMetadata.UUID,
			)
		}
		logger.ErrorContext(ctx, "Failed to create service account.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	if mldErr.Code != 0 {
		logger.ErrorContext(ctx, "Failed to create service account.",
			"mldErr", mldErr,
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
