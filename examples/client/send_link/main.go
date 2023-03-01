package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/client"
	"github.com/MicahParks/magiclinksdev/mldtest"
	"github.com/MicahParks/magiclinksdev/model"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugared := logger.Sugar()

	c, err := client.New(mldtest.APIKey, mldtest.Aud, mldtest.BaseURL, mldtest.Iss, client.Options{})
	if err != nil {
		sugared.Fatalw("Failed to create client.",
			mld.LogErr, err,
		)
	}

	claims, err := json.Marshal(mldtest.TClaims)
	if err != nil {
		sugared.Fatalw("Failed to marshal claims.",
			mld.LogErr, err,
		)
	}

	req := model.LinkCreateRequest{
		LinkArgs: model.LinkCreateArgs{
			JWTCreateArgs: model.JWTCreateArgs{
				JWTClaims:          claims,
				JWTLifespanSeconds: 5,
			},
			LinkLifespan:     100,
			RedirectQueryKey: "",
			RedirectURL:      "https://jwtdebug.micahparks.com",
		},
	}
	resp, mldErr, err := c.LinkCreate(ctx, req)
	if err != nil {
		if mldErr.Code != 0 {
			sugared = sugared.With(
				"code", mldErr.Code,
				"message", mldErr.Message,
				"requestUUID", mldErr.RequestMetadata.UUID,
			)
		}
		sugared.Fatalw("Failed to create link.",
			mld.LogErr, err,
		)
	}

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		sugared.Fatalw("Failed to marshal response.",
			mld.LogErr, err,
		)
	}

	println(string(data))
}
