package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.MultiConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	logger := setup.CreateLogger(conf.Server)
	logger.InfoContext(ctx, "Starting server...")

	middlewareHook := handle.MiddlewareHookFunc(func(options handle.MiddlewareOptions) handle.MiddlewareOptions {
		// Filter out the middleware hooks by HTTP path.
		if options.Path == network.PathJWTCreate {
			options.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				// Overwrite the HTTP handler to return a 404, disabling the endpoint.
				writer.WriteHeader(http.StatusNotFound)
			})
		}

		// Add a custom logging middleware to all endpoints.
		originalHandler := options.Handler
		options.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
			logger.InfoContext(ctx, "Got a request. It passed all built-in middleware.",
				"method", request.Method,
				"path", request.URL.Path,
			)
			originalHandler.ServeHTTP(writer, request)
		})

		return options
	})

	options := setup.ServerOptions{
		Logger:         logger,
		MiddlewareHook: middlewareHook,
	}

	server, err := setup.CreateMultiProviderServer(ctx, conf, options)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to setup server.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	setup.RunServer(ctx, logger, server, conf.Server)
}
