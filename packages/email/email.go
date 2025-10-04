package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// Config 邮件服务配置
type Config struct {
	Host     string `koanf:"host"`     // SMTP 服务器地址，如 smtp.gmail.com
	Port     int    `koanf:"port"`     // SMTP 端口，通常 587 (TLS) 或 465 (SSL)
	Username string `koanf:"username"` // 发件人邮箱
	Password string `koanf:"password"` // 邮箱密码或授权码
	UseTLS   bool   `koanf:"tls"`      // 是否使用 TLS，默认 true
}

// Message 邮件消息
type Message struct {
	From        string   // 发件人显示名称，如 "SSE Wiki <noreply@example.com>"
	To          []string // 收件人列表
	Cc          []string // 抄送列表
	Bcc         []string // 密送列表
	Subject     string   // 邮件主题
	Body        string   // 邮件正文（纯文本或 HTML）
	ContentType string   // 内容类型，默认 "text/plain"，可设为 "text/html"
}

// Client 邮件客户端
type Client struct {
	config *Config
}

// NewClient 创建邮件客户端
func NewClient(config *Config) *Client {
	// 设置默认端口
	if config.Port == 0 {
		config.Port = 587
	}
	return &Client{config: config}
}

// Send 发送邮件
func (c *Client) Send(msg *Message) error {
	if msg.From == "" {
		return fmt.Errorf("发件人不能为空")
	}
	if len(msg.To) == 0 {
		return fmt.Errorf("收件人不能为空")
	}
	if msg.Subject == "" {
		return fmt.Errorf("邮件主题不能为空")
	}

	// 设置默认内容类型
	if msg.ContentType == "" {
		msg.ContentType = "text/plain; charset=UTF-8"
	}

	// 构建邮件内容
	headers := make(map[string]string)
	headers["From"] = msg.From
	headers["To"] = strings.Join(msg.To, ", ")
	if len(msg.Cc) > 0 {
		headers["Cc"] = strings.Join(msg.Cc, ", ")
	}
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = msg.ContentType

	// 组装邮件
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	// 收集所有收件人
	recipients := append([]string{}, msg.To...)
	recipients = append(recipients, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)

	// 发送邮件
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// 根据配置选择是否使用 TLS
	if c.config.UseTLS || c.config.Port == 587 {
		return c.sendWithTLS(addr, auth, msg.From, recipients, []byte(message))
	}

	return smtp.SendMail(addr, auth, msg.From, recipients, []byte(message))
}

// sendWithTLS 使用 TLS 发送邮件
func (c *Client) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// 连接到 SMTP 服务器
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	defer client.Close()

	// 发送 STARTTLS 命令
	if err = client.StartTLS(&tls.Config{ServerName: c.config.Host}); err != nil {
		return fmt.Errorf("启动 TLS 失败: %w", err)
	}

	// 认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %w", err)
	}

	// 设置发件人
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("准备发送邮件内容失败: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("关闭邮件内容写入失败: %w", err)
	}

	return client.Quit()
}

// SendSimple 发送简单文本邮件（便捷方法）
func (c *Client) SendSimple(from string, to string, subject string, body string) error {
	return c.Send(&Message{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Body:    body,
	})
}

// SendHTML 发送 HTML 邮件（便捷方法）
func (c *Client) SendHTML(from string, to string, subject string, htmlBody string) error {
	return c.Send(&Message{
		From:        from,
		To:          []string{to},
		Subject:     subject,
		Body:        htmlBody,
		ContentType: "text/html; charset=UTF-8",
	})
}
