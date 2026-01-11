package discussion

import (
	"testing"

	discussionModel "terminal-terrace/sse-wiki/internal/model/discussion"
	"terminal-terrace/sse-wiki/internal/testutils"
	"gorm.io/gorm"
)

// mockUserService 模拟用户服务
type mockUserService struct {
	users map[uint]*UserInfo
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users: make(map[uint]*UserInfo),
	}
}

func (m *mockUserService) GetUserInfo(userID uint) (*UserInfo, error) {
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	// 返回默认用户信息
	return &UserInfo{
		ID:       userID,
		Username: "user" + string(rune(userID)),
	}, nil
}

func (m *mockUserService) setUser(userID uint, username string) {
	m.users[userID] = &UserInfo{
		ID:       userID,
		Username: username,
	}
}

// setupDiscussionService 创建 DiscussionService 实例用于测试
func setupDiscussionService(t *testing.T) (*discussionService, *gorm.DB, *mockUserService) {
	db := testutils.SetupTestDB(t)
	repo := NewDiscussionRepository(db)
	userService := newMockUserService()
	service := NewDiscussionService(repo, db, userService).(*discussionService)
	return service, db, userService
}

// TestCreateComment_Integration 集成测试：创建评论
func TestCreateComment_Integration(t *testing.T) {
	service, db, userService := setupDiscussionService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	userService.setUser(author.ID, "author")
	commenter := testutils.CreateTestUser(db)
	userService.setUser(commenter.ID, "commenter")

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	tests := []struct {
		name        string
		articleID   uint
		userID      uint
		req         *CreateCommentRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Create comment successfully",
			articleID: testArticle.ID,
			userID:    commenter.ID,
			req: &CreateCommentRequest{
				Content: "This is a test comment",
			},
			expectError: false,
		},
		{
			name:      "Create comment with empty content",
			articleID: testArticle.ID,
			userID:    commenter.ID,
			req: &CreateCommentRequest{
				Content: "",
			},
			expectError: true,
		},
		{
			name:      "Create comment on non-existent article",
			articleID: 99999,
			userID:    commenter.ID,
			req: &CreateCommentRequest{
				Content: "This is a test comment",
			},
			expectError: false, // 会创建讨论区，不会报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := service.CreateComment(tt.articleID, tt.userID, tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if comment == nil {
					t.Errorf("Comment is nil")
				} else {
					// 验证评论已创建
					var dbComment discussionModel.DiscussionComment
					if err := db.First(&dbComment, comment.ID).Error; err != nil {
						t.Errorf("Comment not found in database: %v", err)
					} else {
						if dbComment.Content != tt.req.Content {
							t.Errorf("Comment content = %q, want %q", dbComment.Content, tt.req.Content)
						}
						if dbComment.CreatedBy != tt.userID {
							t.Errorf("Comment CreatedBy = %d, want %d", dbComment.CreatedBy, tt.userID)
						}
						if dbComment.ParentID != nil {
							t.Errorf("Comment ParentID should be nil for top-level comment")
						}
					}
				}
			}
		})
	}
}

// TestReplyComment_Integration 集成测试：回复评论
func TestReplyComment_Integration(t *testing.T) {
	service, db, userService := setupDiscussionService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	userService.setUser(author.ID, "author")
	commenter := testutils.CreateTestUser(db)
	userService.setUser(commenter.ID, "commenter")
	replier := testutils.CreateTestUser(db)
	userService.setUser(replier.ID, "replier")

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建讨论区和顶级评论
	discussion := &discussionModel.Discussion{
		ArticleID:   testArticle.ID,
		Title:       "Test Discussion",
		Description: "Test Description",
		CreatedBy:   author.ID,
	}
	if err := db.Create(discussion).Error; err != nil {
		t.Fatalf("Failed to create discussion: %v", err)
	}

	parentComment := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     nil,
		Content:      "Parent comment",
		CreatedBy:    commenter.ID,
	}
	if err := db.Create(parentComment).Error; err != nil {
		t.Fatalf("Failed to create parent comment: %v", err)
	}

	tests := []struct {
		name           string
		parentCommentID uint
		userID         uint
		req            *CreateCommentRequest
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "Reply comment successfully",
			parentCommentID: parentComment.ID,
			userID:         replier.ID,
			req: &CreateCommentRequest{
				Content: "This is a reply",
			},
			expectError: false,
		},
		{
			name:           "Reply to non-existent comment",
			parentCommentID: 99999,
			userID:         replier.ID,
			req: &CreateCommentRequest{
				Content: "This is a reply",
			},
			expectError: true,
			errorMsg:    "父评论不存在或无效",
		},
		{
			name:           "Reply with empty content",
			parentCommentID: parentComment.ID,
			userID:         replier.ID,
			req: &CreateCommentRequest{
				Content: "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reply, err := service.ReplyComment(tt.parentCommentID, tt.userID, tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if reply == nil {
					t.Errorf("Reply is nil")
				} else {
					// 验证回复已创建
					var dbReply discussionModel.DiscussionComment
					if err := db.First(&dbReply, reply.ID).Error; err != nil {
						t.Errorf("Reply not found in database: %v", err)
					} else {
						if dbReply.Content != tt.req.Content {
							t.Errorf("Reply content = %q, want %q", dbReply.Content, tt.req.Content)
						}
						if dbReply.CreatedBy != tt.userID {
							t.Errorf("Reply CreatedBy = %d, want %d", dbReply.CreatedBy, tt.userID)
						}
						if dbReply.ParentID == nil || *dbReply.ParentID != tt.parentCommentID {
							t.Errorf("Reply ParentID = %v, want %d", dbReply.ParentID, tt.parentCommentID)
						}
						if dbReply.DiscussionID != parentComment.DiscussionID {
							t.Errorf("Reply DiscussionID = %d, want %d", dbReply.DiscussionID, parentComment.DiscussionID)
						}
					}
				}
			}
		})
	}
}

// TestGetArticleComments_Integration 集成测试：获取文章评论
func TestGetArticleComments_Integration(t *testing.T) {
	service, db, userService := setupDiscussionService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	userService.setUser(author.ID, "author")
	commenter1 := testutils.CreateTestUser(db)
	userService.setUser(commenter1.ID, "commenter1")
	commenter2 := testutils.CreateTestUser(db)
	userService.setUser(commenter2.ID, "commenter2")

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建讨论区
	discussion := &discussionModel.Discussion{
		ArticleID:   testArticle.ID,
		Title:       "Test Discussion",
		Description: "Test Description",
		CreatedBy:   author.ID,
	}
	if err := db.Create(discussion).Error; err != nil {
		t.Fatalf("Failed to create discussion: %v", err)
	}

	// 创建顶级评论
	parentComment1 := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     nil,
		Content:      "Parent comment 1",
		CreatedBy:    commenter1.ID,
	}
	if err := db.Create(parentComment1).Error; err != nil {
		t.Fatalf("Failed to create parent comment 1: %v", err)
	}

	parentComment2 := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     nil,
		Content:      "Parent comment 2",
		CreatedBy:    commenter2.ID,
	}
	if err := db.Create(parentComment2).Error; err != nil {
		t.Fatalf("Failed to create parent comment 2: %v", err)
	}

	// 创建回复
	reply1 := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     &parentComment1.ID,
		Content:      "Reply to comment 1",
		CreatedBy:    commenter2.ID,
	}
	if err := db.Create(reply1).Error; err != nil {
		t.Fatalf("Failed to create reply: %v", err)
	}

	tests := []struct {
		name         string
		articleID    uint
		expectCount  int
		expectError  bool
	}{
		{
			name:        "Get comments for article with comments",
			articleID:   testArticle.ID,
			expectCount: 3, // 2 parent + 1 reply
			expectError: false,
		},
		{
			name:        "Get comments for article without discussion",
			articleID:   99999,
			expectCount: 0,
			expectError: false, // 返回空列表，不报错
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetArticleComments(tt.articleID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if result == nil {
					t.Errorf("Result is nil")
				} else {
					if result.Total != tt.expectCount {
						t.Errorf("Expected total count %d, got %d", tt.expectCount, result.Total)
					}
					// 验证树状结构：顶级评论应该包含回复
					if tt.expectCount > 0 {
						topLevelCount := 0
						for _, comment := range result.Comments {
							topLevelCount++
							if comment.ParentID != nil {
								t.Errorf("Top-level comment should have nil ParentID")
							}
						}
						if topLevelCount != 2 {
							t.Errorf("Expected 2 top-level comments, got %d", topLevelCount)
						}
					}
				}
			}
		})
	}
}

// TestUpdateComment_Integration 集成测试：更新评论
func TestUpdateComment_Integration(t *testing.T) {
	service, db, userService := setupDiscussionService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	userService.setUser(author.ID, "author")
	commenter := testutils.CreateTestUser(db)
	userService.setUser(commenter.ID, "commenter")
	otherUser := testutils.CreateTestUser(db)
	userService.setUser(otherUser.ID, "other")

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建讨论区和评论
	discussion := &discussionModel.Discussion{
		ArticleID:   testArticle.ID,
		Title:       "Test Discussion",
		Description: "Test Description",
		CreatedBy:   author.ID,
	}
	if err := db.Create(discussion).Error; err != nil {
		t.Fatalf("Failed to create discussion: %v", err)
	}

	comment := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     nil,
		Content:      "Original comment",
		CreatedBy:    commenter.ID,
	}
	if err := db.Create(comment).Error; err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	tests := []struct {
		name        string
		commentID   uint
		userID      uint
		req         *UpdateCommentRequest
		expectError bool
		errorMsg    string
	}{
		{
			name:      "Update own comment successfully",
			commentID: comment.ID,
			userID:    commenter.ID,
			req: &UpdateCommentRequest{
				Content: "Updated comment",
			},
			expectError: false,
		},
		{
			name:      "Update other user's comment",
			commentID: comment.ID,
			userID:    otherUser.ID,
			req: &UpdateCommentRequest{
				Content: "Updated comment",
			},
			expectError: true,
			errorMsg:   "无权限执行此操作",
		},
		{
			name:      "Update non-existent comment",
			commentID: 99999,
			userID:    commenter.ID,
			req: &UpdateCommentRequest{
				Content: "Updated comment",
			},
			expectError: true,
			errorMsg:   "评论不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedComment, err := service.UpdateComment(tt.commentID, tt.userID, tt.req)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if updatedComment == nil {
					t.Errorf("Updated comment is nil")
				} else {
					// 验证评论已更新
					var dbComment discussionModel.DiscussionComment
					if err := db.First(&dbComment, tt.commentID).Error; err != nil {
						t.Errorf("Comment not found in database: %v", err)
					} else {
						if dbComment.Content != tt.req.Content {
							t.Errorf("Comment content = %q, want %q", dbComment.Content, tt.req.Content)
						}
					}
				}
			}
		})
	}
}

// TestDeleteComment_Integration 集成测试：删除评论
func TestDeleteComment_Integration(t *testing.T) {
	service, db, userService := setupDiscussionService(t)

	// 创建测试数据
	author := testutils.CreateTestUser(db)
	userService.setUser(author.ID, "author")
	commenter := testutils.CreateTestUser(db)
	userService.setUser(commenter.ID, "commenter")
	otherUser := testutils.CreateTestUser(db)
	userService.setUser(otherUser.ID, "other")

	testModule := testutils.CreateTestModule(db, author.ID)
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// 创建讨论区和评论
	discussion := &discussionModel.Discussion{
		ArticleID:   testArticle.ID,
		Title:       "Test Discussion",
		Description: "Test Description",
		CreatedBy:   author.ID,
	}
	if err := db.Create(discussion).Error; err != nil {
		t.Fatalf("Failed to create discussion: %v", err)
	}

	comment := &discussionModel.DiscussionComment{
		DiscussionID: discussion.ID,
		ParentID:     nil,
		Content:      "Comment to delete",
		CreatedBy:    commenter.ID,
	}
	if err := db.Create(comment).Error; err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	tests := []struct {
		name        string
		commentID   uint
		userID      uint
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Delete own comment successfully",
			commentID:   comment.ID,
			userID:      commenter.ID,
			expectError: false,
		},
		{
			name:        "Delete other user's comment",
			commentID:   comment.ID,
			userID:      otherUser.ID,
			expectError: true,
			errorMsg:    "无权限执行此操作",
		},
		{
			name:        "Delete non-existent comment",
			commentID:   99999,
			userID:      commenter.ID,
			expectError: true,
			errorMsg:    "评论不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重新创建评论（如果之前被删除了）
			if tt.commentID == comment.ID {
				var existingComment discussionModel.DiscussionComment
				if err := db.First(&existingComment, comment.ID).Error; err != nil {
					// 评论不存在，重新创建
					newComment := &discussionModel.DiscussionComment{
						DiscussionID: discussion.ID,
						ParentID:     nil,
						Content:      "Comment to delete",
						CreatedBy:    commenter.ID,
					}
					if err := db.Create(newComment).Error; err != nil {
						t.Fatalf("Failed to recreate comment: %v", err)
					}
					tt.commentID = newComment.ID
				}
			}

			err := service.DeleteComment(tt.commentID, tt.userID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证评论已软删除
					var dbComment discussionModel.DiscussionComment
					if err := db.First(&dbComment, tt.commentID).Error; err != nil {
						t.Errorf("Comment not found in database: %v", err)
					} else {
						if !dbComment.IsDeleted {
							t.Errorf("Comment should be marked as deleted")
						}
					}
				}
			}
		})
	}
}

