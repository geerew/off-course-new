package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestFsPath(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		router, _ := setup(t)

		router.config.AppFs.Fs.MkdirAll("/dir1", os.ModePerm)
		router.config.AppFs.Fs.Create("/file1")
		router.config.AppFs.Fs.Create("/file2")
		router.config.AppFs.Fs.Create("/file3")

		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"), nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, 4, respData.Count)
		require.Len(t, respData.Directories, 1)
		require.Len(t, respData.Files, 3)
		require.Equal(t, types.PathClassificationNone, respData.Directories[0].Classification)
	})

	t.Run("200 (path classifications)", func(t *testing.T) {
		router, ctx := setup(t)

		// Create /dir 1, /course 1, /courses/course 2
		router.config.AppFs.Fs.MkdirAll("/dir 1", os.ModePerm)

		course1 := &models.Course{Title: "course 1", Path: "/course 1"}
		require.NoError(t, router.dao.CreateCourse(ctx, course1))
		require.NoError(t, router.config.AppFs.Fs.MkdirAll(course1.Path, os.ModePerm))

		course2 := &models.Course{Title: "course 2", Path: "/courses/course 2"}
		require.NoError(t, router.dao.CreateCourse(ctx, course2))
		require.NoError(t, router.config.AppFs.Fs.MkdirAll(course2.Path, os.ModePerm))

		// Test /
		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"), nil)
		status, body, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.NoError(t, err)
		require.Equal(t, 3, respData.Count)
		require.Len(t, respData.Directories, 3)
		require.Len(t, respData.Files, 0)

		require.Equal(t, types.PathClassificationCourse, respData.Directories[0].Classification)   // /course 1
		require.Equal(t, types.PathClassificationAncestor, respData.Directories[1].Classification) // /courses
		require.Equal(t, types.PathClassificationNone, respData.Directories[2].Classification)     // /dir 1
	})

	t.Run("404 (path not found)", func(t *testing.T) {
		router, _ := setup(t)
		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/other"), nil)
		status, _, err := requestHelper(t, router, req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("400 (decode error)", func(t *testing.T) {
		router, _ := setup(t)

		req := httptest.NewRequest(http.MethodGet, "/api/filesystem/`", nil)
		status, body, err := requestHelper(t, router, req)

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "failed to decode path", string(body))
	})
}
