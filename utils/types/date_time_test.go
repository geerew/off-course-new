package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NowDateTime(t *testing.T) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05") // without ms part for test consistency
	dt := NowDateTime()

	require.Contains(t, dt.String(), now)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ParseDateTime(t *testing.T) {
	nowTime := time.Now().UTC()
	nowDateTime, _ := ParseDateTime(nowTime)
	nowStr := nowTime.Format(DefaultDateLayout)

	scenarios := []struct {
		value    any
		expected string
	}{
		{nil, ""},
		{"", ""},
		{"invalid", ""},
		{nowDateTime, nowStr},
		{nowTime, nowStr},
		{1641024040, "2022-01-01 08:00:40.000Z"},
		{int32(1641024040), "2022-01-01 08:00:40.000Z"},
		{int64(1641024040), "2022-01-01 08:00:40.000Z"},
		{uint(1641024040), "2022-01-01 08:00:40.000Z"},
		{uint64(1641024040), "2022-01-01 08:00:40.000Z"},
		{uint32(1641024040), "2022-01-01 08:00:40.000Z"},
		{"2022-01-01 11:23:45.678", "2022-01-01 11:23:45.678Z"},
	}

	for i, s := range scenarios {
		dt, err := ParseDateTime(s.value)

		require.Nil(t, err, "(%d) Failed to parse %v: %v", i, s.value, err)
		require.Equal(t, s.expected, dt.String(), "(%d) Expected %q, got %q", i, s.expected, dt.String())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_DateTimeTime(t *testing.T) {
	str := "2022-01-01 11:23:45.678Z"

	expected, err := time.Parse(DefaultDateLayout, str)
	require.NoError(t, err)

	dt, err := ParseDateTime(str)
	require.NoError(t, err)

	result := dt.Time()
	require.Equal(t, expected, result)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_IsZero(t *testing.T) {
	dt0 := DateTime{}
	require.True(t, dt0.IsZero())

	dt1 := NowDateTime()
	require.False(t, dt1.IsZero())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_Equal(t *testing.T) {
	scenarios := []struct {
		dt1      DateTime
		dt2      DateTime
		expected bool
	}{
		{DateTime{}, DateTime{}, true},        // Both zero values
		{NowDateTime(), NowDateTime(), false}, // Different current times
		{NowDateTime(), NowDateTime(), false}, // Another set of different times
		{
			dt1:      DateTime{t: time.Date(2022, 1, 1, 11, 23, 45, 678000000, time.UTC)},
			dt2:      DateTime{t: time.Date(2022, 1, 1, 11, 23, 45, 678000000, time.UTC)},
			expected: true, // Matching times
		},
		{
			dt1:      DateTime{t: time.Date(2022, 1, 1, 11, 23, 45, 0, time.UTC)},
			dt2:      DateTime{t: time.Date(2022, 1, 1, 11, 23, 46, 0, time.UTC)},
			expected: false, // Different times
		},
	}

	for i, s := range scenarios {
		require.True(t, s.dt1.Equal(s.dt1), "(%d) Expected %v.Equal(%v) to be true", i, s.dt1, s.dt1)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_String(t *testing.T) {
	dt0 := DateTime{}
	require.Empty(t, dt0.String())

	expected := "2022-01-01 11:23:45.678Z"
	dt1, _ := ParseDateTime(expected)
	require.Equal(t, expected, dt1.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_MarshalJSON(t *testing.T) {
	scenarios := []struct {
		date     string
		expected string
	}{
		{"", `""`},
		{"2022-01-01 11:23:45.678", `"2022-01-01 11:23:45.678Z"`},
	}

	for i, s := range scenarios {
		dt, err := ParseDateTime(s.date)
		require.Nil(t, err, "(%d) %v", i, err)

		result, err := dt.MarshalJSON()
		require.Nil(t, err, "(%d) %v", i, err)
		require.Equal(t, s.expected, string(result), "(%d) Expected %q, got %q", i, s.expected, string(result))

	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_UnmarshalJSON(t *testing.T) {
	scenarios := []struct {
		date     string
		expected string
	}{
		{"", ""},
		{"invalid_json", ""},
		{"'123'", ""},
		{"2022-01-01 11:23:45.678", ""},
		{`"2022-01-01 11:23:45.678"`, "2022-01-01 11:23:45.678Z"},
	}

	for i, s := range scenarios {
		dt := DateTime{}
		dt.UnmarshalJSON([]byte(s.date))
		require.Equal(t, s.expected, dt.String(), "(%d) Expected %q, got %q", i, s.expected, dt.String())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_Value(t *testing.T) {
	scenarios := []struct {
		value    any
		expected string
	}{
		{"", ""},
		{"invalid", ""},
		{1641024040, "2022-01-01 08:00:40.000Z"},
		{"2022-01-01 11:23:45.678", "2022-01-01 11:23:45.678Z"},
		{NowDateTime(), NowDateTime().String()},
	}

	for i, s := range scenarios {
		dt, _ := ParseDateTime(s.value)
		result, err := dt.Value()
		require.Nil(t, err, "(%d) %v", i, err)
		require.Equal(t, s.expected, result, "(%d) Expected %q, got %q", i, s.expected, result)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTime_Scan(t *testing.T) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05") // without ms part for test consistency

	scenarios := []struct {
		value    any
		expected string
	}{
		{nil, ""},
		{"", ""},
		{"invalid", ""},
		{NowDateTime(), now},
		{time.Now(), now},
		{1.0, ""},
		{1641024040, "2022-01-01 08:00:40.000Z"},
		{"2022-01-01 11:23:45.678", "2022-01-01 11:23:45.678Z"},
	}

	for i, s := range scenarios {
		dt := DateTime{}

		err := dt.Scan(s.value)
		require.Nil(t, err, "(%d) %v", i, err)
		require.Contains(t, dt.String(), s.expected, "(%d) Expected %q, got %q", i, s.expected, dt.String())
	}
}
