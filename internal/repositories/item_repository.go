package repositories

import (
	"context"
	"database/sql"

	"github.com/codekop/gomvc/internal/helper"
	"github.com/codekop/gomvc/internal/models"
)

type ItemRepository struct {
	crud *helper.Helper
}

// NewItemRepository creates the sample_items repository and its CRUD allow-list.
func NewItemRepository(db *sql.DB, driver string) *ItemRepository {
	crudHelper, err := helper.New(db, driver, "sample_items", []string{"id", "name", "total", "created_at"})
	if err != nil {
		panic(err)
	}
	return &ItemRepository{crud: crudHelper}
}

// List fetches sample items through the generic CRUD helper.
func (r *ItemRepository) List(ctx context.Context, limit int) ([]models.Item, error) {
	return r.ListPage(ctx, limit, 0)
}

// ListPage fetches one page of sample items.
func (r *ItemRepository) ListPage(ctx context.Context, limit, offset int) ([]models.Item, error) {
	rows, err := r.crud.ListOffset(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]models.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.ItemFromMap(row))
	}
	return items, nil
}

// Count returns the total number of sample items.
func (r *ItemRepository) Count(ctx context.Context) (int64, error) {
	return r.crud.Count(ctx)
}

// Find fetches one sample item by id.
func (r *ItemRepository) Find(ctx context.Context, id int64) (models.Item, error) {
	row, err := r.crud.Find(ctx, id)
	if err != nil {
		return models.Item{}, err
	}
	return models.ItemFromMap(row), nil
}

// Create inserts the writable fields of a sample item.
func (r *ItemRepository) Create(ctx context.Context, name string, total float64) (int64, error) {
	return r.crud.Insert(ctx, map[string]any{"name": name, "total": total})
}

// Update applies allow-listed changes to one sample item.
func (r *ItemRepository) Update(ctx context.Context, id int64, values map[string]any) (bool, error) {
	return r.crud.Update(ctx, id, values)
}

// Delete removes one sample item by id.
func (r *ItemRepository) Delete(ctx context.Context, id int64) (bool, error) {
	return r.crud.Delete(ctx, id)
}
