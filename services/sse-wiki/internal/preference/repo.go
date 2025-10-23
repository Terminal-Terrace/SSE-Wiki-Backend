package preference

import (
	"terminal-terrace/sse-wiki/internal/model/preference"

	"gorm.io/gorm"
)

func (p *PreferHandler) RepoUpdatePrefer(id uint, tag string) error {

	if err := p.db.Model(&preference.UserPreference{}).
		Where("user_id = ?", id).
		UpdateColumn("prefer_index", gorm.Expr("prefer_index * ?", p.decreaseRatio)).Error; err != nil {
		return err
	}

	// 尝试更新特定标签的偏好指数
	result := p.db.Model(&preference.UserPreference{}).
		Where("user_id = ? AND prefer_tag = ?", id, tag).
		Update("prefer_index", gorm.Expr("prefer_index + ?", 1))

	if result.Error != nil {
		return result.Error
	}

	// 如果没有行被更新，说明记录不存在，则创建它
	if result.RowsAffected == 0 {
		newRecord := preference.UserPreference{
			UserID:      id,
			PreferTag:   tag,
			PreferIndex: 1.0,
		}
		return p.db.Create(&newRecord).Error
	}

	return nil
}

func (p *PreferHandler) RepoSetPreference(id uint, tag string, wantedIndex float64) error {

	// 尝试直接更新
	result := p.db.Model(&preference.UserPreference{}).
		Where("user_id = ? AND prefer_tag = ?", id, tag).
		Update("prefer_index", wantedIndex)

	if result.Error != nil {
		return result.Error
	}

	// 如果没有行被更新，说明记录不存在，则创建它
	if result.RowsAffected == 0 {
		newRecord := preference.UserPreference{
			UserID:      id,
			PreferTag:   tag,
			PreferIndex: wantedIndex,
		}
		return p.db.Create(&newRecord).Error
	}

	return nil
}

func (p *PreferHandler) RepoGetPreference(id uint, tag string) (preference.UserPreference, error) {
	var record preference.UserPreference
	err := p.db.Where("user_id = ? AND prefer_tag = ?", id, tag).First(&record).Error

	// 如果记录未找到或发生错误，返回默认值 0
	if err != nil {
		return preference.UserPreference{
			UserID:      id,
			PreferTag:   "",
			PreferIndex: -1,
		}, err
	}
	return record, nil
}

func (p *PreferHandler) RepoGetBestPreference(id uint) (preference.UserPreference, error) {
	var record preference.UserPreference
	err := p.db.Where("user_id = ?", id).Order("prefer_index desc").First(&record).Error

	if err != nil {
		// 如果记录未找到或发生错误，返回一个带有无效索引的记录
		return preference.UserPreference{
			UserID:      id,
			PreferTag:   "",
			PreferIndex: -1,
		}, err
	}

	return record, nil
}
