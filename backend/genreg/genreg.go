// Package genreg 代码生成模块的运行时注册中心。
//
// 生成的 api 文件在 init() 里调用 RegisterModule，把自己声明的路由、菜单登记进来；
// router 启动时把路由挂到鉴权分组，seed 启动时把菜单（含按钮权限）写入数据库。
// 这样生成代码只要放进 backend/model、backend/dao、backend/api 与前端 views 目录，
// 重启服务即可访问，无需手工修改 router.go / seed.go。
package genreg

import "github.com/gin-gonic/gin"

// Route 一条生成接口的路由声明
type Route struct {
	Method  string // GET/POST/PUT/DELETE
	Path    string // 相对模块分组，如 "/list"、"/:id"、""（空表示分组根路径）
	Perm    string // 权限标识，如 system:book:list
	Handler gin.HandlerFunc
}

// MenuButton 按钮级权限菜单（F 类型）
type MenuButton struct {
	Name string
	Perm string
}

// Menu 生成的页面菜单（C 类型）
type Menu struct {
	Name      string // 菜单名称（功能名）
	ParentID  uint   // 上级菜单 ID，0 表示按默认目录（系统工具 1700）挂载
	Path      string // 路由 path，如 book
	Component string // 组件路径，如 system/book/index
	Perm      string // 列表权限，如 system:book:list
	Icon      string
	OrderNum  int
	Buttons   []MenuButton
}

// Module 一个生成模块（一张表/一组主子表）的完整声明
type Module struct {
	Module       string // 模块名（接口前缀，如 system）
	ClassVar     string // 小驼峰（如 book）
	ClassName    string // 结构体名（如 Book）
	FunctionName string // 功能名称（如 图书）
	Routes       []Route
	Menu         Menu
}

var modules []Module

// RegisterModule 登记一个生成模块（由生成的 api 文件 init() 调用）
func RegisterModule(m Module) {
	// 防止同模块重复注册（例如手工复制出两份代码）
	for _, old := range modules {
		if old.Module == m.Module && old.ClassVar == m.ClassVar {
			return
		}
	}
	modules = append(modules, m)
}

// Modules 返回全部已注册的生成模块（按注册顺序）
func Modules() []Module {
	return modules
}
