# pkg 目录

pkg 目录存放可以被外部包导入的公共工具和库。

## 职责

- 提供可复用的工具函数
- 定义公共常量和配置
- 不包含业务逻辑
- 可以被其他服务导入

## pkg vs internal 的区别

| 特性 | pkg | internal |
|-----|-----|----------|
| 可见性 | 可被外部导入 | 只能被父包及其子包导入 |
| 用途 | 通用工具、公共库 | 服务内部实现 |
| 业务逻辑 | 不包含 | 包含 |
| 示例 | 工具函数、加密、验证 | Handler, Service, Repository |

## 典型的 pkg 内容

```
pkg/
├── utils/           # 通用工具函数
│   ├── string.go    # 字符串处理
│   ├── time.go      # 时间处理
│   └── validator.go # 数据验证
├── crypto/          # 加密相关
│   ├── hash.go      # 哈希函数
│   └── jwt.go       # JWT 工具
├── pagination/      # 分页工具
│   └── paginate.go
└── constants/       # 公共常量
    └── constants.go
```

## 示例代码

### 字符串工具 (utils/string.go)

```go
package utils

import (
    "regexp"
    "strings"
)

// IsValidUsername 验证用户名格式
func IsValidUsername(username string) bool {
    if len(username) < 3 || len(username) > 20 {
        return false
    }
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
    return matched
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, email)
    return matched
}

// TruncateString 截断字符串
func TruncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}

// ToSnakeCase 转换为蛇形命名
func ToSnakeCase(s string) string {
    var result strings.Builder
    for i, r := range s {
        if i > 0 && r >= 'A' && r <= 'Z' {
            result.WriteRune('_')
        }
        result.WriteRune(r)
    }
    return strings.ToLower(result.String())
}
```

### 时间工具 (utils/time.go)

```go
package utils

import (
    "time"
)

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
    return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
    return t.Format("2006-01-02")
}

// ParseTime 解析时间字符串
func ParseTime(s string) (time.Time, error) {
    return time.Parse("2006-01-02 15:04:05", s)
}

// IsToday 判断是否为今天
func IsToday(t time.Time) bool {
    now := time.Now()
    return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// DaysBetween 计算两个日期之间的天数
func DaysBetween(start, end time.Time) int {
    return int(end.Sub(start).Hours() / 24)
}

// TimeAgo 返回友好的时间描述（如"3分钟前"）
func TimeAgo(t time.Time) string {
    duration := time.Since(t)

    if duration < time.Minute {
        return "刚刚"
    }
    if duration < time.Hour {
        return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
    }
    if duration < 24*time.Hour {
        return fmt.Sprintf("%d小时前", int(duration.Hours()))
    }
    if duration < 30*24*time.Hour {
        return fmt.Sprintf("%d天前", int(duration.Hours()/24))
    }
    return t.Format("2006-01-02")
}
```

### 密码加密 (crypto/hash.go)

```go
package crypto

import (
    "golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
```

### JWT 工具 (crypto/jwt.go)

```go
package crypto

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(userID uint, username, role, secret string, expireHours int) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expireHours))),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// ParseToken 解析 JWT token
func ParseToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

### 分页工具 (pagination/paginate.go)

```go
package pagination

import (
    "strconv"
)

type Pagination struct {
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
    Total    int64 `json:"total"`
    Pages    int   `json:"pages"`
}

// NewPagination 创建分页对象
func NewPagination(page, pageSize int, total int64) *Pagination {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 100 {
        pageSize = 100
    }

    pages := int(total) / pageSize
    if int(total)%pageSize != 0 {
        pages++
    }

    return &Pagination{
        Page:     page,
        PageSize: pageSize,
        Total:    total,
        Pages:    pages,
    }
}

// Offset 计算数据库查询的偏移量
func (p *Pagination) Offset() int {
    return (p.Page - 1) * p.PageSize
}

// ParsePageParams 从字符串解析分页参数
func ParsePageParams(pageStr, pageSizeStr string) (int, int) {
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }

    pageSize, err := strconv.Atoi(pageSizeStr)
    if err != nil || pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 100 {
        pageSize = 100
    }

    return page, pageSize
}
```

### 公共常量 (constants/constants.go)

```go
package constants

const (
    // 用户角色
    RoleAdmin     = "admin"
    RoleModerator = "moderator"
    RoleUser      = "user"

    // 文章状态
    ArticleDraft     = "draft"
    ArticlePublished = "published"
    ArticleArchived  = "archived"

    // 默认值
    DefaultPageSize = 20
    MaxPageSize     = 100
)
```

## 使用示例

### 在 Service 中使用 pkg

```go
package service

import (
    "terminal-terrace/template/pkg/crypto"
    "terminal-terrace/template/pkg/utils"
)

func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
    // 使用密码加密工具
    hashedPassword, err := crypto.HashPassword(req.Password)
    if err != nil {
        return err
    }

    // 使用字符串验证工具
    if !utils.IsValidUsername(req.Username) {
        return errors.New("invalid username")
    }

    // ...
}
```

### 在其他服务中使用

```go
// 在其他服务中也可以导入使用
import "terminal-terrace/template/pkg/crypto"

password := "secret123"
hashed, _ := crypto.HashPassword(password)
```

## 设计原则

1. **通用性**：pkg 中的代码应该是通用的，不包含业务逻辑
2. **无依赖**：pkg 不应该依赖 internal 中的包
3. **可测试**：提供纯函数，易于测试
4. **文档完善**：公共 API 应该有清晰的注释
5. **向后兼容**：pkg 的修改应该考虑向后兼容性

## pkg vs packages (项目根目录的 packages)

| 位置 | 用途 | 作用域 |
|-----|------|--------|
| `services/xxx/pkg/` | 服务级公共工具 | 可被当前服务和其他服务导入 |
| `packages/` | 项目级共享包 | 所有服务共享（如 auth-sdk, response） |

**选择建议**：
- 只在当前服务使用 → `services/xxx/pkg/`
- 多个服务共享 → `packages/`

## 注意事项

- pkg 中的包应该是稳定的、经过充分测试的
- 不要在 pkg 中依赖 internal 的代码
- pkg 的修改可能影响多个地方，要谨慎
- 相同功能的工具应该合并到一个包中
- 考虑使用 `packages/` 目录来共享跨服务的通用代码
