package code

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model/user"
	"terminal-terrace/email"
	"terminal-terrace/response"
)

const (
	// 验证码有效期（分钟）
	CodeExpireMinutes = 5
	// 验证码长度
	CodeLength = 6
	// Redis Key 前缀
	RedisKeyPrefix = "auth:code:"
	from 		 = "wateringtop <wateringtop@qq.com>"
)

type CodeService struct {
	mailer *email.Client
}

func NewCodeService(mailer *email.Client) *CodeService {
	return &CodeService{mailer: mailer}
}

// generateCode 生成随机验证码
func generateCode() string {
	code := ""
	for i := 0; i < CodeLength; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

// storeCode 存储验证码到 Redis
func (s *CodeService) storeCode(email string, codeType CodeType, code string) *response.BusinessError {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%d", RedisKeyPrefix, email, codeType)

	err := database.RedisDB.Set(ctx, key, code, time.Duration(CodeExpireMinutes)*time.Minute).Err()
	if err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("存储验证码失败"),
			response.WithError(err),
		)
	}

	return nil
}

func (s *CodeService) register(email string) *response.BusinessError {
	// 1. 查看有没有这个邮箱, 有就报错
	var existingUser user.User
	if err := database.PostgresDB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("该邮箱已被注册"),
		)
	}

	// 2. 生成验证码
	code := generateCode()

	// 3. 存储验证码到 Redis
	if err := s.storeCode(email, CodeTypeRegister, code); err != nil {
		return err
	}

	// 4. 发送验证码邮件
	if err := s.mailer.SendRegisterVerificationCode(from, email, code, CodeExpireMinutes); err != nil {
		fmt.Println("Error sending email:", err)
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("发送验证码邮件失败"),
			response.WithError(err),
		)
	}

	return nil
}

func (s *CodeService) resetPassword(email string) *response.BusinessError {
	// 1. 查看有没有这个邮箱, 没有就报错
	var existingUser user.User
	if err := database.PostgresDB.Where("email = ?", email).First(&existingUser).Error; err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("该邮箱未注册"),
		)
	}

	// 2. 生成验证码
	code := generateCode()

	// 3. 存储验证码到 Redis
	if err := s.storeCode(email, CodeTypeResetPassword, code); err != nil {
		return err
	}

	// 4. 发送验证码邮件
	if err := s.mailer.SendResetPasswordCode(from, email, code, CodeExpireMinutes); err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("发送验证码邮件失败"),
			response.WithError(err),
		)
	}

	return nil
}

// TODO: 需要一个中间件, 限制同一 IP 或同一邮箱发送验证码的频率, 防止滥用
func (s *CodeService) SendCode(req SendCodeRequest) *response.BusinessError {
	switch req.Type {
	case CodeTypeRegister:
		// 注册逻辑
		return s.register(req.Email)
	case CodeTypeResetPassword:
		// 重置密码逻辑
		return s.resetPassword(req.Email)
	default:
		return response.NewBusinessError(response.WithErrorCode(response.InvalidParameter), response.WithErrorMessage("invalid code type"))
	}
}

// VerifyCode 验证验证码
func (s *CodeService) VerifyCode(email string, codeType CodeType, code string) *response.BusinessError {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%d", RedisKeyPrefix, email, codeType)

	// 从 Redis 获取验证码
	storedCode, err := database.RedisDB.Get(ctx, key).Result()
	if err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("验证码已过期或不存在"),
		)
	}

	// 验证验证码
	if storedCode != code {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("验证码错误"),
		)
	}

	// 验证成功后删除验证码
	database.RedisDB.Del(ctx, key)

	return nil
}

// VerifyEmailCode 验证邮箱验证码（提供给其他模块使用的便捷函数）
func VerifyEmailCode(email string, codeType CodeType, code string) *response.BusinessError {
	ctx := context.Background()
	key := fmt.Sprintf("%s%s:%d", RedisKeyPrefix, email, codeType)

	// 从 Redis 获取验证码
	storedCode, err := database.RedisDB.Get(ctx, key).Result()
	if err != nil {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("验证码已过期或不存在"),
		)
	}

	// 验证验证码
	if storedCode != code {
		return response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("验证码错误"),
		)
	}

	// 验证成功后删除验证码
	database.RedisDB.Del(ctx, key)

	return nil
}