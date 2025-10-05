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

// GetUserModeratorModuleIDs 获取用户的协作者模块ID列表
func (r *ModuleRepository) GetUserModeratorModuleIDs(userID uint) ([]uint, error) {
	var moderators []moduleModel.ModuleModerator
	err := r.db.Where("user_id = ?", userID).Find(&moderators).Error
	if err != nil {
		return nil, err
	}

	ids := make([]uint, len(moderators))
	for i, m := range moderators {
		ids[i] = m.ModuleID
	}
	return ids, nil
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
func (r *ModuleRepository) GetModeratorsWithUserInfo(moduleID uint) ([]ModeratorWithUser, error) {
	var result []ModeratorWithUser
	err := r.db.Table("module_moderators").
		Select("module_moderators.user_id, auth_users.username, module_moderators.role, module_moderators.created_at").
		Joins("LEFT JOIN auth_users ON auth_users.id = module_moderators.user_id").
		Where("module_moderators.module_id = ?", moduleID).
		Scan(&result).Error
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
