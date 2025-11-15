package repository

import (
	"errors"
	"gojwt-rest-api/internal/domain"

	"gorm.io/gorm"
)

// userRepositoryImpl is the implementation of UserRepository
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

// Create creates a new user
func (r *userRepositoryImpl) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *userRepositoryImpl) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *userRepositoryImpl) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindAll retrieves all users with pagination and search
func (r *userRepositoryImpl) FindAll(pagination *domain.PaginationQuery) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	query := r.db.Model(&domain.User{})

	// Apply search filter if provided
	if pagination.Search != "" {
		searchPattern := "%" + pagination.Search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", searchPattern, searchPattern)
	}

	// Count total items
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (pagination.Page - 1) * pagination.PageSize
	if err := query.Offset(offset).Limit(pagination.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Update updates a user
func (r *userRepositoryImpl) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

// Delete deletes a user by ID
func (r *userRepositoryImpl) Delete(id uint) error {
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}
