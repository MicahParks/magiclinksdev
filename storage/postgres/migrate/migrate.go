package migrate

import (
	"context"
	"fmt"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"time"
)

const (
	logDescription = "description"
	logFile        = "file"
	logVersion     = "version"
)

// Migrator is the interface for applying migrations.
//
// Migrator only migrates in the forward direction. It does not support rolling back migrations. Ensure you have quick
// and robust database backup and restore procedure before running migrations.
//
// No other program should interact with the database while the Migrator is running.
type Migrator interface {
	// Migrate applies all migrations in order. It will automatically skip migrations that have .
	Migrate(ctx context.Context) error
}

// MigratorOptions are options for creating a Migrator.
type MigratorOptions struct {
	SetupTimeout time.Duration
	Sugared      *zap.SugaredLogger
}

// Migration is a database migration.
type Migration interface {
	// Metadata returns metadata about the migration.
	Metadata() Metadata
	// Migrate applies the migration. The setup data should be read to determine if the migration should be applied.
	//
	// A storage.Tx can be retrieved from the context.Context under the key ctxkey.Tx.
	Migrate(ctx context.Context, setup postgres.Setup, tx pgx.Tx) (applied bool, err error)
}

// Metadata is metadata about a migration.
type Metadata struct {
	Description string
	Filename    string
	Version     string
}

type migrator struct {
	migrations []Migration
	pool       *pgxpool.Pool
	setup      postgres.Setup
	Sugared    *zap.SugaredLogger
}

func (m migrator) Migrate(ctx context.Context) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to create transaction for migrations: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer tx.Rollback(ctx)

	for _, migration := range m.migrations {
		meta := migration.Metadata()
		m.Sugared.Infow("Performing migration.",
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.Version,
		)
		applied, err := migration.Migrate(ctx, m.setup, tx)
		if err != nil {
			return fmt.Errorf("failed to apply migration %q: %w", meta.Version, err)
		}
		msg := "Migration not applied."
		if applied {
			msg = "Migration applied."
		}
		m.Sugared.Infow(msg,
			logDescription, meta.Description,
			logFile, meta.Filename,
			logVersion, meta.Version,
		)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit migrations transaction: %w", err)
	}

	return nil
}

// New returns a new Migrator.
func New(pool *pgxpool.Pool, options MigratorOptions) (Migrator, error) {
	timeout := options.SetupTimeout
	if timeout == 0 {
		timeout = time.Second * 30
	}
	if options.Sugared == nil {
		options.Sugared = zap.NewNop().Sugar()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tx, err := pool.Begin(context.Background())
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

	m := migrator{
		migrations: nil, // TODO Populate.
		pool:       pool,
		setup:      setup,
		Sugared:    options.Sugared,
	}

	return m, nil
}
