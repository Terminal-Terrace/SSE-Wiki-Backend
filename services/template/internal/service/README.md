# service 目录

Service 层是核心业务逻辑层，处理所有业务规则和流程。

## 职责

- 实现核心业务逻辑
- 调用 Repository 层进行数据操作
- 处理业务规则验证
- 控制事务边界
- 返回业务错误

## 目录结构

```
service/
├── user_service.go       # 用户业务逻辑
├── article_service.go    # 文章业务逻辑
└── auth_service.go       # 认证业务逻辑
```

## 示例代码

### 用户 Service (user_service.go)

```go
package service

import (
    "errors"
    "gorm.io/gorm"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
    "terminal-terrace/template/internal/model"
    "terminal-terrace/template/internal/repository"
)

type UserService struct {
    userRepo repository.UserRepository
}

// NewUserService 构造函数（依赖注入）
func NewUserService(userRepo repository.UserRepository) *UserService {
    return &UserService{
        userRepo: userRepo,
    }
}

// CreateUser 创建用户
func (s *UserService) CreateUser(req *dto.CreateUserRequest) (*model.User, *response.BusinessError) {
    // 1. 业务规则验证：检查用户名是否已存在
    existingUser, err := s.userRepo.GetByUsername(req.Username)
    if err == nil && existingUser != nil {
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.Conflict),
            response.WithErrorMessage("用户名已存在"),
        )
    }

    // 2. 检查邮箱是否已存在
    existingUser, err = s.userRepo.GetByEmail(req.Email)
    if err == nil && existingUser != nil {
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.Conflict),
            response.WithErrorMessage("邮箱已被使用"),
        )
    }

    // 3. 创建用户对象
    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashPassword(req.Password), // 密码加密
    }

    // 4. 保存到数据库
    if err := s.userRepo.Create(user); err != nil {
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("创建用户失败"),
            response.WithError(err),
        )
    }

    return user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id string) (*model.User, *response.BusinessError) {
    userID := parseID(id) // 转换 ID

    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("用户不存在"),
            )
        }
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询用户失败"),
            response.WithError(err),
        )
    }

    return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id string, req *dto.UpdateUserRequest) (*model.User, *response.BusinessError) {
    userID := parseID(id)

    // 1. 查询用户是否存在
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("用户不存在"),
            )
        }
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询用户失败"),
        )
    }

    // 2. 更新字段（只更新提供的字段）
    if req.Username != nil {
        // 检查用户名是否已被其他用户使用
        existingUser, _ := s.userRepo.GetByUsername(*req.Username)
        if existingUser != nil && existingUser.ID != userID {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.Conflict),
                response.WithErrorMessage("用户名已被使用"),
            )
        }
        user.Username = *req.Username
    }

    if req.Email != nil {
        // 检查邮箱是否已被其他用户使用
        existingUser, _ := s.userRepo.GetByEmail(*req.Email)
        if existingUser != nil && existingUser.ID != userID {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.Conflict),
                response.WithErrorMessage("邮箱已被使用"),
            )
        }
        user.Email = *req.Email
    }

    // 3. 保存更新
    if err := s.userRepo.Update(user); err != nil {
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("更新用户失败"),
            response.WithError(err),
        )
    }

    return user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id string) *response.BusinessError {
    userID := parseID(id)

    // 1. 检查用户是否存在
    _, err := s.userRepo.GetByID(userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("用户不存在"),
            )
        }
        return response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询用户失败"),
        )
    }

    // 2. 执行删除
    if err := s.userRepo.Delete(userID); err != nil {
        return response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("删除用户失败"),
            response.WithError(err),
        )
    }

    return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(page, pageSize string) ([]*model.User, int64, *response.BusinessError) {
    pageNum := parsePageNum(page)
    pageSizeNum := parsePageSize(pageSize)

    users, total, err := s.userRepo.List(pageNum, pageSizeNum)
    if err != nil {
        return nil, 0, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询用户列表失败"),
            response.WithError(err),
        )
    }

    return users, total, nil
}

// 辅助函数
func hashPassword(password string) string {
    // TODO: 实现密码加密逻辑（如 bcrypt）
    return password
}

func parseID(id string) uint {
    // TODO: 实现 ID 转换和验证
    return 0
}

func parsePageNum(page string) int {
    // TODO: 实现分页参数解析
    return 1
}

func parsePageSize(pageSize string) int {
    // TODO: 实现分页参数解析
    return 20
}
```

### 文章 Service (article_service.go)

```go
package service

import (
    "errors"
    "gorm.io/gorm"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
    "terminal-terrace/template/internal/model"
    "terminal-terrace/template/internal/repository"
)

type ArticleService struct {
    articleRepo repository.ArticleRepository
    userRepo    repository.UserRepository
}

func NewArticleService(
    articleRepo repository.ArticleRepository,
    userRepo repository.UserRepository,
) *ArticleService {
    return &ArticleService{
        articleRepo: articleRepo,
        userRepo:    userRepo,
    }
}

// CreateArticle 创建文章
func (s *ArticleService) CreateArticle(req *dto.CreateArticleRequest, authorID uint) (*model.Article, *response.BusinessError) {
    // 1. 验证作者是否存在
    _, err := s.userRepo.GetByID(authorID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("作者不存在"),
            )
        }
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("验证作者失败"),
        )
    }

    // 2. 创建文章对象
    article := &model.Article{
        Title:    req.Title,
        Content:  req.Content,
        Summary:  req.Summary,
        AuthorID: authorID,
        ModuleID: req.ModuleID,
    }

    // 3. 保存文章
    if err := s.articleRepo.Create(article); err != nil {
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("创建文章失败"),
            response.WithError(err),
        )
    }

    return article, nil
}

// GetArticleByID 获取文章详情
func (s *ArticleService) GetArticleByID(id uint) (*model.Article, *response.BusinessError) {
    article, err := s.articleRepo.GetByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("文章不存在"),
            )
        }
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询文章失败"),
        )
    }

    // 增加浏览次数
    _ = s.articleRepo.IncrementViewCount(id)

    return article, nil
}

// GetArticlesByModuleID 获取模块下的文章列表
func (s *ArticleService) GetArticlesByModuleID(moduleID uint, page, pageSize int) ([]*model.Article, int64, *response.BusinessError) {
    articles, total, err := s.articleRepo.GetByModuleID(moduleID, page, pageSize)
    if err != nil {
        return nil, 0, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询文章列表失败"),
            response.WithError(err),
        )
    }

    return articles, total, nil
}
```

## Service 层的典型结构

```go
func (s *XxxService) DoSomething(req *dto.Request) (*model.Result, *response.BusinessError) {
    // 1. 业务规则验证
    if err := s.validateBusinessRules(req); err != nil {
        return nil, err
    }

    // 2. 调用 Repository 层
    result, err := s.repo.Query(req)
    if err != nil {
        return nil, s.handleRepositoryError(err)
    }

    // 3. 业务逻辑处理
    processed := s.processBusinessLogic(result)

    // 4. 返回结果
    return processed, nil
}
```

## 事务处理

### 简单事务

```go
func (s *UserService) CreateUserWithProfile(req *dto.CreateUserRequest) *response.BusinessError {
    return s.userRepo.Transaction(func(tx *gorm.DB) error {
        // 1. 创建用户
        user := &model.User{...}
        if err := tx.Create(user).Error; err != nil {
            return err
        }

        // 2. 创建用户资料
        profile := &model.Profile{
            UserID: user.ID,
            ...
        }
        if err := tx.Create(profile).Error; err != nil {
            return err
        }

        return nil
    })
}
```

### 复杂事务（跨多个 Repository）

```go
// 在 Service 中管理事务
func (s *ArticleService) PublishArticle(articleID uint, userID uint) *response.BusinessError {
    tx := s.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 1. 更新文章状态
    if err := tx.Model(&model.Article{}).Where("id = ?", articleID).
        Update("status", "published").Error; err != nil {
        tx.Rollback()
        return response.NewBusinessError(...)
    }

    // 2. 增加用户积分
    if err := tx.Model(&model.User{}).Where("id = ?", userID).
        Update("points", gorm.Expr("points + ?", 10)).Error; err != nil {
        tx.Rollback()
        return response.NewBusinessError(...)
    }

    // 3. 提交事务
    if err := tx.Commit().Error; err != nil {
        return response.NewBusinessError(...)
    }

    return nil
}
```

## 错误处理模式

```go
func (s *UserService) GetUser(id uint) (*model.User, *response.BusinessError) {
    user, err := s.userRepo.GetByID(id)
    if err != nil {
        // 区分不同的错误类型
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, response.NewBusinessError(
                response.WithErrorCode(response.NotFound),
                response.WithErrorMessage("用户不存在"),
            )
        }
        return nil, response.NewBusinessError(
            response.WithErrorCode(response.DatabaseError),
            response.WithErrorMessage("查询失败"),
            response.WithError(err),
        )
    }
    return user, nil
}
```

## 设计原则

1. **核心业务逻辑在此层**：所有业务规则、验证、流程控制都在 Service 层
2. **依赖 Repository 而非直接依赖 DB**：通过 Repository 接口操作数据
3. **返回 BusinessError**：不要返回原始 error，应该包装为业务错误
4. **事务边界控制**：复杂事务应该在 Service 层控制
5. **依赖注入**：通过构造函数注入依赖

## Service vs Handler vs Repository

| 层级 | 职责 | 返回类型 |
|-----|------|---------|
| Handler | HTTP 请求处理、参数验证 | 调用 Service，返回 HTTP 响应 |
| Service | 业务逻辑、业务规则验证 | 返回 Model 和 BusinessError |
| Repository | 数据访问 | 返回 Model 和 error |

## 注意事项

- Service 应该包含所有业务逻辑，不要把业务逻辑放在 Handler 或 Repository
- 使用 `BusinessError` 而不是原始 error
- 复杂的查询应该封装在 Repository 中
- 事务处理优先在 Service 层控制
- 避免 Service 之间的循环依赖
- Service 层应该是无状态的（除了依赖注入的 Repository）
