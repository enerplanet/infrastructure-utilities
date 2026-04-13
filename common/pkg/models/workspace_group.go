package models

import (
	"time"
)

type WorkspaceGroup struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WorkspaceID uint      `json:"workspace_id" gorm:"not null;index:idx_workspace_group_workspace"`
	GroupID     string    `json:"group_id" gorm:"type:varchar(255);not null;index:idx_workspace_group_group"`
	GroupName   string    `json:"group_name" gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workspace Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
}

func (WorkspaceGroup) TableName() string {
	return "workspace_groups"
}
