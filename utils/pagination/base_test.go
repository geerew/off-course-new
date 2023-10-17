package pagination

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NewPagination(t *testing.T) {
	t.Run("no values", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(1)

		assert.Equal(t, 1, p.page)
		assert.Equal(t, DefaultPerPage, p.perPage)
		assert.Equal(t, 1, p.TotalItems())
	})

	t.Run("values", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=2" + "&" + PerPageQueryParam + "=10")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(24)

		assert.Equal(t, 2, p.page)
		assert.Equal(t, 24, p.TotalItems())
		assert.Equal(t, 3, p.totalPages)
	})

	t.Run("invalid values", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=-20" + "&" + PerPageQueryParam + "=bob")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(24)

		assert.Equal(t, 1, p.page)
		assert.Equal(t, DefaultPerPage, p.perPage)
		assert.Equal(t, 24, p.TotalItems())
		assert.Equal(t, 1, p.totalPages)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Limit(t *testing.T) {
	var tests = []struct {
		in       string
		expected int
	}{
		// Invalid
		{"1", 1},
		{"", DefaultPerPage},
		{"abc", DefaultPerPage},
		{"-1", DefaultPerPage},
		{"0", DefaultPerPage},
		// Valid
		{"5", 5},
	}

	for _, tt := range tests {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=1" + "&" + PerPageQueryParam + "=" + tt.in)
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(1)

		assert.Equal(t, tt.expected, p.Limit())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Offset(t *testing.T) {
	var tests = []struct {
		page     string
		perPage  string
		expected int
	}{
		{"", "", 0},
		{"abc", "def", 0},
		{"-1", "40", 0},
		{"0", "10", 0},
		{"1", "10", 0},
		{"2", "10", 10},
		{"5", "10", 40},
		{"20", "30", 570},
	}

	for _, tt := range tests {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=" + tt.page + "&" + PerPageQueryParam + "=" + tt.perPage)
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(1)

		assert.Equal(t, tt.expected, p.Offset())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Page(t *testing.T) {

	var tests = []struct {
		in       string
		expected int
	}{
		// Invalid
		{"1", 1},
		{"", 1},
		{"abc", 1},
		{"-1", 1},
		{"0", 1},
		// Valid
		{"5", 5},
	}

	for _, tt := range tests {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=" + tt.in)
		defer app.ReleaseCtx(c)

		assert.Equal(t, tt.expected, page(c))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_PerPage(t *testing.T) {
	var tests = []struct {
		in       string
		expected int
	}{
		// Invalid
		{"1", 1},
		{"", DefaultPerPage},
		{"abc", DefaultPerPage},
		{"-1", DefaultPerPage},
		{"0", DefaultPerPage},
		{fmt.Sprintf("%d", MaxPerPage+1), MaxPerPage},
		// Valid
		{"5", 5},
	}

	for _, tt := range tests {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PerPageQueryParam + "=" + tt.in)
		defer app.ReleaseCtx(c)

		assert.Equal(t, tt.expected, perPage(c))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_BuildResult(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=-20" + "&" + PerPageQueryParam + "=bob")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(24)

		type Data struct {
			ID        string         `json:"id"`
			CreatedAt types.DateTime `json:"createdAt"`
		}

		// The data to marshal
		data := []Data{
			{ID: "1", CreatedAt: types.NowDateTime()},
			{ID: "2", CreatedAt: types.NowDateTime()},
		}

		result, err := p.BuildResult(data)
		require.Nil(t, err)
		require.Len(t, result.Items, 2)

		for i, raw := range result.Items {
			var d Data
			require.Nil(t, json.Unmarshal(raw, &d))
			assert.Equal(t, data[i].ID, d.ID)
			assert.Equal(t, data[i].CreatedAt.String(), d.CreatedAt.String())
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=-20" + "&" + PerPageQueryParam + "=bob")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(24)

		result, err := p.BuildResult("data")
		require.EqualError(t, err, "input is not a slice")
		assert.Nil(t, result)
	})

	t.Run("error marshalling", func(t *testing.T) {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().SetRequestURI("/dummy?" + PageQueryParam + "=-20" + "&" + PerPageQueryParam + "=bob")
		defer app.ReleaseCtx(c)

		p := New(c)
		p.SetCount(24)

		// Invalid data
		badData := []struct {
			UnsupportedField chan int `json:"unsupportedField"`
		}{
			{UnsupportedField: make(chan int)},
		}

		result, err := p.BuildResult(badData)
		require.EqualError(t, err, "json: unsupported type: chan int")
		assert.Nil(t, result)
	})
}
