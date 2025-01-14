package api

import "github.com/geerew/off-course/utils/types"

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileSystemResponse struct {
	Count       int                 `json:"count"`
	Directories []*fileInfoResponse `json:"directories"`
	Files       []*fileInfoResponse `json:"files"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileInfoResponse struct {
	Title          string                   `json:"title"`
	Path           string                   `json:"path"`
	Classification types.PathClassification `json:"classification"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseProgressResponse struct {
	Started           bool           `json:"started"`
	StartedAt         types.DateTime `json:"startedAt"`
	Percent           int            `json:"percent"`
	CompletedAt       types.DateTime `json:"completedAt"`
	ProgressUpdatedAt types.DateTime `json:"progressUpdatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	HasCard   bool           `json:"hasCard"`
	Available bool           `json:"available"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Scan status
	ScanStatus string `json:"scanStatus"`

	// Progress
	Progress courseProgressResponse `json:"progress"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTagResponse struct {
	ID       string `json:"id"`
	Tag      string `json:"tag,omitempty"`
	CourseID string `json:"courseId,omitempty"`
	Title    string `json:"title,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetProgressRequest struct {
	VideoPos  int  `json:"videoPos"`
	Completed bool `json:"completed"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetProgressResponse struct {
	VideoPos    int            `json:"videoPos"`
	Completed   bool           `json:"completed"`
	CompletedAt types.DateTime `json:"completedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachmentResponse struct {
	ID        string         `json:"id"`
	AssetId   string         `json:"assetId"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetResponse struct {
	ID        string         `json:"id"`
	CourseID  string         `json:"courseId"`
	Title     string         `json:"title"`
	Prefix    int            `json:"prefix"`
	Chapter   string         `json:"chapter"`
	Path      string         `json:"path"`
	Type      types.Asset    `json:"assetType"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Relations
	Progress    *assetProgressResponse `json:"progress"`
	Attachments []*attachmentResponse  `json:"attachments,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"courseId"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt types.DateTime   `json:"createdAt"`
	UpdatedAt types.DateTime   `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagRequest struct {
	Tag string `json:"tag"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagResponse struct {
	ID          string               `json:"id"`
	Tag         string               `json:"tag"`
	CourseCount int                  `json:"courseCount"`
	Courses     []*courseTagResponse `json:"courses,omitempty"`
	CreatedAt   types.DateTime       `json:"createdAt"`
	UpdatedAt   types.DateTime       `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type UserRequest struct {
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type UserResponse struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	DisplayName string         `json:"displayName"`
	Role        types.UserRole `json:"role"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TokenResponse struct {
	Token string `json:"token"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logResponse struct {
	ID        string         `json:"id"`
	Level     int            `json:"level"`
	Message   string         `json:"message"`
	Data      types.JsonMap  `json:"data"`
	CreatedAt types.DateTime `json:"createdAt"`
}
