package dao

import (
	"context"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse creates a course and course progress
func (dao *DAO) CreateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	return dao.db.RunInTransaction(ctx, func(txCtx context.Context) error {
		err := dao.Create(txCtx, course)
		if err != nil {
			return err
		}

		courseProgress := &models.CourseProgress{CourseID: course.Id()}
		return dao.CreateCourseProgress(txCtx, courseProgress)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourse updates a course
func (dao *DAO) UpdateCourse(ctx context.Context, course *models.Course) error {
	if course == nil {
		return utils.ErrNilPtr
	}

	_, err := dao.Update(ctx, course)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ClassifyCoursePaths classifies the given paths into one of the following categories:
//   - PathClassificationNone: The path does not exist in the courses table
//   - PathClassificationAncestor: The path is an ancestor of a course path
//   - PathClassificationCourse: The path is an exact match to a course path
//   - PathClassificationDescendant: The path is a descendant of a course path
//
// The paths are returned as a map with the original path as the key and the classification as the
// value
func (dao *DAO) ClassifyCoursePaths(ctx context.Context, paths []string) (map[string]types.PathClassification, error) {
	course := &models.Course{}

	paths = slices.DeleteFunc(paths, func(s string) bool {
		return s == ""
	})

	if len(paths) == 0 {
		return nil, nil
	}

	results := make(map[string]types.PathClassification)
	for _, path := range paths {
		results[path] = types.PathClassificationNone
	}

	whereClause := make([]squirrel.Sqlizer, len(paths))
	for i, path := range paths {
		whereClause[i] = squirrel.Like{course.Table() + ".path": path + "%"}
	}

	query, args, _ := squirrel.
		StatementBuilder.
		Select(course.Table() + ".path").
		From(course.Table()).
		Where(squirrel.Or(whereClause)).
		ToSql()

	q := database.QuerierFromContext(ctx, dao.db)
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coursePath string
	coursePaths := []string{}
	for rows.Next() {
		if err := rows.Scan(&coursePath); err != nil {
			return nil, err
		}
		coursePaths = append(coursePaths, coursePath)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Process
	for _, path := range paths {
		for _, coursePath := range coursePaths {
			if coursePath == path {
				results[path] = types.PathClassificationCourse
				break
			} else if strings.HasPrefix(coursePath, path) {
				results[path] = types.PathClassificationAncestor
				break
			} else if strings.HasPrefix(path, coursePath) && results[path] != types.PathClassificationAncestor {
				results[path] = types.PathClassificationDescendant
				break
			}
		}
	}

	return results, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// // ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// // version of this array
// //
// // It will creates a new list of valid table columns based upon columns() for the current
// // DAO. Additionally, it handles the special case of 'scan_status' column, which requires custom
// // sorting logic, via a CASE statement.
// //
// // The custom sorting logic is defined as follows:
// //   - NULL values are treated as the lowest value (sorted first in ASC, last in DESC)
// //   - 'waiting' status is treated as the second value
// //   - 'processing' status is treated as the third value
// func (dao *CourseDao) ProcessOrderBy(orderBy []string, validOrderByColumns []string) []string {
// 	if len(orderBy) == 0 {
// 		return orderBy
// 	}

// 	var processedOrderBy []string

// 	for _, ob := range orderBy {
// 		t, c := extractTableAndColumn(ob)

// 		// Prefix the table with the dao's table if not found
// 		if t == "" {
// 			t = dao.Table()
// 			ob = t + "." + ob
// 		}

// 		if isValidOrderBy(t, c, validOrderByColumns) {
// 			// When the column is 'scan_status', apply the custom sorting logic
// 			if c == "scan_status" {
// 				// Determine the sort direction, defaulting to ASC if not specified
// 				parts := strings.Fields(ob)
// 				sortDirection := "ASC"
// 				if len(parts) > 1 {
// 					sortDirection = strings.ToUpper(parts[1])
// 				}

// 				caseStmt := "CASE " +
// 					"WHEN scan_status IS NULL THEN 1 " +
// 					"WHEN scan_status = 'waiting' THEN 2 " +
// 					"WHEN scan_status = 'processing' THEN 3 " +
// 					"END " + sortDirection

// 				processedOrderBy = append(processedOrderBy, caseStmt)
// 			} else {
// 				processedOrderBy = append(processedOrderBy, ob)
// 			}
// 		}
// 	}

// 	return processedOrderBy
// }
