package daos

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgressDao is the data access object for assets progress
type AssetProgressDao struct {
	db    database.Database
	table string
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAssetProgressDao returns a new AssetProgressDao
func NewAssetProgressDao(db database.Database) *AssetProgressDao {
	return &AssetProgressDao{
		db:    db,
		table: TableAssetsProgress(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssetProgress returns the name of the assets progress table
func TableAssetsProgress() string {
	return "assets_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create inserts a new asset progress
//
// If this is part of a transaction, use `CreateTx`
func (dao *AssetProgressDao) Create(cp *models.AssetProgress) error {
	return dao.create(cp, dao.db.Exec)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateTx inserts a new asset progress in a transaction
func (dao *AssetProgressDao) CreateTx(cp *models.AssetProgress, tx *sql.Tx) error {
	return dao.create(cp, tx.Exec)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get selects an asset progress with the given asset ID
func (dao *AssetProgressDao) Get(assetId string) (*models.AssetProgress, error) {
	return dao.get(assetId, dao.db.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (dao *AssetProgressDao) GetTx(assetId string, tx *sql.Tx) (*models.AssetProgress, error) {
	return dao.get(assetId, tx.QueryRow)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update updates a asset progress
//
// # It is always run in a transaction
//
// Note: Only the `video_pos` and `completed` can be updated
func (dao *AssetProgressDao) Update(ap *models.AssetProgress) error {
	return dao.db.RunInTransaction(func(tx *sql.Tx) error {
		return dao.UpdateTx(ap, tx)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateTx updates the `video_pos` when the type is a video and `completed` of an asset progress
//
// When the `completed` is true, `completed_at` is set to the current time. When the `completed`
// is false, `completed_at` is set to null
//
// This is to be used in a transaction
// Note: Only the `video_pos` and `completed` can be updated
func (dao *AssetProgressDao) UpdateTx(ap *models.AssetProgress, tx *sql.Tx) error {
	if ap.CourseID == "" || ap.AssetID == "" {
		return ErrEmptyId
	}

	if ap.VideoPos < 0 {
		ap.VideoPos = 0
	}

	currentAp, err := dao.GetTx(ap.AssetID, tx)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if currentAp != nil {
		// Do nothing when there is no change
		if ap.VideoPos == currentAp.VideoPos && ap.Completed == currentAp.Completed {
			return nil
		}

		// Set or clear the completed at time
		if ap.Completed && ap.CompletedAt.IsZero() {
			ap.CompletedAt = types.NowDateTime()
		} else if !ap.Completed {
			ap.CompletedAt = types.DateTime{}
		}
	} else {
		// There is no current asset progress, so set an ID (if not already set) and a created at
		// time
		if ap.ID == "" {
			ap.RefreshId()
		}

		ap.RefreshCreatedAt()
	}

	ap.RefreshUpdatedAt()

	// Update (or create if it doesn't exist)
	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(ap)).
		Suffix("ON CONFLICT (asset_id) DO UPDATE SET video_pos = ?, completed = ?, completed_at = ?, updated_at = ?",
			ap.VideoPos, ap.Completed, ap.CompletedAt, ap.UpdatedAt).
		ToSql()

	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}

	// Refresh course progress
	cpDao := NewCourseProgressDao(dao.db)
	return cpDao.RefreshTx(ap.CourseID, tx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Internal
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// create inserts a new asset progress
func (dao *AssetProgressDao) create(ap *models.AssetProgress, execFunc database.ExecFn) error {
	if ap.ID == "" {
		ap.RefreshId()
	}

	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	query, args, _ := squirrel.
		StatementBuilder.
		Insert(dao.table).
		SetMap(dao.data(ap)).
		ToSql()

	_, err := execFunc(query, args...)

	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// get selects an asset progress with the given asset ID
func (dao *AssetProgressDao) get(assetId string, queryRowFn database.QueryRowFn) (*models.AssetProgress, error) {
	generic := NewGenericDao(dao.db, dao.table)

	dbParams := &database.DatabaseParams{
		Columns: dao.selectColumns(),
		Where:   squirrel.Eq{generic.table + ".asset_id": assetId},
	}

	row, err := generic.get(dao.baseSelect(), dbParams, queryRowFn)
	if err != nil {
		return nil, err
	}

	cp, err := dao.scanRow(row)
	if err != nil {
		return nil, err
	}

	return cp, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// baseSelect returns the default select builder
//
// Note: The columns are removed, so you must specify the columns with `.Columns(...)` when using
// this select builder
func (dao *AssetProgressDao) baseSelect() squirrel.SelectBuilder {
	return squirrel.
		StatementBuilder.
		PlaceholderFormat(squirrel.Question).
		Select("").
		From(dao.table).
		RemoveColumns()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *AssetProgressDao) selectColumns() []string {
	return []string{
		dao.table + ".*",
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// selectColumns returns the columns to select
func (dao *AssetProgressDao) data(ap *models.AssetProgress) map[string]any {
	return map[string]any{
		"id":           ap.ID,
		"asset_id":     NilStr(ap.AssetID),
		"course_id":    NilStr(ap.CourseID),
		"video_pos":    ap.VideoPos,
		"completed":    ap.Completed,
		"completed_at": ap.CompletedAt,
		"created_at":   ap.CreatedAt,
		"updated_at":   ap.UpdatedAt,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanRow scans an assets progress row
func (dao *AssetProgressDao) scanRow(scannable Scannable) (*models.AssetProgress, error) {
	var ap models.AssetProgress

	err := scannable.Scan(
		&ap.ID,
		&ap.AssetID,
		&ap.CourseID,
		&ap.VideoPos,
		&ap.Completed,
		&ap.CompletedAt,
		&ap.CreatedAt,
		&ap.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &ap, nil
}
