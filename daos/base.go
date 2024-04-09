package daos

import (
	"errors"
	"strings"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row that can be scanned into a struct
type Scannable interface {
	Scan(dest ...interface{}) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Errors
var (
	ErrEmptyId       = errors.New("id cannot be empty")
	ErrMissingWhere  = errors.New("where clause cannot be empty")
	ErrInvalidPrefix = errors.New("prefix must be greater than 0")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilStr returns nil when a string is empty
//
// Use this when inserting into the database to avoid inserting empty strings
func NilStr(s string) any {
	if s == "" {
		return nil
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// extractTableColumn extracts the table and column name from an orderBy string. If no table prefix
// is found, the table part is returned as an empty string
func extractTableColumn(orderBy string) (string, string) {
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
