package me

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	UserID   uint   `json:"user_id" example:"1"`
	Username string `json:"username" example:"testuser"`
	Email    string `json:"email" example:"test@example.com"`
	Role     string `json:"role" example:"admin"`
}
