package preference

type UserPreference struct {
	UserID      uint    `gorm:"column:user_id;not null;uniqueIndex:idx_user_tag" json:"user_id"`
	PreferTag   string  `gorm:"column:prefer_tag;type:text;not null;uniqueIndex:idx_user_tag" json:"prefer_tag"`
	PreferIndex float64 `gorm:"column:prefer_index;type:double precision;not null" json:"prefer_index"`
}

func (UserPreference) TableName() string {
	return "user_preference"
}
