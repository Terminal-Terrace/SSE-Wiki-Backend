package user

import (
	"terminal-terrace/auth-service/internal/database"
	userModel "terminal-terrace/auth-service/internal/model/user"

	"gorm.io/gorm"
)

// PublicUserInfo 公开用户信息（不含 email）
type PublicUserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库实例
func NewUserRepository() *UserRepository {
	return &UserRepository{db: database.PostgresDB}
}

// SearchUsers 搜索用户（按用户名模糊匹配）
// keyword: 搜索关键词
// excludeUserID: 排除的用户ID（通常是当前用户）
// page: 页码，从1开始
// pageSize: 每页数量
// 返回: 用户列表、总数、错误
func (r *UserRepository) SearchUsers(keyword string, excludeUserID uint, page, pageSize int) ([]PublicUserInfo, int64, error) {
	var users []userModel.User
	var total int64

	// 构建查询
	query := r.db.Model(&userModel.User{})

	// 关键词搜索（用户名模糊匹配）
	if keyword != "" {
		query = query.Where("username ILIKE ?", "%"+keyword+"%")
	}

	// 排除指定用户
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 查询用户列表
	if err := query.Select("id, username, avatar").
		Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// 转换为 PublicUserInfo（不含 email）
	result := make([]PublicUserInfo, len(users))
	for i, u := range users {
		username := ""
		if u.Username != nil {
			username = *u.Username
		}
		avatar := ""
		if u.Avatar != nil {
			avatar = *u.Avatar
		}
		result[i] = PublicUserInfo{
			ID:       uint(u.ID),
			Username: username,
			Avatar:   avatar,
		}
	}

	return result, total, nil
}

// GetUsersByIDs 批量获取用户公开信息
// userIDs: 用户ID列表
// 返回: 用户信息列表、错误
func (r *UserRepository) GetUsersByIDs(userIDs []uint) ([]PublicUserInfo, error) {
	if len(userIDs) == 0 {
		return []PublicUserInfo{}, nil
	}

	var users []userModel.User
	if err := r.db.Select("id, username, avatar").
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// 转换为 PublicUserInfo（不含 email）
	result := make([]PublicUserInfo, len(users))
	for i, u := range users {
		username := ""
		if u.Username != nil {
			username = *u.Username
		}
		avatar := ""
		if u.Avatar != nil {
			avatar = *u.Avatar
		}
		result[i] = PublicUserInfo{
			ID:       uint(u.ID),
			Username: username,
			Avatar:   avatar,
		}
	}

	return result, nil
}

// GetUserByID 根据ID获取单个用户公开信息
func (r *UserRepository) GetUserByID(userID uint) (*PublicUserInfo, error) {
	var user userModel.User
	if err := r.db.Select("id, username, avatar").
		First(&user, userID).Error; err != nil {
		return nil, err
	}

	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	avatar := ""
	if user.Avatar != nil {
		avatar = *user.Avatar
	}

	return &PublicUserInfo{
		ID:       uint(user.ID),
		Username: username,
		Avatar:   avatar,
	}, nil
}
