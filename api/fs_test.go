package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestFsPath(t *testing.T) {
	t.Run("200 (found)", func(t *testing.T) {
		appFs, _, _, _ := setup(t)

		appFs.Fs.MkdirAll("/", os.ModePerm)
		appFs.Fs.MkdirAll("/dir1", os.ModePerm)
		appFs.Fs.Create("/file1")
		appFs.Fs.Create("/file2")
		appFs.Fs.Create("/file3")

		status, body, err := fsRequestHelper(appFs, http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"))
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, status)

		var respData fileSystemResponse
		err = json.Unmarshal(body, &respData)
		require.Nil(t, err)
		require.Equal(t, respData.Count, 4)
		require.Len(t, respData.Directories, 1)
		require.Len(t, respData.Files, 3)
	})

	t.Run("404 (path not found)", func(t *testing.T) {
		appFs, _, _, _ := setup(t)

		status, _, err := fsRequestHelper(appFs, http.MethodGet, "/api/filesystem/"+utils.EncodeString("nonexistent/path"))
		require.Nil(t, err)
		require.Equal(t, http.StatusNotFound, status)
	})

	t.Run("400 (decode error)", func(t *testing.T) {
		appFs, _, _, _ := setup(t)

		status, body, err := fsRequestHelper(appFs, http.MethodGet, "/api/filesystem/`")

		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, status)
		require.Equal(t, "failed to decode path", string(body))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// HELPERS
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func fsRequestHelper(appFs *appFs.AppFs, method string, target string) (int, []byte, error) {
	f := fiber.New()
	bindFsApi(f.Group("/api"), appFs)

	resp, err := f.Test(httptest.NewRequest(method, target, nil))
	if err != nil {
		return -1, nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}
