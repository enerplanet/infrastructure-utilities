package httputil

import (
	"platform.local/common/pkg/constants"

	"github.com/gin-gonic/gin"
)

type UserContext struct {
	UserID      string
	Email       string
	Name        string
	AccessLevel string
	GroupID     string
}

func GetUserContext(c *gin.Context) (*UserContext, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(401, gin.H{"error": "User not authenticated", "code": "SESSION_EXPIRED"})
		return nil, false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		InternalError(c, "Invalid user ID format")
		return nil, false
	}

	email, _ := c.Get("user_email")
	emailStr, _ := email.(string)

	name, _ := c.Get("user_name")
	nameStr, _ := name.(string)

	accessLevel, _ := c.Get("access_level")
	accessLevelStr, _ := accessLevel.(string)

	groupID, _ := c.Get("group_id")
	groupIDStr, _ := groupID.(string)

	return &UserContext{
		UserID:      userIDStr,
		Email:       emailStr,
		Name:        nameStr,
		AccessLevel: accessLevelStr,
		GroupID:     groupIDStr,
	}, true
}

func MustGetUserID(c *gin.Context) (string, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(401, gin.H{"error": "User not authenticated", "code": "SESSION_EXPIRED"})
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		InternalError(c, "Invalid user ID format")
		return "", false
	}

	return userIDStr, true
}

// IsExpertAccess checks if the access level is expert
func IsExpertAccess(accessLevel string) bool {
	return accessLevel == constants.AccessLevelExpert
}

// IsManagerOrExpertAccess checks if the access level is manager or expert
func IsManagerOrExpertAccess(accessLevel string) bool {
	return accessLevel == constants.AccessLevelExpert || accessLevel == constants.AccessLevelManager
}

// RequireExpertAccessFromContext checks if the user has expert access level
func RequireExpertAccessFromContext(userCtx *UserContext, c *gin.Context) bool {
	if userCtx.AccessLevel != constants.AccessLevelExpert {
		Forbidden(c, "Expert access required")
		c.Abort()
		return false
	}
	return true
}

// RequireExpertAccess checks if the current user has expert access level
// Legacy signature for backward compatibility
func RequireExpertAccess(sessionDataOrContext interface{}, c *gin.Context) bool {
	// If called with old signature (sessionData, c), just use context
	accessLevel, exists := c.Get("access_level")
	if !exists {
		Unauthorized(c, "Session not found")
		c.Abort()
		return false
	}

	accessLevelStr, ok := accessLevel.(string)
	if !ok || accessLevelStr != constants.AccessLevelExpert {
		Forbidden(c, "Expert access required")
		c.Abort()
		return false
	}
	return true
}

// RequireManagerOrExpertAccess checks if the user has manager or expert access level
// Legacy signature for backward compatibility
func RequireManagerOrExpertAccess(sessionDataOrContext interface{}, c *gin.Context) bool {
	// If called with old signature (sessionData, c), just use context
	accessLevel, exists := c.Get("access_level")
	if !exists {
		Unauthorized(c, "Session not found")
		c.Abort()
		return false
	}

	accessLevelStr, ok := accessLevel.(string)
	if !ok {
		Forbidden(c, "Invalid access level")
		c.Abort()
		return false
	}

	if accessLevelStr != constants.AccessLevelExpert && accessLevelStr != constants.AccessLevelManager {
		Forbidden(c, "Manager or Expert access required")
		c.Abort()
		return false
	}
	return true
}
