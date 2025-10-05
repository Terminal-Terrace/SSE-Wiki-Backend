package code

import (
	"github.com/gin-gonic/gin"

	"terminal-terrace/email"
	"terminal-terrace/auth-service/config"
)

func RegisterRoutes(r *gin.RouterGroup) {
	mailer := email.NewClient(&config.Conf.Smtp)

	codeService := NewCodeService(mailer)

	h := &CodeHandler{
		service: codeService,
	}
	r.POST("/code", h.handle)
}