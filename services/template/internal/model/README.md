# model 目录

Model 层定义数据库表结构，使用 GORM 作为 ORM。

## 职责

- 定义数据库表的结构体
- 定义表之间的关联关系
- 执行数据库表的自动迁移
- GORM 钩子函数（BeforeCreate, AfterUpdate 等）

## 目录结构

```
model/
├── model.go         # 表迁移和通用模型
├── user.go          # 用户模型
├── article.go       # 文章模型
└── base.go          # 基础模型（可选）
```

## 示例代码

### 基础模型 (base.go)

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// BaseModel 基础模型（包含通用字段）
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
```

### 用户模型 (user.go)

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
    Email     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
    Password  string         `gorm:"type:varchar(255);not null" json:"-"` // json:"-" 表示不序列化
    Avatar    string         `gorm:"type:varchar(255)" json:"avatar"`
    Role      string         `gorm:"type:varchar(20);default:'user'" json:"role"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

    // 关联关系
    Articles []Article `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
    return "users"
}

// BeforeCreate GORM 钩子：创建前执行
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 例如：密码加密
    // u.Password = hashPassword(u.Password)
    return nil
}
```

### 文章模型 (article.go)

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// Article 文章模型
type Article struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    Title     string         `gorm:"type:varchar(200);not null" json:"title"`
    Content   string         `gorm:"type:text" json:"content"`
    Summary   string         `gorm:"type:text" json:"summary"`
    AuthorID  uint           `gorm:"not null;index" json:"author_id"`
    ModuleID  uint           `gorm:"not null;index" json:"module_id"`
    ViewCount int            `gorm:"default:0" json:"view_count"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

    // 关联关系
    Author *User   `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
    Module *Module `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
}

// TableName 指定表名
func (Article) TableName() string {
    return "articles"
}
```

### 表迁移 (model.go)

```go
package model

import (
    "gorm.io/gorm"
)

// InitTable 初始化数据库表
func InitTable(db *gorm.DB) error {
    // 自动迁移所有模型
    return db.AutoMigrate(
        &User{},
        &Article{},
        &Module{},
        // 添加更多模型...
    )
}
```

## GORM 常用标签

### 字段标签

```go
type User struct {
    // 主键
    ID uint `gorm:"primarykey"`

    // 字段类型和约束
    Username string `gorm:"type:varchar(100);not null;uniqueIndex"`

    // 默认值
    Role string `gorm:"default:'user'"`

    // 索引
    Email string `gorm:"index"`                    // 普通索引
    Phone string `gorm:"uniqueIndex"`             // 唯一索引
    Name  string `gorm:"index:idx_name_age"`      // 组合索引

    // 忽略字段（不映射到数据库）
    TempData string `gorm:"-"`

    // JSON 序列化控制
    Password string `json:"-"`                     // 不序列化
    Extra    string `json:"extra,omitempty"`       // 空值时不序列化
}
```

### 关联标签

```go
// 一对多关联
type User struct {
    Articles []Article `gorm:"foreignKey:AuthorID"`
}

// 多对一关联
type Article struct {
    AuthorID uint  `gorm:"not null"`
    Author   *User `gorm:"foreignKey:AuthorID"`
}

// 多对多关联
type User struct {
    Roles []Role `gorm:"many2many:user_roles;"`
}
```

## GORM 钩子函数

```go
// 创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 密码加密、生成 UUID 等
    return nil
}

// 创建后
func (u *User) AfterCreate(tx *gorm.DB) error {
    // 发送欢迎邮件等
    return nil
}

// 更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    return nil
}

// 更新后
func (u *User) AfterUpdate(tx *gorm.DB) error {
    return nil
}

// 删除前
func (u *User) BeforeDelete(tx *gorm.DB) error {
    return nil
}

// 删除后
func (u *User) AfterDelete(tx *gorm.DB) error {
    return nil
}

// 查询后
func (u *User) AfterFind(tx *gorm.DB) error {
    return nil
}
```

## 软删除

GORM 支持软删除，使用 `DeletedAt` 字段：

```go
type User struct {
    ID        uint
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 软删除（设置 DeletedAt 时间）
db.Delete(&user)

// 查询时自动排除软删除的记录
db.Find(&users)

// 包含软删除的记录
db.Unscoped().Find(&users)

// 永久删除
db.Unscoped().Delete(&user)
```

## 表名约定

GORM 默认使用复数形式作为表名：
- `User` → `users`
- `Article` → `articles`

自定义表名：

```go
func (User) TableName() string {
    return "user"  // 使用单数形式
}
```

## 设计原则

1. **使用结构体定义表结构**：不要手写 SQL DDL
2. **利用 GORM 标签**：充分使用 gorm 和 json 标签
3. **敏感字段处理**：密码等字段使用 `json:"-"` 防止序列化
4. **软删除**：使用 `DeletedAt` 字段实现软删除
5. **关联关系**：明确定义表之间的关联关系
6. **钩子函数**：在钩子中处理密码加密、数据校验等

## 注意事项

- Model 只定义数据结构，不包含业务逻辑
- 不要在 Model 中直接操作数据库（应该在 Repository 层）
- 使用 `json:"-"` 防止敏感字段被序列化
- `AutoMigrate` 只会创建表和添加字段，不会删除字段
- 生产环境建议使用数据库迁移工具（如 golang-migrate）
