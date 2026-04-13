package models

import (
	"time"
)

type Workspace struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:text"`
	UserID      string    `json:"user_id" gorm:"type:varchar(255);not null;index:idx_workspace_user"`
	UserEmail   string    `json:"user_email" gorm:"type:varchar(255);index:idx_workspace_user_email"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;index:idx_workspace_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Members []WorkspaceMember `json:"members,omitempty" gorm:"foreignKey:WorkspaceID"`
	Groups  []WorkspaceGroup  `json:"groups,omitempty" gorm:"foreignKey:WorkspaceID"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

// IsShared returns true if the workspace has members or groups (excluding owner)
func (w *Workspace) IsShared() bool {
	return len(w.Members) > 0 || len(w.Groups) > 0
}

type WorkspaceMember struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WorkspaceID uint      `json:"workspace_id" gorm:"not null;index:idx_workspace_member_workspace"`
	UserID      string    `json:"user_id" gorm:"type:varchar(255);not null;index:idx_workspace_member_user"`
	Email       string    `json:"email" gorm:"type:varchar(255);not null"`
	JoinedAt    time.Time `json:"joined_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Workspace Workspace `json:"workspace,omitempty" gorm:"foreignKey:WorkspaceID"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

const (
	WorkspaceRoleExpert       = "expert"
	WorkspaceRoleIntermediate = "intermediate"
	WorkspaceRoleLowAccess    = "low_access"
)
