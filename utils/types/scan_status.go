package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type ScanStatus struct {
	s ScanStatusType
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type ScanStatusType string

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const (
	ScanStatusWaiting    ScanStatusType = "waiting"
	ScanStatusProcessing ScanStatusType = "processing"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ParseScanStatus creates a ScanStatus type with the status of
// waiting
func NewScanStatus(s ScanStatusType) ScanStatus {
	switch s {
	case ScanStatusWaiting:
		return ScanStatus{s: ScanStatusWaiting}
	case ScanStatusProcessing:
		return ScanStatus{s: ScanStatusProcessing}
	}

	return ScanStatus{s: ScanStatusWaiting}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetWaiting updates the state to waiting
func (ss *ScanStatus) SetWaiting() {
	ss.s = ScanStatusWaiting
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetProcessing updates the state to processing
func (ss *ScanStatus) SetProcessing() {
	ss.s = ScanStatusProcessing
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsWaiting returns true is the status is waiting
func (ss ScanStatus) IsWaiting() bool {
	return ss.s == ScanStatusWaiting
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// String implements the `Stringer` interface
func (ss ScanStatus) String() string {
	return fmt.Sprint(ss.s)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// MarshalJSON implements the `json.Marshaler` interface
func (ss ScanStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ss.s + `"`), nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UnmarshalJSON implements the `json.Unmarshaler` interface
func (ss *ScanStatus) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	return ss.Scan(raw)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Value implements the `driver.Valuer` interface
func (ss ScanStatus) Value() (driver.Value, error) {
	return ss.s, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan implements `sql.Scanner` interface
func (ss *ScanStatus) Scan(value any) error {

	vv := cast.ToString(value)

	switch vv {
	case string(ScanStatusWaiting):
		ss.s = ScanStatusWaiting
	case string(ScanStatusProcessing):
		ss.s = ScanStatusProcessing
	default:
		// Default to waiting
		ss.s = ScanStatusWaiting
	}

	return nil
}
