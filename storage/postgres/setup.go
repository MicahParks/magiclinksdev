package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/mod/semver"

	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	databaseVersion = "v0.1.0"
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
func NewWithSetup(ctx context.Context, config Config, setupLogger *slog.Logger) (storage.Storage, *pgxpool.Pool, error) {
	post, p, err := New(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Postgres storage: %w", err)
	}
	if config.AutoMigrate {
		encryptionKey, err := DecodeAES256Base64(config.AES256KeyBase64)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode AES256 key: %w", err)
		}
		options := MigratorOptions{
			EncryptionKey: encryptionKey,
			SetupCtx:      ctx,
			Logger:        setupLogger,
		}
		m, err := NewMigrator(p, options)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create Postgres migrator: %w", err)
		}
		err = m.Migrate(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to migrate Postgres database: %w", err)
		}
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

func compareSemVer(programSemVer, databaseSemVer string) error {
	program := semver.Canonical(programSemVer)
	database := semver.Canonical(databaseSemVer)
	validProgram := semver.IsValid(program)
	validDatabase := semver.IsValid(database)
	if !validProgram || !validDatabase {
		return fmt.Errorf("%w: Postgres database and Go program must have a Semantic Version: Go program %q, Postgres database %q", ErrPostgresSetupCheck, program, database)
	}

	const errFmt = "%w: Go program has Semantic Version %s, but Postgres database has Semantic Version %s: a database migration is likely needed"
	major := semver.Major(program)
	if major != semver.Major(database) {
		return fmt.Errorf(errFmt, ErrPostgresSetupCheck, program, database)
	}

	if major == "v0" {
		const extra = ": development versions must match exactly"
		if semver.Compare(program, database) != 0 {
			return fmt.Errorf(errFmt+extra, ErrPostgresSetupCheck, program, database)
		}
		return nil
	}

	// Compare the minor versions.
	program, database = semver.MajorMinor(program), semver.MajorMinor(database)
	switch v := semver.Compare(program, database); v {
	case -1:
		return nil // Database has newer minor version, which should be backwards compatible.
	case 0:
		return nil
	case 1:
		const extra = ": Go program has newer minor version, it may have newer features incompatible with database"
		return fmt.Errorf(errFmt+extra, ErrPostgresSetupCheck, program, database)
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
		conn, err = pgxpool.NewWithConfig(ctx, c)
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
