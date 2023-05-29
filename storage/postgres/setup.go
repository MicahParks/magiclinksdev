package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/mod/semver"

	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

var (
	// ErrPostgresSetupCheck is the error returned when the Postgres setup check fails.
	ErrPostgresSetupCheck = errors.New("failed to perform Postgres setup check")
)

// Setup is the JSON data that sits in the setup table.
type Setup struct {
	PlaintextClaims bool   `json:"plaintextClaims,omitempty"`
	PlaintextJWK    bool   `json:"plaintextJWK,omitempty"`
	SemVer          string `json:"semver,omitempty"` // https://pkg.go.dev/golang.org/x/mod/semver
}

// NewWithSetup creates a new Postgres storage and returns its connection pool. It also performs a setup check.
func NewWithSetup(ctx context.Context, config Config) (storage.Storage, *pgxpool.Pool, error) {
	post, p, err := New(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Postgres storage: %w", err)
	}
	tx, err := post.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin Postgres setup transaction: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)
	ctx = context.WithValue(ctx, ctxkey.Tx, tx)
	err = post.(postgres).setupCheck(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to Postgres setup check: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to commit Postgres setup transaction: %w", err)
	}
	return post, p, nil
}

// New creates a new Postgres storage and returns its connection pool.
func New(ctx context.Context, config Config) (storage.Storage, *pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, config.InitialTimeout.Get())
	defer cancel()
	p, err := pool(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Postgres connection pool: %w", err)
	}
	post, err := newPostgres(p, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Postgres storage: %w", err)
	}
	return post, p, nil
}

func compareSemVer(fromConfig, inDatabase string) error {
	config := semver.Canonical(fromConfig)
	database := semver.Canonical(inDatabase)
	validConfig := semver.IsValid(config)
	validDatabase := semver.IsValid(database)
	if !validConfig || !validDatabase {
		return fmt.Errorf("%w: Postgres database and configuration must have a Semantic Version: Configuration %q, Postgres database %q", ErrPostgresSetupCheck, config, database)
	}

	const errFmt = "%w: configuration has Semantic Version %s, but Postgres database has Semantic Version %s: a database migration is likely needed"
	major := semver.Major(config)
	if major != semver.Major(database) {
		return fmt.Errorf(errFmt, ErrPostgresSetupCheck, config, database)
	}

	if major == "v0" {
		const extra = ": development versions must match exactly"
		if semver.Compare(config, database) != 0 {
			return fmt.Errorf(errFmt+extra, ErrPostgresSetupCheck, config, database)
		}
		return nil
	}

	// Compare the minor versions.
	config, database = semver.MajorMinor(config), semver.MajorMinor(database)
	switch v := semver.Compare(config, database); v {
	case -1:
		return nil // Database has newer minor version, which should be backwards compatible.
	case 0:
		return nil
	case 1:
		const extra = ": configuration has newer minor version, server may have newer features incompatible with database"
		return fmt.Errorf(errFmt+extra, ErrPostgresSetupCheck, config, database)
	default:
		return fmt.Errorf("unknown semver comparison result %d: %w", v, ErrPostgresSetupCheck)
	}
}

func pool(ctx context.Context, config Config) (*pgxpool.Pool, error) {
	c, err := pgxpool.ParseConfig(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL DSN: %w", err)
	}
	c.HealthCheckPeriod = config.Health.Get()
	c.MaxConnIdleTime = config.MaxIdle.Get()
	c.MinConns = config.MinConns

	var conn *pgxpool.Pool
	const retries = 5
	for i := 0; i < retries; i++ {
		conn, err = pgxpool.ConnectConfig(ctx, c)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("failed to connect to Postgres after waiting %d seconds: %w", retries, err)
			case <-time.After(time.Second):
			}
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	return conn, nil
}
