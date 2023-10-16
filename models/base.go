package models

import (
	"strings"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/uptrace/bun"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// BaseModel defines the base model for all models
type BaseModel struct {
	ID        string         `bun:",pk"`
	CreatedAt types.DateTime `bun:",nullzero,notnull"`
	UpdatedAt types.DateTime `bun:",nullzero,notnull"`
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

func selectWhere(q *bun.SelectQuery, params *database.DatabaseParams, table string) *bun.SelectQuery {
	if params != nil && params.Where != nil {
		for _, where := range params.Where {
			// Update the column with the default table name when a table is not present
			if !strings.Contains(where.Column, ".") {
				where.Column = table + "." + where.Column
			}

			if where.Query == "" {
				q = q.Where("? = ?", bun.Ident(where.Column), where.Value)
			} else {
				q = q.Where(where.Query, bun.Ident(where.Column), where.Value)
			}
		}
	}

	return q
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func selectOrderBy(q *bun.SelectQuery, params *database.DatabaseParams, table string) *bun.SelectQuery {
	if params != nil && params.OrderBy != nil {
		orderBys := []string{}

		// The orderby can come in 2 forms:
		//  - A single string with comma separated columns
		//  - An array of strings with each string being a column
		for _, orderBy := range params.OrderBy {
			// Split the string by comma into slice
			parts := strings.Split(orderBy, ",")
			for _, part := range parts {
				if !strings.Contains(part, ".") {
					orderBys = append(orderBys, table+"."+part)
				}
			}
		}

		q = q.Order(orderBys...)
	}

	return q
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func selectRelation(q *bun.SelectQuery, params *database.DatabaseParams) *bun.SelectQuery {
	if params != nil && params.Relation != nil {
		for _, relation := range params.Relation {
			// Select specific columns from the relation and/or order by specific columns
			if len(relation.Cols) > 0 || len(relation.OrderBy) > 0 {
				q = q.Relation(relation.Struct, func(q *bun.SelectQuery) *bun.SelectQuery {
					for _, col := range relation.Cols {
						q = q.Column(col)
					}

					if len(relation.OrderBy) > 0 {
						q = q.Order(relation.OrderBy...)
					}

					return q
				})
			} else {
				q = q.Relation(relation.Struct)
			}
		}
	}

	return q
}
