package main

import (
	"context"
	"log"
	"os"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.NopConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	logger := setup.CreateLogger(conf.Server)
	logger.InfoContext(ctx, "Starting server...")

	options := setup.ServerOptions{
		Logger: logger,
	}

	server, err := setup.CreateNopProviderServer(ctx, conf, options)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to setup server.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	setup.RunServer(ctx, logger, server)
}
