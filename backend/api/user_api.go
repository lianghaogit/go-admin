package api

import (
	"strings"
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListUsers 分页查询用户列表（数据权限过滤）
func ListUsers(c *gin.Context) {
	q := dao.UserQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		Username: c.Query("username"),
		Phone:    c.Query("phone"),
		Status:   c.Query("status"),
		RoleID:   parseUint(c.Query("roleId")),
	}
	scopeSQL, scopeArgs, _ := dao.DataScopeUserCond(c.GetUint("userId"))
	users, total, err := dao.ListUsers(q, scopeSQL, scopeArgs)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, users, total)
}

// GetUser 查询用户详情（含角色）
func GetUser(c *gin.Context) {
	user, err := dao.GetUserByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	utils.OkData(c, user)
}

// UserReq 用户请求体
type UserReq struct {
	UserID   uint   `json:"userId"`
	DeptID   uint   `json:"deptId"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickName"`
	Email    string `json:"email"`
	Phone    string `json:"phonenumber"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
	RoleIds  []uint `json:"roleIds"`
	PostIds  []uint `json:"postIds"`
}

// CreateUser 新增用户
func CreateUser(c *gin.Context) {
	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		utils.Fail(c, 500, "用户名和密码不能为空")
		return
	}
	if count, _ := dao.CountUserByUsername(req.Username); count > 0 {
		utils.Fail(c, 500, "用户名已存在")
		return
	}
	if !checkAssignableRoles(c, req.RoleIds) {
		utils.Fail(c, 500, "不允许分配超级管理员角色")
		return
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(c, 500, "密码加密失败")
		return
	}
	user := &model.User{
		Username: req.Username, Password: hash, Nickname: req.Nickname,
		Email: req.Email, Phone: req.Phone, Status: req.Status, Remark: req.Remark, DeptID: req.DeptID,
	}
	if user.Status == "" {
		user.Status = "0"
	}
	if err := dao.CreateUser(user, req.RoleIds); err != nil {
		utils.Fail(c, 500, "新增用户失败")
		return
	}
	if err := dao.UpdateUserPosts(user.ID, req.PostIds); err != nil {
		utils.Fail(c, 500, "绑定岗位失败")
		return
	}
	utils.Ok(c)
}

// UpdateUser 修改用户
func UpdateUser(c *gin.Context) {
	var req UserReq
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetUserByID(req.UserID)
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	if req.Username != "" && req.Username != old.Username {
		utils.Fail(c, 500, "用户名不允许修改")
		return
	}
	if old.Username == "admin" && c.GetUint("userId") != old.ID {
		utils.Fail(c, 500, "超级管理员不允许被他人操作")
		return
	}
	if !dao.UserInDataScope(c.GetUint("userId"), req.UserID) {
		utils.Fail(c, 500, "无权操作该用户（超出数据权限范围）")
		return
	}
	if !checkAssignableRoles(c, req.RoleIds) {
		utils.Fail(c, 500, "不允许分配超级管理员角色")
		return
	}
	user := &model.User{
		ID: old.ID, Username: old.Username, Password: old.Password,
		Nickname: req.Nickname, Email: req.Email, Phone: req.Phone,
		Status: req.Status, Remark: req.Remark, DeptID: req.DeptID,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Fail(c, 500, "密码加密失败")
			return
		}
		user.Password = hash
	}
	if user.Status == "" {
		user.Status = "0"
	}
	if err := dao.UpdateUser(user, req.RoleIds); err != nil {
		utils.Fail(c, 500, "修改用户失败")
		return
	}
	if err := dao.UpdateUserPosts(user.ID, req.PostIds); err != nil {
		utils.Fail(c, 500, "绑定岗位失败")
		return
	}
	utils.Ok(c)
}

// DeleteUser 删除用户（支持逗号分隔批量删除）
func DeleteUser(c *gin.Context) {
	ids := strings.Split(c.Param("id"), ",")
	if len(ids) == 0 {
		utils.Fail(c, 500, "参数错误")
		return
	}
	curID := c.GetUint("userId")
	for _, s := range ids {
		id := parseUint(s)
		if id == 0 {
			utils.Fail(c, 500, "参数错误")
			return
		}
		if id == curID {
			utils.Fail(c, 500, "不允许删除当前登录用户")
			return
		}
		user, err := dao.GetUserByID(id)
		if err != nil {
			utils.Fail(c, 500, "用户不存在")
			return
		}
		if user.Username == "admin" {
			utils.Fail(c, 500, "超级管理员不允许删除")
			return
		}
		if !dao.UserInDataScope(curID, id) {
			utils.Fail(c, 500, "无权删除该用户（超出数据权限范围）")
			return
		}
	}
	if err := dao.DeleteUsers(ids); err != nil {
		utils.Fail(c, 500, "删除用户失败")
		return
	}
	utils.Ok(c)
}

// ChangeUserStatus 修改用户状态（启用/停用）
func ChangeUserStatus(c *gin.Context) {
	var req struct {
		UserID uint   `json:"userId"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 || (req.Status != "0" && req.Status != "1") {
		utils.Fail(c, 500, "参数错误")
		return
	}
	user, err := dao.GetUserByID(req.UserID)
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	if user.Username == "admin" {
		utils.Fail(c, 500, "超级管理员不允许停用")
		return
	}
	if !dao.UserInDataScope(c.GetUint("userId"), req.UserID) {
		utils.Fail(c, 500, "无权操作该用户（超出数据权限范围）")
		return
	}
	if err := dao.UpdateUserStatus(req.UserID, req.Status); err != nil {
		utils.Fail(c, 500, "修改状态失败")
		return
	}
	utils.Ok(c)
}

// AuthRole 分配用户角色（仅更新角色关联）
func AuthRole(c *gin.Context) {
	var req struct {
		UserID  uint   `json:"userId"`
		RoleIds []uint `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	user, err := dao.GetUserByID(req.UserID)
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	if user.Username == "admin" && c.GetUint("userId") != user.ID {
		utils.Fail(c, 500, "超级管理员不允许被他人操作")
		return
	}
	if !dao.UserInDataScope(c.GetUint("userId"), req.UserID) {
		utils.Fail(c, 500, "无权操作该用户（超出数据权限范围）")
		return
	}
	if !checkAssignableRoles(c, req.RoleIds) {
		utils.Fail(c, 500, "不允许分配超级管理员角色")
		return
	}
	if err := dao.UpdateUserRoles(req.UserID, req.RoleIds); err != nil {
		utils.Fail(c, 500, "分配角色失败")
		return
	}
	utils.Ok(c)
}

// ResetPwdReq 重置密码请求
type ResetPwdReq struct {
	UserID   uint   `json:"userId" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ResetPwd 重置用户密码
func ResetPwd(c *gin.Context) {
	var req ResetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Fail(c, 500, "密码加密失败")
		return
	}
	if !checkUserOperable(c, req.UserID) {
		utils.Fail(c, 500, "无权重置该用户密码")
		return
	}
	if err := dao.UpdateUserPassword(req.UserID, hash); err != nil {
		utils.Fail(c, 500, "重置密码失败")
		return
	}
	utils.Ok(c)
}

// checkUserOperable 校验当前登录用户是否有权操作目标用户（超管保护 + 数据权限）
func checkUserOperable(c *gin.Context, targetID uint) bool {
	operatorID := c.GetUint("userId")
	if dao.IsAdmin(targetID) && operatorID != targetID {
		return false
	}
	return dao.UserInDataScope(operatorID, targetID)
}

// checkAssignableRoles 非超管不允许分配内置超级管理员角色（roleId=1），防止提权
func checkAssignableRoles(c *gin.Context, roleIDs []uint) bool {
	if dao.IsAdmin(c.GetUint("userId")) {
		return true
	}
	for _, id := range roleIDs {
		if id == 1 {
			return false
		}
	}
	return true
}

// ---------- 个人中心 ----------

// GetProfile 查询当前用户个人信息
func GetProfile(c *gin.Context) {
	user, err := dao.GetUserByID(c.GetUint("userId"))
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	utils.OkData(c, user)
}

// UpdateProfileReq 更新个人信息请求
type UpdateProfileReq struct {
	NickName string `json:"nickName"`
	Email    string `json:"email"`
	Phone    string `json:"phonenumber"`
	Remark   string `json:"remark"`
}

// UpdateProfile 更新当前用户个人信息
func UpdateProfile(c *gin.Context) {
	var req UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	if err := dao.UpdateUserProfile(c.GetUint("userId"), req.NickName, req.Email, req.Phone, req.Remark); err != nil {
		utils.Fail(c, 500, "修改个人信息失败")
		return
	}
	utils.Ok(c)
}

// UpdateAvatar 更新当前用户头像
func UpdateAvatar(c *gin.Context) {
	var req struct {
		Avatar string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Avatar == "" {
		utils.Fail(c, 500, "头像地址不能为空")
		return
	}
	// 安全：仅允许本站上传路径（/uploads/ 前缀），拒绝外链/任意路径注入
	if !strings.HasPrefix(req.Avatar, "/uploads/") {
		utils.Fail(c, 500, "头像地址不合法")
		return
	}
	if err := dao.UpdateUserAvatar(c.GetUint("userId"), req.Avatar); err != nil {
		utils.Fail(c, 500, "头像更新失败")
		return
	}
	utils.Ok(c)
}

// UpdatePwdReq 修改密码请求
type UpdatePwdReq struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// UpdatePwd 修改当前用户密码
func UpdatePwd(c *gin.Context) {
	var req UpdatePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	user, err := dao.GetUserByID(c.GetUint("userId"))
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	if !utils.CheckPassword(user.Password, req.OldPassword) {
		utils.Fail(c, 500, "修改密码失败，旧密码错误")
		return
	}
	hash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.Fail(c, 500, "密码加密失败")
		return
	}
	if err := dao.UpdateUserPassword(user.ID, hash); err != nil {
		utils.Fail(c, 500, "修改密码失败")
		return
	}
	utils.Ok(c)
}
