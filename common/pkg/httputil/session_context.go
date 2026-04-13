package httputil

import (
	"net/http"
	"strings"

	platformsession "platform.local/platform/session"

	"github.com/gin-gonic/gin"
)

// GetSessionFromContext retrieves session data from context (set by auth middleware)
func GetSessionFromContext(c *gin.Context) (*platformsession.SessionData, bool) {
	sessionData, exists := c.Get("session_data")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found in context"})
		c.Abort()
		return nil, false
	}

	session, ok := sessionData.(*platformsession.SessionData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid session data"})
		c.Abort()
		return nil, false
	}

	return session, true
}

// Legacy compatibility - now gets session from context instead of store
func GetSessionOrAbort(c *gin.Context, _ interface{}) (*platformsession.SessionData, bool) {
	return GetSessionFromContext(c)
}

// GetSessionCookie retrieves the session_id cookie from the request
func GetSessionCookie(c *gin.Context) (string, error) {
	return c.Cookie("session_id")
}

// GetSessionCookieOrEmpty retrieves the session_id cookie, returns empty string if not found
func GetSessionCookieOrEmpty(c *gin.Context) string {
	sessionID, _ := c.Cookie("session_id")
	return sessionID
}

// SetSessionContext sets user data in gin context from session data
func SetSessionContext(c *gin.Context, sessionData *platformsession.SessionData) {
	c.Set("session_data", sessionData)
	c.Set("user_id", sessionData.UserID)
	c.Set("access_level", sessionData.AccessLevel)
	c.Set("group_id", sessionData.GroupID)
	if sessionData.UserInfoData != nil {
		c.Set("user_email", sessionData.UserInfoData.Email)
		c.Set("user_name", sessionData.UserInfoData.FullName)
	}
}

// IsPublicPath checks if the given path matches any of the public paths
func IsPublicPath(path string, publicPaths []string) bool {
	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// Common public paths shared across services
var CommonPublicPaths = []string{
	"/api/health",
	"/api/v1/calculation/callback/",
	"/assets/",
	"/images/",
}
