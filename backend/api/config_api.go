package api

import (
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListConfigs 分页查询系统参数
func ListConfigs(c *gin.Context) {
	q := dao.ConfigQuery{
		PageNum:    pageNum(c),
		PageSize:   pageSize(c),
		ConfigName: c.Query("configName"),
		ConfigKey:  c.Query("configKey"),
		ConfigType: c.Query("configType"),
	}
	configs, total, err := dao.ListConfigs(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, configs, total)
}

// GetConfig 查询参数详情
func GetConfig(c *gin.Context) {
	cfg, err := dao.GetConfigByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "参数不存在")
		return
	}
	utils.OkData(c, cfg)
}

// ConfigReq 参数请求体
type ConfigReq struct {
	ConfigID    uint   `json:"configId"`
	ConfigName  string `json:"configName"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType  string `json:"configType"`
	Remark      string `json:"remark"`
}

// CreateConfig 新增参数
func CreateConfig(c *gin.Context) {
	var req ConfigReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ConfigName == "" || req.ConfigKey == "" {
		utils.Fail(c, 500, "参数名称和参数键名不能为空")
		return
	}
	if count, _ := dao.CountConfigByKey(req.ConfigKey); count > 0 {
		utils.Fail(c, 500, "参数键名已存在")
		return
	}
	cfg := &model.Config{ConfigName: req.ConfigName, ConfigKey: req.ConfigKey, ConfigValue: req.ConfigValue, ConfigType: req.ConfigType, Remark: req.Remark}
	if cfg.ConfigType == "" {
		cfg.ConfigType = "N"
	}
	if err := dao.CreateConfig(cfg); err != nil {
		utils.Fail(c, 500, "新增参数失败")
		return
	}
	utils.Ok(c)
}

// UpdateConfig 修改参数
func UpdateConfig(c *gin.Context) {
	var req ConfigReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ConfigID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetConfigByID(req.ConfigID)
	if err != nil {
		utils.Fail(c, 500, "参数不存在")
		return
	}
	// 内置参数不允许修改类型
	if old.ConfigType == "Y" && req.ConfigType != "Y" {
		utils.Fail(c, 500, "内置参数不允许修改类型")
		return
	}
	if req.ConfigKey != old.ConfigKey {
		if count, _ := dao.CountConfigByKey(req.ConfigKey); count > 0 {
			utils.Fail(c, 500, "参数键名已存在")
			return
		}
	}
	cfg := &model.Config{
		ID: old.ID, ConfigName: req.ConfigName, ConfigKey: req.ConfigKey,
		ConfigValue: req.ConfigValue, ConfigType: old.ConfigType, Remark: req.Remark,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if err := dao.UpdateConfig(cfg); err != nil {
		utils.Fail(c, 500, "修改参数失败")
		return
	}
	utils.Ok(c)
}

// DeleteConfig 删除参数（内置参数不允许删除）
func DeleteConfig(c *gin.Context) {
	id := idParam(c)
	cfg, err := dao.GetConfigByID(id)
	if err != nil {
		utils.Fail(c, 500, "参数不存在")
		return
	}
	if cfg.ConfigType == "Y" {
		utils.Fail(c, 500, "内置参数不允许删除")
		return
	}
	if err := dao.DeleteConfig(id); err != nil {
		utils.Fail(c, 500, "删除参数失败")
		return
	}
	utils.Ok(c)
}
