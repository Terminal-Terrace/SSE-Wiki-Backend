package grpc

import (
	"context"

	"terminal-terrace/sse-wiki/internal/article"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/dto"
	articleModel "terminal-terrace/sse-wiki/internal/model/article"
	pb "terminal-terrace/sse-wiki/protobuf/proto/review_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReviewServiceImpl implements the ReviewService gRPC interface
type ReviewServiceImpl struct {
	pb.UnimplementedReviewServiceServer
}

// NewReviewServiceImpl creates a new ReviewService implementation
func NewReviewServiceImpl() *ReviewServiceImpl {
	return &ReviewServiceImpl{}
}

// getArticleService creates an ArticleService instance
func (s *ReviewServiceImpl) getArticleService() *article.ArticleService {
	articleRepo := article.NewArticleRepository(database.PostgresDB)
	versionRepo := article.NewVersionRepository(database.PostgresDB)
	submissionRepo := article.NewSubmissionRepository(database.PostgresDB)
	tagRepo := article.NewTagRepository(database.PostgresDB)
	mergeService := article.NewMergeService()
	return article.NewArticleService(articleRepo, versionRepo, submissionRepo, tagRepo, mergeService)
}

// GetReviews returns the list of submissions for review
func (s *ReviewServiceImpl) GetReviews(ctx context.Context, req *pb.GetReviewsRequest) (*pb.GetReviewsResponse, error) {
	var articleID *uint
	if req.ArticleId > 0 {
		id := uint(req.ArticleId)
		articleID = &id
	}

	submissions, err := s.getArticleService().GetReviews(req.Status, articleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbSubmissions := make([]*pb.Submission, len(submissions))
	for i := range submissions {
		pbSubmissions[i] = convertReviewSubmissionModel(&submissions[i])
	}

	return &pb.GetReviewsResponse{Submissions: pbSubmissions}, nil
}

// GetReviewDetail returns detailed information for a submission
func (s *ReviewServiceImpl) GetReviewDetail(ctx context.Context, req *pb.GetReviewDetailRequest) (*pb.GetReviewDetailResponse, error) {
	detail, err := s.getArticleService().GetReviewDetail(uint(req.SubmissionId), uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	detailMap := detail

	pbDetail := &pb.ReviewDetail{}

	// Convert submission
	if sub, ok := detailMap["submission"].(map[string]interface{}); ok {
		pbDetail.Submission = &pb.Submission{
			Id:                uint32(getUint(sub, "id")),
			ArticleId:         uint32(getUint(sub, "article_id")),
			ArticleTitle:      getString(sub, "article_title"),
			ProposedVersionId: uint32(getUint(sub, "proposed_version_id")),
			BaseVersionId:     uint32(getUint(sub, "base_version_id")),
			SubmittedBy:       uint32(getUint(sub, "submitted_by")),
			SubmittedByName:   getString(sub, "submitted_by_name"),
			Status:            getString(sub, "status"),
			CommitMessage:     getString(sub, "commit_message"),
			HasConflict:       getBool(sub, "has_conflict"),
			CreatedAt:         getString(sub, "created_at"),
		}
	}

	// Convert proposed version
	if ver, ok := detailMap["proposed_version"].(map[string]interface{}); ok {
		pbDetail.ProposedVersion = &pb.Version{
			Id:            uint32(getUint(ver, "id")),
			ArticleId:     uint32(getUint(ver, "article_id")),
			VersionNumber: int32(getInt(ver, "version_number")),
			Content:       getString(ver, "content"),
			CommitMessage: getString(ver, "commit_message"),
			AuthorId:      uint32(getUint(ver, "author_id")),
			Status:        getString(ver, "status"),
			CreatedAt:     getString(ver, "created_at"),
		}
	}

	// Convert base version
	if ver, ok := detailMap["base_version"].(map[string]interface{}); ok {
		pbDetail.BaseVersion = &pb.Version{
			Id:            uint32(getUint(ver, "id")),
			ArticleId:     uint32(getUint(ver, "article_id")),
			VersionNumber: int32(getInt(ver, "version_number")),
			Content:       getString(ver, "content"),
			CommitMessage: getString(ver, "commit_message"),
			AuthorId:      uint32(getUint(ver, "author_id")),
			Status:        getString(ver, "status"),
			CreatedAt:     getString(ver, "created_at"),
		}
	}

	// Convert article
	if art, ok := detailMap["article"].(map[string]interface{}); ok {
		pbDetail.Article = &pb.Article{
			Id:       uint32(getUint(art, "id")),
			Title:    getString(art, "title"),
			ModuleId: uint32(getUint(art, "module_id")),
		}
	}

	return &pb.GetReviewDetailResponse{Detail: pbDetail}, nil
}

// ReviewAction performs a review action (approve/reject)
func (s *ReviewServiceImpl) ReviewAction(ctx context.Context, req *pb.ReviewActionRequest) (*pb.ReviewActionResponse, error) {
	actionReq := dto.ReviewActionRequest{
		Action: req.Action,
		Notes:  req.Notes,
	}
	if req.MergedContent != "" {
		actionReq.MergedContent = &req.MergedContent
	}

	result, err := s.getArticleService().ReviewSubmission(
		uint(req.SubmissionId), uint(req.ReviewerId), req.UserRole, actionReq,
	)

	if err != nil {
		// Check for merge conflict error
		if conflictErr, ok := err.(*article.MergeConflictError); ok {
			cd := conflictErr.ConflictData
			return &pb.ReviewActionResponse{
				Message: "检测到冲突",
				ConflictData: &pb.ConflictData{
					BaseContent:          getString(cd, "base_content"),
					TheirContent:         getString(cd, "their_content"),
					OurContent:           getString(cd, "our_content"),
					HasConflict:          getBool(cd, "has_conflict"),
					BaseVersionNumber:    int32(getInt(cd, "base_version_number")),
					CurrentVersionNumber: int32(getInt(cd, "current_version_number")),
					SubmitterName:        getString(cd, "submitter_name"),
				},
			}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &pb.ReviewActionResponse{
		Message: getString(result.(map[string]interface{}), "message"),
	}

	// Convert published version if present
	if resultMap, ok := result.(map[string]interface{}); ok {
		if pubVer, ok := resultMap["published_version"].(map[string]interface{}); ok {
			response.PublishedVersion = &pb.Version{
				Id:            uint32(getUint(pubVer, "id")),
				ArticleId:     uint32(getUint(pubVer, "article_id")),
				VersionNumber: int32(getInt(pubVer, "version_number")),
				Content:       getString(pubVer, "content"),
				CommitMessage: getString(pubVer, "commit_message"),
				AuthorId:      uint32(getUint(pubVer, "author_id")),
				Status:        getString(pubVer, "status"),
				CreatedAt:     getString(pubVer, "created_at"),
			}
		}
	}

	return response, nil
}

// convertReviewSubmissionModel converts ReviewSubmission model to proto Submission
func convertReviewSubmissionModel(s *articleModel.ReviewSubmission) *pb.Submission {
	if s == nil {
		return &pb.Submission{}
	}
	pbSub := &pb.Submission{
		Id:                uint32(s.ID),
		ArticleId:         uint32(s.ArticleID),
		ProposedVersionId: uint32(s.ProposedVersionID),
		BaseVersionId:     uint32(s.BaseVersionID),
		SubmittedBy:       uint32(s.SubmittedBy),
		Status:            s.Status,
		HasConflict:       s.HasConflict,
		CreatedAt:         s.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if s.ReviewedBy != nil {
		pbSub.ReviewedBy = uint32(*s.ReviewedBy)
	}
	if s.ReviewedAt != nil {
		pbSub.ReviewedAt = s.ReviewedAt.Format("2006-01-02 15:04:05")
	}
	if s.AIScore != nil {
		pbSub.AiScore = int32(*s.AIScore)
	}
	return pbSub
}
