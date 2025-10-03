package pkg

import (
	"crypto/rand"
	"encoding/base64"
)

// 生成一个随机的 state 字符串, 用于 OAuth2 流程中防止 CSRF 攻击
// 理论上在一定时间内不会有重复的, 有的话只能说运气有点好了. :)
func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}