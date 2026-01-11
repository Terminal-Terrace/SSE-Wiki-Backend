package article_test

import (
	"testing"
	"time"

	"terminal-terrace/sse-wiki/internal/dto"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/testutils"
)

// TestAddCollaborator_Integration 集成测试：添加协作者
func TestAddCollaborator_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本（CreateArticle 会自动创建，这里手动创建以简化测试）
	initialVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Initial content",
		CommitMessage: "Initial commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(initialVersion)
	testArticle.CurrentVersionID = &initialVersion.ID
	db.Save(testArticle)

	// 添加作者为 admin 协作者（CreateArticle 会自动添加，这里确保存在）
	db.Table("article_collaborators").FirstOrCreate(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    author.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name         string
		operatorID   uint
		userRole     string
		targetUserID uint
		targetRole   string
		expectError  bool
		errorMsg     string
	}{
		// Author 可以添加 admin
		{
			name:         "Author can add admin",
			operatorID:   author.ID,
			userRole:     "user",
			targetUserID: adminUser.ID,
			targetRole:   "admin",
			expectError:  false,
		},
		// Author 可以添加 moderator
		{
			name:         "Author can add moderator",
			operatorID:   author.ID,
			userRole:     "user",
			targetUserID: moderatorUser.ID,
			targetRole:   "moderator",
			expectError:  false,
		},
		// Admin 协作者可以添加 moderator
		{
			name:         "Admin collaborator can add moderator",
			operatorID:   adminUser.ID,
			userRole:     "user",
			targetUserID: moderatorUser.ID,
			targetRole:   "moderator",
			expectError:  false,
		},
		// Admin 协作者不能添加 admin
		{
			name:         "Admin collaborator cannot add admin",
			operatorID:   adminUser.ID,
			userRole:     "user",
			targetUserID: regularUser.ID,
			targetRole:   "admin",
			expectError:  true,
			errorMsg:     "permission denied: only author can add admin collaborators",
		},
		// Global_Admin 不能添加协作者
		{
			name:         "Global_Admin cannot add admin",
			operatorID:   globalAdmin.ID,
			userRole:     "admin",
			targetUserID: regularUser.ID,
			targetRole:   "admin",
			expectError:  true,
			errorMsg:     "permission denied",
		},
		// Global_Admin 不能添加 moderator
		{
			name:         "Global_Admin cannot add moderator",
			operatorID:   globalAdmin.ID,
			userRole:     "admin",
			targetUserID: regularUser.ID,
			targetRole:   "moderator",
			expectError:  true,
			errorMsg:     "permission denied",
		},
		// 普通用户不能添加协作者
		{
			name:         "Regular user cannot add collaborator",
			operatorID:   regularUser.ID,
			userRole:     "user",
			targetUserID: moderatorUser.ID,
			targetRole:   "moderator",
			expectError:  true,
			errorMsg:     "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 先添加 admin 协作者（如果需要）
			if tt.operatorID == adminUser.ID && tt.name == "Admin collaborator can add moderator" {
				req := dto.AddCollaboratorRequest{
					UserID: adminUser.ID,
					Role:   "admin",
				}
				service.AddCollaborator(testArticle.ID, author.ID, "user", req)
			}
			
			req := dto.AddCollaboratorRequest{
				UserID: tt.targetUserID,
				Role:   tt.targetRole,
			}
			
			err := service.AddCollaborator(testArticle.ID, tt.operatorID, tt.userRole, req)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证协作者已添加
					collaborators, _ := service.GetCollaborators(testArticle.ID, tt.operatorID, tt.userRole)
					found := false
					for _, c := range collaborators {
						if c.UserID == tt.targetUserID && c.Role == tt.targetRole {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Collaborator not found after adding")
					}
				}
			}
		})
	}
}

// TestRemoveCollaborator_Integration 集成测试：移除协作者
func TestRemoveCollaborator_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本
	initialVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Initial content",
		CommitMessage: "Initial commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(initialVersion)
	testArticle.CurrentVersionID = &initialVersion.ID
	db.Save(testArticle)

	// 添加协作者
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    author.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name         string
		operatorID   uint
		userRole     string
		targetUserID uint
		expectError  bool
		errorMsg     string
	}{
		// Author 可以移除 admin
		{
			name:         "Author can remove admin",
			operatorID:   author.ID,
			userRole:     "user",
			targetUserID: adminUser.ID,
			expectError:  false,
		},
		// Author 可以移除 moderator
		{
			name:         "Author can remove moderator",
			operatorID:   author.ID,
			userRole:     "user",
			targetUserID: moderatorUser.ID,
			expectError:  false,
		},
		// Admin 可以移除 moderator（需要先确保 adminUser 是 admin 协作者）
		{
			name:         "Admin can remove moderator",
			operatorID:   adminUser.ID,
			userRole:     "user",
			targetUserID: moderatorUser.ID,
			expectError:  false,
		},
		// 不能移除 Author
		{
			name:         "Cannot remove author",
			operatorID:   adminUser.ID,
			userRole:     "user",
			targetUserID: author.ID,
			expectError:  true,
			errorMsg:     "cannot remove author",
		},
		// Global_Admin 不能移除协作者
		{
			name:         "Global_Admin cannot remove collaborator",
			operatorID:   globalAdmin.ID,
			userRole:     "admin",
			targetUserID: moderatorUser.ID,
			expectError:  true,
			errorMsg:     "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 确保协作者存在
			if tt.targetUserID != author.ID {
				db.Table("article_collaborators").FirstOrCreate(&article.ArticleCollaborator{
					ArticleID: testArticle.ID,
					UserID:    tt.targetUserID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
			}
			// 确保操作者是 admin（对于 "Admin can remove moderator" 测试）
			if tt.name == "Admin can remove moderator" {
				db.Table("article_collaborators").FirstOrCreate(&article.ArticleCollaborator{
					ArticleID: testArticle.ID,
					UserID:    adminUser.ID,
					Role:      "admin",
					CreatedAt: time.Now(),
				})
			}

			err := service.RemoveCollaborator(testArticle.ID, tt.operatorID, tt.userRole, tt.targetUserID)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证协作者已移除
					collaborators, _ := service.GetCollaborators(testArticle.ID, tt.operatorID, tt.userRole)
					for _, c := range collaborators {
						if c.UserID == tt.targetUserID {
							t.Errorf("Collaborator still exists after removal")
						}
					}
				}
			}
		})
	}
}

// TestGetCollaborators_Integration 集成测试：查询协作者
func TestGetCollaborators_Integration(t *testing.T) {
	svc, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 添加协作者
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    author.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("article_collaborators").Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name        string
		userID      uint
		userRole    string
		expectError bool
		expectCount int
	}{
		// Author 可以查看
		{
			name:        "Author can view collaborators",
			userID:      author.ID,
			userRole:    "user",
			expectError: false,
			expectCount: 3,
		},
		// Admin 协作者可以查看
		{
			name:        "Admin collaborator can view collaborators",
			userID:      adminUser.ID,
			userRole:    "user",
			expectError: false,
			expectCount: 3,
		},
		// Global_Admin 可以查看
		{
			name:        "Global_Admin can view collaborators",
			userID:      globalAdmin.ID,
			userRole:    "admin",
			expectError: false,
			expectCount: 3,
		},
		// 普通用户不能查看
		{
			name:        "Regular user cannot view collaborators",
			userID:      regularUser.ID,
			userRole:    "user",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collaborators, err := svc.GetCollaborators(testArticle.ID, tt.userID, tt.userRole)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(collaborators) != tt.expectCount {
					t.Errorf("Expected %d collaborators, got %d", tt.expectCount, len(collaborators))
				}
			}
		})
	}
}

