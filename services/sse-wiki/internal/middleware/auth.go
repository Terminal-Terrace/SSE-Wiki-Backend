package middleware

import (
	"fmt"
	"terminal-terrace/response"
	"terminal-terrace/sse-wiki/config"
	"terminal-terrace/sse-wiki/internal/dto"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 载荷
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"` // TODO: 需确认users表添加role字段
	jwt.RegisteredClaims
}

// parseToken 从 cookie 或 Authorization header 中解析 token
func parseToken(c *gin.Context) (*Claims, error) {
	var tokenString string

	// 优先从 cookie 中获取 access_token
	tokenString, err := c.Cookie("access_token")
	if err != nil || tokenString == "" {
		// 如果 cookie 中没有，尝试从 Authorization header 获取（兼容旧方式）
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			return nil, fmt.Errorf("未提供认证令牌")
		}

		// 验证格式: Bearer <token>
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			return nil, fmt.Errorf("认证格式错误")
		}
	}

	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Conf.JWT.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("无效的认证令牌")
	}

	// 提取 claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("认证令牌无效")
}

// JWTAuth JWT 认证中间件（必需认证）
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseToken(c)
		if err != nil {
			dto.ErrorResponse(c, response.NewBusinessError(
				response.WithErrorCode(response.Unauthorized),
				response.WithErrorMessage(err.Error()),
			))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// OptionalJWTAuth 可选的 JWT 认证中间件（不强制要求认证，但如果有token则解析）
func OptionalJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseToken(c)
		if err == nil && claims != nil {
			// 如果有有效的 token，将用户信息存入上下文
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("user_role", claims.Role)
		}
		// 无论是否有 token，都继续执行
		c.Next()
	}
}
