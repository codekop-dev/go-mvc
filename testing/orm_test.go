package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codekop/gomvc/internal/helper"
)

// TestORMBuilderWithoutDatabase verifies query builder validation and safe DB failure.
func TestORMBuilderWithoutDatabase(t *testing.T) {
	orm := helper.NewORM(nil, "mysql")
	if _, err := orm.Table("items; DROP TABLE users").Get(context.Background()); err == nil {
		t.Fatal("expected invalid table identifier")
	}
	result, err := orm.Table("items").Where("id", 1).Get(context.Background())
	if result != nil || !errors.Is(err, helper.ErrNotConfigured) {
		t.Fatalf("result = %#v, error = %v, want ErrNotConfigured", result, err)
	}
}
