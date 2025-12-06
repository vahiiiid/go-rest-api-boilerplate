package contextutil

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
)

// GetUser retrieves the authenticated user claims from context
// Returns nil if not found or invalid type
func GetUser(c *gin.Context) *auth.Claims {
	value, exists := c.Get(auth.KeyUser)
	if !exists {
		return nil
	}

	claims, ok := value.(*auth.Claims)
	if !ok {
		return nil
	}

	return claims
}

// MustGetUser retrieves user claims or returns error
func MustGetUser(c *gin.Context) (*auth.Claims, error) {
	claims := GetUser(c)
	if claims == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return claims, nil
}

// GetUserID retrieves the authenticated user's ID from context
// Returns 0 if not found
func GetUserID(c *gin.Context) uint {
	claims := GetUser(c)
	if claims == nil {
		return 0
	}
	return claims.UserID
}

// MustGetUserID retrieves user ID or returns error
func MustGetUserID(c *gin.Context) (uint, error) {
	userID := GetUserID(c)
	if userID == 0 {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetEmail retrieves the authenticated user's email from context
func GetEmail(c *gin.Context) string {
	claims := GetUser(c)
	if claims == nil {
		return ""
	}
	return claims.Email
}

// IsAuthenticated checks if request has valid authentication
func IsAuthenticated(c *gin.Context) bool {
	return GetUser(c) != nil
}

// CanAccessUser checks if authenticated user can access target user
func CanAccessUser(c *gin.Context, targetUserID uint) bool {
	if IsAdmin(c) {
		return true
	}
	authenticatedUserID := GetUserID(c)
	return authenticatedUserID == targetUserID
}

// GetUserName retrieves the authenticated user's name from context
func GetUserName(c *gin.Context) string {
	claims := GetUser(c)
	if claims == nil {
		return ""
	}
	return claims.Name
}

// HasRole checks if user has specific role
func HasRole(c *gin.Context, role string) bool {
	claims := GetUser(c)
	if claims == nil {
		return false
	}
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetRoles retrieves user roles from context
func GetRoles(c *gin.Context) []string {
	claims := GetUser(c)
	if claims == nil {
		return []string{}
	}
	return claims.Roles
}

// IsAdmin checks if user has admin role
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin")
}
