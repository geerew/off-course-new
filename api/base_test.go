package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RouterLogger(t *testing.T) {
	router := setup(t)

	_, err := router.router.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	require.NoError(t, err)
}
