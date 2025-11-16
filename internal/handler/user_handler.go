package handler

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/middleware"
	"gojwt-rest-api/internal/service"
	"gojwt-rest-api/pkg/validator"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user requests
type UserHandler struct {
	userService service.UserService
	validator   *validator.Validator
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService, validator *validator.Validator) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator,
	}
}

// GetProfile gets current user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse("user not authenticated", nil))
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse(domain.ErrUserNotFound.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to retrieve user profile", err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("user profile retrieved", user.ToResponse()))
}

// GetUserByID gets a user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("invalid user ID", err))
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse(domain.ErrUserNotFound.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to retrieve user", err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("user retrieved", user.ToResponse()))
}

// GetAllUsers gets all users with pagination
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var pagination domain.PaginationQuery

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")

	pagination.Page = page
	pagination.PageSize = pageSize
	pagination.Search = search

	users, total, err := h.userService.GetAllUsers(&pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to retrieve users", err))
		return
	}

	// Convert to response format
	userResponses := make([]*domain.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))

	response := domain.PaginatedResponse{
		Data:       userResponses,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("users retrieved", response))
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("invalid user ID", err))
		return
	}

	var req domain.UpdateUserRequest

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

	// Update user
	user, err := h.userService.UpdateUser(uint(id), &req)
	if err != nil {
		switch err {
		case domain.ErrEmailAlreadyInUse:
			c.JSON(http.StatusConflict, domain.ErrorResponse(domain.ErrEmailAlreadyInUse.Error(), err))
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse(domain.ErrUserNotFound.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse(domain.ErrFailedToUpdateUser.Error(), err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("user updated successfully", user.ToResponse()))
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse("invalid user ID", err))
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		switch err {
		case domain.ErrUserNotFound:
			c.JSON(http.StatusNotFound, domain.ErrorResponse(domain.ErrUserNotFound.Error(), err))
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse("failed to delete user", err))
		}
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse("user deleted successfully", nil))
}
