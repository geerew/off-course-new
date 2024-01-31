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
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewScanDao returns a new ScanDao
func NewScanDao(db database.Database) *ScanDao {
	return &ScanDao{
		db:    db,
		table: TableScans(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableScans returns the name of the scans table
func TableScans() string {
	return "scans"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new scan
func (dao *ScanDao) Create(s *models.Scan) error {
	if s.ID == "" {
		s.RefreshId()
	}

	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	data := map[string]interface{}{
		"id":         s.ID,
		"course_id":  NilStr(s.CourseID),
		"status":     NilStr(s.Status.String()),
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}

	// Default status to waiting
	s.Status = types.NewScanStatus(types.ScanStatusWaiting)

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(data).
		ToSql()

	_, err := dao.db.Exec(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects a scan with the given course ID
func (dao *ScanDao) Get(courseId string) (*models.Scan, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".course_id": courseId},
	}

	row, err := generic.Get(dao.baseSelect(), dbParams)
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
		Update(dao.table).
		Set("status", NilStr(scan.Status.String())).
		Set("updated_at", scan.UpdatedAt).
		Where("id = ?", scan.ID).
		ToSql()

	_, err := dao.db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes a scan with the given ID
func (dao *ScanDao) Delete(id string) error {
	generic := NewGenericDao(dao.db, dao.table)
	return generic.Delete(id)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Next returns the next scan whose status is `waiting“
func (dao *ScanDao) Next() (*models.Scan, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".status": types.ScanStatusWaiting},
		OrderBy: []string{"created_at ASC"},
	}

	row, err := generic.Get(dao.baseSelect(), dbParams)
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
	return squirrel.StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		LeftJoin(TableCourses() + " ON " + dao.table + ".course_id = " + TableCourses() + ".id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *ScanDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
		TableCourses() + ".path AS course_path",
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
