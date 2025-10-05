package login

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
	"terminal-terrace/auth-service/internal/model/user"
	userproviders "terminal-terrace/auth-service/internal/model/user_providers"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/auth-service/internal/refresh"
	"terminal-terrace/response"

	"gorm.io/gorm"
)

type GithubLoginService struct{
	refreshTokenRepo *refresh.RefreshTokenRepository
}

func init() {
	registerLoginService("github", &GithubLoginService{})
}

func (s *GithubLoginService) getRefreshTokenRepo() *refresh.RefreshTokenRepository {
	if s.refreshTokenRepo == nil {
		s.refreshTokenRepo = refresh.NewRefreshTokenRepository(database.RedisDB)
	}
	return s.refreshTokenRepo
}

func (s *GithubLoginService) Login(req LoginRequest) (LoginResponse, *response.BusinessError) {
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

	// 3. 使用 code 换取 GitHub access token
	accessToken, err := s.getGitHubAccessToken(req.Code)
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取 GitHub access token 失败"),
		)
	}

	// 4. 使用 access token 获取 GitHub 用户信息
	githubUser, err := s.getGitHubUserInfo(accessToken)
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("获取 GitHub 用户信息失败"),
		)
	}

	// 5. 查找或创建用户
	foundUser, err := s.findOrCreateUser(githubUser)
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("用户创建或查询失败"),
		)
	}

	// 6. 生成 refresh token
	token, err := pkg.GenerateRandomToken()
	if err != nil {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成令牌失败"),
		)
	}

	// 7. 存储 refresh token
	username := ""
	if foundUser.Username != nil {
		username = *foundUser.Username
	}
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
		RefreshToken: token,
		RedirectUrl:  redirectUrl,
	}, nil
}

// validateRequest 参数校验
func (s *GithubLoginService) validateRequest(req LoginRequest) *response.BusinessError {
	if req.State == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("state 不能为空"),
		)
	}

	if req.Code == "" {
		return response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("code 不能为空"),
		)
	}

	return nil
}

// GitHubAccessTokenResponse GitHub access token 响应结构
type GitHubAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// getGitHubAccessToken 使用 code 换取 GitHub access token
func (s *GithubLoginService) getGitHubAccessToken(code string) (string, error) {
	// 构建请求参数
	data := url.Values{}
	data.Set("client_id", config.Conf.Github.ClientID)
	data.Set("client_secret", config.Conf.Github.ClientSecret)
	data.Set("code", code)

	// 发送请求到 GitHub
	resp, err := http.PostForm("https://github.com/login/oauth/access_token", data)
	if err != nil {
		return "", fmt.Errorf("请求 GitHub access token 失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 GitHub 响应失败: %w", err)
	}

	// 解析响应 (GitHub 返回的是 URL encoded 格式)
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", fmt.Errorf("解析 GitHub 响应失败: %w", err)
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		return "", fmt.Errorf("未获取到 access token: %s", string(body))
	}

	return accessToken, nil
}

// GitHubUser GitHub 用户信息结构
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// getGitHubUserInfo 使用 access token 获取 GitHub 用户信息
func (s *GithubLoginService) getGitHubUserInfo(accessToken string) (*GitHubUser, error) {
	// 创建请求
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub 用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 GitHub 用户信息失败: %w", err)
	}

	// 解析响应
	var githubUser GitHubUser
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("解析 GitHub 用户信息失败: %w", err)
	}

	// 如果主 API 没有返回 email，尝试获取用户的邮箱列表
	if githubUser.Email == "" {
		email, _ := s.getGitHubUserEmail(accessToken)
		githubUser.Email = email
	}

	return &githubUser, nil
}

// GitHubEmail GitHub 邮箱信息结构
type GitHubEmail struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

// getGitHubUserEmail 获取 GitHub 用户的主邮箱
func (s *GithubLoginService) getGitHubUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []GitHubEmail
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}

	// 找到主邮箱
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}

	// 如果没有主邮箱，返回第一个验证过的邮箱
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("未找到验证过的邮箱")
}

// findOrCreateUser 查找或创建用户
func (s *GithubLoginService) findOrCreateUser(githubUser *GitHubUser) (*user.User, error) {
	// 1. 先查找是否已存在该 GitHub 用户的绑定关系
	var userProvider userproviders.UserProvider
	result := database.PostgresDB.Where("provider = ? AND provider_user_id = ?", "github", strconv.Itoa(githubUser.ID)).First(&userProvider)

	if result.Error == nil {
		// 找到了绑定关系，直接获取用户
		var foundUser user.User
		if err := database.PostgresDB.First(&foundUser, userProvider.UserID).Error; err != nil {
			return nil, err
		}
		return &foundUser, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		// 不是"未找到"错误，是其他数据库错误
		return nil, result.Error
	}

	// 2. 没有绑定关系，创建新用户（仅写入邮箱）
	// TODO: 之后改为重定向到注册页面，让用户补全信息
	var newUser user.User
	email := githubUser.Email

	// 如果邮箱为空，使用 GitHub 提供的 noreply 邮箱
	if email == "" {
		email = fmt.Sprintf("%d+%s@users.noreply.github.com", githubUser.ID, githubUser.Login)
	}

	// 使用事务确保用户和绑定关系同时创建
	err := database.PostgresDB.Transaction(func(tx *gorm.DB) error {
		// 创建用户（仅写入邮箱，其他字段留空或使用默认值）
		newUser = user.User{
			Email:        email,
			Username:     nil,             // 留空，等待用户补全
			PasswordHash: "",              // OAuth 用户不需要密码
			Role:         string(user.RoleStudent),
		}

		if err := tx.Create(&newUser).Error; err != nil {
			return err
		}

		// 创建绑定关系
		providerEmail := githubUser.Email
		newUserProvider := userproviders.UserProvider{
			UserID:         newUser.ID,
			Provider:       "github",
			ProviderUserID: strconv.Itoa(githubUser.ID),
			ProviderEmail:  &providerEmail,
		}

		if err := tx.Create(&newUserProvider).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &newUser, nil
}