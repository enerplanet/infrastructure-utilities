package models

import "time"

type Group struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GroupMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	GroupID   uint      `json:"group_id" gorm:"not null"`
	UserID    string    `json:"user_id" gorm:"not null"`
	Group     *Group    `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	CreatedAt time.Time `json:"created_at"`
}
