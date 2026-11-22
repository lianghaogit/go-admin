package router

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"go-admin/api"
	"go-admin/config"
	"go-admin/genreg"
	"go-admin/middleware"
	"go-admin/web"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	// 访问日志带 request_id，便于与 operlog / 业务错误日志关联
	r.Use(gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		rid, _ := p.Keys[middleware.RequestIDKey].(string)
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s | rid=%s | %s\n",
			p.TimeStamp.Format("2006-01-02 15:04:05"), p.StatusCode, p.Latency, p.ClientIP,
			p.Method, p.Path, rid, p.ErrorMessage)
	}))
	r.Use(gin.Recovery())
	r.Use(middleware.BodyLimit())
	r.Use(middleware.Cors())

	// 上传文件静态访问
	// 安全：svg/html/xml 等可含脚本类型强制下载（Content-Disposition: attachment），
	// 避免浏览器同源内嵌渲染导致存储型 XSS
	if dir := config.Cfg.Server.UploadDir; dir != "" {
		r.Use(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
				switch strings.ToLower(filepath.Ext(c.Request.URL.Path)) {
				case ".svg", ".html", ".htm", ".xml":
					c.Header("Content-Disposition", "attachment")
					c.Header("Content-Type", "application/octet-stream")
				}
			}
			c.Next()
		})
		r.Static("/uploads", dir)
	}

	apiGroup := r.Group("/api")
	{
		// 公共接口
		apiGroup.GET("/captchaImage", api.CaptchaImage)
		apiGroup.POST("/login", api.Login)

		// 需要登录
		auth := apiGroup.Group("")
		auth.Use(middleware.Auth())
		auth.Use(middleware.OperLog())
		auth.Use(middleware.CheckApiPerm())
		{
			auth.GET("/getInfo", api.GetInfo)
			auth.GET("/getRouters", api.GetRouters)
			auth.POST("/logout", api.Logout)

			// 首页统计
			auth.GET("/dashboard/stats", api.DashStats)
			auth.GET("/dashboard/loginTrend", api.LoginTrend)
			auth.GET("/dashboard/operTypeStats", api.OperTypeStats)
			auth.GET("/dashboard/loginHeatmap", api.LoginHeatmap)

			// 个人中心
			auth.GET("/system/user/profile", api.GetProfile)
			auth.PUT("/system/user/profile", api.UpdateProfile)
			auth.PUT("/system/user/profile/avatar", api.UpdateAvatar)
			auth.PUT("/system/user/profile/updatePwd", api.UpdatePwd)

			// 用户管理
			user := auth.Group("/system/user")
			{
				user.GET("/list", middleware.CheckPerm("system:user:list"), api.ListUsers)
				user.GET("/:id", middleware.CheckPerm("system:user:query"), api.GetUser)
				user.POST("", middleware.CheckPerm("system:user:add"), api.CreateUser)
				user.PUT("", middleware.CheckPerm("system:user:edit"), api.UpdateUser)
				user.DELETE("/:id", middleware.CheckPerm("system:user:remove"), api.DeleteUser)
				user.PUT("/resetPwd", middleware.CheckPerm("system:user:resetPwd"), api.ResetPwd)
				user.PUT("/changeStatus", middleware.CheckPerm("system:user:edit"), api.ChangeUserStatus)
				user.PUT("/authRole", middleware.CheckPerm("system:user:edit"), api.AuthRole)
			}

			// 角色管理
			role := auth.Group("/system/role")
			{
				role.GET("/list", middleware.CheckPerm("system:role:list"), api.ListRoles)
				role.GET("/all", middleware.CheckPerm("system:role:list"), api.ListAllRoles)
				role.GET("/:id", middleware.CheckPerm("system:role:query"), api.GetRole)
				role.POST("", middleware.CheckPerm("system:role:add"), api.CreateRole)
				role.PUT("", middleware.CheckPerm("system:role:edit"), api.UpdateRole)
				role.DELETE("/:id", middleware.CheckPerm("system:role:remove"), api.DeleteRole)
				// 角色分配用户
				role.GET("/authUser/allocatedList", middleware.CheckPerm("system:role:list"), api.ListRoleAllocatedUsers)
				role.GET("/authUser/unallocatedList", middleware.CheckPerm("system:role:list"), api.ListRoleUnallocatedUsers)
				role.PUT("/authUser", middleware.CheckPerm("system:role:edit"), api.AssignAuthUsers)
				role.DELETE("/authUser", middleware.CheckPerm("system:role:edit"), api.CancelAuthUsers)
			}

			// 菜单管理
			menu := auth.Group("/system/menu")
			{
				menu.GET("/list", middleware.CheckPerm("system:menu:list"), api.ListMenus)
				menu.GET("/treeselect", middleware.CheckPerm("system:menu:list"), api.MenuTreeSelect)
				menu.GET("/roleMenuTreeselect/:roleId", middleware.CheckPerm("system:role:query"), api.RoleMenuTreeSelect)
				menu.GET("/:id", middleware.CheckPerm("system:menu:query"), api.GetMenu)
				menu.POST("", middleware.CheckPerm("system:menu:add"), api.CreateMenu)
				menu.PUT("", middleware.CheckPerm("system:menu:edit"), api.UpdateMenu)
				menu.DELETE("/:id", middleware.CheckPerm("system:menu:remove"), api.DeleteMenu)
			}

			// 部门管理
			dept := auth.Group("/system/dept")
			{
				dept.GET("/list", middleware.CheckPerm("system:dept:list"), api.ListDepts)
				dept.GET("/treeselect", middleware.CheckPerm("system:user:list"), api.DeptTreeSelect)
				dept.GET("/:id", middleware.CheckPerm("system:dept:query"), api.GetDept)
				dept.POST("", middleware.CheckPerm("system:dept:add"), api.CreateDept)
				dept.PUT("", middleware.CheckPerm("system:dept:edit"), api.UpdateDept)
				dept.DELETE("/:id", middleware.CheckPerm("system:dept:remove"), api.DeleteDept)
			}

			// 字典管理
			dict := auth.Group("/system/dict")
			{
				dict.GET("/type/list", middleware.CheckPerm("system:dict:list"), api.ListDictTypes)
				dict.GET("/type/:id", middleware.CheckPerm("system:dict:query"), api.GetDictType)
				dict.POST("/type", middleware.CheckPerm("system:dict:add"), api.CreateDictType)
				dict.PUT("/type", middleware.CheckPerm("system:dict:edit"), api.UpdateDictType)
				dict.DELETE("/type/:id", middleware.CheckPerm("system:dict:remove"), api.DeleteDictType)

				dict.GET("/data/list", middleware.CheckPerm("system:dict:list"), api.ListDictData)
				dict.GET("/data/type/:dictType", middleware.CheckPerm("system:dict:query"), api.DictDataByType)
				dict.GET("/data/:dictCode", middleware.CheckPerm("system:dict:query"), api.GetDictData)
				dict.POST("/data", middleware.CheckPerm("system:dict:add"), api.CreateDictData)
				dict.PUT("/data", middleware.CheckPerm("system:dict:edit"), api.UpdateDictData)
				dict.DELETE("/data/:dictCode", middleware.CheckPerm("system:dict:remove"), api.DeleteDictData)
			}

			// 岗位管理
			post := auth.Group("/system/post")
			{
				post.GET("/list", middleware.CheckPerm("system:post:list"), api.ListPosts)
				post.GET("/all", middleware.CheckPerm("system:post:list"), api.ListAllPosts)
				post.GET("/:id", middleware.CheckPerm("system:post:query"), api.GetPost)
				post.POST("", middleware.CheckPerm("system:post:add"), api.CreatePost)
				post.PUT("", middleware.CheckPerm("system:post:edit"), api.UpdatePost)
				post.DELETE("/:id", middleware.CheckPerm("system:post:remove"), api.DeletePost)
			}

			// 参数配置
			config := auth.Group("/system/config")
			{
				config.GET("/list", middleware.CheckPerm("system:config:list"), api.ListConfigs)
				config.GET("/:id", middleware.CheckPerm("system:config:query"), api.GetConfig)
				config.POST("", middleware.CheckPerm("system:config:add"), api.CreateConfig)
				config.PUT("", middleware.CheckPerm("system:config:edit"), api.UpdateConfig)
				config.DELETE("/:id", middleware.CheckPerm("system:config:remove"), api.DeleteConfig)
			}

			// 登录日志
			log := auth.Group("/monitor/logininfor")
			{
				log.GET("/list", middleware.CheckPerm("monitor:logininfor:list"), api.ListLoginLogs)
				log.GET("/export", middleware.CheckPerm("monitor:logininfor:list"), api.ExportLoginLogs)
				log.DELETE("/clean", middleware.CheckPerm("monitor:logininfor:remove"), api.CleanLoginLog)
				log.DELETE("/:infoId", middleware.CheckPerm("monitor:logininfor:remove"), api.DeleteLoginLog)
			}

			// 操作日志
			operlog := auth.Group("/monitor/operlog")
			{
				operlog.GET("/list", middleware.CheckPerm("monitor:operlog:list"), api.ListOperLogs)
				operlog.GET("/export", middleware.CheckPerm("monitor:operlog:list"), api.ExportOperLogs)
				operlog.DELETE("/clean", middleware.CheckPerm("monitor:operlog:remove"), api.CleanOperLog)
				operlog.DELETE("/:operId", middleware.CheckPerm("monitor:operlog:remove"), api.DeleteOperLog)
			}

			// 通知公告
			notice := auth.Group("/system/notice")
			{
				notice.GET("/list", middleware.CheckPerm("system:notice:list"), api.ListNotices)
				notice.GET("/active", middleware.CheckPerm("system:notice:list"), api.ListActiveNotices)
				notice.GET("/:id", middleware.CheckPerm("system:notice:query"), api.GetNotice)
				notice.POST("", middleware.CheckPerm("system:notice:add"), api.CreateNotice)
				notice.PUT("", middleware.CheckPerm("system:notice:edit"), api.UpdateNotice)
				notice.DELETE("/:id", middleware.CheckPerm("system:notice:remove"), api.DeleteNotice)
			}

			// 在线用户
			online := auth.Group("/monitor/online")
			{
				online.GET("/list", middleware.CheckPerm("monitor:online:list"), api.ListOnlines)
				online.DELETE("/all", middleware.CheckPerm("monitor:online:forceLogout"), api.ForceLogoutAll)
				online.DELETE("/:tokenHash", middleware.CheckPerm("monitor:online:forceLogout"), api.ForceLogout)
			}

			// 定时任务
			job := auth.Group("/monitor/job")
			{
				job.GET("/list", middleware.CheckPerm("monitor:job:list"), api.ListJobs)
				job.GET("/executors", middleware.CheckPerm("monitor:job:list"), api.ListJobExecutors)
				job.GET("/:id", middleware.CheckPerm("monitor:job:query"), api.GetJob)
				job.POST("", middleware.CheckPerm("monitor:job:add"), api.CreateJob)
				job.PUT("", middleware.CheckPerm("monitor:job:edit"), api.UpdateJob)
				job.PUT("/changeStatus", middleware.CheckPerm("monitor:job:changeStatus"), api.ChangeJobStatus)
				job.PUT("/run", middleware.CheckPerm("monitor:job:run"), api.RunJob)
				job.DELETE("/:id", middleware.CheckPerm("monitor:job:remove"), api.DeleteJob)
			}

			// 任务执行日志
			jobLog := auth.Group("/monitor/jobLog")
			{
				jobLog.GET("/list", middleware.CheckPerm("monitor:job:list"), api.ListJobLogs)
				jobLog.DELETE("/clean", middleware.CheckPerm("monitor:job:remove"), api.CleanJobLogs)
				jobLog.DELETE("/:id", middleware.CheckPerm("monitor:job:remove"), api.DeleteJobLog)
			}

			// 服务监控
			server := auth.Group("/monitor/server")
			{
				server.GET("/info", middleware.CheckPerm("monitor:server:list"), api.ServerInfo)
			}

			// 缓存监控
			cacheG := auth.Group("/monitor/cache")
			{
				cacheG.GET("/info", middleware.CheckPerm("monitor:cache:list"), api.CacheInfo)
				cacheG.GET("/key", middleware.CheckPerm("monitor:cache:list"), api.CacheKeyDetail)
				cacheG.DELETE("/clear", middleware.CheckPerm("monitor:cache:clear"), api.CacheClear)
			}

			// 数据监控（SQLite）
			dbG := auth.Group("/monitor/db")
			{
				dbG.GET("/info", middleware.CheckPerm("monitor:db:list"), api.DbInfo)
			}

			// 接口权限
			apiG := auth.Group("/system/api")
			{
				apiG.GET("/list", middleware.CheckPerm("system:api:list"), api.ListApis)
				apiG.GET("/:id", middleware.CheckPerm("system:api:query"), api.GetApi)
				apiG.POST("", middleware.CheckPerm("system:api:add"), api.CreateApi)
				apiG.PUT("", middleware.CheckPerm("system:api:edit"), api.UpdateApi)
				apiG.PUT("/changeStatus", middleware.CheckPerm("system:api:changeStatus"), api.ChangeApiStatus)
				apiG.DELETE("/:id", middleware.CheckPerm("system:api:remove"), api.DeleteApi)
			}

			// 代码生成
			gen := auth.Group("/tool/gen")
			{
				gen.GET("/db/list", middleware.CheckPerm("tool:gen:list"), api.ListDbTables)
				gen.POST("/importTable", middleware.CheckPerm("tool:gen:import"), api.ImportGenTables)
				gen.GET("/list", middleware.CheckPerm("tool:gen:list"), api.ListGenTables)
				gen.GET("/:tableId", middleware.CheckPerm("tool:gen:query"), api.GetGenTable)
				gen.PUT("", middleware.CheckPerm("tool:gen:edit"), api.SaveGenTable)
				gen.DELETE("/:tableId", middleware.CheckPerm("tool:gen:remove"), api.DeleteGenTable)
				gen.GET("/preview/:tableId", middleware.CheckPerm("tool:gen:preview"), api.PreviewGenTable)
				gen.GET("/code/:tableId", middleware.CheckPerm("tool:gen:code"), api.DownloadGenCode)
			}

			// 用户导出/导入
			user.GET("/export", middleware.CheckPerm("system:user:export"), api.ExportUsers)
			user.GET("/importTemplate", middleware.CheckPerm("system:user:import"), api.ImportUserTemplate)
			user.POST("/importData", middleware.CheckPerm("system:user:import"), api.ImportUsers)

			// 系统接口文档（枚举全部路由）
			doc := auth.Group("/tool/api-doc")
			{
				doc.GET("/list", middleware.CheckPerm("tool:apidoc:list"), func(c *gin.Context) {
					api.ApiDocList(c, r.Routes())
				})
			}

			// 通用文件上传
			common := auth.Group("/common")
			{
				common.POST("/upload", middleware.CheckPerm("common:upload"), api.UploadFile)
			}

			// 工作流/审批（流程中心）
			flow := auth.Group("/flow")
			{
				// 流程定义
				flow.GET("/def/list", middleware.CheckPerm("flow:def:list"), api.ListFlowDefs)
				flow.GET("/def/all", middleware.CheckPerm("flow:def:list"), api.ListEnabledFlowDefs)
				flow.GET("/def/:id", middleware.CheckPerm("flow:def:query"), api.GetFlowDef)
				flow.POST("/def", middleware.CheckPerm("flow:def:add"), api.CreateFlowDef)
				flow.PUT("/def", middleware.CheckPerm("flow:def:edit"), api.UpdateFlowDef)
				flow.PUT("/def/changeStatus", middleware.CheckPerm("flow:def:changeStatus"), api.ChangeFlowDefStatus)
				flow.DELETE("/def/:id", middleware.CheckPerm("flow:def:remove"), api.DeleteFlowDef)
				// 发起与我的申请
				flow.GET("/apply/list", middleware.CheckPerm("flow:apply:list"), api.ListFlowApply)
				flow.POST("/apply", middleware.CheckPerm("flow:apply:add"), api.StartFlow)
				flow.PUT("/apply/withdraw/:id", middleware.CheckPerm("flow:apply:withdraw"), api.WithdrawFlow)
				flow.PUT("/apply/reapply/:id", middleware.CheckPerm("flow:apply:reapply"), api.ReapplyFlow)
				flow.PUT("/apply/urge/:id", middleware.CheckPerm("flow:apply:urge"), api.UrgeFlow)
				flow.PUT("/apply/terminate/:id", middleware.CheckPerm("flow:apply:terminate"), api.TerminateFlow)
				// 待办/已办/抄送
				flow.GET("/todo/list", middleware.CheckPerm("flow:todo:list"), api.ListFlowTodo)
				flow.PUT("/todo/approve", middleware.CheckPerm("flow:todo:approve"), api.ApproveFlowTask)
				flow.PUT("/todo/reject", middleware.CheckPerm("flow:todo:reject"), api.RejectFlowTask)
				flow.PUT("/todo/transfer", middleware.CheckPerm("flow:todo:transfer"), api.TransferFlowTask)
				flow.PUT("/todo/addSign", middleware.CheckPerm("flow:todo:addSign"), api.AddSignFlowTask)
				flow.GET("/done/list", middleware.CheckPerm("flow:todo:list"), api.ListFlowDone)
				flow.GET("/cc/list", middleware.CheckPerm("flow:cc:list"), api.ListFlowCc)
				flow.PUT("/cc/read/:id", middleware.CheckPerm("flow:cc:list"), api.MarkCcRead)
				// 实例
				flow.GET("/instance/list", middleware.CheckPerm("flow:instance:list"), api.ListFlowInstances)
				flow.GET("/instance/:id", middleware.CheckPerm("flow:instance:query"), api.GetFlowInstanceDetail)
			}

			// 代码生成模块自注册路由：生成的 api 文件 init() 登记，无需手工改本文件
			for _, m := range genreg.Modules() {
				mod := auth.Group("/" + m.Module + "/" + m.ClassVar)
				for _, rt := range m.Routes {
					switch rt.Method {
					case "GET":
						mod.GET(rt.Path, middleware.CheckPerm(rt.Perm), rt.Handler)
					case "POST":
						mod.POST(rt.Path, middleware.CheckPerm(rt.Perm), rt.Handler)
					case "PUT":
						mod.PUT(rt.Path, middleware.CheckPerm(rt.Perm), rt.Handler)
					case "DELETE":
						mod.DELETE(rt.Path, middleware.CheckPerm(rt.Perm), rt.Handler)
					}
				}
			}
		}
	}

	// 内嵌前端静态资源：未命中 API 的请求交给内嵌的 Vue 构建产物，
	// 非静态文件路径回退 index.html，以支持 history 模式路由
	if distFS, err := fs.Sub(web.Dist, "dist"); err == nil {
		fileServer := http.FileServer(http.FS(distFS))
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			if strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/uploads/") {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "资源不存在"})
				return
			}
			name := strings.TrimPrefix(p, "/")
			if name != "" {
				if f, err := distFS.Open(name); err == nil {
					_ = f.Close()
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
			}
			// SPA 回退：交给 Vue Router 处理
			c.Request.URL.Path = "/"
			fileServer.ServeHTTP(c.Writer, c.Request)
		})
	}
	return r
}
