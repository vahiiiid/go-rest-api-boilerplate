package tests

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/ctx"
)

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user claims when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		result := ctx.GetUser(c)
		assert.NotNil(t, result)
		assert.Equal(t, uint(42), result.UserID)
		assert.Equal(t, "test@example.com", result.Email)
	})

	t.Run("returns nil when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		result := ctx.GetUser(c)
		assert.Nil(t, result)
	})

	t.Run("returns nil when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		result := ctx.GetUser(c)
		assert.Nil(t, result)
	})
}

func TestMustGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user claims when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		result, err := ctx.MustGetUser(c)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(42), result.UserID)
		assert.Equal(t, "test@example.com", result.Email)
	})

	t.Run("returns error when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		result, err := ctx.MustGetUser(c)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "user not found in context")
	})
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user ID when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		userID := ctx.GetUserID(c)
		assert.Equal(t, uint(42), userID)
	})

	t.Run("returns 0 when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		userID := ctx.GetUserID(c)
		assert.Equal(t, uint(0), userID)
	})

	t.Run("returns 0 when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		userID := ctx.GetUserID(c)
		assert.Equal(t, uint(0), userID)
	})
}

func TestMustGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user ID when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		userID, err := ctx.MustGetUserID(c)
		assert.NoError(t, err)
		assert.Equal(t, uint(42), userID)
	})

	t.Run("returns error when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		userID, err := ctx.MustGetUserID(c)
		assert.Error(t, err)
		assert.Equal(t, uint(0), userID)
		assert.Contains(t, err.Error(), "user ID not found in context")
	})

	t.Run("returns error when user ID is 0", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 0, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		userID, err := ctx.MustGetUserID(c)
		assert.Error(t, err)
		assert.Equal(t, uint(0), userID)
		assert.Contains(t, err.Error(), "user ID not found in context")
	})
}

func TestGetEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns email when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		email := ctx.GetEmail(c)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("returns empty string when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		email := ctx.GetEmail(c)
		assert.Equal(t, "", email)
	})

	t.Run("returns empty string when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		email := ctx.GetEmail(c)
		assert.Equal(t, "", email)
	})
}

func TestIsAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns true when user is present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		authenticated := ctx.IsAuthenticated(c)
		assert.True(t, authenticated)
	})

	t.Run("returns false when user is not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		authenticated := ctx.IsAuthenticated(c)
		assert.False(t, authenticated)
	})

	t.Run("returns false when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		authenticated := ctx.IsAuthenticated(c)
		assert.False(t, authenticated)
	})
}

func TestCanAccessUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns true when user can access own resource", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		canAccess := ctx.CanAccessUser(c, 42)
		assert.True(t, canAccess)
	})

	t.Run("returns false when user cannot access other resource", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com"}
		c.Set(auth.KeyUser, claims)

		canAccess := ctx.CanAccessUser(c, 43)
		assert.False(t, canAccess)
	})

	t.Run("returns false when user is not authenticated", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		canAccess := ctx.CanAccessUser(c, 42)
		assert.False(t, canAccess)
	})
}

func TestGetUserName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns user name when present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com", Name: "John Doe"}
		c.Set(auth.KeyUser, claims)

		userName := ctx.GetUserName(c)
		assert.Equal(t, "John Doe", userName)
	})

	t.Run("returns empty string when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		userName := ctx.GetUserName(c)
		assert.Equal(t, "", userName)
	})

	t.Run("returns empty string when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		userName := ctx.GetUserName(c)
		assert.Equal(t, "", userName)
	})
}

func TestHasRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns false when user is present (roles not implemented)", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &auth.Claims{UserID: 42, Email: "test@example.com", Name: "John Doe"}
		c.Set(auth.KeyUser, claims)

		hasRole := ctx.HasRole(c, "admin")
		assert.False(t, hasRole) // Currently returns false as roles are not implemented
	})

	t.Run("returns false when user is not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		hasRole := ctx.HasRole(c, "admin")
		assert.False(t, hasRole)
	})

	t.Run("returns false when wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(auth.KeyUser, "not-a-claims-struct")

		hasRole := ctx.HasRole(c, "admin")
		assert.False(t, hasRole)
	})
}
