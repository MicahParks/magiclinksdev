package postgres

import (
	"context"
	"fmt"

	"github.com/MicahParks/jwkset"
	"github.com/jackc/pgx/v4"

	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	logKID = "kid"
)

// algMigration is the migration from database version v0.0.1 to v0.1.0. This is the first database migration.
type algMigration struct{}

func (a algMigration) metadata() metadata {
	return metadata{
		Description: `This migrates the database from version v0.0.1 to v0.1.0. This is the first database migration. It adds a column to the "mld.jwk" table to identify the key's algorithm. This is to support a new feature of client key selection.`,
		Filename:    "v0.1.0_alg.go",
		SemVer:      "v0.1.0",
	}
}

func (a algMigration) migrate(ctx context.Context, setup Setup, tx pgx.Tx, options migrationOptions) (applied bool, err error) {
	needed, err := migrationNeeded(a.metadata().SemVer, setup.SemVer)
	if err != nil {
		return false, fmt.Errorf("failed to determine if migration is needed: %w", err)
	}
	if !needed {
		return false, nil
	}

	//language=sql
	query := `
ALTER TABLE mld.jwk
    ADD COLUMN alg TEXT NOT NULL DEFAULT ''
`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to alter table for %q query: %w", a.metadata().Filename, err)
	}
	options.Logger.DebugContext(ctx, `Added "alg" column to "mld.jwk" table.`)

	//language=sql
	query = `
CREATE INDEX ON mld.jwk (alg)
`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to create index for %q query: %w", a.metadata().Filename, err)
	}
	options.Logger.DebugContext(ctx, `Created index on "alg" column of "mld.jwk" table.`)

	//language=sql
	query = `
SELECT id, assets FROM mld.jwk
`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to query for existing JSON Web Keys for %q query: %w", a.metadata().Filename, err)
	}
	defer rows.Close()

	type key struct {
		id   int64
		meta jwkset.KeyWithMeta[storage.JWKSetCustomKeyMeta]
	}
	keys := make([]key, 0)
	for rows.Next() {
		var k key
		var assets []byte
		err = rows.Scan(&k.id, &assets)
		if err != nil {
			return false, fmt.Errorf("failed to scan row for %q query: %w", a.metadata().Filename, err)
		}

		k.meta, err = jwkUnmarshalAssets(options.EncryptionKey, assets, setup.PlaintextJWK)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal JSON Web Key for %q query: %w", a.metadata().Filename, err)
		}

		keys = append(keys, k)
	}
	rows.Close()
	options.Logger.InfoContext(ctx, "Found existing JSON Web Keys.",
		"existingKeys", len(keys),
	)

	//language=sql
	query = `
UPDATE mld.jwk
SET alg = $1
WHERE id = $2
`

	for _, k := range keys {
		alg := k.meta.ALG
		if alg == "" {
			options.Logger.WarnContext(ctx, "Found a JSON Web Key with an empty algorithm. Skipping. Client applications will be unable to select this key explicitly.",
				logKID, k.id,
			)
			continue
		}
		_, err = tx.Exec(ctx, query, alg, k.id)
		if err != nil {
			return false, fmt.Errorf("failed to update JSON Web Key for %q query: %w", a.metadata().Filename, err)
		}
	}

	return true, nil
}
