package pkg

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomToken 生成随机令牌字符串（纯工具函数）
func GenerateRandomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("生成随机令牌失败: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
