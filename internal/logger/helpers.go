package logger

import (
	"go.uber.org/zap"
)

// Common #

// Bad Request
func LogBadRequest(requestID string, err error) {
	Info("Invalid request body",
		zap.String("request_id", requestID),
		zap.Error(err),
	)
}

// Failed to create Token
func FailedToCreateToken(requestID string, err error, email string, password string) {
	Error("Failed to generate token",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("email", email),
		zap.String("password", password),
	)
}

// Invalid userID
func InvalidUserID(requestID string, err error, id string) {
	Error("Invalid user ID",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("user_id", id),
	)
}

// Forbidden ID
func ForbiddenID(requestID string, err error, id uint) {
	Error("Forbidden ID",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.Uint("user_id", id),
	)
}

// User not found
func UserNotFound(requestID string, err error, id uint) {
	Error("User not found",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.Uint("user_id", id),
	)
}

// Register #

// Email Already Exists
func EmailAlreadyExists(requestID string, err error, email string) {
	Error("Email already exists",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("email", email),
	)
}

// Failed to Register
func FailedToRegister(requestID string, err error, name string, email string, password string) {
	Error("Failed to register user",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("name", name),
		zap.String("email", email),
		zap.String("password", password),
	)
}

// Login #

// Invalid Credentials
func InvalidCredentials(requestID string, err error, email string, password string) {
	Error("Invalid email or password",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("email", email),
		zap.String("password", password),
	)
}

// Failed To Login
func FailedToLogin(requestID string, err error, email string, password string) {
	Error("Failed to login user",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("email", email),
		zap.String("password", password),
	)
}

// Get User #

// Failed To Get User By ID
func FailedToGetUser(requestID string, err error, id uint) {
	Error("Failed to get user",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.Uint("user_id", id),
	)
}

// Update User #

// Failed To Update User
func FailedToUpdateUser(requestID string, err error, name string, email string) {
	Error("Failed to update user",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("name", name),
		zap.String("email", email),
	)
}

// Delete User #

// Failed to delete user
func FailedToDeleteUser(requestID string, err error, id uint) {
	Error("Failed to delete user",
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.Uint("user_id", id),
	)
}
