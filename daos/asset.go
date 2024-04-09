package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetDao is the data access object for assets
type AssetDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetDao returns a new AssetDao
func NewAssetDao(db database.Database) *AssetDao {
	return &AssetDao{
		db:    db,
		table: TableAssets(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssets returns the name of the assets table
func TableAssets() string {
	return "assets"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Count returns the number of assets
func (dao *AssetDao) Count(params *database.DatabaseParams) (int, error) {
	generic := NewGenericDao(dao.db, dao.table)
	return generic.Count(dao.baseSelect(), params, nil)
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
		Insert(dao.table).
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
	generic := NewGenericDao(dao.db, dao.table)

	assetDbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".id": id},
	}

	row, err := generic.Get(dao.baseSelect(), assetDbParams, tx)
	if err != nil {
		return nil, err
	}

	asset, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	// Get the attachments
	attachmentDao := NewAttachmentDao(dao.db)

	// Set the DB params
	var attachmentDbParams *database.DatabaseParams

	if dbParams == nil {
		attachmentDbParams = &database.DatabaseParams{
			Where: squirrel.Eq{"asset_id": asset.ID},
		}
	} else {
		attachmentDbParams = &database.DatabaseParams{
			OrderBy: dbParams.OrderBy,
			Where:   squirrel.Eq{"asset_id": asset.ID},
		}
	}

	attachments, err := attachmentDao.List(attachmentDbParams, tx)
	if err != nil {
		return nil, err
	}

	asset.Attachments = attachments

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// List selects assets
//
// `tx` allows for the function to be run within a transaction
func (dao *AssetDao) List(dbParams *database.DatabaseParams, tx *sql.Tx) ([]*models.Asset, error) {
	generic := NewGenericDao(dao.db, dao.table)

	if dbParams == nil {
		dbParams = &database.DatabaseParams{}
	}

	origOrderBy := dbParams.OrderBy

	dbParams.OrderBy = dao.processOrderBy(dbParams.OrderBy)

	// Default the columns if not specified
	if len(dbParams.Columns) == 0 {
		dbParams.Columns = dao.selectColumns()
	}

	rows, err := generic.List(dao.baseSelect(), dbParams, tx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*models.Asset

	for rows.Next() {
		a, err := dao.scanRow(rows)
		if err != nil {
			return nil, err
		}

		assets = append(assets, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get the attachments
	if len(assets) > 0 {
		attachmentDao := NewAttachmentDao(dao.db)

		// Get the asset IDs
		assetIds := []string{}
		for _, a := range assets {
			assetIds = append(assetIds, a.ID)
		}

		// Reduce the order by clause to only include columns specific to the attachments table
		var reducedOrderBy []string
		for _, ob := range origOrderBy {
			table, _ := extractTableColumn(ob)

			if table == attachmentDao.table {
				reducedOrderBy = append(reducedOrderBy, ob)
			}
		}

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

// Delete deletes an asset with the given ID
//
// `tx` allows for the function to be run within a transaction
func (dao *AssetDao) Delete(dbParams *database.DatabaseParams, tx *sql.Tx) error {
	if dbParams == nil || dbParams.Where == nil {
		return ErrMissingWhere
	}

	generic := NewGenericDao(dao.db, dao.table)
	return generic.Delete(dbParams, tx)
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
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		LeftJoin(TableAssetsProgress() + " ON " + TableAssets() + ".id = " + TableAssetsProgress() + ".asset_id").
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *AssetDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
		TableAssetsProgress() + ".video_pos",
		TableAssetsProgress() + ".completed",
		TableAssetsProgress() + ".completed_at",
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

// processOrderBy takes an array of strings representing orderBy clauses and returns a processed
// version of this array
//
// It will creates a new list of valid table columns based upon selectColumns() for the current
// DAO
func (dao *AssetDao) processOrderBy(orderBy []string) []string {
	if len(orderBy) == 0 {
		return orderBy
	}

	validTableColumns := dao.selectColumns()
	var processedOrderBy []string

	for _, ob := range orderBy {
		table, column := extractTableColumn(ob)

		if isValidOrderBy(table, column, validTableColumns) {
			processedOrderBy = append(processedOrderBy, ob)
		}
	}

	return processedOrderBy
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
