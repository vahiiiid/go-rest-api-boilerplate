package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token is expired
	ErrExpiredToken = errors.New("token expired")
	// ErrTokenReuse is returned when a refresh token is reused
	ErrTokenReuse = errors.New("token reuse detected")
	// ErrTokenRevoked is returned when a refresh token has been revoked
	ErrTokenRevoked = errors.New("token has been revoked")
)

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	TokenFamily  uuid.UUID `json:"-"`
}

// Service defines authentication service interface
type Service interface {
	GenerateToken(userID uint, email string, name string) (string, error)
	GenerateTokenPair(ctx context.Context, userID uint, email string, name string) (*TokenPair, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	ValidateToken(tokenString string) (*Claims, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	RevokeUserRefreshToken(ctx context.Context, userID uint, refreshToken string) error
	RevokeAllUserTokens(ctx context.Context, userID uint) error
}

type service struct {
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	refreshTokenRepo RefreshTokenRepository
	db               *gorm.DB
}

// NewService creates a new authentication service using typed config
func NewService(cfg *config.JWTConfig) Service {
	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	accessTokenTTL := cfg.AccessTokenTTL
	if accessTokenTTL == 0 {
		if cfg.TTLHours > 0 {
			accessTokenTTL = time.Duration(cfg.TTLHours) * time.Hour
		} else {
			accessTokenTTL = 15 * time.Minute
		}
	}

	refreshTokenTTL := cfg.RefreshTokenTTL
	if refreshTokenTTL == 0 {
		refreshTokenTTL = 168 * time.Hour
	}

	return &service{
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// NewServiceWithRepo creates a new authentication service with refresh token repository
func NewServiceWithRepo(cfg *config.JWTConfig, db *gorm.DB) Service {
	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	accessTokenTTL := cfg.AccessTokenTTL
	if accessTokenTTL == 0 {
		if cfg.TTLHours > 0 {
			accessTokenTTL = time.Duration(cfg.TTLHours) * time.Hour
		} else {
			accessTokenTTL = 15 * time.Minute
		}
	}

	refreshTokenTTL := cfg.RefreshTokenTTL
	if refreshTokenTTL == 0 {
		refreshTokenTTL = 168 * time.Hour
	}

	return &service{
		jwtSecret:        jwtSecret,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		refreshTokenRepo: NewRefreshTokenRepository(db),
		db:               db,
	}
}

// GenerateToken generates a JWT token for a user (deprecated: use GenerateTokenPair)
func (s *service) GenerateToken(userID uint, email string, name string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.accessTokenTTL)

	var roles []string
	if s.db != nil {
		var roleNames []string
		err := s.db.Table("roles").
			Select("roles.name").
			Joins("JOIN user_roles ON user_roles.role_id = roles.id").
			Where("user_roles.user_id = ?", userID).
			Find(&roleNames).Error
		if err != nil {
			// WHY: Security-critical - token with empty roles bypasses authorization
			return "", fmt.Errorf("failed to fetch user roles: %w", err)
		}
		roles = roleNames
	}

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", userID),
		"email": email,
		"name":  name,
		"roles": roles,
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

	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userID, err := strconv.ParseUint(subStr, 10, 32)
	if err != nil {
		return nil, ErrInvalidToken
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)

	var roles []string
	if rolesInterface, ok := claims["roles"].([]interface{}); ok {
		for _, role := range rolesInterface {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
	}

	return &Claims{
		UserID: uint(userID),
		Email:  email,
		Name:   name,
		Roles:  roles,
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens with rotation support
func (s *service) GenerateTokenPair(ctx context.Context, userID uint, email string, name string) (*TokenPair, error) {
	if s.refreshTokenRepo == nil {
		return nil, errors.New("refresh token repository not initialized")
	}

	accessToken, err := s.GenerateToken(userID, email, name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	tokenFamily := uuid.New()
	refreshTokenHash := HashToken(refreshToken)

	dbToken := &RefreshToken{
		UserID:      userID,
		TokenHash:   refreshTokenHash,
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, dbToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenFamily:  tokenFamily,
	}, nil
}

// RefreshAccessToken validates refresh token and generates new token pair with rotation
func (s *service) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	if s.refreshTokenRepo == nil {
		return nil, errors.New("refresh token repository not initialized")
	}

	tokenHash := HashToken(refreshToken)

	storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	if storedToken.RevokedAt != nil {
		return nil, ErrTokenRevoked
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	if storedToken.UsedAt != nil {
		if err := s.refreshTokenRepo.RevokeTokenFamily(ctx, storedToken.TokenFamily); err != nil {
			return nil, fmt.Errorf("failed to revoke token family: %w", err)
		}
		return nil, ErrTokenReuse
	}

	if err := s.refreshTokenRepo.MarkAsUsed(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	type userModel struct {
		ID    uint
		Email string
		Name  string
	}
	var user userModel
	if err := s.db.WithContext(ctx).Table("users").Select("id, email, name").Where("id = ?", storedToken.UserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user for token claims: %w", err)
	}

	accessToken, err := s.GenerateToken(storedToken.UserID, user.Email, user.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newTokenHash := HashToken(newRefreshToken)
	newDBToken := &RefreshToken{
		UserID:      storedToken.UserID,
		TokenHash:   newTokenHash,
		TokenFamily: storedToken.TokenFamily,
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
	}

	if err := s.refreshTokenRepo.Create(ctx, newDBToken); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenFamily:  storedToken.TokenFamily,
	}, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *service) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	if s.refreshTokenRepo == nil {
		return errors.New("refresh token repository not initialized")
	}

	tokenHash := HashToken(refreshToken)
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed to find refresh token: %w", err)
	}

	return s.refreshTokenRepo.RevokeTokenFamily(ctx, storedToken.TokenFamily)
}

// RevokeUserRefreshToken revokes a specific refresh token for an authenticated user
func (s *service) RevokeUserRefreshToken(ctx context.Context, userID uint, refreshToken string) error {
	if s.refreshTokenRepo == nil {
		return errors.New("refresh token repository not initialized")
	}

	tokenHash := HashToken(refreshToken)
	storedToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed to find refresh token: %w", err)
	}

	if storedToken.UserID != userID {
		return ErrTokenDoesNotBelongToUser
	}

	return s.refreshTokenRepo.RevokeTokenFamily(ctx, storedToken.TokenFamily)
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *service) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	if s.refreshTokenRepo == nil {
		return errors.New("refresh token repository not initialized")
	}

	return s.refreshTokenRepo.RevokeByUserID(ctx, userID)
}

// generateRandomToken generates a cryptographically secure random token
func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
