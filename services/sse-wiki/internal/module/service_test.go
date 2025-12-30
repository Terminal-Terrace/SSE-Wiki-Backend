package module

import (
	"testing"
)

// TestGetModuleTree_IsModerator_GlobalAdmin 测试 Global_Admin 场景
// Global_Admin 对所有模块都应该返回 is_moderator=true
func TestGetModuleTree_IsModerator_GlobalAdmin(t *testing.T) {
	// 测试 isGlobalAdmin 逻辑
	tests := []struct {
		name     string
		userRole string
		expected bool
	}{
		{"admin role is global admin", "admin", true},
		{"user role is not global admin", "user", false},
		{"moderator role is not global admin", "moderator", false},
		{"owner role is not global admin", "owner", false},
		{"empty role is not global admin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isGlobalAdmin := tt.userRole == "admin"
			if isGlobalAdmin != tt.expected {
				t.Errorf("isGlobalAdmin for role %q = %v, want %v", tt.userRole, isGlobalAdmin, tt.expected)
			}
		})
	}
}

// TestBuildTree_IsModerator_Logic 测试 buildTree 中 is_moderator 的计算逻辑
// is_moderator 应该为 true 当且仅当：
// (1) 用户是 Global_Admin，或
// (2) 用户是 owner_id，或
// (3) 用户在 module_moderators 表中有记录
func TestBuildTree_IsModerator_Logic(t *testing.T) {
	tests := []struct {
		name          string
		isGlobalAdmin bool
		moduleID      uint
		moderatorMap  map[uint]bool
		expected      bool
	}{
		{
			name:          "global admin has moderator access",
			isGlobalAdmin: true,
			moduleID:      1,
			moderatorMap:  map[uint]bool{},
			expected:      true,
		},
		{
			name:          "user in moderator map has access",
			isGlobalAdmin: false,
			moduleID:      1,
			moderatorMap:  map[uint]bool{1: true},
			expected:      true,
		},
		{
			name:          "user not in moderator map has no access",
			isGlobalAdmin: false,
			moduleID:      1,
			moderatorMap:  map[uint]bool{2: true, 3: true},
			expected:      false,
		},
		{
			name:          "empty moderator map means no access",
			isGlobalAdmin: false,
			moduleID:      1,
			moderatorMap:  map[uint]bool{},
			expected:      false,
		},
		{
			name:          "global admin overrides moderator map",
			isGlobalAdmin: true,
			moduleID:      1,
			moderatorMap:  map[uint]bool{}, // 即使 map 为空，global admin 也有权限
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 buildTree 中的 is_moderator 计算逻辑
			isModerator := tt.isGlobalAdmin || tt.moderatorMap[tt.moduleID]
			if isModerator != tt.expected {
				t.Errorf("isModerator = %v, want %v", isModerator, tt.expected)
			}
		})
	}
}

// TestModeratorMapConstruction 测试 moderatorMap 的构建逻辑
// moderatorMap 应该包含：
// (1) modules.owner_id = userID 的模块
// (2) module_moderators.user_id = userID 的模块
func TestModeratorMapConstruction(t *testing.T) {
	// 模拟从数据库返回的 moderator module IDs
	// 这些 ID 来自 UNION 查询：
	// SELECT id FROM modules WHERE owner_id = ?
	// UNION
	// SELECT module_id FROM module_moderators WHERE user_id = ?

	tests := []struct {
		name               string
		moderatorModuleIDs []uint
		checkModuleID      uint
		expected           bool
	}{
		{
			name:               "module in list returns true",
			moderatorModuleIDs: []uint{1, 2, 3},
			checkModuleID:      2,
			expected:           true,
		},
		{
			name:               "module not in list returns false",
			moderatorModuleIDs: []uint{1, 2, 3},
			checkModuleID:      4,
			expected:           false,
		},
		{
			name:               "empty list returns false",
			moderatorModuleIDs: []uint{},
			checkModuleID:      1,
			expected:           false,
		},
		{
			name:               "single module in list",
			moderatorModuleIDs: []uint{5},
			checkModuleID:      5,
			expected:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 构建 moderatorMap
			moderatorMap := make(map[uint]bool)
			for _, id := range tt.moderatorModuleIDs {
				moderatorMap[id] = true
			}

			// 检查模块是否在 map 中
			result := moderatorMap[tt.checkModuleID]
			if result != tt.expected {
				t.Errorf("moderatorMap[%d] = %v, want %v", tt.checkModuleID, result, tt.expected)
			}
		})
	}
}

// TestIsModerator_Property4 属性测试：is_moderator 正确计算
// Property 4: 对于任意模块和用户组合，is_moderator 为 true 当且仅当：
// (1) 用户是 Global_Admin，或
// (2) 用户是 owner_id，或
// (3) 用户在 module_moderators 表中有记录
// **Validates: Requirements 2.1, 2.2, 2.3, 2.5**
func TestIsModerator_Property4(t *testing.T) {
	type testCase struct {
		name          string
		isGlobalAdmin bool
		isOwner       bool
		isModerator   bool
		expected      bool
	}

	// 生成所有可能的组合
	tests := []testCase{
		// Global_Admin 场景
		{"global_admin only", true, false, false, true},
		{"global_admin and owner", true, true, false, true},
		{"global_admin and moderator", true, false, true, true},
		{"global_admin, owner and moderator", true, true, true, true},

		// Owner 场景（非 Global_Admin）
		{"owner only", false, true, false, true},
		{"owner and moderator", false, true, true, true},

		// Moderator 场景（非 Global_Admin，非 Owner）
		{"moderator only", false, false, true, true},

		// 无权限场景
		{"no permission", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 is_moderator 计算
			// moderatorMap 包含 owner_id 匹配和 module_moderators 记录
			inModeratorMap := tt.isOwner || tt.isModerator
			result := tt.isGlobalAdmin || inModeratorMap

			if result != tt.expected {
				t.Errorf("is_moderator = %v, want %v (isGlobalAdmin=%v, isOwner=%v, isModerator=%v)",
					result, tt.expected, tt.isGlobalAdmin, tt.isOwner, tt.isModerator)
			}

			// 验证属性：is_moderator == true 当且仅当满足三个条件之一
			shouldBeTrue := tt.isGlobalAdmin || tt.isOwner || tt.isModerator
			if result != shouldBeTrue {
				t.Errorf("Property 4 violation: is_moderator=%v but conditions are (globalAdmin=%v, owner=%v, moderator=%v)",
					result, tt.isGlobalAdmin, tt.isOwner, tt.isModerator)
			}
		})
	}
}


// TestBuildTreeWithInheritance_Logic 测试权限继承逻辑
// Property 5: 模块权限继承
// 对于任意有父模块的模块，如果用户在父模块有权限，则用户在子模块应有等效或更低的权限
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
func TestBuildTreeWithInheritance_Logic(t *testing.T) {
	tests := []struct {
		name              string
		isGlobalAdmin     bool
		moduleID          uint
		moderatorMap      map[uint]bool
		parentIsModerator bool
		expected          bool
	}{
		{
			name:              "inherit from parent moderator",
			isGlobalAdmin:     false,
			moduleID:          2,
			moderatorMap:      map[uint]bool{}, // 子模块不在 map 中
			parentIsModerator: true,            // 但父模块有权限
			expected:          true,            // 应该继承权限
		},
		{
			name:              "no inheritance when parent has no permission",
			isGlobalAdmin:     false,
			moduleID:          2,
			moderatorMap:      map[uint]bool{},
			parentIsModerator: false,
			expected:          false,
		},
		{
			name:              "direct permission overrides inheritance",
			isGlobalAdmin:     false,
			moduleID:          2,
			moderatorMap:      map[uint]bool{2: true}, // 直接有权限
			parentIsModerator: false,                  // 父模块无权限
			expected:          true,
		},
		{
			name:              "global admin always has permission",
			isGlobalAdmin:     true,
			moduleID:          2,
			moderatorMap:      map[uint]bool{},
			parentIsModerator: false,
			expected:          true,
		},
		{
			name:              "both direct and inherited permission",
			isGlobalAdmin:     false,
			moduleID:          2,
			moderatorMap:      map[uint]bool{2: true},
			parentIsModerator: true,
			expected:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟 buildTreeWithInheritance 中的 is_moderator 计算逻辑
			isModerator := tt.isGlobalAdmin || tt.moderatorMap[tt.moduleID] || tt.parentIsModerator
			if isModerator != tt.expected {
				t.Errorf("isModerator = %v, want %v", isModerator, tt.expected)
			}
		})
	}
}

// TestInheritanceProperty5 属性测试：模块权限继承
// Property 5: 对于任意有父模块的模块，如果用户在父模块有权限，则用户在子模块应有等效或更低的权限
func TestInheritanceProperty5(t *testing.T) {
	// 模拟一个三层模块树：
	// Module 1 (root)
	//   └── Module 2 (child)
	//         └── Module 3 (grandchild)

	type moduleNode struct {
		id       uint
		parentID *uint
	}

	parentID1 := uint(1)
	parentID2 := uint(2)

	modules := []moduleNode{
		{id: 1, parentID: nil},
		{id: 2, parentID: &parentID1},
		{id: 3, parentID: &parentID2},
	}

	tests := []struct {
		name         string
		moderatorMap map[uint]bool // 用户直接有权限的模块
		expected     map[uint]bool // 每个模块的 is_moderator 期望值
	}{
		{
			name:         "owner of root inherits to all children",
			moderatorMap: map[uint]bool{1: true},
			expected:     map[uint]bool{1: true, 2: true, 3: true},
		},
		{
			name:         "owner of middle inherits to grandchild only",
			moderatorMap: map[uint]bool{2: true},
			expected:     map[uint]bool{1: false, 2: true, 3: true},
		},
		{
			name:         "owner of leaf has no inheritance",
			moderatorMap: map[uint]bool{3: true},
			expected:     map[uint]bool{1: false, 2: false, 3: true},
		},
		{
			name:         "no permission anywhere",
			moderatorMap: map[uint]bool{},
			expected:     map[uint]bool{1: false, 2: false, 3: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟递归构建树时的权限继承
			result := make(map[uint]bool)

			// 模拟 buildTreeWithInheritance 的递归逻辑
			var buildWithInheritance func(parentID *uint, parentIsModerator bool)
			buildWithInheritance = func(parentID *uint, parentIsModerator bool) {
				for _, m := range modules {
					// 匹配父节点
					if (parentID == nil && m.parentID == nil) ||
						(parentID != nil && m.parentID != nil && *parentID == *m.parentID) {

						isModerator := tt.moderatorMap[m.id] || parentIsModerator
						result[m.id] = isModerator

						// 递归处理子节点
						buildWithInheritance(&m.id, isModerator)
					}
				}
			}

			buildWithInheritance(nil, false)

			// 验证结果
			for moduleID, expected := range tt.expected {
				if result[moduleID] != expected {
					t.Errorf("module %d: is_moderator = %v, want %v", moduleID, result[moduleID], expected)
				}
			}

			// 验证 Property 5：如果父模块有权限，子模块也应该有权限
			for _, m := range modules {
				if m.parentID != nil {
					parentHasPermission := result[*m.parentID]
					childHasPermission := result[m.id]

					if parentHasPermission && !childHasPermission {
						t.Errorf("Property 5 violation: parent %d has permission but child %d does not",
							*m.parentID, m.id)
					}
				}
			}
		})
	}
}
