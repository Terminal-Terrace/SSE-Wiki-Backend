package user

import "time"

type User struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(50);not null;uniqueIndex" json:"username"`
	Email        string     `gorm:"column:email;type:varchar(100);not null;uniqueIndex" json:"email"`
	PasswordHash string     `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	FullName     *string    `gorm:"column:full_name;type:varchar(100)" json:"full_name,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "auth_users"
}