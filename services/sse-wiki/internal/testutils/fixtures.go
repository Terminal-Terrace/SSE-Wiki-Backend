package testutils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/model/user"
)

// CreateTestUser creates a test user with unique username/email
func CreateTestUser(db *gorm.DB, opts ...UserOption) *user.User {
	uniqueID := uuid.New().String()
	username := fmt.Sprintf("test_user_%s", uniqueID)
	email := fmt.Sprintf("test_%s@example.com", uniqueID)
	
	testUser := &user.User{
		ID:        0, // Will be set by database
		Username:  username,
		Email:     email,
		Role:      "student",
		CreatedAt: time.Now(),
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
		u.Username = username
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

// CreateTestModule creates a test module
func CreateTestModule(db *gorm.DB, ownerID uint, opts ...ModuleOption) *module.Module {
	uniqueID := uuid.New().String()
	moduleName := fmt.Sprintf("test_module_%s", uniqueID)
	
	testModule := &module.Module{
		ModuleName: moduleName,
		Description: "Test module description",
		OwnerID:    ownerID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	for _, opt := range opts {
		opt(testModule)
	}

	if err := db.Create(testModule).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test module: %v", err))
	}

	return testModule
}

// ModuleOption configures test module
type ModuleOption func(*module.Module)

// WithModuleName sets the module name
func WithModuleName(name string) ModuleOption {
	return func(m *module.Module) {
		m.ModuleName = name
	}
}

// WithParentID sets the parent module ID
func WithParentID(parentID uint) ModuleOption {
	return func(m *module.Module) {
		m.ParentID = &parentID
	}
}

// CreateTestArticle creates a test article
func CreateTestArticle(db *gorm.DB, moduleID uint, createdBy uint, opts ...ArticleOption) *article.Article {
	uniqueID := uuid.New().String()
	title := fmt.Sprintf("Test Article %s", uniqueID)
	
	testArticle := &article.Article{
		Title:            title,
		ModuleID:         moduleID,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		IsReviewRequired: boolPtr(true),
		ViewCount:        0,
	}

	for _, opt := range opts {
		opt(testArticle)
	}

	if err := db.Create(testArticle).Error; err != nil {
		panic(fmt.Sprintf("Failed to create test article: %v", err))
	}

	return testArticle
}

// ArticleOption configures test article
type ArticleOption func(*article.Article)

// WithTitle sets the article title
func WithTitle(title string) ArticleOption {
	return func(a *article.Article) {
		a.Title = title
	}
}

// WithModuleID sets the module ID
func WithModuleID(moduleID uint) ArticleOption {
	return func(a *article.Article) {
		a.ModuleID = moduleID
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}

