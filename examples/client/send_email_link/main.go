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

	claims, err := json.Marshal(mldtest.TClaims)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to marshal claims.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	req := model.MagicLinkEmailCreateRequest{
		MagicLinkEmailCreateParams: model.MagicLinkEmailCreateParams{
			ButtonText:   "Log in",
			Greeting:     "Hello John Doe,",
			LogoClickURL: "https://magiclinks.dev",
			LogoImageURL: "https://magiclinks.dev/typeface-gray.png",
			ServiceName:  "magiclinks.dev",
			Subject:      "Your login for magiclinks.dev",
			SubTitle:     "No password required!",
			Title:        "Please click the below button to login.",
			ToEmail:      "johndoe@example.com",
			ToName:       "John Doe",
		},
		MagicLinkCreateParams: model.MagicLinkCreateParams{
			JWTCreateParams: model.JWTCreateParams{
				Claims:          claims,
				LifespanSeconds: 5,
			},
			LifespanSeconds:  60 * 60,
			RedirectQueryKey: "",
			RedirectURL:      "https://jwtdebug.micahparks.com",
		},
	}
	resp, mldErr, err := c.EmailLinkCreate(ctx, req)
	if err != nil {
		if mldErr.Code != 0 {
			logger = logger.With(
				"code", mldErr.Code,
				"message", mldErr.Message,
				"requestUUID", mldErr.RequestMetadata.UUID,
			)
		}
		logger.ErrorContext(ctx, "Failed to create email link.",
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
