package domain

import (
	"time"
)

// User represents the user entity
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	IsAdmin   bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// RefreshToken represents the refresh token entity
type RefreshToken struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index"`
	Token        string    `gorm:"unique;not null;type:varchar(500)"`
	TokenFamily  string    `gorm:"not null;index;type:varchar(100)"` // For detecting token reuse
	ExpiresAt    time.Time `gorm:"not null;index"`
	IsRevoked    bool      `gorm:"default:false;index"`
	RevokedAt    *time.Time
	ReplacedBy   *string   `gorm:"type:varchar(500)"` // Track token rotation
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	User         User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for GORM
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsValid checks if the refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && time.Now().Before(rt.ExpiresAt)
}

// TokenBlacklist represents blacklisted access tokens (for logout)
type TokenBlacklist struct {
	ID        uint      `gorm:"primaryKey"`
	Token     string    `gorm:"unique;not null;type:varchar(500);index"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
