package daos

import (
	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AttachmentDao is the data access object for attachments
type AttachmentDao struct {
	BaseDao
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAttachmentDao returns a new AttachmentDao
func NewAttachmentDao(db database.Database) *AttachmentDao {
	return &AttachmentDao{
		BaseDao: BaseDao{
			db:    db,
			table: "attachments",
		},
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count counts the attachments
func (dao *AttachmentDao) Count(dbParams *database.DatabaseParams, tx *database.Tx) (int, error) {
	return GenericCount(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create creates an attachment
func (dao *AttachmentDao) Create(a *models.Attachment, tx *database.Tx) error {
	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table()).
		SetMap(dao.data(a)).
		ToSql()

	execFn := dao.db.Exec
	if tx != nil {
		execFn = tx.Exec
	}

	_, err := execFn(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets attachment with the given ID
func (dao *AttachmentDao) Get(id string, tx *database.Tx) (*models.Attachment, error) {
	dbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table() + ".id": id},
	}

	return GenericGet(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List lists attachments
func (dao *AttachmentDao) List(dbParams *database.DatabaseParams, tx *database.Tx) ([]*models.Attachment, error) {
	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	// Process the order by clauses
	dbParams.OrderBy = GenericProcessOrderBy(dbParams.OrderBy, dao.columns(), false)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	return GenericList(dao, dbParams, dao.scanRow, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes attachments based upon the where clause
func (dao *AttachmentDao) Delete(dbParams *database.DatabaseParams, tx *database.Tx) error {
	return GenericDelete(dao, dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// countSelect returns the default count select builder
func (dao *AttachmentDao) countSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table()).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AttachmentDao) baseSelect() squirrel.SelectBuilder {
	return dao.countSelect()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *AttachmentDao) columns() []string {
	return []string{
		dao.Table() + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for an attachment
func (dao *AttachmentDao) data(a *models.Attachment) map[string]any {
	return map[string]any{
		"id":         a.ID,
		"course_id":  NilStr(a.CourseID),
		"asset_id":   NilStr(a.AssetID),
		"title":      NilStr(a.Title),
		"path":       NilStr(a.Path),
		"created_at": FormatTime(a.CreatedAt),
		"updated_at": FormatTime(a.UpdatedAt),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an attachment row
func (dao *AttachmentDao) scanRow(scannable Scannable) (*models.Attachment, error) {
	var a models.Attachment

	var createdAt string
	var updatedAt string

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.AssetID,
		&a.Title,
		&a.Path,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if a.CreatedAt, err = ParseTime(createdAt); err != nil {
		return nil, err
	}

	if a.UpdatedAt, err = ParseTime(updatedAt); err != nil {
		return nil, err
	}

	return &a, nil
}
