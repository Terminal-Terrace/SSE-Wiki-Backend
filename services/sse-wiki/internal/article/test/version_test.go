package article_test

import (
	"fmt"
	"testing"
	"time"

	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/testutils"
)
func TestGetVersions_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建多个版本
	versions := []*article.ArticleVersion{}
	for i := 0; i < 5; i++ {
		version := &article.ArticleVersion{
			ArticleID:     testArticle.ID,
			VersionNumber: i + 1,
			Content:       fmt.Sprintf("Content %d", i+1),
			CommitMessage: fmt.Sprintf("Commit %d", i+1),
			AuthorID:      author.ID,
			Status:        "published",
			CreatedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		if i > 0 {
			prevID := versions[i-1].ID
			version.BaseVersionID = &prevID
		}
		db.Create(version)
		versions = append(versions, version)
	}

	// 创建另一个文章
	otherArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)
	otherVersion := &article.ArticleVersion{
		ArticleID:     otherArticle.ID,
		VersionNumber: 1,
		Content:       "Other content",
		CommitMessage: "Other commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(otherVersion)

	tests := []struct {
		name        string
		articleID   uint
		expectCount int
		expectError bool
	}{
		{
			name:        "get all versions",
			articleID:   testArticle.ID,
			expectCount: 5,
			expectError: false,
		},
		{
			name:        "get versions for other article",
			articleID:   otherArticle.ID,
			expectCount: 1,
			expectError: false,
		},
		{
			name:        "non-existent article returns empty",
			articleID:   99999,
			expectCount: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions, err := service.GetVersions(tt.articleID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(versions) != tt.expectCount {
				t.Errorf("Expected %d versions, got %d", tt.expectCount, len(versions))
			}

			// 验证版本按版本号倒序排列
			if len(versions) > 1 {
				for i := 0; i < len(versions)-1; i++ {
					if versions[i].VersionNumber < versions[i+1].VersionNumber {
						t.Errorf("Versions should be ordered by version_number DESC")
					}
				}
			}
		})
	}
}

// TestGetVersionByID_Integration 集成测试：获取特定版本
func TestGetVersionByID_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	version := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Test content",
		CommitMessage: "Test commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(version)

	tests := []struct {
		name        string
		versionID   uint
		expectError bool
		checkFields func(t *testing.T, v *article.ArticleVersion)
	}{
		{
			name:        "get existing version",
			versionID:   version.ID,
			expectError: false,
			checkFields: func(t *testing.T, v *article.ArticleVersion) {
				if v.Content != "Test content" {
					t.Errorf("Expected content 'Test content', got %q", v.Content)
				}
				if v.CommitMessage != "Test commit" {
					t.Errorf("Expected commit message 'Test commit', got %q", v.CommitMessage)
				}
				if v.VersionNumber != 1 {
					t.Errorf("Expected version number 1, got %d", v.VersionNumber)
				}
			},
		},
		{
			name:        "non-existent version returns error",
			versionID:   99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := service.GetVersionByID(tt.versionID)

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
				tt.checkFields(t, v)
			}
		})
	}
}

// TestGetVersionDiff_Integration 集成测试：获取版本差异
func TestGetVersionDiff_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本（v1，无base_version）
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Base content",
		CommitMessage: "Base commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	db.Create(baseVersion)

	// 创建基于v1的版本（v2）
	currentVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       "Current content",
		CommitMessage: "Current commit",
		AuthorID:      author.ID,
		Status:        "published",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	db.Create(currentVersion)

	tests := []struct {
		name        string
		versionID   uint
		expectError bool
		checkFields func(t *testing.T, result map[string]interface{})
	}{
		{
			name:        "get diff for version with base",
			versionID:   currentVersion.ID,
			expectError: false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["current_version"] == nil {
					t.Errorf("Expected current_version")
				}
				if result["base_version"] == nil {
					t.Errorf("Expected base_version")
				}
				base, ok := result["base_version"].(*article.ArticleVersion)
				if !ok {
					t.Fatalf("Expected base_version to be *ArticleVersion")
				}
				if base.Content != "Base content" {
					t.Errorf("Expected base content 'Base content', got %q", base.Content)
				}
			},
		},
		{
			name:        "get diff for initial version (no base)",
			versionID:   baseVersion.ID,
			expectError: false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["base_version"] != nil {
					t.Errorf("Expected base_version to be nil for initial version, got %v", result["base_version"])
				}
			},
		},
		{
			name:        "non-existent version returns error",
			versionID:   99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetVersionDiff(tt.versionID)

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

// TestGetReviews_Integration 集成测试：获取审核列表
func TestGetReviews_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	submitter := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Base content",
		CommitMessage: "Base commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	if err := db.Create(baseVersion).Error; err != nil {
		t.Fatalf("Failed to create base version: %v", err)
	}
	testArticle.CurrentVersionID = &baseVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 创建多个提交
	submissions := []*article.ReviewSubmission{}
	for i := 0; i < 3; i++ {
		proposedVersion := &article.ArticleVersion{
			ArticleID:     testArticle.ID,
			VersionNumber: 2 + i,
			Content:       fmt.Sprintf("Proposed content %d", i+1),
			CommitMessage: fmt.Sprintf("Proposed commit %d", i+1),
			AuthorID:      submitter.ID,
			Status:        "pending",
			BaseVersionID: &baseVersion.ID,
			CreatedAt:     time.Now(),
		}
		if err := db.Create(proposedVersion).Error; err != nil {
			t.Fatalf("Failed to create proposed version: %v", err)
		}

		status := "pending"
		if i == 1 {
			status = "merged"
		}
		submission := &article.ReviewSubmission{
			ArticleID:         testArticle.ID,
			ProposedVersionID: proposedVersion.ID,
			BaseVersionID:     baseVersion.ID,
			SubmittedBy:       submitter.ID,
			Status:            status,
			ProposedTags:      "[]",
			CreatedAt:         time.Now(),
		}
		if err := db.Create(submission).Error; err != nil {
			t.Fatalf("Failed to create submission: %v", err)
		}
		submissions = append(submissions, submission)
	}

	tests := []struct {
		name        string
		status      string
		articleID   *uint
		expectCount int
		expectError bool
	}{
		{
			name:        "get all reviews",
			status:      "all",
			articleID:   nil,
			expectCount: 3,
			expectError: false,
		},
		{
			name:        "get pending reviews",
			status:      "pending",
			articleID:   nil,
			expectCount: 2,
			expectError: false,
		},
		{
			name:        "get merged reviews",
			status:      "merged",
			articleID:   nil,
			expectCount: 1,
			expectError: false,
		},
		{
			name:        "get reviews for specific article",
			status:      "all",
			articleID:   &testArticle.ID,
			expectCount: 3,
			expectError: false,
		},
		{
			name:        "get pending reviews for specific article",
			status:      "pending",
			articleID:   &testArticle.ID,
			expectCount: 2,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reviews, err := service.GetReviews(tt.status, tt.articleID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(reviews) != tt.expectCount {
				t.Errorf("Expected %d reviews, got %d", tt.expectCount, len(reviews))
			}
		})
	}
}

// TestGetReviewDetail_Integration 集成测试：获取审核详情
func TestGetReviewDetail_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	submitter := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       "Base content",
		CommitMessage: "Base commit",
		AuthorID:      author.ID,
		Status:        "published",
		CreatedAt:     time.Now(),
	}
	if err := db.Create(baseVersion).Error; err != nil {
		t.Fatalf("Failed to create base version: %v", err)
	}
	testArticle.CurrentVersionID = &baseVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 创建当前版本（已更新）
	currentVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       "Current content",
		CommitMessage: "Current commit",
		AuthorID:      author.ID,
		Status:        "published",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(currentVersion).Error; err != nil {
		t.Fatalf("Failed to create current version: %v", err)
	}
	testArticle.CurrentVersionID = &currentVersion.ID
	if err := db.Save(testArticle).Error; err != nil {
		t.Fatalf("Failed to update article: %v", err)
	}

	// 创建pending提交
	proposedVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       "Proposed content",
		CommitMessage: "Proposed commit",
		AuthorID:      submitter.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(proposedVersion).Error; err != nil {
		t.Fatalf("Failed to create proposed version: %v", err)
	}

	submission := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: proposedVersion.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       submitter.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submission).Error; err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	tests := []struct {
		name        string
		submissionID uint
		userID      uint
		globalRole  string
		expectError bool
		checkFields func(t *testing.T, result map[string]interface{})
	}{
		{
			name:         "get review detail as author",
			submissionID: submission.ID,
			userID:       author.ID,
			globalRole:   "",
			expectError:  false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["id"] != submission.ID {
					t.Errorf("Expected submission ID %d, got %v", submission.ID, result["id"])
				}
				if result["status"] != "pending" {
					t.Errorf("Expected status 'pending', got %v", result["status"])
				}
				if result["proposed_version"] == nil {
					t.Errorf("Expected proposed_version")
				}
				if result["base_version"] == nil {
					t.Errorf("Expected base_version")
				}
				if result["current_version"] == nil {
					t.Errorf("Expected current_version")
				}
			},
		},
		{
			name:         "get review detail as global admin",
			submissionID: submission.ID,
			userID:       globalAdmin.ID,
			globalRole:   "admin",
			expectError:  false,
			checkFields: func(t *testing.T, result map[string]interface{}) {
				if result["current_user_role"] != "admin" {
					t.Errorf("Expected current_user_role 'admin', got %v", result["current_user_role"])
				}
			},
		},
		{
			name:         "non-existent submission returns error",
			submissionID: 99999,
			userID:       author.ID,
			globalRole:   "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetReviewDetail(tt.submissionID, tt.userID, tt.globalRole)

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


// ============================================================================
// 复杂版本管理场景测试（使用真实HTML文章内容）
// ============================================================================

// TestComplexScenario1_ConcurrentSubmissionChainConflicts 场景1: 并发提交链式冲突
// 描述: 5-6个用户基于同一base版本提交，先审核的会通过，后续的会检测到冲突
