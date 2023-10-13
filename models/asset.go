package models

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Asset struct {
	BaseModel
	CourseID string `bun:",notnull"`
	Title    string `bun:",notnull,default:null"`
	Prefix   int    `bun:",notnull,default:null"`
	Chapter  string
	Type     types.Asset `bun:",notnull,default:null"`
	Path     string      `bun:",unique,notnull,default:null"`
	Started  bool
	Finished bool

	// Belongs to
	Course *Course `bun:"rel:belongs-to,join:course_id=id"`

	// Has many
	Attachments []*Attachment `bun:"rel:has-many,join:id=asset_id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountAssets returns the number of attachments
func CountAssets(ctx context.Context, db database.Database, params *database.DatabaseParams) (int, error) {
	q := db.DB().NewSelect().Model((*Asset)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params, "asset")
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssets returns a slice of assets
func GetAssets(ctx context.Context, db database.Database, params *database.DatabaseParams) ([]*Asset, error) {
	var assets []*Asset

	q := db.DB().NewSelect().Model(&assets)

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			if count, err := CountAssets(ctx, db, params); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		if params.Relation != nil {
			q = selectRelation(q, params)
		}

		// Order by
		if len(params.OrderBy) > 0 {
			q = q.Order(params.OrderBy...)
		}

		// Where
		if params.Where != nil {
			if params.Where != nil {
				q = selectWhere(q, params, "asset")
			}
		}
	}

	err := q.Scan(ctx)

	return assets, err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAsset returns an asset based upon the where clause in the database params
func GetAsset(ctx context.Context, db database.Database, params *database.DatabaseParams) (*Asset, error) {
	if params == nil || params.Where == nil {
		return nil, errors.New("where clause required")
	}
	asset := &Asset{}

	q := db.DB().NewSelect().Model(asset)

	// Where
	if params.Where != nil {
		q = selectWhere(q, params, "asset")
	}

	// Relations
	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return asset, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAsset inserts a new asset
func CreateAsset(ctx context.Context, db database.Database, asset *Asset) error {
	if asset.Prefix < 0 {
		return fmt.Errorf("prefix must be greater than 0")
	}

	asset.RefreshId()
	asset.RefreshCreatedAt()
	asset.RefreshUpdatedAt()

	_, err := db.DB().NewInsert().Model(asset).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAsset deletes an asset with the given ID
func DeleteAsset(ctx context.Context, db database.Database, id string) (int, error) {
	asset := &Asset{}
	asset.SetId(id)

	if res, err := db.DB().NewDelete().Model(asset).WherePK().Exec(ctx); err != nil {
		return 0, err
	} else {
		count, _ := res.RowsAffected()
		return int(count), err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestAssets creates n number of assets for each course in the slice. If a db is provided, the
// assets will be inserted into the db
//
// THIS IS FOR TESTING PURPOSES
func NewTestAssets(t *testing.T, db database.Database, courses []*Course, assetsPerCourse int) []*Asset {
	assets := []*Asset{}
	for i := 0; i < len(courses); i++ {
		for j := 0; j < assetsPerCourse; j++ {
			title := fmt.Sprintf("%s.mp4", security.PseudorandomString(8))
			prefix := rand.Intn(100-1) + 1
			chapter := fmt.Sprintf("chapter %s", security.PseudorandomString(2))

			a := &Asset{
				CourseID: courses[i].ID,
				Title:    title,
				Prefix:   prefix,
				Chapter:  chapter,
				Type:     *types.NewAsset("mp4"),
				Path:     fmt.Sprintf("%s/%s/%d %s", courses[i].Path, chapter, prefix, title),
				Started:  false,
				Finished: false,
			}

			a.RefreshId()
			a.RefreshCreatedAt()
			a.RefreshUpdatedAt()

			if db != nil {
				_, err := db.DB().NewInsert().Model(a).Exec(context.Background())
				require.Nil(t, err)
			}

			assets = append(assets, a)
			time.Sleep(1 * time.Millisecond)
		}
	}

	return assets
}
