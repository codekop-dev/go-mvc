package helper

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// ORM is a small CodeIgniter 3-style query facade over database/sql.
type ORM struct {
	db      *sql.DB
	dialect string
}

// QueryBuilder stores one request-scoped SELECT query and is safe to chain.
type QueryBuilder struct {
	orm       *ORM
	table     string
	columns   []string
	wheres    []condition
	orderBy   string
	direction string
	limit     int
	offset    int
	err       error
}

type condition struct {
	column string
	value  any
}

// Result contains rows and exposes familiar result_array/row_array helpers.
type Result struct {
	rows []map[string]any
}

// NewORM creates a request-safe query facade for MySQL/MariaDB or PostgreSQL.
func NewORM(db *sql.DB, driver string) *ORM {
	if driver != "postgres" {
		driver = "mysql"
	}
	return &ORM{db: db, dialect: driver}
}

// Table starts a query for a table, equivalent to CodeIgniter's from().
func (o *ORM) Table(table string) *QueryBuilder {
	return (&QueryBuilder{orm: o}).From(table)
}

// From starts a query without first calling Select().
func (o *ORM) From(table string) *QueryBuilder {
	return o.Table(table)
}

// Select starts a query with comma-separated safe column names or *.
func (o *ORM) Select(columns string) *QueryBuilder {
	return (&QueryBuilder{orm: o}).Select(columns)
}

// Query executes application-owned raw SQL with bound values, like PDO query().
func (o *ORM) Query(ctx context.Context, query string, args ...any) (*Result, error) {
	if err := o.ready(); err != nil {
		return nil, err
	}
	rows, err := o.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanRows(rows)
	if err != nil {
		return nil, err
	}
	return &Result{rows: items}, nil
}

// GetWhere builds a simple equality query from a table and conditions.
func (o *ORM) GetWhere(ctx context.Context, table string, where map[string]any) (*Result, error) {
	query := o.Table(table)
	keys := sortedMapKeys(where)
	for _, key := range keys {
		query.Where(key, where[key])
	}
	return query.Get(ctx)
}

// Insert creates a row using bound values and returns its generated id.
func (o *ORM) Insert(ctx context.Context, table string, values map[string]any) (int64, error) {
	if err := o.ready(); err != nil {
		return 0, err
	}
	if err := validIdentifier(table); err != nil {
		return 0, err
	}
	keys, err := writableKeys(values)
	if err != nil {
		return 0, err
	}
	columns := make([]string, len(keys))
	holders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, key := range keys {
		columns[i] = o.quote(key)
		holders[i] = o.placeholder(i + 1)
		args[i] = values[key]
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", o.quote(table), strings.Join(columns, ", "), strings.Join(holders, ", "))
	if o.dialect == "postgres" {
		var id int64
		if err := o.db.QueryRowContext(ctx, query+" RETURNING id", args...).Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}
	result, err := o.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update changes a row using set values and equality conditions.
func (o *ORM) Update(ctx context.Context, table string, values, where map[string]any) (bool, error) {
	if err := o.ready(); err != nil {
		return false, err
	}
	if err := validIdentifier(table); err != nil {
		return false, err
	}
	setKeys, err := writableKeys(values)
	if err != nil {
		return false, err
	}
	whereKeys := sortedMapKeys(where)
	if len(whereKeys) == 0 {
		return false, fmt.Errorf("update requires where conditions")
	}
	args := make([]any, 0, len(setKeys)+len(whereKeys))
	assignments := make([]string, len(setKeys))
	for i, key := range setKeys {
		assignments[i] = fmt.Sprintf("%s = %s", o.quote(key), o.placeholder(i+1))
		args = append(args, values[key])
	}
	conditions, conditionArgs := o.conditions(whereKeys, where, len(args))
	args = append(args, conditionArgs...)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", o.quote(table), strings.Join(assignments, ", "), strings.Join(conditions, " AND "))
	result, err := o.db.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// Delete removes rows matching equality conditions.
func (o *ORM) Delete(ctx context.Context, table string, where map[string]any) (bool, error) {
	if err := o.ready(); err != nil {
		return false, err
	}
	if err := validIdentifier(table); err != nil {
		return false, err
	}
	keys := sortedMapKeys(where)
	if len(keys) == 0 {
		return false, fmt.Errorf("delete requires where conditions")
	}
	conditions, args := o.conditions(keys, where, 0)
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", o.quote(table), strings.Join(conditions, " AND "))
	result, err := o.db.ExecContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

// Select chooses columns for the current builder.
func (q *QueryBuilder) Select(columns string) *QueryBuilder {
	if q.err != nil {
		return q
	}
	for _, column := range strings.Split(columns, ",") {
		column = strings.TrimSpace(column)
		if column == "*" {
			q.columns = []string{"*"}
			continue
		}
		if err := validIdentifier(column); err != nil {
			q.err = err
			return q
		}
		q.columns = append(q.columns, column)
	}
	return q
}

// From sets the table for the current builder.
func (q *QueryBuilder) From(table string) *QueryBuilder {
	if err := validIdentifier(table); err != nil {
		q.err = err
		return q
	}
	q.table = table
	return q
}

// Where adds an equality condition to the current builder.
func (q *QueryBuilder) Where(column string, value any) *QueryBuilder {
	if err := validIdentifier(column); err != nil {
		q.err = err
		return q
	}
	q.wheres = append(q.wheres, condition{column: column, value: value})
	return q
}

// OrderBy adds a safe ORDER BY clause to the current builder.
func (q *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	if err := validIdentifier(column); err != nil {
		q.err = err
		return q
	}
	direction = strings.ToUpper(strings.TrimSpace(direction))
	if direction != "ASC" && direction != "DESC" {
		q.err = fmt.Errorf("order direction must be ASC or DESC")
		return q
	}
	q.orderBy = column
	q.direction = direction
	return q
}

// Limit caps the number of rows returned by the current builder.
func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	if limit < 1 {
		limit = 1
	}
	if limit > 1000 {
		limit = 1000
	}
	q.limit = limit
	return q
}

// Offset skips rows before returning results.
func (q *QueryBuilder) Offset(offset int) *QueryBuilder {
	if offset < 0 {
		offset = 0
	}
	q.offset = offset
	return q
}

// Get executes the current builder and returns a CodeIgniter-style Result.
func (q *QueryBuilder) Get(ctx context.Context) (*Result, error) {
	if q.err != nil {
		return nil, q.err
	}
	if q.table == "" {
		return nil, fmt.Errorf("query table is required")
	}
	if err := q.orm.ready(); err != nil {
		return nil, err
	}
	columns := "*"
	if len(q.columns) > 0 && !(len(q.columns) == 1 && q.columns[0] == "*") {
		quoted := make([]string, len(q.columns))
		for i, column := range q.columns {
			quoted[i] = q.orm.quote(column)
		}
		columns = strings.Join(quoted, ", ")
	}
	query := fmt.Sprintf("SELECT %s FROM %s", columns, q.orm.quote(q.table))
	args := make([]any, 0, len(q.wheres)+1)
	if len(q.wheres) > 0 {
		conditions := make([]string, len(q.wheres))
		for i, condition := range q.wheres {
			if condition.value == nil {
				conditions[i] = q.orm.quote(condition.column) + " IS NULL"
				continue
			}
			conditions[i] = fmt.Sprintf("%s = %s", q.orm.quote(condition.column), q.orm.placeholder(len(args)+1))
			args = append(args, condition.value)
		}
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	if q.orderBy != "" {
		query += " ORDER BY " + q.orm.quote(q.orderBy) + " " + q.direction
	}
	if q.limit > 0 {
		query += " LIMIT " + q.orm.placeholder(len(args)+1)
		args = append(args, q.limit)
		if q.offset > 0 {
			query += " OFFSET " + q.orm.placeholder(len(args)+1)
			args = append(args, q.offset)
		}
	}
	return q.orm.Query(ctx, query, args...)
}

// First executes the current builder and returns the first row or sql.ErrNoRows.
func (q *QueryBuilder) First(ctx context.Context) (map[string]any, error) {
	result, err := q.Limit(1).Get(ctx)
	if err != nil {
		return nil, err
	}
	return result.RowArrayOrError()
}

// Find adds an id condition and returns one row.
func (q *QueryBuilder) Find(ctx context.Context, id int64) (map[string]any, error) {
	return q.Where("id", id).First(ctx)
}

// ResultArray returns all result rows, similar to CodeIgniter's result_array().
func (r *Result) ResultArray() []map[string]any {
	return r.rows
}

// GetResult returns all rows using a CodeIgniter 4-style method name.
func (r *Result) GetResult() []map[string]any {
	return r.ResultArray()
}

// GetResultArray returns all rows using a CodeIgniter 4-style method name.
func (r *Result) GetResultArray() []map[string]any {
	return r.ResultArray()
}

// RowArray returns the first result row or nil when the result is empty.
func (r *Result) RowArray() map[string]any {
	if len(r.rows) == 0 {
		return nil
	}
	return r.rows[0]
}

// GetRow returns the first row using a CodeIgniter 4-style method name.
func (r *Result) GetRow() map[string]any {
	return r.RowArray()
}

// GetRowArray returns the first row as a map using a CodeIgniter 4-style name.
func (r *Result) GetRowArray() map[string]any {
	return r.RowArray()
}

// RowArrayOrError returns the first row or sql.ErrNoRows for a find-style call.
func (r *Result) RowArrayOrError() (map[string]any, error) {
	row := r.RowArray()
	if row == nil {
		return nil, sql.ErrNoRows
	}
	return row, nil
}

// NumRows returns the number of rows in the result.
func (r *Result) NumRows() int {
	return len(r.rows)
}

// ready checks whether a live database connection exists.
func (o *ORM) ready() error {
	if o == nil || o.db == nil {
		return ErrNotConfigured
	}
	return nil
}

// conditions builds sorted equality predicates for update/delete operations.
func (o *ORM) conditions(keys []string, values map[string]any, offset int) ([]string, []any) {
	conditions := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for _, key := range keys {
		if values[key] == nil {
			conditions = append(conditions, o.quote(key)+" IS NULL")
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s = %s", o.quote(key), o.placeholder(offset+len(args)+1)))
		args = append(args, values[key])
	}
	return conditions, args
}

// quote applies the active SQL dialect's identifier quoting.
func (o *ORM) quote(identifier string) string {
	if o.dialect == "postgres" {
		return `"` + identifier + `"`
	}
	return "`" + identifier + "`"
}

// placeholder returns the parameter syntax for the active SQL dialect.
func (o *ORM) placeholder(index int) string {
	if o.dialect == "postgres" {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

// validIdentifier rejects user-controlled SQL identifiers.
func validIdentifier(identifier string) error {
	if !identifierPattern.MatchString(identifier) {
		return fmt.Errorf("invalid SQL identifier %q", identifier)
	}
	return nil
}

// writableKeys returns deterministic, validated map keys for write operations.
func writableKeys(values map[string]any) ([]string, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("write values are required")
	}
	keys := sortedMapKeys(values)
	for _, key := range keys {
		if err := validIdentifier(key); err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// sortedMapKeys returns map keys in stable order for predictable SQL parameters.
func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
