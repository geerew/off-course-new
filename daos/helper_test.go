package daos

import (
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_IsValidOrderBy(t *testing.T) {
	t.Run("with wildcard", func(t *testing.T) {
		validTableColumns := []string{
			"courses.*",
			"scans.status as scan_status",
			"courses_progress.started",
			"courses_progress.percent",
		}

		tests := []struct {
			name     string
			table    string
			column   string
			expected bool
		}{
			// Wildcard match
			{"valid .*", "", "nonexistent", true},
			{"valid .* and direct", "", "percent", true},
			{"valid .* and alias", "", "scan_status", true},

			// Table.* match
			{"valid table.*", "courses", "title", true},

			// Table.column match
			{"valid table.column", "courses_progress", "started", true},
			{"valid table.column as alias", "scans", "status", true},

			// Invalid
			{"invalid table.column", "test", "invalid", false},
			{"invalid column", "", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := isValidOrderBy(tt.table, tt.column, validTableColumns)
				if result != tt.expected {
					t.Errorf("isValidOrderBy(%s, %s) = %v; expected %v", tt.table, tt.column, result, tt.expected)
				}
			})
		}
	})

	t.Run("without wildcard", func(t *testing.T) {
		validTableColumns := []string{
			"courses.id",
			"data",
			"scans.status as scan_status",
			"courses_progress.started",
			"courses_progress.percent",
		}

		tests := []struct {
			name     string
			table    string
			column   string
			expected bool
		}{
			// Exact
			{"valid direct", "", "data", true},
			{"valid direct as alias", "", "scan_status", true},

			// Table.column
			{"valid table.column", "courses_progress", "started", true},
			{"valid table.column as alias", "scans", "status", true},

			// Wildcard
			{"invalid .*", "", "nonexistent", false},
			{"invalid .* and direct", "", "percent", false},
			{"invalid .* and alias", "", "status", false},

			// Table.*
			{"invalid table.*", "courses", "title", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := isValidOrderBy(tt.table, tt.column, validTableColumns)
				if result != tt.expected {
					t.Errorf("isValidOrderBy(%s, %s) = %v; expected %v", tt.table, tt.column, result, tt.expected)
				}
			})
		}
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ToDBMapOrPanic(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected to panic.")
			}
		}()

		toDBMapOrPanic("test")
	})

	t.Run("success", func(t *testing.T) {
		course := &models.Course{
			BaseModel: models.BaseModel{
				ID:        "1",
				CreatedAt: types.NowDateTime(),
				UpdatedAt: types.NowDateTime(),
			},
			Title:     "Test Course",
			Path:      "test-course",
			CardPath:  "test-card-path",
			Available: true,
		}

		result := toDBMapOrPanic(course)

		require.Equal(t, result["id"], course.ID)
		require.Equal(t, result["created_at"], course.CreatedAt.String())
		require.Equal(t, result["updated_at"], course.UpdatedAt.String())
		require.Equal(t, result["title"], course.Title)
		require.Equal(t, result["path"], course.Path)
		require.Equal(t, result["card_path"], course.CardPath)
		require.Equal(t, result["available"], course.Available)
	})

	t.Run("nil", func(t *testing.T) {
		course := &models.Course{
			BaseModel: models.BaseModel{
				ID:        "",
				CreatedAt: types.DateTime{},
				UpdatedAt: types.DateTime{},
			},
			Title:     "",
			Path:      "test-course",
			CardPath:  "test-card-path",
			Available: true,
		}

		result := toDBMapOrPanic(course)

		require.Nil(t, result["id"])
		require.Nil(t, result["created_at"])
		require.Nil(t, result["updated_at"])
		require.Nil(t, result["title"], course.Title)
		require.Equal(t, result["path"], course.Path)
		require.Equal(t, result["card_path"], course.CardPath)
		require.Equal(t, result["available"], course.Available)
	})
}
