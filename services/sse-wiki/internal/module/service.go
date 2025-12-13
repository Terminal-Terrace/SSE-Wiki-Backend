package module

import (
	"terminal-terrace/response"
	moduleModel "terminal-terrace/sse-wiki/internal/model/module"

	"gorm.io/gorm"
)

type ModuleService struct {
	moduleRepo *ModuleRepository
}

func NewModuleService(db *gorm.DB) *ModuleService {
	return &ModuleService{
		moduleRepo: NewModuleRepository(db),
	}
}

// GetModuleTree 获取模块树
// userID: 用户ID
// userRole: 用户的全局角色（来自 JWT），"admin" 表示全局管理员
func (s *ModuleService) GetModuleTree(userID uint, userRole string) ([]ModuleTreeNode, error) {
	// 1. 获取所有模块
	modules, err := s.moduleRepo.GetAllModules()
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块列表失败"),
		)
	}

	// 2. 检查是否是全局管理员
	isGlobalAdmin := userRole == "admin"

	// 3. 获取用户在每个模块的角色
	roleMap := make(map[uint]string)
	if !isGlobalAdmin {
		roleMap, err = s.moduleRepo.GetUserModuleRoles(userID)
		if err != nil {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.Fail),
				response.WithErrorMessage("获取协作者信息失败"),
			)
		}
	}

	// 4. 构建树形结构
	tree := s.buildTreeWithRoles(modules, nil, roleMap, isGlobalAdmin, "")
	return tree, nil
}

// buildTreeWithRoles 递归构建树形结构（支持权限继承和角色计算）
// roleMap: 用户在每个模块的直接角色 map[moduleID]role
// isGlobalAdmin: 如果为 true，所有模块的角色都是 "admin"
// parentRole: 父模块的角色，用于权限继承
func (s *ModuleService) buildTreeWithRoles(modules []moduleModel.Module, parentID *uint, roleMap map[uint]string, isGlobalAdmin bool, parentRole string) []ModuleTreeNode {
	var tree []ModuleTreeNode

	for _, module := range modules {
		// 匹配父节点
		if (parentID == nil && module.ParentID == nil) ||
			(parentID != nil && module.ParentID != nil && *parentID == *module.ParentID) {

			// 计算用户在该模块的角色
			var role string
			if isGlobalAdmin {
				// 全局管理员对所有模块都有 admin 权限
				role = "admin"
			} else if directRole, ok := roleMap[module.ID]; ok {
				// 用户在该模块有直接角色
				role = directRole
			} else if parentRole != "" {
				// 从父模块继承角色
				role = parentRole
			}

			// isModerator 为 true 表示用户有任何权限
			isModerator := role != ""

			node := ModuleTreeNode{
				ID:          module.ID,
				Name:        module.ModuleName,
				Description: module.Description,
				OwnerID:     module.OwnerID,
				IsModerator: isModerator,
				Role:        role,
				Children:    s.buildTreeWithRoles(modules, &module.ID, roleMap, isGlobalAdmin, role),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// GetModule 获取单个模块
func (s *ModuleService) GetModule(id uint) (*moduleModel.Module, error) {
	module, err := s.moduleRepo.GetModuleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("模块不存在"),
			)
		}
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块失败"),
		)
	}
	return module, nil
}

// GetBreadcrumbs 获取面包屑导航
func (s *ModuleService) GetBreadcrumbs(id uint) ([]BreadcrumbNode, error) {
	var breadcrumbs []BreadcrumbNode
	currentID := id

	for currentID != 0 {
		module, err := s.moduleRepo.GetModuleByID(currentID)
		if err != nil {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.Fail),
				response.WithErrorMessage("获取面包屑失败"),
			)
		}

		// 前置插入（保证根模块在前）
		breadcrumbs = append([]BreadcrumbNode{{
			ID:   		 module.ID,
			Name: 		 module.ModuleName,
		}}, breadcrumbs...)

		// 向上查找父模块
		if module.ParentID == nil {
			break
		}
		currentID = *module.ParentID
	}

	return breadcrumbs, nil
}

// CreateModule 创建模块
func (s *ModuleService) CreateModule(req CreateModuleRequest, userID uint, userRole string) (*moduleModel.Module, error) {
	// 处理 parent_id=0 的情况，将其视为 null
	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	// 权限检查
	if req.ParentID == nil {
		// 创建顶级模块，需要系统管理员权限
		if userRole != "admin" {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.Forbidden),
				response.WithErrorMessage("只有系统管理员可以创建顶级模块"),
			)
		}
	} else {
		// 创建子模块，需要对父模块有管理权限
		hasPermission, err := s.CheckModulePermission(userID, *req.ParentID, userRole)
		if err != nil {
			return nil, err
		}
		if !hasPermission {
			return nil, response.NewBusinessError(
				response.WithErrorCode(response.Forbidden),
				response.WithErrorMessage("您没有在此模块下创建子模块的权限"),
			)
		}
	}

	// 创建模块
	module := &moduleModel.Module{
		ModuleName:  req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		OwnerID:     userID,
	}

	if err := s.moduleRepo.CreateModule(module); err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("创建模块失败"),
		)
	}

	// 权限继承落库：如果有父模块，复制父模块的所有协作者到子模块
	if req.ParentID != nil {
		_ = s.moduleRepo.CopyParentModeratorsToChild(*req.ParentID, module.ID)
	}

	return module, nil
}

// UpdateModule 更新模块
func (s *ModuleService) UpdateModule(id uint, req UpdateModuleRequest, userID uint, userRole string) error {
	// 记录前端是否传了parent_id（在转换之前记录）
	shouldUpdateParentID := req.ParentID != nil

	// 处理 parent_id=0 的情况，将其视为 null
	if req.ParentID != nil && *req.ParentID == 0 {
		req.ParentID = nil
	}

	// 检查权限
	hasPermission, err := s.CheckModulePermission(userID, id, userRole)
	if err != nil {
		return err
	}
	if !hasPermission {
		return response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("您没有权限修改此模块"),
		)
	}

	// 获取当前模块
	module, err := s.moduleRepo.GetModuleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("模块不存在"),
			)
		}
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块失败"),
		)
	}

	// 判断parent_id是否真的改变了
	needUpdateParentID := false
	if shouldUpdateParentID {
		// 前端传了parent_id，检查是否真的改变了
		if module.ParentID == nil && req.ParentID == nil {
			// 都是null，没有改变
			needUpdateParentID = false
		} else if module.ParentID == nil || req.ParentID == nil {
			// 一个是null，一个不是，改变了
			needUpdateParentID = true
		} else if *module.ParentID != *req.ParentID {
			// 都不是null，但值不同，改变了
			needUpdateParentID = true
		}
	}

	// 如果需要修改 parent_id，进行额外检查
	if needUpdateParentID {
		if req.ParentID == nil {
			// 移动到顶级，需要系统管理员权限
			if userRole != "admin" {
				return response.NewBusinessError(
					response.WithErrorCode(response.Forbidden),
					response.WithErrorMessage("只有系统管理员可以将模块移动到顶级"),
				)
			}
		} else {
			// 移动到其他父模块，检查目标父模块权限
			hasTargetPermission, err := s.CheckModulePermission(userID, *req.ParentID, userRole)
			if err != nil {
				return err
			}
			if !hasTargetPermission {
				return response.NewBusinessError(
					response.WithErrorCode(response.Forbidden),
					response.WithErrorMessage("您没有权限将模块移动到目标位置"),
				)
			}

			// 防止循环引用
			isDescendant, err := s.moduleRepo.CheckIsDescendant(id, *req.ParentID)
			if err != nil {
				return response.NewBusinessError(
					response.WithErrorCode(response.Fail),
					response.WithErrorMessage("检查循环引用失败"),
				)
			}
			if isDescendant {
				return response.NewBusinessError(
					response.WithErrorCode(response.ParseError),
					response.WithErrorMessage("不能将模块移动到其子模块下"),
				)
			}
		}
	}

	// 更新模块
	module.ModuleName  = req.Name
	module.Description = req.Description
	// 只有前端明确传了parent_id时才更新（shouldUpdateParentID标志）
	if shouldUpdateParentID {
		module.ParentID = req.ParentID
	}

	if err := s.moduleRepo.UpdateModule(module); err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("更新模块失败"),
		)
	}

	return nil
}

// DeleteModule 删除模块
func (s *ModuleService) DeleteModule(id uint, userID uint, userRole string) (int64, error) {
	// 获取模块
	module, err := s.moduleRepo.GetModuleByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("模块不存在"),
			)
		}
		return 0, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块失败"),
		)
	}

	// 权限检查：只有所有者或系统管理员可以删除
	if userRole != "admin" && module.OwnerID != userID {
		return 0, response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("只有模块所有者或系统管理员可以删除模块"),
		)
	}

	// 统计将要删除的模块数量
	count, err := s.moduleRepo.CountChildModules(id)
	if err != nil {
		return 0, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("统计子模块失败"),
		)
	}

	// 删除模块（数据库级联删除）
	// 依赖数据库外键 ON DELETE CASCADE 自动删除所有子孙模块
	if err := s.moduleRepo.DeleteModule(id); err != nil {
		return 0, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("删除模块失败"),
		)
	}

	return count, nil
}

// CheckModulePermission 检查用户是否对指定模块有管理权限
// 支持权限继承：如果用户是祖先模块的 owner/moderator，也有权限
func (s *ModuleService) CheckModulePermission(userID, moduleID uint, userRole string) (bool, error) {
	// 1. 系统管理员有所有权限
	if userRole == "admin" {
		return true, nil
	}

	// 2. 使用支持继承的权限检查
	role, _, err := s.moduleRepo.GetUserPermissionWithInheritance(moduleID, userID)
	if err != nil {
		return false, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("检查权限失败"),
		)
	}

	// 有任何角色（owner/admin/moderator）都有权限
	return role != "", nil
}

// GetModerators 获取协作者列表
func (s *ModuleService) GetModerators(moduleID uint, userID uint, userRole string) ([]ModeratorInfo, error) {
	// 权限检查：只有所有者或系统管理员可以查看
	hasPermission, err := s.CheckModulePermission(userID, moduleID, userRole)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("您没有权限查看协作者列表"),
		)
	}

	// 获取协作者列表(包含用户信息)
	moderators, err := s.moduleRepo.GetModeratorsWithUserInfo(moduleID)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取协作者列表失败"),
		)
	}

	var result []ModeratorInfo
	for _, m := range moderators {
		result = append(result, ModeratorInfo{
			UserID:    m.UserID,
			Username:  m.Username,
			Role:      m.Role,
			CreatedAt: m.CreatedAt,
		})
	}

	return result, nil
}

// AddModerator 添加协作者
func (s *ModuleService) AddModerator(moduleID uint, req AddModeratorRequest, userID uint, userRole string) error {
	// 获取模块
	module, err := s.moduleRepo.GetModuleByID(moduleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("模块不存在"),
			)
		}
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块失败"),
		)
	}

	// 权限检查：只有所有者或系统管理员可以添加协作者
	if userRole != "admin" && module.OwnerID != userID {
		return response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("只有模块所有者或系统管理员可以添加协作者"),
		)
	}

	// 防止添加自己为协作者
	if req.UserID == userID {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("不能添加自己为协作者"),
		)
	}

	// 添加协作者到当前模块
	moderator := &moduleModel.ModuleModerator{
		ModuleID: moduleID,
		UserID:   req.UserID,
		Role:     req.Role,
	}

	if err := s.moduleRepo.AddModerator(moderator); err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("添加协作者失败"),
		)
	}

	// 权限继承落库：递归添加协作者到所有子模块
	_ = s.moduleRepo.AddModeratorToDescendants(moduleID, req.UserID, req.Role)

	return nil
}

// RemoveModerator 移除协作者
func (s *ModuleService) RemoveModerator(moduleID, targetUserID uint, userID uint, userRole string) error {
	// 获取模块
	module, err := s.moduleRepo.GetModuleByID(moduleID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NewBusinessError(
				response.WithErrorCode(response.NotFound),
				response.WithErrorMessage("模块不存在"),
			)
		}
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取模块失败"),
		)
	}

	// 不能移除 Owner
	if targetUserID == module.OwnerID {
		return response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("不能移除模块所有者"),
		)
	}

	// 权限检查：只有模块所有者或系统管理员可以移除协作者
	if userRole != "admin" && module.OwnerID != userID {
		return response.NewBusinessError(
			response.WithErrorCode(response.Forbidden),
			response.WithErrorMessage("只有模块所有者或系统管理员可以移除协作者"),
		)
	}

	// 移除协作者
	if err := s.moduleRepo.RemoveModerator(moduleID, targetUserID); err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("移除协作者失败"),
		)
	}

	// 权限继承落库：递归从所有子模块移除协作者
	_ = s.moduleRepo.RemoveModeratorFromDescendants(moduleID, targetUserID)

	return nil
}
