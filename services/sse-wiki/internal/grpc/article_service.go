package grpc

import (
	"context"
	"strconv"
	"time"

	"terminal-terrace/sse-wiki/internal/article"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/dto"
	articleModel "terminal-terrace/sse-wiki/internal/model/article"
	pb "terminal-terrace/sse-wiki/protobuf/proto/article_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ArticleServiceImpl implements the ArticleService gRPC interface
type ArticleServiceImpl struct {
	pb.UnimplementedArticleServiceServer
}

// NewArticleServiceImpl creates a new ArticleService implementation
func NewArticleServiceImpl() *ArticleServiceImpl {
	return &ArticleServiceImpl{}
}

// GetArticlesByModule returns articles in a module with pagination
func (s *ArticleServiceImpl) GetArticlesByModule(ctx context.Context, req *pb.GetArticlesByModuleRequest) (*pb.GetArticlesByModuleResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := s.getArticleService().GetArticlesByModule(uint(req.ModuleId), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Extract data from map[string]interface{}
	total := getInt64(result, "total")
	resultPage := getInt(result, "page")
	resultPageSize := getInt(result, "page_size")
	articlesData := getSlice(result, "articles")

	articles := make([]*pb.ArticleListItem, len(articlesData))
	for i, item := range articlesData {
		if a, ok := item.(map[string]interface{}); ok {
			articles[i] = &pb.ArticleListItem{
				Id:               uint32(getUint(a, "id")),
				Title:            getString(a, "title"),
				ModuleId:         uint32(getUint(a, "module_id")),
				CurrentVersionId: uint32(getUintPtr(a, "current_version_id")),
				ViewCount:        uint32(getUint(a, "view_count")),
				Tags:             getStringSlice(a, "tags"),
				CreatedBy:        uint32(getUint(a, "created_by")),
				CreatedAt:        getString(a, "created_at"),
				UpdatedAt:        getString(a, "updated_at"),
			}
		}
	}

	return &pb.GetArticlesByModuleResponse{
		Total:    total,
		Page:     int32(resultPage),
		PageSize: int32(resultPageSize),
		Articles: articles,
	}, nil
}

// Get User favourites
func (s *ArticleServiceImpl) GetUserArticleFavourites(ctx context.Context, req *pb.GetArticleFavouritesRequest) (*pb.GetArticleFavouritesResponse, error) {
	UserId, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, err
	}
	res, err := s.getArticleService().GetUserFavouriteArticle(uint(UserId))

	if err != nil {
		return nil, err
	}

	return &pb.GetArticleFavouritesResponse{
		Id: res,
	}, nil
}

// GetArticle returns article details
func (s *ArticleServiceImpl) GetArticle(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	result, err := s.getArticleService().GetArticle(uint(req.Id), uint(req.UserId), req.UserRole)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	pbArticle := &pb.Article{
		Id:               uint32(getUint(result, "id")),
		Title:            getString(result, "title"),
		ModuleId:         uint32(getUint(result, "module_id")),
		Content:          getString(result, "content"),
		CommitMessage:    getString(result, "commit_message"),
		VersionNumber:    int32(getInt(result, "version_number")),
		CurrentVersionId: uint32(getUint(result, "current_version_id")),
		CurrentUserRole:  getString(result, "current_user_role"),
		IsReviewRequired: getBool(result, "is_review_required"),
		ViewCount:        uint32(getUint(result, "view_count")),
		Tags:             getStringSlice(result, "tags"),
		CreatedBy:        uint32(getUint(result, "created_by")),
		CreatedAt:        getString(result, "created_at"),
		UpdatedAt:        getString(result, "updated_at"),
	}

	// Convert pending submissions if present
	if submissions := getSlice(result, "pending_submissions"); len(submissions) > 0 {
		pbArticle.PendingSubmissions = make([]*pb.PendingSubmission, len(submissions))
		for i, sub := range submissions {
			if subMap, ok := sub.(map[string]interface{}); ok {
				pbArticle.PendingSubmissions[i] = &pb.PendingSubmission{
					Id:              uint32(getUint(subMap, "id")),
					SubmittedBy:     uint32(getUint(subMap, "submitted_by")),
					SubmittedByName: getString(subMap, "submitted_by_name"),
					Status:          getString(subMap, "status"),
					CreatedAt:       getString(subMap, "created_at"),
				}
			}
		}
	}

	// Convert history entries if present
	if historyEntries := getSlice(result, "history"); len(historyEntries) > 0 {
		pbArticle.History = make([]*pb.HistoryEntry, len(historyEntries))
		for i, entry := range historyEntries {
			if entryMap, ok := entry.(map[string]interface{}); ok {
				pbArticle.History[i] = &pb.HistoryEntry{
					EntryType:              getString(entryMap, "entry_type"),
					EntryId:                uint32(getUint(entryMap, "entry_id")),
					VersionId:              uint32(getUint(entryMap, "version_id")),
					SubmissionId:           uint32(getUint(entryMap, "submission_id")),
					Status:                 getString(entryMap, "status"),
					SubmissionStatus:       getString(entryMap, "submission_status"),
					BaseVersionId:          uint32(getUint(entryMap, "base_version_id")),
					MergedAgainstVersionId: uint32(getUint(entryMap, "merged_against_version_id")),
					HasConflict:            getBool(entryMap, "has_conflict"),
					MergeResult:            getString(entryMap, "merge_result"),
					CommitMessage:          getString(entryMap, "commit_message"),
					AuthorId:               uint32(getUint(entryMap, "author_id")),
					ReviewedBy:             uint32(getUint(entryMap, "reviewed_by")),
					ReviewNotes:            getString(entryMap, "review_notes"),
					CreatedAt:              getString(entryMap, "created_at"),
					ReviewedAt:             getString(entryMap, "reviewed_at"),
				}
			}
		}
	}

	return &pb.GetArticleResponse{Article: pbArticle}, nil
}

// GetVersions returns version history for an article
func (s *ArticleServiceImpl) GetVersions(ctx context.Context, req *pb.GetVersionsRequest) (*pb.GetVersionsResponse, error) {
	versions, err := s.getArticleService().GetVersions(uint(req.ArticleId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbVersions := make([]*pb.Version, len(versions))
	for i, v := range versions {
		pbVersions[i] = convertArticleVersion(&v)
	}

	return &pb.GetVersionsResponse{Versions: pbVersions}, nil
}

// GetVersion returns a specific version
func (s *ArticleServiceImpl) GetVersion(ctx context.Context, req *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	version, err := s.getArticleService().GetVersionByID(uint(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetVersionResponse{Version: convertArticleVersion(version)}, nil
}

// GetVersionDiff returns diff information for a version
func (s *ArticleServiceImpl) GetVersionDiff(ctx context.Context, req *pb.GetVersionDiffRequest) (*pb.GetVersionDiffResponse, error) {
	diffData, err := s.getArticleService().GetVersionDiff(uint(req.VersionId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response := &pb.GetVersionDiffResponse{}

	// 提取当前版本信息
	if currentVersion := diffData["current_version"]; currentVersion != nil {
		if v, ok := currentVersion.(*articleModel.ArticleVersion); ok && v != nil {
			response.CurrentVersion = convertArticleVersion(v)
		}
	}

	// 提取基础版本信息
	if baseVersion := diffData["base_version"]; baseVersion != nil {
		if v, ok := baseVersion.(*articleModel.ArticleVersion); ok && v != nil {
			response.BaseVersion = convertArticleVersion(v)
		}
	}

	return response, nil
}

// CreateArticle creates a new article
func (s *ArticleServiceImpl) CreateArticle(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	isReviewRequired := req.IsReviewRequired
	createReq := dto.CreateArticleRequest{
		Title:            req.Title,
		ModuleID:         uint(req.ModuleId),
		Content:          req.Content,
		CommitMessage:    req.CommitMessage,
		IsReviewRequired: &isReviewRequired,
		Tags:             dto.StringSlice(req.Tags),
	}

	result, err := s.getArticleService().CreateArticle(createReq, uint(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert result to proto Article
	pbArticle := &pb.Article{
		Id:       uint32(getUint(result, "id")),
		Title:    getString(result, "title"),
		ModuleId: uint32(getUint(result, "module_id")),
	}

	return &pb.CreateArticleResponse{Article: pbArticle}, nil
}

// CreateSubmission creates a new submission for an article
func (s *ArticleServiceImpl) CreateSubmission(ctx context.Context, req *pb.CreateSubmissionRequest) (*pb.CreateSubmissionResponse, error) {
	subReq := dto.SubmissionRequest{
		Content:       req.Content,
		CommitMessage: req.CommitMessage,
		BaseVersionID: uint(req.BaseVersionId),
	}

	submission, publishedVersion, err := s.getArticleService().CreateSubmission(
		uint(req.ArticleId), subReq, uint(req.UserId), req.UserRole,
	)

	if err != nil {
		// Check for merge conflict error
		if conflictErr, ok := err.(*article.MergeConflictError); ok {
			cd := conflictErr.ConflictData
			return &pb.CreateSubmissionResponse{
				Published:  false,
				NeedReview: false,
				Message:    "合并冲突",
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

	response := &pb.CreateSubmissionResponse{}

	if submission == nil {
		response.Published = true
		response.NeedReview = false
		response.Message = "修改已发布"
		if publishedVersion != nil {
			response.PublishedVersion = convertVersionFromDTO(publishedVersion)
		}
	} else {
		response.Published = false
		response.NeedReview = true
		response.Message = "提交成功，等待审核"
		response.Submission = convertSubmissionFromDTO(submission)
	}

	return response, nil
}

// UpdateBasicInfo updates article basic information
func (s *ArticleServiceImpl) UpdateBasicInfo(ctx context.Context, req *pb.UpdateBasicInfoRequest) (*pb.UpdateBasicInfoResponse, error) {
	updateReq := dto.UpdateArticleBasicInfoRequest{}

	if req.HasTitle {
		title := req.Title
		updateReq.Title = &title
	}
	if req.HasTags {
		tags := dto.StringSlice(req.Tags)
		updateReq.Tags = &tags
	}
	if req.HasIsReviewRequired {
		isReviewRequired := req.IsReviewRequired
		updateReq.IsReviewRequired = &isReviewRequired
	}

	err := s.getArticleService().UpdateBasicInfo(uint(req.ArticleId), uint(req.UserId), req.UserRole, updateReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateBasicInfoResponse{}, nil
}

// AddCollaborator adds a collaborator to an article
func (s *ArticleServiceImpl) AddCollaborator(ctx context.Context, req *pb.AddCollaboratorRequest) (*pb.AddCollaboratorResponse, error) {
	addReq := dto.AddCollaboratorRequest{
		UserID: uint(req.TargetUserId),
		Role:   req.Role,
	}

	err := s.getArticleService().AddCollaborator(uint(req.ArticleId), uint(req.UserId), req.UserRole, addReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddCollaboratorResponse{}, nil
}

// Helper to get the underlying article service
func (s *ArticleServiceImpl) getArticleService() *article.ArticleService {
	// Access the private field through reflection or expose it
	// For now, we'll create a new service instance
	articleRepo := article.NewArticleRepository(database.PostgresDB)
	versionRepo := article.NewVersionRepository(database.PostgresDB)
	submissionRepo := article.NewSubmissionRepository(database.PostgresDB)
	tagRepo := article.NewTagRepository(database.PostgresDB)
	mergeService := article.NewMergeService()
	return article.NewArticleService(articleRepo, versionRepo, submissionRepo, tagRepo, mergeService)
}

// Helper functions

func ptrToUint(ptr *uint) uint {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getUint(m map[string]interface{}, key string) uint {
	if v, ok := m[key]; ok {
		if v == nil {
			return 0
		}
		switch val := v.(type) {
		case uint:
			return val
		case *uint:
			if val == nil {
				return 0
			}
			return *val
		case uint32:
			return uint(val)
		case uint64:
			return uint(val)
		case int:
			return uint(val)
		case int32:
			return uint(val)
		case int64:
			return uint(val)
		case float64:
			return uint(val)
		}
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int32:
			return int(val)
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case time.Time:
			return val.Format("2006-01-02 15:04:05")
		}
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]string); ok {
		return v
	}
	if v, ok := m[key].([]interface{}); ok {
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

func convertVersion(v interface{}) *pb.Version {
	if vMap, ok := v.(map[string]interface{}); ok {
		return &pb.Version{
			Id:            uint32(getUint(vMap, "id")),
			ArticleId:     uint32(getUint(vMap, "article_id")),
			VersionNumber: int32(getInt(vMap, "version_number")),
			Content:       getString(vMap, "content"),
			CommitMessage: getString(vMap, "commit_message"),
			AuthorId:      uint32(getUint(vMap, "author_id")),
			Status:        getString(vMap, "status"),
			CreatedAt:     getString(vMap, "created_at"),
		}
	}
	return &pb.Version{}
}

func convertVersionFromDTO(v interface{}) *pb.Version {
	if vMap, ok := v.(map[string]interface{}); ok {
		return convertVersion(vMap)
	}
	return &pb.Version{}
}

func convertSubmissionFromDTO(s interface{}) *pb.Submission {
	if sMap, ok := s.(map[string]interface{}); ok {
		return &pb.Submission{
			Id:                uint32(getUint(sMap, "id")),
			ArticleId:         uint32(getUint(sMap, "article_id")),
			ArticleTitle:      getString(sMap, "article_title"),
			ProposedVersionId: uint32(getUint(sMap, "proposed_version_id")),
			BaseVersionId:     uint32(getUint(sMap, "base_version_id")),
			SubmittedBy:       uint32(getUint(sMap, "submitted_by")),
			SubmittedByName:   getString(sMap, "submitted_by_name"),
			Status:            getString(sMap, "status"),
			CommitMessage:     getString(sMap, "commit_message"),
			HasConflict:       getBool(sMap, "has_conflict"),
			CreatedAt:         getString(sMap, "created_at"),
		}
	}
	return &pb.Submission{}
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case int32:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

func getSlice(m map[string]interface{}, key string) []interface{} {
	if v, ok := m[key].([]interface{}); ok {
		return v
	}
	if v, ok := m[key].([]map[string]interface{}); ok {
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	}
	return nil
}

func getUintPtr(m map[string]interface{}, key string) uint {
	if v, ok := m[key]; ok {
		if v == nil {
			return 0
		}
		switch val := v.(type) {
		case *uint:
			if val == nil {
				return 0
			}
			return *val
		case uint:
			return val
		case uint32:
			return uint(val)
		case uint64:
			return uint(val)
		case int:
			return uint(val)
		case int32:
			return uint(val)
		case int64:
			return uint(val)
		case float64:
			return uint(val)
		}
	}
	return 0
}

// convertArticleVersion converts ArticleVersion model to proto Version
func convertArticleVersion(v *articleModel.ArticleVersion) *pb.Version {
	if v == nil {
		return &pb.Version{}
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

// convertReviewSubmission converts ReviewSubmission model to proto Submission
func convertReviewSubmission(s *articleModel.ReviewSubmission) *pb.Submission {
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

// getVersionContent extracts content from ArticleVersion object
func getVersionContent(v interface{}) string {
	// 尝试直接类型断言
	if version, ok := v.(*articleModel.ArticleVersion); ok && version != nil {
		return version.Content
	}
	// 如果是 map 类型（兼容旧逻辑）
	if vMap, ok := v.(map[string]interface{}); ok {
		return getString(vMap, "content")
	}
	return ""
}

// getVersionNumber extracts version number from ArticleVersion object
func getVersionNumber(v interface{}) int {
	// 尝试直接类型断言
	if version, ok := v.(*articleModel.ArticleVersion); ok && version != nil {
		return version.VersionNumber
	}
	// 如果是 map 类型（兼容旧逻辑）
	if vMap, ok := v.(map[string]interface{}); ok {
		return getInt(vMap, "version_number")
	}
	return 0
}
