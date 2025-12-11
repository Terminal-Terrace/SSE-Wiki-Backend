package user

import "time"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
)

type User struct {
	ID           int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     *string   `gorm:"column:username;type:varchar(50);uniqueIndex" json:"username"`
	Email        string    `gorm:"column:email;type:varchar(100);not null;uniqueIndex" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"column:role;type:varchar(20);not null;default:'student'" json:"role"`
	Avatar       *string   `gorm:"column:avatar;type:varchar(500)" json:"avatar"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "auth_users"
}
