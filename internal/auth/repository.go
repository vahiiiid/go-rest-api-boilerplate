package auth

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TokenRepository abstracts persistence for password reset tokens
type TokenRepository interface {
	InvalidateUserTokens(ctx context.Context, userID uint) error
	Create(ctx context.Context, userID uint, tokenHash string, expiresAt time.Time) error
	FindValidByHash(ctx context.Context, tokenHash string, now time.Time) (*PasswordResetToken, error)
	MarkUsed(ctx context.Context, tokenID uint) error
}

type gormTokenRepository struct {
	db *gorm.DB
}

// NewGormTokenRepository constructs a GORM-backed token repository
func NewGormTokenRepository(db *gorm.DB) TokenRepository {
	return &gormTokenRepository{db: db}
}

func (r *gormTokenRepository) InvalidateUserTokens(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&PasswordResetToken{}).
		Where("user_id = ? AND used = ?", userID, false).
		Updates(map[string]any{"used": true}).Error
}

func (r *gormTokenRepository) Create(ctx context.Context, userID uint, tokenHash string, expiresAt time.Time) error {
	rec := &PasswordResetToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *gormTokenRepository) FindValidByHash(ctx context.Context, tokenHash string, now time.Time) (*PasswordResetToken, error) {
	var rec PasswordResetToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used = ? AND expires_at > ?", tokenHash, false, now).
		First(&rec).Error; err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *gormTokenRepository) MarkUsed(ctx context.Context, tokenID uint) error {
	return r.db.WithContext(ctx).Model(&PasswordResetToken{}).
		Where("id = ?", tokenID).
		Updates(map[string]any{"used": true}).Error
}
