package daos

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scannable is an interface for a database row that can be scanned into a struct
type Scannable interface {
	Scan(dest ...interface{}) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Errors
var (
	ErrEmptyId       = errors.New("id cannot be empty")
	ErrMissingWhere  = errors.New("where clause cannot be empty")
	ErrInvalidPrefix = errors.New("prefix must be greater than 0")
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NilStr returns nil when a string is empty
//
// Use this when inserting into the database to avoid inserting empty strings
func NilStr(s string) any {
	if s == "" {
		return nil
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// extractTableColumn extracts the table and column name from an orderBy string. If no table prefix
// is found, the table part is returned as an empty string
func extractTableColumn(orderBy string) (string, string) {
	parts := strings.Fields(orderBy)
	tableColumn := strings.Split(parts[0], ".")

	if len(tableColumn) == 2 {
		return tableColumn[0], tableColumn[1]
	}

	return "", tableColumn[0]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// isValidOrderBy returns true if the orderBy string is valid. The table and column are validated
// against the given list of valid table.columns (ex. courses.id, scans.status as scan_status).
func isValidOrderBy(table, column string, validateTableColumns []string) bool {
	// If the column is empty, always return false
	if column == "" {
		return false
	}

	for _, validTc := range validateTableColumns {
		// Wildcard match (ex. courses.* == id)
		if table == "" && strings.HasSuffix(validTc, ".*") {
			return true
		}

		// Exact match (ex. id == id || courses.id == courses.id || courses.id as courses_id == courses.id)
		if validTc == column || validTc == table+"."+column || strings.HasPrefix(validTc, table+"."+column+" as ") {
			return true
		}

		// Table + wildcard match (ex. courses.* == courses.id)
		if strings.HasSuffix(validTc, ".*") && strings.HasPrefix(validTc, table+".") {
			return true
		}

		// courses.id as course_id == course_id
		if strings.HasSuffix(validTc, " as "+column) {
			return true
		}
	}

	return false
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// TEST DATA HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestCourse struct {
	*models.Course
	Scan   *models.Scan
	Assets []*models.Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestData creates test data for the given database. If the database is nil, the data is not
// saved to the database.
func NewTestData(
	t *testing.T,
	db database.Database,
	numberOfCourses int,
	scan bool,
	assetsPerCourse int,
	attachmentsPerAsset int) []*TestCourse {

	var testCourses []*TestCourse

	for i := 0; i < numberOfCourses; i++ {
		tc := &TestCourse{}

		tc.Course = newTestCourse(t, db)

		// Add a scan to the course
		if scan {
			tc.Scan = newTestScan(t, db, tc.Course.ID)
		}

		// Add assets to the course
		if assetsPerCourse > 0 {
			tc.Assets = newTestAssets(t, db, tc.Course, assetsPerCourse, attachmentsPerAsset)
		}

		testCourses = append(testCourses, tc)
	}

	return testCourses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newTestCourse(t *testing.T, db database.Database) *models.Course {
	c := &models.Course{}

	c.RefreshId()
	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	c.Title = fmt.Sprintf("Course %s", security.PseudorandomString(5))
	c.Path = fmt.Sprintf("/%s/%s", security.PseudorandomString(5), c.Title)

	if db != nil {
		dao := NewCourseDao(db)

		err := dao.Create(c)
		require.NoError(t, err, "Failed to create course")

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return c
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newTestScan creates a new scan for the given course id. If the database is nil, the scan is not
// saved to the database
func newTestScan(t *testing.T, db database.Database, courseId string) *models.Scan {
	s := &models.Scan{}

	s.RefreshId()
	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	s.CourseID = courseId
	s.Status = types.NewScanStatus(types.ScanStatusWaiting)

	if db != nil {
		dao := NewScanDao(db)

		err := dao.Create(s)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newTestScan creates n assets for the given course and n attachments for each asset. If the
// database is nil, the asset/attachments will not be saved to the database
func newTestAssets(t *testing.T, db database.Database, course *models.Course, numberOfAssets int, numberOfAttachments int) []*models.Asset {
	assets := []*models.Asset{}

	for i := 0; i < numberOfAssets; i++ {
		a := &models.Asset{}

		a.RefreshId()
		a.RefreshCreatedAt()
		a.RefreshUpdatedAt()

		a.CourseID = course.ID
		a.Title = security.PseudorandomString(6)
		a.Prefix = sql.NullInt16{Int16: int16(rand.Intn(100-1) + 1), Valid: true}
		a.Chapter = fmt.Sprintf("%d chapter %s", i+1, security.PseudorandomString(2))
		a.Type = *types.NewAsset("mp4")
		a.Path = fmt.Sprintf("%s/%s/%d %s.mp4", course.Path, a.Chapter, a.Prefix.Int16, a.Title)

		if db != nil {
			dao := NewAssetDao(db)

			err := dao.Create(a)
			require.Nil(t, err)

			// This allows the created/updated times to be different when inserting multiple rows
			time.Sleep(time.Millisecond * 1)
		}

		// Add attachments to the asset
		if numberOfAttachments > 0 {
			a.Attachments = newTestAttachments(t, db, a, numberOfAttachments)
		}

		assets = append(assets, a)
	}

	return assets
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newTestAttachments creates n attachments for the given asset. If the database is nil, the
// attachments will not be saved to the database
func newTestAttachments(t *testing.T, db database.Database, asset *models.Asset, numberOfAttachments int) []*models.Attachment {
	attachments := []*models.Attachment{}

	for i := 0; i < numberOfAttachments; i++ {
		a := &models.Attachment{}

		a.RefreshId()
		a.RefreshCreatedAt()
		a.RefreshUpdatedAt()

		a.CourseID = asset.CourseID
		a.AssetID = asset.ID
		a.Title = security.PseudorandomString(6)
		a.Path = fmt.Sprintf("%s/%d %s", filepath.Dir(asset.Path), asset.Prefix.Int16, a.Title)

		if db != nil {
			dao := NewAttachmentDao(db)

			err := dao.Create(a)
			require.Nil(t, err)

			// This allows the created/updated times to be different when inserting multiple rows
			time.Sleep(time.Millisecond * 1)
		}

		attachments = append(attachments, a)

	}

	return attachments
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// newTestAssetsProgress creates a new asset progress for the given asset id and course id. If the
// database is nil, the asset progress is not saved to the database
func newTestAssetsProgress(t *testing.T, db database.Database, assetId string, courseId string) *models.AssetProgress {
	ap := &models.AssetProgress{}

	ap.RefreshId()
	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	ap.AssetID = assetId
	ap.CourseID = courseId

	if db != nil {
		dao := NewAssetProgressDao(db)

		err := dao.Create(ap)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return ap
}
