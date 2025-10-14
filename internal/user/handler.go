package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/ctx"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/logger"
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
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	// Get unique requestID
	requestID := logger.GetRequestID(c)

	var req RegisterRequest

	// User registration attempt log
	logger.Info("User registration attempt",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logger.LogBadRequest(requestID, err)
		return
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			logger.EmailAlreadyExists(requestID, err, req.Email)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		logger.FailedToRegister(requestID, err, req.Name, req.Email, req.Password)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		logger.FailedToCreateToken(requestID, err, req.Email, req.Password)
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
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	// Get unique requestID
	requestID := logger.GetRequestID(c)

	var req LoginRequest

	// User login attempt log
	logger.Info("User login attempt",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
		zap.String("password", req.Password),
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logger.LogBadRequest(requestID, err)
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			logger.InvalidCredentials(requestID, err, req.Email, req.Password)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate user"})
		logger.FailedToLogin(requestID, err, req.Email, req.Password)
		return
	}

	token, err := h.authService.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		logger.FailedToCreateToken(requestID, err, req.Email, req.Password)
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
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	// Get unique requestID
	requestID := logger.GetRequestID(c)

	// Get user attempt log
	logger.Info("Get user attempt",
		zap.String("request_id", requestID),
		zap.String("id", c.Param("id")),
	)

	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		logger.InvalidUserID(requestID, err, c.Param("id"))
		return
	}

	if !ctx.CanAccessUser(c, uint(id)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		logger.ForbiddenID(requestID, err, uint(id))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			logger.UserNotFound(requestID, err, uint(id))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		logger.FailedToGetUser(requestID, err, uint(id))
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
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	// Get unique requestID
	requestID := logger.GetRequestID(c)

	var req UpdateUserRequest

	// Update user attempt log
	logger.Info("Update user attempt",
		zap.String("request_id", requestID),
		zap.String("id", c.Param("id")),
		zap.String("email", req.Email),
		zap.String("name", req.Name),
	)

	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		logger.InvalidUserID(requestID, err, c.Param("id"))
		return
	}

	// Authorization check
	if !ctx.CanAccessUser(c, uint(id)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		logger.ForbiddenID(requestID, err, uint(id))
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logger.LogBadRequest(requestID, err)
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			logger.UserNotFound(requestID, err, uint(id))
			return
		}
		if errors.Is(err, ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			logger.EmailAlreadyExists(requestID, err, req.Email)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		logger.FailedToUpdateUser(requestID, err, req.Email, req.Name)
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
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	// Get unique requestID
	requestID := logger.GetRequestID(c)

	// Delete user attempt log
	logger.Info("Delete user attempt",
		zap.String("request_id", requestID),
		zap.String("id", c.Param("id")),
	)

	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		logger.InvalidUserID(requestID, err, c.Param("id"))
		return
	}

	// Authorization check
	if !ctx.CanAccessUser(c, uint(id)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		logger.ForbiddenID(requestID, err, uint(id))
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		logger.FailedToDeleteUser(requestID, err, uint(id))
		return
	}

	c.Status(http.StatusNoContent)
}
