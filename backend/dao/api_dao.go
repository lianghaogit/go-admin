package dao

import (
	"strings"
	"sync"
	"time"

	"go-admin/model"
)

// ApiRule 内存接口规则
type ApiRule struct {
	Method   string
	Segments []string // path 按 / 分段，: 开头为通配
	Perms    string
}

var (
	apiRules   []ApiRule
	apiRulesMu sync.RWMutex
)

// ListApis 分页查询接口
func ListApis(q struct {
	PageNum, PageSize  int
	Name, Path, Status string
}) ([]model.Api, int64, error) {
	var apis []model.Api
	var total int64
	db := DB.Model(&model.Api{})
	if q.Name != "" {
		db = db.Where(likeExprEsc("api_name"), "%"+escapeLike(q.Name)+"%")
	}
	if q.Path != "" {
		db = db.Where(likeExprEsc("path"), "%"+escapeLike(q.Path)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("api_id asc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&apis).Error
	return apis, total, err
}

// GetApiByID 按 ID 查询接口
func GetApiByID(id uint) (*model.Api, error) {
	var api model.Api
	err := DB.First(&api, id).Error
	return &api, err
}

// CreateApi 新增接口
func CreateApi(a *model.Api) error {
	// 创建时间在写入时补齐（列表按创建时间倒序展示，NULL 会排到最后）
	if a.CreateTime.Time().IsZero() {
		a.CreateTime = model.LocalTime(time.Now())
	}
	if err := DB.Create(a).Error; err != nil {
		return err
	}
	return ReloadApiRules()
}

// UpdateApi 修改接口
func UpdateApi(a *model.Api) error {
	if err := DB.Save(a).Error; err != nil {
		return err
	}
	return ReloadApiRules()
}

// ChangeApiStatus 启停接口规则
func ChangeApiStatus(id uint, status string) error {
	if err := DB.Model(&model.Api{}).Where("api_id = ?", id).Update("status", status).Error; err != nil {
		return err
	}
	return ReloadApiRules()
}

// DeleteApi 删除接口
func DeleteApi(id uint) error {
	if err := DB.Delete(&model.Api{}, id).Error; err != nil {
		return err
	}
	return ReloadApiRules()
}

// ReloadApiRules 重新加载接口规则到内存（启用状态的）
func ReloadApiRules() error {
	var apis []model.Api
	if err := DB.Where("status = '0' AND path <> ''").Find(&apis).Error; err != nil {
		return err
	}
	rules := make([]ApiRule, 0, len(apis))
	for _, a := range apis {
		rules = append(rules, ApiRule{
			Method:   strings.ToUpper(a.Method),
			Segments: splitPath(a.Path),
			Perms:    a.Perms,
		})
	}
	apiRulesMu.Lock()
	apiRules = rules
	apiRulesMu.Unlock()
	return nil
}

// ApiRulesCount 当前内存规则数（调试用）
func ApiRulesCount() int {
	apiRulesMu.RLock()
	defer apiRulesMu.RUnlock()
	return len(apiRules)
}

// MatchApiRule 匹配接口规则（方法 + 路径逐段，:param 通配），未命中返回 nil
// 精确匹配优先于通配匹配，避免 /user/:id 吞掉 /user/list
func MatchApiRule(method, path string) *ApiRule {
	apiRulesMu.RLock()
	defer apiRulesMu.RUnlock()
	method = strings.ToUpper(method)
	segs := splitPath(path)
	// 第一轮：精确匹配（无 : 通配段且逐段相等）
	for i := range apiRules {
		r := &apiRules[i]
		if r.Method != method || len(r.Segments) != len(segs) {
			continue
		}
		exact := true
		for j := range r.Segments {
			if strings.HasPrefix(r.Segments[j], ":") || r.Segments[j] != segs[j] {
				exact = false
				break
			}
		}
		if exact {
			return r
		}
	}
	// 第二轮：通配匹配（:param 段任意）
	for i := range apiRules {
		r := &apiRules[i]
		if r.Method != method || len(r.Segments) != len(segs) {
			continue
		}
		match := true
		for j := range r.Segments {
			if strings.HasPrefix(r.Segments[j], ":") {
				continue
			}
			if r.Segments[j] != segs[j] {
				match = false
				break
			}
		}
		if match {
			return r
		}
	}
	return nil
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// GetApiMeta 按路径+方法查询接口名称与权限标识（接口文档用）
func GetApiMeta(path, method string) (name, perms string) {
	var api model.Api
	err := DB.Where("path = ? AND method = ?", path, strings.ToUpper(method)).First(&api).Error
	if err != nil {
		return "", ""
	}
	return api.APIName, api.Perms
}

// CountApiByPathMethod 统计相同路径+方法的接口数
func CountApiByPathMethod(path, method string, excludeID uint) (int64, error) {
	var count int64
	db := DB.Model(&model.Api{}).Where("path = ? AND method = ?", path, strings.ToUpper(method))
	if excludeID > 0 {
		// 主键列为 api_id（字段 APIID），写成 id 会导致 SQL 报错且判重失效
		db = db.Where("api_id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
