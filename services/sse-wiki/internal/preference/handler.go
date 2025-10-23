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
// @Produce json
// @Param user_id query uint64 true "用户ID"
// @Param tag query string true "标签"
// @Success 200 {object} response.Response{data=float64} "偏好指数（数字）"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /get_refer [get]
func (p *PreferHandler) GetPrefer(c *gin.Context) {
	var msg dto.QueryReferRequest
	if err := c.ShouldBindQuery(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	res, err := p.RepoGetPreference(msg.UserId, msg.Tag)

	if err != nil {
		dto.ErrorResponse(c, &response.BusinessError{
			Code: 404,
			Msg:  "user-tag pair not found",
			Err:  err,
		})
		return
	}

	dto.SuccessResponse(c, res)
}

// GetBestPrefer 获取用户最偏好的标签
// @Summary 获取用户最偏好的标签
// @Description 根据用户ID查询偏好指数最高的标签。若该用户无任何偏好记录，返回空字符串。
// @Tags Preference
// @Produce json
// @Param user_id query uint true "用户ID"
// @Success 200 {object} response.Response{data=string} "偏好指数最高的标签"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /get_best_refer [get]
func (p *PreferHandler) GetBestPrefer(c *gin.Context) {
	var msg dto.QueryBestReferRequest
	if err := c.ShouldBindQuery(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	res, err := p.RepoGetBestPreference(msg.UserId)

	if err != nil {
		dto.ErrorResponse(c, &response.BusinessError{
			Code: 404,
			Msg:  "user not found",
			Err:  err,
		})
		return
	}

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

	err := p.RepoUpdatePrefer(msg.UserId, msg.Tag)

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

// SetPreference 设置用户对某标签的偏好
// @Summary 设置用户偏好
// @Description 设置指定用户对某标签的偏好指数为特定值。若记录不存在，则创建新记录。
// @Tags Preference
// @Accept json
// @Produce json
// @Param request body dto.SetReferRequest true "设置参数"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "设置失败（如数据库错误）"
// @Router /set_refer [post]
func (p *PreferHandler) SetPreference(c *gin.Context) {
	var msg dto.SetReferRequest
	if err := c.ShouldBindJSON(&msg); err != nil {
		dto.ValidationErrorResponse(c, err)
		return
	}

	err := p.RepoSetPreference(msg.UserId, msg.Tag, msg.WantedId)

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
