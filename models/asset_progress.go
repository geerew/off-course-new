package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `assets_progress`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AssetProgress defines the model for a course progress
//
// As this changes, update `scanAssetProgressRow()`
type AssetProgress struct {
	BaseModel

	AssetID     string
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssetsProgress returns the table name for the courses progress table
func TableAssetsProgress() string {
	return "assets_progress"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAssetsProgress counts the number of assets progress
func CountAssetsProgress(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(TableAssetsProgress())

	if params != nil && params.Where != "" {
		builder = builder.Where(params.Where)
	}

	// Build the query
	query, args, err := builder.ToSql()
	if err != nil {
		return -1, err
	}

	// Execute the query
	var count int
	err = db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetsProgress selects courses progress
func GetAssetsProgress(db database.Database, params *database.DatabaseParams) ([]*AssetProgress, error) {
	var aps []*AssetProgress

	// Start building the query
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableAssetsProgress() + ".*").
		From(TableAssetsProgress())

	// Add additional clauses
	if params != nil {
		// ORDER BY
		if params != nil && len(params.OrderBy) > 0 {
			builder = builder.OrderBy(params.OrderBy...)
		}

		// WHERE
		if params.Where != "" {
			builder = builder.Where(params.Where)
		}

		// PAGINATION
		if params.Pagination != nil {
			var err error
			if builder, err = paginate(db, params, builder, CountAssetsProgress); err != nil {
				return nil, err
			}
		}
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		ap, err := scanAssetProgressRow(rows)
		if err != nil {
			return nil, err
		}

		aps = append(aps, ap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return aps, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetProgress selects an asset progress for the given asset ID
func GetAssetProgress(db database.Database, assetId string) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(TableAssetsProgress() + ".*").
		From(TableAssetsProgress()).
		Where(sq.Eq{TableAssetsProgress() + ".asset_id": assetId})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	ap, err := scanAssetProgressRow(row)
	if err != nil {
		return nil, err
	}

	return ap, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetProgress inserts a new asset progress
func CreateAssetProgress(db database.Database, ap *AssetProgress) error {
	ap.RefreshId()
	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableAssetsProgress()).
		Columns("id", "asset_id", "video_pos", "completed", "completed_at", "created_at", "updated_at").
		Values(ap.ID, NilStr(ap.AssetID), ap.VideoPos, ap.Completed, ap.CompletedAt, ap.CreatedAt, ap.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetProgressVideoPos updates `video_pos`. When a video progress entry does not exist, it
// will be created. When an entry does exist, it will be updated.
func UpdateAssetProgressVideoPos(db database.Database, assetId string, position int) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	ap, err := GetAssetProgress(db, assetId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Keep the position >= 0
	if position < 0 {
		position = 0
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		ap = &AssetProgress{
			AssetID:  assetId,
			VideoPos: position,
		}

		err = CreateAssetProgress(db, ap)
		if err != nil {
			return nil, err
		}

		return ap, nil
	}

	// --------------------------------
	// Update existing
	// --------------------------------

	// Nothing to do
	if ap.VideoPos == position {
		return ap, nil
	}

	updatedAt := types.NowDateTime()

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableAssetsProgress()).
		Set("video_pos", position).
		Set("updated_at", updatedAt).
		Where("id = ?", ap.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	ap.VideoPos = position
	ap.UpdatedAt = updatedAt

	return ap, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetProgressCompleted updates `completed` and `completed_at`. When a video progress entry
// does not exist, it will be created. When an entry does exist, it will be updated.
//
// When an `completed` is true, `completed_at` is set to the current time. When `completed` is
// false, `completed_at` is set to null
func UpdateAssetProgressCompleted(db database.Database, assetId string, completed bool) (*AssetProgress, error) {
	if assetId == "" {
		return nil, errors.New("id cannot be empty")
	}

	ap, err := GetAssetProgress(db, assetId)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// --------------------------------
	// Create if it does not exist
	// --------------------------------
	if err == sql.ErrNoRows {
		ap = &AssetProgress{
			AssetID:   assetId,
			Completed: completed,
		}

		if completed {
			ap.CompletedAt = types.NowDateTime()
		}

		err = CreateAssetProgress(db, ap)
		if err != nil {
			return nil, err
		}

		return ap, nil
	}

	// --------------------------------
	// Update existing
	// --------------------------------

	// Nothing to do
	if ap.Completed == completed {
		return ap, nil
	}

	updatedAt := types.NowDateTime()

	var completedAt types.DateTime
	if completed {
		completedAt = updatedAt
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Update(TableAssetsProgress()).
		Set("completed", completed).
		Set("completed_at", completedAt).
		Set("updated_at", updatedAt).
		Where("id = ?", ap.ID)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	ap.Completed = completed
	ap.CompletedAt = completedAt
	ap.UpdatedAt = updatedAt

	return ap, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestAssetsProgress creates an asset progress for each asset in the slice. If a db is
// provided, a DB insert will be performed
//
// THIS IS FOR TESTING PURPOSES
func NewTestAssetsProgress(t *testing.T, db database.Database, assets []*Asset) []*AssetProgress {
	aps := []*AssetProgress{}

	for i := 0; i < len(assets); i++ {
		ap := &AssetProgress{}

		ap.RefreshId()
		ap.RefreshCreatedAt()
		ap.RefreshUpdatedAt()

		ap.AssetID = assets[i].ID

		if db != nil {
			err := CreateAssetProgress(db, ap)
			require.Nil(t, err)

			// This allows the created/updated times to be different when inserting multiple rows
			time.Sleep(time.Millisecond * 1)
		}

		aps = append(aps, ap)
	}

	return aps
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanAssetProgressRow scans an asset progress row
func scanAssetProgressRow(scannable Scannable) (*AssetProgress, error) {
	var ap AssetProgress

	err := scannable.Scan(
		&ap.ID,
		&ap.AssetID,
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
