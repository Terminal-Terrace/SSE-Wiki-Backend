package user

import "time"

// User 用户模型(映射到auth_users表,只读)
type User struct {
	ID       uint      `gorm:"column:id;primaryKey" json:"id"`
	Username string    `gorm:"column:username" json:"username"`
	Email    string    `gorm:"column:email" json:"email"`
	Role     string    `gorm:"column:role" json:"role"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (User) TableName() string {
	return "auth_users"
}
