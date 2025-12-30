package authsdk

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
	ErrNoToken      = errors.New("no token provided")
)

// Claims JWT 自定义声明
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// UserContext 用户上下文信息
type UserContext struct {
	UserID   int
	Username string
	Email    string
	Role     string // "admin" 表示全局管理员
}

// ParseToken 解析并验证 JWT token
// secret: JWT 签名密钥
func ParseToken(tokenString, secret string) (*UserContext, error) {
	if tokenString == "" {
		return nil, ErrNoToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &UserContext{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     claims.Role,
		}, nil
	}

	return nil, ErrInvalidToken
}
