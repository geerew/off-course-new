package types

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_String(t *testing.T) {
	assert.Equal(t, "admin", UserRoleAdmin.String())
	assert.Equal(t, "user", UserRoleUser.String())
	assert.Equal(t, "invalid", UserRole("invalid").String())
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected bool
	}{
		{UserRoleAdmin, true},
		{UserRoleUser, true},
		{UserRole("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.role.IsValid())
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_MarshalJSON(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected string
		hasError bool
	}{
		{UserRoleAdmin, `"admin"`, false},
		{UserRoleUser, `"user"`, false},
		{UserRole("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			data, err := json.Marshal(tt.role)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, string(data))
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		data     string
		expected UserRole
		hasError bool
	}{
		{`"admin"`, UserRoleAdmin, false},
		{`"user"`, UserRoleUser, false},
		{`"invalid"`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.data, func(t *testing.T) {
			var role UserRole
			err := json.Unmarshal([]byte(tt.data), &role)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, role)
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_Scan(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected UserRole
		hasError bool
	}{
		{"admin", UserRoleAdmin, false},
		{"user", UserRoleUser, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input.(string), func(t *testing.T) {
			var role UserRole
			err := role.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, role)
			}
		})
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestUserRole_Value(t *testing.T) {
	tests := []struct {
		role     UserRole
		expected driver.Value
		hasError bool
	}{
		{UserRoleAdmin, "admin", false},
		{UserRoleUser, "user", false},
		{UserRole("invalid"), nil, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			value, err := tt.role.Value()
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, value)
			}
		})
	}
}
