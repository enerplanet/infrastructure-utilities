package session

import (
	"context"
	"time"
)

type UserInfoData struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

type SessionData struct {
	UserID               string        `json:"user_id"` // Keycloak user ID
	AccessToken          string        `json:"access_token"`
	RefreshToken         string        `json:"refresh_token"`
	TokenExpiresAt       time.Time     `json:"token_expires_at"`
	UserInfoData         *UserInfoData `json:"user_info_data"`
	AccessLevel          string        `json:"access_level"`
	GroupID              string        `json:"group_id,omitempty"` // Primary group for managers
	ProductTourCompleted bool          `json:"product_tour_completed"`
	CreatedAt            time.Time     `json:"created_at"`
	LastSeenAt           time.Time     `json:"last_seen_at,omitempty"`
}

type SessionStore interface {
	SaveSession(ctx context.Context, userID string, session *SessionData) error
	GetSession(ctx context.Context, userID string) (*SessionData, error)
	DeleteSession(ctx context.Context, userID string) error
	// DeleteSessionsByUser removes all active sessions for the given Keycloak user ID
	// This is used to forcefully log out a user across all devices immediately.
	DeleteSessionsByUser(ctx context.Context, userID string) error
	CheckSession(ctx context.Context, userID string) (bool, error)
	RefreshSessionTTL(ctx context.Context, userID string) error
	// CountActiveSessions returns the number of active sessions in the store.
	CountActiveSessions(ctx context.Context) (int64, error)
	// CountActiveUsers returns the number of distinct users with recent session activity.
	// When userIDs are provided, only those users are counted.
	CountActiveUsers(ctx context.Context, userIDs ...string) (int64, error)
}
