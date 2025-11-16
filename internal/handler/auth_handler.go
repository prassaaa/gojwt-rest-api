package handler

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/pkg/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userService service.UserService
	validator   *validator.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userService service.UserService, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		validator:   validator,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrInvalidRequest.Error(), err))
		return
	}

	// Validate request
	if validationErrors := h.validator.Validate(&req); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrValidationFailed.Error(), validationErrors))
		return
	}

	// Register user
	user, err := h.userService.Register(&req)
	if err != nil {
		switch err {
		case domain.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, domain.ErrorResponse(domain.ErrUserAlreadyExists.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse(domain.ErrRegistrationFailed.Error(), err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse("user registered successfully", user.ToResponse()))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrInvalidRequest.Error(), err))
		return
	}

	// Validate request
	if validationErrors := h.validator.Validate(&req); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrValidationFailed.Error(), validationErrors))
		return
	}

	// Login user
	response, err := h.userService.Login(&req)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrInvalidCredentials.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse(domain.ErrLoginFailed.Error(), err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("login successful", response))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshTokenRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrInvalidRequest.Error(), err))
		return
	}

	// Validate request
	if validationErrors := h.validator.Validate(&req); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse(domain.ErrValidationFailed.Error(), validationErrors))
		return
	}

	// Refresh token
	response, err := h.userService.RefreshToken(&req)
	if err != nil {
		switch err {
		case domain.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrInvalidRefreshToken.Error(), err))
		case domain.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrTokenExpired.Error(), err))
		case domain.ErrTokenReused:
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrTokenReused.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to refresh token", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("token refreshed successfully", response))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.LogoutRequest

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse(domain.ErrAuthHeaderRequired.Error(), nil))
		return
	}

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		// Logout can work without refresh token (just invalidates current session)
		req.RefreshToken = ""
	}

	// Logout user
	if err := h.userService.Logout(userID.(uint), &req); err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to logout", err.Error()))
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("logout successful", nil))
}

