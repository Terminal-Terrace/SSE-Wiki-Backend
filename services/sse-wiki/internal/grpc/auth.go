package grpc

import (
	"context"
	"strings"

	authsdk "terminal-terrace/auth-sdk"
	"terminal-terrace/sse-wiki/config"

	"google.golang.org/grpc/metadata"
)

// GetUserFromContext 从 gRPC context 获取用户信息
// 使用 auth-sdk 解析 JWT token
// 如果没有 token 或解析失败，返回空的 UserContext（UserID=0）
func GetUserFromContext(ctx context.Context) *authsdk.UserContext {
	return authsdk.GetUserFromContext(ctx, config.Conf.JWT.Secret)
}

// ExtractToken extracts JWT token from gRPC metadata
// The token is expected in the "authorization" metadata key with "Bearer " prefix
func ExtractToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}

	// Remove "Bearer " prefix if present
	token := values[0]
	if strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}

	return token
}

// ExtractUserInfo extracts user_id and user_role from gRPC metadata
// These are passed from Node.js Gateway after JWT validation
// Deprecated: 使用 GetUserFromContext 代替，它直接从 JWT 解析用户信息
func ExtractUserInfo(ctx context.Context) (userID uint, userRole string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, ""
	}

	// Extract user_id
	if values := md.Get("x-user-id"); len(values) > 0 {
		// Parse user_id as uint
		var id uint64
		if _, err := parseUint(values[0], &id); err == nil {
			userID = uint(id)
		}
	}

	// Extract user_role
	if values := md.Get("x-user-role"); len(values) > 0 {
		userRole = values[0]
	}

	return userID, userRole
}

// parseUint is a helper to parse string to uint64
func parseUint(s string, result *uint64) (bool, error) {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + uint64(c-'0')
	}
	*result = n
	return true, nil
}
