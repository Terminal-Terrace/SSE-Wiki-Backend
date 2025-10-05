package module

import (
	"strconv"
	"terminal-terrace/response"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/dto"

	"github.com/gin-gonic/gin"
)

type ModuleHandler struct {
	moduleService *ModuleService
}

func NewModuleHandler() *ModuleHandler {
	return &ModuleHandler{
		moduleService: NewModuleService(database.PostgresDB),
	}
}

// GetModuleTree 获取模块树
// @Summary 获取完整模块树
// @Description 返回完整的树形结构，后端负责构建树
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]ModuleTreeNode}
// @Router /modules [get]
func (h *ModuleHandler) GetModuleTree(c *gin.Context) {
	// 从上下文获取用户信息（由认证中间件设置）
	userID, _ := c.Get("user_id")
	uid := userID.(uint)

	tree, err := h.moduleService.GetModuleTree(uid)
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, tree)
}

// GetModule 获取单个模块信息
// @Summary 获取单个模块
// @Description 获取指定模块的详细信息
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Success 200 {object} response.Response{data=moduleModel.Module}
// @Router /modules/{id} [get]
func (h *ModuleHandler) GetModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	module, err := h.moduleService.GetModule(uint(id))
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, module)
}

// GetBreadcrumbs 获取面包屑导航
// @Summary 获取面包屑导航
// @Description 从根模块到当前模块的路径
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Success 200 {object} response.Response{data=[]BreadcrumbNode}
// @Router /modules/{id}/breadcrumbs [get]
func (h *ModuleHandler) GetBreadcrumbs(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	breadcrumbs, err := h.moduleService.GetBreadcrumbs(uint(id))
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, breadcrumbs)
}

// CreateModule 创建模块
// @Summary 创建新模块
// @Description 创建顶级模块或子模块
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateModuleRequest true "创建模块请求"
// @Success 200 {object} response.Response{data=moduleModel.Module}
// @Router /modules [post]
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("参数错误"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	module, err := h.moduleService.CreateModule(req, userID.(uint), userRole.(string))
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, module)
}

// UpdateModule 更新模块
// @Summary 更新模块信息
// @Description 更新模块名称或移动模块位置
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Param request body UpdateModuleRequest true "更新模块请求"
// @Success 200 {object} response.Response
// @Router /modules/{id} [put]
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("参数错误"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	if err := h.moduleService.UpdateModule(uint(id), req, userID.(uint), userRole.(string)); err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "更新成功"})
}

// DeleteModule 删除模块
// @Summary 删除模块
// @Description 删除模块及其所有子模块（级联删除）
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Success 200 {object} response.Response{data=DeleteModuleResponse}
// @Router /modules/{id} [delete]
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	count, err := h.moduleService.DeleteModule(uint(id), userID.(uint), userRole.(string))
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, DeleteModuleResponse{
		DeletedModules: int(count),
	})
}

// GetModerators 获取协作者列表
// @Summary 获取协作者列表
// @Description 获取指定模块的协作者列表
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Success 200 {object} response.Response{data=[]ModeratorInfo}
// @Router /modules/{id}/moderators [get]
func (h *ModuleHandler) GetModerators(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	moderators, err := h.moduleService.GetModerators(uint(id), userID.(uint), userRole.(string))
	if err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, moderators)
}

// AddModerator 添加协作者
// @Summary 添加协作者
// @Description 为模块添加协作者
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Param request body AddModeratorRequest true "添加协作者请求"
// @Success 200 {object} response.Response
// @Router /modules/{id}/moderators [post]
func (h *ModuleHandler) AddModerator(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	var req AddModeratorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("参数错误"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	if err := h.moduleService.AddModerator(uint(id), req, userID.(uint), userRole.(string)); err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "添加成功"})
}

// RemoveModerator 移除协作者
// @Summary 移除协作者
// @Description 从模块移除协作者
// @Tags 模块管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模块ID"
// @Param userId path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /modules/{id}/moderators/{userId} [delete]
func (h *ModuleHandler) RemoveModerator(c *gin.Context) {
	moduleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的模块ID"),
		))
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		dto.ErrorResponse(c, response.NewBusinessError(
			response.WithErrorCode(response.ParseError),
			response.WithErrorMessage("无效的用户ID"),
		))
		return
	}

	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")

	if err := h.moduleService.RemoveModerator(uint(moduleID), uint(targetUserID), userID.(uint), userRole.(string)); err != nil {
		dto.ErrorResponse(c, err.(*response.BusinessError))
		return
	}

	dto.SuccessResponse(c, gin.H{"message": "移除成功"})
}
