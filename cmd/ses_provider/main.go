package main

import (
	"context"
	"log"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.SESConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	logger := setup.CreateLogger(conf.Server)
	logger.InfoContext(ctx, "Starting server...")

	options := setup.ServerOptions{
		Logger: logger,
	}

	server, err := setup.CreateSESProvider(ctx, conf, options)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to setup server.",
			mld.LogErr, err,
		)
	}

	setup.RunServer(ctx, logger, server, conf.Server)
}
