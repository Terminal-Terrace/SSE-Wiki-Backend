package logout

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	handler := &LogoutHandler{}

	// 退出登录（不需要认证中间件，因为只是清除 Cookie）
	r.POST("/logout", handler.Logout)
}
