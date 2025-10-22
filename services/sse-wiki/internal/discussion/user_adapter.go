package discussion

import (
	"errors"
	"gorm.io/gorm"
)

// ========== 临时用户服务实现 ==========
// 这是一个简化的实现，用于在没有完整用户模块时快速集成
// 后续应该替换为真实的用户服务

// simpleUserService 简单的用户服务实现
type simpleUserService struct {
	db *gorm.DB
}

// NewSimpleUserService 创建简单的用户服务
// 这是一个临时方案，仅用于开发测试
func NewSimpleUserService(db *gorm.DB) UserService {
	return &simpleUserService{db: db}
}

// GetUserInfo 获取用户信息（简化实现）
func (s *simpleUserService) GetUserInfo(userID uint) (*UserInfo, error) {
	// 临时实现：假设你有一个 users 表
	// 根据实际的用户表结构调整
	var user struct {
		ID       uint   `gorm:"column:id"`
		Username string `gorm:"column:username"`
		Avatar   string `gorm:"column:avatar"`
	}

	err := s.db.Table("users").
		Select("id, username, avatar").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果用户不存在，返回一个默认值而不是错误
			return &UserInfo{
				ID:       userID,
				Username: "Unknown User",
				Avatar:   "",
			}, nil
		}
		return nil, err
	}

	return &UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}, nil
}

// ========== 使用说明 ==========
/*
在 route.go 中使用：

// 方式1: 使用简单的用户服务（临时方案）
discussion.SetupDiscussionRoutesWithUserService(apiV1, db, discussion.NewSimpleUserService(db))

// 方式2: 如果已有完整的用户服务
// import "terminal-terrace/sse-wiki/internal/user"
// userService := user.NewUserService(...)
// discussion.SetupDiscussionRoutesWithUserService(apiV1, db, userService)
*/