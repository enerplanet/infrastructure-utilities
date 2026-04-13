package models

import (
	"time"

	"gorm.io/gorm"
)

type Feedback struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        string         `json:"user_id" gorm:"type:varchar(255);not null;index"`
	UserEmail     string         `json:"user_email" gorm:"type:varchar(255);not null;index"`
	UserName      string         `json:"user_name" gorm:"type:varchar(255);not null"`
	Category      string         `json:"category" gorm:"type:varchar(50);not null;index"`
	Subject       string         `json:"subject" gorm:"type:varchar(255);not null"`
	Message       string         `json:"message" gorm:"type:text;not null"`
	Rating        int            `json:"rating" gorm:"type:int;default:0;check:rating >= 0 AND rating <= 5"`
	Status        string         `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	Priority      string         `json:"priority" gorm:"type:varchar(20);default:'medium';index"`
	AdminResponse *string        `json:"admin_response" gorm:"type:text"`
	RespondedAt   *time.Time     `json:"responded_at"`
	RespondedBy   *string        `json:"responded_by" gorm:"type:varchar(255);index"`
	ImagePath     *string        `json:"image_path,omitempty" gorm:"type:varchar(500)"`
	ImageMimeType *string        `json:"image_mime_type,omitempty" gorm:"type:varchar(100)"`
	ImageSize     *int64         `json:"image_size,omitempty"`
	Images        *string        `json:"images,omitempty" gorm:"type:text"` // JSON array of {path, mime_type, size}
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	User            UserInfo  `json:"user" gorm:"-"`
	RespondedByUser *UserInfo `json:"responded_by_user,omitempty" gorm:"-"`
}

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type FeedbackStatus string
type FeedbackCategory string
type FeedbackPriority string

const (
	StatusPending    FeedbackStatus = "pending"
	StatusInProgress FeedbackStatus = "in_progress"
	StatusResolved   FeedbackStatus = "resolved"
	StatusClosed     FeedbackStatus = "closed"
)

const (
	CategoryBug         FeedbackCategory = "bug"
	CategoryFeature     FeedbackCategory = "feature"
	CategoryImprovement FeedbackCategory = "improvement"
	CategoryGeneral     FeedbackCategory = "general"
)

const (
	PriorityLow      FeedbackPriority = "low"
	PriorityMedium   FeedbackPriority = "medium"
	PriorityHigh     FeedbackPriority = "high"
	PriorityCritical FeedbackPriority = "critical"
)

var (
	ValidStatuses = map[string]struct{}{
		string(StatusPending):    {},
		string(StatusInProgress): {},
		string(StatusResolved):   {},
		string(StatusClosed):     {},
	}
	ValidCategories = map[string]struct{}{
		string(CategoryBug):         {},
		string(CategoryFeature):     {},
		string(CategoryImprovement): {},
		string(CategoryGeneral):     {},
	}
	ValidPriorities = map[string]struct{}{
		string(PriorityLow):      {},
		string(PriorityMedium):   {},
		string(PriorityHigh):     {},
		string(PriorityCritical): {},
	}
)

func (Feedback) TableName() string {
	return "feedbacks"
}

func (f *Feedback) BeforeCreate(tx *gorm.DB) error {
	if f.Status == "" {
		f.Status = string(StatusPending)
	}
	if f.Priority == "" {
		f.Priority = string(PriorityMedium)
	}
	return nil
}

func (f *Feedback) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("admin_response") && f.AdminResponse != nil && *f.AdminResponse != "" {
		now := time.Now()
		f.RespondedAt = &now
	}
	return nil
}
