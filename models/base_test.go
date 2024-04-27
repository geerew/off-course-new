package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestRefreshId(t *testing.T) {
	model := BaseModel{}
	model.RefreshId()

	require.NotEmpty(t, model.ID)

	currentId := model.ID

	model.RefreshId()

	require.NotEmpty(t, model.ID)
	require.NotEqual(t, currentId, model.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestSetId(t *testing.T) {
	model := BaseModel{}
	model.SetId("testId")
	require.Equal(t, "testId", model.ID)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestRefreshCreatedAt(t *testing.T) {
	model := BaseModel{}
	model.RefreshCreatedAt()

	require.False(t, model.CreatedAt.IsZero())

	currentCreatedAt := model.CreatedAt
	time.Sleep(1 * time.Millisecond)

	model.RefreshCreatedAt()
	require.False(t, model.CreatedAt.IsZero())
	require.NotEqual(t, currentCreatedAt, model.CreatedAt)

}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestRefreshUpdatedAt(t *testing.T) {
	model := BaseModel{}
	model.RefreshUpdatedAt()

	require.False(t, model.UpdatedAt.IsZero())

	currentUpdatedAt := model.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	model.RefreshUpdatedAt()
	require.False(t, model.UpdatedAt.IsZero())
	require.NotEqual(t, currentUpdatedAt, model.UpdatedAt)
}
