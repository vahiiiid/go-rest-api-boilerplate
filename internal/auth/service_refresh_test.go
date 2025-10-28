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

func setupServiceTest(t *testing.T) (*service, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&RefreshToken{})
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
