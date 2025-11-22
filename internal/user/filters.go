package user

import (
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// UserFilterParams represents filtering parameters for user list
type UserFilterParams struct {
	Role   string
	Search string
	Sort   string
	Order  string
}

// ParseUserFilters parses and validates user filter parameters from request
func ParseUserFilters(c *gin.Context) UserFilterParams {
	role := c.Query("role")
	if role != "" && role != RoleUser && role != RoleAdmin {
		role = ""
	}

	// Sanitize search parameter: limit length and strip dangerous characters
	search := c.Query("search")
	if search != "" {
		// Limit search length to prevent DoS
		if utf8.RuneCountInString(search) > 100 {
			search = string([]rune(search)[:100])
		}
		// Trim whitespace
		search = strings.TrimSpace(search)
	}

	sort := c.DefaultQuery("sort", "created_at")
	validSorts := map[string]bool{
		"name":       true,
		"email":      true,
		"created_at": true,
		"updated_at": true,
	}
	if !validSorts[sort] {
		sort = "created_at"
	}

	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return UserFilterParams{
		Role:   role,
		Search: search,
		Sort:   sort,
		Order:  order,
	}
}
