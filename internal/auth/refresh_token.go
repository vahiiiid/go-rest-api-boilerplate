package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTokenDoesNotBelongToUser = errors.New("token does not belong to user")
)

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	UserID      uint      `gorm:"not null;index"`
	TokenHash   string    `gorm:"type:varchar(64);not null;index"`
	TokenFamily uuid.UUID `gorm:"type:uuid;not null;index"`
	ExpiresAt   time.Time `gorm:"not null;index"`
	UsedAt      *time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// BeforeCreate is a GORM hook that sets the ID and CreatedAt before creating the record
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	if rt.CreatedAt.IsZero() {
		rt.CreatedAt = time.Now()
	}
	return nil
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	FindByTokenFamily(ctx context.Context, tokenFamily uuid.UUID) ([]*RefreshToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	RevokeTokenFamily(ctx context.Context, tokenFamily uuid.UUID) error
	RevokeByUserID(ctx context.Context, userID uint) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// HashToken creates a SHA256 hash of the token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) FindByTokenFamily(ctx context.Context, tokenFamily uuid.UUID) ([]*RefreshToken, error) {
	var tokens []*RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_family = ?", tokenFamily).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *refreshTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("id = ?", id).
		Where("used_at IS NULL").
		Update("used_at", now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("token already used or not found")
	}

	return nil
}

func (r *refreshTokenRepository) RevokeTokenFamily(ctx context.Context, tokenFamily uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("token_family = ?", tokenFamily).
		Where("revoked_at IS NULL").
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) RevokeByUserID(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshToken{}).Error
}
