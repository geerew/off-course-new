package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_NewAsset(t *testing.T) {
	// Valid
	tests := []struct {
		ext      string
		expected AssetType
	}{
		// Video
		{"avi", AssetVideo},
		{"mkv", AssetVideo},
		{"flac", AssetVideo},
		{"mp4", AssetVideo},
		{"m4a", AssetVideo},
		{"mp3", AssetVideo},
		{"ogv", AssetVideo},
		{"ogm", AssetVideo},
		{"ogg", AssetVideo},
		{"oga", AssetVideo},
		{"opus", AssetVideo},
		{"webm", AssetVideo},
		{"wav", AssetVideo},
		// document
		{"html", AssetHTML},
		{"htm", AssetHTML},
		{"pdf", AssetPDF},
	}

	for _, tt := range tests {
		a := NewAsset(tt.ext)
		assert.Equal(t, tt.expected, a.s)
	}

	// Invalid
	a := NewAsset("test")
	assert.Nil(t, a)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Set(t *testing.T) {
	a := NewAsset("html")
	assert.Equal(t, AssetHTML, a.s)

	// Set to video
	a.SetVideo()
	assert.Equal(t, AssetVideo, a.s)

	// Set to HTML
	a.SetHTML()
	assert.Equal(t, AssetHTML, a.s)

	// Set to PDF
	a.SetPDF()
	assert.Equal(t, AssetPDF, a.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Is(t *testing.T) {
	// Is video
	a := NewAsset("mp4")
	assert.True(t, a.IsVideo())

	// Is HTML
	a = NewAsset("html")
	assert.True(t, a.IsHTML())

	// Is PDF
	a = NewAsset("pdf")
	assert.True(t, a.IsPDF())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_String(t *testing.T) {
	a := NewAsset("mp4")
	assert.Equal(t, "video", a.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_MarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mp4", `"video"`},
		{"html", `"html"`},
		{"pdf", `"pdf"`},
	}

	for _, tt := range tests {
		a := NewAsset(tt.input)
		require.NotNil(t, a)

		res, err := a.MarshalJSON()
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, string(res))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected AssetType
		err      string
	}{
		// Errors
		{"", "", "unexpected end of JSON input"},
		{"xxx", "", "invalid character 'x' looking for beginning of value"},
		// Invalid asset types
		{`""`, "", "invalid asset type"},
		{`"bob"`, "", "invalid asset type"},
		// Success
		{`"video"`, AssetVideo, ""},
		{`"html"`, AssetHTML, ""},
		{`"pdf"`, AssetPDF, ""},
	}

	for _, tt := range tests {
		a := Asset{}
		err := a.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			assert.Equal(t, tt.expected, a.s)
		} else {
			assert.EqualError(t, err, tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Value(t *testing.T) {
	tests := []struct {
		input    string
		expected AssetType
	}{
		{"mp4", AssetVideo},
		{"html", AssetHTML},
		{"pdf", AssetPDF},
	}

	for _, tt := range tests {
		a := NewAsset(tt.input)
		require.NotNil(t, a)

		res, err := a.Value()
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, res)
	}

	// Nil
	a := Asset{}
	res, err := a.Value()
	assert.Nil(t, err)
	assert.Nil(t, res)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestAsset_Scan(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := []struct {
			value    any
			expected string
		}{
			{"video", "video"},
			{"html", "html"},
			{"pdf", "pdf"},
		}

		for _, tt := range tests {
			a := Asset{}

			err := a.Scan(tt.value)
			assert.Nil(t, err)
			assert.Contains(t, a.s, tt.expected)
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := []struct {
			value any
		}{
			{nil},
			{""},
			{"invalid"},
		}

		for _, tt := range tests {
			a := Asset{}

			err := a.Scan(tt.value)
			assert.EqualError(t, err, "invalid asset type")
		}
	})
}
