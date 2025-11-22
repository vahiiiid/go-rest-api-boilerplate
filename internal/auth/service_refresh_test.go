package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
)

// testUser is a minimal user struct for testing
type testUser struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (testUser) TableName() string {
	return "users"
}

type testRole struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (testRole) TableName() string {
	return "roles"
}

type testUserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (testUserRole) TableName() string {
	return "user_roles"
}

func setupServiceTest(t *testing.T) (*service, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&RefreshToken{}, &testUser{}, &testRole{}, &testUserRole{})
	require.NoError(t, err)

	testRoleData := &testRole{
		ID:          1,
		Name:        "user",
		Description: "Standard user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(testRoleData).Error
	require.NoError(t, err)

	testUserData := &testUser{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(testUserData).Error
	require.NoError(t, err)

	testUserRoleData := &testUserRole{
		UserID: 1,
		RoleID: 1,
	}
	err = db.Create(testUserRoleData).Error
	require.NoError(t, err)

	cfg := &config.JWTConfig{
		Secret:          "test-secret",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	svc := &service{
		jwtSecret:        cfg.Secret,
		accessTokenTTL:   cfg.AccessTokenTTL,
		refreshTokenTTL:  cfg.RefreshTokenTTL,
		refreshTokenRepo: NewRefreshTokenRepository(db),
		db:               db,
	}

	return svc, db
}

func TestService_GenerateTokenPair(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	tokenPair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn)

	claims, err := svc.ValidateToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "Test User", claims.Name)
}

func TestService_RefreshAccessToken_Success(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	originalPair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	require.NoError(t, err)

	time.Sleep(time.Second)

	newPair, err := svc.RefreshAccessToken(ctx, originalPair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
	assert.NotEqual(t, originalPair.AccessToken, newPair.AccessToken)
	assert.NotEqual(t, originalPair.RefreshToken, newPair.RefreshToken)
	assert.Equal(t, originalPair.TokenFamily, newPair.TokenFamily)
}

func TestService_RefreshAccessToken_ReuseDetection(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	originalPair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	require.NoError(t, err)

	_, err = svc.RefreshAccessToken(ctx, originalPair.RefreshToken)
	require.NoError(t, err)

	_, err = svc.RefreshAccessToken(ctx, originalPair.RefreshToken)
	assert.ErrorIs(t, err, ErrTokenReuse)

	var tokens []RefreshToken
	err = db.Where("token_family = ?", originalPair.TokenFamily).Find(&tokens).Error
	require.NoError(t, err)
	for _, token := range tokens {
		assert.NotNil(t, token.RevokedAt, "All tokens in family should be revoked")
	}
}

func TestService_RefreshAccessToken_InvalidToken(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	_, err := svc.RefreshAccessToken(ctx, "invalid-token")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestService_RefreshAccessToken_ExpiredToken(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	tokenFamily := uuid.New()
	expiredToken := &RefreshToken{
		UserID:      1,
		TokenHash:   HashToken("expired-refresh-token"),
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}

	err := db.Create(expiredToken).Error
	require.NoError(t, err)

	_, err = svc.RefreshAccessToken(ctx, "expired-refresh-token")
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestService_RefreshAccessToken_RevokedToken(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	tokenFamily := uuid.New()
	now := time.Now()
	revokedToken := &RefreshToken{
		UserID:      1,
		TokenHash:   HashToken("revoked-refresh-token"),
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		RevokedAt:   &now,
	}

	err := db.Create(revokedToken).Error
	require.NoError(t, err)

	_, err = svc.RefreshAccessToken(ctx, "revoked-refresh-token")
	assert.ErrorIs(t, err, ErrTokenRevoked)
}

func TestService_RevokeRefreshToken(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	tokenPair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	require.NoError(t, err)

	err = svc.RevokeRefreshToken(ctx, tokenPair.RefreshToken)
	assert.NoError(t, err)

	var tokens []RefreshToken
	err = db.Where("token_family = ?", tokenPair.TokenFamily).Find(&tokens).Error
	require.NoError(t, err)
	assert.NotEmpty(t, tokens)
	for _, token := range tokens {
		assert.NotNil(t, token.RevokedAt)
	}
}

func TestService_RevokeAllUserTokens(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	pair1, err := svc.GenerateTokenPair(ctx, 1, "user1@example.com", "User 1")
	require.NoError(t, err)
	pair2, err := svc.GenerateTokenPair(ctx, 1, "user1@example.com", "User 1")
	require.NoError(t, err)
	pair3, err := svc.GenerateTokenPair(ctx, 2, "user2@example.com", "User 2")
	require.NoError(t, err)

	err = svc.RevokeAllUserTokens(ctx, 1)
	assert.NoError(t, err)

	var user1Tokens []RefreshToken
	err = db.Where("user_id = ?", 1).Find(&user1Tokens).Error
	require.NoError(t, err)
	for _, token := range user1Tokens {
		assert.NotNil(t, token.RevokedAt)
	}

	var user2Tokens []RefreshToken
	err = db.Where("user_id = ?", 2).Find(&user2Tokens).Error
	require.NoError(t, err)
	for _, token := range user2Tokens {
		assert.Nil(t, token.RevokedAt)
	}

	_ = pair1
	_ = pair2
	_ = pair3
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := generateRandomToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := generateRandomToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	assert.NotEqual(t, token1, token2, "Each token should be unique")
}

func TestService_GenerateTokenPair_NilRepository(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-for-jwt-tokens-min-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	svc := NewService(cfg)
	ctx := context.Background()

	_, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token repository not initialized")
}

func TestService_RefreshAccessToken_NilRepository(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-for-jwt-tokens-min-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	svc := NewService(cfg)
	ctx := context.Background()

	_, err := svc.RefreshAccessToken(ctx, "some-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token repository not initialized")
}

func TestService_RevokeRefreshToken_NilRepository(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-for-jwt-tokens-min-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	svc := NewService(cfg)
	ctx := context.Background()

	err := svc.RevokeRefreshToken(ctx, "some-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token repository not initialized")
}

func TestService_RevokeAllUserTokens_NilRepository(t *testing.T) {
	cfg := &config.JWTConfig{
		Secret:          "test-secret-for-jwt-tokens-min-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	}

	svc := NewService(cfg)
	ctx := context.Background()

	err := svc.RevokeAllUserTokens(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token repository not initialized")
}

func TestService_RevokeRefreshToken_TokenNotFound(t *testing.T) {
	svc, _ := setupServiceTest(t)
	ctx := context.Background()

	err := svc.RevokeRefreshToken(ctx, "non-existent-token")
	assert.NoError(t, err)
}

func TestService_RevokeUserRefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T, svc *service, db *gorm.DB) (userID uint, refreshToken string)
		userID        uint
		expectedError string
	}{
		{
			name: "successful_revocation",
			setupFunc: func(t *testing.T, svc *service, db *gorm.DB) (uint, string) {
				ctx := context.Background()
				pair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
				require.NoError(t, err)
				return 1, pair.RefreshToken
			},
			expectedError: "",
		},
		{
			name: "token_does_not_belong_to_user",
			setupFunc: func(t *testing.T, svc *service, db *gorm.DB) (uint, string) {
				ctx := context.Background()
				pair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
				require.NoError(t, err)
				return 2, pair.RefreshToken
			},
			expectedError: "token does not belong to user",
		},
		{
			name: "token_not_found_returns_nil",
			setupFunc: func(t *testing.T, svc *service, db *gorm.DB) (uint, string) {
				return 1, "non-existent-token-12345"
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, db := setupServiceTest(t)
			ctx := context.Background()

			userID, refreshToken := tt.setupFunc(t, svc, db)

			err := svc.RevokeUserRefreshToken(ctx, userID, refreshToken)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RevokeUserRefreshToken_NilRepository(t *testing.T) {
	svc := &service{
		jwtSecret:        "test-secret",
		accessTokenTTL:   15 * time.Minute,
		refreshTokenTTL:  7 * 24 * time.Hour,
		refreshTokenRepo: nil,
	}
	ctx := context.Background()

	err := svc.RevokeUserRefreshToken(ctx, 1, "some-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token repository not initialized")
}

func TestService_RefreshAccessToken_UserNotFound(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	pair, err := svc.GenerateTokenPair(ctx, 999, "nonexistent@example.com", "Ghost User")
	require.NoError(t, err)

	err = db.Exec("DELETE FROM users WHERE id = 999").Error
	require.NoError(t, err)

	_, err = svc.RefreshAccessToken(ctx, pair.RefreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch user for token claims")
}

func TestService_GenerateTokenPair_DatabaseError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&RefreshToken{}, &testUser{}, &testRole{}, &testUserRole{})
	require.NoError(t, err)

	testRoleData := &testRole{
		ID:          1,
		Name:        "user",
		Description: "Standard user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(testRoleData).Error
	require.NoError(t, err)

	testUserData := &testUser{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(testUserData).Error
	require.NoError(t, err)

	testUserRoleData := &testUserRole{
		UserID: 1,
		RoleID: 1,
	}
	err = db.Create(testUserRoleData).Error
	require.NoError(t, err)

	svc := &service{
		jwtSecret:        "test-secret",
		accessTokenTTL:   15 * time.Minute,
		refreshTokenTTL:  7 * 24 * time.Hour,
		refreshTokenRepo: NewRefreshTokenRepository(db),
		db:               db,
	}

	db.Exec("DROP TABLE refresh_tokens")

	ctx := context.Background()
	_, err = svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store refresh token")
}

func TestService_RefreshAccessToken_MarkAsUsedError(t *testing.T) {
	svc, db := setupServiceTest(t)
	ctx := context.Background()

	pair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	require.NoError(t, err)

	db.Exec("DROP TABLE refresh_tokens")

	_, err = svc.RefreshAccessToken(ctx, pair.RefreshToken)
	assert.Error(t, err)
}

func TestService_GenerateTokenPair_InvalidSecret(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&RefreshToken{}, &testUser{}, &testRole{}, &testUserRole{})
	require.NoError(t, err)

	testRoleData := &testRole{
		ID:          1,
		Name:        "user",
		Description: "Standard user",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = db.Create(testRoleData).Error
	require.NoError(t, err)

	testUserData := &testUser{
		ID:           1,
		Name:         "Test User",
		Email:        "test@example.com",
		PasswordHash: "hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = db.Create(testUserData).Error
	require.NoError(t, err)

	testUserRoleData := &testUserRole{
		UserID: 1,
		RoleID: 1,
	}
	err = db.Create(testUserRoleData).Error
	require.NoError(t, err)

	svc := &service{
		jwtSecret:        "",
		accessTokenTTL:   15 * time.Minute,
		refreshTokenTTL:  7 * 24 * time.Hour,
		refreshTokenRepo: NewRefreshTokenRepository(db),
		db:               db,
	}

	ctx := context.Background()
	pair, err := svc.GenerateTokenPair(ctx, 1, "test@example.com", "Test User")
	assert.NoError(t, err)
	assert.NotNil(t, pair)
}
