package main

import (
	"bytes"
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/magiclinksdev/mldtest"
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
		exit(ctx, l, time.Now(), 1)
	}
	u, err = u.Parse("/api/v1/" + network.PathLinkCreate)
	if err != nil {
		l.ErrorContext(ctx, "Failed to parse URL.",
			mld.LogErr, err,
		)
		exit(ctx, l, time.Now(), 1)
	}

	g, ctx := errgroup.WithContext(ctx)

	start := time.Now()
	defer func() {
		l.InfoContext(ctx, "Stress test complete.",
			"duration", time.Since(start).String(),
		)
	}()
	for i := 0; i < 1000; i++ {
		g.Go(func() error {
			for j := 0; j < 1000; j++ {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(linkJSON))
				if err != nil {
					l.ErrorContext(ctx, "Failed to create request.",
						mld.LogErr, err,
					)
					exit(ctx, l, start, 1)
				}

				req.Header.Set(middleware.APIKeyHeader, mldtest.APIKey.String())
				req.Header.Set(mld.HeaderContentType, mld.ContentTypeJSON)

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					l.ErrorContext(ctx, "Failed to send request.",
						mld.LogErr, err,
					)
					exit(ctx, l, start, 1)
				}
				//goland:noinspection GoUnhandledErrorResult
				_ = resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					l.ErrorContext(ctx, "Failed to create link.",
						"status", resp.StatusCode,
					)
					exit(ctx, l, start, 1)
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
		exit(ctx, l, start, 1)
	}

	exit(ctx, l, start, 0)
}

func exit(ctx context.Context, l *slog.Logger, start time.Time, code int) {
	msg := strings.Builder{}
	msg.WriteString("Stress test finish")
	if code == 0 {
		msg.WriteString("ed successfully.")
	} else {
		msg.WriteString("ed with errors.")
	}
	l.InfoContext(ctx, msg.String(),
		"duration", time.Since(start).String(),
	)
	os.Exit(code)
}
