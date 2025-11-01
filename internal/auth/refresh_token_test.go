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
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&RefreshToken{})
	require.NoError(t, err)

	return db
}

func TestHashToken(t *testing.T) {
	token := "test-token-123"
	hash1 := HashToken(token)
	hash2 := HashToken(token)

	assert.Equal(t, hash1, hash2, "Same token should produce same hash")
	assert.Len(t, hash1, 64, "SHA256 hash should be 64 characters")

	differentToken := "different-token"
	hash3 := HashToken(differentToken)
	assert.NotEqual(t, hash1, hash3, "Different tokens should produce different hashes")
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	tokenFamily := uuid.New()
	token := &RefreshToken{
		UserID:      1,
		TokenHash:   "test-hash",
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, token.ID)
}

func TestRefreshTokenRepository_FindByTokenHash(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	tests := []struct {
		name          string
		setupToken    *RefreshToken
		searchHash    string
		expectedError bool
	}{
		{
			name: "valid token found",
			setupToken: &RefreshToken{
				UserID:      1,
				TokenHash:   "valid-hash",
				TokenFamily: uuid.New(),
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
			searchHash:    "valid-hash",
			expectedError: false,
		},
		{
			name: "expired token still returned by repository",
			setupToken: &RefreshToken{
				UserID:      1,
				TokenHash:   "expired-hash",
				TokenFamily: uuid.New(),
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			searchHash:    "expired-hash",
			expectedError: false,
		},
		{
			name: "revoked token still returned by repository",
			setupToken: &RefreshToken{
				UserID:      1,
				TokenHash:   "revoked-hash",
				TokenFamily: uuid.New(),
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
				RevokedAt:   ptrTime(time.Now()),
			},
			searchHash:    "revoked-hash",
			expectedError: false,
		},
		{
			name:          "non-existent token",
			setupToken:    nil,
			searchHash:    "non-existent",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupToken != nil {
				err := repo.Create(ctx, tt.setupToken)
				require.NoError(t, err)
			}

			found, err := repo.FindByTokenHash(ctx, tt.searchHash)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, found)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, tt.searchHash, found.TokenHash)
			}
		})
	}
}

func TestRefreshTokenRepository_FindByTokenFamily(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	tokenFamily := uuid.New()

	token1 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash1",
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	token2 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash2",
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token1)
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = repo.Create(ctx, token2)
	require.NoError(t, err)

	tokens, err := repo.FindByTokenFamily(ctx, tokenFamily)
	assert.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "hash2", tokens[0].TokenHash)
	assert.Equal(t, "hash1", tokens[1].TokenHash)
}

func TestRefreshTokenRepository_MarkAsUsed(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	token := &RefreshToken{
		UserID:      1,
		TokenHash:   "test-hash",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token)
	require.NoError(t, err)
	assert.Nil(t, token.UsedAt)

	err = repo.MarkAsUsed(ctx, token.ID)
	assert.NoError(t, err)

	var updated RefreshToken
	err = db.First(&updated, token.ID).Error
	require.NoError(t, err)
	assert.NotNil(t, updated.UsedAt)
}

func TestRefreshTokenRepository_RevokeTokenFamily(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	tokenFamily := uuid.New()

	token1 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash1",
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	token2 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash2",
		TokenFamily: tokenFamily,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token1)
	require.NoError(t, err)
	err = repo.Create(ctx, token2)
	require.NoError(t, err)

	err = repo.RevokeTokenFamily(ctx, tokenFamily)
	assert.NoError(t, err)

	var tokens []RefreshToken
	err = db.Where("token_family = ?", tokenFamily).Find(&tokens).Error
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.NotNil(t, tokens[0].RevokedAt)
	assert.NotNil(t, tokens[1].RevokedAt)
}

func TestRefreshTokenRepository_RevokeByUserID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	token1 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash1",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	token2 := &RefreshToken{
		UserID:      1,
		TokenHash:   "hash2",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	token3 := &RefreshToken{
		UserID:      2,
		TokenHash:   "hash3",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, token1)
	require.NoError(t, err)
	err = repo.Create(ctx, token2)
	require.NoError(t, err)
	err = repo.Create(ctx, token3)
	require.NoError(t, err)

	err = repo.RevokeByUserID(ctx, 1)
	assert.NoError(t, err)

	var user1Tokens []RefreshToken
	err = db.Where("user_id = ?", 1).Find(&user1Tokens).Error
	require.NoError(t, err)
	assert.Len(t, user1Tokens, 2)
	assert.NotNil(t, user1Tokens[0].RevokedAt)
	assert.NotNil(t, user1Tokens[1].RevokedAt)

	var user2Tokens []RefreshToken
	err = db.Where("user_id = ?", 2).Find(&user2Tokens).Error
	require.NoError(t, err)
	assert.Len(t, user2Tokens, 1)
	assert.Nil(t, user2Tokens[0].RevokedAt)
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshTokenRepository(db)
	ctx := context.Background()

	expiredToken := &RefreshToken{
		UserID:      1,
		TokenHash:   "expired",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	validToken := &RefreshToken{
		UserID:      1,
		TokenHash:   "valid",
		TokenFamily: uuid.New(),
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	err := repo.Create(ctx, expiredToken)
	require.NoError(t, err)
	err = repo.Create(ctx, validToken)
	require.NoError(t, err)

	err = repo.DeleteExpired(ctx)
	assert.NoError(t, err)

	var tokens []RefreshToken
	err = db.Find(&tokens).Error
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, "valid", tokens[0].TokenHash)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
