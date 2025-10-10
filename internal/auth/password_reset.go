package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken represents a password reset token stored in DB
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	TokenHash string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"not null;default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for PasswordResetToken model
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// PasswordResetService defines operations for password reset tokens
type PasswordResetService interface {
	CreateToken(ctx context.Context, userID uint, ttl time.Duration) (string, time.Time, error)
	ValidateAndConsume(ctx context.Context, token string) (uint, error)
}

type passwordResetService struct {
	db   *gorm.DB
	repo TokenRepository
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(db *gorm.DB) PasswordResetService {
	return &passwordResetService{db: db, repo: NewGormTokenRepository(db)}
}

// CreateToken creates and persists a password reset token for a user
func (s *passwordResetService) CreateToken(ctx context.Context, userID uint, ttl time.Duration) (string, time.Time, error) {
	if userID == 0 {
		return "", time.Time{}, errors.New("invalid user id")
	}

	// Invalidate previous unused tokens for this user
	if err := s.repo.InvalidateUserTokens(ctx, userID); err != nil {
		return "", time.Time{}, err
	}

	// Generate a secure random token (32 bytes -> 64 hex chars)
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(randomBytes)

	// Hash the token before storing
	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	record := &PasswordResetToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(ttl),
	}

	if err := s.repo.Create(ctx, record.UserID, record.TokenHash, record.ExpiresAt); err != nil {
		return "", time.Time{}, err
	}
	return token, record.ExpiresAt, nil
}

// ValidateAndConsume validates a token, marks it used, and returns the associated user ID
func (s *passwordResetService) ValidateAndConsume(ctx context.Context, token string) (uint, error) {
	if token == "" {
		return 0, errors.New("token required")
	}

	// Hash the provided token and lookup by hash
	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	rec, err := s.repo.FindValidByHash(ctx, tokenHash, time.Now().UTC())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("invalid or expired token")
		}
		return 0, err
	}

	// Mark as used
	if err := s.repo.MarkUsed(ctx, rec.ID); err != nil {
		return 0, err
	}

	return rec.UserID, nil
}
