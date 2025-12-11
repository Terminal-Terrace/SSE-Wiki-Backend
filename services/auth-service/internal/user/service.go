package user

import (
	"strings"

	"terminal-terrace/response"
)

// UserService 用户服务层
type UserService struct {
	repo *UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{
		repo: NewUserRepository(),
	}
}

// SearchUsersRequest 搜索用户请求
type SearchUsersRequest struct {
	Keyword       string
	ExcludeUserID uint
	Page          int
	PageSize      int
}

// SearchUsersResponse 搜索用户响应
type SearchUsersResponse struct {
	Users []PublicUserInfo
	Total int64
}

// SearchUsers 搜索用户
// 返回 PublicUserInfo（不含 email）
func (s *UserService) SearchUsers(req SearchUsersRequest) (*SearchUsersResponse, *response.BusinessError) {
	// 处理空白关键词
	keyword := strings.TrimSpace(req.Keyword)

	// 默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.repo.SearchUsers(keyword, req.ExcludeUserID, page, pageSize)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("搜索用户失败"),
		)
	}

	return &SearchUsersResponse{
		Users: users,
		Total: total,
	}, nil
}

// GetUsersByIDsRequest 批量获取用户请求
type GetUsersByIDsRequest struct {
	UserIDs []uint
}

// GetUsersByIDsResponse 批量获取用户响应
type GetUsersByIDsResponse struct {
	Users []PublicUserInfo
}

// GetUsersByIDs 批量获取用户公开信息
// 返回 PublicUserInfo（不含 email）
func (s *UserService) GetUsersByIDs(req GetUsersByIDsRequest) (*GetUsersByIDsResponse, *response.BusinessError) {
	if len(req.UserIDs) == 0 {
		return &GetUsersByIDsResponse{Users: []PublicUserInfo{}}, nil
	}

	// 去重
	uniqueIDs := make(map[uint]bool)
	var deduped []uint
	for _, id := range req.UserIDs {
		if !uniqueIDs[id] {
			uniqueIDs[id] = true
			deduped = append(deduped, id)
		}
	}

	users, err := s.repo.GetUsersByIDs(deduped)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取用户信息失败"),
		)
	}

	return &GetUsersByIDsResponse{Users: users}, nil
}

// GetUserByID 获取单个用户公开信息
func (s *UserService) GetUserByID(userID uint) (*PublicUserInfo, *response.BusinessError) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.NotFound),
			response.WithErrorMessage("用户不存在"),
		)
	}
	return user, nil
}

// UserExists 检查用户是否存在
func (s *UserService) UserExists(userID uint) bool {
	_, err := s.repo.GetUserByID(userID)
	return err == nil
}
