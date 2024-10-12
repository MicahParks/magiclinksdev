package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// otpMigration is the migration from database version v0.1.0 to v0.2.0. This is the first database migration.
type otpMigration struct{}

func (a otpMigration) metadata() metadata {
	return metadata{
		Description: `This migrates the database from version v0.1.0 to v0.2.0. This is the second database migration. It adds the "mld.otp" table for One-Time Password support.`,
		Filename:    "v0.2.0_otp.go",
		SemVer:      "v0.2.0",
	}
}

func (a otpMigration) migrate(ctx context.Context, setup Setup, tx pgx.Tx, options migrationOptions) (applied bool, err error) {
	needed, err := migrationNeeded(a.metadata().SemVer, setup.SemVer)
	if err != nil {
		return false, fmt.Errorf("failed to determine if migration is needed: %w", err)
	}
	if !needed {
		return false, nil
	}

	//language=sql
	query := `
CREATE TABLE mld.otp
(
    id        BIGSERIAL PRIMARY KEY,
    sa_id     BIGINT                   NOT NULL REFERENCES mld.service_account (id),
    expires   TIMESTAMP WITH TIME ZONE NOT NULL,
    id_public UUID                     NOT NULL UNIQUE,
    otp       TEXT                     NOT NULL,
    used      TIMESTAMP WITH TIME ZONE,
    created   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
)
`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to create new table for %q query: %w", a.metadata().Filename, err)
	}
	options.Logger.DebugContext(ctx, `Added "mld.otp" table.`)

	//language=sql
	query = `
CREATE INDEX ON mld.jwk (alg)
`
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to create index for %q query: %w", a.metadata().Filename, err)
	}
	options.Logger.DebugContext(ctx, `Created index on "alg" column of "mld.jwk" table.`)

	indexes := []string{
		"CREATE INDEX ON mld.otp (sa_id)",
		"CREATE INDEX ON mld.otp (expires)",
		"CREATE INDEX ON mld.otp (id_public)",
		"CREATE INDEX ON mld.otp (used)",
		"CREATE INDEX ON mld.otp (created)",
	}
	for _, index := range indexes {
		_, err = tx.Exec(ctx, index)
		if err != nil {
			return false, fmt.Errorf("failed to create index for %q: %q, %w", a.metadata().Filename, index, err)
		}
		options.Logger.DebugContext(ctx, fmt.Sprintf(`Created index on "mld.otp" table: %q.`, index))
	}

	return true, nil
}
