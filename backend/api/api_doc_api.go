package api

import (
	"sort"
	"strings"

	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ApiDocItem 接口文档条目
type ApiDocItem struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	Name     string `json:"name"`
	Perms    string `json:"perms"`
	NeedAuth bool   `json:"needAuth"`
}

// ApiDocModule 按模块分组的接口文档
type ApiDocModule struct {
	Name  string       `json:"name"`
	Items []ApiDocItem `json:"items"`
}

// ApiDocList 系统接口文档：枚举全部已注册 /api 路由，附接口权限标识
func ApiDocList(c *gin.Context, routes gin.RoutesInfo) {
	doc := make([]ApiDocItem, 0, len(routes))
	for _, r := range routes {
		if !strings.HasPrefix(r.Path, "/api") {
			continue
		}
		// 仅收录需鉴权的接口（排除验证码/登录/文档自身）
		if r.Path == "/api/captchaImage" || r.Path == "/api/login" || r.Path == "/api/tool/api-doc/list" {
			continue
		}
		name, perms := dao.GetApiMeta(r.Path, r.Method)
		doc = append(doc, ApiDocItem{
			Method:   r.Method,
			Path:     r.Path,
			Name:     name,
			Perms:    perms,
			NeedAuth: true,
		})
	}

	// 按模块前缀分组：system / monitor / tool / common / 其他
	modules := make([]ApiDocModule, 0)
	group := func(prefix, name string) *ApiDocModule {
		for i := range modules {
			if modules[i].Name == name {
				return &modules[i]
			}
		}
		modules = append(modules, ApiDocModule{Name: name})
		return &modules[len(modules)-1]
	}

	rest := make([]ApiDocItem, 0)
	for _, item := range doc {
		switch {
		case strings.HasPrefix(item.Path, "/api/system/"):
			group(item.Path, "system").Items = append(group(item.Path, "system").Items, item)
		case strings.HasPrefix(item.Path, "/api/monitor/"):
			group(item.Path, "monitor").Items = append(group(item.Path, "monitor").Items, item)
		case strings.HasPrefix(item.Path, "/api/tool/"):
			group(item.Path, "tool").Items = append(group(item.Path, "tool").Items, item)
		case strings.HasPrefix(item.Path, "/api/common/"):
			group(item.Path, "common").Items = append(group(item.Path, "common").Items, item)
		default:
			rest = append(rest, item)
		}
	}
	if len(rest) > 0 {
		group("", "base").Items = rest
	}

	for i := range modules {
		sort.Slice(modules[i].Items, func(a, b int) bool {
			if modules[i].Items[a].Path == modules[i].Items[b].Path {
				return modules[i].Items[a].Method < modules[i].Items[b].Method
			}
			return modules[i].Items[a].Path < modules[i].Items[b].Path
		})
	}

	utils.OkData(c, gin.H{"modules": modules, "total": len(doc)})
}
