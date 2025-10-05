package module

import "time"

// Module 模块表
type Module struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ModuleName string    `gorm:"type:varchar(100);not null" json:"module_name"`
	ParentID   *uint     `gorm:"index;default:null" json:"parent_id"` // NULL表示顶级模块
	OwnerID    uint      `gorm:"not null;index" json:"owner_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联（仅用于查询，不会在数据库中创建字段）
	// 添加级联删除约束：删除父模块时自动删除所有子模块
	Children []Module `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE" json:"children,omitempty"`
}

// ModuleModerator 模块协作者表
type ModuleModerator struct {
	ModuleID  uint      `gorm:"primaryKey" json:"module_id"`
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	Role      string    `gorm:"type:varchar(50);not null;default:'moderator'" json:"role"` // admin, moderator
	CreatedAt time.Time `json:"created_at"`
}

func (Module) TableName() string {
	return "modules"
}

func (ModuleModerator) TableName() string {
	return "module_moderators"
}
