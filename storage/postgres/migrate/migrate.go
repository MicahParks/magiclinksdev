package migrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/mod/semver"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
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
	SetupTimeout  time.Duration
	Sugared       *zap.SugaredLogger
}

// MigrationOptions hold optional data for database migrations.
type MigrationOptions struct {
	EncryptionKey [32]byte
	Sugared       *zap.SugaredLogger
}

// Migration is a database migration.
type Migration interface {
	// Metadata returns metadata about the migration.
	Metadata() Metadata
	// Migrate applies the migration. The setup data should be read to determine if the migration should be applied.
	//
	// A storage.Tx can be retrieved from the context.Context under the key ctxkey.Tx.
	Migrate(ctx context.Context, setup postgres.Setup, tx pgx.Tx, options MigrationOptions) (applied bool, err error)
}

// Metadata is metadata about a migration.
type Metadata struct {
	Description string
	Filename    string
	SemVer      string
}

type postgresMigrator struct {
	encryptionKey [32]byte
	migrations    []Migration
	pool          *pgxpool.Pool
	setup         postgres.Setup
	sugared       *zap.SugaredLogger
}

func (p postgresMigrator) Migrate(ctx context.Context) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction for migrations: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	options := MigrationOptions{
		EncryptionKey: p.encryptionKey,
		Sugared:       p.sugared,
	}

	for _, migration := range p.migrations {
		meta := migration.Metadata()
		options.Sugared = options.Sugared.With(
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.SemVer,
		)

		options.Sugared.Infow("Performing migration.")
		applied, err := migration.Migrate(ctx, p.setup, tx, options)
		if err != nil {
			options.Sugared.Infow("Failed to apply migration.",
				mld.LogErr, err,
			)
			return fmt.Errorf("failed to apply migration %q: %w", meta.SemVer, err)
		}

		msg := "Migration not applied."
		if applied {
			msg = "Migration applied."

			p.setup.SemVer = meta.SemVer
			data, err := json.Marshal(p.setup)
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
		}
		options.Sugared.Infow(msg,
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.SemVer,
		)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit migrations transaction: %w", err)
	}

	return nil
}

// NewPostgresMigrator returns a new Migrator for a Postgres storage implementation.
func NewPostgresMigrator(pool *pgxpool.Pool, options MigratorOptions) (Migrator, error) {
	timeout := options.SetupTimeout
	if timeout == 0 {
		timeout = time.Second * 30
	}
	if options.Sugared == nil {
		options.Sugared = zap.NewNop().Sugar()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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

	setup, err := postgres.ReadSetup(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to read setup: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit migrator setup transaction: %w", err)
	}

	migrations := []Migration{
		Alg{},
	}

	m := postgresMigrator{
		encryptionKey: options.EncryptionKey,
		migrations:    migrations,
		pool:          pool,
		setup:         setup,
		sugared:       options.Sugared,
	}

	return m, nil
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
