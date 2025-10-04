# Email Package

邮件发送工具包，提供统一的邮件发送能力。

## 功能特性

- ✅ 支持 SMTP 协议发送邮件
- ✅ 支持 TLS/SSL 加密连接
- ✅ 支持纯文本和 HTML 格式邮件
- ✅ 支持邮件模板（内置验证码、欢迎邮件等模板）
- ✅ 支持多收件人、抄送、密送
- ✅ 提供便捷方法简化常用场景

## 快速开始

### 1. 基本使用

```go
import "terminal-terrace/email"

// 创建邮件客户端
client := email.NewClient(&email.Config{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-app-password",
    From:     "SSE Wiki <your-email@gmail.com>",
})

// 发送简单文本邮件
err := client.SendSimple(
    "recipient@example.com",
    "测试邮件",
    "这是一封测试邮件",
)
```

### 2. 发送 HTML 邮件

```go
htmlContent := `
<html>
<body>
    <h1>欢迎</h1>
    <p>这是一封 HTML 邮件</p>
</body>
</html>
`

err := client.SendHTML(
    "recipient@example.com",
    "HTML 邮件",
    htmlContent,
)
```

### 3. 发送验证码邮件

```go
// 使用内置验证码模板
err := client.SendVerificationCode(
    "user@example.com",
    "123456",  // 验证码
    5,         // 过期时间（分钟）
)
```

### 4. 使用自定义模板

```go
// 创建模板
tmpl, err := email.NewTemplate(`
<html>
<body>
    <h1>Hello {{.Name}}</h1>
    <p>Your order ID is: {{.OrderID}}</p>
</body>
</html>
`)

// 渲染并发送
data := map[string]interface{}{
    "Name":    "张三",
    "OrderID": "12345",
}

err = client.SendWithTemplate(
    "user@example.com",
    "订单确认",
    tmpl,
    data,
)
```

### 5. 发送给多个收件人

```go
err := client.Send(&email.Message{
    To:      []string{"user1@example.com", "user2@example.com"},
    Cc:      []string{"manager@example.com"},
    Subject: "团队通知",
    Body:    "这是一条团队通知",
})
```

## 配置说明

### Config 结构

```go
type Config struct {
    Host     string // SMTP 服务器地址
    Port     int    // SMTP 端口（默认 587）
    Username string // 发件人邮箱
    Password string // 邮箱密码或授权码
    From     string // 发件人显示名称（默认使用 Username）
    UseTLS   bool   // 是否使用 TLS（默认 true）
}
```

### 常见 SMTP 配置

**Gmail:**
```go
Config{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "your-email@gmail.com",
    Password: "your-app-password", // 需要开启两步验证并生成应用专用密码
}
```

**QQ 邮箱:**
```go
Config{
    Host:     "smtp.qq.com",
    Port:     587,
    Username: "your-qq@qq.com",
    Password: "your-authorization-code", // 需要在 QQ 邮箱设置中获取授权码
}
```

**163 邮箱:**
```go
Config{
    Host:     "smtp.163.com",
    Port:     25,
    Username: "your-email@163.com",
    Password: "your-authorization-code", // 需要在邮箱设置中获取授权码
}
```

**Outlook:**
```go
Config{
    Host:     "smtp-mail.outlook.com",
    Port:     587,
    Username: "your-email@outlook.com",
    Password: "your-password",
}
```

## 环境变量配置

在服务的 `config.yaml` 中添加邮件配置：

```yaml
email:
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: ${EMAIL_PASSWORD}  # 建议使用环境变量
  from: "SSE Wiki <your-email@gmail.com>"
```

可以通过环境变量覆盖：
```bash
export APP_EMAIL_PASSWORD="your-app-password"
```

## 内置模板

### 验证码模板

```go
client.SendVerificationCode("user@example.com", "123456", 5)
```

### 欢迎邮件模板

```go
tmpl, _ := email.NewTemplate(email.WelcomeTemplate)
data := email.WelcomeData{
    AppName:    "SSE Wiki",
    Username:   "张三",
    Message:    "感谢您的注册，现在可以开始使用我们的服务了。",
    ActionURL:  "https://example.com/login",
    ActionText: "立即登录",
}
client.SendWithTemplate("user@example.com", "欢迎加入", tmpl, data)
```

## 在服务中使用

### 1. 在 config 中添加邮件配置

```go
// config/config.go
type Config struct {
    // ... 其他配置
    Email EmailConfig `koanf:"email"`
}

type EmailConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    Username string `koanf:"username"`
    Password string `koanf:"password"`
    From     string `koanf:"from"`
}
```

### 2. 初始化邮件客户端

```go
// cmd/server/main.go
import "terminal-terrace/email"

func main() {
    config.Load("config.yaml")

    // 初始化邮件客户端
    emailClient := email.NewClient(&email.Config{
        Host:     config.Conf.Email.Host,
        Port:     config.Conf.Email.Port,
        Username: config.Conf.Email.Username,
        Password: config.Conf.Email.Password,
        From:     config.Conf.Email.From,
    })

    // 将 emailClient 传递给需要发送邮件的服务
    // ...
}
```

### 3. 在业务代码中使用

```go
// internal/service/auth_service.go
type AuthService struct {
    userRepo    *repository.UserRepository
    emailClient *email.Client
}

func (s *AuthService) SendVerificationEmail(userEmail string) error {
    // 生成验证码
    code := generateCode()

    // 存储验证码到 Redis
    // ...

    // 发送邮件
    return s.emailClient.SendVerificationCode(userEmail, code, 5)
}
```

## 错误处理

所有发送方法都返回 `error`，应当进行适当的错误处理：

```go
err := client.SendSimple("user@example.com", "测试", "内容")
if err != nil {
    log.Printf("发送邮件失败: %v", err)
    // 根据业务需求决定是否重试或返回错误
}
```

常见错误：
- 配置错误（服务器地址、端口、认证信息）
- 网络连接问题
- SMTP 服务器拒绝（发件人/收件人地址无效、超出配额等）
- 模板渲染错误

## 最佳实践

1. **安全性**
   - 不要在代码中硬编码邮箱密码，使用环境变量
   - 使用应用专用密码而不是邮箱登录密码
   - 对于生产环境，考虑使用专业的邮件服务（如 SendGrid、AWS SES）

2. **性能优化**
   - 发送邮件是 I/O 操作，考虑使用异步队列
   - 对于批量发送，使用消息队列避免阻塞主线程

3. **用户体验**
   - 邮件标题和内容清晰明确
   - 使用美观的 HTML 模板
   - 提供退订链接（如果是营销邮件）

4. **监控和日志**
   - 记录发送成功/失败的日志
   - 监控发送失败率
   - 设置告警机制

## 测试

建议在测试环境中使用真实的邮箱进行测试，或使用 Mailtrap、MailHog 等邮件测试工具。

```go
// 在测试中使用
func TestEmail(t *testing.T) {
    client := email.NewClient(&email.Config{
        Host:     "smtp.mailtrap.io",
        Port:     2525,
        Username: "test-username",
        Password: "test-password",
    })

    err := client.SendSimple("test@example.com", "测试", "内容")
    assert.NoError(t, err)
}
```
