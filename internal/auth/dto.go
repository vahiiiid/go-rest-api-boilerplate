package auth

// Claims represents JWT token claims
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// TokenResponse represents token response
type TokenResponse struct {
	Token string `json:"token"`
}
