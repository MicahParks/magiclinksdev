package main

import (
	"bytes"
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

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

	l := slog.Default()

	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		l.ErrorContext(ctx, "Failed to parse URL.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	u, err = u.Parse(network.PathLinkCreate)
	if err != nil {
		l.ErrorContext(ctx, "Failed to parse URL.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	g, ctx := errgroup.WithContext(ctx)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		g.Go(func() error {
			for j := 0; j < 1000; j++ {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(linkJSON))
				if err != nil {
					l.ErrorContext(ctx, "Failed to create request.",
						mld.LogErr, err,
					)
					os.Exit(1)
				}

				req.Header.Set(middleware.APIKeyHeader, "40084740-0bc3-455d-b298-e23a31561580") // Admin API key from config.

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					l.ErrorContext(ctx, "Failed to send request.",
						mld.LogErr, err,
					)
					os.Exit(1)
				}
				//goland:noinspection GoUnhandledErrorResult
				_ = resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					l.ErrorContext(ctx, "Failed to create service account.",
						"status", resp.StatusCode,
					)
					os.Exit(1)
				}
			}
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		l.ErrorContext(ctx, "Failed to send requests.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	total := time.Since(start)

	println(total.String())
}
