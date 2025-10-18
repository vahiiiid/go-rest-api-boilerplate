package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token is expired
	ErrExpiredToken = errors.New("token expired")
)

// Service defines authentication service interface
type Service interface {
	GenerateToken(userID uint, email string, name string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type service struct {
	jwtSecret string
	jwtTTL    time.Duration
}

// NewService creates a new authentication service using typed config
func NewService(cfg *config.JWTConfig) Service {
	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	ttlHours := cfg.TTLHours
	if ttlHours == 0 {
		ttlHours = 24
	}

	return &service{
		jwtSecret: jwtSecret,
		jwtTTL:    time.Duration(ttlHours) * time.Hour,
	}
}

// GenerateToken generates a JWT token for a user
func (s *service) GenerateToken(userID uint, email string, name string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.jwtTTL)

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"email": email,
		"name":  name,
		"exp":   expirationTime.Unix(),
		"iat":   now.Unix(),
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

	// Extract email from "email" claim
	email, _ := claims["email"].(string) // email is optional

	// Extract name from "name" claim
	name, _ := claims["name"].(string) // name is optional

	return &Claims{
		UserID: uint(userID),
		Email:  email,
		Name:   name,
	}, nil
}
