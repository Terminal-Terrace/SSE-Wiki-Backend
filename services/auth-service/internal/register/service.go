package register

import (
	"regexp"

	"terminal-terrace/auth-service/internal/code"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model/user"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/response"

	"golang.org/x/crypto/bcrypt"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	upperRegex    = regexp.MustCompile(`[A-Z]`)
	lowerRegex    = regexp.MustCompile(`[a-z]`)
	digitRegex    = regexp.MustCompile(`[0-9]`)
)

type RegisterService struct{}

// 只支持账号密码注册
func (s *RegisterService) Register(req RegisterRequest) (RegisterResponse, *response.BusinessError) {
	// 1. 参数校验
	if err := s.validateRequest(req); err != nil {
		return RegisterResponse{}, err
	}

	// 2. 检查用户名和邮箱是否已存在
	var existingUser user.User
	if err := database.PostgresDB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		if existingUser.Username == req.Username {
			return RegisterResponse{}, response.NewBusinessError(
				response.WithErrorCode(response.Fail),
				response.WithErrorMessage("用户名已存在"),
			)
		}
		if existingUser.Email == req.Email {
			return RegisterResponse{}, response.NewBusinessError(
				response.WithErrorCode(response.Fail),
				response.WithErrorMessage("邮箱已被注册"),
			)
		}
	}

	// 3. 检查验证码
	if err := code.VerifyEmailCode(req.Email, code.CodeTypeRegister, req.Code); err != nil {
		return RegisterResponse{}, err
	}

	// 4. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("密码加密失败"),
		)
	}

	// 5. 创建用户
	newUser := user.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := database.PostgresDB.Create(&newUser).Error; err != nil {
		return RegisterResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("用户创建失败"),
		)
	}

	// 6. 生成 refresh token
	refreshToken, err := pkg.GenerateRefreshToken(newUser.ID, newUser.Username, newUser.Email)
	if err != nil {
		return RegisterResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成刷新令牌失败"),
		)
	}

	// 7. 返回结果
	return RegisterResponse{
		RefreshToken: refreshToken,
		RedirectUrl:  "/",
	}, nil
}

// 参数校验
func (s *RegisterService) validateRequest(req RegisterRequest) *response.BusinessError {
	// 校验用户名
	if req.Username == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("用户名不能为空"),
		)
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("用户名长度必须在3-50个字符之间"),
		)
	}
	if !usernameRegex.MatchString(req.Username) {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("用户名只能包含字母、数字和下划线"),
		)
	}

	// 校验邮箱
	if req.Email == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("邮箱不能为空"),
		)
	}
	if !emailRegex.MatchString(req.Email) {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("邮箱格式不正确"),
		)
	}

	// 校验密码
	if req.Password == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("密码不能为空"),
		)
	}
	if len(req.Password) < 6 || len(req.Password) > 100 {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("密码长度必须在6-100个字符之间"),
		)
	}

	// 校验确认密码
	if req.ConfirmPassword != req.Password {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("两次密码输入不一致"),
		)
	}

	// 校验密码强度
	if !s.isStrongPassword(req.Password) {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("密码强度不足，需包含大小写字母、数字"),
		)
	}

	return nil
}

// 密码强度校验
func (s *RegisterService) isStrongPassword(password string) bool {
	hasUpper := upperRegex.MatchString(password)
	hasLower := lowerRegex.MatchString(password)
	hasDigit := digitRegex.MatchString(password)

	return hasUpper && hasLower && hasDigit
}
