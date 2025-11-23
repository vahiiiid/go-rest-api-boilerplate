package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParsePaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		expected PaginationParams
	}{
		{
			name:  "default values",
			query: "",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "valid page and per_page",
			query: "page=2&per_page=50",
			expected: PaginationParams{
				Page:    2,
				PerPage: 50,
			},
		},
		{
			name:  "page less than 1 defaults to 1",
			query: "page=0",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "negative page defaults to 1",
			query: "page=-5",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "per_page less than 1 defaults to 20",
			query: "per_page=0",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "negative per_page defaults to 20",
			query: "per_page=-10",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "per_page exceeding max capped at 100",
			query: "per_page=200",
			expected: PaginationParams{
				Page:    1,
				PerPage: 100,
			},
		},
		{
			name:  "per_page at max boundary",
			query: "per_page=100",
			expected: PaginationParams{
				Page:    1,
				PerPage: 100,
			},
		},
		{
			name:  "per_page at min boundary",
			query: "per_page=1",
			expected: PaginationParams{
				Page:    1,
				PerPage: 1,
			},
		},
		{
			name:  "invalid page string defaults to 1",
			query: "page=abc",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "invalid per_page string defaults to 20",
			query: "per_page=xyz",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "large page number",
			query: "page=999999",
			expected: PaginationParams{
				Page:    999999,
				PerPage: 20,
			},
		},
		{
			name:  "both invalid",
			query: "page=invalid&per_page=invalid",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "page with decimal",
			query: "page=2.5",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "per_page with decimal",
			query: "per_page=25.7",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
		{
			name:  "empty strings",
			query: "page=&per_page=",
			expected: PaginationParams{
				Page:    1,
				PerPage: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)

			result := ParsePaginationParams(c)

			assert.Equal(t, tt.expected.Page, result.Page)
			assert.Equal(t, tt.expected.PerPage, result.PerPage)
		})
	}
}

func TestPaginationConstants(t *testing.T) {
	assert.Equal(t, 1, DefaultPage)
	assert.Equal(t, 20, DefaultPerPage)
	assert.Equal(t, 100, MaxPerPage)
}
