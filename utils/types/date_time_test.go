package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestNowDateTime(t *testing.T) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	dt := NowDateTime()
	assert.Contains(t, dt.String(), now)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestParseDateTime(t *testing.T) {

	nowTime := time.Now().UTC()
	nowDateTime, _ := ParseDateTime(nowTime)
	nowStr := nowTime.Format(DefaultDateLayout)

	tests := []struct {
		value    any
		expected string
	}{
		{nil, ""},
		{"", ""},
		{"invalid", ""},
		{nowDateTime, nowStr},
		{nowTime, nowStr},
		{1677592032, "2023-02-28 13:47:12.000Z"},
		{"2023-02-28 11:23:45.678", "2023-02-28 11:23:45.678Z"},
	}

	for _, tt := range tests {
		dt, err := ParseDateTime(tt.value)
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, dt.String())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestTime(t *testing.T) {
	str := "2023-02-28 11:23:45.678Z"

	expected, err := time.Parse(DefaultDateLayout, str)
	assert.Nil(t, err)

	dt, err := ParseDateTime(str)
	assert.Nil(t, err)

	res := dt.Time()
	assert.Equal(t, expected, res)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeIsZero(t *testing.T) {
	dt0 := DateTime{}
	assert.True(t, dt0.IsZero())

	dt1 := NowDateTime()
	assert.False(t, dt1.IsZero())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeString(t *testing.T) {
	dt0 := DateTime{}
	assert.Empty(t, dt0.String())

	expected := "2023-02-28 11:23:45.678Z"
	dt1, err := ParseDateTime(expected)
	assert.Nil(t, err)
	assert.Equal(t, expected, dt1.String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeMarshalJSON(t *testing.T) {
	tests := []struct {
		date     string
		expected string
	}{
		{"", `""`},
		{"2023-02-28 11:23:45.678", `"2023-02-28 11:23:45.678Z"`},
	}

	for _, tt := range tests {
		dt, err := ParseDateTime(tt.date)
		assert.Nil(t, err)

		res, err := dt.MarshalJSON()
		assert.Nil(t, err)
		assert.Equal(t, tt.expected, string(res))
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		date     string
		expected string
	}{
		{"", ""},
		{"invalid_json", ""},
		{"'123'", ""},
		{"2023-02-28 11:23:45.678", ""},
		{`"2023-02-28 11:23:45.678"`, "2023-02-28 11:23:45.678Z"},
	}

	for _, tt := range tests {
		dt := DateTime{}
		dt.UnmarshalJSON([]byte(tt.date))
		assert.Equal(t, tt.expected, dt.String())
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeValue(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		{"", ""},
		{"invalid", ""},
		{1641024040, "2022-01-01 08:00:40.000Z"},
		{"2023-02-28 11:23:45.678", "2023-02-28 11:23:45.678Z"},
		{NowDateTime(), NowDateTime().String()},
	}

	for _, tt := range tests {
		dt, _ := ParseDateTime(tt.value)
		res, err := dt.Value()

		assert.Nil(t, err)
		assert.Equal(t, tt.expected, res)
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestDateTimeScan(t *testing.T) {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	tests := []struct {
		value    any
		expected string
	}{
		{nil, ""},
		{"", ""},
		{"invalid", ""},
		{NowDateTime(), now},
		{time.Now(), now},
		{1.0, ""},
		{1677592032, "2023-02-28 13:47:12.000Z"},
		{"2023-02-28 11:23:45.678", "2023-02-28 11:23:45.678Z"},
	}

	for _, tt := range tests {
		dt := DateTime{}

		err := dt.Scan(tt.value)
		assert.Nil(t, err)
		assert.Contains(t, dt.String(), tt.expected)
	}
}
