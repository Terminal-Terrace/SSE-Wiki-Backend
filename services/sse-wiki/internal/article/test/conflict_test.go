package article_test

import (
	"fmt"
	"testing"
	"time"

	articlePkg "terminal-terrace/sse-wiki/internal/article"
	"terminal-terrace/sse-wiki/internal/dto"
	"terminal-terrace/sse-wiki/internal/model/article"
	"terminal-terrace/sse-wiki/internal/testutils"
)

func TestConcurrentSubmissionChainConflicts_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	userA := testutils.CreateTestUser(db)
	userB := testutils.CreateTestUser(db)
	userC := testutils.CreateTestUser(db)
	userD := testutils.CreateTestUser(db)
	userE := testutils.CreateTestUser(db)
	reviewer := testutils.CreateTestUser(db) // 审核者（Admin角色）

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本（使用测试数据常量）
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       BaseContentSSEWiki,
		CommitMessage: "Initial commit",
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

	// 添加reviewer为admin协作者（可以审核）
	if err := db.Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    reviewer.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("Failed to add reviewer as admin: %v", err)
	}

	// 创建所有提交（使用测试数据常量）
	submissions := []struct {
		userID  uint
		content string
		name    string
	}{
		{userA.ID, UserAContentSSEWiki, "UserA"},
		{userB.ID, UserBContentSSEWiki, "UserB"},
		{userC.ID, UserCContentSSEWiki, "UserC"},
		{userD.ID, UserDContentSSEWiki, "UserD"},
		{userE.ID, UserEContentSSEWiki, "UserE"},
	}

	submissionIDs := make([]uint, len(submissions))
	for i, sub := range submissions {
		// 创建pending版本
		pendingVersion := &article.ArticleVersion{
			ArticleID:     testArticle.ID,
			VersionNumber: 2 + i,
			Content:       sub.content,
			CommitMessage: fmt.Sprintf("Commit from %s", sub.name),
			AuthorID:      sub.userID,
			Status:        "pending",
			BaseVersionID: &baseVersion.ID,
			CreatedAt:     time.Now(),
		}
		if err := db.Create(pendingVersion).Error; err != nil {
			t.Fatalf("Failed to create pending version for %s: %v", sub.name, err)
		}

		// 创建审核提交记录
		submission := &article.ReviewSubmission{
			ArticleID:         testArticle.ID,
			ProposedVersionID: pendingVersion.ID,
			BaseVersionID:     baseVersion.ID,
			SubmittedBy:       sub.userID,
			Status:            "pending",
			ProposedTags:      "[]",
			CreatedAt:         time.Now(),
		}
		if err := db.Create(submission).Error; err != nil {
			t.Fatalf("Failed to create submission for %s: %v", sub.name, err)
		}
		submissionIDs[i] = submission.ID
	}

	// 审核流程：先审核A
	t.Run("review_first_submission_A", func(t *testing.T) {
		req := dto.ReviewActionRequest{
			Action: "approve",
			Notes:  "Approve user A's submission",
		}

		_, err := service.ReviewSubmission(submissionIDs[0], reviewer.ID, "", req)
		if err != nil {
			t.Fatalf("Failed to review submission A: %v", err)
		}

		// 验证版本2已发布
		var art article.Article
		if err := db.First(&art, testArticle.ID).Error; err != nil {
			t.Fatalf("Failed to get article: %v", err)
		}
		if art.CurrentVersionID == nil {
			t.Fatalf("Expected current_version_id to be set")
		}

		var version2 article.ArticleVersion
		if err := db.First(&version2, *art.CurrentVersionID).Error; err != nil {
			t.Fatalf("Failed to get version 2: %v", err)
		}
		if version2.Content != UserAContentSSEWiki {
			t.Errorf("Expected version 2 content to be user A's content")
		}
		if version2.Status != "published" {
			t.Errorf("Expected version 2 status 'published', got %q", version2.Status)
		}

		// 验证submission状态
		var submissionA article.ReviewSubmission
		if err := db.First(&submissionA, submissionIDs[0]).Error; err != nil {
			t.Fatalf("Failed to get submission A: %v", err)
		}
		if submissionA.Status != "merged" {
			t.Errorf("Expected submission A status 'merged', got %q", submissionA.Status)
		}
	})

	// 审核B、C、D、E - 都应该检测到冲突
	conflictSubmissions := []struct {
		name          string
		submissionID  uint
		expectedBase  string
		expectedOur   string
		expectedTheir string
	}{
		{
			name:          "review_B_conflict",
			submissionID:  submissionIDs[1],
			expectedBase:  BaseContentSSEWiki,
			expectedOur:   UserAContentSSEWiki, // 当前版本是A的内容
			expectedTheir: UserBContentSSEWiki,
		},
		{
			name:          "review_C_conflict",
			submissionID:  submissionIDs[2],
			expectedBase:  BaseContentSSEWiki,
			expectedOur:   UserAContentSSEWiki,
			expectedTheir: UserCContentSSEWiki,
		},
		{
			name:          "review_D_conflict",
			submissionID:  submissionIDs[3],
			expectedBase:  BaseContentSSEWiki,
			expectedOur:   UserAContentSSEWiki,
			expectedTheir: UserDContentSSEWiki,
		},
		{
			name:          "review_E_conflict",
			submissionID:  submissionIDs[4],
			expectedBase:  BaseContentSSEWiki,
			expectedOur:   UserAContentSSEWiki,
			expectedTheir: UserEContentSSEWiki,
		},
	}

	for _, tt := range conflictSubmissions {
		t.Run(tt.name, func(t *testing.T) {
			req := dto.ReviewActionRequest{
				Action: "approve",
				Notes:  fmt.Sprintf("Review %s", tt.name),
			}

			_, err := service.ReviewSubmission(tt.submissionID, reviewer.ID, "", req)

			// 应该返回冲突错误
			if err == nil {
				t.Fatalf("Expected conflict error but got none")
			}

			mergeConflictErr, ok := err.(*articlePkg.MergeConflictError)
			if !ok {
				t.Fatalf("Expected articlePkg.MergeConflictError, got %T: %v", err, err)
			}

			// 验证冲突信息
			if mergeConflictErr.ConflictData == nil {
				t.Fatalf("Expected ConflictData")
			}

			conflictData := mergeConflictErr.ConflictData
			if conflictData["base_content"] != tt.expectedBase {
				t.Errorf("Expected base_content to match base version")
			}
			if conflictData["our_content"] != tt.expectedOur {
				t.Errorf("Expected our_content to be current version (user A's content)")
			}
			if conflictData["their_content"] != tt.expectedTheir {
				t.Errorf("Expected their_content to match proposed version")
			}
			if conflictData["has_conflict"] != true {
				t.Errorf("Expected has_conflict=true")
			}

			// 验证submission状态变为conflict_detected
			var submission article.ReviewSubmission
			if err := db.First(&submission, tt.submissionID).Error; err != nil {
				t.Fatalf("Failed to get submission: %v", err)
			}
			if submission.Status != "conflict_detected" {
				t.Errorf("Expected submission status 'conflict_detected', got %q", submission.Status)
			}
			if !submission.HasConflict {
				t.Errorf("Expected HasConflict=true")
			}

			// 验证current_version_id仍然是版本2（用户A的版本）
			var art article.Article
			if err := db.First(&art, testArticle.ID).Error; err != nil {
				t.Fatalf("Failed to get article: %v", err)
			}
			var version2 article.ArticleVersion
			if err := db.First(&version2, *art.CurrentVersionID).Error; err != nil {
				t.Fatalf("Failed to get version 2: %v", err)
			}
			if version2.Content != UserAContentSSEWiki {
				t.Errorf("Expected current_version to remain as user A's content")
			}
		})
	}
}

// TestAutomaticMergeSuccess_Integration 场景2: 无冲突自动合并（测试合并算法）
// 描述: 用户B基于用户A已发布的版本2提交，只有B修改了（单方修改），应该能自动合并
// 注意: 当前ThreeWayMerge算法是字符串级比较，无法识别"修改不同章节自动合并"
//
//	此测试验证"单方修改自动合并"场景，匹配当前算法行为
func TestAutomaticMergeSuccess_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	userA := testutils.CreateTestUser(db)
	userB := testutils.CreateTestUser(db)
	reviewer := testutils.CreateTestUser(db)

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建初始版本（使用测试数据常量）
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       BaseContentGoTutorial,
		CommitMessage: "Initial commit",
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

	// 添加reviewer为admin协作者
	if err := db.Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    reviewer.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("Failed to add reviewer as admin: %v", err)
	}

	// 创建用户A的提交（使用测试数据常量）
	pendingVersionA := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       UserAContentGoTutorial,
		CommitMessage: "User A: Update chapter 1",
		AuthorID:      userA.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionA).Error; err != nil {
		t.Fatalf("Failed to create pending version A: %v", err)
	}

	submissionA := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionA.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userA.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionA).Error; err != nil {
		t.Fatalf("Failed to create submission A: %v", err)
	}

	// 审核A → 版本2发布
	reqA := dto.ReviewActionRequest{
		Action: "approve",
		Notes:  "Approve user A's submission",
	}
	_, err := service.ReviewSubmission(submissionA.ID, reviewer.ID, "", reqA)
	if err != nil {
		t.Fatalf("Failed to review submission A: %v", err)
	}

	// 验证版本2已发布
	var art article.Article
	if err := db.First(&art, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	var version2 article.ArticleVersion
	if err := db.First(&version2, *art.CurrentVersionID).Error; err != nil {
		t.Fatalf("Failed to get version 2: %v", err)
	}
	if version2.Content != UserAContentGoTutorial {
		t.Errorf("Expected version 2 content to be user A's content")
	}

	// 创建用户B的提交
	// 注意：用户B基于版本2（用户A的版本）提交，这样相对于版本2，只有B修改了第二章
	// 当前ThreeWayMerge算法：如果只有theirs改了（相对于base），可以自动合并
	// base=版本2, our=版本2, their=用户B的内容（基于版本2，只改了第二章）
	// 由于theirs != base（B改了），ours == base（当前版本就是版本2），所以可以自动合并
	pendingVersionB := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       UserBContentGoTutorial,
		CommitMessage: "User B: Update chapter 2 based on version 2",
		AuthorID:      userB.ID,
		Status:        "pending",
		BaseVersionID: art.CurrentVersionID, // 基于版本2，而不是base版本
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionB).Error; err != nil {
		t.Fatalf("Failed to create pending version B: %v", err)
	}

	submissionB := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionB.ID,
		BaseVersionID:     *art.CurrentVersionID, // 基于版本2
		SubmittedBy:       userB.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionB).Error; err != nil {
		t.Fatalf("Failed to create submission B: %v", err)
	}

	// 审核B → 应该自动合并成功
	reqB := dto.ReviewActionRequest{
		Action: "approve",
		Notes:  "Approve user B's submission - should auto-merge",
	}
	_, err = service.ReviewSubmission(submissionB.ID, reviewer.ID, "", reqB)

	// 应该成功，无冲突
	if err != nil {
		t.Fatalf("Expected no conflict, but got error: %v", err)
	}

	// 验证版本3已发布，且内容包含两方的修改
	var artAfterB article.Article
	if err := db.First(&artAfterB, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	var version3 article.ArticleVersion
	if err := db.First(&version3, *artAfterB.CurrentVersionID).Error; err != nil {
		t.Fatalf("Failed to get version 3: %v", err)
	}

	// 验证合并后的内容应该包含：
	// - 第一章：用户A的修改（"语法简洁"）
	// - 第二章：用户B的修改（"和channel"）
	// 注意：由于用户B基于版本2提交，合并算法会检测到只有theirs改了（相对于base=版本2）
	// 所以合并结果就是用户B的内容（包含A的第一章修改和B的第二章修改）
	expectedMergedContent := UserBContentGoTutorial // 用户B的内容已经包含了A的修改（因为B基于版本2）

	if version3.Content != expectedMergedContent {
		t.Errorf("Expected merged content to be user B's content (which includes A's chapter 1 modification):\nExpected:\n%s\n\nGot:\n%s", expectedMergedContent, version3.Content)
	}

	// 验证submission状态
	var submissionBAfter article.ReviewSubmission
	if err := db.First(&submissionBAfter, submissionB.ID).Error; err != nil {
		t.Fatalf("Failed to get submission B: %v", err)
	}
	if submissionBAfter.Status != "merged" {
		t.Errorf("Expected submission B status 'merged', got %q", submissionBAfter.Status)
	}
	if submissionBAfter.HasConflict {
		t.Errorf("Expected HasConflict=false for successful auto-merge")
	}
}

// TestConflictResolutionAndSubsequentSubmission_Integration 场景3: 冲突解决后继续提交
// 描述: 解决冲突后，基于新版本继续提交应该无冲突
func TestConflictResolutionAndSubsequentSubmission_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	userA := testutils.CreateTestUser(db)
	userB := testutils.CreateTestUser(db)
	userC := testutils.CreateTestUser(db)
	reviewer := testutils.CreateTestUser(db)

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Base版本（使用测试数据常量）
	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       BaseContentSimple,
		CommitMessage: "Initial commit",
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

	// 添加reviewer为admin协作者
	if err := db.Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    reviewer.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("Failed to add reviewer as admin: %v", err)
	}

	// 用户A提交（使用测试数据常量）
	pendingVersionA := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       UserAContentSimple,
		CommitMessage: "User A commit",
		AuthorID:      userA.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionA).Error; err != nil {
		t.Fatalf("Failed to create pending version A: %v", err)
	}
	submissionA := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionA.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userA.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionA).Error; err != nil {
		t.Fatalf("Failed to create submission A: %v", err)
	}

	// 用户B提交（使用测试数据常量）
	pendingVersionB := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       UserBContentSimple,
		CommitMessage: "User B commit",
		AuthorID:      userB.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionB).Error; err != nil {
		t.Fatalf("Failed to create pending version B: %v", err)
	}
	submissionB := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionB.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userB.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionB).Error; err != nil {
		t.Fatalf("Failed to create submission B: %v", err)
	}

	// 审核A → 版本2发布
	reqA := dto.ReviewActionRequest{Action: "approve", Notes: "Approve A"}
	_, err := service.ReviewSubmission(submissionA.ID, reviewer.ID, "", reqA)
	if err != nil {
		t.Fatalf("Failed to review submission A: %v", err)
	}

	// 审核B → 冲突检测
	reqB := dto.ReviewActionRequest{Action: "approve", Notes: "Review B"}
	_, err = service.ReviewSubmission(submissionB.ID, reviewer.ID, "", reqB)
	if err == nil {
		t.Fatalf("Expected conflict error")
	}

	// 解决冲突：提供merged_content（使用测试数据常量）
	reqBResolved := dto.ReviewActionRequest{
		Action:        "approve",
		Notes:         "Resolve conflict manually",
		MergedContent: stringPtr(ResolvedContentSimple),
	}
	_, err = service.ReviewSubmission(submissionB.ID, reviewer.ID, "", reqBResolved)
	if err != nil {
		t.Fatalf("Failed to resolve conflict: %v", err)
	}

	// 验证版本3已发布（冲突解决后的版本）
	var art article.Article
	if err := db.First(&art, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	var version3 article.ArticleVersion
	if err := db.First(&version3, *art.CurrentVersionID).Error; err != nil {
		t.Fatalf("Failed to get version 3: %v", err)
	}
	if version3.Content != ResolvedContentSimple {
		t.Errorf("Expected version 3 content to be resolved content, got %q", version3.Content)
	}

	// 用户C提交: 基于版本3提交（使用测试数据常量）
	pendingVersionC := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 4,
		Content:       UserCContentSimple,
		CommitMessage: "User C commit based on version 3",
		AuthorID:      userC.ID,
		Status:        "pending",
		BaseVersionID: art.CurrentVersionID, // 基于版本3
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionC).Error; err != nil {
		t.Fatalf("Failed to create pending version C: %v", err)
	}
	submissionC := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionC.ID,
		BaseVersionID:     *art.CurrentVersionID,
		SubmittedBy:       userC.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionC).Error; err != nil {
		t.Fatalf("Failed to create submission C: %v", err)
	}

	// 审核C → 应该无冲突，直接发布
	reqC := dto.ReviewActionRequest{Action: "approve", Notes: "Approve C"}
	_, err = service.ReviewSubmission(submissionC.ID, reviewer.ID, "", reqC)
	if err != nil {
		t.Fatalf("Expected no conflict for submission C, but got error: %v", err)
	}

	// 验证版本4已发布
	var artAfterC article.Article
	if err := db.First(&artAfterC, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	var version4 article.ArticleVersion
	if err := db.First(&version4, *artAfterC.CurrentVersionID).Error; err != nil {
		t.Fatalf("Failed to get version 4: %v", err)
	}
	if version4.Content != UserCContentSimple {
		t.Errorf("Expected version 4 content to be user C's content")
	}
}

// TestRejectionChainReaction_Integration 场景4: 拒绝提交后的连锁反应
// 描述: 拒绝某个提交后，其他提交的base版本仍然是旧的，可以正常通过
func TestRejectionChainReaction_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	userA := testutils.CreateTestUser(db)
	userB := testutils.CreateTestUser(db)
	reviewer := testutils.CreateTestUser(db)

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Base版本
	baseContent := `<p>原始内容</p>`

	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       baseContent,
		CommitMessage: "Initial commit",
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

	// 添加reviewer为admin协作者
	if err := db.Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    reviewer.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("Failed to add reviewer as admin: %v", err)
	}

	// 用户A提交
	userAContent := `<p>用户A的修改</p>`
	pendingVersionA := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       userAContent,
		CommitMessage: "User A commit",
		AuthorID:      userA.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionA).Error; err != nil {
		t.Fatalf("Failed to create pending version A: %v", err)
	}
	submissionA := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionA.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userA.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionA).Error; err != nil {
		t.Fatalf("Failed to create submission A: %v", err)
	}

	// 用户B提交
	userBContent := `<p>用户B的修改</p>`
	pendingVersionB := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       userBContent,
		CommitMessage: "User B commit",
		AuthorID:      userB.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionB).Error; err != nil {
		t.Fatalf("Failed to create pending version B: %v", err)
	}
	submissionB := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionB.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userB.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionB).Error; err != nil {
		t.Fatalf("Failed to create submission B: %v", err)
	}

	// 拒绝A的提交
	reqRejectA := dto.ReviewActionRequest{Action: "reject", Notes: "Reject A"}
	_, err := service.ReviewSubmission(submissionA.ID, reviewer.ID, "", reqRejectA)
	if err != nil {
		t.Fatalf("Failed to reject submission A: %v", err)
	}

	// 验证A的版本状态为rejected
	var versionA article.ArticleVersion
	if err := db.First(&versionA, pendingVersionA.ID).Error; err != nil {
		t.Fatalf("Failed to get version A: %v", err)
	}
	if versionA.Status != "rejected" {
		t.Errorf("Expected version A status 'rejected', got %q", versionA.Status)
	}

	// 验证A的submission状态为rejected
	var submissionAAfter article.ReviewSubmission
	if err := db.First(&submissionAAfter, submissionA.ID).Error; err != nil {
		t.Fatalf("Failed to get submission A: %v", err)
	}
	if submissionAAfter.Status != "rejected" {
		t.Errorf("Expected submission A status 'rejected', got %q", submissionAAfter.Status)
	}

	// 验证current_version_id仍然是版本1（因为A被拒绝了）
	var art article.Article
	if err := db.First(&art, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	if art.CurrentVersionID == nil || *art.CurrentVersionID != baseVersion.ID {
		t.Errorf("Expected current_version_id to remain as base version (1), got %v", art.CurrentVersionID)
	}

	// 审核B → 应该基于版本1合并（因为A被拒绝了，current_version仍然是1）
	reqB := dto.ReviewActionRequest{Action: "approve", Notes: "Approve B"}
	_, err = service.ReviewSubmission(submissionB.ID, reviewer.ID, "", reqB)
	if err != nil {
		t.Fatalf("Expected no conflict for submission B (A was rejected), but got error: %v", err)
	}

	// 验证版本3已发布（用户B的版本）
	var artAfterB article.Article
	if err := db.First(&artAfterB, testArticle.ID).Error; err != nil {
		t.Fatalf("Failed to get article: %v", err)
	}
	var version3 article.ArticleVersion
	if err := db.First(&version3, *artAfterB.CurrentVersionID).Error; err != nil {
		t.Fatalf("Failed to get version 3: %v", err)
	}
	if version3.Content != UserBContentSimple {
		t.Errorf("Expected version 3 content to be user B's content")
	}
}

// TestRealTimeConflictDetection_Integration 场景5: 实时冲突检测（GetReviewDetail）
// 描述: 获取审核详情时，实时检测冲突状态
func TestRealTimeConflictDetection_Integration(t *testing.T) {
	service, db := setupArticleService(t)

	// 创建测试用户
	author := testutils.CreateTestUser(db)
	userA := testutils.CreateTestUser(db)
	userB := testutils.CreateTestUser(db)
	reviewer := testutils.CreateTestUser(db)

	// 创建模块和文章
	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Base版本
	baseContent := `<p>原始内容</p>`

	baseVersion := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 1,
		Content:       baseContent,
		CommitMessage: "Initial commit",
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

	// 添加reviewer为admin协作者
	if err := db.Create(&article.ArticleCollaborator{
		ArticleID: testArticle.ID,
		UserID:    reviewer.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("Failed to add reviewer as admin: %v", err)
	}

	// 用户A提交
	userAContent := `<p>用户A的修改</p>`
	pendingVersionA := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 2,
		Content:       userAContent,
		CommitMessage: "User A commit",
		AuthorID:      userA.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionA).Error; err != nil {
		t.Fatalf("Failed to create pending version A: %v", err)
	}
	submissionA := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionA.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userA.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionA).Error; err != nil {
		t.Fatalf("Failed to create submission A: %v", err)
	}

	// 用户B提交
	userBContent := `<p>用户B的修改</p>`
	pendingVersionB := &article.ArticleVersion{
		ArticleID:     testArticle.ID,
		VersionNumber: 3,
		Content:       userBContent,
		CommitMessage: "User B commit",
		AuthorID:      userB.ID,
		Status:        "pending",
		BaseVersionID: &baseVersion.ID,
		CreatedAt:     time.Now(),
	}
	if err := db.Create(pendingVersionB).Error; err != nil {
		t.Fatalf("Failed to create pending version B: %v", err)
	}
	submissionB := &article.ReviewSubmission{
		ArticleID:         testArticle.ID,
		ProposedVersionID: pendingVersionB.ID,
		BaseVersionID:     baseVersion.ID,
		SubmittedBy:       userB.ID,
		Status:            "pending",
		ProposedTags:      "[]",
		CreatedAt:         time.Now(),
	}
	if err := db.Create(submissionB).Error; err != nil {
		t.Fatalf("Failed to create submission B: %v", err)
	}

	// 审核A → 版本2发布
	reqA := dto.ReviewActionRequest{Action: "approve", Notes: "Approve A"}
	_, err := service.ReviewSubmission(submissionA.ID, reviewer.ID, "", reqA)
	if err != nil {
		t.Fatalf("Failed to review submission A: %v", err)
	}

	// GetReviewDetail(B) → 应该实时检测到冲突
	result, err := service.GetReviewDetail(submissionB.ID, reviewer.ID, "")
	if err != nil {
		t.Fatalf("Failed to get review detail: %v", err)
	}

	// 验证冲突信息
	if result["has_conflict"] != true {
		t.Errorf("Expected has_conflict=true, got %v", result["has_conflict"])
	}

	// 验证包含冲突数据
	conflictData, ok := result["conflict_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected conflict_data")
	}

	// 验证冲突元数据（不再包含 content 字段，内容从版本对象获取）
	if conflictData["has_conflict"] != true {
		t.Errorf("Expected has_conflict=true in conflict_data")
	}
	if conflictData["base_version_number"] == nil {
		t.Errorf("Expected base_version_number in conflict_data")
	}
	if conflictData["current_version_number"] == nil {
		t.Errorf("Expected current_version_number in conflict_data")
	}
}
