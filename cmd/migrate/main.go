package main

import (
	"context"
	"log"
	"os"

	jt "github.com/MicahParks/jsontype"
	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/setup"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.MultiConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	var logger *zap.Logger
	if os.Getenv("DEV_MODE") == "true" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to create logger.", err)
	}
	sugared := logger.Sugar()

	_, pool, err := postgres.New(ctx, conf.Storage)
	if err != nil {
		sugared.Fatalw("Failed to create postgres pool.",
			mld.LogErr, err,
		)
	}

	k, err := postgres.DecodeAES256Base64(conf.Storage.AES256KeyBase64)
	if err != nil {
		sugared.Fatalw("Failed to decode AES256 key.",
			mld.LogErr, err,
		)
	}
	options := postgres.MigratorOptions{
		EncryptionKey: k,
		Sugared:       sugared,
	}
	migrator, err := postgres.NewMigrator(pool, options)
	if err != nil {
		sugared.Fatalw("Failed to create migrator.",
			mld.LogErr, err,
		)
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		sugared.Fatalw("Failed to migrate.",
			mld.LogErr, err,
		)
	}
}
