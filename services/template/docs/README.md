# Template Service API 文档

## 快速开始

### 访问 Swagger UI

启动服务后，访问：**http://localhost:8082/swagger/index.html**

### 生成文档

每次修改 API 注释后，需要重新生成文档：

```bash
# 在 template 目录下执行
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

或者使用已安装的 swag：

```bash
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

## API 端点

这是一个模板服务，暂无实际 API。

### 如何添加新接口

1. **在 handler 方法上添加注释**：

```go
// HandleExample 接口说明
// @Summary 简短描述
// @Description 详细描述
// @Tags 标签分组
// @Accept json
// @Produce json
// @Param request body RequestDTO true "请求参数"
// @Success 200 {object} response.Response{data=ResponseDTO} "成功描述"
// @Failure 400 {object} response.Response "失败描述"
// @Router /api/v1/path [post]
func (h *Handler) HandleExample(c *gin.Context) {
    // ...
}
```

2. **在 DTO 结构体添加示例值**：

```go
type RequestDTO struct {
    Field string `json:"field" binding:"required" example:"示例值"` // 字段说明
}
```

3. **重新生成文档**（见上方命令）

4. **刷新浏览器**查看更新

## 注意事项

- 文档自动生成，无需手动编辑 `docs/` 目录下的文件
- `docs/` 目录已添加到 `.gitignore`，不会提交到版本控制
- API 路由前缀：`/api/v1`
- 认证方式：Bearer Token（在 Header 中添加 `Authorization: Bearer <token>`）
