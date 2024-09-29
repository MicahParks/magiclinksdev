package magiclinksdev_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/mldtest"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/setup"
	"github.com/MicahParks/magiclinksdev/storage"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
)

var (
	assets *testAssets
	//go:embed config.test.json
	testConfig []byte
)

type testAssets struct {
	conf setup.TestConfig
	keys []jwkset.JWK
	mux  *http.ServeMux
	sa   model.ServiceAccount
}

func readTestConfig() (setup.TestConfig, error) {
	var conf setup.TestConfig
	err := json.Unmarshal(testConfig, &conf)
	if err != nil {
		return setup.TestConfig{}, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}
	conf, err = conf.DefaultsAndValidate()
	if err != nil {
		return setup.TestConfig{}, fmt.Errorf("failed to validate and apply defaults to confiuration: %w", err)
	}

	return conf, nil
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := log.New(os.Stdout, "", 0)

	conf, err := readTestConfig()
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to read config.", err)
	}

	truncateDatabase(ctx, conf.Storage, logger)

	server, err := setup.CreateTestingProvider(ctx, conf, setup.ServerOptions{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to create server.", err)
	}

	keys := createKeyIfNotExists(ctx, server.Store, logger)

	mux, err := network.CreateHTTPHandlers(server)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to create HTTP handlers.", err)
	}

	sa := model.ServiceAccount{
		UUID:   mldtest.SAUUID,
		APIKey: mldtest.APIKey,
		Aud:    mldtest.Aud,
		Admin:  true,
	}

	assets = &testAssets{
		conf: conf,
		keys: keys,
		mux:  mux,
		sa:   sa,
	}

	os.Exit(m.Run())
}

func createKeyIfNotExists(ctx context.Context, store storage.Storage, logger *log.Logger) []jwkset.JWK {
	tx, err := store.Begin(ctx)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to begin transaction.", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, ctxkey.Tx, tx)

	keys, existed, err := setup.CreateKeysIfNotExists(ctx, store)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to create keys.", err)
	}
	if !existed {
		logger.Fatalf("Keys should have been created by setup.CreateMultiProviderServer function call.")
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to commit transaction.", err)
	}

	return keys
}

func truncateDatabase(ctx context.Context, config postgres.Config, logger *log.Logger) {
	store, _, err := postgres.NewWithSetup(ctx, config, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to create storage.", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer store.Close(ctx)

	tx, err := store.Begin(ctx)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to begin transaction.", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, ctxkey.Tx, tx)

	err = store.TestingTruncate(ctx)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to truncate tables.", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		logger.Fatalf(mld.LogFmt, "Failed to commit transaction.", err)
	}
}
