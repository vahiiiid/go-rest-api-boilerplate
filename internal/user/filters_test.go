package user

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParseUserFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		expected UserFilterParams
	}{
		{
			name:  "default values",
			query: "",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "valid user role",
			query: "role=user",
			expected: UserFilterParams{
				Role:   "user",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "valid admin role",
			query: "role=admin",
			expected: UserFilterParams{
				Role:   "admin",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "invalid role filtered out",
			query: "role=superadmin",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "search parameter",
			query: "search=john",
			expected: UserFilterParams{
				Role:   "",
				Search: "john",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "search with whitespace trimmed",
			query: "search=" + url.QueryEscape("  john  "),
			expected: UserFilterParams{
				Role:   "",
				Search: "john",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "search exceeding 100 characters truncated",
			query: "search=" + strings.Repeat("a", 150),
			expected: UserFilterParams{
				Role:   "",
				Search: strings.Repeat("a", 100),
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "valid sort by name",
			query: "sort=name",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "name",
				Order:  "desc",
			},
		},
		{
			name:  "valid sort by email",
			query: "sort=email",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "email",
				Order:  "desc",
			},
		},
		{
			name:  "valid sort by created_at",
			query: "sort=created_at",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "valid sort by updated_at",
			query: "sort=updated_at",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "updated_at",
				Order:  "desc",
			},
		},
		{
			name:  "invalid sort defaults to created_at",
			query: "sort=invalid_column",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "valid order asc",
			query: "order=asc",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "asc",
			},
		},
		{
			name:  "valid order desc",
			query: "order=desc",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "invalid order defaults to desc",
			query: "order=random",
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "all valid parameters",
			query: "role=admin&search=john&sort=email&order=asc",
			expected: UserFilterParams{
				Role:   "admin",
				Search: "john",
				Sort:   "email",
				Order:  "asc",
			},
		},
		{
			name:  "mixed valid and invalid parameters",
			query: "role=invalid&search=test&sort=invalid&order=invalid",
			expected: UserFilterParams{
				Role:   "",
				Search: "test",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "empty search after whitespace trim",
			query: "search=" + url.QueryEscape("   "),
			expected: UserFilterParams{
				Role:   "",
				Search: "",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "unicode characters in search",
			query: "search=用户测试",
			expected: UserFilterParams{
				Role:   "",
				Search: "用户测试",
				Sort:   "created_at",
				Order:  "desc",
			},
		},
		{
			name:  "unicode search exceeding 100 characters",
			query: "search=" + strings.Repeat("用", 150),
			expected: UserFilterParams{
				Role:   "",
				Search: strings.Repeat("用", 100),
				Sort:   "created_at",
				Order:  "desc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)

			result := ParseUserFilters(c)

			assert.Equal(t, tt.expected.Role, result.Role)
			assert.Equal(t, tt.expected.Search, result.Search)
			assert.Equal(t, tt.expected.Sort, result.Sort)
			assert.Equal(t, tt.expected.Order, result.Order)
		})
	}
}
