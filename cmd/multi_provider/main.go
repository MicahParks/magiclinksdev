package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	jt "github.com/MicahParks/jsontype"
	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/setup"
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
	logger.Info("Starting server...")
	sugared := logger.Sugar()

	options := setup.ServerOptions{
		Sugared: sugared,
	}

	server, err := setup.CreateMultiProviderServer(ctx, conf, options)
	if err != nil {
		sugared.Fatalw("Failed to setup server.",
			mld.LogErr, err,
		)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux, err := network.CreateHTTPHandlers(server)
	if err != nil {
		sugared.Fatalw("Failed to create HTTP handlers.",
			mld.LogErr, err,
		)
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})
	go serverShutdown(ctx, conf.Server, sugared, idleConnsClosed, httpServer)

	sugared.Infow("Server is listening.",
		"port", port,
	)
	err = httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		sugared.Fatalw("Failed to listen and serve.",
			mld.LogErr, err,
		)
	}

	<-idleConnsClosed
}

func serverShutdown(ctx context.Context, conf config.Config, sugared *zap.SugaredLogger, idleConnsClosed chan struct{}, srv *http.Server) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutdown:
		sugared.Info("Got a SIGINT or SIGTERM.")
	case <-ctx.Done():
		sugared.Infow("Context over.",
			mld.LogErr, ctx.Err(),
		)
	}

	err := sugared.Sync()
	if err != nil {
		log.Printf(mld.LogFmt, "Failed to sync logger on server shutdown.", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout.Get())
	defer cancel()
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		sugared.Infow("Couldn't shut down server before time ended.",
			mld.LogErr, err,
		)
	}

	close(idleConnsClosed)
}
