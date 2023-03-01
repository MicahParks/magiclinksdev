package main

import (
	"bytes"
	"context"
	_ "embed"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware"
)

//go:embed link.prod.json
var linkJSON []byte

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugared := logger.Sugar()

	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		sugared.Fatalw("Failed to parse URL.",
			mld.LogErr, err,
		)
	}
	u, err = u.Parse(network.PathLinkCreate)
	if err != nil {
		sugared.Fatalw("Failed to parse URL.",
			mld.LogErr, err,
		)
	}

	g, ctx := errgroup.WithContext(ctx)
	if err != nil {
		sugared.Fatalw("Failed to create errgroup.",
			mld.LogErr, err,
		)
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		g.Go(func() error {
			for j := 0; j < 1000; j++ {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(linkJSON))
				if err != nil {
					sugared.Fatalw("Failed to create request.",
						mld.LogErr, err,
					)
				}

				req.Header.Set(middleware.APIKeyHeader, "40084740-0bc3-455d-b298-e23a31561580") // Admin API key from config.

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					sugared.Fatalw("Failed to send request.",
						mld.LogErr, err,
					)
				}
				//goland:noinspection GoUnhandledErrorResult
				_ = resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					sugared.Fatalw("Failed to create service account.",
						"status", resp.StatusCode,
					)
				}
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		sugared.Fatalw("Failed to send requests.",
			mld.LogErr, err,
		)
	}

	total := time.Since(start)

	println(total.String())
}
