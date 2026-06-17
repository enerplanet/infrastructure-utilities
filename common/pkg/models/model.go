// Package models holds the platform schema shared by all SpatialHub apps; app-specific entities live in each app's backend.
package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"platform.local/common/pkg/contracts"
)

type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID      string     `gorm:"not null;size:255;index" json:"user_id"`
	UserEmail   string     `gorm:"not null;size:255;index" json:"user_email"`
	WorkspaceID *uint      `gorm:"index" json:"workspace_id"`
	Workspace   *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`

	Title       string  `gorm:"not null;size:255" json:"title"`
	Description *string `gorm:"type:text" json:"description,omitempty"`
	Status      string  `gorm:"not null;size:32;default:'draft';index" json:"status"`

	Coordinates datatypes.JSON `gorm:"type:jsonb" json:"coordinates,omitempty"`

	Region         *string `gorm:"size:255" json:"region,omitempty"`
	Country        *string `gorm:"size:255" json:"country,omitempty"`
	BufferDistance *int    `json:"buffer_distance,omitempty"`
	Resolution     *int    `json:"resolution,omitempty"`

	FromDate time.Time `gorm:"not null" json:"from_date"`
	ToDate   time.Time `gorm:"not null" json:"to_date"`

	Config  datatypes.JSON `gorm:"type:jsonb" json:"config,omitempty"`
	Results datatypes.JSON `gorm:"type:jsonb" json:"results,omitempty"`

	SessionID   *int64  `gorm:"index" json:"session_id,omitempty"`
	CallbackURL *string `gorm:"type:text" json:"callback_url,omitempty"`

	WebserviceID *uint `gorm:"index" json:"webservice_id,omitempty"`

	GroupID       *uint  `gorm:"index" json:"group_id,omitempty"`
	ParentModelID *uint  `gorm:"index" json:"parent_model_id,omitempty"`
	ParentModel   *Model `gorm:"foreignKey:ParentModelID" json:"parent_model,omitempty"`
	IsCopy        bool   `gorm:"default:false" json:"is_copy"`


	ChildModelIDs []uint `gorm:"-" json:"child_model_ids,omitempty"`


	ParentModelTitle *string `gorm:"-" json:"parent_model_title,omitempty"`

	IsActive bool `gorm:"default:false" json:"is_active"`

	CalculationStartedAt   *time.Time `json:"calculation_started_at,omitempty"`
	CalculationCompletedAt *time.Time `json:"calculation_completed_at,omitempty"`

	// Relationships
	Shares []ModelShare `gorm:"foreignKey:ModelID" json:"shares,omitempty"`

	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Model status constants, re-exported from the contracts package.
const (
	ModelStatusDraft      = contracts.StatusDraft
	ModelStatusQueue      = contracts.StatusQueue
	ModelStatusRunning    = contracts.StatusRunning
	ModelStatusProcessing = contracts.StatusProcessing
	ModelStatusCompleted  = contracts.StatusCompleted
	ModelStatusPublished  = contracts.StatusPublished
	ModelStatusFailed     = contracts.StatusFailed
	ModelStatusCancelled  = contracts.StatusCancelled
)

func (Model) TableName() string {
	return "models"
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if len(m.Coordinates) == 0 {
		m.Coordinates = datatypes.JSON([]byte("null"))
	}
	if len(m.Config) == 0 {
		m.Config = datatypes.JSON([]byte("null"))
	}
	if len(m.Results) == 0 {
		m.Results = datatypes.JSON([]byte("null"))
	}
	return nil
}

func (m *Model) IsOwner(userID string) bool {
	return m.UserID == userID
}

func (m *Model) CanBeEditedByUser(userID string) bool {
	return m.IsOwner(userID)
}

func (m *Model) IsCompleted() bool {
	return m.Status == ModelStatusCompleted || m.Status == ModelStatusPublished
}

func (m *Model) IsPending() bool {
	return m.Status == ModelStatusQueue
}

func (m *Model) IsRunning() bool {
	return m.Status == ModelStatusRunning
}

func (m *Model) IsDraft() bool {
	return m.Status == ModelStatusDraft
}

func (m *Model) IsProcessing() bool {
	return m.Status == ModelStatusProcessing
}

func (m *Model) HasFailed() bool {
	return m.Status == ModelStatusFailed
}

// IsShared returns true if the model has been shared with other users
func (m *Model) IsShared() bool {
	return len(m.Shares) > 0
}

type ModelShare struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ModelID    uint      `gorm:"not null;index" json:"model_id"`
	UserID     string    `gorm:"size:255;index" json:"user_id"`
	Email      string    `gorm:"not null;size:255;index" json:"email"`
	Permission string    `gorm:"not null;size:32;default:'view'" json:"permission"`
	SharedBy   string    `gorm:"not null;size:255" json:"shared_by"`
	SharedAt   time.Time `json:"shared_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Model *Model `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}

const (
	ModelSharePermissionView = "view"
	ModelSharePermissionEdit = "edit"
)

func (ModelShare) TableName() string {
	return "model_shares"
}

type ModelResult struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	ModelID uint   `gorm:"not null;index" json:"model_id"`
	UserID  string `gorm:"size:255;index" json:"user_id"`

	ZipPath       string `gorm:"type:text;not null" json:"zip_path"`
	ExtractedPath string `gorm:"type:text;not null" json:"extracted_path"`
	TifFilePath   string `gorm:"type:text;not null" json:"tif_file_path"`
	TifFileName   string `gorm:"size:255;not null" json:"tif_file_name"`

	GeoserverWorkspace string `gorm:"size:255" json:"geoserver_workspace,omitempty"`
	GeoserverLayerName string `gorm:"size:255;index" json:"geoserver_layer_name,omitempty"`
	GeoserverStoreName string `gorm:"size:255" json:"geoserver_store_name,omitempty"`

	FileSizeBytes    int64  `json:"file_size_bytes"`
	ExtractionStatus string `gorm:"size:32;default:'pending';index" json:"extraction_status"`
	GeoserverStatus  string `gorm:"size:32;default:'pending';index" json:"geoserver_status"`
	ErrorMessage     string `gorm:"type:text" json:"error_message,omitempty"`

	Metadata datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Model *Model `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}

// Result-processing statuses, re-exported from the contracts package.
const (
	ResultExtractionPending    = contracts.ResultExtractionPending
	ResultExtractionProcessing = contracts.ResultExtractionProcessing
	ResultExtractionCompleted  = contracts.ResultExtractionCompleted
	ResultExtractionFailed     = contracts.ResultExtractionFailed

	ResultGeoserverPending    = contracts.ResultGeoserverPending
	ResultGeoserverProcessing = contracts.ResultGeoserverProcessing
	ResultGeoserverConfigured = contracts.ResultGeoserverConfigured
	ResultGeoserverFailed     = contracts.ResultGeoserverFailed
)

func (ModelResult) TableName() string {
	return "model_results"
}
