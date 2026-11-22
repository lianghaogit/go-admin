package dao

import (
	"go-admin/model"
)

// ConfigQuery 参数查询条件
type ConfigQuery struct {
	PageNum    int
	PageSize   int
	ConfigName string
	ConfigKey  string
	ConfigType string
}

// ListConfigs 分页查询系统参数
func ListConfigs(q ConfigQuery) ([]model.Config, int64, error) {
	var configs []model.Config
	var total int64
	db := DB.Model(&model.Config{})
	if q.ConfigName != "" {
		db = db.Where(likeExprEsc("config_name"), "%"+escapeLike(q.ConfigName)+"%")
	}
	if q.ConfigKey != "" {
		db = db.Where(likeExprEsc("config_key"), "%"+escapeLike(q.ConfigKey)+"%")
	}
	if q.ConfigType != "" {
		db = db.Where("config_type = ?", q.ConfigType)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("id asc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&configs).Error
	return configs, total, err
}

// GetConfigByID 按 ID 查询参数
func GetConfigByID(id uint) (*model.Config, error) {
	var c model.Config
	err := DB.First(&c, id).Error
	return &c, err
}

// GetConfigByKey 按 Key 查询参数
func GetConfigByKey(key string) (*model.Config, error) {
	var c model.Config
	err := DB.Where("config_key = ?", key).First(&c).Error
	return &c, err
}

// CountConfigByKey 统计相同 key 的数量
func CountConfigByKey(key string) (int64, error) {
	var count int64
	err := DB.Model(&model.Config{}).Where("config_key = ?", key).Count(&count).Error
	return count, err
}

// CreateConfig 新增参数
func CreateConfig(c *model.Config) error {
	return DB.Create(c).Error
}

// UpdateConfig 更新参数
func UpdateConfig(c *model.Config) error {
	return DB.Save(c).Error
}

// DeleteConfig 删除参数
func DeleteConfig(id uint) error {
	return DB.Delete(&model.Config{}, id).Error
}
