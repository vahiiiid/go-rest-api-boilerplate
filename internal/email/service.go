package email

import (
	"fmt"
	"log"
	"time"
)

// EmailService defines operations to send application emails
type EmailService interface {
	SendPasswordResetEmail(to string, token string, expiresAt time.Time) error
}

// ConsoleEmailService logs emails to stdout (development fallback)
type ConsoleEmailService struct{}

func (s *ConsoleEmailService) SendPasswordResetEmail(to string, token string, expiresAt time.Time) error {
	link := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	log.Printf("[PASSWORD RESET] Email would be sent to: %s", to)
	log.Printf("[PASSWORD RESET] Reset link: %s", link)
	log.Printf("[PASSWORD RESET] Token expires at: %s", expiresAt.Format("2006-01-02 15:04:05"))
	return nil
}
