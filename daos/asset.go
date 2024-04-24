package daos

import (
	"database/sql"
	"slices"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetDao is the data access object for assets
type AssetDao struct {
	db    database.Database
	Table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetDao returns a new AssetDao
func NewAssetDao(db database.Database) *AssetDao {
	return &AssetDao{
		db:    db,
		Table: "assets",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of assets
func (dao *AssetDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())
	return generic.Count(params, nil)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new asset
func (dao *AssetDao) Create(a *models.Asset) error {
	if a.Prefix.Valid && a.Prefix.Int16 < 0 {
		return ErrInvalidPrefix
	}

	if a.ID == "" {
		a.RefreshId()
	}

	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.Table).
		SetMap(dao.data(a)).
		ToSql()

	_, err := dao.db.Exec(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects an asset with the given ID.
//
// `dbparams` can be used to order the attachments
//
// `tx` allows for the function to be run within a transaction
func (dao *AssetDao) Get(id string, dbParams *database.DatabaseParams, tx *sql.Tx) (*models.Asset, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())

	assetDbParams := &database.DatabaseParams{
		Columns: dao.columns(),
		Where:   squirrel.Eq{dao.Table + ".id": id},
	}

	row, err := generic.Get(assetDbParams, tx)
	if err != nil {
		return nil, err
	}

	asset, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	// Get the attachments
	attachmentDao := NewAttachmentDao(dao.db)
	if dbParams != nil && slices.Contains(dbParams.IncludeRelations, attachmentDao.Table) {
		// Set the DB params
		attachmentDbParams := &database.DatabaseParams{
			OrderBy: dbParams.OrderBy,
			Where:   squirrel.Eq{"asset_id": asset.ID},
		}

		attachments, err := attachmentDao.List(attachmentDbParams, tx)
		if err != nil {
			return nil, err
		}

		asset.Attachments = attachments
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects assets
//
// `tx` allows for the function to be run within a transaction
func (dao *AssetDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Asset, error) {
	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	origOrderBy := dbParams.OrderBy
	dbParams.OrderBy = dao.ProcessOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.columns()
	}

	rows, err := generic.List(dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*models.Asset
	assetIds := []string{}

	for rows.Next() {
		a, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		assets = append(assets, a)
		assetIds = append(assetIds, a.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get the attachments
	attachmentDao := NewAttachmentDao(dao.db)
	if len(assets) > 0 && slices.Contains(dbParams.IncludeRelations, attachmentDao.Table) {

		// Reduce the order by clause to only include columns specific to the attachments table
		reducedOrderBy := attachmentDao.ProcessOrderBy(origOrderBy)

		dbParams = &database.DatabaseParams{
			OrderBy: reducedOrderBy,
			Where:   squirrel.Eq{"asset_id": assetIds},
		}

		// Get the attachments
		attachments, err := attachmentDao.List(dbParams, tx)
		if err != nil {
			return nil, err
		}

		// Store in a map for easy lookup
		attachmentsMap := map[string][]*models.Attachment{}
		for _, a := range attachments {
			attachmentsMap[a.AssetID] = append(attachmentsMap[a.AssetID], a)
		}

		// Add attachments to its asset
		for _, a := range assets {
			a.Attachments = attachmentsMap[a.ID]
		}
	}

	return assets, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete deletes an asset based upon the where clause
//
// `tx` allows for the function to be run within a transaction
func (dao *AssetDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())
	return generic.Delete(dbParams, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ProcessOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon columns() for the current
// DAO
func (dao *AssetDao) ProcessOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	generic := NewGenericDao(dao.db, dao.Table, dao.baseSelect())
	return generic.ProcessOrderBy(orderBy, dao.columns())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// It performs 1 left join
//   - assets progress table to get `video_pos`, `completed` and `completed_at`
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AssetDao) baseSelect() squirrel.SelectBuilder {
	apDao := NewAssetProgressDao(dao.db)

	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.Table).
		LeftJoin(apDao.Table + " ON " + dao.Table + ".id = " + apDao.Table + ".asset_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// columns returns the columns to select
func (dao *AssetDao) columns() []string {
	apDao := NewAssetProgressDao(dao.db)

	return []string{
		dao.Table + ".*",
		apDao.Table + ".video_pos",
		apDao.Table + ".completed",
		apDao.Table + ".completed_at",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// data generates a map of key/values for an asset
func (dao *AssetDao) data(a *models.Asset) map[string]any {
	return map[string]any{
		"id":         a.ID,
		"course_id":  NilStr(a.CourseID),
		"title":      NilStr(a.Title),
		"prefix":     a.Prefix,
		"chapter":    NilStr(a.Chapter),
		"type":       NilStr(a.Type.String()),
		"path":       NilStr(a.Path),
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an asset row
func (dao *AssetDao) scanRow(scannable Scannable) (*models.Asset, error) {
	var a models.Asset

	// Nullable fields
	var chapter sql.NullString
	var videoPos sql.NullInt16
	var completed sql.NullBool

	err := scannable.Scan(
		&a.ID,
		&a.CourseID,
		&a.Title,
		&a.Prefix,
		&chapter,
		&a.Type,
		&a.Path,
		&a.CreatedAt,
		&a.UpdatedAt,
		// Asset progress
		&videoPos,
		&completed,
		&a.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	if chapter.Valid {
		a.Chapter = chapter.String
	}

	a.VideoPos = int(videoPos.Int16)
	a.Completed = completed.Bool

	return &a, nil
}
