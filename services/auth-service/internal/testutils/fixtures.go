package testutils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"terminal-terrace/auth-service/internal/model/user"
)

// CreateTestUser creates a test user with unique username/email
func CreateTestUser(db *gorm.DB, opts ...UserOption) *user.User {
	uniqueID := uuid.New().String()
	username := fmt.Sprintf("test_user_%s", uniqueID)
	email := fmt.Sprintf("test_%s@example.com", uniqueID)
	
	// Default password hash
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	
	testUser := &user.User{
		Username:     &username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         "student",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	for _, opt := range opts {
		opt(testUser)
	}

	if err := db.Create(testUser).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test user: %v", err))
	}

	return testUser
}

// UserOption configures test user
type UserOption func(*user.User)

// WithUsername sets the username
func WithUsername(username string) UserOption {
	return func(u *user.User) {
		u.Username = &username
	}
}

// WithEmail sets the email
func WithEmail(email string) UserOption {
	return func(u *user.User) {
		u.Email = email
	}
}

// WithRole sets the role
func WithRole(role string) UserOption {
	return func(u *user.User) {
		u.Role = role
	}
}

// WithPassword sets the password (will be hashed)
func WithPassword(password string) UserOption {
	return func(u *user.User) {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		u.PasswordHash = string(hash)
	}
}

// WithPasswordHash sets the password hash directly
func WithPasswordHash(hash string) UserOption {
	return func(u *user.User) {
		u.PasswordHash = hash
	}
}

