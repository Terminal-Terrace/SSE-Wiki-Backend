package preference

import (
	"terminal-terrace/response"
	"terminal-terrace/sse-wiki/internal/dto"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PreferHandler struct {
	db            *gorm.DB
	decreaseRatio float64
}

func NewPreferHandler(db *gorm.DB, ratio float64) *PreferHandler {
	return &PreferHandler{
		db:            db,
		decreaseRatio: ratio,
	}
}

// GetPrefer 获取指定用户对某标签的偏好指数
// @Summary 获取用户偏好指数
// @Description 根据用户ID和标签名称查询其偏好指数（PreferIndex）。若记录不存在，返回0。
// @Tags Preference
// @Accept json
// @Produce json
// @Param request body dto.QueryReferRequest true "查询参数"
// @Success 200 {object} response.Response{data=float64} "偏好指数（数字）"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /get_refer [post]
func (p *PreferHandler) GetPrefer(c *gin.Context) {
	var msg dto.QueryReferRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	res := p.GetPreference(msg.UserId, msg.Tag)

	dto.SuccessResponse(c, res)
}

// UpdatePreference 更新用户对某标签的偏好
// @Summary 更新用户偏好
// @Description 对所有偏好记录执行全局衰减（PreferIndex *= decreaseRatio），然后对指定用户ID和标签的记录执行 PreferIndex += 1；若不存在则创建新记录（PreferIndex=1）。
// @Tags Preference
// @Accept json
// @Produce json
// @Param request body dto.UpdateReferRequest true "更新参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "更新失败（如数据库错误）"
// @Router /update_refer [post]
func (p *PreferHandler) UpdatePreference(c *gin.Context) {
	var msg dto.UpdateReferRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	err := p.UpdatePrefer(msg.UserId, msg.Tag)

	// fmt.Println(err)

	if err != nil {
		dto.ErrorResponse(c, &response.BusinessError{
			Code: 404,
			Msg:  "update unsuccess",
			Err:  err,
		})
		return
	}

	dto.SuccessResponse(c, nil)
}

// GetPrefer 获取指定用户对某标签的偏好指数
// @Summary 获取用户偏好指数
// @Description 根据用户ID查询偏好指数最高的标签。若记录不存在，返回0。
// @Tags Preference
// @Accept json
// @Produce json
// @Param request body dto.QueryBestReferRequest true "查询参数"
// @Success 200 {object} response.Response{data=float64} "偏好指数（数字）"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /get_best_refer [post]
func (p *PreferHandler) GetBestRefer(c *gin.Context) {
	var msg dto.QueryBestReferRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	res := p.GetBestPreference(msg.UserId)

	dto.SuccessResponse(c, res)
}
