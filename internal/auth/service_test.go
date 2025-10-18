package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.JWTConfig
		expected struct {
			secret string
			ttl    time.Duration
		}
	}{
		{
			name: "with provided config",
			cfg: &config.JWTConfig{
				Secret:   "test-secret",
				TTLHours: 48,
			},
			expected: struct {
				secret string
				ttl    time.Duration
			}{
				secret: "test-secret",
				ttl:    48 * time.Hour,
			},
		},
		{
			name: "with empty secret defaults to default",
			cfg: &config.JWTConfig{
				Secret:   "",
				TTLHours: 12,
			},
			expected: struct {
				secret string
				ttl    time.Duration
			}{
				secret: "default-secret-change-in-production",
				ttl:    12 * time.Hour,
			},
		},
		{
			name: "with zero TTL defaults to 24 hours",
			cfg: &config.JWTConfig{
				Secret:   "test-secret",
				TTLHours: 0,
			},
			expected: struct {
				secret string
				ttl    time.Duration
			}{
				secret: "test-secret",
				ttl:    24 * time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService := NewService(tt.cfg)
			assert.NotNil(t, authService)
			assert.Implements(t, (*Service)(nil), authService)

			// Test that we can generate and validate tokens
			token, err := authService.GenerateToken(123, "test@example.com", "Test User")
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			claims, err := authService.ValidateToken(token)
			assert.NoError(t, err)
			assert.Equal(t, uint(123), claims.UserID)
		})
	}
}

func TestService_GenerateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	service := NewService(cfg)

	tests := []struct {
		name     string
		userID   uint
		email    string
		userName string
	}{
		{
			name:     "successful token generation",
			userID:   123,
			email:    "test@example.com",
			userName: "Test User",
		},
		{
			name:     "token generation with empty email",
			userID:   456,
			email:    "",
			userName: "User Name",
		},
		{
			name:     "token generation with empty name",
			userID:   789,
			email:    "user@example.com",
			userName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.email, tt.userName)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify token can be parsed
			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})

			assert.NoError(t, err)
			assert.True(t, parsedToken.Valid)

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			assert.True(t, ok)

			// Verify claims
			expectedUserID := fmt.Sprintf("%d", tt.userID)
			assert.Equal(t, expectedUserID, claims["sub"])
			assert.Equal(t, tt.email, claims["email"])
			assert.Equal(t, tt.userName, claims["name"])

			// Verify expiration is set
			exp, ok := claims["exp"].(float64)
			assert.True(t, ok)
			assert.True(t, exp > float64(time.Now().Unix()))

			// Verify issued at is set
			iat, ok := claims["iat"].(float64)
			assert.True(t, ok)
			assert.True(t, iat <= float64(time.Now().Unix()))
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:   "test-secret",
		TTLHours: 24,
	}
	service := NewService(cfg)

	t.Run("valid token", func(t *testing.T) {
		// Generate a valid token
		token, err := service.GenerateToken(123, "test@example.com", "Test User")
		assert.NoError(t, err)

		// Validate the token
		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, uint(123), claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)
		assert.Equal(t, "Test User", claims.Name)
	})

	t.Run("invalid token format", func(t *testing.T) {
		claims, err := service.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, claims)
	})

	t.Run("token with wrong signature", func(t *testing.T) {
		// Create token with different secret
		wrongService := NewService(&config.JWTConfig{
			Secret:   "wrong-secret",
			TTLHours: 24,
		})
		token, err := wrongService.GenerateToken(123, "test@example.com", "Test User")
		assert.NoError(t, err)

		// Try to validate with correct service
		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create service with very short TTL
		shortTTLService := NewService(&config.JWTConfig{
			Secret:   "test-secret",
			TTLHours: 0, // This will default to 24, but we'll create token manually
		})

		// Create an expired token manually
		now := time.Now()
		expiredTime := now.Add(-time.Hour) // Expired 1 hour ago

		claims := jwt.MapClaims{
			"sub":   "123",
			"email": "test@example.com",
			"name":  "Test User",
			"exp":   expiredTime.Unix(),
			"iat":   now.Add(-2 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		// Try to validate expired token
		validatedClaims, err := shortTTLService.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredToken, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("token with invalid user ID", func(t *testing.T) {
		// Create token with invalid user ID
		claims := jwt.MapClaims{
			"sub":   "invalid-id",
			"email": "test@example.com",
			"name":  "Test User",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("token without sub claim", func(t *testing.T) {
		// Create token without sub claim
		claims := jwt.MapClaims{
			"email": "test@example.com",
			"name":  "Test User",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("token with wrong signing method", func(t *testing.T) {
		// Create token with RS256 instead of HS256
		claims := jwt.MapClaims{
			"sub":   "123",
			"email": "test@example.com",
			"name":  "Test User",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
		assert.Nil(t, validatedClaims)
	})

	t.Run("token with optional missing fields", func(t *testing.T) {
		// Create token with only required sub claim
		claims := jwt.MapClaims{
			"sub": "456",
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("test-secret"))
		assert.NoError(t, err)

		validatedClaims, err := service.ValidateToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, validatedClaims)
		assert.Equal(t, uint(456), validatedClaims.UserID)
		assert.Equal(t, "", validatedClaims.Email)
		assert.Equal(t, "", validatedClaims.Name)
	})
}
