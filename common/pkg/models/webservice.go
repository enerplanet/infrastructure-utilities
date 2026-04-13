package models

import (
	"time"

	"gorm.io/gorm"
)

type WebserviceInstance struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Name         *string `gorm:"size:255" json:"name"`
	IP           string  `gorm:"not null;size:45" json:"ip"`
	Port         int     `gorm:"not null" json:"port"`
	Protocol     string  `gorm:"not null;size:10;default:'http'" json:"protocol"`
	Available    bool    `gorm:"default:false" json:"available"`
	Busy         bool    `gorm:"default:false" json:"busy"`
	Status       string  `gorm:"size:20;default:'inactive'" json:"status"`
	StatusReason *string `gorm:"type:text" json:"status_reason"`

	LastCheck     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"last_check"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`

	Endpoint           *string `gorm:"size:255" json:"endpoint"`
	AutoScaling        bool    `gorm:"default:false" json:"auto_scaling"`
	MaxConcurrency     int     `gorm:"default:1;not null" json:"max_concurrency" binding:"min=1,max=100"`
	CurrentConcurrency int     `gorm:"default:0;not null" json:"current_concurrency"`

	CpuUsage    *float64 `gorm:"column:cpu_usage;type:numeric(5,2)" json:"cpu_usage"`
	MemoryUsage *float64 `gorm:"column:memory_usage;type:numeric(5,2)" json:"memory_usage"`

	CreatedByID    *string `gorm:"column:created_by_id;size:64" json:"created_by_id"`
	CreatedByName  *string `gorm:"column:created_by_user;size:255" json:"created_by_user"`
	CreatedByEmail *string `gorm:"column:created_by_user_email;size:255" json:"created_by_user_email"`

	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"-"`
}

const (
	StatusActive      = "active"
	StatusInactive    = "inactive"
	StatusMaintenance = "maintenance"

	ProtocolHTTP  = "http"
	ProtocolHTTPS = "https"
)

func (w *WebserviceInstance) IsHealthy() bool {
	if w.LastHeartbeat == nil {
		return false
	}
	return time.Since(*w.LastHeartbeat) < 5*time.Minute
}

// CanAcceptRequest checks if the webservice is available
func (w *WebserviceInstance) CanAcceptRequest() bool {
	return w.Status == StatusActive && w.Available
}

// IncrementConcurrency increases the current concurrency counter (for monitoring only)
func (w *WebserviceInstance) IncrementConcurrency() {
	w.CurrentConcurrency++
}

// DecrementConcurrency decreases the current concurrency counter (for monitoring only)
func (w *WebserviceInstance) DecrementConcurrency() {
	if w.CurrentConcurrency > 0 {
		w.CurrentConcurrency--
	}
}
