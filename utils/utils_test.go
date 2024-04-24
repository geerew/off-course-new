package utils

import (
	"reflect"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rzajac/zltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_TrimQuotes(t *testing.T) {
	tests := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"1", "1"},
		{`""`, ""},
		{"123", "123"},
		{`"123"`, "123"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, TrimQuotes(tt.in))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DecodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res, err := DecodeString("")
		assert.Nil(t, err)
		assert.Equal(t, "", res)
	})

	t.Run("decode error", func(t *testing.T) {
		res, err := DecodeString("`")
		assert.EqualError(t, err, "failed to decode path")
		assert.Empty(t, res)
	})

	t.Run("unescape error", func(t *testing.T) {
		res, err := DecodeString("dGVzdCUyMDElMiUyNiUyMHRlc3QlMjAy")
		assert.EqualError(t, err, "failed to unescape path")
		assert.Empty(t, res)
	})

	t.Run("success", func(t *testing.T) {
		res, err := DecodeString("JTJGdGVzdCUyRmRhdGE=")
		assert.Nil(t, err)
		assert.Equal(t, "/test/data", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_EncodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res := EncodeString("")
		assert.Equal(t, "", res)
	})

	t.Run("success", func(t *testing.T) {
		res := EncodeString("/test/data")
		assert.Equal(t, "JTJGdGVzdCUyRmRhdGE=", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DiffStructs(t *testing.T) {
	// Set test logger
	loggerHook := zltest.New(t)
	log.Logger = zerolog.New(loggerHook)

	// Struct for testing
	type testStruct struct {
		ID    int
		Title string
	}

	t.Run("not a struct empty", func(t *testing.T) {
		leftDiff, rightDiff := DiffStructs([]string{"left"}, []string{"right"}, "")
		assert.Nil(t, leftDiff)
		assert.Nil(t, rightDiff)

		loggerHook.LastEntry().ExpMsg("invalid struct or invalid key")
		loggerHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
		loggerHook.Reset()
	})

	t.Run("invalid key", func(t *testing.T) {
		leftDiff, rightDiff := DiffStructs(
			[]testStruct{{ID: 0, Title: "Test"}},
			[]testStruct{{ID: 0, Title: "Test"}},
			"Name")

		assert.Nil(t, leftDiff)
		assert.Nil(t, rightDiff)

		loggerHook.LastEntry().ExpMsg("invalid struct or invalid key")
		loggerHook.LastEntry().ExpLevel(zerolog.ErrorLevel)
		loggerHook.Reset()
	})

	t.Run("both empty", func(t *testing.T) {
		leftDiff, rightDiff := DiffStructs[testStruct](nil, nil, "")
		assert.Nil(t, leftDiff)
		assert.Nil(t, rightDiff)
	})

	t.Run("left empty", func(t *testing.T) {
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			right = append(right, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff := DiffStructs(nil, right, "ID")
		assert.Empty(t, leftDiff)
		require.Len(t, rightDiff, 5)
	})

	t.Run("right empty", func(t *testing.T) {
		left := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff := DiffStructs(left, nil, "ID")
		require.Len(t, leftDiff, 5)
		assert.Empty(t, rightDiff)
	})

	t.Run("same", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff := DiffStructs(left, right, "ID")
		assert.Empty(t, leftDiff)
		assert.Empty(t, rightDiff)
	})

	t.Run("completely different", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i + 5, Title: "Test"})
		}

		leftDiff, rightDiff := DiffStructs(left, right, "ID")
		require.Len(t, leftDiff, 5)
		require.Len(t, rightDiff, 5)
	})

	t.Run("mixture", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i + 5, Title: "Test"})
		}

		// Give left 2 from right. This means right now only has 3 that left does not have
		left = append(left, right[0], right[3])

		// Give right 1 from left. This means left now only has 4 that right does not have
		right = append(right, left[0])

		leftDiff, rightDiff := DiffStructs(left, right, "ID")
		require.Len(t, leftDiff, 4)
		require.Len(t, rightDiff, 3)
	})

	t.Run("1 extra", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i, Title: "Test"})
		}

		// Give left 1 extra
		left = append(left, &testStruct{ID: 5, Title: "Test"})

		leftDiff, rightDiff := DiffStructs(left, right, "ID")
		require.Len(t, leftDiff, 1)
		require.Len(t, rightDiff, 0)

		// Give right 1 extra (plus the new left one)
		right = append(right, left[len(left)-1], &testStruct{ID: 6, Title: "Test"})
		leftDiff, rightDiff = DiffStructs(left, right, "ID")
		require.Len(t, leftDiff, 0)
		require.Len(t, rightDiff, 1)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_IsStructWithKey(t *testing.T) {
	type testStruct struct {
		ID    int
		Title string
	}

	t.Run("valid struct", func(t *testing.T) {
		test := testStruct{ID: 1, Title: "Test"}
		assert.True(t, IsStructWithKey(test, "ID"))
	})

	t.Run("valid pointer struct", func(t *testing.T) {
		test := &testStruct{ID: 1, Title: "Test"}
		assert.True(t, IsStructWithKey(test, "ID"))
	})

	t.Run("string", func(t *testing.T) {
		assert.False(t, IsStructWithKey("test", "ID"))
	})

	t.Run("invalid key", func(t *testing.T) {
		test := testStruct{ID: 1, Title: "Test"}
		assert.False(t, IsStructWithKey(test, "Name"))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ValueToString(t *testing.T) {
	// Test integers
	tests := []struct {
		input  interface{}
		output string
	}{
		{int(42), "42"},
		{int8(42), "42"},
		{int16(42), "42"},
		{int32(42), "42"},
		{int64(42), "42"},
		{uint(42), "42"},
		{uint8(42), "42"},
		{uint16(42), "42"},
		{uint32(42), "42"},
		{uint64(42), "42"},
		{float32(42.5), "42.5"},
		{float64(42.5), "42.5"},
		{"Hello", "Hello"},
		{true, "true"},
		{false, "false"},
		{struct{}{}, ""}, // Unsupported type
	}

	for _, test := range tests {
		assert.Equal(t, test.output, ValueToString(reflect.ValueOf(test.input)))
	}
}
