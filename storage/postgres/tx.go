package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Transaction is a storage.Tx implementation for Postgres.
type Transaction struct {
	Tx pgx.Tx
}

// Commit helps implement the storage.Tx interface.
func (t *Transaction) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

// Rollback helps implement the storage.Tx interface.
func (t *Transaction) Rollback(ctx context.Context) error {
	err := t.Tx.Rollback(ctx)
	if err != nil {
		err = fmt.Errorf("failed to rollback Postgres transaction: %w", err)
	}
	return err
}
