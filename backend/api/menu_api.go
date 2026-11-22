package api

import (
	"strconv"
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListMenus 菜单列表（树形）
func ListMenus(c *gin.Context) {
	menus, err := dao.ListMenus()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, dao.BuildMenuTree(menus))
}

// MenuTreeSelect 菜单下拉树
func MenuTreeSelect(c *gin.Context) {
	menus, err := dao.ListMenus()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, dao.BuildMenuTree(menus))
}

// RoleMenuTreeSelect 角色菜单树（含已选菜单 ID）
func RoleMenuTreeSelect(c *gin.Context) {
	menus, err := dao.ListMenus()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	roleID, _ := strconv.ParseUint(c.Param("roleId"), 10, 64)
	checked, err := dao.GetCheckedMenuIDs(uint(roleID))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, gin.H{"menus": dao.BuildMenuTree(menus), "checkedKeys": checked})
}

// GetMenu 查询菜单详情
func GetMenu(c *gin.Context) {
	menu, err := dao.GetMenuByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "菜单不存在")
		return
	}
	utils.OkData(c, menu)
}

// MenuReq 菜单请求体
type MenuReq struct {
	MenuID    uint   `json:"menuId"`
	MenuName  string `json:"menuName"`
	ParentID  uint   `json:"parentId"`
	OrderNum  int    `json:"orderNum"`
	Path      string `json:"path"`
	Component string `json:"component"`
	MenuType  string `json:"menuType"`
	Perms     string `json:"perms"`
	Icon      string `json:"icon"`
	Visible   string `json:"visible"`
	Status    string `json:"status"`
}

// CreateMenu 新增菜单
func CreateMenu(c *gin.Context) {
	var req MenuReq
	if err := c.ShouldBindJSON(&req); err != nil || req.MenuName == "" {
		utils.Fail(c, 500, "菜单名称不能为空")
		return
	}
	menu := &model.Menu{
		MenuName: req.MenuName, ParentID: req.ParentID, OrderNum: req.OrderNum,
		Path: req.Path, Component: req.Component, MenuType: req.MenuType,
		Perms: req.Perms, Icon: req.Icon, Visible: req.Visible, Status: req.Status,
	}
	if menu.MenuType == "" {
		menu.MenuType = "C"
	}
	if menu.Visible == "" {
		menu.Visible = "0"
	}
	if menu.Status == "" {
		menu.Status = "0"
	}
	if err := dao.CreateMenu(menu); err != nil {
		utils.Fail(c, 500, "新增菜单失败")
		return
	}
	utils.Ok(c)
}

// UpdateMenu 修改菜单
func UpdateMenu(c *gin.Context) {
	var req MenuReq
	if err := c.ShouldBindJSON(&req); err != nil || req.MenuID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetMenuByID(req.MenuID)
	if err != nil {
		utils.Fail(c, 500, "菜单不存在")
		return
	}
	if req.ParentID == req.MenuID {
		utils.Fail(c, 500, "上级菜单不能选择自己")
		return
	}
	menu := &model.Menu{
		ID: old.ID, MenuName: req.MenuName, ParentID: req.ParentID, OrderNum: req.OrderNum,
		Path: req.Path, Component: req.Component, MenuType: req.MenuType,
		Perms: req.Perms, Icon: req.Icon, Visible: req.Visible, Status: req.Status,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if menu.MenuType == "" {
		menu.MenuType = "C"
	}
	if menu.Visible == "" {
		menu.Visible = "0"
	}
	if menu.Status == "" {
		menu.Status = "0"
	}
	if err := dao.UpdateMenu(menu); err != nil {
		utils.Fail(c, 500, "修改菜单失败")
		return
	}
	utils.Ok(c)
}

// DeleteMenu 删除菜单
func DeleteMenu(c *gin.Context) {
	id := idParam(c)
	if count, _ := dao.CountChildren(id); count > 0 {
		utils.Fail(c, 500, "存在子菜单，不允许删除")
		return
	}
	if count, _ := dao.CountMenuInRole(id); count > 0 {
		utils.Fail(c, 500, "菜单已分配，不允许删除")
		return
	}
	if err := dao.DeleteMenu(id); err != nil {
		utils.Fail(c, 500, "删除菜单失败")
		return
	}
	utils.Ok(c)
}
