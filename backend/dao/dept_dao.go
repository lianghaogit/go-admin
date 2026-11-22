package dao

import (
	"go-admin/model"

	"gorm.io/gorm"
)

// ListDepts 查询全部部门（按父子关系与排序）
func ListDepts() ([]model.Dept, error) {
	var depts []model.Dept
	err := DB.Order("parent_id asc, order_num asc, id asc").Find(&depts).Error
	return depts, err
}

// GetDeptByID 按 ID 查询部门
func GetDeptByID(id uint) (*model.Dept, error) {
	var dept model.Dept
	err := DB.First(&dept, id).Error
	return &dept, err
}

// CreateDept 新增部门
func CreateDept(dept *model.Dept) error {
	return DB.Create(dept).Error
}

// UpdateDept 更新部门
func UpdateDept(dept *model.Dept) error {
	return DB.Save(dept).Error
}

// DeleteDept 删除部门（表结构不建外键，应用层清理角色的部门数据权限关联）
func DeleteDept(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM sys_role_dept WHERE dept_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Dept{}, id).Error
	})
}

// CountDeptChildren 统计子部门数量
func CountDeptChildren(parentID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Dept{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// CountUserByDept 统计部门下用户数量
func CountUserByDept(deptID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.User{}).Where("dept_id = ?", deptID).Count(&count).Error
	return count, err
}

// BuildDeptTree 构建部门树（递归全深度，空 children 返回空数组）
func BuildDeptTree(depts []model.Dept) []model.Dept {
	childrenMap := make(map[uint][]model.Dept)
	for _, d := range depts {
		childrenMap[d.ParentID] = append(childrenMap[d.ParentID], d)
	}
	var build func(uint) []model.Dept
	build = func(parentID uint) []model.Dept {
		list := childrenMap[parentID]
		res := make([]model.Dept, 0, len(list))
		for i := range list {
			list[i].Children = build(list[i].ID)
			res = append(res, list[i])
		}
		return res
	}
	return build(0)
}
