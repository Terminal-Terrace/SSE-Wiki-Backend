package discussion

import (
	"errors"
	"fmt"
	"time"
	"sort"

	"gorm.io/gorm"
	discussionModel "terminal-terrace/sse-wiki/internal/model/discussion"
)

var (
	ErrDiscussionNotFound = errors.New("讨论区不存在")
	ErrCommentNotFound    = errors.New("评论不存在")
	ErrUnauthorized       = errors.New("无权限执行此操作")
	ErrInvalidParentID    = errors.New("父评论不存在或无效")
	ErrArticleNotFound    = errors.New("文章不存在")
)

// DiscussionService 讨论服务接口
type DiscussionService interface {
	// 获取文章的所有评论（树状结构）
	GetArticleComments(articleID uint) (*CommentsListResponse, error)

	// 创建顶级评论
	CreateComment(articleID uint, userID uint, req *CreateCommentRequest) (*CommentResponse, error)

	// 回复评论
	ReplyComment(parentCommentID uint, userID uint, req *CreateCommentRequest) (*CommentResponse, error)

	// 更新评论
	UpdateComment(commentID uint, userID uint, req *UpdateCommentRequest) (*CommentResponse, error)

	// 删除评论
	DeleteComment(commentID uint, userID uint) error
}

type discussionService struct {
	repo   DiscussionRepository
	db     *gorm.DB
	userService UserService // 用于获取用户信息
}

// UserService 用户服务接口（需要从其他包引入或定义）
type UserService interface {
	GetUserInfo(userID uint) (*UserInfo, error)
}

// NewDiscussionService 创建服务实例
func NewDiscussionService(repo DiscussionRepository, db *gorm.DB, userService UserService) DiscussionService {
	return &discussionService{
		repo:        repo,
		db:          db,
		userService: userService,
	}
}

// GetArticleComments 获取文章的所有评论（树状结构）
func (s *discussionService) GetArticleComments(articleID uint) (*CommentsListResponse, error) {
	if s == nil {
		return &CommentsListResponse{
			Discussion: nil,
			Comments:   []*CommentResponse{},
			Total:      0,
		}, nil
	}

	// 1. 查找讨论区
	discussion, err := s.repo.FindDiscussionByArticleID(articleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文章还没有讨论区，返回空列表
			return &CommentsListResponse{
				Discussion: nil,
				Comments:   []*CommentResponse{},
				Total:      0,
			}, nil
		}
		return nil, err
	}

	// 2. 获取所有评论（扁平列表）
	comments, err := s.repo.FindCommentsByDiscussionID(discussion.ID)
	if err != nil {
		return nil, err
	}

	// 3. 填充用户信息
	commentsWithUser, err := s.enrichCommentsWithUserInfo(comments)
	if err != nil {
		return nil, err
	}

	// 4. 构建树状结构
	commentTree := s.buildCommentTree(commentsWithUser)

	// 5. 返回响应
	return &CommentsListResponse{
		Discussion: ToDiscussionResponse(discussion),
		Comments:   commentTree,
		Total:      len(comments),
	}, nil
}

// CreateComment 创建顶级评论
func (s *discussionService) CreateComment(articleID uint, userID uint, req *CreateCommentRequest) (*CommentResponse, error) {
	if s == nil {
		return nil, errors.New("service is nil")
	}
	
	var comment *discussionModel.DiscussionComment

	// 使用事务：可能需要创建讨论区
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 查找或创建讨论区
		discussion, err := s.repo.FindDiscussionByArticleID(articleID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新的讨论区
				discussion = &discussionModel.Discussion{
					ArticleID:   articleID,
					Title:       fmt.Sprintf("Article %d Discussion", articleID),
					Description: "Article discussion area",
					CreatedBy:   userID,
				}
				if err := s.repo.CreateDiscussion(discussion); err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2. 创建评论
		comment = &discussionModel.DiscussionComment{
			DiscussionID: discussion.ID,
			ParentID:     nil, // 顶级评论
			Content:      req.Content,
			CreatedBy:    userID,
		}

		return s.repo.CreateComment(comment)
	})

	if err != nil {
		return nil, err
	}

	// 3. 获取用户信息
	if s.userService != nil {
		userInfo, err := s.userService.GetUserInfo(userID)
		if err == nil {
			comment.Creator = userInfo
		}
	}

	return ToCommentResponse(comment), nil
}

// ReplyComment 回复评论
func (s *discussionService) ReplyComment(parentCommentID uint, userID uint, req *CreateCommentRequest) (*CommentResponse, error) {
	// 1. 验证父评论是否存在
	parentComment, err := s.repo.FindCommentByID(parentCommentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidParentID
		}
		return nil, err
	}

	// 2. 创建回复评论
	comment := &discussionModel.DiscussionComment{
		DiscussionID: parentComment.DiscussionID,
		ParentID:     &parentCommentID,
		Content:      req.Content,
		CreatedBy:    userID,
	}

	if err := s.repo.CreateComment(comment); err != nil {
		return nil, err
	}

	// 3. 获取用户信息
	if s.userService != nil {
		userInfo, err := s.userService.GetUserInfo(userID)
		if err == nil {
			comment.Creator = userInfo
		}
	}

	return ToCommentResponse(comment), nil
}

// UpdateComment 更新评论
func (s *discussionService) UpdateComment(commentID uint, userID uint, req *UpdateCommentRequest) (*CommentResponse, error) {
	// 1. 查找评论
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	// 2. 权限验证：只能编辑自己的评论
	if comment.CreatedBy != userID {
		return nil, ErrUnauthorized
	}

	// 3. 更新内容
	comment.Content = req.Content
	comment.UpdatedAt = time.Now()

	if err := s.repo.UpdateComment(comment); err != nil {
		return nil, err
	}

	// 4. 获取用户信息
	if s.userService != nil {
		userInfo, err := s.userService.GetUserInfo(userID)
		if err == nil {
			comment.Creator = userInfo
		}
	}

	return ToCommentResponse(comment), nil
}

// DeleteComment 删除评论
func (s *discussionService) DeleteComment(commentID uint, userID uint) error {
	// 1. 查找评论
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	// 2. 权限验证：只能删除自己的评论
	if comment.CreatedBy != userID {
		return ErrUnauthorized
	}

	// 3. 删除评论（软删除）
	return s.repo.DeleteComment(commentID)
}

// ========== 辅助方法 ==========

// enrichCommentsWithUserInfo 为评论列表填充用户信息
func (s *discussionService) enrichCommentsWithUserInfo(comments []discussionModel.DiscussionComment) ([]discussionModel.DiscussionComment, error) {
	if s == nil {
		return comments, nil
	}
	
	if s.userService == nil {
		return comments, nil
	}
	
	if len(comments) == 0 {
		return comments, nil
	}

	// 收集所有唯一的用户ID
	userIDs := make(map[uint]bool)
	for _, comment := range comments {
		userIDs[comment.CreatedBy] = true
	}

	// 批量获取用户信息（这里简化处理，实际应该批量查询）
	userInfoMap := make(map[uint]*UserInfo)
	for userID := range userIDs {
		userInfo, err := s.userService.GetUserInfo(userID)
		if err == nil {
			userInfoMap[userID] = userInfo
		}
	}

	// 填充用户信息
	for i := range comments {
		if userInfo, exists := userInfoMap[comments[i].CreatedBy]; exists {
			comments[i].Creator = userInfo
		}
	}

	return comments, nil
}

// buildCommentTree 将输入的 comments 转为 CommentResponse 的树，但最后会把每个顶级节点的子孙扁平化为直接子节点（按 CreatedAt 升序），
 // 并清空这些子孙的 Replies，以防重复/嵌套渲染。
func (s *discussionService) buildCommentTree(comments []discussionModel.DiscussionComment) []*CommentResponse {
    if len(comments) == 0 {
        return []*CommentResponse{}
    }

    // 1. id -> resp 映射
    commentMap := make(map[uint]*CommentResponse, len(comments))
    for i := range comments {
        resp := ToCommentResponse(&comments[i])
        commentMap[comments[i].ID] = resp
    }

    // 2. 构建父子关系（保留原始嵌套结构）
    for i := range comments {
        c := &comments[i]
        resp := commentMap[c.ID]
        if c.ParentID != nil {
            if parent, ok := commentMap[*c.ParentID]; ok {
                parent.Replies = append(parent.Replies, resp)
                parent.ReplyCount++
            }
        }
    }

    // 3. 收集顶级评论
    var roots []*CommentResponse
    for i := range comments {
        if comments[i].ParentID == nil {
            roots = append(roots, commentMap[comments[i].ID])
        }
    }

    // 4. 对每个顶级，扁平化其所有子孙为一个切片，并按 CreatedAt 升序排序。
    for _, root := range roots {
		// 广度优先或深度优先都可以；这里使用队列（广度优先）或简单的切片遍历均可。
		var queue []*CommentResponse
		queue = append(queue, root.Replies...) 

		var flat []*CommentResponse
		for len(queue) > 0 {
			node := queue[0]
			queue = queue[1:]
			flat = append(flat, node)
			// 将 node 的子节点加入队列，继续采集（保持原始结构遍历）
			queue = append(queue, node.Replies...)
		}

        // 按 CreatedAt 升序排序（假设 CreatedAt 是 time.Time）
        sort.SliceStable(flat, func(i, j int) bool {
            return flat[i].CreatedAt.Before(flat[j].CreatedAt)
        })

        // 清空每个扁平化节点的 Replies，防止前端再次渲染嵌套
        for _, n := range flat {
            n.Replies = nil
        }

        // 将 root 的 Replies 替换为扁平化后的列表，并更新 ReplyCount
        root.Replies = flat
        root.ReplyCount = len(flat)
    }

    return roots
}