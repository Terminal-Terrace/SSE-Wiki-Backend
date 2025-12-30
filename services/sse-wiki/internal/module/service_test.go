package module

import (
	"context"
	"strings"
	"testing"
	"time"

	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/testutils"
	"gorm.io/gorm"
)

// setupModuleService 创建 ModuleService 实例用于测试
func setupModuleService(t *testing.T) (*ModuleService, *gorm.DB) {
	db := testutils.SetupTestDB(t)
	service := NewModuleService(db)
	return service, db
}

// TestGetModuleTree_Integration 集成测试：获取模块树
func TestGetModuleTree_Integration(t *testing.T) {
	service, db := setupModuleService(t)
	
	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	
	// 创建模块树
	// Root
	//   └── Child1
	//         └── Grandchild1
	rootModule := testutils.CreateTestModule(db, owner.ID)
	childModule := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(rootModule.ID))
	grandchildModule := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(childModule.ID))
	
	// 添加 moderator 到 rootModule
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  rootModule.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})
	
	tests := []struct {
		name          string
		userID        uint
		userRole      string
		expectRootCount int // 根节点数量（应该是1）
		expectModerator map[uint]bool // moduleID -> isModerator
	}{
		// Global_Admin 对所有模块都有权限
		{
			name:        "Global_Admin has access to all modules",
			userID:      globalAdmin.ID,
			userRole:    "admin",
			expectRootCount: 1, // 只有1个根节点
			expectModerator: map[uint]bool{
				rootModule.ID:      true,
				childModule.ID:     true,
				grandchildModule.ID: true,
			},
		},
		// Owner 对所有模块都有权限
		{
			name:        "Owner has access to all modules",
			userID:      owner.ID,
			userRole:    "user",
			expectRootCount: 1, // 只有1个根节点
			expectModerator: map[uint]bool{
				rootModule.ID:      true,
				childModule.ID:     true,
				grandchildModule.ID: true,
			},
		},
		// Moderator 通过继承获得子模块权限
		{
			name:        "Moderator inherits permissions to child modules",
			userID:      moderatorUser.ID,
			userRole:    "user",
			expectRootCount: 1, // 只有1个根节点
			expectModerator: map[uint]bool{
				rootModule.ID:      true,
				childModule.ID:     true,  // 继承自父模块
				grandchildModule.ID: true, // 继承自父模块
			},
		},
		// 普通用户无权限
		{
			name:        "Regular user has no access",
			userID:      regularUser.ID,
			userRole:    "user",
			expectRootCount: 1, // 只有1个根节点
			expectModerator: map[uint]bool{
				rootModule.ID:      false,
				childModule.ID:     false,
				grandchildModule.ID: false,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := service.GetModuleTree(tt.userID, tt.userRole)
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(tree) != tt.expectRootCount {
				t.Errorf("Expected %d root modules, got %d", tt.expectRootCount, len(tree))
				return
			}
			
			// 验证 isModerator 字段
			var checkNode func(node ModuleTreeNode)
			checkNode = func(node ModuleTreeNode) {
				if expected, ok := tt.expectModerator[node.ID]; ok {
					if node.IsModerator != expected {
						t.Errorf("Module %d: IsModerator = %v, want %v", node.ID, node.IsModerator, expected)
					}
				}
				for _, child := range node.Children {
					checkNode(child)
				}
			}
			
			for _, root := range tree {
				checkNode(root)
			}
		})
	}
}

// TestCreateModule_Integration 集成测试：创建模块
func TestCreateModule_Integration(t *testing.T) {
	service, db := setupModuleService(t)
	
	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	
	// 创建父模块
	parentModule := testutils.CreateTestModule(db, owner.ID)
	
	// 添加 admin 和 moderator 到父模块
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  parentModule.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  parentModule.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})
	
	tests := []struct {
		name        string
		userID      uint
		userRole    string
		req         CreateModuleRequest
		expectError bool
		errorMsg    string
	}{
		// Global_Admin 可以创建顶级模块
		{
			name:     "Global_Admin can create root module",
			userID:   globalAdmin.ID,
			userRole: "admin",
			req: CreateModuleRequest{
				Name:        "Root Module",
				Description: "Root module description",
				ParentID:    nil,
			},
			expectError: false,
		},
		// Owner 可以创建子模块
		{
			name:     "Owner can create child module",
			userID:   owner.ID,
			userRole: "user",
			req: CreateModuleRequest{
				Name:        "Child Module",
				Description: "Child module description",
				ParentID:    &parentModule.ID,
			},
			expectError: false,
		},
		// Admin 可以创建子模块
		{
			name:     "Admin can create child module",
			userID:   adminUser.ID,
			userRole: "user",
			req: CreateModuleRequest{
				Name:        "Admin Child Module",
				Description: "Admin child module description",
				ParentID:    &parentModule.ID,
			},
			expectError: false,
		},
		// Moderator 可以创建子模块
		{
			name:     "Moderator can create child module",
			userID:   moderatorUser.ID,
			userRole: "user",
			req: CreateModuleRequest{
				Name:        "Moderator Child Module",
				Description: "Moderator child module description",
				ParentID:    &parentModule.ID,
			},
			expectError: false,
		},
		// 普通用户不能创建顶级模块
		{
			name:     "Regular user cannot create root module",
			userID:   regularUser.ID,
			userRole: "user",
			req: CreateModuleRequest{
				Name:        "Root Module",
				Description: "Root module description",
				ParentID:    nil,
			},
			expectError: true,
			errorMsg:    "只有系统管理员可以创建顶级模块",
		},
		// 普通用户不能创建子模块
		{
			name:     "Regular user cannot create child module",
			userID:   regularUser.ID,
			userRole: "user",
			req: CreateModuleRequest{
				Name:        "Child Module",
				Description: "Child module description",
				ParentID:    &parentModule.ID,
			},
			expectError: true,
			errorMsg:    "您没有在此模块下创建子模块的权限",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdModule, err := service.CreateModule(tt.req, tt.userID, tt.userRole)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if createdModule == nil {
					t.Errorf("Created module is nil")
				} else {
					// 验证模块已创建
					var m module.Module
					if err := db.First(&m, createdModule.ID).Error; err != nil {
						t.Errorf("Module not found in database: %v", err)
					}
				}
			}
		})
	}
}

// TestAddModerator_Integration 集成测试：添加协作者
func TestAddModerator_Integration(t *testing.T) {
	service, db := setupModuleService(t)
	
	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	targetUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	
	// 创建模块
	testModule := testutils.CreateTestModule(db, owner.ID)
	
	// 添加 admin 和 moderator 到模块
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  testModule.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  testModule.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})
	
	tests := []struct {
		name        string
		userID      uint
		userRole    string
		targetUserID uint
		targetRole   string
		expectError bool
		errorMsg    string
	}{
		// Owner 可以添加 admin
		{
			name:        "Owner can add admin",
			userID:      owner.ID,
			userRole:    "user",
			targetUserID: targetUser.ID,
			targetRole:   "admin",
			expectError: false,
		},
		// Owner 可以添加 moderator
		{
			name:        "Owner can add moderator",
			userID:      owner.ID,
			userRole:    "user",
			targetUserID: targetUser.ID,
			targetRole:   "moderator",
			expectError: false,
		},
		// Admin 可以添加 moderator（根据 README）
		{
			name:        "Admin can add moderator",
			userID:      adminUser.ID,
			userRole:    "user",
			targetUserID: targetUser.ID,
			targetRole:   "moderator",
			expectError: false,
		},
		// Admin 不能添加 admin
		{
			name:        "Admin cannot add admin",
			userID:      adminUser.ID,
			userRole:    "user",
			targetUserID: targetUser.ID,
			targetRole:   "admin",
			expectError: true,
			errorMsg:    "只有模块所有者可以添加 Admin 协作者",
		},
		// Global_Admin 可以添加 admin
		{
			name:        "Global_Admin can add admin",
			userID:      globalAdmin.ID,
			userRole:    "admin",
			targetUserID: targetUser.ID,
			targetRole:   "admin",
			expectError: false,
		},
		// Moderator 不能添加协作者
		{
			name:        "Moderator cannot add collaborator",
			userID:      moderatorUser.ID,
			userRole:    "user",
			targetUserID: targetUser.ID,
			targetRole:   "moderator",
			expectError: true,
			errorMsg:    "只有模块所有者或 Admin 协作者可以添加协作者",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 清理之前的协作者记录
			db.Table("module_moderators").Where("module_id = ? AND user_id = ?", testModule.ID, tt.targetUserID).Delete(&module.ModuleModerator{})
			
			req := AddModeratorRequest{
				UserID: tt.targetUserID,
				Role:   tt.targetRole,
			}
			err := service.AddModerator(testModule.ID, req, tt.userID, tt.userRole)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证协作者已添加
					var moderator module.ModuleModerator
					if err := db.Table("module_moderators").Where("module_id = ? AND user_id = ?", testModule.ID, tt.targetUserID).First(&moderator).Error; err != nil {
						t.Errorf("Moderator not found after adding: %v", err)
					} else if moderator.Role != tt.targetRole {
						t.Errorf("Moderator role = %q, want %q", moderator.Role, tt.targetRole)
					}
				}
			}
		})
	}
}

// TestRemoveModerator_Integration 集成测试：移除协作者
func TestRemoveModerator_Integration(t *testing.T) {
	service, db := setupModuleService(t)
	
	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	targetUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	
	// 创建模块
	testModule := testutils.CreateTestModule(db, owner.ID)
	
	tests := []struct {
		name        string
		userID      uint
		userRole    string
		targetUserID uint
		setupModerator func() // 设置协作者
		expectError bool
		errorMsg    string
	}{
		// Owner 可以移除协作者
		{
			name:     "Owner can remove moderator",
			userID:   owner.ID,
			userRole: "user",
			targetUserID: targetUser.ID,
			setupModerator: func() {
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    targetUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
			},
			expectError: false,
		},
		// Admin 可以移除 moderator（根据 README）
		{
			name:     "Admin can remove moderator",
			userID:   adminUser.ID,
			userRole: "user",
			targetUserID: targetUser.ID,
			setupModerator: func() {
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    adminUser.ID,
					Role:      "admin",
					CreatedAt: time.Now(),
				})
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    targetUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
			},
			expectError: false,
		},
		// Global_Admin 可以移除协作者
		{
			name:     "Global_Admin can remove moderator",
			userID:   globalAdmin.ID,
			userRole: "admin",
			targetUserID: targetUser.ID,
			setupModerator: func() {
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    targetUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
			},
			expectError: false,
		},
		// Moderator 不能移除协作者
		{
			name:     "Moderator cannot remove collaborator",
			userID:   moderatorUser.ID,
			userRole: "user",
			targetUserID: targetUser.ID,
			setupModerator: func() {
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    moderatorUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
				db.Table("module_moderators").FirstOrCreate(&module.ModuleModerator{
					ModuleID:  testModule.ID,
					UserID:    targetUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
			},
			expectError: true,
			errorMsg:    "只有模块所有者或 Admin 协作者可以移除协作者",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 设置协作者
			tt.setupModerator()
			
			err := service.RemoveModerator(testModule.ID, tt.targetUserID, tt.userID, tt.userRole)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证协作者已移除
					var moderator module.ModuleModerator
					if err := db.Table("module_moderators").Where("module_id = ? AND user_id = ?", testModule.ID, tt.targetUserID).First(&moderator).Error; err == nil {
						t.Errorf("Moderator still exists after removal")
					}
				}
			}
		})
	}
}

// TestUpdateModule_MoveModule_Integration 集成测试：移动模块权限
func TestUpdateModule_MoveModule_Integration(t *testing.T) {
	service, db := setupModuleService(t)

	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// 创建模块树
	// Root1
	//   └── Child1 (要移动的模块)
	// Root2 (目标父模块)
	root1 := testutils.CreateTestModule(db, owner.ID)
	child1 := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(root1.ID))
	root2 := testutils.CreateTestModule(db, owner.ID)

	// 添加 admin 和 moderator 到 child1
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  child1.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  child1.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})
	// 添加 admin 到 root2（目标父模块），以便 Admin 可以移动 child1 到 root2
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  root2.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name        string
		userID      uint
		userRole    string
		targetParentID *uint
		expectError bool
		errorMsg    string
	}{
		// Owner 可以移动模块
		{
			name:        "Owner can move module",
			userID:      owner.ID,
			userRole:    "user",
			targetParentID: &root2.ID,
			expectError: false,
		},
		// Admin 可以移动模块
		{
			name:        "Admin can move module",
			userID:      adminUser.ID,
			userRole:    "user",
			targetParentID: &root2.ID,
			expectError: false,
		},
		// Global_Admin 可以移动模块
		{
			name:        "Global_Admin can move module",
			userID:      globalAdmin.ID,
			userRole:    "admin",
			targetParentID: &root2.ID,
			expectError: false,
		},
		// Moderator 不能移动模块（根据 README）
		{
			name:        "Moderator cannot move module",
			userID:      moderatorUser.ID,
			userRole:    "user",
			targetParentID: &root2.ID,
			expectError: true,
			errorMsg:    "您没有权限将模块移动到目标位置",
		},
		// 普通用户不能移动模块
		{
			name:        "Regular user cannot move module",
			userID:      regularUser.ID,
			userRole:    "user",
			targetParentID: &root2.ID,
			expectError: true,
			errorMsg:    "您没有权限修改此模块",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置 child1 的 parent_id
			db.Model(&module.Module{}).Where("id = ?", child1.ID).Update("parent_id", root1.ID)

			req := UpdateModuleRequest{
				Name:        child1.ModuleName,
				Description: child1.Description,
				ParentID:    tt.targetParentID,
			}
			err := service.UpdateModule(child1.ID, req, tt.userID, tt.userRole)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证模块已移动
					var updatedModule module.Module
					if err := db.First(&updatedModule, child1.ID).Error; err != nil {
						t.Errorf("Failed to get module: %v", err)
					} else {
						if tt.targetParentID == nil {
							if updatedModule.ParentID != nil {
								t.Errorf("Expected ParentID to be nil, got %v", updatedModule.ParentID)
							}
						} else if updatedModule.ParentID == nil || *updatedModule.ParentID != *tt.targetParentID {
							t.Errorf("Expected ParentID to be %d, got %v", *tt.targetParentID, updatedModule.ParentID)
						}
					}
				}
			}
		})
	}
}

// TestDeleteModule_Integration 集成测试：删除模块权限
func TestDeleteModule_Integration(t *testing.T) {
	service, db := setupModuleService(t)

	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	tests := []struct {
		name        string
		userID      uint
		userRole    string
		setupModule func() uint // 返回要删除的模块ID
		expectError bool
		errorMsg    string
	}{
		// Owner 可以删除模块
		{
			name:     "Owner can delete module",
			userID:   owner.ID,
			userRole: "user",
			setupModule: func() uint {
				mod := testutils.CreateTestModule(db, owner.ID)
				return mod.ID
			},
			expectError: false,
		},
		// Global_Admin 可以删除模块
		{
			name:     "Global_Admin can delete module",
			userID:   globalAdmin.ID,
			userRole: "admin",
			setupModule: func() uint {
				mod := testutils.CreateTestModule(db, owner.ID)
				return mod.ID
			},
			expectError: false,
		},
		// Admin 协作者不能删除模块（根据 README，只有 Owner 和 Global_Admin 可以删除）
		{
			name:     "Admin collaborator cannot delete module",
			userID:   adminUser.ID,
			userRole: "user",
			setupModule: func() uint {
				mod := testutils.CreateTestModule(db, owner.ID)
				db.Table("module_moderators").Create(&module.ModuleModerator{
					ModuleID:  mod.ID,
					UserID:    adminUser.ID,
					Role:      "admin",
					CreatedAt: time.Now(),
				})
				return mod.ID
			},
			expectError: true,
			errorMsg:    "只有模块所有者或系统管理员可以删除模块",
		},
		// Moderator 不能删除模块（根据 README）
		{
			name:     "Moderator cannot delete module",
			userID:   moderatorUser.ID,
			userRole: "user",
			setupModule: func() uint {
				mod := testutils.CreateTestModule(db, owner.ID)
				db.Table("module_moderators").Create(&module.ModuleModerator{
					ModuleID:  mod.ID,
					UserID:    moderatorUser.ID,
					Role:      "moderator",
					CreatedAt: time.Now(),
				})
				return mod.ID
			},
			expectError: true,
			errorMsg:    "只有模块所有者或系统管理员可以删除模块",
		},
		// 普通用户不能删除模块
		{
			name:     "Regular user cannot delete module",
			userID:   regularUser.ID,
			userRole: "user",
			setupModule: func() uint {
				mod := testutils.CreateTestModule(db, owner.ID)
				return mod.ID
			},
			expectError: true,
			errorMsg:    "只有模块所有者或系统管理员可以删除模块",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleID := tt.setupModule()

			_, err := service.DeleteModule(moduleID, tt.userID, tt.userRole)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					// 验证模块已删除
					var deletedModule module.Module
					if err := db.First(&deletedModule, moduleID).Error; err == nil {
						t.Errorf("Module should be deleted but still exists")
					}
				}
			}
		})
	}
}

// TestGetModule_Integration 集成测试：获取单个模块
func TestGetModule_Integration(t *testing.T) {
	service, db := setupModuleService(t)

	// 创建测试用户和模块
	owner := testutils.CreateTestUser(db)
	testModule := testutils.CreateTestModule(db, owner.ID)

	tests := []struct {
		name        string
		moduleID    uint
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Get existing module",
			moduleID:    testModule.ID,
			expectError: false,
		},
		{
			name:        "Get non-existent module",
			moduleID:    99999,
			expectError: true,
			errorMsg:    "模块不存在",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := service.GetModule(tt.moduleID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if module == nil {
					t.Errorf("Module is nil")
				} else if module.ID != tt.moduleID {
					t.Errorf("Expected module ID %d, got %d", tt.moduleID, module.ID)
				}
			}
		})
	}
}

// TestGetBreadcrumbs_Integration 集成测试：获取面包屑导航
func TestGetBreadcrumbs_Integration(t *testing.T) {
	service, db := setupModuleService(t)

	// 创建测试用户和模块树
	// Root -> Child -> Grandchild
	owner := testutils.CreateTestUser(db)
	rootModule := testutils.CreateTestModule(db, owner.ID)
	childModule := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(rootModule.ID))
	grandchildModule := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(childModule.ID))

	tests := []struct {
		name           string
		moduleID       uint
		expectError    bool
		expectedLength int
		expectedNames  []string
	}{
		{
			name:           "Get breadcrumbs for root module",
			moduleID:       rootModule.ID,
			expectError:    false,
			expectedLength: 1,
			expectedNames:  []string{rootModule.ModuleName},
		},
		{
			name:           "Get breadcrumbs for child module",
			moduleID:       childModule.ID,
			expectError:    false,
			expectedLength: 2,
			expectedNames:  []string{rootModule.ModuleName, childModule.ModuleName},
		},
		{
			name:           "Get breadcrumbs for grandchild module",
			moduleID:       grandchildModule.ID,
			expectError:    false,
			expectedLength: 3,
			expectedNames:  []string{rootModule.ModuleName, childModule.ModuleName, grandchildModule.ModuleName},
		},
		{
			name:        "Get breadcrumbs for non-existent module",
			moduleID:    99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			breadcrumbs, err := service.GetBreadcrumbs(tt.moduleID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(breadcrumbs) != tt.expectedLength {
					t.Errorf("Expected %d breadcrumbs, got %d", tt.expectedLength, len(breadcrumbs))
				} else {
					// 验证面包屑顺序和名称
					for i, expectedName := range tt.expectedNames {
						if breadcrumbs[i].Name != expectedName {
							t.Errorf("Breadcrumb[%d]: expected name %q, got %q", i, expectedName, breadcrumbs[i].Name)
						}
					}
				}
			}
		})
	}
}

// TestGetModerators_Integration 集成测试：获取协作者列表
func TestGetModerators_Integration(t *testing.T) {
	service, db := setupModuleService(t)

	// 创建测试用户
	owner := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db)
	moderatorUser := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	globalAdmin := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// 创建测试模块
	testModule := testutils.CreateTestModule(db, owner.ID)

	// 添加协作者
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  testModule.ID,
		UserID:    adminUser.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	})
	db.Table("module_moderators").Create(&module.ModuleModerator{
		ModuleID:  testModule.ID,
		UserID:    moderatorUser.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	})

	tests := []struct {
		name        string
		userID      uint
		userRole    string
		expectError bool
		errorMsg    string
		expectCount int
	}{
		{
			name:        "Owner can get moderators",
			userID:      owner.ID,
			userRole:    "user",
			expectError: false,
			expectCount: 3, // owner + admin + moderator (GetModeratorsWithUserInfo 返回 owner 和所有 moderators)
		},
		{
			name:        "Admin can get moderators",
			userID:      adminUser.ID,
			userRole:    "user",
			expectError: false,
			expectCount: 3, // owner + admin + moderator
		},
		{
			name:        "Global_Admin can get moderators",
			userID:      globalAdmin.ID,
			userRole:    "admin",
			expectError: false,
			expectCount: 3, // owner + admin + moderator
		},
		{
			name:        "Moderator can get moderators (has permission)",
			userID:      moderatorUser.ID,
			userRole:    "user",
			expectError: false,
			expectCount: 3, // owner + admin + moderator (CheckModulePermission 返回 true 因为用户是 moderator)
		},
		{
			name:        "Regular user cannot get moderators",
			userID:      regularUser.ID,
			userRole:    "user",
			expectError: true,
			errorMsg:    "您没有权限查看协作者列表",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moderators, err := service.GetModerators(testModule.ID, tt.userID, tt.userRole)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if len(moderators) != tt.expectCount {
					t.Errorf("Expected %d moderators, got %d", tt.expectCount, len(moderators))
				} else {
					// 验证协作者信息
					userIDs := make(map[uint]bool)
					roles := make(map[string]bool)
					for _, m := range moderators {
						userIDs[m.UserID] = true
						roles[m.Role] = true
						if m.Username == "" {
							t.Errorf("Moderator username should not be empty")
						}
						if m.Role != "owner" && m.Role != "admin" && m.Role != "moderator" {
							t.Errorf("Invalid role: %s", m.Role)
						}
					}
					// 验证包含owner、admin和moderator
					if !userIDs[owner.ID] {
						t.Errorf("Expected owner in moderators list")
					}
					if !userIDs[adminUser.ID] {
						t.Errorf("Expected admin user in moderators list")
					}
					if !userIDs[moderatorUser.ID] {
						t.Errorf("Expected moderator user in moderators list")
					}
					// 验证包含owner角色
					if !roles["owner"] {
						t.Errorf("Expected owner role in moderators list")
					}
				}
			}
		})
	}
}

// TestLockService_Integration 集成测试：编辑锁服务
func TestLockService_Integration(t *testing.T) {
	redisClient := testutils.SetupTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available, skipping LockService tests")
	}

	lockService := NewLockService(redisClient)

	// 创建测试用户
	user1 := &struct {
		ID       uint
		Username string
	}{
		ID:       1,
		Username: "user1",
	}
	user2 := &struct {
		ID       uint
		Username string
	}{
		ID:       2,
		Username: "user2",
	}

	t.Run("Acquire lock when no lock exists", func(t *testing.T) {
		// 清理之前的锁
		ctx := context.Background()
		redisClient.Del(ctx, LockKey)

		result, err := lockService.AcquireLock(user1.ID, user1.Username)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("Expected lock acquisition to succeed")
		}
		if result.LockedBy != nil {
			t.Errorf("Expected LockedBy to be nil when lock is acquired")
		}
	})

	t.Run("Acquire lock when already held by same user", func(t *testing.T) {
		ctx := context.Background()
		// 确保user1持有锁
		redisClient.Del(ctx, LockKey)
		lockService.AcquireLock(user1.ID, user1.Username)

		// 同一用户再次获取锁应该成功
		result, err := lockService.AcquireLock(user1.ID, user1.Username)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("Expected lock acquisition to succeed for same user")
		}
	})

	t.Run("Acquire lock when held by different user", func(t *testing.T) {
		ctx := context.Background()
		// user1持有锁
		redisClient.Del(ctx, LockKey)
		lockService.AcquireLock(user1.ID, user1.Username)

		// user2尝试获取锁应该失败
		result, err := lockService.AcquireLock(user2.ID, user2.Username)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result.Success {
			t.Errorf("Expected lock acquisition to fail when held by different user")
		}
		if result.LockedBy == nil {
			t.Errorf("Expected LockedBy to be set when lock is held by different user")
		} else {
			if result.LockedBy.ID != user1.ID {
				t.Errorf("Expected LockedBy.ID = %d, got %d", user1.ID, result.LockedBy.ID)
			}
			if result.LockedBy.Username != user1.Username {
				t.Errorf("Expected LockedBy.Username = %q, got %q", user1.Username, result.LockedBy.Username)
			}
		}
		if result.LockedAt == "" {
			t.Errorf("Expected LockedAt to be set")
		}
	})

	t.Run("Release lock when held by same user", func(t *testing.T) {
		ctx := context.Background()
		// user1持有锁
		redisClient.Del(ctx, LockKey)
		lockService.AcquireLock(user1.ID, user1.Username)

		// user1释放锁应该成功
		err := lockService.ReleaseLock(user1.ID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// 验证锁已释放
		exists, err := redisClient.Exists(ctx, LockKey).Result()
		if err != nil {
			t.Fatalf("Failed to check lock existence: %v", err)
		}
		if exists > 0 {
			t.Errorf("Expected lock to be released, but it still exists")
		}
	})

	t.Run("Release lock when held by different user", func(t *testing.T) {
		ctx := context.Background()
		// user1持有锁
		redisClient.Del(ctx, LockKey)
		lockService.AcquireLock(user1.ID, user1.Username)

		// user2尝试释放锁应该失败
		err := lockService.ReleaseLock(user2.ID)
		if err == nil {
			t.Errorf("Expected error when releasing lock held by different user")
		} else if !strings.Contains(err.Error(), "不能释放他人持有的锁") {
			t.Errorf("Expected error message containing '不能释放他人持有的锁', got %q", err.Error())
		}
	})

	t.Run("Release lock when no lock exists", func(t *testing.T) {
		ctx := context.Background()
		// 确保没有锁
		redisClient.Del(ctx, LockKey)

		// 释放不存在的锁应该成功（幂等操作）
		err := lockService.ReleaseLock(user1.ID)
		if err != nil {
			t.Fatalf("Unexpected error when releasing non-existent lock: %v", err)
		}
	})
}

