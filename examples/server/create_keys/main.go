package main

import (
	"context"
	"time"

	jt "github.com/MicahParks/jsontype"
	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/setup"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	sugared := logger.Sugar()

	config, err := jt.Read[setup.MultiConfig]()
	if err != nil {
		sugared.Fatalw("Failed to read config.",
			mld.LogErr, err,
		)
	}

	store, _, err := postgres.New(ctx, config.Storage)
	if err != nil {
		sugared.Fatalw("Failed to create storage.",
			mld.LogErr, err,
		)
	}

	tx, err := store.Begin(ctx)
	if err != nil {
		sugared.Fatalw("Failed to begin transaction.",
			mld.LogErr, err,
		)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, ctxkey.Tx, tx)

	_, _, err = setup.CreateKeysIfNotExists(ctx, store)
	if err != nil {
		sugared.Fatalw("Failed to truncate database.",
			mld.LogErr, err,
		)
	}

	err = tx.Commit(ctx)
	if err != nil {
		sugared.Fatalw("Failed to commit transaction.",
			mld.LogErr, err,
		)
	}
}
