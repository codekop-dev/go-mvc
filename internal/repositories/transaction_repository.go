package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/codekop/gomvc/internal/models"
)

type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates the read-only penjualan repository.
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Latest reads simple transaction columns ordered by newest id.
func (r *TransactionRepository) Latest(ctx context.Context, limit int) ([]models.Transaction, error) {
	return r.ListPage(ctx, limit, 0)
}

// ListPage fetches one page of simple transactions.
func (r *TransactionRepository) ListPage(ctx context.Context, limit, offset int) ([]models.Transaction, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, no_trx, total, created_at
		FROM penjualan
		ORDER BY id DESC
		LIMIT %d OFFSET %d`, limit, offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Transaction, 0, limit)
	for rows.Next() {
		var item models.Transaction
		if err := rows.Scan(&item.ID, &item.Number, &item.Total, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Count returns the total number of simple transactions.
func (r *TransactionRepository) Count(ctx context.Context) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database is not configured")
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM penjualan").Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
