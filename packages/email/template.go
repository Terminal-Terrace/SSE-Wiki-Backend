package email

import (
	"bytes"
	"fmt"
	"html/template"
)

// Template 邮件模板
type Template struct {
	tmpl *template.Template
}

// NewTemplate 从 HTML 字符串创建模板
func NewTemplate(htmlContent string) (*Template, error) {
	tmpl, err := template.New("email").Parse(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("解析邮件模板失败: %w", err)
	}
	return &Template{tmpl: tmpl}, nil
}

// NewTemplateFromFile 从文件创建模板
func NewTemplateFromFile(filePath string) (*Template, error) {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return nil, fmt.Errorf("解析邮件模板文件失败: %w", err)
	}
	return &Template{tmpl: tmpl}, nil
}

// Render 渲染模板
func (t *Template) Render(data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := t.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染邮件模板失败: %w", err)
	}
	return buf.String(), nil
}

// SendWithTemplate 使用模板发送邮件
func (c *Client) SendWithTemplate(to string, subject string, tmpl *Template, data interface{}) error {
	body, err := tmpl.Render(data)
	if err != nil {
		return err
	}
	return c.SendHTML(to, subject, body)
}

// 预定义常用邮件模板

// VerificationCodeTemplate 验证码邮件模板
const VerificationCodeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; }
        .code { font-size: 32px; font-weight: bold; color: #4CAF50; text-align: center;
                letter-spacing: 5px; padding: 20px; background-color: #fff; border: 2px dashed #4CAF50; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>{{.Message}}</p>
            <div class="code">{{.Code}}</div>
            <p>该验证码将在 {{.ExpireMinutes}} 分钟后过期，请尽快使用。</p>
            <p>如果这不是您的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`

// VerificationCodeData 验证码模板数据
type VerificationCodeData struct {
	Title         string // 邮件标题，如 "邮箱验证码"
	Message       string // 提示信息，如 "您正在进行邮箱验证，验证码为："
	Code          string // 验证码
	ExpireMinutes int    // 过期时间（分钟）
}

// SendVerificationCode 发送验证码邮件（便捷方法）
func (c *Client) SendVerificationCode(to string, code string, expireMinutes int) error {
	tmpl, err := NewTemplate(VerificationCodeTemplate)
	if err != nil {
		return err
	}

	data := VerificationCodeData{
		Title:         "邮箱验证码",
		Message:       "您正在进行邮箱验证，验证码为：",
		Code:          code,
		ExpireMinutes: expireMinutes,
	}

	return c.SendWithTemplate(to, "【SSE Wiki】邮箱验证码", tmpl, data)
}

// WelcomeTemplate 欢迎邮件模板
const WelcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; }
        .button { display: inline-block; padding: 12px 24px; background-color: #2196F3;
                  color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>欢迎加入 {{.AppName}}</h1>
        </div>
        <div class="content">
            <p>Hi {{.Username}}，</p>
            <p>欢迎注册 {{.AppName}}！您的账号已成功创建。</p>
            <p>{{.Message}}</p>
            {{if .ActionURL}}
            <div style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </div>
            {{end}}
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
        </div>
    </div>
</body>
</html>
`

// WelcomeData 欢迎邮件模板数据
type WelcomeData struct {
	AppName    string // 应用名称
	Username   string // 用户名
	Message    string // 欢迎信息
	ActionURL  string // 操作链接（可选）
	ActionText string // 操作按钮文字（可选）
}

// RegisterVerificationTemplate 注册验证码邮件模板
const RegisterVerificationTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #FF9800; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; }
        .code { font-size: 32px; font-weight: bold; color: #FF9800; text-align: center;
                letter-spacing: 5px; padding: 20px; background-color: #fff; border: 2px dashed #FF9800;
                margin: 20px 0; }
        .highlight { color: #FF9800; font-weight: bold; }
        .warning { background-color: #fff3cd; border-left: 4px solid #FF9800; padding: 12px;
                   margin: 20px 0; font-size: 14px; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 欢迎注册 SSE Wiki</h1>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>感谢您注册 <span class="highlight">SSE Wiki</span>！为了确保您的账号安全，请使用以下验证码完成注册：</p>
            <div class="code">{{.Code}}</div>
            <p style="text-align: center; color: #666; font-size: 14px;">
                验证码有效期：<span class="highlight">{{.ExpireMinutes}} 分钟</span>
            </p>
            <div class="warning">
                <strong>⚠️ 安全提示：</strong>
                <ul style="margin: 8px 0; padding-left: 20px;">
                    <li>请勿将验证码泄露给他人</li>
                    <li>SSE Wiki 工作人员不会向您索要验证码</li>
                    <li>如非本人操作，请忽略此邮件</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p style="margin-top: 10px;">© SSE Wiki - 软件学院知识共享平台</p>
        </div>
    </div>
</body>
</html>
`

// RegisterVerificationData 注册验证码模板数据
type RegisterVerificationData struct {
	Code          string // 验证码
	ExpireMinutes int    // 过期时间（分钟）
}

// SendRegisterVerificationCode 发送注册验证码邮件
func (c *Client) SendRegisterVerificationCode(to string, code string, expireMinutes int) error {
	tmpl, err := NewTemplate(RegisterVerificationTemplate)
	if err != nil {
		return err
	}

	data := RegisterVerificationData{
		Code:          code,
		ExpireMinutes: expireMinutes,
	}

	return c.SendWithTemplate(to, "【SSE Wiki】注册验证码", tmpl, data)
}

// ResetPasswordTemplate 重置密码验证码邮件模板
const ResetPasswordTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #F44336; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border: 1px solid #ddd; }
        .code { font-size: 32px; font-weight: bold; color: #F44336; text-align: center;
                letter-spacing: 5px; padding: 20px; background-color: #fff; border: 2px dashed #F44336;
                margin: 20px 0; }
        .highlight { color: #F44336; font-weight: bold; }
        .info-box { background-color: #e3f2fd; border-left: 4px solid #2196F3; padding: 12px;
                    margin: 20px 0; font-size: 14px; }
        .warning { background-color: #ffebee; border-left: 4px solid #F44336; padding: 12px;
                   margin: 20px 0; font-size: 14px; }
        .footer { text-align: center; padding: 20px; color: #888; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 密码重置验证</h1>
        </div>
        <div class="content">
            <p>您好，</p>
            <p>我们收到了您重置 <span class="highlight">SSE Wiki</span> 账号密码的请求。请使用以下验证码完成密码重置：</p>
            <div class="code">{{.Code}}</div>
            <p style="text-align: center; color: #666; font-size: 14px;">
                验证码有效期：<span class="highlight">{{.ExpireMinutes}} 分钟</span>
            </p>
            <div class="info-box">
                <strong>📋 操作步骤：</strong>
                <ol style="margin: 8px 0; padding-left: 20px;">
                    <li>返回密码重置页面</li>
                    <li>输入上方的验证码</li>
                    <li>设置新密码并确认</li>
                </ol>
            </div>
            <div class="warning">
                <strong>⚠️ 重要提醒：</strong>
                <ul style="margin: 8px 0; padding-left: 20px;">
                    <li>如果您没有申请重置密码，<span class="highlight">请立即忽略此邮件</span></li>
                    <li>为了您的账号安全，建议定期更换密码</li>
                    <li>请勿将验证码透露给任何人，包括 SSE Wiki 工作人员</li>
                    <li>完成密码重置后，所有设备将自动登出，需重新登录</li>
                </ul>
            </div>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复。</p>
            <p style="margin-top: 10px;">如有疑问，请联系我们的支持团队</p>
            <p style="margin-top: 10px;">© SSE Wiki - 软件学院知识共享平台</p>
        </div>
    </div>
</body>
</html>
`

// ResetPasswordData 重置密码验证码模板数据
type ResetPasswordData struct {
	Code          string // 验证码
	ExpireMinutes int    // 过期时间（分钟）
}

// SendResetPasswordCode 发送重置密码验证码邮件
func (c *Client) SendResetPasswordCode(to string, code string, expireMinutes int) error {
	tmpl, err := NewTemplate(ResetPasswordTemplate)
	if err != nil {
		return err
	}

	data := ResetPasswordData{
		Code:          code,
		ExpireMinutes: expireMinutes,
	}

	return c.SendWithTemplate(to, "【SSE Wiki】密码重置验证码", tmpl, data)
}
