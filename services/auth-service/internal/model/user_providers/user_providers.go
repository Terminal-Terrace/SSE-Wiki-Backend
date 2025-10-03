package userproviders

type UserProvider struct {
	UserID         int     `gorm:"column:user_id;primaryKey" json:"user_id"`
	Provider       string  `gorm:"column:provider;type:varchar(50);not null" json:"provider"`
	ProviderUserID string  `gorm:"column:provider_user_id;type:varchar(50);not null" json:"provider_user_id"`
	ProviderEmail  *string `gorm:"column:provider_email;type:varchar(50);index" json:"provider_email,omitempty"`
}

func (UserProvider) TableName() string {
	return "user_providers"
}