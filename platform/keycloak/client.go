package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pkgauth "platform.local/platform/auth"
	applogger "platform.local/platform/logger"
)

const (
	pathGroups = "groups/%s"
	pathUsers  = "users/%s"
)

type Client struct {
	baseURL            string
	realm              string
	adminTokenProvider *pkgauth.AdminTokenProvider
	httpClient         *http.Client
}

type Group struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Path       string              `json:"path"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Enabled  bool   `json:"enabled"`
}

type GroupDetail struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Path       string              `json:"path"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

func NewClient(baseURL, realm string, adminTokenProvider *pkgauth.AdminTokenProvider) *Client {
	return &Client{
		baseURL:            baseURL,
		realm:              realm,
		adminTokenProvider: adminTokenProvider,
		httpClient:         &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	resp, err := c.doOnce(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	// Retry once on 401 with a fresh token
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.adminTokenProvider.Invalidate()
		return c.doOnce(ctx, method, path, body)
	}
	return resp, nil
}

func (c *Client) doOnce(ctx context.Context, method, path string, body any) (*http.Response, error) {
	token, err := c.adminTokenProvider.GetToken()
	if err != nil {
		return nil, fmt.Errorf("get admin token: %w", err)
	}
	url := fmt.Sprintf("%s/admin/realms/%s/%s", c.baseURL, c.realm, path)

	var rbody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		rbody = bytes.NewBuffer(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rbody)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}

func (c *Client) GetUserGroups(ctx context.Context, userID string) ([]Group, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf("users/%s/groups", userID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user groups: status %d", resp.StatusCode)
	}
	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *Client) GetAllGroups(ctx context.Context) ([]Group, error) {
	resp, err := c.do(ctx, http.MethodGet, "groups", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get groups: status %d", resp.StatusCode)
	}
	var groups []Group
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (c *Client) FindGroupByName(ctx context.Context, name string) (*Group, error) {
	groups, err := c.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}
	for _, g := range groups {
		if g.Name == name {
			gg := g
			return &gg, nil
		}
	}
	return nil, nil
}

func (c *Client) EnsureGroupByName(ctx context.Context, name string) (string, error) {
	if g, err := c.FindGroupByName(ctx, name); err == nil && g != nil {
		return g.ID, nil
	} else if err != nil {
		return "", err
	}

	if _, err := c.CreateGroup(ctx, name); err != nil {
		return "", err
	}
	g2, err := c.FindGroupByName(ctx, name)
	if err != nil {
		return "", err
	}
	if g2 == nil {
		return "", fmt.Errorf("group %s not found after creation", name)
	}
	return g2.ID, nil
}

func (c *Client) EnsureManagerDefaultGroup(ctx context.Context, managerID, managerEmail, managerName string) (string, error) {
	name := fmt.Sprintf("Default_%s", managerID)
	groupID, err := c.EnsureGroupByName(ctx, name)
	if err != nil {
		return "", err
	}

	existingGroup, err := c.GetGroup(ctx, groupID)
	if err != nil {
		applogger.ForComponent("keycloak").Warnf("Failed to get group details: %v", err)
	} else {
		attrs := map[string][]string{
			"owner_email": {managerEmail},
			"owner_name":  {managerName},
		}

		// Only set display_name if it doesn't already exist (preserve custom names)
		if existingGroup.Attributes == nil || len(existingGroup.Attributes["display_name"]) == 0 {
			attrs["display_name"] = []string{"Manager_default"}
		}

		if err := c.UpdateGroupAttributes(ctx, groupID, attrs); err != nil {
			applogger.ForComponent("keycloak").Warnf("Failed to set attributes for manager default group: %v", err)
		}
	}

	if err := c.AddUserToGroup(ctx, managerID, groupID); err != nil {
		return groupID, err
	}

	return groupID, nil
}

func (c *Client) IsDefaultGroup(groupName string) bool {
	if strings.EqualFold(groupName, "Default") {
		return true
	}
	return strings.HasPrefix(groupName, "Default_")
}

func (c *Client) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("users/%s/groups/%s", userID, groupID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add user to group: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("users/%s/groups/%s", userID, groupID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("remove user from group: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetUserPrimaryGroup(ctx context.Context, userID string) (string, error) {
	groups, err := c.GetUserGroups(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(groups) == 0 {
		return "", nil
	}

	if len(groups) > 1 {
		sharedDefaultID := ""
		if g, err := c.FindGroupByName(ctx, "Default"); err == nil && g != nil {
			sharedDefaultID = g.ID
		}

		for _, g := range groups {
			if g.ID != sharedDefaultID {
				return g.ID, nil
			}
		}
	}

	return groups[0].ID, nil
}

func (c *Client) GetManagerGroupSet(ctx context.Context, managerUserID string) (map[string]bool, error) {
	groups, err := c.GetUserGroups(ctx, managerUserID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(groups))
	for _, g := range groups {
		set[g.ID] = true
	}
	return set, nil
}

func (c *Client) CreateGroup(ctx context.Context, name string) (string, error) {
	payload := map[string]string{"name": name}
	resp, err := c.do(ctx, http.MethodPost, "groups", payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("create group: status %d", resp.StatusCode)
	}

	groups, err := c.GetAllGroups(ctx)
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		if g.Name == name {
			return g.ID, nil
		}
	}
	return "", nil
}

func (c *Client) doGroupUpdate(ctx context.Context, groupID string, payload interface{}, errorMsg string) error {
	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf(pathGroups, groupID), payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: status %d", errorMsg, resp.StatusCode)
	}
	return nil
}

func (c *Client) UpdateGroupName(ctx context.Context, groupID, name string) error {
	group, err := c.GetGroup(ctx, groupID)
	if err != nil {
		return fmt.Errorf("get group before update: %w", err)
	}

	group.Name = name

	return c.doGroupUpdate(ctx, groupID, group, "update group")
}

func (c *Client) DeleteGroup(ctx context.Context, groupID string) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf(pathGroups, groupID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete group: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetGroupMembers(ctx context.Context, groupID string) ([]map[string]any, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathGroups+"/members", groupID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get group members: status %d", resp.StatusCode)
	}
	var members []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&members); err != nil {
		return nil, err
	}
	return members, nil
}

func (c *Client) GetGroup(ctx context.Context, groupID string) (*GroupDetail, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathGroups, groupID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get group: status %d", resp.StatusCode)
	}
	var g GroupDetail
	if err := json.NewDecoder(resp.Body).Decode(&g); err != nil {
		return nil, err
	}
	if g.Attributes == nil {
		g.Attributes = make(map[string][]string)
	}
	return &g, nil
}

func (c *Client) UpdateGroupAttributes(ctx context.Context, groupID string, attrs map[string][]string) error {
	g, err := c.GetGroup(ctx, groupID)
	if err != nil {
		return err
	}

	if g.Attributes == nil {
		g.Attributes = make(map[string][]string)
	}
	for k, v := range attrs {
		g.Attributes[k] = v
	}

	payload := map[string]any{
		"name":       g.Name,
		"attributes": g.Attributes,
	}

	return c.doGroupUpdate(ctx, groupID, payload, "update group attributes")
}

func (c *Client) SetGroupDisabled(ctx context.Context, groupID string, disabled bool) error {
	val := "false"
	if disabled {
		val = "true"
	}
	return c.UpdateGroupAttributes(ctx, groupID, map[string][]string{"disabled": {val}})
}

func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf(pathUsers, userID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete user: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) SetUserEnabled(ctx context.Context, userID string, enabled bool) error {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathUsers, userID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get user: status %d", resp.StatusCode)
	}
	var current map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&current); err != nil {
		return err
	}
	current["enabled"] = enabled

	resp2, err := c.do(ctx, http.MethodPut, fmt.Sprintf(pathUsers, userID), current)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNoContent && resp2.StatusCode != http.StatusOK {
		return fmt.Errorf("update user enabled: status %d", resp2.StatusCode)
	}
	return nil
}

func (c *Client) LogoutUser(ctx context.Context, userID string) error {
	resp, err := c.do(ctx, http.MethodPost, fmt.Sprintf(pathUsers+"/logout", userID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("logout user: status %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *Client) FindUsers(ctx context.Context, username string) ([]User, error) {
	path := "users"
	if username != "" {
		path += fmt.Sprintf("?username=%s", username)
	}
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("find users: status %d", resp.StatusCode)
	}
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// FindUserByEmail looks up a user by exact email match in Keycloak.
func (c *Client) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	path := fmt.Sprintf("users?email=%s&exact=true", email)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("find user by email: status %d", resp.StatusCode)
	}
	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return &users[0], nil
}

func (c *Client) GetUserAttributes(ctx context.Context, userID string) (map[string][]string, error) {
	resp, err := c.do(ctx, http.MethodGet, fmt.Sprintf(pathUsers, userID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user: status %d", resp.StatusCode)
	}
	var user map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	attrs, ok := user["attributes"].(map[string]any)
	if !ok {
		return make(map[string][]string), nil
	}

	result := make(map[string][]string)
	for k, v := range attrs {
		if arr, ok := v.([]any); ok {
			strArr := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					strArr = append(strArr, s)
				}
			}
			result[k] = strArr
		}
	}
	return result, nil
}

func (c *Client) CreateUser(ctx context.Context, user map[string]interface{}) (string, error) {
	resp, err := c.do(ctx, http.MethodPost, "users", user)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", fmt.Errorf("user already exists")
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create user: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Get the created user ID
	// Location header contains the URL of the new user
	location := resp.Header.Get("Location")
	if location != "" {
		parts := strings.Split(location, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: search by username
	username, ok := user["username"].(string)
	if ok {
		users, err := c.FindUsers(ctx, username)
		if err == nil && len(users) > 0 {
			for _, u := range users {
				if u.Username == username {
					return u.ID, nil
				}
			}
		}
	}

	return "", nil
}
