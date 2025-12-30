package permission

import (
	"testing"
	"time"

	"terminal-terrace/sse-wiki/internal/model/module"
	"terminal-terrace/sse-wiki/internal/testutils"
)

// TestGetRoleLevel 测试角色等级获取
func TestGetRoleLevel(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected int
	}{
		{"owner role", "owner", RoleLevelOwner},
		{"author role", "author", RoleLevelOwner},
		{"admin role", "admin", RoleLevelAdmin},
		{"moderator role", "moderator", RoleLevelModerator},
		{"user role", "user", RoleLevelUser},
		{"unknown role", "unknown", RoleLevelUnknown},
		{"empty role", "", RoleLevelUnknown},
		{"case sensitive owner", "Owner", RoleLevelUnknown}, // Case sensitive
		{"case sensitive ADMIN", "ADMIN", RoleLevelUnknown}, // Case sensitive
		{"whitespace role", " owner ", RoleLevelUnknown},     // Whitespace
		{"numeric role", "123", RoleLevelUnknown},            // Numeric
		{"special chars role", "owner@admin", RoleLevelUnknown}, // Special chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRoleLevel(tt.role)
			if result != tt.expected {
				t.Errorf("GetRoleLevel(%q) = %d, want %d", tt.role, result, tt.expected)
			}
		})
	}
}

// TestCompareRoles 测试角色比较
func TestCompareRoles(t *testing.T) {
	tests := []struct {
		name     string
		roleA    string
		roleB    string
		positive bool // true if roleA > roleB
		zero     bool // true if roleA == roleB
	}{
		// Standard comparisons
		{"owner > admin", "owner", "admin", true, false},
		{"owner > moderator", "owner", "moderator", true, false},
		{"owner > user", "owner", "user", true, false},
		{"admin > moderator", "admin", "moderator", true, false},
		{"admin > user", "admin", "user", true, false},
		{"moderator > user", "moderator", "user", true, false},
		
		// Equality cases
		{"owner == author", "owner", "author", false, true},
		{"author == owner", "author", "owner", false, true},
		{"admin == admin", "admin", "admin", false, true},
		{"moderator == moderator", "moderator", "moderator", false, true},
		{"user == user", "user", "user", false, true},
		
		// Reverse comparisons
		{"user < moderator", "user", "moderator", false, false},
		{"user < admin", "user", "admin", false, false},
		{"user < owner", "user", "owner", false, false},
		{"moderator < admin", "moderator", "admin", false, false},
		{"moderator < owner", "moderator", "owner", false, false},
		{"admin < owner", "admin", "owner", false, false},
		
		// Unknown role comparisons
		{"unknown == unknown", "unknown", "unknown", false, true},
		{"unknown < user", "unknown", "user", false, false},
		{"user > unknown", "user", "unknown", true, false},
		{"empty == empty", "", "", false, true},
		{"empty < user", "", "user", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareRoles(tt.roleA, tt.roleB)
			if tt.positive && result <= 0 {
				t.Errorf("CompareRoles(%q, %q) = %d, expected positive", tt.roleA, tt.roleB, result)
			}
			if tt.zero && result != 0 {
				t.Errorf("CompareRoles(%q, %q) = %d, expected zero", tt.roleA, tt.roleB, result)
			}
			if !tt.positive && !tt.zero && result >= 0 {
				t.Errorf("CompareRoles(%q, %q) = %d, expected negative", tt.roleA, tt.roleB, result)
			}
		})
	}
}


// TestHasRequiredRole 测试权限要求检查
func TestHasRequiredRole(t *testing.T) {
	tests := []struct {
		name         string
		actualRole   string
		requiredRole string
		expected     bool
	}{
		// owner 可以做任何事
		{"owner has owner permission", "owner", "owner", true},
		{"owner has admin permission", "owner", "admin", true},
		{"owner has moderator permission", "owner", "moderator", true},
		{"owner has user permission", "owner", "user", true},
		{"author has owner permission", "author", "owner", true},
		{"author has admin permission", "author", "admin", true},
		{"author has moderator permission", "author", "moderator", true},
		{"author has user permission", "author", "user", true},

		// admin 权限
		{"admin has admin permission", "admin", "admin", true},
		{"admin has moderator permission", "admin", "moderator", true},
		{"admin has user permission", "admin", "user", true},
		{"admin lacks owner permission", "admin", "owner", false},

		// moderator 权限
		{"moderator has moderator permission", "moderator", "moderator", true},
		{"moderator has user permission", "moderator", "user", true},
		{"moderator lacks admin permission", "moderator", "admin", false},
		{"moderator lacks owner permission", "moderator", "owner", false},

		// user 权限
		{"user has user permission", "user", "user", true},
		{"user lacks moderator permission", "user", "moderator", false},
		{"user lacks admin permission", "user", "admin", false},
		{"user lacks owner permission", "user", "owner", false},

		// Unknown role cases
		// Note: Unknown/empty roles should be rejected (unauthenticated users should be handled at BFF controller layer)
		// Permission layer assumes valid input and rejects unknown roles for security
		{"unknown lacks owner permission", "unknown", "owner", false},
		{"unknown lacks admin permission", "unknown", "admin", false},
		{"unknown lacks moderator permission", "unknown", "moderator", false},
		{"unknown lacks user permission", "unknown", "user", false}, // Unknown role (level 0) should be rejected, not auto-upgraded to user
		{"empty lacks owner permission", "", "owner", false},
		{"empty lacks user permission", "", "user", false}, // Empty role (level 0) should be rejected, unauthenticated handled at BFF layer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasRequiredRole(tt.actualRole, tt.requiredRole)
			if result != tt.expected {
				t.Errorf("HasRequiredRole(%q, %q) = %v, want %v", tt.actualRole, tt.requiredRole, result, tt.expected)
			}
		})
	}
}

// TestNormalizeRole 测试角色标准化
func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected string
	}{
		{"owner stays owner", "owner", "owner"},
		{"author becomes owner", "author", "owner"},
		{"admin stays admin", "admin", "admin"},
		{"moderator stays moderator", "moderator", "moderator"},
		{"user stays user", "user", "user"},
		{"unknown becomes user", "unknown", "user"},
		{"empty becomes user", "", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRole(tt.role)
			if result != tt.expected {
				t.Errorf("NormalizeRole(%q) = %q, want %q", tt.role, result, tt.expected)
			}
		})
	}
}

// TestIsGlobalAdmin 测试全局管理员检查
func TestIsGlobalAdmin(t *testing.T) {
	tests := []struct {
		name     string
		userRole string
		expected bool
	}{
		{"admin is global admin", "admin", true},
		{"user is not global admin", "user", false},
		{"moderator is not global admin", "moderator", false},
		{"owner is not global admin", "owner", false},
		{"empty is not global admin", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGlobalAdmin(tt.userRole)
			if result != tt.expected {
				t.Errorf("IsGlobalAdmin(%q) = %v, want %v", tt.userRole, result, tt.expected)
			}
		})
	}
}

// TestNewPermissionResult 测试权限结果创建
func TestNewPermissionResult(t *testing.T) {
	result := NewPermissionResult(true, "owner", PermissionSourceDirect)

	if !result.HasPermission {
		t.Error("HasPermission should be true")
	}
	if result.EffectiveRole != "owner" {
		t.Errorf("EffectiveRole = %q, want %q", result.EffectiveRole, "owner")
	}
	if result.PermissionSource != PermissionSourceDirect {
		t.Errorf("PermissionSource = %q, want %q", result.PermissionSource, PermissionSourceDirect)
	}
}

// TestNoPermission 测试无权限结果
func TestNoPermission(t *testing.T) {
	result := NoPermission()

	if result.HasPermission {
		t.Error("HasPermission should be false")
	}
	if result.EffectiveRole != "user" {
		t.Errorf("EffectiveRole = %q, want %q", result.EffectiveRole, "user")
	}
	if result.PermissionSource != PermissionSourceNone {
		t.Errorf("PermissionSource = %q, want %q", result.PermissionSource, PermissionSourceNone)
	}
}

// TestRoleLevelHierarchy 测试角色等级层次结构的一致性
// Property 1: 角色等级比较一致性
// 对于任意两个角色 A 和 B，如果 RoleLevel(A) > RoleLevel(B)，则角色 A 应拥有角色 B 的所有权限
func TestRoleLevelHierarchy(t *testing.T) {
	roles := []string{"owner", "admin", "moderator", "user"}

	for i, roleA := range roles {
		for j, roleB := range roles {
			levelA := GetRoleLevel(roleA)
			levelB := GetRoleLevel(roleB)

			// 如果 levelA > levelB，则 roleA 应该有 roleB 的所有权限
			if levelA > levelB {
				if !HasRequiredRole(roleA, roleB) {
					t.Errorf("Role hierarchy violation: %s (level %d) should have %s (level %d) permissions",
						roleA, levelA, roleB, levelB)
				}
			}

			// 如果 levelA < levelB，则 roleA 不应该有 roleB 的权限
			if levelA < levelB {
				if HasRequiredRole(roleA, roleB) {
					t.Errorf("Role hierarchy violation: %s (level %d) should NOT have %s (level %d) permissions",
						roleA, levelA, roleB, levelB)
				}
			}

			// 验证索引顺序与等级顺序一致（roles 数组按权限从高到低排列）
			if i < j && levelA <= levelB {
				t.Errorf("Role order violation: %s should have higher level than %s", roleA, roleB)
			}
		}
	}
}

// TestPermissionInheritance 测试权限继承逻辑
func TestPermissionInheritance(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	owner := testutils.CreateTestUser(db)
	moderator := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)

	// Create parent module
	parentModule := testutils.CreateTestModule(db, owner.ID)

	// Create child module
	childModule := testutils.CreateTestModule(db, owner.ID, testutils.WithParentID(parentModule.ID))

	// Add moderator to parent module
	moderatorRecord := &module.ModuleModerator{
		ModuleID:  parentModule.ID,
		UserID:    moderator.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	}
	db.Create(moderatorRecord)

	tests := []struct {
		name         string
		moduleID     uint
		userID       uint
		userRole     string
		requiredRole string
		expected     bool
		source       PermissionSource
	}{
		{
			name:         "owner has direct permission on parent",
			moduleID:     parentModule.ID,
			userID:       owner.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceDirect,
		},
		{
			name:         "moderator has direct permission on parent",
			moduleID:     parentModule.ID,
			userID:       moderator.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     true,
			source:       PermissionSourceDirect,
		},
		{
			name:         "owner inherits permission on child",
			moduleID:     childModule.ID,
			userID:       owner.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceDirect, // Owner of child module, not inherited
		},
		{
			name:         "moderator inherits permission on child",
			moduleID:     childModule.ID,
			userID:       moderator.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     true,
			source:       PermissionSourceInherited,
		},
		{
			name:         "regular user has no permission on child",
			moduleID:     childModule.ID,
			userID:       regularUser.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     false,
			source:       PermissionSourceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckModulePermissionWithInheritance(tt.moduleID, tt.userID, tt.userRole, tt.requiredRole)
			if result.HasPermission != tt.expected {
				t.Errorf("CheckModulePermissionWithInheritance() HasPermission = %v, want %v", result.HasPermission, tt.expected)
			}
			if result.PermissionSource != tt.source {
				t.Errorf("CheckModulePermissionWithInheritance() PermissionSource = %v, want %v", result.PermissionSource, tt.source)
			}
		})
	}
}

// TestPermissionInheritance_PureInheritance 测试纯继承场景
// 验证：父模块 owner（不是子模块创建者）→ 子模块继承为 owner
func TestPermissionInheritance_PureInheritance(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	parentOwner := testutils.CreateTestUser(db)
	childCreator := testutils.CreateTestUser(db) // 子模块的创建者（不同用户）
	parentModerator := testutils.CreateTestUser(db)

	// Create parent module owned by parentOwner
	parentModule := testutils.CreateTestModule(db, parentOwner.ID)

	// Create child module owned by childCreator (different from parentOwner)
	childModule := testutils.CreateTestModule(db, childCreator.ID, testutils.WithParentID(parentModule.ID))

	// Add moderator to parent module
	moderatorRecord := &module.ModuleModerator{
		ModuleID:  parentModule.ID,
		UserID:    parentModerator.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	}
	db.Create(moderatorRecord)

	tests := []struct {
		name         string
		moduleID     uint
		userID       uint
		userRole     string
		requiredRole string
		expected     bool
		source       PermissionSource
		expectedRole string
	}{
		{
			name:         "parent owner inherits owner role on child (pure inheritance)",
			moduleID:     childModule.ID,
			userID:       parentOwner.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceInherited,
			expectedRole: "owner", // 父模块 owner 继承到子模块仍为 owner
		},
		{
			name:         "parent moderator inherits moderator role on child",
			moduleID:     childModule.ID,
			userID:       parentModerator.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     true,
			source:       PermissionSourceInherited,
			expectedRole: "moderator",
		},
		{
			name:         "child creator has direct owner role (not inherited)",
			moduleID:     childModule.ID,
			userID:       childCreator.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceDirect, // 直接权限，不是继承
			expectedRole: "owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckModulePermissionWithInheritance(tt.moduleID, tt.userID, tt.userRole, tt.requiredRole)
			if result.HasPermission != tt.expected {
				t.Errorf("CheckModulePermissionWithInheritance() HasPermission = %v, want %v", result.HasPermission, tt.expected)
			}
			if result.PermissionSource != tt.source {
				t.Errorf("CheckModulePermissionWithInheritance() PermissionSource = %v, want %v", result.PermissionSource, tt.source)
			}
			if result.EffectiveRole != tt.expectedRole {
				t.Errorf("CheckModulePermissionWithInheritance() EffectiveRole = %v, want %v", result.EffectiveRole, tt.expectedRole)
			}
		})
	}
}

// TestGlobalAdminPermission 测试全局管理员权限检查
func TestGlobalAdminPermission(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	adminUser := testutils.CreateTestUser(db, testutils.WithRole("admin"))
	regularUser := testutils.CreateTestUser(db)
	moduleOwner := testutils.CreateTestUser(db)

	// Create module owned by regular user
	testModule := testutils.CreateTestModule(db, moduleOwner.ID)

	tests := []struct {
		name         string
		moduleID     uint
		userID       uint
		userRole     string
		requiredRole string
		expected     bool
		source       PermissionSource
	}{
		{
			name:         "global admin bypasses module-level permissions",
			moduleID:     testModule.ID,
			userID:       adminUser.ID,
			userRole:     "admin",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceGlobal,
		},
		{
			name:         "global admin has admin permission",
			moduleID:     testModule.ID,
			userID:       adminUser.ID,
			userRole:     "admin",
			requiredRole: "admin",
			expected:     true,
			source:       PermissionSourceGlobal,
		},
		{
			name:         "regular user lacks permission",
			moduleID:     testModule.ID,
			userID:       regularUser.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     false,
			source:       PermissionSourceNone,
		},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckModulePermission(tt.moduleID, tt.userID, tt.userRole, tt.requiredRole)
			if result.HasPermission != tt.expected {
				t.Errorf("CheckModulePermission() HasPermission = %v, want %v", result.HasPermission, tt.expected)
			}
			if result.PermissionSource != tt.source {
				t.Errorf("CheckModulePermission() PermissionSource = %v, want %v", result.PermissionSource, tt.source)
			}
		})
	}
}

// TestArticlePermission 测试文章权限检查
func TestArticlePermission(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	author := testutils.CreateTestUser(db)
	collaborator := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// Create module for article
	testModule := testutils.CreateTestModule(db, author.ID)

	// Create article
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Add collaborator to article
	collaboratorRecord := &struct {
		ArticleID uint      `gorm:"primaryKey"`
		UserID    uint      `gorm:"primaryKey"`
		Role      string    `gorm:"type:varchar(50);not null"`
		CreatedAt time.Time
	}{
		ArticleID: testArticle.ID,
		UserID:    collaborator.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	}
	db.Table("article_collaborators").Create(collaboratorRecord)

	tests := []struct {
		name         string
		articleID    uint
		userID       uint
		userRole     string
		requiredRole string
		expected     bool
		source       PermissionSource
	}{
		{
			name:         "author has owner permission",
			articleID:    testArticle.ID,
			userID:       author.ID,
			userRole:     "user",
			requiredRole: "owner",
			expected:     true,
			source:       PermissionSourceDirect,
		},
		{
			name:         "author has author permission",
			articleID:    testArticle.ID,
			userID:       author.ID,
			userRole:     "user",
			requiredRole: "author",
			expected:     true,
			source:       PermissionSourceDirect,
		},
		{
			name:         "collaborator has moderator permission",
			articleID:    testArticle.ID,
			userID:       collaborator.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     true,
			source:       PermissionSourceDirect,
		},
		{
			name:         "regular user has no permission",
			articleID:    testArticle.ID,
			userID:       regularUser.ID,
			userRole:     "user",
			requiredRole: "moderator",
			expected:     false,
			source:       PermissionSourceNone,
		},
		{
			name:         "admin has no direct permission (only delete)",
			articleID:    testArticle.ID,
			userID:       adminUser.ID,
			userRole:     "admin",
			requiredRole: "moderator",
			expected:     false,
			source:       PermissionSourceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CheckArticlePermission(tt.articleID, tt.userID, tt.userRole, tt.requiredRole)
			if result.HasPermission != tt.expected {
				t.Errorf("CheckArticlePermission() HasPermission = %v, want %v", result.HasPermission, tt.expected)
			}
			if result.PermissionSource != tt.source {
				t.Errorf("CheckArticlePermission() PermissionSource = %v, want %v", result.PermissionSource, tt.source)
			}
		})
	}
}

// TestCanDeleteArticle 测试文章删除权限
func TestCanDeleteArticle(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	author := testutils.CreateTestUser(db)
	collaborator := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// Create module for article
	testModule := testutils.CreateTestModule(db, author.ID)

	// Create article
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Add collaborator with admin role
	collaboratorRecord := &struct {
		ArticleID uint      `gorm:"primaryKey"`
		UserID    uint      `gorm:"primaryKey"`
		Role      string    `gorm:"type:varchar(50);not null"`
		CreatedAt time.Time
	}{
		ArticleID: testArticle.ID,
		UserID:    collaborator.ID,
		Role:      "admin",
		CreatedAt: time.Now(),
	}
	db.Table("article_collaborators").Create(collaboratorRecord)

	tests := []struct {
		name      string
		articleID uint
		userID    uint
		userRole  string
		expected  bool
	}{
		{
			name:      "author can delete article",
			articleID: testArticle.ID,
			userID:    author.ID,
			userRole:  "user",
			expected:  true,
		},
		{
			name:      "admin collaborator can delete article",
			articleID: testArticle.ID,
			userID:    collaborator.ID,
			userRole:  "user",
			expected:  true,
		},
		{
			name:      "global admin can delete article",
			articleID: testArticle.ID,
			userID:    adminUser.ID,
			userRole:  "admin",
			expected:  true,
		},
		{
			name:      "regular user cannot delete article",
			articleID: testArticle.ID,
			userID:    regularUser.ID,
			userRole:  "user",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CanDeleteArticle(tt.articleID, tt.userID, tt.userRole)
			if result != tt.expected {
				t.Errorf("CanDeleteArticle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetEffectiveArticleRole 测试获取文章有效角色
func TestGetEffectiveArticleRole(t *testing.T) {
	db := testutils.SetupTestDB(t)
	service := NewPermissionService(db)

	// Create test users
	author := testutils.CreateTestUser(db)
	collaborator := testutils.CreateTestUser(db)
	regularUser := testutils.CreateTestUser(db)
	adminUser := testutils.CreateTestUser(db, testutils.WithRole("admin"))

	// Create module for article
	testModule := testutils.CreateTestModule(db, author.ID)

	// Create article
	testArticle := testutils.CreateTestArticle(db, testModule.ID, author.ID)

	// Add collaborator
	collaboratorRecord := &struct {
		ArticleID uint      `gorm:"primaryKey"`
		UserID    uint      `gorm:"primaryKey"`
		Role      string    `gorm:"type:varchar(50);not null"`
		CreatedAt time.Time
	}{
		ArticleID: testArticle.ID,
		UserID:    collaborator.ID,
		Role:      "moderator",
		CreatedAt: time.Now(),
	}
	db.Table("article_collaborators").Create(collaboratorRecord)

	tests := []struct {
		name      string
		articleID uint
		userID    uint
		userRole  string
		expected  string
		source    PermissionSource
	}{
		{
			name:      "author has author role",
			articleID: testArticle.ID,
			userID:    author.ID,
			userRole:  "user",
			expected:  "author",
			source:    PermissionSourceDirect,
		},
		{
			name:      "collaborator has moderator role",
			articleID: testArticle.ID,
			userID:    collaborator.ID,
			userRole:  "user",
			expected:  "moderator",
			source:    PermissionSourceDirect,
		},
		{
			name:      "regular user has user role",
			articleID: testArticle.ID,
			userID:    regularUser.ID,
			userRole:  "user",
			expected:  "user",
			source:    PermissionSourceNone,
		},
		{
			name:      "global admin has user role (can submit edits, can delete, but cannot directly publish or edit metadata)",
			articleID: testArticle.ID,
			userID:    adminUser.ID,
			userRole:  "admin",
			expected:  "user", // Global_Admin 返回 user 级别：可以提交修改（需审核）和删除，但不能直接发布或编辑基础信息
			source:    PermissionSourceNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, source := service.GetEffectiveArticleRole(tt.articleID, tt.userID, tt.userRole)
			if role != tt.expected {
				t.Errorf("GetEffectiveArticleRole() role = %v, want %v", role, tt.expected)
			}
			if source != tt.source {
				t.Errorf("GetEffectiveArticleRole() source = %v, want %v", source, tt.source)
			}
		})
	}
}
