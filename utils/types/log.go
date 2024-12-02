package types

import "log/slog"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type LogType int

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	LogTypeRequest LogType = iota
	LogTypeCron
	LogTypeCourseScan
	LogTypeFileSystem
	LogTypeDB
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AllLogTypes returns a slice of all log types in string form
func AllLogTypes() []string {
	return []string{
		LogTypeRequest.String(),
		LogTypeCron.String(),
		LogTypeCourseScan.String(),
		LogTypeFileSystem.String(),
		LogTypeDB.String(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String returns the string representation of the LogType
func (lt LogType) String() string {
	names := [...]string{"request", "cron", "course scan", "file system", "db"}

	if int(lt) < 0 || int(lt) >= len(names) {
		return "unknown"
	}

	return names[lt]
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// LogValue implements the slog.Value interface
func (lt LogType) LogValue() slog.Value {
	return slog.StringValue(lt.String())
}
