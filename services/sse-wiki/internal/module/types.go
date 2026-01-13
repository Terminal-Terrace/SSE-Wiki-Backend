package module

// ModuleTreeNode 模块树节点
type ModuleTreeNode struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	OwnerID     uint             `json:"owner_id"`
	IsModerator bool             `json:"isModerator"`
	Children    []ModuleTreeNode `json:"children"`
	Role        string           `json:"role"` // 当前用户在该模块的角色: owner, admin, moderator, 空字符串表示无权限
}

// CreateModuleRequest 创建模块请求
type CreateModuleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=512"`
	ParentID    *uint   `json:"parent_id"`
}

// UpdateModuleRequest 更新模块请求
type UpdateModuleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=512"`
	ParentID    *uint   `json:"parent_id"`
}

// BreadcrumbNode 面包屑节点
type BreadcrumbNode struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// LockRequest 编辑锁请求
type LockRequest struct {
	Action string `json:"action" binding:"required,oneof=acquire release"`
}

// LockResponse 编辑锁响应
type LockResponse struct {
	Success  bool      `json:"success"`
	LockedBy *UserInfo `json:"locked_by,omitempty"`
	LockedAt string    `json:"locked_at,omitempty"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// AddModeratorRequest 添加协作者请求
type AddModeratorRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin moderator"`
}

// ModeratorInfo 协作者信息
type ModeratorInfo struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

// DeleteModuleResponse 删除模块响应
type DeleteModuleResponse struct {
	DeletedModules int `json:"deleted_modules"`
}
