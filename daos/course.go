package daos

import (
	"database/sql"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CourseDao is the data access object for courses
type CourseDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewCourseDao returns a new CourseDao
func NewCourseDao(db database.Database) *CourseDao {
	return &CourseDao{
		BaseDao: BaseDao{
			db:    db,
			table: "courses",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the courses
func (dao *CourseDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return genericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates a course and courses progress
//
// # A new transaction is created if `tx` is nil
//
// NOTE: There is currently no support for users, but when there is, the default courses_progress
// should be inserted for the admin user
func (dao *CourseDao) Create(c *models.Course, tx *database.Tx) error {
	createFn := func(tx *database.Tx) error {
		if c.ID == "" {
			c.RefreshId()
		}

		c.RefreshCreatedAt()
		c.RefreshUpdatedAt()

		query, args, _ := squirrel.
			StatementBuilder.
			Insert(dao.Table()).
			SetMap(modelToMapOrPanic(c)).
			ToSql()

		// Create the course
		if _, err := tx.Exec(query, args...); err != nil {
			return err
		}

		// Create the course progress
		cp := &models.CourseProgress{
			CourseID: c.ID,
		}

		cpDao := NewCourseProgressDao(dao.db)
		return cpDao.Create(cp, tx)
	}

	if tx == nil {
		return dao.db.RunInTransaction(func(tx *database.Tx) error {
			return createFn(tx)
		})
	} else {
		return createFn(tx)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a course with the given ID
func (dao *CourseDao) Get(id string, tx *database.Tx) (*models.Course, error) {
	selectColumns, _ := tableColumnsOrPanic(models.Course{}, dao.Table())

	courseDbParams := &database.DatabaseParams{
		Columns: selectColumns,
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	return genericGet(dao, courseDbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists courses
func (dao *CourseDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Course, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	selectColumns, orderByColumns := tableColumnsOrPanic(models.Course{}, dao.Table())

	dbParams.Columns = selectColumns

	// Remove invalid orderBy columns
	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy, orderByColumns)

	return genericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates selected columns of a course
//
//   - card_path
//   - available
func (dao *CourseDao) Update(course *models.Course, tx *database.Tx) error {
	if course.ID == "" {
		return ErrEmptyId
	}

	course.RefreshUpdatedAt()

	// Convert to a map so we have the rendered values
	data := modelToMapOrPanic(course)

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("card_path", data["card_path"]).
		Set("available", data["available"]).
		Set("updated_at", data["updated_at"]).
		Where("id = ?", data["id"]).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes courses based upon the where clause
func (dao *CourseDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ClassifyPaths classifies the given paths into one of the following categories:
//   - PathClassificationNone: The path does not exist in the courses table
//   - PathClassificationAncestor: The path is an ancestor of a course path
//   - PathClassificationCourse: The path is an exact match to a course path
//   - PathClassificationDescendant: The path is a descendant of a course path
//
// The paths are returned as a map with the original path as the key and the classification as the
// value
func (dao *CourseDao) ClassifyPaths(paths []string) (map[string]types.PathClassification, error) {
	paths = slices.DeleteFunc(paths, func(s string) bool {
		return s == ""
	})

	if len(paths) == 0 {
		return nil, nil
	}

	// Initialize the results map
	results := make(map[string]types.PathClassification)
	for _, path := range paths {
		results[path] = types.PathClassificationNone
	}

	// Build the where clause
	whereClause := make([]squirrel.Sqlizer, len(paths))
	for i, path := range paths {
		whereClause[i] = squirrel.Like{dao.Table() + ".path": path + "%"}
	}

	query, args, _ := squirrel.
		StatementBuilder.
		Select(dao.Table() + ".path").
		From(dao.table).
		Where(squirrel.Or(whereClause)).
		ToSql()

	rows, err := dao.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Store the found course paths
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

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO. Additionally, it handles the special case of 'scan_status' column, which requires custom
// sorting logic, via a CASE statement.
//
// The custom sorting logic is defined as follows:
//   - NULL values are treated as the lowest value (sorted first in ASC, last in DESC)
//   - 'waiting' status is treated as the second value
//   - 'processing' status is treated as the third value
func (dao *CourseDao) ProcessOrderBy(orderBy []string, validOrderByColumns []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	var processedOrderBy []string

	for _, ob := range orderBy {
		t, c := extractTableAndColumn(ob)

		// Prefix the table with the dao's table if not found
		if t == "" {
			t = dao.Table()
			ob = t + "." + ob
		}

		if isValidOrderBy(t, c, validOrderByColumns) {
			// When the column is 'scan_status', apply the custom sorting logic
			if c == "scan_status" {
				// Determine the sort direction, defaulting to ASC if not specified
				parts := strings.Fields(ob)
				sortDirection := "ASC"
				if len(parts) > 1 {
					sortDirection = strings.ToUpper(parts[1])
				}

				caseStmt := "CASE " +
					"WHEN scan_status IS NULL THEN 1 " +
					"WHEN scan_status = 'waiting' THEN 2 " +
					"WHEN scan_status = 'processing' THEN 3 " +
					"END " + sortDirection

				processedOrderBy = append(processedOrderBy, caseStmt)
			} else {
				processedOrderBy = append(processedOrderBy, ob)
			}
		}
	}

	return processedOrderBy
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default select builder for counting
//
// It performs 2 left joins
//   - scans table to get `scan_status`
//   - courses progress table to get `started`, `started_at`, `percent`, and `completed_at`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *CourseDao) countSelect() squirrel.SelectBuilder {
	sDao := NewScanDao(dao.db)
	cpDao := NewCourseProgressDao(dao.db)

	return dao.BaseDao.countSelect().
		LeftJoin(sDao.Table() + " ON " + dao.Table() + ".id = " + sDao.Table() + ".course_id").
		LeftJoin(cpDao.Table() + " ON " + dao.Table() + ".id = " + cpDao.Table() + ".course_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
func (dao *CourseDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a course row
func (dao *CourseDao) scanRow(scannable Scannable) (*models.Course, error) {
	var c models.Course

	// Nullable fields
	var cardPath sql.NullString
	var scanStatus sql.NullString

	err := scannable.Scan(
		&c.ID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.Title,
		&c.Path,
		&cardPath,
		&c.Available,

		// Scan
		&scanStatus,

		// Course progress
		&c.Started,
		&c.StartedAt,
		&c.Percent,
		&c.CompletedAt,
		&c.ProgressUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if cardPath.Valid {
		c.CardPath = cardPath.String
	}

	if scanStatus.Valid {
		c.ScanStatus = scanStatus.String
	}

	return &c, nil
}
