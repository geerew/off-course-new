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
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewScanDao returns a new ScanDao
func NewScanDao(db database.Database) *ScanDao {
	return &ScanDao{
		BaseDao: BaseDao{
			db:    db,
			table: "scans",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates a scan
func (dao *ScanDao) Create(s *models.Scan, tx *database.Tx) error {
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
		Insert(dao.Table()).
		SetMap(toDBMapOrPanic(s)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets a scan with the given course ID
func (dao *ScanDao) Get(courseId string, tx *database.Tx) (*models.Scan, error) {
	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".course_id": courseId},
	}

	return genericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates status of a scan
func (dao *ScanDao) Update(scan *models.Scan, tx *database.Tx) error {
	if scan.ID == "" {
		return ErrEmptyId
	}

	scan.RefreshUpdatedAt()

	// Convert to a map so we have the rendered values
	data := toDBMapOrPanic(scan)

	query, args, _ := squirrel.
		StatementBuilder.
		Update(dao.Table()).
		Set("status", data["status"]).
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

// Delete deletes scans based upon the where clause
func (dao *ScanDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return genericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Next gets the next scan whose status is `waiting“
func (dao *ScanDao) Next(tx *database.Tx) (*models.Scan, error) {
	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".status": types.ScanStatusWaiting},
		OrderBy: []string{"created_at ASC"},
	}

	scan, err := genericGet(dao, dbParams, dao.scanRow, tx)
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

// countSelect returns the default count select builder
//
// It performs 1 left join
//   - courses table to get `course_path`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *ScanDao) countSelect() squirrel.SelectBuilder {
	courseDao := NewCourseDao(dao.db)

	return dao.BaseDao.countSelect().
		LeftJoin(courseDao.Table() + " ON " + dao.Table() + ".course_id = " + courseDao.Table() + ".id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
func (dao *ScanDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *ScanDao) columns() []string {
	courseDao := NewCourseDao(dao.db)

	return append(
		dao.BaseDao.columns(),
		courseDao.Table()+".path AS course_path",
	)
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
