package dao

import (
	"go-admin/model"
)

// ListDictTypes 分页查询字典类型
func ListDictTypes(pageNum, pageSize int, dictName, dictType string) ([]model.DictType, int64, error) {
	var types []model.DictType
	var total int64
	db := DB.Model(&model.DictType{})
	if dictName != "" {
		db = db.Where(likeExprEsc("dict_name"), "%"+escapeLike(dictName)+"%")
	}
	if dictType != "" {
		db = db.Where(likeExprEsc("dict_type"), "%"+escapeLike(dictType)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("id asc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&types).Error
	return types, total, err
}

// GetDictTypeByID 按 ID 查询字典类型
func GetDictTypeByID(id uint) (*model.DictType, error) {
	var t model.DictType
	err := DB.First(&t, id).Error
	return &t, err
}

// CountDictTypeByType 统计相同字典类型编码的数量
func CountDictTypeByType(dictType string) (int64, error) {
	var count int64
	err := DB.Model(&model.DictType{}).Where("dict_type = ?", dictType).Count(&count).Error
	return count, err
}

// CreateDictType 新增字典类型
func CreateDictType(t *model.DictType) error {
	return DB.Create(t).Error
}

// UpdateDictType 更新字典类型
func UpdateDictType(t *model.DictType) error {
	return DB.Save(t).Error
}

// DeleteDictType 删除字典类型
func DeleteDictType(id uint) error {
	return DB.Delete(&model.DictType{}, id).Error
}

// ListDictData 分页查询字典数据
func ListDictData(pageNum, pageSize int, dictType, label string) ([]model.DictData, int64, error) {
	var data []model.DictData
	var total int64
	db := DB.Model(&model.DictData{})
	if dictType != "" {
		db = db.Where("dict_type = ?", dictType)
	}
	if label != "" {
		db = db.Where(likeExprEsc("dict_label"), "%"+escapeLike(label)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("dict_sort asc, id asc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&data).Error
	return data, total, err
}

// ListDictDataByType 按类型查询全部字典数据（前端下拉用，仅正常）
func ListDictDataByType(dictType string) ([]model.DictData, error) {
	var data []model.DictData
	err := DB.Where("dict_type = ? AND status = '0'", dictType).Order("dict_sort asc, id asc").Find(&data).Error
	return data, err
}

// GetDictDataByID 按 ID 查询字典数据
func GetDictDataByID(id uint) (*model.DictData, error) {
	var d model.DictData
	err := DB.First(&d, id).Error
	return &d, err
}

// CountDictDataByType 统计某类型下的数据条数
func CountDictDataByType(dictType string) (int64, error) {
	var count int64
	err := DB.Model(&model.DictData{}).Where("dict_type = ?", dictType).Count(&count).Error
	return count, err
}

// CreateDictData 新增字典数据
func CreateDictData(d *model.DictData) error {
	return DB.Create(d).Error
}

// UpdateDictData 更新字典数据
func UpdateDictData(d *model.DictData) error {
	return DB.Save(d).Error
}

// DeleteDictData 删除字典数据
func DeleteDictData(id uint) error {
	return DB.Delete(&model.DictData{}, id).Error
}

// DeleteDictDatas 批量删除字典数据
func DeleteDictDatas(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return DB.Where("id IN ?", ids).Delete(&model.DictData{}).Error
}
