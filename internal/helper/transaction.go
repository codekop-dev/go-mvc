package helper

import (
	"context"
	"database/sql"
)

// WithTransaction starts a database transaction, rolls back on any error, and commits on success.
func WithTransaction(ctx context.Context, db *sql.DB, work func(*sql.Tx) error) error {
	if db == nil {
		return ErrNotConfigured
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if err := work(tx); err != nil {
		return err
	}
	return tx.Commit()
}
