package preference

import (
	"terminal-terrace/sse-wiki/internal/model/preference"

	"gorm.io/gorm"
)

func (p *PreferHandler) UpdatePrefer(id uint, tag string) error {

	if err := p.db.Model(&preference.UserPreference{}).
		Where("1=1").
		UpdateColumn("prefer_index", gorm.Expr("prefer_index * ?", p.decreaseRatio)).Error; err != nil {
		return err
	}
	var record preference.UserPreference
	err := p.db.Where("user_id = ? AND prefer_tag = ?", id, tag).First(&record).Error

	// 找到了，PreferIndex += 1
	if err == nil {
		return p.db.Model(&preference.UserPreference{}).
			Where("user_id = ? AND prefer_tag = ?", id, tag).
			Update("prefer_index", gorm.Expr("prefer_index + ?", 1)).Error
	}

	// 没找到，创建新记录，PreferIndex 设为 1
	if err == gorm.ErrRecordNotFound {
		// fmt.Println("Not found")
		newRecord := preference.UserPreference{
			UserID:      id,
			PreferTag:   tag,
			PreferIndex: 1.0,
		}

		return p.db.Create(&newRecord).Error
	}

	return err
}

func (p *PreferHandler) GetPreference(id uint, tag string) preference.UserPreference {
	var record preference.UserPreference
	err := p.db.Where("user_id = ? AND prefer_tag = ?", id, tag).First(&record).Error

	// 如果记录未找到或发生错误，返回默认值 0
	if err != nil {
		return preference.UserPreference{
			UserID:      id,
			PreferTag:   "",
			PreferIndex: -1,
		}
	}
	return record
}

func (p *PreferHandler) GetBestPreference(id uint) preference.UserPreference {
	var record preference.UserPreference
	err := p.db.Where("user_id = ?", id).Order("prefer_index desc").First(&record).Error

	if err != nil {
		// 如果记录未找到或发生错误，返回一个带有无效索引的记录
		return preference.UserPreference{
			UserID:      id,
			PreferTag:   "",
			PreferIndex: -1,
		}
	}
	return record
}
