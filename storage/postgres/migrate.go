package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/mod/semver"

	mld "github.com/MicahParks/magiclinksdev"
)

const (
	logDescription = "description"
	logFile        = "file"
	logVersion     = "version"
)

var (
	// ErrSemVer indicates a semver is invalid.
	ErrSemVer = errors.New("invalid semver")
)

// Migrator is the interface for applying migrations.
//
// Migrator only migrates in the forward direction. It does not support rolling back migrations. Ensure you have quick
// and robust database backup and restore procedure before running migrations.
//
// No other program should interact with the database while the Migrator is running.
//
// No Migrator implementation should depend on code elsewhere in this project.
type Migrator interface {
	// Migrate applies all migrations in order. It will automatically skip migrations that have .
	Migrate(ctx context.Context) error
}

// MigratorOptions are options for creating a Migrator.
type MigratorOptions struct {
	EncryptionKey [32]byte
	Logger        *slog.Logger
	SetupCtx      context.Context
}

func (m migrator) Migrate(ctx context.Context) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction for migrations: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	setup, err := ReadSetup(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to read setup: %w", err)
	}
	err = compareSemVer(databaseVersion, setup.SemVer)
	if err == nil {
		m.logger.DebugContext(ctx, "No database migrations required.")
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("failed to commit migrations transaction: %w", err)
		}
		return nil
	}

	options := migrationOptions{
		EncryptionKey: m.encryptionKey,
		Logger:        m.logger,
	}

	migrationsApplied := 0
	for _, mig := range m.migrations {
		meta := mig.metadata()
		options.Logger = options.Logger.With(
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.SemVer,
		)

		options.Logger.InfoContext(ctx, "Performing migration.")
		applied, err := mig.migrate(ctx, m.setup, tx, options)
		if err != nil {
			options.Logger.InfoContext(ctx, "Failed to apply migration.",
				mld.LogErr, err,
			)
			return fmt.Errorf("failed to apply migration %q: %w", meta.SemVer, err)
		}

		msg := "Migration not applied."
		if applied {
			msg = "Migration applied."

			m.setup.SemVer = meta.SemVer
			data, err := json.Marshal(m.setup)
			if err != nil {
				return fmt.Errorf("failed to marshal setup after successful migration: %w", err)
			}

			//language=sql
			const query = `
UPDATE mld.setup
SET setup=$1
WHERE id
`
			_, err = tx.Exec(ctx, query, data)
			if err != nil {
				return fmt.Errorf("failed to update setup after successful migration: %w", err)
			}
			migrationsApplied++
		}
		options.Logger.InfoContext(ctx, msg,
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.SemVer,
		)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit migrations transaction: %w", err)
	}

	m.logger.InfoContext(ctx, "Migrations complete.",
		"migrationsApplied", migrationsApplied,
	)

	return nil
}

type migrator struct {
	encryptionKey [32]byte
	logger        *slog.Logger
	migrations    []migration
	pool          *pgxpool.Pool
	setup         Setup
}

// NewMigrator returns a new Migrator for a Postgres storage implementation.
func NewMigrator(pool *pgxpool.Pool, options MigratorOptions) (Migrator, error) {
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	ctx := options.SetupCtx
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
	}

	tx, err := pool.BeginTx(context.Background(), pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction for migrator setup: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	setup, err := ReadSetup(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to read setup: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit migrator setup transaction: %w", err)
	}

	migrations := []migration{
		algMigration{},
	}

	m := migrator{
		encryptionKey: options.EncryptionKey,
		migrations:    migrations,
		pool:          pool,
		setup:         setup,
		logger:        options.Logger,
	}

	return m, nil
}

type metadata struct {
	Description string
	Filename    string
	SemVer      string
}

type migrationOptions struct {
	EncryptionKey [32]byte
	Logger        *slog.Logger
}

type migration interface {
	// metadata returns metadata about the migration.
	metadata() metadata
	// migrate applies the migration. The setup data should be read to determine if the migration should be applied.
	//
	// A storage.Tx can be retrieved from the context.Context under the key ctxkey.Tx.
	migrate(ctx context.Context, setup Setup, tx pgx.Tx, options migrationOptions) (applied bool, err error)
}

func migrationNeeded(migration, setup string) (bool, error) {
	m := semver.Canonical(migration)
	s := semver.Canonical(setup)
	if !semver.IsValid(m) {
		return false, fmt.Errorf("%w: migration version %q is not valid semver", ErrSemVer, m)
	}
	if !semver.IsValid(s) {
		return false, fmt.Errorf("%w: setup version %q is not valid semver", ErrSemVer, s)
	}
	return semver.Compare(m, s) == 1, nil
}
