package models

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Defines a model for the table `assets`
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset defines the model for a course
//
// As this changes, update `scanAssetRow()`
type Asset struct {
	BaseModel

	CourseID string
	Title    string
	Prefix   sql.NullInt16
	Chapter  string
	Type     types.Asset
	Path     string

	// --------------------------------
	// Not in this table, but added via a join
	// --------------------------------

	// Asset Progress
	VideoPos    int
	Completed   bool
	CompletedAt types.DateTime

	// Attachments
	Attachments []*Attachment
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TableAssets returns the table name for the assets table
func TableAssets() string {
	return "assets"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAssets counts the number of assets
func CountAssets(db database.Database, params *database.DatabaseParams) (int, error) {
	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select("COUNT(*)").
		From(TableAssets())

	// Add where clauses if necessary
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

// GetAssets selects assets
//
// It performs lefts joins
//   - assets progress table to set `video_pos`, `completed`, and `completed_at`
//
// It then gets the attachments for each asset
func GetAssets(db database.Database, params *database.DatabaseParams) ([]*Asset, error) {
	var assets []*Asset

	cols := []string{
		TableAssets() + ".*",
		TableAssetsProgress() + ".video_pos",
		TableAssetsProgress() + ".completed",
		TableAssetsProgress() + ".completed_at",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableAssets()).
		LeftJoin(TableAssetsProgress() + " ON " + TableAssets() + ".id = " + TableAssetsProgress() + ".asset_id")

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
			if builder, err = paginate(db, params, builder, CountAssets); err != nil {
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
		a, err := scanAssetRow(rows)
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

		// Get the asset IDs
		assetIds := []string{}
		for _, a := range assets {
			assetIds = append(assetIds, a.ID)
		}

		// Get the attachments
		attachments, err := GetAttachments(db, &database.DatabaseParams{Where: sq.Eq{"asset_id": assetIds}})
		if err != nil {
			return nil, err
		}

		// Store in a map for easy lookup
		attachmentsMap := map[string][]*Attachment{}
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

// GetAsset selects an asset for the given ID
//
// It performs a left join to  set `video_pos`, `completed`, and `completed_at` from assets
// progress and then gets the attachments for the asset
func GetAsset(db database.Database, id string) (*Asset, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	cols := []string{
		TableAssets() + ".*",
		TableAssetsProgress() + ".video_pos",
		TableAssetsProgress() + ".completed",
		TableAssetsProgress() + ".completed_at",
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Select(cols...).
		From(TableAssets()).
		LeftJoin(TableAssetsProgress() + " ON " + TableAssets() + ".id = " + TableAssetsProgress() + ".asset_id").
		Where(sq.Eq{TableAssets() + ".id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := db.QueryRow(query, args...)
	if row.Err() != nil {
		return nil, row.Err()
	}

	asset, err := scanAssetRow(row)
	if err != nil {
		return nil, err
	}

	// Get the attachments
	attachments, err := GetAttachments(db, &database.DatabaseParams{Where: sq.Eq{"asset_id": asset.ID}})
	if err != nil {
		return nil, err
	}

	asset.Attachments = attachments

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset inserts a new asset
func CreateAsset(db database.Database, a *Asset) error {
	if a.Prefix.Valid && a.Prefix.Int16 < 0 {
		return fmt.Errorf("prefix must be greater than 0")
	}

	a.RefreshId()
	a.RefreshCreatedAt()
	a.RefreshUpdatedAt()

	builder := sq.StatementBuilder.
		Insert(TableAssets()).
		Columns("id", "course_id", "title", "prefix", "chapter", "type", "path", "created_at", "updated_at").
		Values(a.ID, NilStr(a.CourseID), NilStr(a.Title), a.Prefix, NilStr(a.Chapter), NilStr(a.Type.String()), NilStr(a.Path), a.CreatedAt, a.UpdatedAt)

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAsset deletes an asset with the given ID
func DeleteAsset(db database.Database, id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	builder := sq.StatementBuilder.
		PlaceholderFormat(sq.Question).
		Delete(TableAssets()).
		Where(sq.Eq{"id": id})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.Exec(query, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestAssets creates n number of assets for each course in the slice. If a db is provided, a DB
// insert will be performed
//
// THIS IS FOR TESTING PURPOSES
func NewTestAssets(t *testing.T, db database.Database, courses []*Course, assetsPerCourse int) []*Asset {
	assets := []*Asset{}

	for i := 0; i < len(courses); i++ {
		for j := 0; j < assetsPerCourse; j++ {
			a := &Asset{}

			a.RefreshId()
			a.RefreshCreatedAt()
			a.RefreshUpdatedAt()

			a.CourseID = courses[i].ID
			a.Title = security.PseudorandomString(6)
			a.Prefix = sql.NullInt16{Int16: int16(rand.Intn(100-1) + 1), Valid: true}
			a.Chapter = fmt.Sprintf("%d chapter %s", j+1, security.PseudorandomString(2))
			a.Type = *types.NewAsset("mp4")
			a.Path = fmt.Sprintf("%s/%s/%d %s.mp4", courses[i].Path, a.Chapter, a.Prefix.Int16, a.Title)

			if db != nil {
				err := CreateAsset(db, a)
				require.Nil(t, err)

				// This allows the created/updated times to be different when inserting multiple rows
				time.Sleep(time.Millisecond * 1)
			}

			assets = append(assets, a)
		}
	}

	return assets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// scanAssetRow scans an asset row
func scanAssetRow(scannable Scannable) (*Asset, error) {
	var a Asset

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
		// Course progress
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
