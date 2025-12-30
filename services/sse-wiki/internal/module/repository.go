package module

import (
	moduleModel "terminal-terrace/sse-wiki/internal/model/module"

	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

// GetAllModules 获取所有模块
func (r *ModuleRepository) GetAllModules() ([]moduleModel.Module, error) {
	var modules []moduleModel.Module
	err := r.db.Order("id ASC").Find(&modules).Error
	return modules, err
}

// GetModuleByID 获取单个模块
func (r *ModuleRepository) GetModuleByID(id uint) (*moduleModel.Module, error) {
	var module moduleModel.Module
	err := r.db.First(&module, id).Error
	return &module, err
}

// CreateModule 创建模块
func (r *ModuleRepository) CreateModule(module *moduleModel.Module) error {
	return r.db.Create(module).Error
}

// UpdateModule 更新模块
func (r *ModuleRepository) UpdateModule(module *moduleModel.Module) error {
	return r.db.Save(module).Error
}

// DeleteModule 删除模块（依赖数据库级联删除）
// 注意：需要数据库外键设置了 ON DELETE CASCADE
func (r *ModuleRepository) DeleteModule(id uint) error {
	return r.db.Delete(&moduleModel.Module{}, id).Error
}

// DeleteModuleRecursive 递归删除模块及其所有子孙模块（代码级实现）
// 备用方案：当数据库没有级联删除时使用
func (r *ModuleRepository) DeleteModuleRecursive(id uint) error {
	// 使用递归 CTE 删除所有子孙模块
	return r.db.Exec(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		DELETE FROM modules WHERE id IN (SELECT id FROM module_tree)
	`, id).Error
}

// GetAllChildModuleIDs 获取所有子孙模块ID（用于批量操作）
func (r *ModuleRepository) GetAllChildModuleIDs(id uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		SELECT id FROM module_tree
	`, id).Scan(&ids).Error
	return ids, err
}

// CountChildModules 统计子模块数量（递归）
func (r *ModuleRepository) CountChildModules(id uint) (int64, error) {
	var count int64
	// 使用递归 CTE 查询所有子孙模块
	err := r.db.Raw(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		SELECT COUNT(*) FROM module_tree
	`, id).Scan(&count).Error
	return count, err
}

// GetUserModeratorModuleIDs 获取用户有管理权限的模块ID列表
// 同时查询 modules.owner_id 和 module_moderators 表，使用 UNION 合并结果
// 修复：之前只查询 module_moderators 表，导致 owner 的 is_moderator 返回 false
func (r *ModuleRepository) GetUserModeratorModuleIDs(userID uint) ([]uint, error) {
	var ids []uint
	// 使用 UNION 合并两个来源的模块ID：
	// 1. modules.owner_id = userID（模块所有者）
	// 2. module_moderators.user_id = userID（协作者表）
	err := r.db.Raw(`
		SELECT id FROM modules WHERE owner_id = ?
		UNION
		SELECT module_id FROM module_moderators WHERE user_id = ?
	`, userID, userID).Scan(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// UserModuleRole 用户在模块中的角色信息
type UserModuleRole struct {
	ModuleID uint   `gorm:"column:module_id"`
	Role     string `gorm:"column:role"`
}

// GetUserModuleRoles 获取用户在所有模块的角色
// 返回 map[moduleID]role，role 可能是 "owner", "admin", "moderator"
func (r *ModuleRepository) GetUserModuleRoles(userID uint) (map[uint]string, error) {
	var results []UserModuleRole
	// 查询用户作为 owner 的模块
	err := r.db.Raw(`
		SELECT id as module_id, 'owner' as role FROM modules WHERE owner_id = ?
		UNION ALL
		SELECT module_id, role FROM module_moderators WHERE user_id = ?
	`, userID, userID).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// 转换为 map，owner 优先级最高
	roleMap := make(map[uint]string)
	for _, r := range results {
		existing, ok := roleMap[r.ModuleID]
		if !ok || r.Role == "owner" || (existing != "owner" && r.Role == "admin") {
			roleMap[r.ModuleID] = r.Role
		}
	}
	return roleMap, nil
}

// GetModerators 获取协作者列表
func (r *ModuleRepository) GetModerators(moduleID uint) ([]moduleModel.ModuleModerator, error) {
	var moderators []moduleModel.ModuleModerator
	err := r.db.Where("module_id = ?", moduleID).Find(&moderators).Error
	return moderators, err
}

// ModeratorWithUser 协作者及用户信息
type ModeratorWithUser struct {
	UserID    uint   `gorm:"column:user_id"`
	Username  string `gorm:"column:username"`
	Role      string `gorm:"column:role"`
	CreatedAt string `gorm:"column:created_at"`
}

// GetModeratorsWithUserInfo 获取协作者列表(包含用户信息)
// 同时返回 Owner 和 Moderators，Owner 的 role 为 "owner"
// 注意：排除 module_moderators 中与 owner_id 相同的记录，避免重复
func (r *ModuleRepository) GetModeratorsWithUserInfo(moduleID uint) ([]ModeratorWithUser, error) {
	var result []ModeratorWithUser
	// 使用 UNION 合并 owner 和 moderators
	// 1. 从 modules 表获取 owner
	// 2. 从 module_moderators 表获取协作者（排除 owner）
	err := r.db.Raw(`
		SELECT m.owner_id as user_id, u.username, 'owner' as role, m.created_at::text as created_at
		FROM modules m
		LEFT JOIN auth_users u ON u.id = m.owner_id
		WHERE m.id = ?
		UNION ALL
		SELECT mm.user_id, u.username, mm.role, mm.created_at::text as created_at
		FROM module_moderators mm
		LEFT JOIN auth_users u ON u.id = mm.user_id
		WHERE mm.module_id = ? AND mm.user_id != (SELECT owner_id FROM modules WHERE id = ?)
		ORDER BY role ASC
	`, moduleID, moduleID, moduleID).Scan(&result).Error
	return result, err
}

// AddModerator 添加协作者
func (r *ModuleRepository) AddModerator(moderator *moduleModel.ModuleModerator) error {
	return r.db.Create(moderator).Error
}

// RemoveModerator 删除协作者
func (r *ModuleRepository) RemoveModerator(moduleID, userID uint) error {
	return r.db.Where("module_id = ? AND user_id = ?", moduleID, userID).
		Delete(&moduleModel.ModuleModerator{}).Error
}

// IsModerator 检查是否是协作者
func (r *ModuleRepository) IsModerator(moduleID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&moduleModel.ModuleModerator{}).
		Where("module_id = ? AND user_id = ?", moduleID, userID).
		Count(&count).Error
	return count > 0, err
}

// CheckIsDescendant 检查 potentialDescendantID 是否是 moduleID 的子孙节点
func (r *ModuleRepository) CheckIsDescendant(moduleID, potentialDescendantID uint) (bool, error) {
	var count int64
	err := r.db.Raw(`
		WITH RECURSIVE descendants AS (
			SELECT id, parent_id FROM modules WHERE parent_id = ?
			UNION ALL
			SELECT m.id, m.parent_id FROM modules m
			INNER JOIN descendants d ON m.parent_id = d.id
		)
		SELECT COUNT(*) FROM descendants WHERE id = ?
	`, moduleID, potentialDescendantID).Scan(&count).Error
	return count > 0, err
}

// GetAncestorModuleIDs 获取模块的所有祖先模块ID（从直接父模块到根模块）
// 使用递归 CTE 向上查询父模块链
func (r *ModuleRepository) GetAncestorModuleIDs(moduleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
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

// GetUserPermissionWithInheritance 获取用户在模块的权限
// 由于权限已落库，直接查询 module_moderators 表即可
// 返回值：role（角色名）, inherited（是否继承，现在总是 false）, error
func (r *ModuleRepository) GetUserPermissionWithInheritance(moduleID uint, userID uint) (string, bool, error) {
	// 检查是否是模块所有者
	var ownerID uint
	err := r.db.Table("modules").
		Select("owner_id").
		Where("id = ?", moduleID).
		Scan(&ownerID).Error
	if err == nil && ownerID == userID {
		return "owner", false, nil
	}

	// 检查 module_moderators 表（权限已落库，包含继承的权限）
	var role string
	err = r.db.Table("module_moderators").
		Select("role").
		Where("module_id = ? AND user_id = ?", moduleID, userID).
		Scan(&role).Error
	if err == nil && role != "" {
		return role, false, nil
	}

	// 无权限
	return "", false, nil
}

// CopyParentModeratorsToChild 将父模块的权限复制到子模块（权限继承落库）
// 继承规则：
//   - 父模块 Owner → 子模块 Admin（如果不是子模块 Owner）
//   - 父模块 Admin → 子模块 Admin（如果不是子模块 Owner）
//   - 父模块 Moderator → 子模块 Moderator（如果不是子模块 Owner）
//
// 注意：子模块创建者自动成为 Owner（在 modules.owner_id 中），不需要再添加到 module_moderators
func (r *ModuleRepository) CopyParentModeratorsToChild(parentID, childID uint) error {
	// 获取子模块的 owner_id，用于排除
	var childOwnerID uint
	if err := r.db.Table("modules").Select("owner_id").Where("id = ?", childID).Scan(&childOwnerID).Error; err != nil {
		return err
	}

	// 1. 将父模块的 Owner 添加为子模块的 Admin（排除子模块 Owner）
	if err := r.db.Exec(`
		INSERT INTO module_moderators (module_id, user_id, role, created_at)
		SELECT ?, owner_id, 'admin', NOW()
		FROM modules
		WHERE id = ? AND owner_id != ?
		ON CONFLICT (module_id, user_id) DO NOTHING
	`, childID, parentID, childOwnerID).Error; err != nil {
		return err
	}

	// 2. 复制父模块的协作者（排除子模块 Owner）
	return r.db.Exec(`
		INSERT INTO module_moderators (module_id, user_id, role, created_at)
		SELECT ?, user_id, role, NOW()
		FROM module_moderators
		WHERE module_id = ? AND user_id != ?
		ON CONFLICT (module_id, user_id) DO NOTHING
	`, childID, parentID, childOwnerID).Error
}

// GetAllDescendantModuleIDs 获取所有子孙模块ID（不包含自身）
func (r *ModuleRepository) GetAllDescendantModuleIDs(moduleID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE parent_id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		SELECT id FROM module_tree
	`, moduleID).Scan(&ids).Error
	return ids, err
}

// AddModeratorToDescendants 递归添加协作者到所有子孙模块
// 用于在父模块添加协作者时，同步到所有子模块
func (r *ModuleRepository) AddModeratorToDescendants(moduleID uint, userID uint, role string) error {
	// 使用 CTE 获取所有子孙模块并批量插入
	return r.db.Exec(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE parent_id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		INSERT INTO module_moderators (module_id, user_id, role, created_at)
		SELECT id, ?, ?, NOW()
		FROM module_tree
		ON CONFLICT (module_id, user_id) DO UPDATE SET role = EXCLUDED.role
	`, moduleID, userID, role).Error
}

// RemoveModeratorFromDescendants 递归从所有子孙模块移除协作者
// 用于在父模块移除协作者时，同步到所有子模块
func (r *ModuleRepository) RemoveModeratorFromDescendants(moduleID uint, userID uint) error {
	// 使用 CTE 获取所有子孙模块并批量删除
	return r.db.Exec(`
		WITH RECURSIVE module_tree AS (
			SELECT id FROM modules WHERE parent_id = ?
			UNION ALL
			SELECT m.id FROM modules m
			INNER JOIN module_tree mt ON m.parent_id = mt.id
		)
		DELETE FROM module_moderators
		WHERE module_id IN (SELECT id FROM module_tree) AND user_id = ?
	`, moduleID, userID).Error
}
