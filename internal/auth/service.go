package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token is expired
	ErrExpiredToken = errors.New("token expired")
)

// Service defines authentication service interface
type Service interface {
	GenerateToken(userID uint) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type service struct {
	jwtSecret string
	jwtTTL    time.Duration
}

// NewService creates a new authentication service
func NewService() Service {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	ttlHours := 24
	if ttlStr := os.Getenv("JWT_TTL_HOURS"); ttlStr != "" {
		if hours, err := strconv.Atoi(ttlStr); err == nil {
			ttlHours = hours
		}
	}

	return &service{
		jwtSecret: jwtSecret,
		jwtTTL:    time.Duration(ttlHours) * time.Hour,
	}
}

// GenerateToken generates a JWT token for a user
func (s *service) GenerateToken(userID uint) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.jwtTTL)

	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"exp": expirationTime.Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract user ID from "sub" claim
	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, err := strconv.ParseUint(subStr, 10, 32)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &Claims{
		UserID: uint(userID),
	}, nil
}
