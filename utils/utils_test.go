package utils

import (
	"reflect"
	"runtime"
	"testing"

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
		require.Equal(t, tt.expected, TrimQuotes(tt.in))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DecodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res, err := DecodeString("")
		require.NoError(t, err)
		require.Equal(t, "", res)
	})

	t.Run("decode error", func(t *testing.T) {
		res, err := DecodeString("`")
		require.EqualError(t, err, "failed to decode path")
		require.Empty(t, res)
	})

	t.Run("unescape error", func(t *testing.T) {
		res, err := DecodeString("dGVzdCUyMDElMiUyNiUyMHRlc3QlMjAy")
		require.EqualError(t, err, "failed to unescape path")
		require.Empty(t, res)
	})

	t.Run("success", func(t *testing.T) {
		res, err := DecodeString("JTJGdGVzdCUyRmRhdGE=")
		require.NoError(t, err)
		require.Equal(t, "/test/data", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_EncodeString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res := EncodeString("")
		require.Equal(t, "", res)
	})

	t.Run("success", func(t *testing.T) {
		res := EncodeString("/test/data")
		require.Equal(t, "JTJGdGVzdCUyRmRhdGE=", res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CheckTruth(t *testing.T) {
	checkTruthTests := []struct {
		v   string
		out bool
	}{
		{"123", true},
		{"true", true},
		{"", false},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"\u0046alse", false},
	}

	for _, test := range checkTruthTests {
		t.Run(test.v, func(t *testing.T) {
			assert.Equal(t, test.out, CheckTruth(test.v))
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_SliceIntersection(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res := SliceIntersection([]string{}, []string{})
		require.Empty(t, res)
	})

	t.Run("no intersection", func(t *testing.T) {
		res := SliceIntersection([]string{"1", "2", "3"}, []string{"4", "5", "6"})
		require.Empty(t, res)
	})

	t.Run("intersection", func(t *testing.T) {
		res := SliceIntersection([]string{"1", "2", "3"}, []string{"2", "3", "4"})
		require.ElementsMatch(t, []string{"2", "3"}, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DiffSliceOfStructsByKey(t *testing.T) {
	// Struct for testing
	type testStruct struct {
		ID    int
		Title string
	}

	t.Run("not a struct empty", func(t *testing.T) {
		leftDiff, rightDiff, err := DiffSliceOfStructsByKey([]string{"left"}, []string{"right"}, "")

		require.Error(t, err)
		require.EqualError(t, err, "invalid struct or key")
		require.Nil(t, leftDiff)
		require.Nil(t, rightDiff)
	})

	t.Run("invalid key", func(t *testing.T) {
		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(
			[]testStruct{{ID: 0, Title: "Test"}},
			[]testStruct{{ID: 0, Title: "Test"}},
			"Name")

		require.Error(t, err)
		require.EqualError(t, err, "invalid struct or key")
		require.Nil(t, leftDiff)
		require.Nil(t, rightDiff)
	})

	t.Run("both empty", func(t *testing.T) {
		leftDiff, rightDiff, err := DiffSliceOfStructsByKey[testStruct](nil, nil, "")
		require.NoError(t, err)
		require.Nil(t, leftDiff)
		require.Nil(t, rightDiff)
	})

	t.Run("left empty", func(t *testing.T) {
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			right = append(right, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(nil, right, "ID")
		require.NoError(t, err)
		require.Empty(t, leftDiff)
		require.Len(t, rightDiff, 5)
	})

	t.Run("right empty", func(t *testing.T) {
		left := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(left, nil, "ID")
		require.NoError(t, err)
		require.Len(t, leftDiff, 5)
		require.Empty(t, rightDiff)
	})

	t.Run("same", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i, Title: "Test"})
		}

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(left, right, "ID")
		require.NoError(t, err)
		require.Empty(t, leftDiff)
		require.Empty(t, rightDiff)
	})

	t.Run("completely different", func(t *testing.T) {
		left := []*testStruct{}
		right := []*testStruct{}

		for i := 0; i < 5; i++ {
			left = append(left, &testStruct{ID: i, Title: "Test"})
			right = append(right, &testStruct{ID: i + 5, Title: "Test"})
		}

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(left, right, "ID")
		require.NoError(t, err)
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

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(left, right, "ID")
		require.NoError(t, err)
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

		leftDiff, rightDiff, err := DiffSliceOfStructsByKey(left, right, "ID")
		require.NoError(t, err)
		require.Len(t, leftDiff, 1)
		require.Zero(t, len(rightDiff))

		// Give right 1 extra (plus the new left one)
		right = append(right, left[len(left)-1], &testStruct{ID: 6, Title: "Test"})
		leftDiff, rightDiff, err = DiffSliceOfStructsByKey(left, right, "ID")
		require.NoError(t, err)
		require.Zero(t, len(leftDiff))
		require.Len(t, rightDiff, 1)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CompareStructs(t *testing.T) {
	// Struct for testing
	type testStruct struct {
		ID        int
		Title     string
		Path      string
		CreatedAt string
		UpdatedAt string
	}

	t.Run("not a struct", func(t *testing.T) {
		require.False(t, CompareStructs("test", "test", nil))
	})

	t.Run("different types", func(t *testing.T) {
		require.False(t, CompareStructs(testStruct{}, &testStruct{}, nil))
	})

	t.Run("empty", func(t *testing.T) {
		require.True(t, CompareStructs(testStruct{}, testStruct{}, nil))
	})

	t.Run("same", func(t *testing.T) {
		require.True(t, CompareStructs(testStruct{ID: 1, Title: "Test"}, testStruct{ID: 1, Title: "Test"}, nil))
	})

	t.Run("different", func(t *testing.T) {
		require.False(t, CompareStructs(testStruct{ID: 1, Title: "Test"}, testStruct{ID: 1, Title: "Test 2"}, nil))
	})

	t.Run("ignore", func(t *testing.T) {
		require.True(t, CompareStructs(
			testStruct{ID: 1, Title: "Test", Path: "/test", CreatedAt: "2021-01-01", UpdatedAt: "2021-01-01"},
			testStruct{ID: 1, Title: "Test", Path: "/test", CreatedAt: "2021-01-02", UpdatedAt: "2021-01-02"},
			[]string{"CreatedAt", "UpdatedAt"},
		))
	})

	t.Run("nested struct", func(t *testing.T) {
		type nestedStruct struct {
			ID int
		}

		type testStruct struct {
			ID     int
			Title  string
			Nested nestedStruct
		}

		require.True(t, CompareStructs(
			testStruct{ID: 1, Title: "Test", Nested: nestedStruct{ID: 1}},
			testStruct{ID: 1, Title: "Test", Nested: nestedStruct{ID: 1}},
			nil,
		))

		require.False(t, CompareStructs(
			testStruct{ID: 1, Title: "Test", Nested: nestedStruct{ID: 1}},
			testStruct{ID: 1, Title: "Test", Nested: nestedStruct{ID: 2}},
			nil,
		))
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
		require.True(t, IsStructWithKey(test, "ID"))
	})

	t.Run("valid pointer struct", func(t *testing.T) {
		test := &testStruct{ID: 1, Title: "Test"}
		require.True(t, IsStructWithKey(test, "ID"))
	})

	t.Run("string", func(t *testing.T) {
		require.False(t, IsStructWithKey("test", "ID"))
	})

	t.Run("invalid key", func(t *testing.T) {
		test := testStruct{ID: 1, Title: "Test"}
		require.False(t, IsStructWithKey(test, "Name"))
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
		require.Equal(t, test.output, ValueToString(reflect.ValueOf(test.input)))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NormalizeWindowsDrive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Test cases for Windows drive paths
		{"C:", "C:\\"},
		{"C:\\", "C:\\"},
		{"C:folder", "C:\\folder"},
		{"C:\\folder", "C:\\folder"},
	}

	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific tests on non-Windows systems")
	}

	for _, test := range tests {
		got := NormalizeWindowsDrive(test.input)
		require.Equal(t, test.expected, got)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_EscapeBackslashes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`C:\Users\example\path`, `C:\\Users\\example\\path`},
		{`C:\Users\\example\\path`, `C:\\Users\\example\\path`},
		{`C:\\Users\\example\\path`, `C:\\Users\\example\\path`},
		{`\\`, `\\`},
		{`\`, `\\`},
		{`C:`, `C:`},
		{``, ``},
		{`C:\Users\ex\ample\pa\th`, `C:\\Users\\ex\\ample\\pa\\th`},
	}

	for _, test := range tests {
		result := EscapeBackslashes(test.input)
		require.Equal(t, test.expected, result)
	}
}

// -------------------------------------------------------

func Test_SnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"camelCase", "camel_case"},
		{"PascalCase", "pascal_case"},
		{"snake_case", "snake_case"},
		{"Already_Snake_Case", "already_snake_case"},
		{"UpperCase", "upper_case"},
		{"lowercase", "lowercase"},
		{"SimpleTest", "simple_test"},
		{"TestID", "test_id"},
		{"AnotherTestCase", "another_test_case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := SnakeCase(tt.input)
			if actual != tt.expected {
				t.Errorf("snakeCase(%q) = %q; want %q", tt.input, actual, tt.expected)
			}
		})
	}
}
