package types

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_String(t *testing.T) {
	tests := []struct {
		logType  LogType
		expected string
	}{
		{LogTypeRequest, "request"},
		{LogTypeCron, "cron"},
		{LogTypeScanner, "scanner"},
		{LogTypeFileSystem, "file system"},
		{LogTypeDB, "db"},
		{LogType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.logType.String())
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_AllLogTypes(t *testing.T) {
	expected := []string{"request", "cron", "scanner", "file system", "db"}
	require.Equal(t, expected, AllLogTypes())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestLog_LogValue(t *testing.T) {
	tests := []struct {
		logType  LogType
		expected string
	}{
		{LogTypeRequest, "request"},
		{LogTypeCron, "cron"},
		{LogTypeScanner, "scanner"},
		{LogTypeFileSystem, "file system"},
		{LogTypeDB, "db"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			require.Equal(t, slog.StringValue(tt.expected), tt.logType.LogValue())
		})
	}
}
