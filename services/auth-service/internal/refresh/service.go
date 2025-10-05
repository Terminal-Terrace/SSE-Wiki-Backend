package refresh

import (
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/response"
)

type RefreshTokenService struct {
	repo *RefreshTokenRepository
}

// NewRefreshTokenService 创建刷新令牌服务实例
func NewRefreshTokenService(repo *RefreshTokenRepository) *RefreshTokenService {
	return &RefreshTokenService{
		repo: repo,
	}
}

// refreshTokenResult 内部返回结果（包含新的 refresh token）
type refreshTokenResult struct {
	AccessToken     string
	newRefreshToken string // 小写，不导出，仅内部使用
}

// RefreshToken 刷新访问令牌
func (s *RefreshTokenService) RefreshToken(req RefreshTokenRequest) (*refreshTokenResult, *response.BusinessError) {
	// 1. 验证 refresh token
	tokenData, err := s.repo.Get(req.RefreshToken)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("刷新令牌无效或已过期"),
		)
	}

	// 2. 撤销旧的 refresh token
	if err := s.repo.Delete(req.RefreshToken); err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("撤销旧令牌失败"),
		)
	}

	// 3. 生成新的 access token
	accessToken, err := pkg.GenerateAccessToken(tokenData.UserID, tokenData.Username, tokenData.Email)
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成访问令牌失败"),
		)
	}

	// 4. 生成新的 refresh token
	newRefreshToken, err := pkg.GenerateRandomToken()
	if err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("生成刷新令牌失败"),
		)
	}

	// 5. 存储新的 refresh token
	if err := s.repo.Create(newRefreshToken, *tokenData); err != nil {
		return nil, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("存储刷新令牌失败"),
		)
	}

	return &refreshTokenResult{
		AccessToken:     accessToken,
		newRefreshToken: newRefreshToken,
	}, nil
}
