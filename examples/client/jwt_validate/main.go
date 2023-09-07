package main

import (
	"log/slog"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/client"
	"github.com/MicahParks/magiclinksdev/mldtest"
)

func main() {
	logger := slog.Default()

	c, err := client.New(mldtest.APIKey, mldtest.Aud, mldtest.BaseURL, mldtest.Iss, client.Options{})
	if err != nil {
		logger.Error("Failed to create client.",
			mld.LogErr, err,
		)
	}

	const rawJWT = "eyJhbGciOiJIUzI1NiIsImtpZCI6InRoaW5nIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.iOTw9YfyVMf19aUEht19XIKX7jhgh_rl2hlsrYx2gao"

	var claims mldtest.TestClaims
	token, err := c.LocalJWTValidate(rawJWT, &claims)
	if err != nil {
		logger.Error("Failed to validate JWT. This is normal for the default example.",
			mld.LogErr, err,
		)
	}

	logger.Info("JWT is valid.",
		"claims", claims,
		"token", token,
	)
}
