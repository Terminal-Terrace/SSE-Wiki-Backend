package article_test

import (
	"fmt"
	"testing"
	"time"

	"terminal-terrace/sse-wiki/internal/dto"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/testutils"
)
func TestCreateArticle_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户和模块
	author := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)

	tests := []struct {
		name              string
		req               dto.CreateArticleRequest
		userID            uint
		expectError       bool
		expectVersion     bool
		expectTags        bool
		expectCollaborator bool
	}{
		{
			name: "successful creation with tags and review required",
			req: dto.CreateArticleRequest{
				Title:            "Test Article",
				ModuleID:         testModule.ID,
				Content:          "<p>Initial content</p>",
				CommitMessage:    "Initial commit",
				Tags:             []string{"tag1", "tag2"},
				IsReviewRequired: boolPtr(true),
			},
			userID:             author.ID,
			expectError:        false,
			expectVersion:      true,
			expectTags:         true,
			expectCollaborator: true,
		},
		{
			name: "successful creation without tags",
			req: dto.CreateArticleRequest{
				Title:            "Test Article 2",
				ModuleID:         testModule.ID,
				Content:          "<p>Content without tags</p>",
				CommitMessage:    "Second commit",
				Tags:             []string{},
				IsReviewRequired: boolPtr(false),
			},
			userID:             author.ID,
			expectError:        false,
			expectVersion:      true,
			expectTags:         false,
			expectCollaborator: true,
		},
		{
			name: "successful creation with empty tag filtered out",
			req: dto.CreateArticleRequest{
				Title:            "Test Article 3",
				ModuleID:         testModule.ID,
				Content:          "<p>Content</p>",
				CommitMessage:    "Third commit",
				Tags:             []string{"tag1", "", "tag2"},
				IsReviewRequired: boolPtr(true),
			},
			userID:             author.ID,
			expectError:        false,
			expectVersion:      true,
			expectTags:         true,
			expectCollaborator: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateArticle(tt.req, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// 验证返回结果
			if result == nil {
				t.Fatalf("Expected result but got nil")
			}

			articleID, ok := result["id"].(uint)
			if !ok {
				t.Fatalf("Expected article ID in result")
			}

			// 验证文章已创建
			var art article.Article
			if err := db.First(&art, articleID).Error; err != nil {
				t.Fatalf("Article not found: %v", err)
			}

			if art.Title != tt.req.Title {
				t.Errorf("Expected title %q, got %q", tt.req.Title, art.Title)
			}
			if art.ModuleID != tt.req.ModuleID {
				t.Errorf("Expected module_id %d, got %d", tt.req.ModuleID, art.ModuleID)
			}
			if art.CreatedBy != tt.userID {
				t.Errorf("Expected created_by %d, got %d", tt.userID, art.CreatedBy)
			}

			// 验证版本已创建
			if tt.expectVersion {
				if art.CurrentVersionID == nil {
					t.Fatalf("Expected current_version_id to be set")
				}

				var version article.ArticleVersion
				if err := db.First(&version, *art.CurrentVersionID).Error; err != nil {
					t.Fatalf("Version not found: %v", err)
				}

				if version.VersionNumber != 1 {
					t.Errorf("Expected version number 1, got %d", version.VersionNumber)
				}
				if version.Content != tt.req.Content {
					t.Errorf("Expected content %q, got %q", tt.req.Content, version.Content)
				}
				if version.CommitMessage != tt.req.CommitMessage {
					t.Errorf("Expected commit message %q, got %q", tt.req.CommitMessage, version.CommitMessage)
				}
				if version.Status != "published" {
					t.Errorf("Expected status 'published', got %q", version.Status)
				}
				if version.BaseVersionID != nil {
					t.Errorf("Expected base_version_id to be nil for initial version, got %v", version.BaseVersionID)
				}
			}

			// 验证标签
			if tt.expectTags {
				var tags []article.ArticleTag
				db.Where("article_id = ?", articleID).Find(&tags)
				if len(tags) == 0 {
					t.Errorf("Expected tags but found none")
				}
			}

			// 验证协作者
			if tt.expectCollaborator {
				var collaborator article.ArticleCollaborator
				if err := db.Where("article_id = ? AND user_id = ?", articleID, tt.userID).First(&collaborator).Error; err != nil {
					t.Errorf("Expected collaborator not found: %v", err)
				} else if collaborator.Role != "admin" {
					t.Errorf("Expected collaborator role 'admin', got %q", collaborator.Role)
				}
			}
		})
	}
}

// TestGetArticle_Integration 集成测试：获取文章详情（包括版本历史）
func TestGetArticle_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	otherUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	testModule := testutils.CreateTestModule(db, author.ID)

	// 创建文章和版本
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)
	initialVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Initial content",
		CommitMessage: "Initial commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	if err := db.Create(initialVersion).Error; err != nil {
		t.Fatalf("Failed to create initial version: %v", err)
	}
	testArticle.CurrentVersionID = &initialVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 创建第二个版本
	secondVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       "Updated content",
		CommitMessage: "Second commit",
		AuthorID:      author.ID,
		Status:        "published",
		BaseVersionID: &initialVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(secondVersion).Error; err != nil {
		t.Fatalf("Failed to create second version: %v", err)
	}
	testArticle.CurrentVersionID = &secondVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 创建一个pending提交
	pendingVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       "Pending content",
		CommitMessage: "Pending commit",
		AuthorID:      otherUser.ID,
		Status:        "pending",
		BaseVersionID: &secondVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersion).Error; err != nil {
		t.Fatalf("Failed to create pending version: %v", err)
	}
	pendingSubmission := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersion.ID,
		BaseVersionID:     secondVersion.ID,
		SubmittedBy:       otherUser.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(pendingSubmission).Error; err != nil {
		t.Fatalf("Failed to create pending submission: %v", err)
	}

	// 创建标签
	tag1 := &article.Tag{Name: "tag1"}
	tag2 := &article.Tag{Name: "tag2"}
	if err := db.Create(tag1).Error; err != nil {
		t.Fatalf("Failed to create tag1: %v", err)
	}
	if err := db.Create(tag2).Error; err != nil {
		t.Fatalf("Failed to create tag2: %v", err)
	}
	if err := db.Create(&article.ArticleTag{ArticleID: testArticle.ID, TagID: tag1.ID}).Error; err != nil {
		t.Fatalf("Failed to create article tag1: %v", err)
	}
	if err := db.Create(&article.ArticleTag{ArticleID: testArticle.ID, TagID: tag2.ID}).Error; err != nil {
		t.Fatalf("Failed to create article tag2: %v", err)
	}

	tests := []struct {
		name           string
		articleID      uint
		userID         uint
		globalUserRole string
		expectError    bool
		checkFields    func(t *testing.T, result map[string]interface{})
	}{
		{
			name:           "author can get article",
			articleID:      testArticle.ID,
			userID:         author.ID,
			globalUserRole: "",
			expectError:    false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["id"] != testArticle.ID {
					t.Errorf("Expected article ID %d, got %v", testArticle.ID, result["id"])
				}
				if result["content"] != "Updated content" {
					t.Errorf("Expected content 'Updated content', got %v", result["content"])
				}
				if result["version_number"] != 2 {
					t.Errorf("Expected version_number 2, got %v", result["version_number"])
				}
				if result["is_author"] != true {
					t.Errorf("Expected is_author true, got %v", result["is_author"])
				}
				if result["can_delete"] != true {
					t.Errorf("Expected can_delete true, got %v", result["can_delete"])
				}
				// 检查历史记录
				history, ok := result["history"].([]map[string]interface{})
				if !ok || len(history) == 0 {
					t.Errorf("Expected history entries")
				}
				// 检查标签
				tags, ok := result["tags"].([]string)
				if !ok || len(tags) != 2 {
					t.Errorf("Expected 2 tags, got %v", tags)
				}
			},
		},
		{
			name:           "global admin can get article",
			articleID:      testArticle.ID,
			userID:         globalAdmin.ID,
			globalUserRole: "admin",
			expectError:    false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["current_user_role"] != "admin" {
					t.Errorf("Expected current_user_role 'admin', got %v", result["current_user_role"])
				}
				if result["can_delete"] != true {
					t.Errorf("Expected can_delete true for global admin, got %v", result["can_delete"])
				}
			},
		},
		{
			name:           "regular user can get article",
			articleID:      testArticle.ID,
			userID:         otherUser.ID,
			globalUserRole: "",
			expectError:    false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["is_author"] != false {
					t.Errorf("Expected is_author false, got %v", result["is_author"])
				}
				if result["can_delete"] != false {
					t.Errorf("Expected can_delete false, got %v", result["can_delete"])
				}
			},
		},
		{
			name:           "non-existent article returns error",
			articleID:      99999,
			userID:         author.ID,
			globalUserRole: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetArticle(tt.articleID, tt.userID, tt.globalUserRole)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkFields != nil {
				tt.checkFields(t, result)
			}
		})
	}
}

// TestGetArticlesByModule_Integration 集成测试：获取模块下的文章列表
func TestGetArticlesByModule_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)
	otherModule := testutils.CreateTestModule(db, author.ID)

	// 创建多篇文章
	articles := []*article.Article{}
	for i := 0; i < 5; i++ {
		art := testutils.CreateTestArticle(db, testModule.ID, author.ID, testutils.WithTitle(fmt.Sprintf("Article %d", i+1)))
		version := &article.ArticleVersion{
			ArticleID:     art.ID,
			VersionNumber: 1,
			Content:       fmt.Sprintf("<p>Content %d</p>", i+1),
			CommitMessage: fmt.Sprintf("Commit %d", i+1),
			AuthorID:      author.ID,
			Status:        "published",
			CreatedAt:     time.Now(),
		}
		db.Create(version)
		art.CurrentVersionID = &version.ID
		db.Save(art)
		articles = append(articles, art)
	}

	// 创建其他模块的文章
	otherArt := testutils.CreateTestArticle(db, otherModule.ID, author.ID)
	otherVersion := &article.ArticleVersion{
		ArticleID:     otherArt.ID,
		VersionNumber: 1,
		Content:       "Other content",
		CommitMessage: "Other commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(otherVersion)
	otherArt.CurrentVersionID = &otherVersion.ID
	db.Save(otherArt)

	tests := []struct {
		name        string
		moduleID    uint
		page        int
		pageSize    int
		expectTotal  int64
		expectCount  int
		expectError  bool
	}{
		{
			name:        "first page",
			moduleID:    testModule.ID,
			page:        1,
			pageSize:    3,
			expectTotal: 5,
			expectCount: 3,
			expectError: false,
		},
		{
			name:        "second page",
			moduleID:    testModule.ID,
			page:        2,
			pageSize:    3,
			expectTotal: 5,
			expectCount: 2,
			expectError: false,
		},
		{
			name:        "empty module",
			moduleID:    otherModule.ID,
			page:        1,
			pageSize:    10,
			expectTotal: 1,
			expectCount: 1,
			expectError: false,
		},
		{
			name:        "non-existent module",
			moduleID:    99999,
			page:        1,
			pageSize:    10,
			expectTotal: 0,
			expectCount: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetArticlesByModule(tt.moduleID, tt.page, tt.pageSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result["total"] != tt.expectTotal {
				t.Errorf("Expected total %d, got %v", tt.expectTotal, result["total"])
			}
			if result["page"] != tt.page {
				t.Errorf("Expected page %d, got %v", tt.page, result["page"])
			}
			if result["page_size"] != tt.pageSize {
				t.Errorf("Expected page_size %d, got %v", tt.pageSize, result["page_size"])
			}

			articles, ok := result["articles"].([]map[string]interface{})
			if !ok {
				t.Fatalf("Expected articles array")
			}
			if len(articles) != tt.expectCount {
				t.Errorf("Expected %d articles, got %d", tt.expectCount, len(articles))
			}
		})
	}
}

// TestGetVersions_Integration 集成测试：获取版本列表
func TestUpdateBasicInfo_Integration(t *testing.T) {
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
	
	newTitle := "Updated Title"
	newTags := dto.StringSlice{"tag1", "tag2"}
	isReviewRequired := false
	
	tests := []struct {
		name        string
		userID      uint
		userRole    string
		req         dto.UpdateArticleBasicInfoRequest
		expectError bool
		errorMsg    string
	}{
		// Author 可以更新
		{
			name:     "Author can update",
			userID:   author.ID,
			userRole: "user",
			req: dto.UpdateArticleBasicInfoRequest{
				Title: &newTitle,
			},
			expectError: false,
		},
		// Admin 协作者可以更新
		{
			name:     "Admin collaborator can update",
			userID:   adminUser.ID,
			userRole: "user",
			req: dto.UpdateArticleBasicInfoRequest{
				Tags: &newTags,
			},
			expectError: false,
		},
		// Moderator 协作者可以更新
		{
			name:     "Moderator collaborator can update",
			userID:   moderatorUser.ID,
			userRole: "user",
			req: dto.UpdateArticleBasicInfoRequest{
				IsReviewRequired: &isReviewRequired,
			},
			expectError: false,
		},
		// Global_Admin 不能更新
		{
			name:     "Global_Admin cannot update",
			userID:   globalAdmin.ID,
			userRole: "admin",
			req: dto.UpdateArticleBasicInfoRequest{
				Title: &newTitle,
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
		// 普通用户不能更新
		{
			name:     "Regular user cannot update",
			userID:   regularUser.ID,
			userRole: "user",
			req: dto.UpdateArticleBasicInfoRequest{
				Title: &newTitle,
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateBasicInfo(testArticle.ID, tt.userID, tt.userRole, tt.req)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestDeleteArticle_Integration 集成测试：删除文章
func TestDeleteArticle_Integration(t *testing.T) {
	service, db := setupArticleService(t)
	
	// 创建测试用户
	author := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	
	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	
	tests := []struct {
		name        string
		userID      uint
		userRole    string
		setupArticle func() uint
		expectError bool
		errorMsg    string
	}{
		// Author 可以删除
		{
			name:     "Author can delete",
			userID:   author.ID,
			userRole: "user",
			setupArticle: func() uint {
				art := testutils.CreateTestArticle(db, testModule.ID, author.ID)
				return art.ID
			},
			expectError: false,
		},
		// Admin 协作者可以删除
		{
			name:     "Admin collaborator can delete",
			userID:   adminUser.ID,
			userRole: "user",
			setupArticle: func() uint {
				art := testutils.CreateTestArticle(db, testModule.ID, author.ID)
				db.Table("article_collaborators").Create(&article.ArticleCollaborator{
					ArticleID: art.ID,
					UserID:    adminUser.ID,
					Role:      "admin",
					CreatedAt: time.Now(),
				})
				return art.ID
			},
			expectError: false,
		},
		// Global_Admin 可以删除
		{
			name:     "Global_Admin can delete",
			userID:   globalAdmin.ID,
			userRole: "admin",
			setupArticle: func() uint {
				art := testutils.CreateTestArticle(db, testModule.ID, author.ID)
				return art.ID
			},
			expectError: false,
		},
		// 普通用户不能删除
		{
			name:     "Regular user cannot delete",
			userID:   regularUser.ID,
			userRole: "user",
			setupArticle: func() uint {
				art := testutils.CreateTestArticle(db, testModule.ID, author.ID)
				return art.ID
			},
			expectError: true,
			errorMsg:    "permission denied",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articleID := tt.setupArticle()
			
			err := service.DeleteArticle(articleID, tt.userID, tt.userRole)
			
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
					// 验证文章已删除
					// DeleteArticleWithCascade 使用事务，删除操作在事务中执行
					// 由于测试数据库也是事务，删除后在同一事务中查询应该能看到结果
					var art article.Article
					// 先尝试正常查询（应该查不到，因为软删除）
					if err := db.First(&art, articleID).Error; err == nil {
						t.Errorf("Article should not be found with normal query (soft deleted)")
					}
					// 使用 Unscoped 查询包括软删除的记录
					if err := db.Unscoped().First(&art, articleID).Error; err != nil {
						// 如果 Unscoped 也查不到，说明是硬删除（不符合软删除规范）
						t.Errorf("Article not found even with Unscoped (hard delete detected, should use soft delete): %v", err)
					} else if !art.DeletedAt.Valid {
						// 如果记录存在但 DeletedAt 无效，说明删除操作没有正确设置 DeletedAt
						t.Errorf("Article should be soft deleted (DeletedAt should be valid)")
					}
					// 如果 DeletedAt.Valid 为 true，说明软删除成功
				}
			}
		})
	}
}

// TestCreateSubmission_ReviewWorkflow_Integration 集成测试：审核流程
// 测试 Global_Admin 提交需审核 vs Author/Admin/Moderator 直接发布
