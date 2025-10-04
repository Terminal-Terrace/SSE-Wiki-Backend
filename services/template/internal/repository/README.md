# repository 目录

Repository 层（也称 DAO 层）负责数据访问，封装所有数据库操作。

## 职责

- 封装数据库的 CRUD 操作
- 提供数据查询接口
- 处理数据库事务
- 缓存操作（可选）
- 不包含业务逻辑

## 为什么需要 Repository？

1. **解耦**：将数据访问逻辑与业务逻辑分离
2. **复用**：同一个查询可以被多个 Service 使用
3. **测试**：可以通过接口 mock Repository 进行单元测试
4. **维护**：所有数据库操作集中管理，易于维护

## 目录结构

```
repository/
├── user_repository.go       # 用户数据访问
├── article_repository.go    # 文章数据访问
└── interfaces.go            # Repository 接口定义（可选）
```

## 示例代码

### 接口定义 (interfaces.go)

```go
package repository

import "terminal-terrace/template/internal/model"

// UserRepository 用户数据访问接口
type UserRepository interface {
    Create(user *model.User) error
    GetByID(id uint) (*model.User, error)
    GetByUsername(username string) (*model.User, error)
    GetByEmail(email string) (*model.User, error)
    Update(user *model.User) error
    Delete(id uint) error
    List(page, pageSize int) ([]*model.User, int64, error)
}
```

### Repository 实现 (user_repository.go)

```go
package repository

import (
    "gorm.io/gorm"
    "terminal-terrace/template/internal/model"
)

type userRepository struct {
    db *gorm.DB
}

// NewUserRepository 构造函数
func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        db: db,
    }
}

// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
    return r.db.Create(user).Error
}

// GetByID 根据 ID 查询用户
func (r *userRepository) GetByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetByUsername 根据用户名查询
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
    var user model.User
    err := r.db.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetByEmail 根据邮箱查询
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
    var user model.User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Update 更新用户信息
func (r *userRepository) Update(user *model.User) error {
    return r.db.Save(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uint) error {
    return r.db.Delete(&model.User{}, id).Error
}

// List 分页查询用户列表
func (r *userRepository) List(page, pageSize int) ([]*model.User, int64, error) {
    var users []*model.User
    var total int64

    // 计算总数
    if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    offset := (page - 1) * pageSize
    err := r.db.Offset(offset).Limit(pageSize).Find(&users).Error
    if err != nil {
        return nil, 0, err
    }

    return users, total, nil
}
```

### 文章 Repository (article_repository.go)

```go
package repository

import (
    "gorm.io/gorm"
    "terminal-terrace/template/internal/model"
)

type ArticleRepository interface {
    Create(article *model.Article) error
    GetByID(id uint) (*model.Article, error)
    GetByModuleID(moduleID uint, page, pageSize int) ([]*model.Article, int64, error)
    Update(article *model.Article) error
    Delete(id uint) error
    IncrementViewCount(id uint) error
}

type articleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
    return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *model.Article) error {
    return r.db.Create(article).Error
}

func (r *articleRepository) GetByID(id uint) (*model.Article, error) {
    var article model.Article
    // 预加载关联数据
    err := r.db.Preload("Author").Preload("Module").First(&article, id).Error
    if err != nil {
        return nil, err
    }
    return &article, nil
}

func (r *articleRepository) GetByModuleID(moduleID uint, page, pageSize int) ([]*model.Article, int64, error) {
    var articles []*model.Article
    var total int64

    query := r.db.Model(&model.Article{}).Where("module_id = ?", moduleID)

    // 计算总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询并预加载作者信息
    offset := (page - 1) * pageSize
    err := query.Preload("Author").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&articles).Error

    if err != nil {
        return nil, 0, err
    }

    return articles, total, nil
}

func (r *articleRepository) Update(article *model.Article) error {
    return r.db.Save(article).Error
}

func (r *articleRepository) Delete(id uint) error {
    return r.db.Delete(&model.Article{}, id).Error
}

// IncrementViewCount 增加浏览次数
func (r *articleRepository) IncrementViewCount(id uint) error {
    return r.db.Model(&model.Article{}).Where("id = ?", id).
        UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}
```

## GORM 常用操作

### 基础查询

```go
// 根据主键查询
db.First(&user, 1)                    // SELECT * FROM users WHERE id = 1;

// 根据条件查询单条
db.Where("username = ?", "alice").First(&user)

// 查询多条
db.Where("role = ?", "admin").Find(&users)

// 查询所有
db.Find(&users)
```

### 条件查询

```go
// WHERE
db.Where("name = ?", "alice").Find(&users)
db.Where("age > ?", 18).Find(&users)
db.Where("name IN ?", []string{"alice", "bob"}).Find(&users)

// AND
db.Where("role = ? AND age > ?", "admin", 18).Find(&users)

// OR
db.Where("role = ?", "admin").Or("age > ?", 18).Find(&users)

// NOT
db.Not("role = ?", "admin").Find(&users)
```

### 排序和分页

```go
// 排序
db.Order("created_at DESC").Find(&articles)
db.Order("age DESC, name ASC").Find(&users)

// 分页
db.Offset(10).Limit(20).Find(&users)  // LIMIT 20 OFFSET 10
```

### 预加载关联

```go
// 预加载单个关联
db.Preload("Author").Find(&articles)

// 预加载多个关联
db.Preload("Author").Preload("Module").Find(&articles)

// 条件预加载
db.Preload("Articles", "status = ?", "published").Find(&users)
```

### 创建和更新

```go
// 创建
db.Create(&user)

// 批量创建
db.Create(&users)

// 更新所有字段
db.Save(&user)

// 更新指定字段
db.Model(&user).Update("name", "alice")
db.Model(&user).Updates(map[string]interface{}{"name": "alice", "age": 20})

// 更新多条记录
db.Model(&User{}).Where("role = ?", "user").Update("active", true)
```

### 删除

```go
// 软删除（有 DeletedAt 字段）
db.Delete(&user, 1)

// 永久删除
db.Unscoped().Delete(&user, 1)

// 批量删除
db.Where("age < ?", 18).Delete(&User{})
```

### 事务处理

```go
// 自动事务
func (r *userRepository) CreateWithProfile(user *model.User, profile *model.Profile) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        profile.UserID = user.ID
        if err := tx.Create(profile).Error; err != nil {
            return err
        }

        return nil
    })
}

// 手动事务
tx := r.db.Begin()

if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}

if err := tx.Create(&profile).Error; err != nil {
    tx.Rollback()
    return err
}

tx.Commit()
```

### 原生 SQL

```go
// 执行原生 SQL
db.Exec("UPDATE users SET age = age + 1 WHERE role = ?", "admin")

// 原生查询
db.Raw("SELECT * FROM users WHERE age > ?", 18).Scan(&users)
```

## 设计原则

1. **单一职责**：每个 Repository 只负责一个实体的数据访问
2. **接口抽象**：使用接口定义 Repository，便于测试和替换实现
3. **依赖注入**：通过构造函数注入 `*gorm.DB`
4. **无业务逻辑**：Repository 只做数据存取，不包含业务判断
5. **错误透传**：将数据库错误返回给上层处理

## Repository vs Service

| Repository | Service |
|-----------|---------|
| 数据访问 | 业务逻辑 |
| 单表操作为主 | 可能涉及多表 |
| 返回 Model | 返回 DTO |
| 返回 error | 返回 BusinessError |

## 注意事项

- Repository 不应该包含业务逻辑
- 复杂查询应该在 Repository 中封装，而不是在 Service 中直接写 GORM 代码
- 使用 `Preload` 避免 N+1 查询问题
- 事务处理应该在 Service 层控制，Repository 提供支持
- 错误处理：将 `gorm.ErrRecordNotFound` 等数据库错误返回给 Service 层处理
