package dto

// UpdateReferRequest 用于更新用户偏好的请求体
// @Description 请求更新指定用户对某标签的偏好。系统将先对所有偏好值做全局衰减，再对该用户-标签组合的偏好值加1（若不存在则创建，初始值为1）。
type UpdateReferRequest struct {
	// 用户的唯一ID
	// Required: true
	// Example: 123
	UserId uint `json:"user_id" binding:"gte=0"`
	// 偏好标签名称（如 "sports", "tech"）
	// Required: true
	// Example: "movies"
	Tag string `json:"tag" binding:"required"`
}

// QueryReferRequest 用于查询用户偏好的请求体
// @Description 请求查询指定用户对某标签的当前偏好指数（PreferIndex）。
type QueryReferRequest struct {
	// 用户的唯一ID
	// Required: true
	// Example: 123
	UserId uint `json:"user_id" binding:"gte=0"`
	// 偏好标签名称
	// Required: true
	// Example: "music"
	Tag string `json:"tag" binding:"required"`
}

// QueryReferRequest 用于查询用户偏好的请求体
// @Description 请求查询指定用户最高偏好指数的标签
type QueryBestReferRequest struct {
	// 用户的唯一ID
	// Required: true
	// Example: 123
	UserId uint `json:"user_id" binding:"gte=0"`
}
