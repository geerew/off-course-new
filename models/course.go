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

	Title       string `bun:",notnull,default:null"`
	Path        string `bun:",unique,notnull,default:null"`
	CardPath    string
	Started     bool
	Percent     int `bun:",notnull,default:0"`
	CompletedAt types.DateTime
	ScanStatus  string

	// Has many
	Assets []*Asset `bun:"rel:has-many,join:id=course_id"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CountCourses counts the number of courses
func CountCourses(ctx context.Context, db database.Database, params *database.DatabaseParams) (int, error) {
	q := db.DB().NewSelect().Model((*Course)(nil))

	if params != nil && params.Where != nil {
		q = selectWhere(q, params.Where, "course")
	}

	return q.Count(ctx)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourses selects courses
func GetCourses(ctx context.Context, db database.Database, params *database.DatabaseParams) ([]*Course, error) {
	var courses []*Course

	// Create a query that joins the scans table, selecting the scan status
	q := db.DB().
		NewSelect().
		Model(&courses).
		ColumnExpr("Course.*").
		ColumnExpr("?.? AS ?",
			bun.Ident("scans"),
			bun.Ident("status"),
			bun.Ident("scan_status")).
		Join("LEFT JOIN ?0 ON (?0.?1 = ?2.?3)",
			bun.Ident("scans"),
			bun.Ident("course_id"),
			bun.Ident("course"),
			bun.Ident("id"))

	if params != nil {
		// Pagination
		if params.Pagination != nil {
			if count, err := CountCourses(ctx, db, params); err != nil {
				return nil, err
			} else {
				params.Pagination.SetCount(count)
			}

			q = q.Offset(params.Pagination.Offset()).Limit(params.Pagination.Limit())
		}

		// Relations
		if params.Relation != nil {
			q = selectRelation(q, params.Relation)
		}

		// Order by
		if len(params.OrderBy) > 0 {
			selectOrderBy(q, params.OrderBy, "course")
		}
		// Where
		if params.Where != nil {
			q = selectWhere(q, params.Where, "course")
		}
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return courses, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourse selects a course based upon the where clause in the database params
func GetCourse(ctx context.Context, db database.Database, params *database.DatabaseParams) (*Course, error) {
	if params == nil || params.Where == nil {
		return nil, errors.New("where clause required")
	}

	course := &Course{}

	q := db.DB().
		NewSelect().
		Model(course).
		ColumnExpr("Course.*").
		ColumnExpr("?.? AS ?",
			bun.Ident("scans"),
			bun.Ident("status"),
			bun.Ident("scan_status")).
		Join("LEFT JOIN ?0 ON (?0.?1 = ?2.?3)",
			bun.Ident("scans"),
			bun.Ident("course_id"),
			bun.Ident("course"),
			bun.Ident("id"))

	// Where
	if params.Where != nil {
		q = selectWhere(q, params.Where, "course")
	}

	// Relations
	if params.Relation != nil {
		q = selectRelation(q, params.Relation)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetCourseById selects a course for the given ID
func GetCourseById(ctx context.Context, db database.Database, params *database.DatabaseParams, id string) (*Course, error) {
	course := &Course{}

	q := db.DB().
		NewSelect().
		Model(course).
		ColumnExpr("Course.*").
		ColumnExpr("?.? AS ?",
			bun.Ident("scans"),
			bun.Ident("status"),
			bun.Ident("scan_status")).
		Join("LEFT JOIN ?0 ON (?0.?1 = ?2.?3)",
			bun.Ident("scans"),
			bun.Ident("course_id"),
			bun.Ident("course"),
			bun.Ident("id")).
		Where("Course.id = ?", id)

	if params != nil && params.Relation != nil {
		q = selectRelation(q, params.Relation)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateCourse inserts a new course
func CreateCourse(ctx context.Context, db database.Database, course *Course) error {
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

// UpdateCourseCardPath updates `card_path`
func UpdateCourseCardPath(ctx context.Context, db database.Database, id string, newCardPath string) (*Course, error) {
	// Require an ID
	if id == "" {
		return nil, errors.New("course ID cannot be empty")
	}

	// Get the course
	course, err := GetCourseById(ctx, db, nil, id)
	if err != nil {
		return nil, err
	}

	// Nothing to do
	if course.CardPath == newCardPath {
		return course, nil
	}

	ts := types.NowDateTime()

	if res, err := db.DB().NewUpdate().Model(course).
		Set("card_path = ?", newCardPath).
		Set("updated_at = ?", ts).
		WherePK().Exec(ctx); err != nil {
		return nil, err
	} else {
		count, _ := res.RowsAffected()
		if count == 0 {
			return course, nil
		}
	}

	// Update the original course struct
	course.CardPath = newCardPath
	course.UpdatedAt = ts

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCoursePercent updates `percent` and `completed_at` (if percent is 100)
func UpdateCoursePercent(ctx context.Context, db database.Database, id string, percent int) (*Course, error) {
	// Require an ID
	if id == "" {
		return nil, errors.New("course ID cannot be empty")
	}

	// Get the course
	course, err := GetCourseById(ctx, db, nil, id)
	if err != nil {
		return nil, err
	}

	// Nothing to do
	if course.Percent == percent {
		return course, nil
	}

	// Keep the percent between 0 and 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	updatedAt := types.NowDateTime()

	// Set the completed at date when the percent is 100
	var completedAt types.DateTime
	if percent == 100 {
		completedAt = updatedAt
	}

	if res, err := db.DB().NewUpdate().Model(course).
		Set("percent = ?", percent).
		Set("completed_at = ?", completedAt).
		Set("updated_at = ?", updatedAt).
		WherePK().Exec(ctx); err != nil {
		return nil, err
	} else {
		// Nothing was changed so return
		count, _ := res.RowsAffected()
		if count == 0 {
			return course, nil
		}
	}

	course.Percent = percent
	course.CompletedAt = completedAt
	course.UpdatedAt = updatedAt

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateCourseCompleted updates `completed` and `completed_at`
func UpdateCourseUpdatedAt(ctx context.Context, db database.Database, id string) (*Course, error) {
	// Require an ID
	if id == "" {
		return nil, errors.New("course ID cannot be empty")
	}

	// Get the course
	course, err := GetCourseById(ctx, db, nil, id)
	if err != nil {
		return nil, err
	}

	ts := types.NowDateTime()

	if res, err := db.DB().NewUpdate().Model(course).
		Set("updated_at = ?", ts).
		WherePK().Exec(ctx); err != nil {
		return nil, err
	} else {
		// Nothing was changed so return
		count, _ := res.RowsAffected()
		if count == 0 {
			return course, nil
		}
	}

	course.UpdatedAt = ts

	return course, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteCourse deletes a course with the given ID
func DeleteCourse(ctx context.Context, db database.Database, id string) (int, error) {
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
			Path:  fmt.Sprintf("/%s/%s", security.PseudorandomString(5), security.PseudorandomString(5)),
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
