package models

import (
	"time"
)

// UserSetting stores user preferences
type UserSetting struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	UserID string `json:"user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_user_settings_user_id"`
	Email  string `json:"email" gorm:"type:varchar(255);index:idx_user_settings_email"`

	PrivacyAccepted         bool `json:"privacy_accepted" gorm:"default:false"`
	ProductTourCompleted    bool `json:"product_tour_completed" gorm:"default:false"`
	AreaSelectTourCompleted bool `json:"area_select_tour_completed" gorm:"default:false"`
	ModelIntroCardDismissed bool `json:"model_intro_card_dismissed" gorm:"default:false"`
	OnboardingCompleted     bool `json:"onboarding_completed" gorm:"default:false"`

	MapLocation *string `json:"map_location,omitempty" gorm:"type:jsonb"`

	WeatherLocation *string `json:"weather_location,omitempty" gorm:"type:jsonb"`

	Theme    string `json:"theme" gorm:"type:varchar(20);default:'system'"`
	Language string `json:"language" gorm:"type:varchar(10);default:'en'"`

	PreferredWorkspaceID *uint `json:"preferred_workspace_id,omitempty" gorm:"index:idx_user_settings_preferred_workspace"`

	EmailNotifications   bool `json:"email_notifications" gorm:"default:true"`
	BrowserNotifications bool `json:"browser_notifications" gorm:"default:true"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserSetting) TableName() string {
	return "user_settings"
}

// Theme constants
const (
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"
)

// Language constants
const (
	LanguageEN = "en"
	LanguageES = "es"
	LanguageDE = "de"
)
