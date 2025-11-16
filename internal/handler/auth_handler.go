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
