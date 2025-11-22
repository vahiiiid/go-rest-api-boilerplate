package user

import (
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

	search := c.Query("search")

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
