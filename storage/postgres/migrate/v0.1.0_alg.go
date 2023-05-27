package migrate

import (
	"context"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
	"github.com/jackc/pgx/v4"
)

// Alg is the migration to database version v0.1.0.
type Alg struct{}

func (a Alg) Metadata() Metadata {
	return Metadata{
		Description: `This migration brings the database version to v0.1.0. This is the first database migration. It adds a column to the "jwk" table to identify the key's algorithm. This is to support a new feature of client key selection.`,
		Filename:    "v0.1.0_alg.go",
		Version:     "v0.1.0",
	}
}

func (a Alg) Migrate(ctx context.Context, setup postgres.Setup, tx pgx.Tx) (applied bool, err error) {
	return false, nil // TODO
}
