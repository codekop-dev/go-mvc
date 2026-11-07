package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codekop/gomvc/internal/helper"
)

// TestCRUDHelperWithoutDatabase verifies safe behavior before DB configuration.
func TestCRUDHelperWithoutDatabase(t *testing.T) {
	if _, err := helper.New(nil, "mysql", "items; DROP TABLE users", []string{"id", "name"}); err == nil {
		t.Fatal("expected unsafe table name to be rejected")
	}
	crudHelper, err := helper.New(nil, "mysql", "items", []string{"id", "name"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := crudHelper.List(context.Background(), 20); !errors.Is(err, helper.ErrNotConfigured) {
		t.Fatalf("error = %v, want ErrNotConfigured", err)
	}
}
