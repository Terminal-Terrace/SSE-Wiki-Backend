package login

import (
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model/user"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/auth-service/internal/refresh"
	"terminal-terrace/response"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SSEWikiLoginService struct {
	refreshTokenRepo *refresh.RefreshTokenRepository
}

func init() {
	registerLoginService("sse-wiki", &SSEWikiLoginService{})
}

// getRefreshTokenRepo 延迟初始化 refreshTokenRepo
func (s *SSEWikiLoginService) getRefreshTokenRepo() *refresh.RefreshTokenRepository {
	if s.refreshTokenRepo == nil {
		s.refreshTokenRepo = refresh.NewRefreshTokenRepository(database.RedisDB)
	}
	return s.refreshTokenRepo
}

// 我们自己的登录服务, 使用账号密码登录
func (s *SSEWikiLoginService) Login(req LoginRequest) (LoginResponse, *response.BusinessError) {
	// 1. 检查参数
	if err := s.validateRequest(req); err != nil {
		return LoginResponse{}, err
	}

	// 2. 验证 state 并获取重定向地址
	redirectUrl, err := pkg.GetRedirectByState(req.State)
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("无效的 state 或 state 已过期"),
		)
	}

	// 3. 查询用户（支持用户名或邮箱登录）
	var foundUser user.User
	result := database.PostgresDB.Where("username = ? OR email = ?", req.Username, req.Username).First(&foundUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return LoginResponse{}, response.NewBusinessError(
				response.WithErrorCode(response.Fail),
				response.WithErrorMessage("用户名或密码错误"),
			)
		}
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("登录失败"),
		)
	}

	// 4. 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password)); err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("用户名或密码错误"),
		)
	}

	// 5. 生成 refresh token
	token, err := pkg.GenerateRandomToken()
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成令牌失败"),
		)
	}

	// 6. 生成 access token (JWT)
	username := ""
	if foundUser.Username != nil {
		username = *foundUser.Username
	}
	accessToken, err := pkg.GenerateAccessToken(foundUser.ID, username, foundUser.Email, foundUser.Role)
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成访问令牌失败"),
		)
	}

	// 7. 存储 refresh token
	tokenData := refresh.TokenData{
		UserID:   foundUser.ID,
		Username: username,
		Email:    foundUser.Email,
		Role:     foundUser.Role,
	}
	if err := s.getRefreshTokenRepo().Create(token, tokenData); err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("存储令牌失败"),
		)
	}

	// 8. 删除已使用的 state（防止重复使用）
	pkg.DeleteState(req.State)

	// 9. 返回结果
	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: token,
		RedirectUrl:  redirectUrl,
	}, nil
}

// 参数校验
func (s *SSEWikiLoginService) validateRequest(req LoginRequest) *response.BusinessError {
	if req.State == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("state 不能为空"),
		)
	}

	if req.Username == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("用户名不能为空"),
		)
	}

	if req.Password == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("密码不能为空"),
		)
	}

	return nil
}