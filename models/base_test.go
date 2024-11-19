package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RefreshId(t *testing.T) {
	model := Base{}
	model.RefreshId()

	require.NotEmpty(t, model.ID)

	currentId := model.ID

	model.RefreshId()

	require.NotEmpty(t, model.ID)
	require.NotEqual(t, currentId, model.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_SetId(t *testing.T) {
	model := Base{}
	model.SetId("testId")
	require.Equal(t, "testId", model.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RefreshCreatedAt(t *testing.T) {
	model := Base{}
	model.RefreshCreatedAt()

	require.False(t, model.CreatedAt.IsZero())

	currentCreatedAt := model.CreatedAt
	time.Sleep(1 * time.Millisecond)

	model.RefreshCreatedAt()
	require.False(t, model.CreatedAt.IsZero())
	require.NotEqual(t, currentCreatedAt, model.CreatedAt)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_RefreshUpdatedAt(t *testing.T) {
	model := Base{}
	model.RefreshUpdatedAt()

	require.False(t, model.UpdatedAt.IsZero())

	currentUpdatedAt := model.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	model.RefreshUpdatedAt()
	require.False(t, model.UpdatedAt.IsZero())
	require.NotEqual(t, currentUpdatedAt, model.UpdatedAt)
}
