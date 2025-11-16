package domain

// RegisterRequest represents registration request
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents login response with tokens
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"` // seconds until access token expires
	TokenType    string        `json:"token_type"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}

// PaginationQuery represents pagination parameters
type PaginationQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Search   string `json:"search" form:"search"`
}

// PaginatedResponse represents paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// ChangePasswordRequest represents change password request for self-service
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// UpdateProfileRequest represents update own profile request for self-service
type UpdateProfileRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}
