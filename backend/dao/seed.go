package dao

import (
	"go-admin/genreg"
	"go-admin/model"
	"go-admin/utils"
)

// seed 初始化数据：部门/字典每次启动确保存在，角色/菜单/管理员仅首次写入
func seed() error {
	if err := ensureDepts(); err != nil {
		return err
	}

	var count int64
	if err := DB.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		if err := seedBase(); err != nil {
			return err
		}
	}

	if err := ensureMenus(); err != nil {
		return err
	}
	if err := ensureDicts(); err != nil {
		return err
	}
	if err := ensurePosts(); err != nil {
		return err
	}
	if err := ensureConfigs(); err != nil {
		return err
	}
	if err := ensureNotices(); err != nil {
		return err
	}
	if err := ensureJobs(); err != nil {
		return err
	}
	if err := ensureApis(); err != nil {
		return err
	}
	if err := ensureFlows(); err != nil {
		return err
	}
	return nil
}

// ensureDepts 部门树不存在时写入（幂等），并回填 admin 部门归属
func ensureDepts() error {
	var count int64
	if err := DB.Model(&model.Dept{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		depts := []model.Dept{
			{ID: 100, ParentID: 0, DeptName: "总公司", OrderNum: 1, Leader: "管理员", Phone: "010-00000000", Email: "boss@example.com", Status: "0"},
			{ID: 101, ParentID: 100, DeptName: "研发部", OrderNum: 1, Leader: "研发主管", Phone: "010-00000001", Email: "rd@example.com", Status: "0"},
			{ID: 102, ParentID: 100, DeptName: "测试部", OrderNum: 2, Leader: "测试主管", Phone: "010-00000002", Email: "qa@example.com", Status: "0"},
			{ID: 103, ParentID: 100, DeptName: "产品部", OrderNum: 3, Leader: "产品主管", Phone: "010-00000003", Email: "pm@example.com", Status: "0"},
		}
		for i := range depts {
			if err := DB.Create(&depts[i]).Error; err != nil {
				return err
			}
		}
	}
	// 回填：老库升级后 admin 的部门归属（幂等，仅当值为 0 时）
	if err := DB.Model(&model.User{}).Where("username = ? AND dept_id = 0", "admin").
		Update("dept_id", 100).Error; err != nil {
		return err
	}
	return nil
}

// seedBase 首次启动写入角色、菜单、管理员账号（默认 admin / admin123）
func seedBase() error {
	adminRole := model.Role{RoleName: "超级管理员", RoleKey: "admin", Sort: 1, Status: "0", Remark: "系统内置超级管理员，拥有全部权限"}
	commonRole := model.Role{RoleName: "普通角色", RoleKey: "common", Sort: 2, Status: "0", Remark: "普通测试角色"}
	if err := DB.Create(&adminRole).Error; err != nil {
		return err
	}
	if err := DB.Create(&commonRole).Error; err != nil {
		return err
	}

	// 管理员用户
	hash, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}
	admin := model.User{
		DeptID: 100, Username: "admin", Password: hash, Nickname: "管理员",
		Email: "admin@example.com", Phone: "13800000000", Status: "0", Remark: "系统内置管理员",
	}
	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	// 绑定：admin -> 超级管理员角色；超级管理员角色 -> 全部菜单
	if err := DB.Model(&admin).Association("Roles").Append(&adminRole); err != nil {
		return err
	}
	var allMenus []model.Menu
	if err := DB.Find(&allMenus).Error; err != nil {
		return err
	}
	if err := DB.Model(&adminRole).Association("Menus").Append(allMenus); err != nil {
		return err
	}
	return nil
}

// ensureMenus 补齐缺失的系统菜单（按固定 ID 幂等插入），并将新菜单绑定给超级管理员角色
func ensureMenus() error {
	menus := []model.Menu{
		{ID: 1, MenuName: "系统管理", ParentID: 0, OrderNum: 1, Path: "system", MenuType: "M", Icon: "Setting", Visible: "0", Status: "0"},
		{ID: 100, MenuName: "用户管理", ParentID: 1, OrderNum: 1, Path: "user", Component: "system/user/index", MenuType: "C", Perms: "system:user:list", Icon: "User", Visible: "0", Status: "0"},
		{ID: 101, MenuName: "用户新增", ParentID: 100, OrderNum: 1, MenuType: "F", Perms: "system:user:add", Visible: "0", Status: "0"},
		{ID: 102, MenuName: "用户修改", ParentID: 100, OrderNum: 2, MenuType: "F", Perms: "system:user:edit", Visible: "0", Status: "0"},
		{ID: 103, MenuName: "用户删除", ParentID: 100, OrderNum: 3, MenuType: "F", Perms: "system:user:remove", Visible: "0", Status: "0"},
		{ID: 104, MenuName: "重置密码", ParentID: 100, OrderNum: 4, MenuType: "F", Perms: "system:user:resetPwd", Visible: "0", Status: "0"},
		{ID: 105, MenuName: "用户导出", ParentID: 100, OrderNum: 5, MenuType: "F", Perms: "system:user:export", Visible: "0", Status: "0"},
		{ID: 106, MenuName: "用户导入", ParentID: 100, OrderNum: 6, MenuType: "F", Perms: "system:user:import", Visible: "0", Status: "0"},
		{ID: 200, MenuName: "角色管理", ParentID: 1, OrderNum: 2, Path: "role", Component: "system/role/index", MenuType: "C", Perms: "system:role:list", Icon: "Avatar", Visible: "0", Status: "0"},
		{ID: 201, MenuName: "角色新增", ParentID: 200, OrderNum: 1, MenuType: "F", Perms: "system:role:add", Visible: "0", Status: "0"},
		{ID: 202, MenuName: "角色修改", ParentID: 200, OrderNum: 2, MenuType: "F", Perms: "system:role:edit", Visible: "0", Status: "0"},
		{ID: 203, MenuName: "角色删除", ParentID: 200, OrderNum: 3, MenuType: "F", Perms: "system:role:remove", Visible: "0", Status: "0"},
		{ID: 300, MenuName: "菜单管理", ParentID: 1, OrderNum: 3, Path: "menu", Component: "system/menu/index", MenuType: "C", Perms: "system:menu:list", Icon: "Menu", Visible: "0", Status: "0"},
		{ID: 301, MenuName: "菜单新增", ParentID: 300, OrderNum: 1, MenuType: "F", Perms: "system:menu:add", Visible: "0", Status: "0"},
		{ID: 302, MenuName: "菜单修改", ParentID: 300, OrderNum: 2, MenuType: "F", Perms: "system:menu:edit", Visible: "0", Status: "0"},
		{ID: 303, MenuName: "菜单删除", ParentID: 300, OrderNum: 3, MenuType: "F", Perms: "system:menu:remove", Visible: "0", Status: "0"},
		{ID: 400, MenuName: "登录日志", ParentID: 1990, OrderNum: 1, Path: "loginlog", Component: "system/loginlog/index", MenuType: "C", Perms: "monitor:logininfor:list", Icon: "Document", Visible: "0", Status: "0"},
		{ID: 500, MenuName: "部门管理", ParentID: 1, OrderNum: 4, Path: "dept", Component: "system/dept/index", MenuType: "C", Perms: "system:dept:list", Icon: "OfficeBuilding", Visible: "0", Status: "0"},
		{ID: 501, MenuName: "部门新增", ParentID: 500, OrderNum: 1, MenuType: "F", Perms: "system:dept:add", Visible: "0", Status: "0"},
		{ID: 502, MenuName: "部门修改", ParentID: 500, OrderNum: 2, MenuType: "F", Perms: "system:dept:edit", Visible: "0", Status: "0"},
		{ID: 503, MenuName: "部门删除", ParentID: 500, OrderNum: 3, MenuType: "F", Perms: "system:dept:remove", Visible: "0", Status: "0"},
		{ID: 600, MenuName: "字典管理", ParentID: 1, OrderNum: 6, Path: "dict", Component: "system/dict/index", MenuType: "C", Perms: "system:dict:list", Icon: "Notebook", Visible: "0", Status: "0"},
		{ID: 601, MenuName: "字典新增", ParentID: 600, OrderNum: 1, MenuType: "F", Perms: "system:dict:add", Visible: "0", Status: "0"},
		{ID: 602, MenuName: "字典修改", ParentID: 600, OrderNum: 2, MenuType: "F", Perms: "system:dict:edit", Visible: "0", Status: "0"},
		{ID: 603, MenuName: "字典删除", ParentID: 600, OrderNum: 3, MenuType: "F", Perms: "system:dict:remove", Visible: "0", Status: "0"},
		{ID: 700, MenuName: "操作日志", ParentID: 1990, OrderNum: 2, Path: "operlog", Component: "system/operlog/index", MenuType: "C", Perms: "monitor:operlog:list", Icon: "Operation", Visible: "0", Status: "0"},
		{ID: 701, MenuName: "日志删除", ParentID: 700, OrderNum: 1, MenuType: "F", Perms: "monitor:operlog:remove", Visible: "0", Status: "0"},
		{ID: 800, MenuName: "岗位管理", ParentID: 1, OrderNum: 5, Path: "post", Component: "system/post/index", MenuType: "C", Perms: "system:post:list", Icon: "Suitcase", Visible: "0", Status: "0"},
		{ID: 801, MenuName: "岗位新增", ParentID: 800, OrderNum: 1, MenuType: "F", Perms: "system:post:add", Visible: "0", Status: "0"},
		{ID: 802, MenuName: "岗位修改", ParentID: 800, OrderNum: 2, MenuType: "F", Perms: "system:post:edit", Visible: "0", Status: "0"},
		{ID: 803, MenuName: "岗位删除", ParentID: 800, OrderNum: 3, MenuType: "F", Perms: "system:post:remove", Visible: "0", Status: "0"},
		{ID: 900, MenuName: "参数配置", ParentID: 1, OrderNum: 7, Path: "config", Component: "system/config/index", MenuType: "C", Perms: "system:config:list", Icon: "Tools", Visible: "0", Status: "0"},
		{ID: 901, MenuName: "参数新增", ParentID: 900, OrderNum: 1, MenuType: "F", Perms: "system:config:add", Visible: "0", Status: "0"},
		{ID: 902, MenuName: "参数修改", ParentID: 900, OrderNum: 2, MenuType: "F", Perms: "system:config:edit", Visible: "0", Status: "0"},
		{ID: 903, MenuName: "参数删除", ParentID: 900, OrderNum: 3, MenuType: "F", Perms: "system:config:remove", Visible: "0", Status: "0"},
		{ID: 1000, MenuName: "通知公告", ParentID: 1, OrderNum: 8, Path: "notice", Component: "system/notice/index", MenuType: "C", Perms: "system:notice:list", Icon: "Bell", Visible: "0", Status: "0"},
		{ID: 1001, MenuName: "公告新增", ParentID: 1000, OrderNum: 1, MenuType: "F", Perms: "system:notice:add", Visible: "0", Status: "0"},
		{ID: 1002, MenuName: "公告修改", ParentID: 1000, OrderNum: 2, MenuType: "F", Perms: "system:notice:edit", Visible: "0", Status: "0"},
		{ID: 1003, MenuName: "公告删除", ParentID: 1000, OrderNum: 3, MenuType: "F", Perms: "system:notice:remove", Visible: "0", Status: "0"},
		{ID: 1100, MenuName: "在线用户", ParentID: 1991, OrderNum: 1, Path: "online", Component: "monitor/online/index", MenuType: "C", Perms: "monitor:online:list", Icon: "Monitor", Visible: "0", Status: "0"},
		{ID: 1101, MenuName: "在线用户强退", ParentID: 1100, OrderNum: 1, MenuType: "F", Perms: "monitor:online:forceLogout", Visible: "0", Status: "0"},
		{ID: 1200, MenuName: "定时任务", ParentID: 1991, OrderNum: 2, Path: "job", Component: "monitor/job/index", MenuType: "C", Perms: "monitor:job:list", Icon: "Timer", Visible: "0", Status: "0"},
		{ID: 1201, MenuName: "任务新增", ParentID: 1200, OrderNum: 1, MenuType: "F", Perms: "monitor:job:add", Visible: "0", Status: "0"},
		{ID: 1202, MenuName: "任务修改", ParentID: 1200, OrderNum: 2, MenuType: "F", Perms: "monitor:job:edit", Visible: "0", Status: "0"},
		{ID: 1203, MenuName: "任务删除", ParentID: 1200, OrderNum: 3, MenuType: "F", Perms: "monitor:job:remove", Visible: "0", Status: "0"},
		{ID: 1204, MenuName: "任务状态修改", ParentID: 1200, OrderNum: 4, MenuType: "F", Perms: "monitor:job:changeStatus", Visible: "0", Status: "0"},
		{ID: 1205, MenuName: "任务立即执行", ParentID: 1200, OrderNum: 5, MenuType: "F", Perms: "monitor:job:run", Visible: "0", Status: "0"},
		{ID: 1300, MenuName: "服务监控", ParentID: 1991, OrderNum: 3, Path: "server", Component: "monitor/server/index", MenuType: "C", Perms: "monitor:server:list", Icon: "DataLine", Visible: "0", Status: "0"},
		{ID: 1400, MenuName: "缓存监控", ParentID: 1991, OrderNum: 4, Path: "cache", Component: "monitor/cache/index", MenuType: "C", Perms: "monitor:cache:list", Icon: "Coin", Visible: "0", Status: "0"},
		{ID: 1401, MenuName: "缓存清空", ParentID: 1400, OrderNum: 1, MenuType: "F", Perms: "monitor:cache:clear", Visible: "0", Status: "0"},
		{ID: 1500, MenuName: "数据监控", ParentID: 1991, OrderNum: 5, Path: "db", Component: "monitor/db/index", MenuType: "C", Perms: "monitor:db:list", Icon: "Odometer", Visible: "0", Status: "0"},
		{ID: 1600, MenuName: "接口权限", ParentID: 1, OrderNum: 9, Path: "api", Component: "system/api/index", MenuType: "C", Perms: "system:api:list", Icon: "Key", Visible: "0", Status: "0"},
		{ID: 1601, MenuName: "接口新增", ParentID: 1600, OrderNum: 1, MenuType: "F", Perms: "system:api:add", Visible: "0", Status: "0"},
		{ID: 1602, MenuName: "接口修改", ParentID: 1600, OrderNum: 2, MenuType: "F", Perms: "system:api:edit", Visible: "0", Status: "0"},
		{ID: 1603, MenuName: "接口删除", ParentID: 1600, OrderNum: 3, MenuType: "F", Perms: "system:api:remove", Visible: "0", Status: "0"},
		{ID: 1604, MenuName: "接口状态修改", ParentID: 1600, OrderNum: 4, MenuType: "F", Perms: "system:api:changeStatus", Visible: "0", Status: "0"},
		{ID: 1700, MenuName: "系统工具", ParentID: 0, OrderNum: 3, Path: "tool", MenuType: "M", Icon: "Operation", Visible: "0", Status: "0"},
		{ID: 1701, MenuName: "代码生成", ParentID: 1700, OrderNum: 1, Path: "gen", Component: "tool/gen/index", MenuType: "C", Perms: "tool:gen:list", Icon: "MagicStick", Visible: "0", Status: "0"},
		{ID: 1702, MenuName: "代码生成-导入", ParentID: 1701, OrderNum: 1, MenuType: "F", Perms: "tool:gen:import", Visible: "0", Status: "0"},
		{ID: 1703, MenuName: "代码生成-修改", ParentID: 1701, OrderNum: 2, MenuType: "F", Perms: "tool:gen:edit", Visible: "0", Status: "0"},
		{ID: 1704, MenuName: "代码生成-删除", ParentID: 1701, OrderNum: 3, MenuType: "F", Perms: "tool:gen:remove", Visible: "0", Status: "0"},
		{ID: 1705, MenuName: "代码生成-预览", ParentID: 1701, OrderNum: 4, MenuType: "F", Perms: "tool:gen:preview", Visible: "0", Status: "0"},
		{ID: 1706, MenuName: "代码生成-生成", ParentID: 1701, OrderNum: 5, MenuType: "F", Perms: "tool:gen:code", Visible: "0", Status: "0"},
		{ID: 1707, MenuName: "接口文档", ParentID: 1700, OrderNum: 2, Path: "api-doc", Component: "tool/api-doc/index", MenuType: "C", Perms: "tool:apidoc:list", Icon: "Link", Visible: "0", Status: "0"},
		{ID: 1708, MenuName: "文件上传", ParentID: 1700, OrderNum: 3, Path: "upload", Component: "tool/upload/index", MenuType: "C", Perms: "tool:upload:list", Icon: "Upload", Visible: "0", Status: "0"},
		{ID: 1709, MenuName: "通用上传", ParentID: 1708, OrderNum: 1, MenuType: "F", Perms: "common:upload", Visible: "0", Status: "0"},
		// 日志管理（系统管理下）：登录日志 / 操作日志
		{ID: 1990, MenuName: "日志管理", ParentID: 1, OrderNum: 10, Path: "log", MenuType: "M", Icon: "Document", Visible: "0", Status: "0"},
		// 系统监控（顶级目录）：在线用户 / 定时任务 / 服务监控 / 缓存监控 / 数据监控
		{ID: 1991, MenuName: "系统监控", ParentID: 0, OrderNum: 2, Path: "monitor", MenuType: "M", Icon: "Monitor", Visible: "0", Status: "0"},
		// 流程中心（顶级目录）：流程定义 / 发起申请 / 待办审批 / 已办任务 / 抄送我的
		{ID: 2000, MenuName: "流程中心", ParentID: 0, OrderNum: 4, Path: "flow", MenuType: "M", Icon: "Share", Visible: "0", Status: "0"},
		{ID: 2001, MenuName: "流程定义", ParentID: 2000, OrderNum: 1, Path: "def", Component: "flow/def/index", MenuType: "C", Perms: "flow:def:list", Icon: "SetUp", Visible: "0", Status: "0"},
		{ID: 2002, MenuName: "流程新增", ParentID: 2001, OrderNum: 1, MenuType: "F", Perms: "flow:def:add", Visible: "0", Status: "0"},
		{ID: 2003, MenuName: "流程修改", ParentID: 2001, OrderNum: 2, MenuType: "F", Perms: "flow:def:edit", Visible: "0", Status: "0"},
		{ID: 2004, MenuName: "流程删除", ParentID: 2001, OrderNum: 3, MenuType: "F", Perms: "flow:def:remove", Visible: "0", Status: "0"},
		{ID: 2005, MenuName: "流程启停", ParentID: 2001, OrderNum: 4, MenuType: "F", Perms: "flow:def:changeStatus", Visible: "0", Status: "0"},
		{ID: 2006, MenuName: "流程详情", ParentID: 2001, OrderNum: 5, MenuType: "F", Perms: "flow:def:query", Visible: "0", Status: "0"},
		{ID: 2010, MenuName: "发起申请", ParentID: 2000, OrderNum: 2, Path: "apply", Component: "flow/apply/index", MenuType: "C", Perms: "flow:apply:list", Icon: "EditPen", Visible: "0", Status: "0"},
		{ID: 2011, MenuName: "发起申请", ParentID: 2010, OrderNum: 1, MenuType: "F", Perms: "flow:apply:add", Visible: "0", Status: "0"},
		{ID: 2012, MenuName: "撤回申请", ParentID: 2010, OrderNum: 2, MenuType: "F", Perms: "flow:apply:withdraw", Visible: "0", Status: "0"},
		{ID: 2013, MenuName: "重新提交", ParentID: 2010, OrderNum: 3, MenuType: "F", Perms: "flow:apply:reapply", Visible: "0", Status: "0"},
		{ID: 2014, MenuName: "催办", ParentID: 2010, OrderNum: 4, MenuType: "F", Perms: "flow:apply:urge", Visible: "0", Status: "0"},
		{ID: 2015, MenuName: "全部实例", ParentID: 2010, OrderNum: 5, MenuType: "F", Perms: "flow:instance:list", Visible: "0", Status: "0"},
		{ID: 2016, MenuName: "实例详情", ParentID: 2010, OrderNum: 6, MenuType: "F", Perms: "flow:instance:query", Visible: "0", Status: "0"},
		{ID: 2017, MenuName: "终止申请", ParentID: 2010, OrderNum: 7, MenuType: "F", Perms: "flow:apply:terminate", Visible: "0", Status: "0"},
		{ID: 2020, MenuName: "待办审批", ParentID: 2000, OrderNum: 3, Path: "todo", Component: "flow/todo/index", MenuType: "C", Perms: "flow:todo:list", Icon: "Stamp", Visible: "0", Status: "0"},
		{ID: 2021, MenuName: "审批通过", ParentID: 2020, OrderNum: 1, MenuType: "F", Perms: "flow:todo:approve", Visible: "0", Status: "0"},
		{ID: 2022, MenuName: "审批驳回", ParentID: 2020, OrderNum: 2, MenuType: "F", Perms: "flow:todo:reject", Visible: "0", Status: "0"},
		{ID: 2023, MenuName: "转办", ParentID: 2020, OrderNum: 3, MenuType: "F", Perms: "flow:todo:transfer", Visible: "0", Status: "0"},
		{ID: 2024, MenuName: "加签", ParentID: 2020, OrderNum: 4, MenuType: "F", Perms: "flow:todo:addSign", Visible: "0", Status: "0"},
		{ID: 2030, MenuName: "已办任务", ParentID: 2000, OrderNum: 4, Path: "done", Component: "flow/done/index", MenuType: "C", Perms: "flow:todo:list", Icon: "Finished", Visible: "0", Status: "0"},
		{ID: 2040, MenuName: "抄送我的", ParentID: 2000, OrderNum: 5, Path: "cc", Component: "flow/cc/index", MenuType: "C", Perms: "flow:cc:list", Icon: "Notification", Visible: "0", Status: "0"},
	}

	created := make([]model.Menu, 0)
	for i := range menus {
		var m model.Menu
		err := DB.Where("id = ?", menus[i].ID).First(&m).Error
		if err == nil {
			continue // 已存在，不做覆盖，保留用户修改
		}
		if err := DB.Create(&menus[i]).Error; err != nil {
			return err
		}
		created = append(created, menus[i])
	}
	// 将新菜单绑定给超级管理员角色（id=1）
	for i := range created {
		var cnt int64
		if err := DB.Table("sys_role_menu").Where("role_id = ? AND menu_id = ?", 1, created[i].ID).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			if err := DB.Exec("INSERT INTO sys_role_menu (role_id, menu_id) VALUES (?, ?)", 1, created[i].ID).Error; err != nil {
				return err
			}
		}
	}
	return ensureGenMenus()
}

// ensureGenMenus 为代码生成模块登记的菜单自动建库（幂等）。
// 生成的代码放进目录并重启后，菜单与按钮权限即自动出现，无需手工往 seed.go 里加。
func ensureGenMenus() error {
	for _, m := range genreg.Modules() {
		if m.Menu.Perm == "" || m.Menu.Path == "" {
			continue
		}
		// 生成模块默认挂到“系统工具”(1700) 下，避免全部堆在顶级目录；
		// 生成代码可在 Menu.ParentID 中显式指定其他上级。
		parentID := m.Menu.ParentID
		if parentID == 0 {
			parentID = 1700
		}
		var existing model.Menu
		err := DB.Where("perms = ? AND menu_type = ?", m.Menu.Perm, "C").First(&existing).Error
		if err == nil {
			// 历史生成菜单曾挂在顶级，原地迁移；用户手工改到其他目录的不动
			if existing.ParentID == 0 && parentID != 0 {
				if err := DB.Model(&existing).Update("parent_id", parentID).Error; err != nil {
					return err
				}
			}
			if err := bindRoleMenu(1, existing.ID); err != nil {
				return err
			}
			if err := ensureGenButtons(existing.ID, m.Menu.Buttons); err != nil {
				return err
			}
			continue
		}
		name := m.Menu.Name
		if name == "" {
			name = m.FunctionName
		}
		icon := m.Menu.Icon
		if icon == "" {
			icon = "Grid"
		}
		orderNum := m.Menu.OrderNum
		if orderNum == 0 {
			orderNum = 20
		}
		parent := model.Menu{
			MenuName:  name,
			ParentID:  parentID,
			OrderNum:  orderNum,
			Path:      m.Menu.Path,
			Component: m.Menu.Component,
			MenuType:  "C",
			Perms:     m.Menu.Perm,
			Icon:      icon,
			Visible:   "0",
			Status:    "0",
		}
		if err := DB.Create(&parent).Error; err != nil {
			return err
		}
		if err := bindRoleMenu(1, parent.ID); err != nil {
			return err
		}
		if err := ensureGenButtons(parent.ID, m.Menu.Buttons); err != nil {
			return err
		}
	}
	return nil
}

// ensureGenButtons 为生成模块补齐按钮权限（幂等，已存在的权限标识不重复创建）
func ensureGenButtons(parentID uint, buttons []genreg.MenuButton) error {
	for i, btn := range buttons {
		if btn.Perm == "" {
			continue
		}
		var count int64
		if err := DB.Model(&model.Menu{}).Where("perms = ? AND menu_type = ?", btn.Perm, "F").Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		child := model.Menu{
			MenuName: btn.Name,
			ParentID: parentID,
			OrderNum: i + 1,
			MenuType: "F",
			Perms:    btn.Perm,
			Visible:  "0",
			Status:   "0",
		}
		if err := DB.Create(&child).Error; err != nil {
			return err
		}
		if err := bindRoleMenu(1, child.ID); err != nil {
			return err
		}
	}
	return nil
}

// bindRoleMenu 幂等绑定角色与菜单
func bindRoleMenu(roleID, menuID uint) error {
	var cnt int64
	if err := DB.Table("sys_role_menu").Where("role_id = ? AND menu_id = ?", roleID, menuID).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}
	return DB.Exec("INSERT INTO sys_role_menu (role_id, menu_id) VALUES (?, ?)", roleID, menuID).Error
}

// ensureDicts 字典类型与数据不存在时写入（幂等）
func ensureDicts() error {
	var count int64
	if err := DB.Model(&model.DictType{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	types := []model.DictType{
		{ID: 1, DictName: "用户状态", DictType: "sys_user_status", Status: "0", Remark: "用户状态列表"},
		{ID: 2, DictName: "系统开关", DictType: "sys_normal_disable", Status: "0", Remark: "系统开关列表"},
		{ID: 3, DictName: "显示状态", DictType: "sys_show_hide", Status: "0", Remark: "显示/隐藏列表"},
		{ID: 4, DictName: "是否状态", DictType: "sys_yes_no", Status: "0", Remark: "是/否列表"},
	}
	for i := range types {
		if err := DB.Create(&types[i]).Error; err != nil {
			return err
		}
	}
	data := []model.DictData{
		{ID: 1, DictSort: 1, DictLabel: "正常", DictValue: "0", DictType: "sys_user_status", Status: "0"},
		{ID: 2, DictSort: 2, DictLabel: "停用", DictValue: "1", DictType: "sys_user_status", Status: "0"},
		{ID: 3, DictSort: 1, DictLabel: "正常", DictValue: "0", DictType: "sys_normal_disable", Status: "0"},
		{ID: 4, DictSort: 2, DictLabel: "停用", DictValue: "1", DictType: "sys_normal_disable", Status: "0"},
		{ID: 5, DictSort: 1, DictLabel: "显示", DictValue: "0", DictType: "sys_show_hide", Status: "0"},
		{ID: 6, DictSort: 2, DictLabel: "隐藏", DictValue: "1", DictType: "sys_show_hide", Status: "0"},
		{ID: 7, DictSort: 1, DictLabel: "是", DictValue: "Y", DictType: "sys_yes_no", Status: "0"},
		{ID: 8, DictSort: 2, DictLabel: "否", DictValue: "N", DictType: "sys_yes_no", Status: "0"},
	}
	for i := range data {
		if err := DB.Create(&data[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensurePosts 岗位不存在时写入，并绑定 admin 默认岗位（幂等）
func ensurePosts() error {
	var count int64
	if err := DB.Model(&model.Post{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		posts := []model.Post{
			{ID: 1, PostCode: "ceo", PostName: "董事长", PostSort: 1, Status: "0", Remark: "公司最高决策者"},
			{ID: 2, PostCode: "se", PostName: "项目经理", PostSort: 2, Status: "0", Remark: "项目负责人"},
			{ID: 3, PostCode: "user", PostName: "员工", PostSort: 3, Status: "0", Remark: "普通员工"},
		}
		for i := range posts {
			if err := DB.Create(&posts[i]).Error; err != nil {
				return err
			}
		}
	}
	// 回填：admin 默认绑定董事长岗位（幂等）
	var cnt int64
	if err := DB.Table("sys_user_post").Where("user_id = ? AND post_id = ?", 1, 1).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		if err := DB.Exec("INSERT INTO sys_user_post (user_id, post_id) VALUES (1, 1)").Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureConfigs 系统参数不存在时写入（幂等，内置参数）
func ensureConfigs() error {
	var count int64
	if err := DB.Model(&model.Config{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	configs := []model.Config{
		{ID: 1, ConfigName: "用户管理-初始密码", ConfigKey: "sys.user.initPassword", ConfigValue: "123456", ConfigType: "Y", Remark: "创建用户时使用的初始密码"},
		{ID: 2, ConfigName: "主框架页-侧边栏主题", ConfigKey: "sys.index.sidebarTheme", ConfigValue: "dark", ConfigType: "Y", Remark: "深色主题"},
		{ID: 3, ConfigName: "账号自助-验证码开关", ConfigKey: "sys.account.captchaEnabled", ConfigValue: "true", ConfigType: "Y", Remark: "是否开启登录验证码"},
		{ID: 4, ConfigName: "账号自助-注册开关", ConfigKey: "sys.account.registerUser", ConfigValue: "false", ConfigType: "Y", Remark: "是否开放自助注册"},
	}
	for i := range configs {
		if err := DB.Create(&configs[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureNotices 通知公告不存在时写入示例公告（幂等）
func ensureNotices() error {
	var count int64
	if err := DB.Model(&model.Notice{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	notice := model.Notice{
		Title:    "欢迎使用 Go Admin 权限管理系统",
		Type:     "1",
		Content:  "本系统基于 Gin + Vue3 + Element Plus + SQLite 构建，实现了 RuoYi 核心权限功能：登录验证码、JWT 鉴权、RBAC 角色权限、用户/角色/菜单/部门/字典/岗位/参数配置管理、操作日志、动态路由与按钮级权限控制。",
		Status:   "0",
		CreateBy: "admin",
		Remark:   "系统初始化公告",
	}
	return DB.Create(&notice).Error
}

// ensureJobs 定时任务不存在时写入示例任务（幂等；已有库按调用目标补插新示例）
func ensureJobs() error {
	var count int64
	if err := DB.Model(&model.Job{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		jobs := []model.Job{
			{ID: 1, JobName: "演示任务", JobGroup: "DEFAULT", InvokeTarget: "demoTask", CronExpression: "0 */5 * * * *", MisfirePolicy: "3", Concurrent: "1", Status: "0", CreateBy: "admin", Remark: "每 5 分钟执行一次演示任务"},
			{ID: 2, JobName: "日志清理", JobGroup: "SYSTEM", InvokeTarget: "clearExpiredLogs", CronExpression: "0 0 1 * * *", MisfirePolicy: "3", Concurrent: "1", Status: "1", CreateBy: "admin", Remark: "每天凌晨 1 点清理 30 天前操作/登录日志（示例暂停）"},
			{ID: 3, JobName: "在线会话清理", JobGroup: "SYSTEM", InvokeTarget: "cleanExpiredOnlines", CronExpression: "0 0 */2 * * *", MisfirePolicy: "3", Concurrent: "1", Status: "0", CreateBy: "admin", Remark: "每 2 小时清理一次过期在线会话"},
			{ID: 4, JobName: "缓存统计快照", JobGroup: "SYSTEM", InvokeTarget: "cacheSnapshot", CronExpression: "0 30 9 * * *", MisfirePolicy: "3", Concurrent: "1", Status: "1", CreateBy: "admin", Remark: "每天 9:30 记录缓存命中率与占用统计（示例暂停）"},
		}
		for i := range jobs {
			if err := DB.Create(&jobs[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}
	// 已有库：按调用目标幂等补插新示例任务
	extra := []model.Job{
		{JobName: "在线会话清理", JobGroup: "SYSTEM", InvokeTarget: "cleanExpiredOnlines", CronExpression: "0 0 */2 * * *", MisfirePolicy: "3", Concurrent: "1", Status: "0", CreateBy: "admin", Remark: "每 2 小时清理一次过期在线会话"},
		{JobName: "缓存统计快照", JobGroup: "SYSTEM", InvokeTarget: "cacheSnapshot", CronExpression: "0 30 9 * * *", MisfirePolicy: "3", Concurrent: "1", Status: "1", CreateBy: "admin", Remark: "每天 9:30 记录缓存命中率与占用统计（示例暂停）"},
	}
	for i := range extra {
		var n int64
		if err := DB.Model(&model.Job{}).Where("invoke_target = ?", extra[i].InvokeTarget).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			if err := DB.Create(&extra[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureFlows 示例审批流程（按 def_key 幂等写入；首次启动插入两条演示流程）
func ensureFlows() error {
	var count int64
	if err := DB.Model(&model.FlowDef{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	flows := []struct {
		def  model.FlowDef
		node []model.FlowNode
	}{
		{
			def: model.FlowDef{DefName: "通用审批", DefKey: "general", Category: "审批", Status: "0", Remark: "一级审批：指定超级管理员审批（演示用）"},
			node: []model.FlowNode{
				{NodeOrder: 1, NodeName: "管理员审批", ApproverType: "user", ApproverIDs: "1", NodeMode: "one", CanAddSign: "1", RejectTo: "0", Remark: "或签：管理员通过即完成"},
			},
		},
		{
			def: model.FlowDef{DefName: "请假审批", DefKey: "leave", Category: "请假", Status: "0", Remark: "两级审批：部门负责人 → 超级管理员（演示用）"},
			node: []model.FlowNode{
				{NodeOrder: 1, NodeName: "部门负责人审批", ApproverType: "dept_leader", ApproverIDs: "", NodeMode: "one", CanAddSign: "1", RejectTo: "1", Remark: "自动路由发起人部门负责人"},
				{NodeOrder: 2, NodeName: "管理员终审", ApproverType: "role", ApproverIDs: "1", NodeMode: "one", CanAddSign: "1", RejectTo: "0", Remark: "超级管理员角色或签"},
			},
		},
	}
	for _, f := range flows {
		if err := DB.Create(&f.def).Error; err != nil {
			return err
		}
		for i := range f.node {
			f.node[i].FlowDefID = f.def.ID
			if err := DB.Create(&f.node[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureApis 接口权限种子（按 path+method 幂等补齐缺失，权限字符与路由注册一致；perms 空表示仅需登录）
func ensureApis() error {
	apis := []model.Api{
		{APIName: "用户信息", Path: "/api/getInfo", Method: "GET"},
		{APIName: "动态路由", Path: "/api/getRouters", Method: "GET"},
		{APIName: "退出登录", Path: "/api/logout", Method: "POST"},
		{APIName: "个人信息查询", Path: "/api/system/user/profile", Method: "GET"},
		{APIName: "个人信息更新", Path: "/api/system/user/profile", Method: "PUT"},
		{APIName: "修改密码", Path: "/api/system/user/profile/updatePwd", Method: "PUT"},
		{APIName: "用户列表", Path: "/api/system/user/list", Method: "GET", Perms: "system:user:list"},
		{APIName: "用户导出", Path: "/api/system/user/export", Method: "GET", Perms: "system:user:export"},
		{APIName: "用户导入模板", Path: "/api/system/user/importTemplate", Method: "GET", Perms: "system:user:import"},
		{APIName: "用户导入", Path: "/api/system/user/importData", Method: "POST", Perms: "system:user:import"},
		{APIName: "用户详情", Path: "/api/system/user/:id", Method: "GET", Perms: "system:user:query"},
		{APIName: "新增用户", Path: "/api/system/user", Method: "POST", Perms: "system:user:add"},
		{APIName: "修改用户", Path: "/api/system/user", Method: "PUT", Perms: "system:user:edit"},
		{APIName: "删除用户", Path: "/api/system/user/:id", Method: "DELETE", Perms: "system:user:remove"},
		{APIName: "重置密码", Path: "/api/system/user/resetPwd", Method: "PUT", Perms: "system:user:resetPwd"},
		{APIName: "角色列表", Path: "/api/system/role/list", Method: "GET", Perms: "system:role:list"},
		{APIName: "角色下拉", Path: "/api/system/role/all", Method: "GET", Perms: "system:role:list"},
		{APIName: "角色详情", Path: "/api/system/role/:id", Method: "GET", Perms: "system:role:query"},
		{APIName: "新增角色", Path: "/api/system/role", Method: "POST", Perms: "system:role:add"},
		{APIName: "修改角色", Path: "/api/system/role", Method: "PUT", Perms: "system:role:edit"},
		{APIName: "删除角色", Path: "/api/system/role/:id", Method: "DELETE", Perms: "system:role:remove"},
		{APIName: "菜单列表", Path: "/api/system/menu/list", Method: "GET", Perms: "system:menu:list"},
		{APIName: "菜单树", Path: "/api/system/menu/treeselect", Method: "GET", Perms: "system:menu:list"},
		{APIName: "角色菜单树", Path: "/api/system/menu/roleMenuTreeselect/:roleId", Method: "GET", Perms: "system:role:query"},
		{APIName: "菜单详情", Path: "/api/system/menu/:id", Method: "GET", Perms: "system:menu:query"},
		{APIName: "新增菜单", Path: "/api/system/menu", Method: "POST", Perms: "system:menu:add"},
		{APIName: "修改菜单", Path: "/api/system/menu", Method: "PUT", Perms: "system:menu:edit"},
		{APIName: "删除菜单", Path: "/api/system/menu/:id", Method: "DELETE", Perms: "system:menu:remove"},
		{APIName: "部门列表", Path: "/api/system/dept/list", Method: "GET", Perms: "system:dept:list"},
		{APIName: "部门树", Path: "/api/system/dept/treeselect", Method: "GET", Perms: "system:user:list"},
		{APIName: "部门详情", Path: "/api/system/dept/:id", Method: "GET", Perms: "system:dept:query"},
		{APIName: "新增部门", Path: "/api/system/dept", Method: "POST", Perms: "system:dept:add"},
		{APIName: "修改部门", Path: "/api/system/dept", Method: "PUT", Perms: "system:dept:edit"},
		{APIName: "删除部门", Path: "/api/system/dept/:id", Method: "DELETE", Perms: "system:dept:remove"},
		{APIName: "字典类型列表", Path: "/api/system/dict/type/list", Method: "GET", Perms: "system:dict:list"},
		{APIName: "字典类型详情", Path: "/api/system/dict/type/:id", Method: "GET", Perms: "system:dict:query"},
		{APIName: "新增字典类型", Path: "/api/system/dict/type", Method: "POST", Perms: "system:dict:add"},
		{APIName: "修改字典类型", Path: "/api/system/dict/type", Method: "PUT", Perms: "system:dict:edit"},
		{APIName: "删除字典类型", Path: "/api/system/dict/type/:id", Method: "DELETE", Perms: "system:dict:remove"},
		{APIName: "字典数据列表", Path: "/api/system/dict/data/list", Method: "GET", Perms: "system:dict:list"},
		{APIName: "字典数据按类型", Path: "/api/system/dict/data/type/:dictType", Method: "GET", Perms: "system:dict:query"},
		{APIName: "字典数据详情", Path: "/api/system/dict/data/:dictCode", Method: "GET", Perms: "system:dict:query"},
		{APIName: "新增字典数据", Path: "/api/system/dict/data", Method: "POST", Perms: "system:dict:add"},
		{APIName: "修改字典数据", Path: "/api/system/dict/data", Method: "PUT", Perms: "system:dict:edit"},
		{APIName: "删除字典数据", Path: "/api/system/dict/data/:dictCode", Method: "DELETE", Perms: "system:dict:remove"},
		{APIName: "岗位列表", Path: "/api/system/post/list", Method: "GET", Perms: "system:post:list"},
		{APIName: "岗位下拉", Path: "/api/system/post/all", Method: "GET", Perms: "system:post:list"},
		{APIName: "岗位详情", Path: "/api/system/post/:id", Method: "GET", Perms: "system:post:query"},
		{APIName: "新增岗位", Path: "/api/system/post", Method: "POST", Perms: "system:post:add"},
		{APIName: "修改岗位", Path: "/api/system/post", Method: "PUT", Perms: "system:post:edit"},
		{APIName: "删除岗位", Path: "/api/system/post/:id", Method: "DELETE", Perms: "system:post:remove"},
		{APIName: "参数列表", Path: "/api/system/config/list", Method: "GET", Perms: "system:config:list"},
		{APIName: "参数详情", Path: "/api/system/config/:id", Method: "GET", Perms: "system:config:query"},
		{APIName: "新增参数", Path: "/api/system/config", Method: "POST", Perms: "system:config:add"},
		{APIName: "修改参数", Path: "/api/system/config", Method: "PUT", Perms: "system:config:edit"},
		{APIName: "删除参数", Path: "/api/system/config/:id", Method: "DELETE", Perms: "system:config:remove"},
		{APIName: "登录日志列表", Path: "/api/monitor/logininfor/list", Method: "GET", Perms: "monitor:logininfor:list"},
		{APIName: "操作日志列表", Path: "/api/monitor/operlog/list", Method: "GET", Perms: "monitor:operlog:list"},
		{APIName: "清空操作日志", Path: "/api/monitor/operlog/clean", Method: "DELETE", Perms: "monitor:operlog:remove"},
		{APIName: "删除操作日志", Path: "/api/monitor/operlog/:operId", Method: "DELETE", Perms: "monitor:operlog:remove"},
		{APIName: "公告列表", Path: "/api/system/notice/list", Method: "GET", Perms: "system:notice:list"},
		{APIName: "首页公告", Path: "/api/system/notice/active", Method: "GET", Perms: "system:notice:list"},
		{APIName: "公告详情", Path: "/api/system/notice/:id", Method: "GET", Perms: "system:notice:query"},
		{APIName: "新增公告", Path: "/api/system/notice", Method: "POST", Perms: "system:notice:add"},
		{APIName: "修改公告", Path: "/api/system/notice", Method: "PUT", Perms: "system:notice:edit"},
		{APIName: "删除公告", Path: "/api/system/notice/:id", Method: "DELETE", Perms: "system:notice:remove"},
		{APIName: "在线用户列表", Path: "/api/monitor/online/list", Method: "GET", Perms: "monitor:online:list"},
		{APIName: "全部强退", Path: "/api/monitor/online/all", Method: "DELETE", Perms: "monitor:online:forceLogout"},
		{APIName: "单条强退", Path: "/api/monitor/online/:tokenHash", Method: "DELETE", Perms: "monitor:online:forceLogout"},
		{APIName: "任务列表", Path: "/api/monitor/job/list", Method: "GET", Perms: "monitor:job:list"},
		{APIName: "任务执行器列表", Path: "/api/monitor/job/executors", Method: "GET", Perms: "monitor:job:list"},
		{APIName: "任务详情", Path: "/api/monitor/job/:id", Method: "GET", Perms: "monitor:job:query"},
		{APIName: "新增任务", Path: "/api/monitor/job", Method: "POST", Perms: "monitor:job:add"},
		{APIName: "修改任务", Path: "/api/monitor/job", Method: "PUT", Perms: "monitor:job:edit"},
		{APIName: "任务启停", Path: "/api/monitor/job/changeStatus", Method: "PUT", Perms: "monitor:job:changeStatus"},
		{APIName: "任务立即执行", Path: "/api/monitor/job/run", Method: "PUT", Perms: "monitor:job:run"},
		{APIName: "删除任务", Path: "/api/monitor/job/:id", Method: "DELETE", Perms: "monitor:job:remove"},
		{APIName: "执行日志列表", Path: "/api/monitor/jobLog/list", Method: "GET", Perms: "monitor:job:list"},
		{APIName: "清空执行日志", Path: "/api/monitor/jobLog/clean", Method: "DELETE", Perms: "monitor:job:remove"},
		{APIName: "删除执行日志", Path: "/api/monitor/jobLog/:id", Method: "DELETE", Perms: "monitor:job:remove"},
		{APIName: "服务监控信息", Path: "/api/monitor/server/info", Method: "GET", Perms: "monitor:server:list"},
		{APIName: "缓存统计", Path: "/api/monitor/cache/info", Method: "GET", Perms: "monitor:cache:list"},
		{APIName: "缓存键详情", Path: "/api/monitor/cache/key", Method: "GET", Perms: "monitor:cache:list"},
		{APIName: "清空缓存", Path: "/api/monitor/cache/clear", Method: "DELETE", Perms: "monitor:cache:clear"},
		{APIName: "数据监控信息", Path: "/api/monitor/db/info", Method: "GET", Perms: "monitor:db:list"},
		{APIName: "接口列表", Path: "/api/system/api/list", Method: "GET", Perms: "system:api:list"},
		{APIName: "接口详情", Path: "/api/system/api/:id", Method: "GET", Perms: "system:api:query"},
		{APIName: "新增接口", Path: "/api/system/api", Method: "POST", Perms: "system:api:add"},
		{APIName: "修改接口", Path: "/api/system/api", Method: "PUT", Perms: "system:api:edit"},
		{APIName: "接口启停", Path: "/api/system/api/changeStatus", Method: "PUT", Perms: "system:api:changeStatus"},
		{APIName: "删除接口", Path: "/api/system/api/:id", Method: "DELETE", Perms: "system:api:remove"},
		{APIName: "生成表列表", Path: "/api/tool/gen/db/list", Method: "GET", Perms: "tool:gen:list"},
		{APIName: "导入生成表", Path: "/api/tool/gen/importTable", Method: "POST", Perms: "tool:gen:import"},
		{APIName: "生成配置列表", Path: "/api/tool/gen/list", Method: "GET", Perms: "tool:gen:list"},
		{APIName: "生成配置详情", Path: "/api/tool/gen/:tableId", Method: "GET", Perms: "tool:gen:query"},
		{APIName: "保存生成配置", Path: "/api/tool/gen", Method: "PUT", Perms: "tool:gen:edit"},
		{APIName: "删除生成配置", Path: "/api/tool/gen/:tableId", Method: "DELETE", Perms: "tool:gen:remove"},
		{APIName: "预览生成代码", Path: "/api/tool/gen/preview/:tableId", Method: "GET", Perms: "tool:gen:preview"},
		{APIName: "下载生成代码", Path: "/api/tool/gen/code/:tableId", Method: "GET", Perms: "tool:gen:code"},
		{APIName: "接口文档", Path: "/api/tool/api-doc/list", Method: "GET", Perms: "tool:apidoc:list"},
		{APIName: "通用文件上传", Path: "/api/common/upload", Method: "POST", Perms: "common:upload"},
		// 工作流/审批（流程中心）
		{APIName: "流程定义列表", Path: "/api/flow/def/list", Method: "GET", Perms: "flow:def:list"},
		{APIName: "启用流程列表", Path: "/api/flow/def/all", Method: "GET", Perms: "flow:def:list"},
		{APIName: "流程定义详情", Path: "/api/flow/def/:id", Method: "GET", Perms: "flow:def:query"},
		{APIName: "新增流程定义", Path: "/api/flow/def", Method: "POST", Perms: "flow:def:add"},
		{APIName: "修改流程定义", Path: "/api/flow/def", Method: "PUT", Perms: "flow:def:edit"},
		{APIName: "流程启停", Path: "/api/flow/def/changeStatus", Method: "PUT", Perms: "flow:def:changeStatus"},
		{APIName: "删除流程定义", Path: "/api/flow/def/:id", Method: "DELETE", Perms: "flow:def:remove"},
		{APIName: "我的申请列表", Path: "/api/flow/apply/list", Method: "GET", Perms: "flow:apply:list"},
		{APIName: "发起申请", Path: "/api/flow/apply", Method: "POST", Perms: "flow:apply:add"},
		{APIName: "撤回申请", Path: "/api/flow/apply/withdraw/:id", Method: "PUT", Perms: "flow:apply:withdraw"},
		{APIName: "重新提交", Path: "/api/flow/apply/reapply/:id", Method: "PUT", Perms: "flow:apply:reapply"},
		{APIName: "催办", Path: "/api/flow/apply/urge/:id", Method: "PUT", Perms: "flow:apply:urge"},
		{APIName: "终止申请", Path: "/api/flow/apply/terminate/:id", Method: "PUT", Perms: "flow:apply:terminate"},
		{APIName: "待办列表", Path: "/api/flow/todo/list", Method: "GET", Perms: "flow:todo:list"},
		{APIName: "审批通过", Path: "/api/flow/todo/approve", Method: "PUT", Perms: "flow:todo:approve"},
		{APIName: "审批驳回", Path: "/api/flow/todo/reject", Method: "PUT", Perms: "flow:todo:reject"},
		{APIName: "转办", Path: "/api/flow/todo/transfer", Method: "PUT", Perms: "flow:todo:transfer"},
		{APIName: "加签", Path: "/api/flow/todo/addSign", Method: "PUT", Perms: "flow:todo:addSign"},
		{APIName: "已办列表", Path: "/api/flow/done/list", Method: "GET", Perms: "flow:todo:list"},
		{APIName: "抄送列表", Path: "/api/flow/cc/list", Method: "GET", Perms: "flow:cc:list"},
		{APIName: "抄送已读", Path: "/api/flow/cc/read/:id", Method: "PUT", Perms: "flow:cc:list"},
		{APIName: "全部实例列表", Path: "/api/flow/instance/list", Method: "GET", Perms: "flow:instance:list"},
		{APIName: "实例详情", Path: "/api/flow/instance/:id", Method: "GET", Perms: "flow:instance:query"},
	}
	for i := range apis {
		var cnt int64
		if err := DB.Model(&model.Api{}).Where("path = ? AND method = ?", apis[i].Path, apis[i].Method).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			continue
		}
		if err := DB.Create(&apis[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
