package models

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type countFn = func(database.Database, *database.DatabaseParams) (int, error)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row
type Scannable interface {
	Scan(dest ...interface{}) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseModel defines the base model for all models
type BaseModel struct {
	ID        string
	CreatedAt types.DateTime
	UpdatedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshId generates and sets a new model ID
func (b *BaseModel) RefreshId() {
	b.ID = security.PseudorandomString(10)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetId sets the model ID
func (b *BaseModel) SetId(id string) {
	b.ID = id
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshCreated updates the Created At field to the current date/time
func (b *BaseModel) RefreshCreatedAt() {
	b.CreatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// RefreshUpdatedAt updates the Updated At field to the current date/time
func (b *BaseModel) RefreshUpdatedAt() {
	b.UpdatedAt = types.NowDateTime()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// paginate applies pagination to the query
func paginate(db database.Database, params *database.DatabaseParams, builder sq.SelectBuilder, count countFn) (sq.SelectBuilder, error) {
	if count, err := count(db, params); err != nil {
		return builder, err
	} else {
		params.Pagination.SetCount(count)
		builder = params.Pagination.Apply(builder)
	}

	return builder, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilStr returns nil when a string is empty
func NilStr(s string) any {
	if s == "" {
		return nil
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func selectRelation(q *bun.SelectQuery, relations []database.Relation) *bun.SelectQuery {
// 	for _, relation := range relations {
// 		// Make a copy so that it can be used in the SelectQuery closure
// 		relation := relation

// 		// Select specific columns from the relation and/or order by specific columns
// 		if len(relation.Columns) > 0 || len(relation.OrderBy) > 0 {
// 			q = q.Relation(relation.Table, func(q *bun.SelectQuery) *bun.SelectQuery {
// 				for _, col := range relation.Columns {
// 					q = q.Column(col)
// 				}

// 				if len(relation.OrderBy) > 0 {
// 					q = q.Order(relation.OrderBy...)
// 				}

// 				return q
// 			})
// 		} else {
// 			q = q.Relation(relation.Table)
// 		}
// 	}

// 	return q
// }
