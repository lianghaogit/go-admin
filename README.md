# Go Admin

一个用 Go + Vue 3 写的后台权限管理脚手架，后端 Gin + GORM，前端 Vue 3 + Element Plus，另外带一套 uni-app 移动端。权限模型参照 RuoYi 的思路实现：用户、角色、菜单三级关联，菜单下再分目录、菜单、按钮；除了菜单/按钮权限，还做了数据权限（五档数据范围）和接口权限（规则存在数据库里，可在页面上配置启停）。

项目里除了常规的增删改查页面，还包含代码生成、Excel 导入导出、定时任务、服务/缓存/数据监控、操作与登录日志、可视化流程审批、在线用户管理、接口文档自动枚举等模块。数据库支持 SQLite、MySQL、PostgreSQL，默认用 SQLite，纯 Go 驱动，不需要装 CGO 或本地数据库就能跑起来；配置文件里改一行类型即可切换到 MySQL 或 PostgreSQL。

## 界面预览

登录页采用左右分栏，左侧是品牌区，右侧是登录表单。

![登录页](docs/images/login.png)

首页仪表盘汇总了用户、角色、菜单、部门、公告数量，近 7 天登录趋势、登录热力分布、操作类型占比和当前服务器资源占用。

![首页仪表盘](docs/images/dashboard.png)

支持暗黑模式，图表颜色会跟随主题切换。

![暗黑模式](docs/images/dashboard-dark.png)

右上角齿轮打开布局配置抽屉，可以切换语言、主题色、九种布局模式、侧边栏深浅、页签/面包屑/Logo 等开关，以及组件尺寸，配置存在浏览器 localStorage 里，刷新不丢。

![布局配置抽屉](docs/images/settings-drawer.png)

## 主要页面截图

### 系统管理

<table>
  <tr>
    <td align="center"><img src="docs/images/system-user.png" width="100%"><br><sub>用户管理</sub></td>
    <td align="center"><img src="docs/images/system-role.png" width="100%"><br><sub>角色管理（含数据范围）</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/system-menu.png" width="100%"><br><sub>菜单管理（目录 / 菜单 / 按钮）</sub></td>
    <td align="center"><img src="docs/images/system-dept.png" width="100%"><br><sub>部门管理（树形）</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/system-dict.png" width="100%"><br><sub>字典管理</sub></td>
    <td align="center"><img src="docs/images/system-notice.png" width="100%"><br><sub>通知公告</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/system-post.png" width="100%"><br><sub>岗位管理</sub></td>
    <td align="center"><img src="docs/images/system-config.png" width="100%"><br><sub>参数配置</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/system-api.png" width="100%"><br><sub>接口权限规则</sub></td>
    <td align="center"><img src="docs/images/monitor-online.png" width="100%"><br><sub>在线用户</sub></td>
  </tr>
</table>

### 监控与日志

<table>
  <tr>
    <td align="center"><img src="docs/images/monitor-job.png" width="100%"><br><sub>定时任务（cron + 执行器）</sub></td>
    <td align="center"><img src="docs/images/monitor-server.png" width="100%"><br><sub>服务监控</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/monitor-cache.png" width="100%"><br><sub>缓存监控</sub></td>
    <td align="center"><img src="docs/images/monitor-db.png" width="100%"><br><sub>数据监控</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/system-operlog.png" width="100%"><br><sub>操作日志</sub></td>
    <td align="center"><img src="docs/images/system-loginlog.png" width="100%"><br><sub>登录日志</sub></td>
  </tr>
</table>

### 开发工具

<table>
  <tr>
    <td align="center"><img src="docs/images/tool-gen.png" width="100%"><br><sub>代码生成</sub></td>
    <td align="center"><img src="docs/images/tool-apidoc.png" width="100%"><br><sub>系统接口文档</sub></td>
  </tr>
  <tr>
    <td align="center"><img src="docs/images/tool-upload.png" width="100%"><br><sub>文件上传演示</sub></td>
    <td align="center"><img src="docs/images/flow-def.png" width="100%"><br><sub>流程定义</sub></td>
  </tr>
</table>

### 流程审批

![流程编排画布](docs/images/flow-editor.png)

## 技术栈

| 端 | 技术 | 说明 |
| --- | --- | --- |
| 后端 | Go 1.27、Gin、GORM、JWT、bcrypt、base64Captcha、robfig/cron、gopsutil | REST 接口、ORM 持久化、令牌鉴权、图形验证码、定时调度、系统指标采集 |
| 数据库 | SQLite（glebarez/sqlite，纯 Go，无需 CGO）/ MySQL / PostgreSQL | 配置 `database.type` 切换，首次启动自动建表并写入种子数据 |
| Web 前端 | Vue 3.5、Vite 5、TypeScript、Element Plus 2.8、Pinia、Vue Router 4、Axios、ECharts 6、AntV X6、vue-i18n | 动态路由、按钮级权限指令、图表、流程画布、中英文切换 |
| 移动端 | uni-app（Vue 3）、Vite 5、Pinia、uni.request | 同一套代码编译 H5、微信小程序和 App |
| 工程化 | ESLint、Prettier、Vitest、vue-tsc、go vet/gofmt、vite-plugin-compression | 前端 lint/类型检查/单测，产物 gzip + brotli 预压缩，依赖分包 |

## 功能一览

### 账号与鉴权

- 图形验证码登录，验证码可在配置文件里关闭；登录按账号和 IP 双维度限流，连续失败达到阈值后锁定 15 分钟。
- JWT Bearer 令牌，过期时间可配；鉴权中间件每次请求都会校验用户是否存在、是否停用，删除或停用账号后旧令牌立即失效。
- 在线会话表记录当前登录用户，支持单条强退和全部强退，强退后令牌马上不可用；会话到期自动清理。
- 登录日志记录用户名、IP、状态、提示信息和登录时间，支持按用户名、IP、状态、时间范围查询、导出和清空。

### 权限体系

- RBAC 模型：用户关联角色，角色关联菜单，菜单分目录（M）、菜单（C）、按钮（F）三级。
- 数据权限五档：全部数据、自定义部门、本部门、本部门及以下、仅本人；多角色命中时取最宽松的规则，自定义部门在 `sys_role_dept` 表里绑定。
- 接口权限：规则存 `sys_api` 表，记录请求方法、路径和权限字符，支持 `:param` 路径模板；全局中间件按规则校验，规则停用或未配置时放行，精确匹配优先于通配匹配。管理页面可以新增、修改、删除和启停规则。
- 按钮级权限：后端路由注册时标注权限字符，前端用 `v-hasPermi` 指令控制按钮显隐；内置超级管理员 `admin` 直接放行所有权限。

### 系统管理

- **用户管理**：分页查询、新增、修改、删除、重置密码、分配角色、岗位多选、归属部门；支持导出 xlsx、导入模板下载和 Excel 导入，逐行校验并返回成功/失败明细；列表带多选批量删除、行内状态开关和列显隐设置。
- **角色管理**：菜单权限分配、数据权限分配（五档 + 自定义部门）、分配用户弹窗、按角色反查已分配/未分配用户。
- **菜单管理**：树形表格维护目录、菜单、按钮，配置路由组件、权限标识、图标和排序；图标选择器内置常用 Element Plus 图标，支持搜索点选。
- **部门管理**：树形结构，有下级或有用户的部门不允许删除。
- **岗位管理**：岗位编码唯一、排序、状态，已分配用户的岗位不可删，用户与岗位多对多关联。
- **字典管理**：字典类型 + 字典数据，按类型查询，类型删除时有引用保护；字典下拉接口带 60 秒内存缓存，增删改后主动失效。
- **参数配置**：键名唯一，内置参数受保护（不可删、不可改类型，可改值）。
- **通知公告**：区分通知/公告，管理状态和详情，首页展示最新公告。
- **个人中心**：修改密码（校验旧密码）、更新昵称/手机/邮箱/备注、上传头像。

### 监控与运维

- **在线用户**：查看当前会话、浏览器、操作系统、登录 IP 和时间，支持强退单个或全部。
- **定时任务**：cron 表达式调度（秒 分 时 日 月 周），新增、修改、启停、删除、立即执行；内置演示任务、过期日志清理、内存快照、过期会话清理、缓存快照等执行器，也支持填 HTTP(S) 地址发起 GET 调用；每次执行写任务日志，记录成功/失败、消息和异常。
- **服务监控**：CPU、内存、磁盘使用率环形图，操作系统、CPU 核数、Go 版本、进程内存、启动时间和运行时长，定时自动刷新。
- **缓存监控**：内置内存缓存的命中/未命中次数、命中率、内存占用，按键前缀分组，可查看键详情和一键清空。
- **数据监控**：按数据库方言统计版本、库大小、连接数/上限、缓存命中率、各表行数和字段注释；SQLite 下展示文件大小和日志模式。
- **操作日志**：自动记录增删改查操作的模块、类型、请求方式、操作人、IP、URL、参数、状态、耗时和错误信息，异步写库不阻塞业务；支持详情弹窗、按条件查询、导出和清空。高频轮询接口不写操作日志。

### 开发工具

- **代码生成**：读取当前库的业务表（自动排除系统表和已导入表），导入为生成配置，自动识别主键、自增、Go 类型、查询方式和表单类型；可配置列注释、是否必填/列表/查询/表单/排序。支持单表、主子表、树形三类模板，预览生成的 Go model/dao/api、路由片段和前端 api/页面代码，打包 zip 下载。生成的模块在 `init()` 里自注册路由和菜单，重启后端即可访问。
- **接口文档**：从 Gin 路由表自动枚举全部 `/api` 接口，按模块分组展示请求方法、地址、名称和权限字符，支持关键词搜索，新增路由自动收录。
- **文件上传**：multipart 通用上传接口，本地按日期目录存储，UUID 重命名防冲突；扩展名白名单（图片、PDF、Office、文本、压缩包、音视频），单文件上限 10MB；SVG/HTML 等危险类型以附件形式返回，不内嵌渲染。前端演示页支持拖拽上传、结果回显和图片预览。
- **Excel 导入导出**：用户列表按当前搜索条件导出，导入按表头匹配字段，主键存在且允许更新时覆盖；主子表导出两张 sheet。

### 流程审批

自研轻量状态机，没有接 Flowable/Activiti，覆盖审批的完整生命周期。

- **流程定义**：AntV X6 画布编排，节点拖拽调整顺序，点击节点配置审批人类型（指定人员、按角色、发起人部门负责人）、或签/会签、驳回策略、是否允许加签、节点超时；支持添加/删除节点、自动排列、流程启停和删除（有申请记录的流程禁删）。
- **流转**：发起申请（可填抄送人）、通过、驳回、重新提交；支持转办、并行加签、抄送（已读标记）、撤回、催办、管理员终止。
- **权限**：实例详情仅申请人、当前节点处理人（含审批、抄送、被转办人）或管理员可见；部门负责人路由使用发起时的部门快照。
- **时间线**：每次操作写记录，前端按时间线展示发起、通过、驳回、转办、加签、撤回、终止、催办、重新提交、抄送、退回上一节点。
- 内置两个示例流程：通用审批（一级指定管理员审批）和请假审批（部门负责人 → 管理员终审）。

### 移动端

`app/` 目录是一套独立的 uni-app 工程，复用后端同一套接口：

- 验证码登录，支持记住账号；`getInfo` 拉取用户、角色和权限，`getRouters` 拉菜单树。
- 首页菜单宫格、菜单下钻页、我的页面（用户信息、角色权限、退出登录），权限用 `hasPerm` 判断。
- 登录态守卫：无令牌跳登录页，请求 401 统一清理并重定向；令牌键 `Admin-Token` 与 Web 端一致。
- 可编译 H5、微信小程序和 App；真机联调时把请求基址改成电脑局域网 IP。

### 界面与体验

- 九种布局模式（详见下文「布局风格」一节）。
- 主题色提供预设色板和取色器，选中后自动生成 Element Plus 色阶 CSS 变量；侧边栏支持深色/浅色。
- 多页签：访问过的页面自动生成标签，支持关闭、右键刷新/关闭当前/关闭其他/关闭全部；页面 keep-alive 缓存，切换页签保留搜索条件和列表状态，关闭标签同步清缓存。
- 菜单搜索：导航栏放大镜或 `Ctrl+K` 弹出，扁平化全部菜单，实时过滤，上下键选择回车直达。
- 面包屑按路由层级生成，父级可点击；路由切换带淡入位移动画和顶部进度条。
- 全屏切换、侧边栏折叠、组件尺寸（大/默认/小）、中英文国际化。
- 窄屏响应式：顶栏自动换行、侧边栏转抽屉、表格高度自适应、弹窗限宽防溢出。

## 快速开始

### 环境要求

- Go 1.27+
- Node.js 18+（推荐 20/22），npm
- 数据库：用 SQLite 不需要额外安装；用 MySQL/PostgreSQL 需要先准备好数据库实例

### 启动后端

```bash
cd backend
go run .
```

后端默认监听 `8080` 端口。首次启动自动建表并写入种子数据（部门、字典、岗位、参数、菜单、admin 账号），之后每次启动会幂等补齐缺失的种子，不会覆盖已修改的数据。

### 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认监听 `5173`，开发环境通过 Vite 代理把 `/api` 和 `/uploads` 转发到 `8080`。浏览器访问 <http://localhost:5173>。

### 启动移动端（H5 预览）

```bash
cd app
npm install
npm run dev:h5
```

### 默认账号

| 账号 | 密码 | 说明 |
| --- | --- | --- |
| admin | admin123 | 内置超级管理员，拥有全部权限 |

登录需要图形验证码；在 `backend/config.yaml` 里把 `captcha.enabled` 改成 `false` 可关闭。

### 一键脚本（Windows）

仓库根目录提供了两个批处理：

- `start.bat`：分别在独立窗口启动后端和前端
- `stop.bat`：按端口停止前后端服务

### 打包成单文件后端

前端先构建：

```bash
cd frontend
npm run build
```

把 `frontend/dist` 复制到 `backend/web/dist`（`embed.go` 通过 `go:embed` 内嵌），再在 `backend` 目录下 `go build`，产出的二进制同时提供 API 和 Web 界面。

## 目录结构

```
go-admin/
├── backend/                  # Go 后端
│   ├── main.go               # 入口：自动建表、种子数据、优雅退出
│   ├── config.yaml           # 服务、JWT、验证码、数据库配置
│   ├── config/                # 配置加载
│   ├── model/                 # 数据模型
│   ├── dao/                   # 数据访问层（GORM + 幂等种子）
│   ├── middleware/            # CORS、JWT 鉴权、权限校验、操作日志、接口权限
│   ├── api/                  # HTTP 接口层
│   ├── router/               # 路由注册与权限标注
│   ├── scheduler/            # 定时任务调度器与执行器
│   ├── templates/            # 代码生成模板
│   ├── genreg/               # 生成模块的路由自注册
│   ├── cmd/migrate/          # 独立迁移命令
│   ├── scripts/              # PostgreSQL 注释、账号、备份、慢查询脚本
│   └── utils/                # JWT、密码、响应、验证码、限流、Excel
├── frontend/                 # Vue 3 前端
│   └── src/
│       ├── api/              # 接口封装
│       ├── store/            # Pinia：用户权限、布局设置
│       ├── router/           # 静态路由 + 动态路由构建
│       ├── directive/        # v-hasPermi、v-table-height 指令
│       ├── layout/           # 布局框架与各布局组件
│       ├── views/            # 各业务页面
│       ├── components/       # 共享组件（导入弹窗、列设置、图标选择器等）
│       └── utils/            # 请求封装、主题色、时间格式化等
├── app/                      # uni-app 移动端
└── docs/                     # 设计与加固记录、截图素材
```

## 布局风格

布局模式在右侧设置抽屉的「布局模式」里切换，共有九种。下面每种都附了实际运行截图。

### 1. 左侧边栏（side，默认）

最常见的后台布局：左侧深色侧边栏放完整菜单，可折叠成 60px 的图标条；右侧上方是导航栏（折叠按钮、菜单搜索、全屏、布局配置、用户下拉），下方是内容区，内容区顶部带面包屑。

![左侧边栏布局](docs/images/layout-side.png)

### 2. 右侧边栏（rightSide）

左侧边栏的镜像：完整菜单放在屏幕右侧，内容区在左，导航栏位置不变。适合习惯右手侧操作或需要把主操作空间靠左的场景。

![右侧边栏布局](docs/images/layout-right-side.png)

### 3. 顶部导航（top）

没有侧边栏，Logo、一级目录菜单、功能区和用户信息排在一条通栏顶栏里；子菜单 hover 弹出层展开，当前目录保持高亮。内容区顶部单独一行面包屑，适合菜单层级不多、希望主区域更宽的后台。

![顶部导航布局](docs/images/layout-top.png)

### 4. 底部导航（bottom）

顶部导航的上下镜像：导航条固定在页面底部，内容区在上方。菜单、搜索、全屏和账号入口都收在底部栏里。

![底部导航布局](docs/images/layout-bottom.png)

### 5. 混合布局（mix）

顶栏放一级目录，左侧边栏只显示当前目录下的二级菜单，按当前路由第一段匹配；顶栏和侧边栏都可以折叠。窄窗口下侧边栏收起为抽屉，避免挤压主内容区。兼顾顶部导航的横向空间和侧边栏的层级展开。

![混合布局](docs/images/layout-mix.png)

### 6. 转盘（dial）

内容区全屏铺开，菜单收在右上角一个四分之一圆的转盘里。收起态是一个小圆钮，hover 或点击后沿圆环展开：一级目录排在内圈，选中目录的子菜单在外圈展开；转盘外的工具条补上搜索、全屏和账号入口，点击转盘以外区域整圈收起。

![转盘布局（展开态）](docs/images/layout-dial-open.png)

### 7. 桌面（desktop）

仿操作系统桌面的布局：没有传统顶栏和侧边栏，整屏是桌面画布，各业务页面作为可拖拽、可缩放、可最大化的窗口层叠摆放；左侧是应用图标墙，底部是任务栏（开始菜单、当前窗口、托盘区、时间和账号）。画布右键有快捷菜单：回到首页、全部最小化、重新排列窗口、更换桌面背景、刷新。窗口尺寸和状态按路由持久化，切换窗口不会卸载组件、不丢页面状态。

![桌面布局](docs/images/layout-desktop.png)

### 8. Dock 坞（dock）

内容区全屏，底部中央悬浮一条 macOS 风格的应用坞。图标 hover 放大并显示名称，有页面打开的目录下方带运行指示点；坞内还有启动台（半透明全菜单网格）、设置和账号入口，鼠标移开后坞自动沉底。

![Dock 坞布局](docs/images/layout-dock.png)

### 9. 垂直页签（vtab）

左侧是一条垂直页签栏，风格类似浏览器侧栏标签：访问过的页面竖向排列，点击切换、hover 出现关闭按钮；右侧是内容区。适合打开页面多、习惯用标签来回切换的使用方式。

![垂直页签布局](docs/images/layout-vtab.png)

## 数据库

| 能力 | SQLite（默认） | MySQL | PostgreSQL |
| --- | --- | --- | --- |
| 驱动 | `glebarez/sqlite`，纯 Go，无需 CGO/gcc | `gorm.io/driver/mysql` | `gorm.io/driver/postgres` |
| 建表迁移 | AutoMigrate | 同左 | 同左 |
| 种子数据 | 首次启动写入 | 同左 | 同左 |
| 代码生成 | 读 `sqlite_master` / PRAGMA | 读 information_schema | 读 `pg_tables` / `pg_attribute` |
| 数据监控 | 版本、文件大小、日志模式、表行数 | 版本、引擎、表行数 | 版本、库大小、连接数、命中率、表行数 |

切换方式：编辑 `backend/config.yaml` 的 `database` 段，把 `type` 改成 `sqlite` / `mysql` / `postgres` 并填好连接参数，重启后端即可。当前仓库配置示例指向本地 PostgreSQL 容器（库名 `go_admin`）；原有 SQLite 文件保留在 `backend/data/go-admin.db`，把类型改回 `sqlite` 即可继续使用。

## 权限模型

```
sys_user ──< sys_user_role >── sys_role ──< sys_role_menu >── sys_menu
 用户表        用户角色关联        角色表       角色菜单关联          菜单表
    │              │                                      ├─ 目录(M)  path/icon
    │              └──< sys_role_dept >（数据权限自定义部门）├─ 菜单(C)  路由/组件/权限标识
    ├─ sys_user_post（用户岗位多对多）                      └─ 按钮(F)  perms 权限字符
    └─ sys_dept（用户归属部门）      sys_api（接口权限：方法 + 路径 + perms + 启停）
```

权限字符示例：`system:user:list`、`system:user:add`、`system:dept:remove`、`monitor:job:run`、`tool:gen:preview`、`flow:apply:terminate`。后端在路由注册时通过 `CheckPerm("system:user:add")` 标注，前端按钮上写 `v-hasPermi="['system:user:add']"`。

## 主要配置

`backend/config.yaml` 常用项：

```yaml
server:
  port: 8080
  mode: debug            # debug / release

jwt:
  secret: "go-admin-jwt-secret-2026-change-me"   # 生产环境务必修改，也可用环境变量 GO_ADMIN_JWT_SECRET 覆盖
  expire: 480            # 过期时间（分钟）

captcha:
  enabled: true          # 登录验证码开关

database:
  type: postgres         # sqlite / mysql / postgres
  host: "127.0.0.1"
  port: 5432
  user: "postgres"
  password: "root"
  dbname: "go_admin"
  timezone: "Asia/Shanghai"
  maxOpenConns: 25
  maxIdleConns: 10
  autoMigrate: true      # 启动时自动建表/迁移
  seedData: true         # 启动时写入/补齐种子数据

logs:
  retentionDays: 90      # 日志保留天数，0 表示不自动清理
```

## 测试与校验

```bash
# 后端：编译、vet、单测
cd backend
go build ./...
go vet ./...
go test ./dao/ ./router/ ./utils/ ./middleware/ -count=1

# 前端：类型检查、lint、单测、构建
cd frontend
npm run type-check
npm run lint
npm run test
npm run build
```

## 一些说明

- 种子数据里的密码、JWT 密钥和数据库连接信息都是本地演示用的，上线前请替换。
- 表结构不建主外键约束：关联关系由 dao 层在删除时保证清理，需要强约束时可在迁移脚本里补。
- 后端接口路由、响应结构的快照测试和验收记录见 `docs/architecture-hardening-plan.md`；开发过程中的逐项修复与验证记录见 `docs/code-style-cleanup.md`。
- 项目目前定位为单节点、单库的中小型后台脚手架；多实例部署需要补 Redis 共享会话、分布式调度选主和共享存储，这部分在加固计划文档里有分阶段安排。

## 许可证

本项目基于 [MIT License](LICENSE) 发布，允许免费用于个人和商业场景：可以使用、复制、修改、合并、发布、分发、再授权和销售副本，条件是在软件副本或实质性部分中保留版权声明和许可证文本。软件按"原样"提供，不附带任何明示或默示担保，完整条款见 [LICENSE](LICENSE)。
