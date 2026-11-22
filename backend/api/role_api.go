package api

import (
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListRoles 分页查询角色
func ListRoles(c *gin.Context) {
	q := dao.RoleQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		RoleName: c.Query("roleName"),
		Status:   c.Query("status"),
	}
	roles, total, err := dao.ListRoles(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, roles, total)
}

// ListAllRoles 查询全部角色（下拉选择用）
func ListAllRoles(c *gin.Context) {
	roles, err := dao.ListAllRoles()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, roles)
}

// GetRole 查询角色详情
func GetRole(c *gin.Context) {
	role, err := dao.GetRoleByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "角色不存在")
		return
	}
	utils.OkData(c, gin.H{
		"roleId": role.ID, "roleName": role.RoleName, "roleKey": role.RoleKey,
		"roleSort": role.Sort, "status": role.Status, "dataScope": role.DataScope,
		"remark": role.Remark, "createTime": role.CreatedAt, "updateTime": role.UpdatedAt,
		"deptIds": dao.GetRoleDeptIDs(role.ID),
	})
}

// RoleReq 角色请求体
type RoleReq struct {
	RoleID    uint   `json:"roleId"`
	RoleName  string `json:"roleName"`
	RoleKey   string `json:"roleKey"`
	Sort      int    `json:"roleSort"`
	Status    string `json:"status"`
	DataScope string `json:"dataScope"`
	Remark    string `json:"remark"`
	MenuIds   []uint `json:"menuIds"`
	DeptIds   []uint `json:"deptIds"`
}

// CreateRole 新增角色
func CreateRole(c *gin.Context) {
	var req RoleReq
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleName == "" || req.RoleKey == "" {
		utils.Fail(c, 500, "角色名称和权限字符不能为空")
		return
	}
	if count, _ := dao.CountRoleByKey(req.RoleKey); count > 0 {
		utils.Fail(c, 500, "角色权限字符已存在")
		return
	}
	role := &model.Role{RoleName: req.RoleName, RoleKey: req.RoleKey, Sort: req.Sort, Status: req.Status, DataScope: req.DataScope, Remark: req.Remark}
	if role.Status == "" {
		role.Status = "0"
	}
	if role.DataScope == "" {
		role.DataScope = "1"
	}
	if err := dao.CreateRoleWithMenus(role, req.MenuIds, req.DeptIds); err != nil {
		utils.Fail(c, 500, "新增角色失败")
		return
	}
	utils.Ok(c)
}

// UpdateRole 修改角色
func UpdateRole(c *gin.Context) {
	var req RoleReq
	if err := c.ShouldBindJSON(&req); err != nil || req.RoleID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetRoleByID(req.RoleID)
	if err != nil {
		utils.Fail(c, 500, "角色不存在")
		return
	}
	// 超级管理员角色（id=1）仅允许管理员本人修改，防止普通角色管理员借此提权
	if req.RoleID == 1 && !dao.IsAdmin(c.GetUint("userId")) {
		utils.Fail(c, 500, "超级管理员角色不允许修改")
		return
	}
	if req.RoleKey != old.RoleKey {
		if count, _ := dao.CountRoleByKey(req.RoleKey); count > 0 {
			utils.Fail(c, 500, "角色权限字符已存在")
			return
		}
	}
	role := &model.Role{
		ID: old.ID, RoleName: req.RoleName, RoleKey: req.RoleKey, Sort: req.Sort,
		Status: req.Status, DataScope: req.DataScope, Remark: req.Remark, CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if role.Status == "" {
		role.Status = "0"
	}
	if role.DataScope == "" {
		role.DataScope = "1"
	}
	if err := dao.UpdateRoleWithMenus(role, req.MenuIds, req.DeptIds); err != nil {
		utils.Fail(c, 500, "修改角色失败")
		return
	}
	utils.Ok(c)
}

// ListRoleAllocatedUsers 查询角色已分配用户
func ListRoleAllocatedUsers(c *gin.Context) {
	roleID := parseUint(c.Query("roleId"))
	if roleID == 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	users, total, err := dao.ListUsersByRole(roleID, true, pageNum(c), pageSize(c), c.Query("username"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, users, total)
}

// ListRoleUnallocatedUsers 查询角色未分配用户
func ListRoleUnallocatedUsers(c *gin.Context) {
	roleID := parseUint(c.Query("roleId"))
	if roleID == 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	users, total, err := dao.ListUsersByRole(roleID, false, pageNum(c), pageSize(c), c.Query("username"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, users, total)
}

// AuthUserReq 分配/取消用户请求体
type AuthUserReq struct {
	RoleID  uint   `json:"roleId" binding:"required"`
	UserIDs []uint `json:"userIds"`
}

// AssignAuthUsers 角色批量分配用户
func AssignAuthUsers(c *gin.Context) {
	var req AuthUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	// 超级管理员角色仅允许超管本人分配，防止普通角色管理员借分配用户提权
	if req.RoleID == 1 && !dao.IsAdmin(c.GetUint("userId")) {
		utils.Fail(c, 500, "超级管理员角色不允许分配")
		return
	}
	operatorID := c.GetUint("userId")
	for _, uid := range req.UserIDs {
		if !dao.UserInDataScope(operatorID, uid) {
			utils.Fail(c, 500, "无权操作目标用户（超出数据权限范围）")
			return
		}
	}
	if err := dao.AssignUsersToRole(req.RoleID, req.UserIDs); err != nil {
		utils.Fail(c, 500, "分配用户失败")
		return
	}
	utils.Ok(c)
}

// CancelAuthUsers 角色批量取消用户
func CancelAuthUsers(c *gin.Context) {
	var req AuthUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	if req.RoleID == 1 && !dao.IsAdmin(c.GetUint("userId")) {
		utils.Fail(c, 500, "超级管理员角色不允许操作")
		return
	}
	operatorID := c.GetUint("userId")
	for _, uid := range req.UserIDs {
		if !dao.UserInDataScope(operatorID, uid) {
			utils.Fail(c, 500, "无权操作目标用户（超出数据权限范围）")
			return
		}
	}
	if err := dao.RemoveUsersFromRole(req.RoleID, req.UserIDs); err != nil {
		utils.Fail(c, 500, "取消分配失败")
		return
	}
	utils.Ok(c)
}

// DeleteRole 删除角色
func DeleteRole(c *gin.Context) {
	id := idParam(c)
	if id == 1 {
		utils.Fail(c, 500, "超级管理员角色不允许删除")
		return
	}
	if count, _ := dao.CountUserByRole(id); count > 0 {
		utils.Fail(c, 500, "该角色已分配用户，不能删除")
		return
	}
	if err := dao.DeleteRole(id); err != nil {
		utils.Fail(c, 500, "删除角色失败")
		return
	}
	utils.Ok(c)
}
