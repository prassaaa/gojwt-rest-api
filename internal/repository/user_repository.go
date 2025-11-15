package repository

import (
	"gojwt-rest-api/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id uint) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindAll(pagination *domain.PaginationQuery) ([]*domain.User, int64, error)
	Update(user *domain.User) error
	Delete(id uint) error
}
