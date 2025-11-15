package domain

import (
	"time"
)

// User represents the user entity
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	Email     string    `json:"email" gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"` // "-" excludes from JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
