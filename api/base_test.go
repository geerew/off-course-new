package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RouterLogger(t *testing.T) {
	appFs, db, courseScanner, lh := setup(t)

	router := New(&RouterConfig{
		Db:            db,
		AppFs:         appFs,
		CourseScanner: courseScanner,
		Port:          ":1234",
		IsProduction:  false,
	})

	_, err := router.router.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	assert.NoError(t, err)

	lh.LastEntry().ExpMsg("Request completed")
	lh.LastEntry().ExpStr("path", "/")
	lh.LastEntry().ExpStr("method", "GET")
}
