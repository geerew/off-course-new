package models

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type Course struct {
	BaseModel

	Title      string `bun:",notnull,default:null"`
	Path       string `bun:",unique,notnull,default:null"`
	CardPath   string
	Started    bool `bun:",notnull,default:false"`
	Finished   bool `bun:",notnull,default:false"`
	ScanStatus string

	// Has many
	Assets []*Asset `bun:"rel:has-many,join:id=course_id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountCourses returns the number of courses
func CountCourses(db database.Database, params *database.DatabaseParams, ctx context.Context) (int, error) {
	q := db.DB().NewSelect().Model((*Course)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params)
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourses returns a slice of courses
func GetCourses(db database.Database, params *database.DatabaseParams, ctx context.Context) ([]*Course, error) {
	var courses []*Course

	// Create a query that joins the scans table, selecting the scan status
	q := db.DB().
		NewSelect().
		Model(&courses).
		ColumnExpr("Course.*").
		ColumnExpr("?.? AS ?", bun.Ident("scans"), bun.Ident("status"), bun.Ident("scan_status")).
		Join("LEFT JOIN ?0 ON (?0.?1 = ?2.?3)", bun.Ident("scans"), bun.Ident("course_id"), bun.Ident("course"), bun.Ident("id"))

	if params != nil {
		// 		// Pagination
		// 		if params.Pagination != nil {
		// 			if count, err := CountCourses(db, params); err != nil {
		// 				return nil, err
		// 			} else {
		// 				params.Pagination.SetCount(count)
		// 			}

		// 			q = q.Scopes(params.Pagination.Paginate())
		// 		}

		// Relations
		if params.Relation != nil {
			q = selectRelation(q, params)
		}

		// Order by
		if len(params.OrderBy) > 0 {
			q = q.Order(params.OrderBy...)
		}
		// Where
		if params.Where != nil {
			q = selectWhere(q, params)
		}
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return courses, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseById returns a course with the given ID
func GetCourseById(db database.Database, id string, params *database.DatabaseParams, ctx context.Context) (*Course, error) {
	course := &Course{}
	course.SetId(id)

	q := db.DB().
		NewSelect().
		Model(course).
		ColumnExpr("Course.*").
		ColumnExpr("?.? AS ?", bun.Ident("scans"), bun.Ident("status"), bun.Ident("scan_status")).
		Join("LEFT JOIN ?0 ON (?0.?1 = ?2.?3)", bun.Ident("scans"), bun.Ident("course_id"), bun.Ident("course"), bun.Ident("id")).
		WherePK()

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse creates a new course
func CreateCourse(db database.Database, course *Course, ctx context.Context) error {
	course.RefreshId()
	course.RefreshCreatedAt()
	course.RefreshUpdatedAt()

	_, err := db.DB().NewInsert().Model(course).Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseCardPath updates the course `card_path`
func UpdateCourseCardPath(db database.Database, course *Course, newCardPath string, ctx context.Context) error {
	// Do nothing when the card path is the same
	if course.CardPath == newCardPath {
		return nil
	}

	// Require an ID
	if course.ID == "" {
		return errors.New("course ID cannot be empty")
	}

	ts := types.NowDateTime()

	// Update the status
	if res, err := db.DB().NewUpdate().Model(course).
		Set("card_path = ?", newCardPath).
		Set("updated_at = ?", ts).
		WherePK().Exec(ctx); err != nil {
		return err
	} else {
		count, _ := res.RowsAffected()
		if count == 0 {
			return nil
		}
	}

	// Update the original scan struct
	course.CardPath = newCardPath
	course.UpdatedAt = ts

	return nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteCourse deletes a course with the given ID
func DeleteCourse(db database.Database, id string, ctx context.Context) (int, error) {
	course := &Course{}
	course.SetId(id)

	if res, err := db.DB().NewDelete().Model(course).WherePK().Exec(ctx); err != nil {
		return 0, err
	} else {
		count, _ := res.RowsAffected()
		return int(count), err
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestCourses creates n number of courses. If a db if provided, the courses will be inserted
// into the db
//
// THIS IS FOR TESTING PURPOSES
func NewTestCourses(t *testing.T, db database.Database, count int) []*Course {
	courses := []*Course{}

	for i := 0; i < count; i++ {
		c := &Course{
			Title: security.PseudorandomString(8),
			Path:  fmt.Sprintf("%s/%s", security.PseudorandomString(5), security.PseudorandomString(5)),
		}

		c.RefreshId()
		c.RefreshCreatedAt()
		c.RefreshUpdatedAt()

		if db != nil {
			_, err := db.DB().NewInsert().Model(c).Exec(context.Background())
			require.Nil(t, err)
		}

		courses = append(courses, c)
		time.Sleep(time.Millisecond * 1)
	}

	return courses
}
