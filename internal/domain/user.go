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
