package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	jt "github.com/MicahParks/jsontype"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/setup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := jt.Read[setup.NopConfig]()
	if err != nil {
		log.Fatalf(mld.LogFmt, "Failed to read configuration.", err)
	}

	logger := slog.Default()
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

	mux, err := network.CreateHTTPHandlers(server)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create HTTP handlers.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.Server.Port),
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})
	go serverShutdown(ctx, conf.Server, logger, idleConnsClosed, httpServer)

	logger.InfoContext(ctx, "Server is listening.",
		"port", conf.Server.Port,
	)
	err = httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "Failed to listen and serve.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	<-idleConnsClosed
}

func serverShutdown(ctx context.Context, conf config.Config, logger *slog.Logger, idleConnsClosed chan struct{}, srv *http.Server) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutdown:
		logger.InfoContext(ctx, "Got a SIGINT or SIGTERM.")
	case <-ctx.Done():
		logger.InfoContext(ctx, "Context over.",
			mld.LogErr, ctx.Err(),
		)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout.Get())
	defer cancel()
	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.InfoContext(ctx, "Couldn't shut down server before time ended.",
			mld.LogErr, err,
		)
	}

	close(idleConnsClosed)
}
