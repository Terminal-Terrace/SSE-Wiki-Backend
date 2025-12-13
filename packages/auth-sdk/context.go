package authsdk

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

// ExtractTokenFromContext 从 gRPC context 的 metadata 中提取 JWT token
// 支持两种方式：
// 1. authorization header (Bearer token)
// 2. x-access-token header
func ExtractTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrNoToken
	}

	// 尝试从 authorization header 获取
	if values := md.Get("authorization"); len(values) > 0 {
		token := values[0]
		// 移除 "Bearer " 前缀
		if strings.HasPrefix(token, "Bearer ") {
			return strings.TrimPrefix(token, "Bearer "), nil
		}
		return token, nil
	}

	// 尝试从 x-access-token 获取
	if values := md.Get("x-access-token"); len(values) > 0 {
		return values[0], nil
	}

	return "", ErrNoToken
}

// GetUserFromContext 从 gRPC context 获取用户信息
// 如果没有 token 或解析失败，返回空的 UserContext（UserID=0）
// secret: JWT 签名密钥
func GetUserFromContext(ctx context.Context, secret string) *UserContext {
	token, err := ExtractTokenFromContext(ctx)
	if err != nil {
		return &UserContext{} // 未登录用户
	}

	user, err := ParseToken(token, secret)
	if err != nil {
		return &UserContext{} // token 无效
	}

	return user
}
