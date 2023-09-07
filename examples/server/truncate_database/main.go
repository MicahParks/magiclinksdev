package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/setup"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := slog.Default()

	conf, err := jt.Read[setup.MultiConfig]()
	if err != nil {
		logger.ErrorContext(ctx, "Failed to read configuration.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	store, _, err := postgres.NewWithSetup(ctx, conf.Storage, logger.With("postgresSetup", true))
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create storage.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	tx, err := store.Begin(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to begin transaction.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, ctxkey.Tx, tx)

	err = store.TestingTruncate(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to truncate database.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to commit transaction.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
}
