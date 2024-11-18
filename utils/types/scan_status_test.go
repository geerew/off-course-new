package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_NewScanStatusWaiting(t *testing.T) {
	require.Equal(t, ScanStatusWaiting, NewScanStatusWaiting().s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_NewScanStatusProcessing(t *testing.T) {
	require.Equal(t, ScanStatusProcessing, NewScanStatusProcessing().s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_SetWaiting(t *testing.T) {
	s := NewScanStatusProcessing()
	s.SetWaiting()
	require.Equal(t, ScanStatusWaiting, s.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_SetProcessing(t *testing.T) {
	s := NewScanStatusWaiting()
	s.SetProcessing()
	require.Equal(t, ScanStatusProcessing, s.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_IsWaiting(t *testing.T) {
	require.True(t, NewScanStatusWaiting().IsWaiting())
	require.False(t, NewScanStatusProcessing().IsWaiting())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_MarshalJSON(t *testing.T) {
	waiting := NewScanStatusWaiting()
	res, err := waiting.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, `"waiting"`, string(res))

	processing := NewScanStatusProcessing()
	res, err = processing.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, `"processing"`, string(res))

	empty := ScanStatus{}
	res, err = empty.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, `""`, string(res))
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected ScanStatusType
		err      string
	}{
		// Errors
		{"", "", "unexpected end of JSON input"},
		{"xxx", "", "invalid character 'x' looking for beginning of value"},
		// Defaults
		{`""`, "", ""},
		{`"bob"`, "", ""},
		// Success
		{`"waiting"`, ScanStatusWaiting, ""},
		{`"processing"`, ScanStatusProcessing, ""},
	}

	for _, tt := range tests {
		ss := ScanStatus{}
		err := ss.UnmarshalJSON([]byte(tt.input))

		if tt.err == "" {
			require.Equal(t, tt.expected, ss.s)
		} else {
			require.EqualError(t, err, tt.err)
		}
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Value(t *testing.T) {
	waiting := NewScanStatusWaiting()
	res, err := waiting.Value()
	require.Nil(t, err)
	require.Equal(t, "waiting", res)

	processing := NewScanStatusProcessing()
	res, err = processing.Value()
	require.Nil(t, err)
	require.Equal(t, "processing", res)

	empty := ScanStatus{}
	res, err = empty.Value()
	require.Nil(t, err)
	require.Equal(t, "", res)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestScanStatus_Scan(t *testing.T) {
	tests := []struct {
		value    any
		expected string
	}{
		// defaults
		{nil, ""},
		{"", ""},
		{"invalid", ""},
		// Values
		{"waiting", "waiting"},
		{"processing", "processing"},
	}

	for _, tt := range tests {
		ss := ScanStatus{}

		err := ss.Scan(tt.value)
		require.Nil(t, err)
		require.Contains(t, ss.s, tt.expected)
	}
}
