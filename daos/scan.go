package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanDao is the data access object for scans
type ScanDao struct {
	db    database.Database
	Table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewScanDao returns a new ScanDao
func NewScanDao(db database.Database) *ScanDao {
	return &ScanDao{
		db:    db,
		Table: "scans",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new scan
func (dao *ScanDao) Create(s *models.Scan) error {
	if s.ID == "" {
		s.RefreshId()
	}

	if s.Status.String() == "" {
		s.Status = types.NewScanStatus(types.ScanStatusWaiting)
	}

	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	// Default status to waiting
	s.Status = types.NewScanStatus(types.ScanStatusWaiting)

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table).
		SetMap(dao.data(s)).
		ToSql()

	_, err := dao.db.Exec(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a scan with the given course ID
func (dao *ScanDao) Get(courseId string) (*models.Scan, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())

	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table + ".course_id": courseId},
	}

	row, err := generic.Get(dbParams, nil)
	if err != nil {
		return nil, err
	}

	scan, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return scan, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a scan
//
// Note: Only the `status` can be updated
func (dao *ScanDao) Update(scan *models.Scan) error {
	if scan.ID == "" {
		return ErrEmptyId
	}

	scan.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table).
		Set("status", NilStr(scan.Status.String())).
		Set("updated_at", scan.UpdatedAt).
		Where("id = ?", scan.ID).
		ToSql()

	_, err := dao.db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a scan based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *ScanDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Next returns the next scan whose status is `waiting“
func (dao *ScanDao) Next() (*models.Scan, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())

	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table + ".status": types.ScanStatusWaiting},
		OrderBy: []string{"created_at ASC"},
	}

	row, err := generic.Get(dbParams, nil)
	if err != nil {
		return nil, err
	}

	scan, err := dao.scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return scan, nil

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 1 left join
//   - courses table to get `course_path`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *ScanDao) baseSelect() squirrel.SelectBuilder {
	courseDao := NewCourseDao(dao.db)

	return squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table).
		LeftJoin(courseDao.Table + " ON " + dao.Table + ".course_id = " + courseDao.Table + ".id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *ScanDao) columns() []string {
	courseDao := NewCourseDao(dao.db)

	return []string{
		dao.Table + ".*",
		courseDao.Table + ".path AS course_path",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for a scan
func (dao *ScanDao) data(s *models.Scan) map[string]any {
	return map[string]any{
		"id":         s.ID,
		"course_id":  NilStr(s.CourseID),
		"status":     NilStr(s.Status.String()),
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans a scan row
func (dao *ScanDao) scanRow(scannable Scannable) (*models.Scan, error) {
	var s models.Scan

	err := scannable.Scan(
		&s.ID,
		&s.CourseID,
		&s.Status,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.CoursePath,
	)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
