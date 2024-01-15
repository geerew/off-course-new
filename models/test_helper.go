package models

import (
	"database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"testing"
	"time"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TestCourses struct {
	*Course
	Scan   *Scan
	Assets []*Asset
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewTestData creates test data for the given database. If the database is nil, the data is not
// saved to the database.
func NewTestData(t *testing.T, db database.Database, numberOfCourses int, scan bool, assetsPerCourse int, attachmentsPerAsset int) []*TestCourses {
	var testCourses []*TestCourses

	for i := 0; i < numberOfCourses; i++ {
		tc := &TestCourses{}

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
func newTestCourse(t *testing.T, db database.Database) *Course {
	c := &Course{}

	c.RefreshId()
	c.RefreshCreatedAt()
	c.RefreshUpdatedAt()

	c.Title = fmt.Sprintf("Course %s", security.PseudorandomString(5))
	c.Path = fmt.Sprintf("/%s/%s", security.PseudorandomString(5), c.Title)

	if db != nil {
		err := CreateCourse(db, c)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return c
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newTestScan(t *testing.T, db database.Database, courseId string) *Scan {
	s := &Scan{}

	s.RefreshId()
	s.RefreshCreatedAt()
	s.RefreshUpdatedAt()

	s.CourseID = courseId
	s.Status = types.NewScanStatus(types.ScanStatusWaiting)

	if db != nil {
		err := CreateScan(db, s)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return s
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newTestAssets(t *testing.T, db database.Database, course *Course, numberOfAssets int, numberOfAttachments int) []*Asset {
	assets := []*Asset{}

	for i := 0; i < numberOfAssets; i++ {
		a := &Asset{}

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
			err := CreateAsset(db, a)
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

func newTestAttachments(t *testing.T, db database.Database, asset *Asset, numberOfAttachments int) []*Attachment {
	attachments := []*Attachment{}

	for i := 0; i < numberOfAttachments; i++ {
		a := &Attachment{}

		a.RefreshId()
		a.RefreshCreatedAt()
		a.RefreshUpdatedAt()

		a.CourseID = asset.CourseID
		a.AssetID = asset.ID
		a.Title = security.PseudorandomString(6)
		a.Path = fmt.Sprintf("%s/%d %s", filepath.Dir(asset.Path), asset.Prefix.Int16, a.Title)

		if db != nil {
			err := CreateAttachment(db, a)
			require.Nil(t, err)

			// This allows the created/updated times to be different when inserting multiple rows
			time.Sleep(time.Millisecond * 1)
		}

		attachments = append(attachments, a)

	}

	return attachments
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newTestCoursesProgress(t *testing.T, db database.Database, courseId string) *CourseProgress {
	cp := &CourseProgress{}

	cp.RefreshId()
	cp.RefreshCreatedAt()
	cp.RefreshUpdatedAt()

	cp.CourseID = courseId

	if db != nil {
		err := CreateCourseProgress(db, cp)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return cp
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func newTestAssetsProgress(t *testing.T, db database.Database, assetId string, courseId string) *AssetProgress {
	ap := &AssetProgress{}

	ap.RefreshId()
	ap.RefreshCreatedAt()
	ap.RefreshUpdatedAt()

	ap.AssetID = assetId
	ap.CourseID = courseId

	if db != nil {
		err := CreateAssetProgress(db, ap)
		require.Nil(t, err)

		// This allows the created/updated times to be different when inserting multiple rows
		time.Sleep(time.Millisecond * 1)
	}

	return ap
}
