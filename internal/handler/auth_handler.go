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
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("invalid request body", err))
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("validation failed", err))
		return
	}

	// Register user
	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("registration failed", err))
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse("user registered successfully", user.ToResponse()))
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("invalid request body", err))
		return
	}

	// Validate request
	if err := h.validator.Validate(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("validation failed", err))
		return
	}

	// Login user
	response, _, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse("login failed", err))
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("login successful", response))
}
