package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/contextutil"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

// Handler handles user-related HTTP requests
type Handler struct {
	userService Service
	authService auth.Service
}

// NewHandler creates a new user handler
func NewHandler(userService Service, authService auth.Service) *Handler {
	return &Handler{
		userService: userService,
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 200 {object} errors.Response{success=bool,data=AuthResponse} "Success response with user data and tokens"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Email already exists"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to register user or generate token"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         ToUserResponse(user),
	}))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} errors.Response{success=bool,data=AuthResponse} "Success response with user data and tokens"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid email or password"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to authenticate user or generate token"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			_ = c.Error(apiErrors.Unauthorized("Invalid email or password"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         ToUserResponse(user),
	}))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with user data"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to get user"
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Update request"
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with updated user data"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID or Validation error"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Email already exists"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to update user"
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		if errors.Is(err, ErrEmailExists) {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to delete user"
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange refresh token for new access and refresh tokens with automatic rotation
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} errors.Response{success=bool,data=auth.TokenPairResponse} "Success response with new token pair"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid or expired refresh token"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Token reuse detected - all tokens revoked"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	tokenPair, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) {
			_ = c.Error(apiErrors.Unauthorized("Invalid or expired refresh token"))
			return
		}
		if errors.Is(err, auth.ErrTokenReuse) {
			_ = c.Error(apiErrors.Forbidden("Token reuse detected. All tokens have been revoked for security."))
			return
		}
		if errors.Is(err, auth.ErrTokenRevoked) {
			_ = c.Error(apiErrors.Unauthorized("Token has been revoked"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(auth.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}))
}

// Logout godoc
// @Summary Logout user
// @Description Revoke refresh token and invalidate user session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.RefreshTokenRequest true "Refresh token to revoke"
// @Success 200 {object} errors.Response{success=bool,data=object} "Successfully logged out"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Token does not belong to user"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to logout"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("user not authenticated"))
		return
	}

	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	if err := h.authService.RevokeUserRefreshToken(c.Request.Context(), userID, req.RefreshToken); err != nil {
		if errors.Is(err, auth.ErrTokenDoesNotBelongToUser) {
			_ = c.Error(apiErrors.Forbidden("token does not belong to user"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Successfully logged out"}))
}

// GetMe godoc
// @Summary Get current user
// @Description Get the currently authenticated user's information with roles
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with current user data"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to get user"
// @Router /api/v1/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// ListUsers godoc
// @Summary List all users (Admin only)
// @Description Get paginated list of all users with optional filtering (requires admin role)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page (max 100)" default(20)
// @Param role query string false "Filter by role (user or admin)"
// @Param search query string false "Search by name or email"
// @Param sort query string false "Sort by field (created_at, updated_at, name, email)" default(created_at)
// @Param order query string false "Sort order (asc or desc)" default(desc)
// @Success 200 {object} errors.Response{success=bool,data=UserListResponse} "Success response with paginated user list"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid parameters"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Admin access required"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to list users"
// @Router /api/v1/admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	pagination := middleware.ParsePaginationParams(c)
	filters := ParseUserFilters(c)

	users, total, err := h.userService.ListUsers(c.Request.Context(), filters, pagination.Page, pagination.PerPage)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			_ = c.Error(apiErrors.BadRequest("Invalid role filter"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = ToUserResponse(&user)
	}

	totalPages := int(total) / pagination.PerPage
	if int(total)%pagination.PerPage > 0 {
		totalPages++
	}

	response := UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, apiErrors.Success(response))
}
