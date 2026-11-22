package api

import (
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListDepts 部门列表（树形，按数据权限过滤）
func ListDepts(c *gin.Context) {
	depts, err := dao.ListDepts()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	tree := dao.BuildDeptTree(depts)
	if ids, limited := dao.DataScopeDeptIDs(c.GetUint("userId")); limited {
		tree = dao.FilterDeptTree(tree, dao.VisibleDeptMap(ids))
	}
	utils.OkData(c, tree)
}

// DeptTreeSelect 部门下拉树（按数据权限过滤）
func DeptTreeSelect(c *gin.Context) {
	depts, err := dao.ListDepts()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	tree := dao.BuildDeptTree(depts)
	if ids, limited := dao.DataScopeDeptIDs(c.GetUint("userId")); limited {
		tree = dao.FilterDeptTree(tree, dao.VisibleDeptMap(ids))
	}
	utils.OkData(c, tree)
}

// GetDept 查询部门详情
func GetDept(c *gin.Context) {
	dept, err := dao.GetDeptByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "部门不存在")
		return
	}
	utils.OkData(c, dept)
}

// DeptReq 部门请求体
type DeptReq struct {
	DeptID   uint   `json:"deptId"`
	ParentID uint   `json:"parentId"`
	DeptName string `json:"deptName"`
	OrderNum int    `json:"orderNum"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// CreateDept 新增部门
func CreateDept(c *gin.Context) {
	var req DeptReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DeptName == "" {
		utils.Fail(c, 500, "部门名称不能为空")
		return
	}
	dept := &model.Dept{
		ParentID: req.ParentID, DeptName: req.DeptName, OrderNum: req.OrderNum,
		Leader: req.Leader, Phone: req.Phone, Email: req.Email, Status: req.Status,
	}
	if dept.Status == "" {
		dept.Status = "0"
	}
	if err := dao.CreateDept(dept); err != nil {
		utils.Fail(c, 500, "新增部门失败")
		return
	}
	utils.Ok(c)
}

// UpdateDept 修改部门
func UpdateDept(c *gin.Context) {
	var req DeptReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DeptID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetDeptByID(req.DeptID)
	if err != nil {
		utils.Fail(c, 500, "部门不存在")
		return
	}
	if req.ParentID == req.DeptID {
		utils.Fail(c, 500, "上级部门不能选择自己")
		return
	}
	dept := &model.Dept{
		ID: old.ID, ParentID: req.ParentID, DeptName: req.DeptName, OrderNum: req.OrderNum,
		Leader: req.Leader, Phone: req.Phone, Email: req.Email, Status: req.Status,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if dept.Status == "" {
		dept.Status = "0"
	}
	if err := dao.UpdateDept(dept); err != nil {
		utils.Fail(c, 500, "修改部门失败")
		return
	}
	utils.Ok(c)
}

// DeleteDept 删除部门
func DeleteDept(c *gin.Context) {
	id := idParam(c)
	if count, _ := dao.CountDeptChildren(id); count > 0 {
		utils.Fail(c, 500, "存在下级部门，不允许删除")
		return
	}
	if count, _ := dao.CountUserByDept(id); count > 0 {
		utils.Fail(c, 500, "部门下存在用户，不允许删除")
		return
	}
	if err := dao.DeleteDept(id); err != nil {
		utils.Fail(c, 500, "删除部门失败")
		return
	}
	utils.Ok(c)
}
