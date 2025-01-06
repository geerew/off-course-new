package database

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	loggerType = slog.Any("type", types.LogTypeDB)
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type contextKey string

const querierKey = contextKey("querier")

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// WithQuerier adds a querier to the context
func WithQuerier(ctx context.Context, querier Querier) context.Context {
	return context.WithValue(ctx, querierKey, querier)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// QuerierFromContext returns the querier from the context, defaulting to a defaulted querier if
// not found
func QuerierFromContext(ctx context.Context, defaultQuerier Querier) Querier {
	if querier, ok := ctx.Value(querierKey).(Querier); ok && querier != nil {
		return querier
	}

	return defaultQuerier
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Defines the sql functions
type (
	ExecFn     = func(query string, args ...interface{}) (sql.Result, error)
	QueryFn    = func(query string, args ...interface{}) (*sql.Rows, error)
	QueryRowFn = func(query string, args ...interface{}) *sql.Row
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Database defines the interface for a database
type Database interface {
	Querier
	RunInTransaction(context.Context, func(context.Context) error) error
	SetLogger(*slog.Logger)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Options defines optional params for a database query
type Options struct {
	// A slice of columns to order by (ex ["id DESC", "title ASC"])
	OrderBy []string

	// A slice of columns to select (ex ["id", "title", "courses.col"])
	Columns []string

	// Any valid squirrel WHERE expression
	//
	//
	// Examples:
	//
	//   EQ:   squirrel.Eq{"id": "123"}
	//   IN:   squirrel.Eq{"id": []string{"123", "456"}}
	//   OR:   squirrel.Or{squirrel.Expr("id = ?", "123"), squirrel.Expr("id = ?", "456")}
	//   AND:  squirrel.And{squirrel.Eq{"id": "123"}, squirrel.Eq{"title": "devops"}}
	//   LIKE: squirrel.Like{"title": "%dev%"}
	//   NOT:  squirrel.NotEq{"id": "123"}
	Where squirrel.Sqlizer

	// Columns to group by
	GroupBy []string

	// Limit the results
	Having squirrel.Sqlizer

	// Used to paginate the results
	Pagination *pagination.Pagination
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseManager manages the database connections
type DatabaseManager struct {
	DataDb Database
	LogsDb Database
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DatabaseConfig defines the configuration for a database
type DatabaseConfig struct {
	DataDir    string
	DSN        string
	MigrateDir string
	AppFs      *appFs.AppFs
	InMemory   bool
	Logger     *slog.Logger
}
