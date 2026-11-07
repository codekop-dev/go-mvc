package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/codekop/gomvc/internal/helper"
)

// TestDatabaseTransactionWithoutDatabase verifies the transaction guard.
func TestDatabaseTransactionWithoutDatabase(t *testing.T) {
	err := helper.WithTransaction(context.Background(), nil, func(_ *sql.Tx) error { return nil })
	if !errors.Is(err, helper.ErrNotConfigured) {
		t.Fatalf("error = %v, want ErrNotConfigured", err)
	}
}
