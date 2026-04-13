package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"platform.local/platform/session"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	MaxSessionLifetime = 8 * time.Hour
)

type RedisSessionManager struct {
	client      *goredis.Client
	PrefixState string
	defaultTTL  time.Duration
}

func NewSessionRedisManager(rds *goredis.Client, ttlMinutes int) *RedisSessionManager {
	if ttlMinutes <= 0 {
		ttlMinutes = 60
	}

	return &RedisSessionManager{
		client:      rds,
		PrefixState: "session",
		defaultTTL:  time.Duration(ttlMinutes) * time.Minute,
	}
}
func (s *RedisSessionManager) buildKey(userID string) string {
	return fmt.Sprintf("%s:%s", s.PrefixState, userID)
}

func (s *RedisSessionManager) buildUserSessionsKey(userID string) string {
	return fmt.Sprintf("user_sessions:%s", userID)
}

// SaveSession stores session data in Redis
func (s *RedisSessionManager) SaveSession(ctx context.Context,
	userID string,
	session *session.SessionData) error {

	// Here userID is actually the sessionID which we use as the Redis key suffix
	sessionID := userID
	key := s.buildKey(sessionID)

	exists, err := s.client.Exists(ctx, key).Result()
	if err == nil && exists == 0 {
		if session.CreatedAt.IsZero() {
			session.CreatedAt = time.Now()
		}
	} else {
		existingSession, err := s.GetSession(ctx, userID)
		if err == nil && existingSession != nil && !existingSession.CreatedAt.IsZero() {
			session.CreatedAt = existingSession.CreatedAt
		} else if session.CreatedAt.IsZero() {
			session.CreatedAt = time.Now()
		}
	}

	jsonData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("error marshaling session data: %w", err)
	}

	err = s.client.SetEx(ctx, key, string(jsonData), s.defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("error saving session to redis: %w", err)
	}

	// Maintain reverse index: user -> session IDs
	if session != nil && session.UserID != "" {
		setKey := s.buildUserSessionsKey(session.UserID)
		if err := s.client.SAdd(ctx, setKey, sessionID).Err(); err != nil {
			return fmt.Errorf("error indexing session for user: %w", err)
		}
		// Align set TTL with session key TTL
		_ = s.client.Expire(ctx, setKey, s.defaultTTL).Err()
	}

	return nil
}

// GetSession retrieves session data from Redis
func (s *RedisSessionManager) GetSession(ctx context.Context, userID string) (*session.SessionData, error) {
	// Here userID is actually the sessionID cookie value
	key := s.buildKey(userID)

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting session from redis: %w", err)
	}

	var session session.SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("error unmarshaling session data: %w", err)
	}

	return &session, nil
}

// DeleteSession removes session from Redis
func (s *RedisSessionManager) DeleteSession(ctx context.Context, userID string) error {
	// Here userID is the sessionID
	sessionID := userID
	key := s.buildKey(sessionID)

	// Try to fetch to remove reverse index cleanly
	sess, _ := s.GetSession(ctx, sessionID)

	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error deleting session from redis: %w", err)
	}

	if sess != nil && sess.UserID != "" {
		setKey := s.buildUserSessionsKey(sess.UserID)
		// Remove this sessionID from the user's set
		_ = s.client.SRem(ctx, setKey, sessionID).Err()
		// Optionally expire the set if empty
		if count, err := s.client.SCard(ctx, setKey).Result(); err == nil && count == 0 {
			_ = s.client.Del(ctx, setKey).Err()
		}
	}

	return nil
}

// CheckSession verifies if a session exists
func (s *RedisSessionManager) CheckSession(ctx context.Context, userID string) (bool, error) {
	// Here userID is the sessionID
	key := s.buildKey(userID)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("error checking session existence: %w", err)
	}

	return exists == 1, nil
}

// RefreshSessionTTL extends the session TTL but respects max lifetime
func (s *RedisSessionManager) RefreshSessionTTL(ctx context.Context, userID string) error {
	// Here userID is the sessionID
	sessionID := userID
	key := s.buildKey(sessionID)

	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}

	if session == nil {
		return fmt.Errorf("session does not exist")
	}

	if !session.CreatedAt.IsZero() {
		sessionAge := time.Since(session.CreatedAt)
		if sessionAge >= MaxSessionLifetime {
			_ = s.DeleteSession(ctx, userID)
			return fmt.Errorf("session exceeded maximum lifetime of %v", MaxSessionLifetime)
		}
	}

	err = s.client.Expire(ctx, key, s.defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("error refreshing session TTL: %w", err)
	}

	// Also bump the TTL on the reverse index set
	if session.UserID != "" {
		setKey := s.buildUserSessionsKey(session.UserID)
		_ = s.client.Expire(ctx, setKey, s.defaultTTL).Err()
	}

	return nil
}

// DeleteSessionsByUser removes all sessions for a given Keycloak user ID
func (s *RedisSessionManager) DeleteSessionsByUser(ctx context.Context, userID string) error {
	setKey := s.buildUserSessionsKey(userID)
	sessionIDs, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil && err != goredis.Nil {
		return fmt.Errorf("error fetching user sessions: %w", err)
	}
	// Delete each session key
	for _, sid := range sessionIDs {
		key := s.buildKey(sid)
		_ = s.client.Del(ctx, key).Err()
	}
	// Remove the set mapping
	_ = s.client.Del(ctx, setKey).Err()
	return nil
}

// CountActiveSessions counts the number of active session keys in Redis.
func (s *RedisSessionManager) CountActiveSessions(ctx context.Context) (int64, error) {
	var count int64
	var cursor uint64
	pattern := fmt.Sprintf("%s:*", s.PrefixState)
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, fmt.Errorf("error scanning sessions: %w", err)
		}
		count += int64(len(keys))
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return count, nil
}
