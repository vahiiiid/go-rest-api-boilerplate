package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/email"
)

// UserPasswordPort abstracts minimal user operations to avoid import cycles
type UserPasswordPort interface {
	FindUserIDByEmail(ctx context.Context, email string) (uint, error)
	UpdatePassword(ctx context.Context, id uint, newPlainPassword string) error
}

// Handler handles auth-related HTTP requests that are not login/register
type Handler struct {
	resetService PasswordResetService
	userPort     UserPasswordPort
	mailer       email.EmailService
}

// NewHandler creates a new auth handler
func NewHandler(reset PasswordResetService, userPort UserPasswordPort) *Handler {
	return &Handler{resetService: reset, userPort: userPort, mailer: &email.ConsoleEmailService{}}
}

// ForgotPassword initiates a password reset by generating a token
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up user by email; do not reveal whether email exists
	if h.userPort != nil {
		if userID, _ := h.userPort.FindUserIDByEmail(c.Request.Context(), req.Email); userID != 0 {
			// Create token for existing user with 1 hour TTL
			token, expiresAt, _ := h.resetService.CreateToken(c.Request.Context(), userID, time.Hour)
			// Send email with the raw token (not hashed)
			_ = h.mailer.SendPasswordResetEmail(req.Email, token, expiresAt)
			// Errors intentionally ignored to avoid leaking information
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "if the email exists, a reset link has been sent"})
}

// ResetPassword validates the token and updates the user's password
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := h.resetService.ValidateAndConsume(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	if h.userPort == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user service not configured"})
		return
	}

	if err := h.userPort.UpdatePassword(c.Request.Context(), userID, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}
