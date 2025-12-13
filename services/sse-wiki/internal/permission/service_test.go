package permission

import (
	"testing"
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
		{"owner > admin", "owner", "admin", true, false},
		{"owner > moderator", "owner", "moderator", true, false},
		{"admin > moderator", "admin", "moderator", true, false},
		{"moderator > user", "moderator", "user", true, false},
		{"owner == author", "owner", "author", false, true},
		{"user < moderator", "user", "moderator", false, false},
		{"admin < owner", "admin", "owner", false, false},
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
