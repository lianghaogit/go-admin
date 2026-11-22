package dao

import (
	"go-admin/model"
)

// ListMenus 查询全部菜单（按父子关系与排序）
func ListMenus() ([]model.Menu, error) {
	var menus []model.Menu
	err := DB.Order("parent_id asc, order_num asc, id asc").Find(&menus).Error
	return menus, err
}

// GetMenuByID 按 ID 查询菜单
func GetMenuByID(id uint) (*model.Menu, error) {
	var menu model.Menu
	err := DB.First(&menu, id).Error
	return &menu, err
}

// CreateMenu 新增菜单
func CreateMenu(menu *model.Menu) error {
	defer InvalidatePermCache()
	return DB.Create(menu).Error
}

// UpdateMenu 更新菜单
func UpdateMenu(menu *model.Menu) error {
	defer InvalidatePermCache()
	return DB.Save(menu).Error
}

// DeleteMenu 删除菜单
func DeleteMenu(id uint) error {
	defer InvalidatePermCache()
	return DB.Delete(&model.Menu{}, id).Error
}

// CountChildren 统计子菜单数量
func CountChildren(parentID uint) (int64, error) {
	var count int64
	err := DB.Model(&model.Menu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// CountMenuInRole 统计菜单被角色引用次数
func CountMenuInRole(menuID uint) (int64, error) {
	var count int64
	err := DB.Table("sys_role_menu").Where("menu_id = ?", menuID).Count(&count).Error
	return count, err
}

// GetMenusByUser 查询用户可见菜单（admin 返回全部，其他用户按角色关联查询）
func GetMenusByUser(userID uint) ([]model.Menu, error) {
	var menus []model.Menu
	if IsAdmin(userID) {
		err := DB.Where("status = '0'").Order("parent_id asc, order_num asc, id asc").Find(&menus).Error
		return menus, err
	}
	err := DB.Raw(`SELECT DISTINCT m.* FROM sys_menu m
		JOIN sys_role_menu rm ON m.id = rm.menu_id
		JOIN sys_user_role ur ON rm.role_id = ur.role_id
		JOIN sys_role r ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.status = '0' AND m.status = '0'
		ORDER BY m.parent_id asc, m.order_num asc, m.id asc`, userID).Scan(&menus).Error
	return menus, err
}

// BuildMenuTree 构建菜单树（递归全深度，空 children 返回空数组）
func BuildMenuTree(menus []model.Menu) []model.Menu {
	childrenMap := make(map[uint][]model.Menu)
	for _, m := range menus {
		childrenMap[m.ParentID] = append(childrenMap[m.ParentID], m)
	}
	var build func(uint) []model.Menu
	build = func(parentID uint) []model.Menu {
		list := childrenMap[parentID]
		res := make([]model.Menu, 0, len(list))
		for i := range list {
			list[i].Children = build(list[i].ID)
			res = append(res, list[i])
		}
		return res
	}
	return build(0)
}

// GetCheckedMenuIDs 查询角色已分配的菜单 ID
func GetCheckedMenuIDs(roleID uint) ([]uint, error) {
	var ids []uint
	err := DB.Table("sys_role_menu").Where("role_id = ?", roleID).Pluck("menu_id", &ids).Error
	return ids, err
}
