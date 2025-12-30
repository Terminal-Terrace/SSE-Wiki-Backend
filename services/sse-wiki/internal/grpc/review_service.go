package grpc

import (
	"context"
	"time"

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
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	detail, err := s.getArticleService().GetReviewDetail(uint(req.SubmissionId), uint(user.UserID), user.Role)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	pbDetail := &pb.ReviewDetail{}

	// 服务层返回的是扁平化的 map，submission 信息直接在顶层
	// 构建 Submission 对象
	pbDetail.Submission = &pb.Submission{
		Id:                uint32(getUint(detail, "id")),
		ArticleId:         uint32(getUint(detail, "article_id")),
		ProposedVersionId: uint32(getUint(detail, "proposed_version_id")),
		BaseVersionId:     uint32(getUint(detail, "base_version_id")),
		SubmittedBy:       uint32(getUint(detail, "submitted_by")),
		Status:            getString(detail, "status"),
		HasConflict:       getBool(detail, "has_conflict"),
	}
	// 处理 created_at（可能是 time.Time 类型）
	if createdAt, ok := detail["created_at"].(time.Time); ok {
		pbDetail.Submission.CreatedAt = createdAt.Format("2006-01-02 15:04:05")
	}
	// 处理 reviewed_by（可能是 *uint 类型）
	if reviewedBy, ok := detail["reviewed_by"].(*uint); ok && reviewedBy != nil {
		pbDetail.Submission.ReviewedBy = uint32(*reviewedBy)
	}

	// Convert proposed_version（服务层返回的是 *article.ArticleVersion 结构体）
	if ver, ok := detail["proposed_version"].(*articleModel.ArticleVersion); ok && ver != nil {
		pbDetail.ProposedVersion = convertVersionModelToReviewPb(ver)
	}

	// Convert base_version
	if ver, ok := detail["base_version"].(*articleModel.ArticleVersion); ok && ver != nil {
		pbDetail.BaseVersion = convertVersionModelToReviewPb(ver)
	}

	// Convert current_version（服务层返回的是 current_version，不是 article）
	// 注意：proto 定义中是 article 字段，但服务层返回的是 current_version
	// 我们需要从 article_id 获取文章信息来填充 Article 字段
	if articleID := getUint(detail, "article_id"); articleID > 0 {
		// 获取文章基本信息
		articleRepo := article.NewArticleRepository(database.PostgresDB)
		if art, err := articleRepo.GetByID(uint(articleID)); err == nil {
			pbDetail.Article = &pb.Article{
				Id:       uint32(art.ID),
				Title:    art.Title,
				ModuleId: uint32(art.ModuleID),
			}
			// 服务层已经计算了 current_user_role，直接使用
			if currentUserRole := getString(detail, "current_user_role"); currentUserRole != "" {
				pbDetail.Article.CurrentUserRole = currentUserRole
			}
		}
	}

	return &pb.GetReviewDetailResponse{Detail: pbDetail}, nil
}

// ReviewAction performs a review action (approve/reject)
func (s *ReviewServiceImpl) ReviewAction(ctx context.Context, req *pb.ReviewActionRequest) (*pb.ReviewActionResponse, error) {
	// 从 JWT 获取用户信息
	user := GetUserFromContext(ctx)

	actionReq := dto.ReviewActionRequest{
		Action: req.Action,
		Notes:  req.Notes,
	}
	if req.MergedContent != "" {
		actionReq.MergedContent = &req.MergedContent
	}

	result, err := s.getArticleService().ReviewSubmission(
		uint(req.SubmissionId), uint(user.UserID), user.Role, actionReq,
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
	// 服务层返回的 published_version 是 *article.ArticleVersion 结构体
	if resultMap, ok := result.(map[string]interface{}); ok {
		if pubVer, ok := resultMap["published_version"].(*articleModel.ArticleVersion); ok && pubVer != nil {
			response.PublishedVersion = convertVersionModelToReviewPb(pubVer)
		}
	}

	return response, nil
}

// convertVersionModelToReviewPb converts ArticleVersion model to review_service proto Version
func convertVersionModelToReviewPb(v *articleModel.ArticleVersion) *pb.Version {
	if v == nil {
		return nil
	}
	return &pb.Version{
		Id:            uint32(v.ID),
		ArticleId:     uint32(v.ArticleID),
		VersionNumber: int32(v.VersionNumber),
		Content:       v.Content,
		CommitMessage: v.CommitMessage,
		AuthorId:      uint32(v.AuthorID),
		Status:        v.Status,
		CreatedAt:     v.CreatedAt.Format("2006-01-02 15:04:05"),
	}
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
