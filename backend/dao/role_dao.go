package dao

import (
	"errors"

	"go-admin/model"

	"gorm.io/gorm"
)

// ErrRoleAssigned 角色已分配用户，不允许删除
var ErrRoleAssigned = errors.New("角色已分配用户，不允许删除")

// RoleQuery 角色列表查询条件
type RoleQuery struct {
	PageNum  int
	PageSize int
	RoleName string
	Status   string
}

// ListRoles 分页查询角色
func ListRoles(q RoleQuery) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	db := DB.Model(&model.Role{})
	if q.RoleName != "" {
		db = db.Where(likeExprEsc("role_name"), "%"+escapeLike(q.RoleName)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("sort asc, id asc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&roles).Error
	return roles, total, err
}

// ListAllRoles 查询全部角色（下拉选择用）
func ListAllRoles() ([]model.Role, error) {
	var roles []model.Role
	err := DB.Order("sort asc").Find(&roles).Error
	return roles, err
}

// GetRoleByID 按 ID 查询角色
func GetRoleByID(id uint) (*model.Role, error) {
	var role model.Role
	err := DB.Preload("Menus").First(&role, id).Error
	return &role, err
}

// CountRoleByKey 统计相同权限字符的角色数
func CountRoleByKey(key string) (int64, error) {
	var count int64
	err := DB.Model(&model.Role{}).Where("role_key = ?", key).Count(&count).Error
	return count, err
}

// CountUserByRole 统计关联某角色的用户数
func CountUserByRole(roleID uint) (int64, error) {
	var count int64
	err := DB.Table("sys_user_role").Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

// CreateRoleWithMenus 创建角色并绑定菜单与数据权限部门
func CreateRoleWithMenus(role *model.Role, menuIDs []uint, deptIDs []uint) error {
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		if err := bindRoleMenus(tx, role, menuIDs); err != nil {
			return err
		}
		return bindRoleDepts(tx, role, deptIDs)
	})
}

// UpdateRoleWithMenus 更新角色并绑定菜单与数据权限部门
func UpdateRoleWithMenus(role *model.Role, menuIDs []uint, deptIDs []uint) error {
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(role).Error; err != nil {
			return err
		}
		if err := bindRoleMenus(tx, role, menuIDs); err != nil {
			return err
		}
		return bindRoleDepts(tx, role, deptIDs)
	})
}

// bindRoleDepts 绑定角色数据权限部门（自定义范围）
func bindRoleDepts(tx *gorm.DB, role *model.Role, deptIDs []uint) error {
	if deptIDs == nil {
		return nil
	}
	if len(deptIDs) == 0 {
		return tx.Model(role).Association("Depts").Clear()
	}
	var depts []model.Dept
	if err := tx.Where("id IN ?", deptIDs).Find(&depts).Error; err != nil {
		return err
	}
	return tx.Model(role).Association("Depts").Replace(depts)
}

// GetRoleDeptIDs 查询角色数据权限部门 ID
func GetRoleDeptIDs(roleID uint) []uint {
	var ids []uint
	_ = DB.Table("sys_role_dept").Where("role_id = ?", roleID).Pluck("dept_id", &ids).Error
	return ids
}

func bindRoleMenus(tx *gorm.DB, role *model.Role, menuIDs []uint) error {
	if menuIDs == nil {
		return nil
	}
	if len(menuIDs) == 0 {
		return tx.Model(role).Association("Menus").Clear()
	}
	var menus []model.Menu
	if err := tx.Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
		return err
	}
	return tx.Model(role).Association("Menus").Replace(menus)
}

// AssignUsersToRole 批量分配用户给角色
func AssignUsersToRole(roleID uint, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		var role model.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}
		var users []model.User
		if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
		// 追加关联（不覆盖已有）
		cur := tx.Model(&role).Association("Users")
		for i := range users {
			if err := cur.Append(&users[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveUsersFromRole 批量移除角色的用户
func RemoveUsersFromRole(roleID uint, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		var role model.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return err
		}
		var users []model.User
		if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
		cur := tx.Model(&role).Association("Users")
		for i := range users {
			if err := cur.Delete(&users[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteRole 删除角色（同时清理菜单与数据权限部门关联）
// 对标 RuoYi：已分配用户的角色不允许删除（避免 sys_user_role 孤儿关联）
func DeleteRole(id uint) error {
	defer InvalidatePermCache()
	var assigned int64
	if err := DB.Table("sys_user_role").Where("role_id = ?", id).Count(&assigned).Error; err != nil {
		return err
	}
	if assigned > 0 {
		return ErrRoleAssigned
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var role model.Role
		if err := tx.First(&role, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&role).Association("Menus").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&role).Association("Depts").Clear(); err != nil {
			return err
		}
		return tx.Delete(&role).Error
	})
}
