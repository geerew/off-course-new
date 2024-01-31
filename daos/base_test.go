package daos

import "testing"

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
