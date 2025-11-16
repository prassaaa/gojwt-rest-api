package handler

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/pkg/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileHandler handles user profile self-service endpoints
type ProfileHandler struct {
	userService service.UserService
	validator   *validator.Validator
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(userService service.UserService, validator *validator.Validator) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
		validator:   validator,
	}
}

// GetOwnProfile gets the authenticated user's profile
// @Summary Get own profile
// @Description Get the authenticated user's profile information
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response
// @Router /api/v1/profile [get]
func (h *ProfileHandler) GetOwnProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse("Unauthorized", nil))
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, domain.ErrorResponse("User not found", err))
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("Profile retrieved successfully", user.ToResponse()))
}

// UpdateOwnProfile updates the authenticated user's profile
// @Summary Update own profile
// @Description Update the authenticated user's profile (name and email)
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response
// @Failure 401 {object} domain.Response
// @Failure 409 {object} domain.Response
// @Router /api/v1/profile [put]
func (h *ProfileHandler) UpdateOwnProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse("Unauthorized", nil))
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("Invalid request body", err))
		return
	}

	// Validate request
	if validationErrors := h.validator.Validate(&req); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("Validation failed", validationErrors))
		return
	}

	user, err := h.userService.UpdateOwnProfile(userID.(uint), &req)
	if err != nil {
		switch err {
		case domain.ErrEmailAlreadyInUse:
			c.JSON(http.StatusConflict, domain.ErrorResponse("Email already in use", err))
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse("User not found", err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("Failed to update profile", err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("Profile updated successfully", user.ToResponse()))
}

// ChangePassword changes the authenticated user's password
// @Summary Change password
// @Description Change the authenticated user's password
// @Tags profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.ChangePasswordRequest true "Change password request"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response
// @Failure 401 {object} domain.Response
// @Router /api/v1/profile/password [put]
func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse("Unauthorized", nil))
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("Invalid request body", err))
		return
	}

	// Validate request
	if validationErrors := h.validator.Validate(&req); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("Validation failed", validationErrors))
		return
	}

	err := h.userService.ChangePassword(userID.(uint), &req)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, domain.ErrorResponse("Old password is incorrect", err))
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse("User not found", err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("Failed to change password", err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("Password changed successfully", nil))
}
