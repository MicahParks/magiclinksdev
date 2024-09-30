package main

import (
	"context"
	"log"
	"os"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/setup"
	"github.com/MicahParks/magiclinksdev/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.MultiConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	logger := setup.CreateLogger(conf.Server)

	_, pool, err := storage.New(ctx, conf.Storage)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create postgres pool.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	k, err := storage.DecodeAES256Base64(conf.Storage.AES256KeyBase64)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to decode AES256 key.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	options := storage.MigratorOptions{
		EncryptionKey: k,
		Logger:        logger,
	}
	migrator, err := storage.NewMigrator(pool, options)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create migrator.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to migrate.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
}
