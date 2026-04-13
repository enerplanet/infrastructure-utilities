package authclient

import (
	"context"
	"fmt"
	"net/http"
	"os"

	httpclient "platform.local/common/pkg/httpclient"
)

// Client handles communication with auth-service
type Client struct {
	http *httpclient.Client
}

// NewClient creates a new auth-service client
func NewClient() *Client {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}

	return &Client{
		http: httpclient.New(baseURL),
	}
}

// DeleteUserSessions deletes all sessions for a user
func (c *Client) DeleteUserSessions(ctx context.Context, userID string) error {
	resp, err := c.http.Do(ctx, http.MethodDelete, fmt.Sprintf("/internal/sessions/user/%s", userID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	return nil
}

// UpdateSessionGroup updates the group_id for a session
func (c *Client) UpdateSessionGroup(ctx context.Context, sessionID, groupID string) error {
	payload := map[string]string{"group_id": groupID}
	resp, err := c.http.DoJSON(ctx, http.MethodPatch, fmt.Sprintf("/internal/sessions/%s/group", sessionID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service returned status %d", resp.StatusCode)
	}

	return nil
}
