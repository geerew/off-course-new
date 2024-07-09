package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geerew/off-course/daos"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestFsPath(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router := setup(t)

		router.config.AppFs.Fs.MkdirAll(string(filepath.Separator)+"dir1", os.ModePerm)
		router.config.AppFs.Fs.Create(string(filepath.Separator) + "file1")
		router.config.AppFs.Fs.Create(string(filepath.Separator) + "file2")
		router.config.AppFs.Fs.Create(string(filepath.Separator) + "file3")

		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"), nil)
		status, body, err := requestHelper(t, router, req)
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
		router := setup(t)

		testData := daos.NewTestBuilder(t).Db(router.config.DbManager.DataDb).Courses(3).Build()

		// Create directories for the courses above
		for _, data := range testData {
			router.config.AppFs.Fs.MkdirAll(data.Path, os.ModePerm)
		}

		// Create additional directories at the root
		router.config.AppFs.Fs.MkdirAll(string(filepath.Separator)+"dir1", os.ModePerm)
		router.config.AppFs.Fs.MkdirAll(string(filepath.Separator)+"dir2", os.ModePerm)

		// Create sub-directory for course 3
		router.config.AppFs.Fs.MkdirAll(filepath.Join(testData[2].Path, "dir1"), os.ModePerm)

		// ----------------------------
		// Get / (test ancestors and none)
		// ----------------------------
		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"), nil)
		status, body, err := requestHelper(t, router, req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, 3, respData.Count)
		require.Len(t, respData.Directories, 3)
		require.Len(t, respData.Files, 0)

		for _, dir := range respData.Directories {
			if dir.Path == string(filepath.Separator)+"dir1" || dir.Path == string(filepath.Separator)+"dir2" {
				require.Equal(t, types.PathClassificationNone, dir.Classification)
			} else {
				require.Equal(t, types.PathClassificationAncestor, dir.Classification)
			}
		}

		// ----------------------------
		// Get directory above 'course 2'
		// ----------------------------
		req = httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString(strings.TrimSuffix(testData[1].Path, string(filepath.Separator)+"Course 2")), nil)
		status, body, err = requestHelper(t, router, req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, 3, respData.Count)
		require.Len(t, respData.Directories, 3)
		require.Len(t, respData.Files, 0)
		require.Equal(t, types.PathClassificationCourse, respData.Directories[0].Classification)
		require.Equal(t, types.PathClassificationCourse, respData.Directories[1].Classification)
		require.Equal(t, types.PathClassificationCourse, respData.Directories[2].Classification)

	})

	t.Run("404 (path not found)", func(t *testing.T) {
		router := setup(t)
		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("nonexistent/path"), nil)
		status, _, err := requestHelper(t, router, req)
		require.Nil(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("400 (decode error)", func(t *testing.T) {
		router := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/`", nil)
		status, body, err := requestHelper(t, router, req)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "failed to decode path", string(body))
	})
}
