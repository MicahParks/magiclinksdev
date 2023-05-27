package migrate

import (
	"context"
	"fmt"
	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/magiclinksdev/storage"
	"github.com/MicahParks/magiclinksdev/storage/postgres"
	"github.com/jackc/pgx/v4"
)

const (
	logKID = "kid"
)

// Alg is the migration from database version v0.0.1 to v0.1.0. This is the first database migration.
type Alg struct{}

func (a Alg) Metadata() Metadata {
	return Metadata{
		Description: `This migrates the database from version v0.0.1 to v0.1.0. This is the first database migration. It adds a column to the "mld.jwk" table to identify the key's algorithm. This is to support a new feature of client key selection.`,
		Filename:    "v0.1.0_alg.go",
		SemVer:      "v0.1.0",
	}
}

func (a Alg) Migrate(ctx context.Context, setup postgres.Setup, tx pgx.Tx, options Options) (applied bool, err error) {
	needed, err := migrationNeeded(a.Metadata().SemVer, setup.SemVer)
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
		return false, fmt.Errorf("failed to alter table for %q query: %w", a.Metadata().Filename, err)
	}
	options.Sugared.Debug(`Added "alg" column to "mld.jwk" table.`)

	//language=sql
	query = `
SELECT id, assets FROM mld.jwk
`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to query for existing JSON Web Keys for %q query: %w", a.Metadata().Filename, err)
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
			return false, fmt.Errorf("failed to scan row for %q query: %w", a.Metadata().Filename, err)
		}

		if !setup.PlaintextJWK {
			assets, err = decrypt(options.EncryptionKey, assets)
			if err != nil {
				return false, fmt.Errorf("failed to decrypt assets for %q query: %w", a.Metadata().Filename, err)
			}
		}

		k.meta, err = jwkUnmarshalAssets(assets)
		if err != nil {
			return false, fmt.Errorf("failed to unmarshal assets for %q query: %w", a.Metadata().Filename, err)
		}

		keys = append(keys, k)
	}
	rows.Close()
	options.Sugared.Infof("Found %d existing JSON Web Keys.", len(keys))

	//language=sql
	query = `
UPDATE mld.jwk
SET alg = $1
WHERE id = $2
`

	for _, k := range keys {
		alg := k.meta.ALG
		if alg == "" {
			options.Sugared.Warnw("Found a JSON Web Key with an empty algorithm. Skipping. Client applications will be unable to select this key explicitly.",
				logKID, k.id,
			)
			continue
		}
		_, err = tx.Exec(ctx, query, alg, k.id)
		if err != nil {
			return false, fmt.Errorf("failed to update JSON Web Key for %q query: %w", a.Metadata().Filename, err)
		}
	}

	return true, nil
}
