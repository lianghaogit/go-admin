package dao

import (
	"strconv"

	"go-admin/model"

	"gorm.io/gorm"
)

// UserQuery 用户列表查询条件
type UserQuery struct {
	PageNum  int
	PageSize int
	Username string
	Phone    string
	Status   string
	RoleID   uint // 传入时返回「未分配该角色」的用户（角色分配用户弹窗用）
}

// GetUserByUsername 按用户名查找用户
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := DB.Preload("Roles").Where("username = ?", username).First(&user).Error
	return &user, err
}

// GetUserAuthInfo 只取鉴权需要的字段，避免每次请求预加载 Roles/Posts/Dept。
// 需要角色/岗位/部门的业务接口仍使用 GetUserByID。
func GetUserAuthInfo(id uint) (*model.User, error) {
	var user model.User
	err := DB.Select("id", "username", "status").First(&user, id).Error
	return &user, err
}

// GetUserByID 按 ID 查找用户
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := DB.Preload("Roles").Preload("Posts").Preload("Dept").First(&user, id).Error
	return &user, err
}

// SuperAdminUsername 内置超级管理员账号名（当前系统唯一超管口径）
const SuperAdminUsername = "admin"

// IsSuperAdminName 统一的超管判定：只认内置账号名。
// 需要 userID 的场景请用 IsAdmin，二者最终走同一判定。
func IsSuperAdminName(username string) bool {
	return username == SuperAdminUsername
}

// IsAdmin 判断是否为内置超级管理员（拥有全部权限）
func IsAdmin(userID uint) bool {
	var user model.User
	if err := DB.Select("username").First(&user, userID).Error; err != nil {
		return false
	}
	return IsSuperAdminName(user.Username)
}

// ListUsers 分页查询用户（按数据权限注入部门/本人过滤）
func ListUsers(q UserQuery, scopeSQL string, scopeArgs []interface{}) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	db := DB.Model(&model.User{})
	if q.Username != "" {
		db = db.Where(likeExprEsc("username"), "%"+escapeLike(q.Username)+"%")
	}
	if q.Phone != "" {
		db = db.Where(likeExprEsc("phone"), "%"+escapeLike(q.Phone)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.RoleID > 0 {
		// 未分配该角色的用户（角色分配用户弹窗）
		db = db.Where("id NOT IN (SELECT user_id FROM sys_user_role WHERE role_id = ?)", q.RoleID)
	}
	if scopeSQL != "" {
		db = db.Where(scopeSQL, scopeArgs...)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Dept").Order("id asc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&users).Error
	return users, total, err
}

// ListUsersAll 导出全部符合条件的用户（不分页，数据权限过滤）
func ListUsersAll(q UserQuery, scopeSQL string, scopeArgs []interface{}) ([]model.User, error) {
	var users []model.User
	db := DB.Model(&model.User{})
	if q.Username != "" {
		db = db.Where(likeExprEsc("username"), "%"+escapeLike(q.Username)+"%")
	}
	if q.Phone != "" {
		db = db.Where(likeExprEsc("phone"), "%"+escapeLike(q.Phone)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if scopeSQL != "" {
		db = db.Where(scopeSQL, scopeArgs...)
	}
	err := db.Preload("Dept").Order("id asc").Find(&users).Error
	return users, err
}

// UpdateUserImport 导入更新用户（昵称/邮箱/手机/状态/备注/部门）
func UpdateUserImport(id uint, nickname, email, phone, status, remark string, deptID uint) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"nickname": nickname,
		"email":    email,
		"phone":    phone,
		"status":   status,
		"remark":   remark,
		"dept_id":  deptID,
	}).Error
}

// GetDeptByName 按部门名称查询
func GetDeptByName(name string) (*model.Dept, error) {
	var dept model.Dept
	err := DB.Where("dept_name = ?", name).First(&dept).Error
	return &dept, err
}

// GetConfigValue 读取系统参数值
func GetConfigValue(key string) string {
	var config model.Config
	if err := DB.Where("config_key = ?", key).First(&config).Error; err != nil {
		return ""
	}
	return config.ConfigValue
}

// UpdateUserPosts 替换用户岗位关联（清空后按 ID 加载模型重建）
func UpdateUserPosts(userID uint, postIDs []uint) error {
	if postIDs == nil {
		return nil
	}
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return err
	}
	if len(postIDs) == 0 {
		return DB.Model(&user).Association("Posts").Clear()
	}
	var posts []model.Post
	if err := DB.Where("id IN ?", postIDs).Find(&posts).Error; err != nil {
		return err
	}
	return DB.Model(&user).Association("Posts").Replace(posts)
}

// UpdateUserProfile 更新当前用户个人信息（昵称/邮箱/手机/备注）
func UpdateUserProfile(id uint, nickname, email, phone, remark string) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"nickname": nickname,
		"email":    email,
		"phone":    phone,
		"remark":   remark,
	}).Error
}

// UpdateUserAvatar 更新当前用户头像
func UpdateUserAvatar(id uint, avatar string) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Update("avatar", avatar).Error
}

// CountUserByUsername 统计同名用户数
func CountUserByUsername(username string) (int64, error) {
	var count int64
	err := DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count, err
}

// CreateUser 创建用户并绑定角色
func CreateUser(user *model.User, roleIDs []uint) error {
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return replaceUserRoles(tx, user, roleIDs)
	})
}

// UpdateUser 更新用户（含角色）
func UpdateUser(user *model.User, roleIDs []uint) error {
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(user).Error; err != nil {
			return err
		}
		return replaceUserRoles(tx, user, roleIDs)
	})
}

func replaceUserRoles(tx *gorm.DB, user *model.User, roleIDs []uint) error {
	if roleIDs == nil {
		return nil
	}
	if len(roleIDs) == 0 {
		return tx.Model(user).Association("Roles").Clear()
	}
	var roles []model.Role
	if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	return tx.Model(user).Association("Roles").Replace(roles)
}

// DeleteUser 删除用户（同时清理角色、岗位关联）
func DeleteUser(id uint) error {
	defer InvalidatePermCache()
	return DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}
		// 表结构不建外键，岗位关联需在应用层清理，避免 sys_user_post 出现孤儿行
		if err := tx.Model(&user).Association("Posts").Clear(); err != nil {
			return err
		}
		return tx.Delete(&user).Error
	})
}

// DeleteUsers 批量删除用户（同时清理角色、岗位关联）
func DeleteUsers(ids []string) error {
	defer InvalidatePermCache()
	var userIDs []uint
	for _, s := range ids {
		v, _ := strconv.ParseUint(s, 10, 64)
		if v > 0 {
			userIDs = append(userIDs, uint(v))
		}
	}
	if len(userIDs) == 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var users []model.User
		if err := tx.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return err
		}
		for i := range users {
			if err := tx.Model(&users[i]).Association("Roles").Clear(); err != nil {
				return err
			}
			if err := tx.Model(&users[i]).Association("Posts").Clear(); err != nil {
				return err
			}
		}
		return tx.Where("id IN ?", userIDs).Delete(&model.User{}).Error
	})
}

// UpdateUserStatus 修改用户状态
func UpdateUserStatus(id uint, status string) error {
	defer InvalidatePermCache()
	return DB.Model(&model.User{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(id uint, hash string) error {
	return DB.Model(&model.User{}).Where("id = ?", id).Update("password", hash).Error
}

// UpdateUserRoles 仅更新用户角色关联（分配角色弹窗）
func UpdateUserRoles(userID uint, roleIDs []uint) error {
	defer InvalidatePermCache()
	var user model.User
	if err := DB.First(&user, userID).Error; err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		return replaceUserRoles(tx, &user, roleIDs)
	})
}

// ListUsersByRole 查询角色的已分配/未分配用户
func ListUsersByRole(roleID uint, allocated bool, pageNum, pageSize int, username string) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	if allocated {
		db := DB.Model(&model.User{}).
			Joins("JOIN sys_user_role ur ON ur.user_id = sys_user.id").
			Where("ur.role_id = ?", roleID)
		if username != "" {
			db = db.Where(likeExprEsc("sys_user.username"), "%"+escapeLike(username)+"%")
		}
		if err := db.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		err := db.Preload("Dept").Order("sys_user.id asc").
			Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&users).Error
		return users, total, err
	}
	db := DB.Model(&model.User{}).Where("id NOT IN (SELECT user_id FROM sys_user_role WHERE role_id = ?)", roleID)
	if username != "" {
		db = db.Where(likeExprEsc("username"), "%"+escapeLike(username)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Dept").Order("id asc").
		Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&users).Error
	return users, total, err
}

// GetRoleKeysByUser 查询用户角色标识
func GetRoleKeysByUser(userID uint) ([]string, error) {
	var keys []string
	err := DB.Raw(`SELECT DISTINCT r.role_key FROM sys_role r
		JOIN sys_user_role ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.status = '0'`, userID).Scan(&keys).Error
	return keys, err
}

// GetPermsByUser 查询用户按钮权限标识（含 C 型菜单 perms，本项目权限字符挂在菜单/按钮上）
// 优先读 30 秒进程内缓存；角色/菜单/授权变更时由 InvalidatePermCache 失效。
func GetPermsByUser(userID uint) ([]string, error) {
	if perms, ok := getCachedPerms(userID); ok {
		return perms, nil
	}
	var perms []string
	err := DB.Raw(`SELECT DISTINCT m.perms FROM sys_menu m
		JOIN sys_role_menu rm ON m.id = rm.menu_id
		JOIN sys_user_role ur ON rm.role_id = ur.role_id
		JOIN sys_role r ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.status = '0' AND m.menu_type IN ('C','F') AND m.status = '0' AND m.perms <> ''`, userID).Scan(&perms).Error
	if err != nil {
		return nil, err
	}
	cachePerms(userID, perms)
	return cloneStrings(perms), nil
}
