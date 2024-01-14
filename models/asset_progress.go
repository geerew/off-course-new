package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `assets_progress`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
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
	builder := assetsProgressBaseSelect().Columns("COUNT(DISTINCT " + TableAssetsProgress() + ".id)")

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
	builder := assetsProgressBaseSelect().Columns(TableAssetsProgress() + ".*")

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

	builder := assetsProgressBaseSelect().
		Columns(TableAssetsProgress() + ".*").
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

	//
	// TODO
	//
	// Update the course percent by first calculating the percentage of completed assets and then
	// updating the course
	// var percent float64

	// // Calculate the percentage of completed assets. IF this fails just log the error and return
	// if err = db.DB().NewSelect().
	// 	Table("assets").
	// 	ColumnExpr("CAST(COUNT(CASE WHEN completed THEN 1 END) * 100 AS FLOAT) / COUNT(*) as completion_percentage").
	// 	Where("course_id = ?", asset.CourseID).
	// 	Scan(ctx, &percent); err != nil {
	// 	log.Err(err).Msg("failed to calculate the percentage of completed assets")
	// 	return asset, nil
	// }

	// // Update the course percent. If this fails just log the error and return
	// if _, err = UpdateCoursePercent(ctx, db, asset.CourseID, int(percent)); err != nil {
	// 	log.Err(err).Msg("failed to update the course `percent`")
	// 	return asset, nil
	// }

	return ap, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetsProgressBaseSelect returns a select builder for the assets_progress table. It does not
// include any columns by default and as such, you must specify the columns with `.Columns(...)`
func assetsProgressBaseSelect() sq.SelectBuilder {
	return sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("").
		From(TableAssetsProgress()).
		RemoveColumns()
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
