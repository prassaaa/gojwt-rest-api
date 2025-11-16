package domain

import "errors"

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrUserAlreadyExists          = errors.New("user with this email already exists")
	ErrInvalidCredentials         = errors.New("invalid email or password")
	ErrInvalidRequest             = errors.New("invalid request body")
	ErrValidationFailed           = errors.New("validation failed")
	ErrRegistrationFailed         = errors.New("registration failed")
	ErrLoginFailed                = errors.New("login failed")
	ErrFailedToHashPassword       = errors.New("failed to hash password")
	ErrFailedToGenerateToken      = errors.New("failed to generate token")
	ErrFailedToCreateUser         = errors.New("failed to create user")
	ErrEmailAlreadyInUse          = errors.New("email already in use")
	ErrFailedToUpdateUser         = errors.New("failed to update user")
	ErrInvalidToken               = errors.New("invalid token")
	ErrInvalidSigningMethod       = errors.New("invalid signing method")
	ErrAuthHeaderRequired         = errors.New("authorization header required")
	ErrInvalidAuthHeaderFormat    = errors.New("invalid authorization header format")
	ErrInvalidOrExpiredToken      = errors.New("invalid or expired token")
	ErrRateLimitExceeded          = errors.New("rate limit exceeded")
)

type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}
