package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/ctx"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
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
// @Description Register a new user with name, email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 409 {object} errors.APIError "Email already exists"
// @Failure 500 {object} errors.APIError "Failed to register user or Failed to generate token"
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

	token, err := h.authService.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  ToUserResponse(user),
	})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Invalid email or password"
// @Failure 500 {object} errors.APIError "Failed to authenticate user or Failed to generate token"
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

	token, err := h.authService.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  ToUserResponse(user),
	})
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.APIError "Invalid user ID"
// @Failure 403 {object} errors.APIError "Forbidden user ID"
// @Failure 404 {object} errors.APIError "User not found"
// @Failure 429 {object} errors.RateLimitError "Rate limit exceeded"
// @Failure 500 {object} errors.APIError "Failed to get user"
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	if !ctx.CanAccessUser(c, uint(id)) {
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

	c.JSON(http.StatusOK, ToUserResponse(user))
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
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.APIError "Invalid user ID or Validation error"
// @Failure 403 {object} errors.APIError "Forbidden user ID"
// @Failure 404 {object} errors.APIError "User not found"
// @Failure 409 {object} errors.APIError "Email already exists"
// @Failure 429 {object} errors.RateLimitError "Rate limit exceeded"
// @Failure 500 {object} errors.APIError "Failed to update user"
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !ctx.CanAccessUser(c, uint(id)) {
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

	c.JSON(http.StatusOK, ToUserResponse(user))
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
// @Failure 400 {object} errors.APIError "Invalid user ID"
// @Failure 403 {object} errors.APIError "Forbidden user ID"
// @Failure 404 {object} errors.APIError "User not found"
// @Failure 429 {object} errors.RateLimitError "Rate limit exceeded"
// @Failure 500 {object} errors.APIError "Failed to delete user"
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !ctx.CanAccessUser(c, uint(id)) {
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
