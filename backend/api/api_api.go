package api

import (
	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListApis 分页查询接口权限
func ListApis(c *gin.Context) {
	q := struct {
		PageNum  int
		PageSize int
		Name     string
		Path     string
		Status   string
	}{PageNum: pageNum(c), PageSize: pageSize(c), Name: c.Query("apiName"), Path: c.Query("path"), Status: c.Query("status")}
	apis, total, err := dao.ListApis(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, apis, total)
}

// GetApi 查询接口详情
func GetApi(c *gin.Context) {
	api, err := dao.GetApiByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "接口不存在")
		return
	}
	utils.OkData(c, api)
}

// ApiReq 接口请求体
type ApiReq struct {
	APIID   uint   `json:"apiId"`
	APIName string `json:"apiName"`
	Path    string `json:"path"`
	Method  string `json:"method"`
	Perms   string `json:"perms"`
	Status  string `json:"status"`
	Remark  string `json:"remark"`
}

// CreateApi 新增接口
func CreateApi(c *gin.Context) {
	var req ApiReq
	if err := c.ShouldBindJSON(&req); err != nil || req.APIName == "" || req.Path == "" || req.Method == "" {
		utils.Fail(c, 500, "接口名称、路径和方法不能为空")
		return
	}
	dup, err := dao.CountApiByPathMethod(req.Path, req.Method, 0)
	if err != nil {
		utils.Fail(c, 500, "校验接口唯一性失败")
		return
	}
	if dup > 0 {
		utils.Fail(c, 500, "该路径与方法的接口已存在")
		return
	}
	a := &model.Api{APIName: req.APIName, Path: req.Path, Method: req.Method, Perms: req.Perms, Status: req.Status, Remark: req.Remark}
	if a.Status == "" {
		a.Status = "0"
	}
	if err := dao.CreateApi(a); err != nil {
		utils.Fail(c, 500, "新增接口失败")
		return
	}
	utils.Ok(c)
}

// UpdateApi 修改接口
func UpdateApi(c *gin.Context) {
	var req ApiReq
	if err := c.ShouldBindJSON(&req); err != nil || req.APIID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetApiByID(req.APIID)
	if err != nil {
		utils.Fail(c, 500, "接口不存在")
		return
	}
	if req.Path != old.Path || req.Method != old.Method {
		dup, err := dao.CountApiByPathMethod(req.Path, req.Method, old.APIID)
		if err != nil {
			utils.Fail(c, 500, "校验接口唯一性失败")
			return
		}
		if dup > 0 {
			utils.Fail(c, 500, "该路径与方法的接口已存在")
			return
		}
	}
	a := &model.Api{APIID: old.APIID, APIName: req.APIName, Path: req.Path, Method: req.Method, Perms: req.Perms, Status: req.Status, Remark: req.Remark}
	if a.Status == "" {
		a.Status = "0"
	}
	if err := dao.UpdateApi(a); err != nil {
		utils.Fail(c, 500, "修改接口失败")
		return
	}
	utils.Ok(c)
}

// ChangeApiStatus 启停接口
func ChangeApiStatus(c *gin.Context) {
	var req struct {
		APIID  uint   `json:"apiId"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.APIID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	if err := dao.ChangeApiStatus(req.APIID, req.Status); err != nil {
		utils.Fail(c, 500, "修改接口状态失败")
		return
	}
	utils.Ok(c)
}

// DeleteApi 删除接口
func DeleteApi(c *gin.Context) {
	if err := dao.DeleteApi(idParam(c)); err != nil {
		utils.Fail(c, 500, "删除接口失败")
		return
	}
	utils.Ok(c)
}
