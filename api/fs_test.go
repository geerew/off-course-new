package api

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"

// 	"github.com/geerew/off-course/utils"
// 	"github.com/gofiber/fiber/v2"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// func TestFsPath(t *testing.T) {
// 	t.Run("200 (found)", func(t *testing.T) {
// 		appFs, _, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		appFs.Fs.MkdirAll("/", os.ModePerm)
// 		appFs.Fs.MkdirAll("/dir1", os.ModePerm)
// 		appFs.Fs.Create("/file1")
// 		appFs.Fs.Create("/file2")
// 		appFs.Fs.Create("/file3")

// 		f := fiber.New()
// 		bindFsApi(f.Group("/api"), appFs)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("/"), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusOK, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)

// 		var respData fileSystemResponse
// 		err = json.Unmarshal(body, &respData)
// 		require.Nil(t, err)
// 		assert.Equal(t, respData.Count, 4)
// 		assert.Len(t, respData.Directories, 1)
// 		assert.Len(t, respData.Files, 3)
// 	})

// 	t.Run("404 (path not found)", func(t *testing.T) {
// 		appFs, _, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindFsApi(f.Group("/api"), appFs)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/filesystem/"+utils.EncodeString("nonexistent/path"), nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusNotFound, resp.StatusCode)
// 	})

// 	t.Run("400 (decode error)", func(t *testing.T) {
// 		appFs, _, _, _, teardown := setup(t)
// 		defer teardown(t)

// 		f := fiber.New()
// 		bindFsApi(f.Group("/api"), appFs)

// 		resp, err := f.Test(httptest.NewRequest(http.MethodGet, "/api/filesystem/`", nil))
// 		assert.NoError(t, err)
// 		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

// 		body, _ := io.ReadAll(resp.Body)
// 		assert.Equal(t, "failed to decode path", string(body))
// 	})
// }
