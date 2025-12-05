package grpc

import (
	"context"

	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/discussion"
	pb "terminal-terrace/sse-wiki/protobuf/proto/ssewiki"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DiscussionServiceImpl implements the DiscussionService gRPC interface
type DiscussionServiceImpl struct {
	pb.UnimplementedDiscussionServiceServer
	discussionService discussion.DiscussionService
}

// NewDiscussionServiceImpl creates a new DiscussionService implementation
func NewDiscussionServiceImpl() *DiscussionServiceImpl {
	repo := discussion.NewDiscussionRepository(database.PostgresDB)
	userService := discussion.NewSimpleUserService(database.PostgresDB)
	svc := discussion.NewDiscussionService(repo, database.PostgresDB, userService)
	return &DiscussionServiceImpl{
		discussionService: svc,
	}
}

// GetArticleComments returns all comments for an article
func (s *DiscussionServiceImpl) GetArticleComments(ctx context.Context, req *pb.GetArticleCommentsRequest) (*pb.GetArticleCommentsResponse, error) {
	result, err := s.discussionService.GetArticleComments(uint(req.ArticleId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbComments := make([]*pb.Comment, len(result.Comments))
	for i, c := range result.Comments {
		pbComments[i] = convertComment(c)
	}

	return &pb.GetArticleCommentsResponse{
		Comments: pbComments,
		Total:    int32(result.Total),
	}, nil
}

// CreateComment creates a new top-level comment
func (s *DiscussionServiceImpl) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CreateCommentResponse, error) {
	createReq := &discussion.CreateCommentRequest{
		Content: req.Content,
	}

	result, err := s.discussionService.CreateComment(uint(req.ArticleId), uint(req.UserId), createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCommentResponse{
		Comment: convertComment(result),
	}, nil
}

// ReplyComment replies to an existing comment
func (s *DiscussionServiceImpl) ReplyComment(ctx context.Context, req *pb.ReplyCommentRequest) (*pb.ReplyCommentResponse, error) {
	createReq := &discussion.CreateCommentRequest{
		Content: req.Content,
	}

	result, err := s.discussionService.ReplyComment(uint(req.CommentId), uint(req.UserId), createReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ReplyCommentResponse{
		Comment: convertComment(result),
	}, nil
}

// UpdateComment updates an existing comment
func (s *DiscussionServiceImpl) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.UpdateCommentResponse, error) {
	updateReq := &discussion.UpdateCommentRequest{
		Content: req.Content,
	}

	result, err := s.discussionService.UpdateComment(uint(req.CommentId), uint(req.UserId), updateReq)
	if err != nil {
		if err == discussion.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if err == discussion.ErrCommentNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateCommentResponse{
		Comment: convertComment(result),
	}, nil
}

// DeleteComment deletes a comment
func (s *DiscussionServiceImpl) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.DeleteCommentResponse, error) {
	err := s.discussionService.DeleteComment(uint(req.CommentId), uint(req.UserId))
	if err != nil {
		if err == discussion.ErrUnauthorized {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		if err == discussion.ErrCommentNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteCommentResponse{}, nil
}

// convertComment converts a CommentResponse to proto Comment
func convertComment(c *discussion.CommentResponse) *pb.Comment {
	if c == nil {
		return nil
	}

	pbComment := &pb.Comment{
		Id:           uint32(c.ID),
		DiscussionId: uint32(c.DiscussionID),
		Content:      c.Content,
		CreatedBy:    uint32(c.CreatedBy),
		CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		IsDeleted:    c.IsDeleted,
		ReplyCount:   int32(c.ReplyCount),
	}

	if c.ParentID != nil {
		pbComment.ParentId = uint32(*c.ParentID)
	}

	if c.Creator != nil {
		pbComment.Creator = &pb.UserInfo{
			Id:       uint32(c.Creator.ID),
			Username: c.Creator.Username,
		}
	}

	if len(c.Replies) > 0 {
		pbComment.Replies = make([]*pb.Comment, len(c.Replies))
		for i, reply := range c.Replies {
			pbComment.Replies[i] = convertComment(reply)
		}
	}

	return pbComment
}
