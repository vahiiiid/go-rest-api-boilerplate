package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int
	PerPage int
}

// ParsePaginationParams parses and validates pagination parameters from request
func ParsePaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = DefaultPage
	}

	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}
