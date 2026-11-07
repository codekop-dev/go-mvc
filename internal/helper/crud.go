package helper

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var identifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

var (
	ErrNotConfigured = errors.New("database is not configured")
	ErrNoFields      = errors.New("no fields supplied")
	ErrInvalidField  = errors.New("invalid CRUD field")
)

type Helper struct {
	db      *sql.DB
	dialect string
	table   string
	columns map[string]struct{}
}

// New creates a safe CRUD helper. Table and column names are allow-listed once;
// values are always sent through database/sql parameters.
func New(db *sql.DB, driver, table string, columns []string) (*Helper, error) {
	if !identifierPattern.MatchString(table) {
		return nil, fmt.Errorf("invalid table name %q", table)
	}
	allowed := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		if !identifierPattern.MatchString(column) {
			return nil, fmt.Errorf("invalid column name %q", column)
		}
		allowed[column] = struct{}{}
	}
	if _, ok := allowed["id"]; !ok {
		return nil, fmt.Errorf("CRUD table must include id")
	}
	if driver != "postgres" {
		driver = "mysql"
	}
	return &Helper{db: db, dialect: driver, table: table, columns: allowed}, nil
}

// List returns the newest rows from the allow-listed table.
func (h *Helper) List(ctx context.Context, limit int) ([]map[string]any, error) {
	return h.ListOffset(ctx, limit, 0)
}

// ListOffset returns rows with both limit and offset for pagination.
func (h *Helper) ListOffset(ctx context.Context, limit, offset int) ([]map[string]any, error) {
	if err := h.ready(); err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s DESC LIMIT %s OFFSET %s", h.selectColumns(), h.quote(h.table), h.quote("id"), h.placeholder(1), h.placeholder(2))
	rows, err := h.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// Count returns the total number of rows in the allow-listed table.
func (h *Helper) Count(ctx context.Context) (int64, error) {
	if err := h.ready(); err != nil {
		return 0, err
	}
	var total int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", h.quote(h.table))
	if err := h.db.QueryRowContext(ctx, query).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// Find returns one row by its integer id or sql.ErrNoRows.
func (h *Helper) Find(ctx context.Context, id int64) (map[string]any, error) {
	if err := h.ready(); err != nil {
		return nil, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s LIMIT 1", h.selectColumns(), h.quote(h.table), h.quote("id"), h.placeholder(1))
	rows, err := h.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	return items[0], nil
}

// Insert creates a row using sorted allow-listed fields and bound values.
func (h *Helper) Insert(ctx context.Context, values map[string]any) (int64, error) {
	if err := h.ready(); err != nil {
		return 0, err
	}
	keys, err := h.keys(values)
	if err != nil {
		return 0, err
	}
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	quoted := make([]string, len(keys))
	for i, key := range keys {
		quoted[i] = h.quote(key)
		placeholders[i] = h.placeholder(i + 1)
		args[i] = values[key]
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", h.quote(h.table), strings.Join(quoted, ", "), strings.Join(placeholders, ", "))
	if h.dialect == "postgres" {
		query += " RETURNING id"
		var id int64
		if err := h.db.QueryRowContext(ctx, query, args...).Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}
	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update changes a row by id and reports whether a row was affected.
func (h *Helper) Update(ctx context.Context, id int64, values map[string]any) (bool, error) {
	if err := h.ready(); err != nil {
		return false, err
	}
	keys, err := h.keys(values)
	if err != nil {
		return false, err
	}
	assignments := make([]string, len(keys))
	args := make([]any, 0, len(keys)+1)
	for i, key := range keys {
		assignments[i] = fmt.Sprintf("%s = %s", h.quote(key), h.placeholder(i+1))
		args = append(args, values[key])
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", h.quote(h.table), strings.Join(assignments, ", "), h.quote("id"), h.placeholder(len(args)))
	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// Delete removes a row by id and reports whether a row was affected.
func (h *Helper) Delete(ctx context.Context, id int64) (bool, error) {
	if err := h.ready(); err != nil {
		return false, err
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", h.quote(h.table), h.quote("id"), h.placeholder(1))
	result, err := h.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// ready prevents CRUD calls when the application intentionally has no database.
func (h *Helper) ready() error {
	if h.db == nil {
		return ErrNotConfigured
	}
	return nil
}

// keys validates and sorts writable columns before SQL is assembled.
func (h *Helper) keys(values map[string]any) ([]string, error) {
	if len(values) == 0 {
		return nil, ErrNoFields
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if _, ok := h.columns[key]; !ok || key == "id" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidField, key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys, nil
}

// selectColumns returns deterministic, quoted columns for SELECT statements.
func (h *Helper) selectColumns() string {
	keys := make([]string, 0, len(h.columns))
	for key := range h.columns {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	quoted := make([]string, len(keys))
	for i, key := range keys {
		quoted[i] = h.quote(key)
	}
	return strings.Join(quoted, ", ")
}

// quote applies the identifier quoting style for the active SQL dialect.
func (h *Helper) quote(identifier string) string {
	if h.dialect == "postgres" {
		return `"` + identifier + `"`
	}
	return "`" + identifier + "`"
}

// placeholder returns ? for MySQL or $N for PostgreSQL.
func (h *Helper) placeholder(index int) string {
	if h.dialect == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// scanRows converts database rows into generic maps for a reusable CRUD helper.
func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		item := make(map[string]any, len(columns))
		for i, column := range columns {
			if value, ok := values[i].([]byte); ok {
				item[column] = string(value)
			} else {
				item[column] = values[i]
			}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
