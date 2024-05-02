package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestFsPath(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		appFs.Fs.MkdirAll("/", os.ModePerm)
		appFs.Fs.MkdirAll("/dir1", os.ModePerm)
		appFs.Fs.Create("/file1")
		appFs.Fs.Create("/file2")
		appFs.Fs.Create("/file3")

		status, body, err := fsRequestHelper(appFs, db, http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"))
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, 4, respData.Count)
		require.Len(t, respData.Directories, 1)
		require.Len(t, respData.Files, 3)
		require.Equal(t, types.PathClassificationNone, respData.Directories[0].Classification)
	})

	t.Run("200 (path classifications)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		testData := daos.NewTestBuilder(t).Db(db).Courses([]string{"course 1", "course 2", "course 3"}).Build()

		// Create directories for the courses above
		for _, data := range testData {
			appFs.Fs.MkdirAll(data.Path, os.ModePerm)
		}

		// Create additional directories at the root
		appFs.Fs.MkdirAll("/dir1", os.ModePerm)
		appFs.Fs.MkdirAll("/dir2", os.ModePerm)

		// Create sub-directory for course 3
		appFs.Fs.MkdirAll(testData[2].Path+"/dir1", os.ModePerm)

		// ----------------------------
		// Get / (test ancestors and none)
		// ----------------------------
		status, body, err := fsRequestHelper(appFs, db, http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"))
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, 5, respData.Count)
		require.Len(t, respData.Directories, 5)
		require.Len(t, respData.Files, 0)

		for _, dir := range respData.Directories {
			if dir.Path == "/dir1" || dir.Path == "/dir2" {
				require.Equal(t, types.PathClassificationNone, dir.Classification)
			} else {
				require.Equal(t, types.PathClassificationAncestor, dir.Classification)
			}
		}

		// ----------------------------
		// Get directory above 'course 2'
		// ----------------------------
		status, body, err = fsRequestHelper(appFs, db, http.MethodGet, "/api/filesystem/"+utils.EncodeString(strings.TrimSuffix(testData[1].Path, "/course 2")))
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, 1, respData.Count)
		require.Len(t, respData.Directories, 1)
		require.Len(t, respData.Files, 0)
		require.Equal(t, types.PathClassificationCourse, respData.Directories[0].Classification)

	})

	t.Run("404 (path not found)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		status, _, err := fsRequestHelper(appFs, db, http.MethodGet, "/api/filesystem/"+utils.EncodeString("nonexistent/path"))
		require.Nil(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("400 (decode error)", func(t *testing.T) {
		appFs, db, _, _ := setup(t)

		status, body, err := fsRequestHelper(appFs, db, http.MethodGet, "/api/filesystem/`")

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "failed to decode path", string(body))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func fsRequestHelper(appFs *appFs.AppFs, db database.Database, method string, target string) (int, []byte, error) {
	f := fiber.New()
	bindFsApi(f.Group("/api"), appFs, db)

	resp, err := f.Test(httptest.NewRequest(method, target, nil))
	if err != nil {
		return -1, nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}
