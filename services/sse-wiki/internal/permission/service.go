// Package permission 统一权限检查服务
// 提供模块和文章的权限检查功能，支持角色等级比较和权限继承
package permission

import (
	"gorm.io/gorm"
)

// 角色等级常量
// 数值越大权限越高
const (
	RoleLevelOwner     = 100 // owner/author 级别
	RoleLevelAdmin     = 80  // admin 级别
	RoleLevelModerator = 50  // moderator 级别
	RoleLevelUser      = 10  // 普通用户级别
	RoleLevelUnknown   = 0   // 未知角色的 fallback 值（防御性设计）
)

// RoleLevelMap 角色名称到等级的映射
var RoleLevelMap = map[string]int{
	"owner":     RoleLevelOwner,
	"author":    RoleLevelOwner,
	"admin":     RoleLevelAdmin,
	"moderator": RoleLevelModerator,
	"user":      RoleLevelUser,
}

// PermissionSource 权限来源类型
type PermissionSource string

const (
	PermissionSourceGlobal    PermissionSource = "global"    // 全局管理员权限
	PermissionSourceDirect    PermissionSource = "direct"    // 直接权限（owner_id 或协作者表）
	PermissionSourceInherited PermissionSource = "inherited" // 继承权限（从父模块继承）
	PermissionSourceNone      PermissionSource = "none"      // 无权限
)

// PermissionResult 权限检查结果
type PermissionResult struct {
	HasPermission    bool             `json:"has_permission"`    // 是否有权限
	EffectiveRole    string           `json:"effective_role"`    // 有效角色: owner/admin/moderator/user
	PermissionSource PermissionSource `json:"permission_source"` // 权限来源: global/direct/inherited/none
}

// PermissionService 统一权限检查服务
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService 创建权限服务实例
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}


// GetRoleLevel 获取角色的权限等级
// 如果角色不存在于映射中，返回 RoleLevelUnknown 作为 fallback
func GetRoleLevel(role string) int {
	if level, ok := RoleLevelMap[role]; ok {
		return level
	}
	return RoleLevelUnknown
}

// CompareRoles 比较两个角色的权限等级
// 返回值: 正数表示 roleA > roleB，负数表示 roleA < roleB，0 表示相等
func CompareRoles(roleA, roleB string) int {
	return GetRoleLevel(roleA) - GetRoleLevel(roleB)
}

// HasRequiredRole 检查实际角色是否满足所需角色的权限要求
// actualRole: 用户的实际角色（应为有效的角色名称：owner/admin/moderator/user）
// requiredRole: 操作所需的最低角色
//
// 注意：
// - Unknown/empty 角色会被拒绝（返回 false），因为权限等级为 0
// - 未登录用户应在 BFF controller 层拦截，不应传递到 permission 层
// - 此函数假设输入是有效的角色名称，不进行自动标准化
func HasRequiredRole(actualRole, requiredRole string) bool {
	return GetRoleLevel(actualRole) >= GetRoleLevel(requiredRole)
}

// NormalizeRole 标准化角色名称
// 将 "owner" 和 "author" 统一处理，其他角色保持不变
//
// 注意：
// - 此函数用于显示/日志/标准化用途，不是权限检查
// - Unknown/empty 角色会被标准化为 "user"，但这不影响权限检查
// - 权限检查应使用 HasRequiredRole，它不会自动标准化 unknown 角色
func NormalizeRole(role string) string {
	switch role {
	case "owner", "author":
		return "owner"
	case "admin":
		return "admin"
	case "moderator":
		return "moderator"
	default:
		return "user"
	}
}

// IsGlobalAdmin 检查用户是否是全局管理员
// userRole: 来自 JWT 的用户角色
func IsGlobalAdmin(userRole string) bool {
	return userRole == "admin"
}

// NewPermissionResult 创建权限检查结果
func NewPermissionResult(hasPermission bool, effectiveRole string, source PermissionSource) PermissionResult {
	return PermissionResult{
		HasPermission:    hasPermission,
		EffectiveRole:    effectiveRole,
		PermissionSource: source,
	}
}

// NoPermission 返回无权限的结果
func NoPermission() PermissionResult {
	return PermissionResult{
		HasPermission:    false,
		EffectiveRole:    "user",
		PermissionSource: PermissionSourceNone,
	}
}

// GetEffectiveModuleRole 获取用户在模块的有效角色（不含继承）
// 检查顺序：
// 1. Global_Admin（JWT role="admin"）→ 返回 owner 级别
// 2. modules.owner_id == userID → 返回 owner 级别
// 3. module_moderators 表 → 返回对应角色
// 4. 无权限 → 返回 user 级别
func (s *PermissionService) GetEffectiveModuleRole(moduleID uint, userID uint, userRole string) (string, PermissionSource) {
	// 1. 检查是否是全局管理员
	if IsGlobalAdmin(userRole) {
		return "owner", PermissionSourceGlobal
	}

	// 2. 检查是否是模块所有者
	var ownerID uint
	err := s.db.Table("modules").
		Select("owner_id").
		Where("id = ?", moduleID).
		Scan(&ownerID).Error

	if err == nil && ownerID == userID {
		return "owner", PermissionSourceDirect
	}

	// 3. 检查 module_moderators 表
	var role string
	err = s.db.Table("module_moderators").
		Select("role").
		Where("module_id = ? AND user_id = ?", moduleID, userID).
		Scan(&role).Error

	if err == nil && role != "" {
		return role, PermissionSourceDirect
	}

	// 4. 无权限，返回普通用户
	return "user", PermissionSourceNone
}

// GetEffectiveModuleRoleWithInheritance 获取用户在模块的有效角色（含继承）
// 检查顺序：
// 1. Global_Admin（JWT role="admin"）→ 返回 owner 级别
// 2. modules.owner_id == userID → 返回 owner 级别
// 3. module_moderators 表 → 返回对应角色
// 4. 祖先模块权限继承 → 返回继承的角色
// 5. 无权限 → 返回 user 级别
func (s *PermissionService) GetEffectiveModuleRoleWithInheritance(moduleID uint, userID uint, userRole string) (string, PermissionSource) {
	// 1. 先检查直接权限
	role, source := s.GetEffectiveModuleRole(moduleID, userID, userRole)
	if source != PermissionSourceNone {
		return role, source
	}

	// 2. 检查继承权限（从祖先模块）
	ancestorIDs, err := s.getAncestorModuleIDs(moduleID)
	if err != nil {
		return "user", PermissionSourceNone
	}

	for _, ancestorID := range ancestorIDs {
		// 检查是否是祖先模块所有者
		var ancestorOwnerID uint
		err := s.db.Table("modules").
			Select("owner_id").
			Where("id = ?", ancestorID).
			Scan(&ancestorOwnerID).Error
		if err == nil && ancestorOwnerID == userID {
			return "owner", PermissionSourceInherited
		}

		// 检查祖先模块的 module_moderators 表
		var ancestorRole string
		err = s.db.Table("module_moderators").
			Select("role").
			Where("module_id = ? AND user_id = ?", ancestorID, userID).
			Scan(&ancestorRole).Error
		if err == nil && ancestorRole != "" {
			return ancestorRole, PermissionSourceInherited
		}
	}

	// 3. 无权限
	return "user", PermissionSourceNone
}

// getAncestorModuleIDs 获取模块的所有祖先模块ID
func (s *PermissionService) getAncestorModuleIDs(moduleID uint) ([]uint, error) {
	var ids []uint
	err := s.db.Raw(`
		WITH RECURSIVE ancestors AS (
			SELECT id, parent_id FROM modules WHERE id = ?
			UNION ALL
			SELECT m.id, m.parent_id FROM modules m
			INNER JOIN ancestors a ON m.id = a.parent_id
		)
		SELECT id FROM ancestors WHERE id != ?
	`, moduleID, moduleID).Scan(&ids).Error
	return ids, err
}

// CheckModulePermissionWithInheritance 检查用户对模块的权限（含继承）
func (s *PermissionService) CheckModulePermissionWithInheritance(moduleID uint, userID uint, userRole string, requiredRole string) PermissionResult {
	effectiveRole, source := s.GetEffectiveModuleRoleWithInheritance(moduleID, userID, userRole)

	hasPermission := HasRequiredRole(effectiveRole, requiredRole)

	return PermissionResult{
		HasPermission:    hasPermission,
		EffectiveRole:    effectiveRole,
		PermissionSource: source,
	}
}


// GetEffectiveArticleRole 获取用户在文章的有效角色
// 检查顺序：
// 1. articles.created_by == userID → 返回 author 级别
// 2. article_collaborators 表 → 返回对应角色
// 3. Global_Admin 特殊处理：可以提交内容修改（与 user 相同，需要审核），可以删除文章
//    但不能直接发布、编辑基础信息或审核他人的提交
// 4. 无权限 → 返回 user 级别
//
// 注意：Global_Admin 对文章的权限：
// - ✅ 可以提交内容修改（与普通用户相同，需要审核）
// - ✅ 可以删除文章（在 CanDeleteArticle 中单独处理）
// - ❌ 不能直接发布（跳过审核）
// - ❌ 不能编辑基础信息（标题、标签、审核开关）
// - ❌ 不能审核他人的提交
// 因此 GetEffectiveArticleRole 返回 "user" 级别，允许提交修改但不允许其他高权限操作
func (s *PermissionService) GetEffectiveArticleRole(articleID uint, userID uint, userRole string) (string, PermissionSource) {
	// 1. 检查是否是文章作者
	var createdBy uint
	err := s.db.Table("articles").
		Select("created_by").
		Where("id = ?", articleID).
		Scan(&createdBy).Error

	if err == nil && createdBy == userID {
		return "author", PermissionSourceDirect
	}

	// 2. 检查 article_collaborators 表
	var role string
	err = s.db.Table("article_collaborators").
		Select("role").
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Scan(&role).Error

	if err == nil && role != "" {
		return role, PermissionSourceDirect
	}

	// 3. Global_Admin 对文章的权限：
	// - 可以提交内容修改（与 user 相同，需要审核）
	// - 可以删除文章（在 CanDeleteArticle 中单独处理）
	// - 不能直接发布、编辑基础信息或审核他人的提交
	// 因此返回 "user" 级别，允许提交修改但不允许其他高权限操作

	// 4. 无权限，返回普通用户
	return "user", PermissionSourceNone
}

// CanDeleteArticle 检查用户是否可以删除文章
// Global_Admin 或 Author/Admin 可以删除
func (s *PermissionService) CanDeleteArticle(articleID uint, userID uint, userRole string) bool {
	// Global_Admin 可以删除任何文章
	if IsGlobalAdmin(userRole) {
		return true
	}

	// 检查文章角色
	role, _ := s.GetEffectiveArticleRole(articleID, userID, userRole)
	return role == "author" || role == "admin"
}


// CheckModulePermission 检查用户对模块的权限
// moduleID: 模块ID
// userID: 用户ID
// userRole: 用户的全局角色（来自 JWT）
// requiredRole: 操作所需的最低角色
func (s *PermissionService) CheckModulePermission(moduleID uint, userID uint, userRole string, requiredRole string) PermissionResult {
	effectiveRole, source := s.GetEffectiveModuleRole(moduleID, userID, userRole)

	hasPermission := HasRequiredRole(effectiveRole, requiredRole)

	return PermissionResult{
		HasPermission:    hasPermission,
		EffectiveRole:    effectiveRole,
		PermissionSource: source,
	}
}

// CheckArticlePermission 检查用户对文章的权限
// articleID: 文章ID
// userID: 用户ID
// userRole: 用户的全局角色（来自 JWT）
// requiredRole: 操作所需的最低角色
func (s *PermissionService) CheckArticlePermission(articleID uint, userID uint, userRole string, requiredRole string) PermissionResult {
	effectiveRole, source := s.GetEffectiveArticleRole(articleID, userID, userRole)

	hasPermission := HasRequiredRole(effectiveRole, requiredRole)

	return PermissionResult{
		HasPermission:    hasPermission,
		EffectiveRole:    effectiveRole,
		PermissionSource: source,
	}
}
