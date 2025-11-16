package helpers

import (
	"gojwt-rest-api/internal/domain"
	"gojwt-rest-api/internal/utils"
	"time"
)

// CreateTestUser creates a test user with default values
func CreateTestUser(id uint, email string) *domain.User {
	hashedPassword, _ := utils.HashPassword("password123")
	return &domain.User{
		ID:        id,
		Name:      "Test User",
		Email:     email,
		Password:  hashedPassword,
		IsAdmin:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateAdminUser creates a test admin user
func CreateAdminUser(id uint, email string) *domain.User {
	user := CreateTestUser(id, email)
	user.IsAdmin = true
	user.Name = "Admin User"
	return user
}

// CreateTestUsers creates multiple test users
func CreateTestUsers(count int) []*domain.User {
	users := make([]*domain.User, count)
	for i := 0; i < count; i++ {
		users[i] = CreateTestUser(uint(i+1), "user"+string(rune(i+1))+"@test.com")
	}
	return users
}

// CreateRegisterRequest creates a test registration request
func CreateRegisterRequest(name, email, password string) *domain.RegisterRequest {
	return &domain.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}
}

// CreateLoginRequest creates a test login request
func CreateLoginRequest(email, password string) *domain.LoginRequest {
	return &domain.LoginRequest{
		Email:    email,
		Password: password,
	}
}

// CreateUpdateUserRequest creates a test update user request
func CreateUpdateUserRequest(name, email string) *domain.UpdateUserRequest {
	return &domain.UpdateUserRequest{
		Name:  name,
		Email: email,
	}
}

// CreatePaginationQuery creates a test pagination query
func CreatePaginationQuery(page, pageSize int, search string) *domain.PaginationQuery {
	return &domain.PaginationQuery{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}
}
