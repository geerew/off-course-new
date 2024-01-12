package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `scans`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan defines the model for a scan
//
// As this changes, update `scanScanRow()`
type Scan struct {
	BaseModel

	CourseID string
	Status   types.ScanStatus

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------
	CoursePath string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableScans returns the table name for the scans table
func TableScans() string {
	return "scans"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetScan selects a scan for the given course ID
func GetScan(db database.Database, courseId string) (*Scan, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	cols := []string{
		TableScans() + ".*",
		TableCourses() + ".path AS course_path",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableScans()).
		Where(sq.Eq{TableScans() + ".course_id": courseId}).
		LeftJoin(TableCourses() + " ON " + TableScans() + ".course_id = " + TableCourses() + ".id")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	s, err := scanScanRow(row)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateScan inserts a new scan
func CreateScan(db database.Database, s *Scan) error {
	s.RefreshId()
	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	// Default status to waiting
	status := s.Status
	if status == (types.ScanStatus{}) {
		status = types.NewScanStatus(types.ScanStatusWaiting)
	}

	builder := sq.StatementBuilder.
		Insert(TableScans()).
		Columns("id", "course_id", "status", "created_at", "updated_at").
		Values(s.ID, NilStr(s.CourseID), NilStr(status.String()), s.CreatedAt, s.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)

	s.Status = status

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateScanStatus updates `status` for the scan with the given course ID
func UpdateScanStatus(db database.Database, courseId string, newStatus types.ScanStatusType) (*Scan, error) {
	if courseId == "" {
		return nil, errors.New("id cannot be empty")
	}

	s, err := GetScan(db, courseId)
	if err != nil {
		return nil, err
	}

	newScanStatus := types.NewScanStatus(newStatus)

	// Nothing to do
	if s.Status == newScanStatus {
		return s, nil
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableScans()).
		Set("status", newScanStatus).
		Set("updated_at", updatedAt).
		Where("id = ?", s.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	s.Status = newScanStatus
	s.UpdatedAt = updatedAt

	return s, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteScan deletes a scan with the given course ID
func DeleteScan(db database.Database, courseId string) error {
	if courseId == "" {
		return errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Delete(TableScans()).
		Where(sq.Eq{"course_id": courseId})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NextScan returns the next scan to be processed whose status is `waiting“
func NextScan(db database.Database) (*Scan, error) {

	cols := []string{
		TableScans() + ".*",
		TableCourses() + ".path AS course_path",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableScans()).
		Where(sq.Eq{TableScans() + ".status": types.ScanStatusWaiting}).
		LeftJoin(TableCourses() + " ON " + TableScans() + ".course_id = " + TableCourses() + ".id").
		OrderBy(TableScans() + ".created_at ASC")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	s, err := scanScanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return s, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanScanRow scans a scan row
func scanScanRow(scannable Scannable) (*Scan, error) {
	var s Scan

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
