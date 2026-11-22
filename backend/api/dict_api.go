package api

import (
	"strings"
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ---------- 字典类型 ----------

// ListDictTypes 分页查询字典类型
func ListDictTypes(c *gin.Context) {
	types, total, err := dao.ListDictTypes(pageNum(c), pageSize(c), c.Query("dictName"), c.Query("dictType"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, types, total)
}

// GetDictType 查询字典类型详情
func GetDictType(c *gin.Context) {
	t, err := dao.GetDictTypeByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "字典类型不存在")
		return
	}
	utils.OkData(c, t)
}

// DictTypeReq 字典类型请求体
type DictTypeReq struct {
	DictID   uint   `json:"dictId"`
	DictName string `json:"dictName"`
	DictType string `json:"dictType"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
}

// CreateDictType 新增字典类型
func CreateDictType(c *gin.Context) {
	var req DictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DictName == "" || req.DictType == "" {
		utils.Fail(c, 500, "字典名称和字典类型不能为空")
		return
	}
	if count, _ := dao.CountDictTypeByType(req.DictType); count > 0 {
		utils.Fail(c, 500, "字典类型已存在")
		return
	}
	t := &model.DictType{DictName: req.DictName, DictType: req.DictType, Status: req.Status, Remark: req.Remark}
	if t.Status == "" {
		t.Status = "0"
	}
	if err := dao.CreateDictType(t); err != nil {
		utils.Fail(c, 500, "新增字典类型失败")
		return
	}
	utils.Ok(c)
}

// UpdateDictType 修改字典类型
func UpdateDictType(c *gin.Context) {
	var req DictTypeReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DictID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetDictTypeByID(req.DictID)
	if err != nil {
		utils.Fail(c, 500, "字典类型不存在")
		return
	}
	if req.DictType != old.DictType {
		if count, _ := dao.CountDictTypeByType(req.DictType); count > 0 {
			utils.Fail(c, 500, "字典类型已存在")
			return
		}
	}
	t := &model.DictType{
		ID: old.ID, DictName: req.DictName, DictType: req.DictType,
		Status: req.Status, Remark: req.Remark, CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if t.Status == "" {
		t.Status = "0"
	}
	if err := dao.UpdateDictType(t); err != nil {
		utils.Fail(c, 500, "修改字典类型失败")
		return
	}
	utils.Ok(c)
}

// DeleteDictType 删除字典类型（先校验是否有数据引用）
func DeleteDictType(c *gin.Context) {
	id := idParam(c)
	t, err := dao.GetDictTypeByID(id)
	if err != nil {
		utils.Fail(c, 500, "字典类型不存在")
		return
	}
	if count, _ := dao.CountDictDataByType(t.DictType); count > 0 {
		utils.Fail(c, 500, "该字典类型存在数据，不允许删除")
		return
	}
	if err := dao.DeleteDictType(id); err != nil {
		utils.Fail(c, 500, "删除字典类型失败")
		return
	}
	dao.ClearDictCache(t.DictType)
	utils.Ok(c)
}

// ---------- 字典数据 ----------

// ListDictData 分页查询字典数据
func ListDictData(c *gin.Context) {
	data, total, err := dao.ListDictData(pageNum(c), pageSize(c), c.Query("dictType"), c.Query("dictLabel"))
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, data, total)
}

// DictDataByType 按类型查询字典数据（下拉用，内存缓存 60s）
func DictDataByType(c *gin.Context) {
	dictType := c.Param("dictType")
	if cached, ok := dao.GetDictDataCache(dictType); ok {
		utils.OkData(c, cached)
		return
	}
	data, err := dao.ListDictDataByType(dictType)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	dao.SetDictDataCache(dictType, data)
	utils.OkData(c, data)
}

// GetDictData 查询字典数据详情
func GetDictData(c *gin.Context) {
	d, err := dao.GetDictDataByID(paramUint(c, "dictCode"))
	if err != nil {
		utils.Fail(c, 500, "字典数据不存在")
		return
	}
	utils.OkData(c, d)
}

// DictDataReq 字典数据请求体
type DictDataReq struct {
	DictCode  uint   `json:"dictCode"`
	DictSort  int    `json:"dictSort"`
	DictLabel string `json:"dictLabel"`
	DictValue string `json:"dictValue"`
	DictType  string `json:"dictType"`
	Status    string `json:"status"`
}

// CreateDictData 新增字典数据
func CreateDictData(c *gin.Context) {
	var req DictDataReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DictLabel == "" || req.DictType == "" {
		utils.Fail(c, 500, "字典标签和字典类型不能为空")
		return
	}
	d := &model.DictData{DictSort: req.DictSort, DictLabel: req.DictLabel, DictValue: req.DictValue, DictType: req.DictType, Status: req.Status}
	if d.Status == "" {
		d.Status = "0"
	}
	if err := dao.CreateDictData(d); err != nil {
		utils.Fail(c, 500, "新增字典数据失败")
		return
	}
	dao.ClearDictCache(req.DictType)
	utils.Ok(c)
}

// UpdateDictData 修改字典数据
func UpdateDictData(c *gin.Context) {
	var req DictDataReq
	if err := c.ShouldBindJSON(&req); err != nil || req.DictCode == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetDictDataByID(req.DictCode)
	if err != nil {
		utils.Fail(c, 500, "字典数据不存在")
		return
	}
	d := &model.DictData{
		ID: old.ID, DictSort: req.DictSort, DictLabel: req.DictLabel,
		DictValue: req.DictValue, DictType: req.DictType, Status: req.Status,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if d.Status == "" {
		d.Status = "0"
	}
	if err := dao.UpdateDictData(d); err != nil {
		utils.Fail(c, 500, "修改字典数据失败")
		return
	}
	// 类型被修改时，旧类型的下拉缓存也要失效
	if old.DictType != d.DictType {
		dao.ClearDictCache(old.DictType)
	}
	dao.ClearDictCache(d.DictType)
	utils.Ok(c)
}

// DeleteDictData 删除字典数据（支持逗号分隔批量）
func DeleteDictData(c *gin.Context) {
	ids := strings.Split(c.Param("dictCode"), ",")
	var idList []uint
	var affectedTypes []string
	for _, s := range ids {
		id := parseUint(s)
		if id == 0 {
			utils.Fail(c, 500, "参数错误")
			return
		}
		if d, err := dao.GetDictDataByID(id); err == nil {
			affectedTypes = append(affectedTypes, d.DictType)
		}
		idList = append(idList, id)
	}
	if err := dao.DeleteDictDatas(idList); err != nil {
		utils.Fail(c, 500, "删除字典数据失败")
		return
	}
	for _, t := range affectedTypes {
		dao.ClearDictCache(t)
	}
	utils.Ok(c)
}
