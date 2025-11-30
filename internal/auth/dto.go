package auth

// Claims represents JWT token claims
type Claims struct {
	UserID uint     `json:"user_id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
}

// TokenResponse represents token response (deprecated: use TokenPairResponse)
type TokenResponse struct {
	Token string `json:"token"`
}

// TokenPairResponse represents access and refresh token pair response
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
