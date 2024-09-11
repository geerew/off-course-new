package daos

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// extractTableAndColumn extracts the table and column name from a string. If a table prefix
// is not found, it will return as an empty string
func extractTableAndColumn(orderBy string) (string, string) {
	parts := strings.Fields(orderBy)
	tableColumn := strings.Split(parts[0], ".")

	if len(tableColumn) == 2 {
		return tableColumn[0], tableColumn[1]
	}

	return "", tableColumn[0]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isValidOrderBy returns true if the orderBy string is valid. The table and column are validated
// against the given list of valid table.columns (ex. courses.id, scans.status as scan_status).
func isValidOrderBy(table, column string, validateTableColumns []string) bool {
	// If the column is empty, always return false
	if column == "" {
		return false
	}

	for _, validTc := range validateTableColumns {
		// Wildcard match (ex. courses.* == id)
		if table == "" && strings.HasSuffix(validTc, ".*") {
			return true
		}

		// Exact match (ex. id == id || courses.id == courses.id || courses.id as courses_id == courses.id)
		if validTc == column || validTc == table+"."+column || strings.HasPrefix(validTc, table+"."+column+" as ") {
			return true
		}

		// Table + wildcard match (ex. courses.* == courses.id)
		if strings.HasSuffix(validTc, ".*") && strings.HasPrefix(validTc, table+".") {
			return true
		}

		// courses.id as course_id == course_id
		if strings.HasSuffix(validTc, " as "+column) {
			return true
		}
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericProcessOrderBy processes an orderBy slice and returns a new slice with only the valid
// order by strings
//
// When an orderBy string does not contain a table prefix, it will be set to the dao's table
func genericProcessOrderBy(orderBy []string, validColumns []string, dao daoer, exactMatch bool) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	var processedOrderBy []string

	for _, ob := range orderBy {
		t, c := extractTableAndColumn(ob)

		// If exact match is set, the table must match the dao's table
		if exactMatch && t == "" {
			continue
		}

		// Prefix the table with the dao's table if not found
		if t == "" {
			t = dao.Table()
			ob = t + "." + ob
		}

		if isValidOrderBy(t, c, validColumns) {
			processedOrderBy = append(processedOrderBy, ob)
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericCount counts the number of rows in a table
func genericCount(dao daoer, dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	queryRowFn := dao.Db().QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	builder := dao.countSelect().
		Columns("COUNT(DISTINCT " + dao.Table() + ".id)")

	if dbParams != nil && dbParams.Where != nil {
		builder = builder.Where(dbParams.Where)
	}

	query, args, _ := builder.ToSql()

	var count int
	err := queryRowFn(query, args...).Scan(&count)

	return count, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericGet gets a row from the given table
func genericGet[T any](
	dao daoer,
	dbParams *database.DatabaseParams,
	scanFn ScanFn[T],
	tx *database.Tx,
) (*T, error) {
	queryRowFn := dao.Db().QueryRow
	if tx != nil {
		queryRowFn = tx.QueryRow
	}

	if dbParams == nil || dbParams.Where == nil {
		return nil, ErrMissingWhere
	}

	builder := dao.baseSelect()

	if dbParams.Columns == nil {
		builder = builder.Columns(dao.Table() + ".*")
	} else {
		builder = builder.Columns(dbParams.Columns...)
	}

	if dbParams.OrderBy != nil {
		builder = builder.OrderBy(dbParams.OrderBy...)
	}

	query, args, _ := builder.
		Where(dbParams.Where).
		ToSql()

	row := queryRowFn(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	res, err := scanFn(row)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericList lists rows from the given table
func genericList[T any](
	dao daoer,
	dbParams *database.DatabaseParams,
	scanFn ScanFn[T],
	tx *database.Tx,
) ([]*T, error) {
	rows, err := genericListWithoutScan(dao, dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*T
	for rows.Next() {
		r, err := scanFn(rows)
		if err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericListWithoutScan lists rows from the given table but leaves the scanning to the
// caller
func genericListWithoutScan(
	dao daoer,
	dbParams *database.DatabaseParams,
	tx *database.Tx,
) (*sql.Rows, error) {
	queryFn := dao.Db().Query
	if tx != nil {
		queryFn = tx.Query
	}

	builder := dao.baseSelect()

	if dbParams != nil {
		if dbParams.Columns == nil {
			builder = builder.Columns(dao.Table() + ".*")
		} else {
			builder = builder.Columns(dbParams.Columns...)
		}

		if dbParams.Where != "" {
			builder = builder.Where(dbParams.Where)
		}

		if dbParams.OrderBy != nil {
			builder = builder.OrderBy(dbParams.OrderBy...)
		}

		if dbParams.Pagination != nil {
			if count, err := dao.Count(dbParams, tx); err != nil {
				return nil, err
			} else {
				dbParams.Pagination.SetCount(count)
				builder = builder.
					Offset(uint64(dbParams.Pagination.Offset())).
					Limit(uint64(dbParams.Pagination.Limit()))
			}
		}

		if dbParams.GroupBys != nil {
			builder = builder.GroupBy(dbParams.GroupBys...)
		}

		if dbParams.Having != nil {
			builder = builder.Having(dbParams.Having)
		}
	}

	query, args, _ := builder.ToSql()

	return queryFn(query, args...)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// genericDelete deletes a row from a table based upon a where clause
func genericDelete(dao daoer, dbParams *database.DatabaseParams, tx *database.Tx) error {
	execFn := dao.Db().Exec
	if tx != nil {
		execFn = tx.Exec
	}

	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	query, args, _ := squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Delete(dao.Table()).
		Where(dbParams.Where).
		ToSql()

	_, err := execFn(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// modelToMapOrPanic takes a model and builds a map where the key is the name of the column
// and the value is the value
//
// Only fields with a `db` tag will be included in the map
func modelToMapOrPanic(model any) map[string]any {
	result := make(map[string]any)

	v, t := utils.ReflectValueAndType(model)

	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("input is not a struct: %v", v.Kind()))
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		dbTag := fieldType.Tag.Get(models.DB_TAG)
		if dbTag == "" {
			continue
		}

		tagParts := strings.Split(dbTag, ":")
		columnName := tagParts[0]

		if slices.Contains(tagParts[1:], "nested") && field.Kind() == reflect.Struct {
			nestedMap := modelToMapOrPanic(field.Interface())

			for k, v := range nestedMap {
				result[k] = v
			}
		} else {
			var processedValue any
			processedValue = valueOrPanic(columnName, field.Interface(), field.Kind())

			// Loop over the extra tags and format the field accordingly
			for _, extraTag := range tagParts[1:] {
				processedValue = processDbTags(processedValue, extraTag)
			}

			result[columnName] = processedValue
		}
	}

	return result
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// valueOrPanic handles getting values from base types and custom types. Custom
// types must implement the driver.Valuer interface from sql
//
// If the value is a struct and it does NOT implement driver.Valuer, it will panic
func valueOrPanic(columnName string, value any, kind reflect.Kind) any {
	var processedValue any

	if kind == reflect.Struct {
		if valuer, ok := value.(driver.Valuer); ok {
			var err error
			processedValue, err = valuer.Value()
			if err != nil {
				panic(fmt.Errorf("error converting value for column %s: %v", columnName, err))
			}
		} else {
			panic(fmt.Errorf("struct for column %s does not implement driver.Valuer", columnName))
		}
	} else {
		processedValue = value
	}

	return processedValue
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// processDbTags applies additional tag logic
//
// Currently supports:
//   - required: if the field is a string and has a length of 0, it will be set to nil
func processDbTags(value any, tag string) any {
	v := reflect.ValueOf(value)

	switch tag {
	case "required":
		if v.Kind() == reflect.String && v.Len() == 0 {
			return nil
		}
	}
	return value
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// tableColumnsOrPanic takes a model and looks up the `db` and `db_join` tags to build a list of
// valid table.column names. It will return 2 slices:
//
// Slice 1 will contain the table.column names with an optional alias. This can be used
// when building a select statement to ensure the correct table.column is selected.
//
// Slice 2 will contain the column names or alias names if applicable. This can be used when
// reducing the orderBy clause to only include valid columns
func tableColumnsOrPanic(model any, table string) ([]string, []string) {
	var selectColumns []string
	var orderByColumns []string

	v, t := utils.ReflectValueAndType(model)

	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("input is not a struct: %v", v.Kind()))
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		dbTag := fieldType.Tag.Get(models.DB_TAG)
		dbJoinTag := fieldType.Tag.Get(models.DB_JOIN_TAG)

		if dbTag == "" && dbJoinTag == "" {
			continue
		}

		if dbTag != "" {
			tagParts := strings.Split(dbTag, ":")

			if len(tagParts) > 1 && slices.Contains(tagParts[1:], "nested") && field.Kind() == reflect.Struct {
				// Recursively get the columns for the nested struct
				nestedSelect, nestedOrderBy := tableColumnsOrPanic(field.Interface(), table)
				selectColumns = append(selectColumns, nestedSelect...)
				orderByColumns = append(orderByColumns, nestedOrderBy...)
			} else {
				selectColumns = append(selectColumns, table+"."+tagParts[0])
				orderByColumns = append(orderByColumns, table+"."+tagParts[0])
			}
		} else {
			tagParts := strings.Split(dbJoinTag, ":")

			if len(tagParts) < 2 {
				panic(fmt.Errorf("db_join tag must contain table:column - %s", dbJoinTag))
			}

			joinTable := tagParts[0]
			column := tagParts[1]
			alias := ""
			expression := ""

			// Check for alias and expression (like COALESCE) in the db_join tag
			if len(tagParts) > 2 {
				alias = tagParts[2]

				if len(tagParts) > 3 {
					expression = tagParts[3]
				}
			}

			// Build the select column, either with the expression or as a normal table.column
			selectColumn := joinTable + "." + column
			if expression != "" {
				selectColumn = expression
			}

			// Add alias if present
			if alias != "" {
				selectColumn += " AS " + alias
				orderByColumns = append(orderByColumns, alias)
			} else {
				orderByColumns = append(orderByColumns, selectColumn)
			}

			selectColumns = append(selectColumns, selectColumn)
		}
	}

	return selectColumns, orderByColumns
}
