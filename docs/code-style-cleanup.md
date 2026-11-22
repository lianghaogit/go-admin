# 代码规范整理变更摘要

> 批次：2026-10-30 代码规范全面整理（后端 Go + 前端 Vue/TS）
> 二轮深检：2026-10-30 追加（后端错误处理/占位代码/退出流程/配置安全）
> 原则：只做格式、命名（安全范围）、注释与工具链；**未改变任何业务逻辑、接口签名、路由/API 路径、组件 props/emit、localStorage key（go-admin-settings）、token 键（Admin-Token）**。

## 〇、二轮深检修复（本轮追加）

| # | 问题 | 位置 | 修复 |
| --- | --- | --- | --- |
| 1 | **无效引用占位 hack**：`var _ = regexp.MustCompile // 占位，防止误删引用`，regexp 实际从未使用 | `dao/gen_dao.go` | 删除占位行与 `"regexp"` import |
| 2 | **GetInfo 静默降权**：非 admin 用户角色/权限查询错误被 `_ =` 丢弃，失败时返回空权限（用户登录后无任何权限且无提示） | `api/auth_api.go:111` | 查询失败记录日志并返回 500，不再静默降权 |
| 3 | **裸 goroutine 无 recover**：登录日志、操作日志异步写库 goroutine 若 panic 会拖垮整个进程 | `api/auth_api.go` `logLoginLog`、`middleware/operlog.go` | defer recover + log 捕获 |
| 4 | **JWT 密钥无环境变量覆盖**：密钥仅配置在 config.yaml | `config/config.go`、`config.yaml` | 支持 `GO_ADMIN_JWT_SECRET` 环境变量覆盖（优先于配置文件），注释更新 |
| 5 | **无优雅退出**：Ctrl+C 时调度器 goroutine 不停止、SQLite 连接不关闭（进程直接退出） | `main.go` | `http.Server` + 信号监听（SIGINT/SIGTERM）→ `Shutdown` → `scheduler.Stop` → 关闭 DB |
| 6 | **.gitignore 不全**：缺上传目录与 IDE 目录 | 根目录 `.gitignore` | 补 `backend/uploads/`、`.vscode/`、`.idea/` |

**二轮审计保留说明（不修，有意为之）**：
- `_ = dao.CreateOnline` / `DeleteOnlineByTokenHash` / `InsertJobLog` 等非关键路径错误丢弃：会话/日志写入失败不影响主业务响应，为 RuoYi 生态惯例
- `views/system/user|role/index.vue`（20KB+）为 RuoYi 风格 CRUD 聚合页：抽取组合式 API 重构收益小、回归风险大，暂不拆分
- `dao/seed.go`（32KB）为种子数据集中管理，按表分块 + 幂等设计，可读性尚可

## 〇.1、三轮优化（追加）

### 后端
| 项 | 位置 | 说明 |
| --- | --- | --- |
| CORS 生产收紧 | `middleware/cors.go` | 支持 `GO_ADMIN_ALLOW_ORIGIN` 环境变量（逗号分隔多来源白名单），未设置时保持 `*`（开发默认）；命中来源时回 `Vary: Origin` |

### 前端 TS 类型收紧（渐进策略推进）
- 新建 `src/types/api.ts`：通用响应类型 `PageResult<T>` / `DataResult<T>` / `PlainResult`
- `api/user.ts`：定义 `UserItem` 接口（userId/deptId/userName/nickName/avatar/phone/email/status/remark/createTime/dept/posts/roles），list/get/add/update/delete/changeStatus/authRole/resetPwd 全部标注返回类型
- `views/system/user/index.vue`：`userList = ref<UserItem[]>([])`
- **类型标注暴露并修复真实 bug**：`openAuthRole` 中 `getUser(row.userId)` 的响应体为 `{ code, msg, data: {...} }`，原代码误用 `detail.roles`（恒为 undefined），导致**分配角色弹窗的用户当前角色勾选一直为空** → 改为 `detail.data?.roles || []`
- `dashboard/index.vue`：`StatCard` 接口（value: number | string），解决统计卡片占位符与数字赋值混用
- 验证：`vue-tsc` 0 错误

### 工程便利
- 新增 **`start.bat`**（一键启动后端 + 前端两个独立窗口，URL 提示）与 **`stop.bat`**（按 8080/5173 端口杀进程），CRLF 编码，双击即用

### 三轮回归
- 前端 `eslint` 0 error / `vue-tsc` 0 错 / `build` 通过；后端 `go build` / `go vet` 全绿

## 〇.2、个人中心（头像/基本信息）修复与设计进阶（追加）

### 修复的真实 bug（浏览器实证）
| # | bug | 影响 | 修复 |
| --- | --- | --- | --- |
| 1 | `fetchProfile` 用 `res.nickName` 取值，而响应体为 `{code,msg,data}` | **个人资料表单全部字段加载为空**（昵称/手机号/邮箱/备注/创建时间，页面只显示 placeholder） | 改用 `res.data.xxx` |
| 2 | 头像上传后 `res.fileName || res.url` 取到纯文件名（`uuid.ext` 无路径） | 头像 URL 存错，刷新后头像无法显示 | 改为 `res.url \|\| res.fileName` |
| 3 | vite proxy 缺 `/uploads` | dev 环境头像请求 5173 直接 404（生产后端静态服务正常） | vite.config.ts 增加 `/uploads` → 8080 代理 |
| 4 | `updateAvatar` 接受任意字符串 | 可存外链/任意路径（存储型安全面扩大） | 后端仅允许 `/uploads/` 前缀，其余返回"头像地址不合法" |
| 5 | 修改密码成功后 `router.push('/login')` 为死代码 | 无害冗余 | 移除（`userStore.logout()` 内部已整页跳转） |

### 设计进阶（个人中心页面重构）
- 头像区：96px 大尺寸 + **hover 遮罩"更换头像"**（点击遮罩/头像直接唤起文件选择），无头像时显示昵称首字
- 信息展示：昵称 + 角色标签 + **所属部门**（`dept.deptName`，此前未展示）
- 表单：输入框加 clearable、保存按钮旁新增"重置"（回到已保存值）、密码新/确认提示区分文案
- 上传：错误兜底 try/catch、上传后立即清空 file input（支持重复选同一文件）

### 验证（全链路实证）
- 上传接口返回 `{ fileName, url: /uploads/… }` → `updateAvatar` 存库成功 → 浏览器 `5173/uploads/...` 经代理返回图片 → 头像真实渲染（不再 fallback 首字）
- 外部 URL（`https://evil.example/x.png`）提交 → 后端返回 `{"code":500,"msg":"头像地址不合法"}` 拒绝
- 回归：前端 `eslint`/`vue-tsc`/`build` 全绿；后端 `go build` 全绿
- 测试数据已清理（admin 头像恢复为空、测试上传文件已删）

## 〇.3、全站设计进阶（追加：圆角 ≤ 2px 收敛 + 全局视觉细节）

### 圆角体系收敛（所有组件圆角 ≤ 2px）
- **全局令牌**（`frontend/src/styles/index.css`）：覆盖 Element Plus 圆角变量 `--el-border-radius-base/small/round/large` 全部为 2px
- **兜底规则**：对 25+ 组件类（按钮/卡片/输入/下拉/弹窗/表格/分页/标签/树/消息/通知/进度条等）强制 `border-radius: 2px !important`，覆盖组件内部固定圆角
- **页面硬编码收敛**：20 处 >2px 圆角替换为 2px —— 桌面模拟（DesktopArea 7 处 6/8/10px、DesktopWindow 4 处 6/4px、StartMenu 2 处 10/6px）、TagsView（3/4px）、登录页验证码、图标选择器、菜单搜索、设置抽屉、缓存监控（8/4px）、上传预览、代码生成、角色分配树、弹窗头部
- **圆形控件保留**：头像 / 开关滑块 / 单选点 / 圆形按钮 / 标签关闭钮维持圆形（语义为"圆"而非"圆角"）

### 全局视觉细节提升
| 项 | 效果 |
| --- | --- |
| 卡片 | 细边框（`--go-border-color`）+ 极轻阴影，在浅灰主区上增强层级 |
| 表格表头 | 浅底（`--go-table-header-bg` 亮/暗双值）+ 中粗字重 600 |
| 侧边栏菜单激活 | 左侧 2px 主色指示条 + 主色淡渐变背景（展开态与折叠 popup 均生效） |
| 弹窗 | header 加细分隔线，标题层级更明确 |
| 输入控件 | focus 描边统一主色 1px |

### 验证（浏览器 + 构建实证）
- 用户管理页实测计算值：card/button/input/table/pager/dialog 均 `2px`；avatar `50%`（圆形保留）；弹窗 header 分割线 1px；表头 `rgb(245,247,250)` + weight 600；卡片边框 1px `#e4e7ed` + 轻阴影
- 侧边栏激活项：展开态与折叠 popup 均为 `border-left: 2px solid #409eff` + 渐变背景
- 全站 `src/` 复扫 `border-radius > 2px` 硬编码：0 命中
- 回归：`eslint` 0 / `vue-tsc` 0 / `build` 通过

## 〇.4、设计再升级：紧凑布局 + 高级质感（追加）

### 布局紧凑化（紧凑但不拥挤）
| 区域 | 原 | 现 |
| --- | --- | --- |
| 侧边栏宽度 | 210px / 折叠 64px | **200px / 60px** |
| 顶栏高度（top/desktop） | 50px | **46px** |
| 面包屑条高度 | 40px | **36px**（横向 padding 24→18px） |
| 页签高度 / 标签项 | 34px / 24px | **32px / 22px** |
| 侧边栏 Logo | 50px | 46px（+白色微渐变） |
| 菜单项高度 | 56px | **46px**（垂直菜单含折叠 popup） |
| 表格单元格 | 8px 垂直 | **6px** |
| 卡片内边距 | 20px | **body 16/18px · header 12/18px** |
| 页面容器 | 20px | **16px**（全局统一 `html .app-container`） |
| 弹窗内边距 | EP 默认 | **body 14/20px · footer 8/20/16px** |
| 分页按钮 | 32px | 26px |
| inline 表单项间距 | 32px | **16px**（底部 12px） |

### 高级质感提升
| 项 | 效果 |
| --- | --- |
| 主区背景 | `#f0f2f5` → `#f6f8fc`，更干净柔和 |
| 首页欢迎卡 | 主色浅渐变横幅 + 右侧装饰圆环 |
| 首页统计卡 | 44px 主色渐变图标块（白色内高光+投影）+ 数值 24px/700，hover 图标上浮 |
| 按钮 | 全色系内高光渐变 + 轻投影，hover 增强 |
| 表格行 hover | 淡主色（亮/暗双值） |
| 卡片 hover | 微阴影提升（is-hover-shadow） |
| 滚动条 | 6px 细条 + 柔和 thumb |

### 验证（浏览器 + 构建实证）
- 展开态侧边栏实测 **200px**、折叠态 **60px**；表格 td/th 垂直 **6px**；表头浅底 `rgb(245,247,250)`；按钮内高光+投影；页面容器 **16px**；统计图标块主色底+白色渐变+投影；欢迎卡渐变横幅
- 窄屏（<768px）响应式正常（统计卡单列/图标上排）
- 回归：`eslint` 0 / `vue-tsc` 0 / `build` 通过

## 〇.5、双主题适配 + 语义化配色体系（追加）

### 问题扫描
大量页面 scoped 样式硬编码：深色文本 `#303133`（31 处/11 文件）、浅色底 `#fff/#f5f7fa/#fafafa/#f0f0f0`、浅色边框 `#e4e7ed/#ebeef5/#dcdfe6`——暗黑主题下保持硬值，导致刺眼/不可读/突兀。

### 修复：语义化配色变量（浅/暗双值）
`styles/index.css` 追加 `--go-text-primary/regular/secondary`、`--go-fill-light`、`--go-border-regular`、`--go-bg-page`：

| 变量 | 浅色 | 暗黑 |
| --- | --- | --- |
| `--go-text-primary` | #303133 | #e5eaf3 |
| `--go-text-regular` | #606266 | #cfd3dc |
| `--go-text-secondary` | #909399 | #a3a6ad |
| `--go-fill-light` | #f5f7fa | #1d1e1f |
| `--go-border-regular` | → `var(--go-border-color)` 别名（暗黑自动 #2b2b2b） | 同左 |
| `--go-bg-page` | #ffffff | #141414 |

批量替换 **17 个文件 47 处**硬编码：dashboard 12、monitor/cache 5、db 3、server 3、job 1、notice 3、operlog 1、role 1、api-doc 2、upload 1、403 2、profile 1、SearchMenu 4、SettingsDrawer 2、TagsView 3、IconSelect 1、ColumnSetting 1。登录页手工适配（页面/表单面板/标题/副标题变量化）。

### 暗黑适配浏览器实证（计算值）
| 页面 | 关键值 | 结论 |
| --- | --- | --- |
| 首页 | 卡片深色、数字 `#e5eaf3`、欢迎渐变保留 | ✓ |
| 用户管理 | 表头 `#1d1e1f`/文字 `#a3a6ad`、输入框深色、按钮主色 | ✓ |
| 个人中心 | 卡片 `#1d1e1f`、label `#cfd3dc`、描述深色 | ✓ |
| 登录页 | 页面/面板/输入框 `#141414`、标题浅色、品牌渐变保留 | ✓ |
| 全局 | `--el-bg-color=#141414`、侧边栏 #304156 双主题一致、顶栏/页签 `#141414` | ✓ |

### 配色体系评估（高级感）
- 主色 `#409EFF`（品牌蓝）与侧边栏 `#304156`、暗黑 `#1d1e1f` 协调；文本三档、背景三档、边框、状态五色均由变量承载，浅/暗一键切换
- 统计卡 7 色数据分类色 + 渐变图标块，暗黑下保持饱和度
- 可选加分项（未做）：主色微调为靛蓝系提升辨识度——涉及 EP 主色衍生变量与大量 `#409eff` 硬编码，改动大，留作后续可选项

- 回归：`eslint` 0 / `vue-tsc` 0 / `build` 通过

## 〇.6、数据库多方言支持（SQLite / MySQL / PostgreSQL，追加）

- **配置**：`config.yaml` `database.type`（sqlite 默认 / mysql / postgres）+ host/port/user/password/dbname；旧配置无 type 自动回退 sqlite，零破坏
- **接入**：`dao/db.go` 按方言选择驱动（glebarez/sqlite / gorm mysql / gorm postgres），MySQL/PG 加连接池（MaxOpen 20 / MaxIdle 5），建表、种子、接口规则加载共用 `afterOpen()`
- **方言适配三处**：
  1. `db.go fixApiCreateTimes`：`datetime('now')` / `NOW()` / `now()` 回填 sys_api 创建时间
  2. `gen_dao.go`：表列表与列元数据——sqlite_master+PRAGMA / information_schema.tables+columns / pg_tables+obj_description+pg_index，统一映射为 DbTableMeta/ColumnMeta
  3. `api/db_api.go` 数据监控：版本、引擎/日志模式、表列表按方言
- **验证**：`go build ./...` 0、`go vet ./...` 0；SQLite 默认配置全量冒烟（登录、首页 7 统计、代码生成页、数据监控 21 表）浏览器实证正常；MySQL/PG 实连冒烟需目标服务（本机未安装），所用 SQL 均为标准 information_schema/pg_* 查询
- **切换方式**：改 `config.yaml` 的 type 与连接参数后重启后端即可，首次启动自动建表 + 种子

## 〇.7、P0 / P2 质量加固（本轮追加：WAL · 登录限流 · 自动化测试 · 代码生成 TS · 构建优化 · 日志导出过滤 · i18n）

### P0-1：SQLite 切 WAL（#63）
- `dao/db.go initSQLite`：gorm.Open 后执行 `PRAGMA journal_mode=WAL`（取返回值确认）+ `PRAGMA busy_timeout=5000`，失败即报错返回
- 收益：读写并发下 WAL 模式显著降低锁竞争；数据监控页「日志模式=wal」、router 测试 `TestDbInfo` 同口径锁住
- 实测：运行实例数据监控页显示 wal ✓

### P0-2：登录防爆破限流（#64）
- 新建 `backend/utils/limiter.go`：内存 map+mutex，常量 **账号 5 次 / IP 20 次 / 锁 15 分钟**，key 前缀 `u:` / `i:`；暴露 `LoginUserAllowed(username, ip)`（双维度均放行才放行）、`LoginUserFail(username, ip)`（双维度各自累计，返回剩余锁定时长）、`LoginUserReset(username)`（登录成功清账号维度，IP 维度保留防换账号爆破）
- `api/auth_api.go Login`：验证码校验前插 allow 判断（锁定文案「登录失败次数过多，请稍后再试」）；验证码错 / 密码错 / 账号停用三个失败分支各调 `LoginUserFail`；成功后 `LoginUserReset`
- **测试暴露并修复的两个真实 bug**（勿再改回）：
  1. `fail()` 原实现 `if !ok || now.After(e.lockUntil)` 对零值 lockUntil 恒真 → 计数永远到不了阈值
  2. `allow()` 对「有失败记录但未锁定」（lockUntil 零值）的条目 `now.Before(零值)` 恒 false → 误判已过期 → 每次请求进 Login 先删计数 → 限流形同虚设
- 实测：生产实例连打 5 次错误登录，第 6 次返回锁定文案 ✓

### P0-3：自动化测试（#65）
- 后端 `backend/router/router_test.go` **13 用例**：TestMain 用 `os.TempDir()` 全新 SQLite + 种子，直连 `config.Cfg` / `dao.InitDB(config.Cfg.Database)` / `InitRouter()`；`doJSON` 默认 IP 198.51.100.7；覆盖验证码 / 登录成功 / 密码错 / **连错 5 次锁定（lockuser 专用账号 + 独立 IP）** / 鉴权 / 用户列表 / 角色 / 代码生成列表 / 数据监控（断言 journalMode=wal）/ 登录成功重置限流 / 工具函数
- `backend/utils/limiter_test.go` 2 用例；运行 `& "E:\Go\bin\go.exe" test ./dao/ ./router/ ./utils/ -count=1` 全绿
- 前端 vitest：`vitest@^2.1.9`（**vitest@5 与 vite 5 peer 冲突装不上**）+ jsdom + @vue/test-utils；`vitest.config.ts`（jsdom/globals）、`tests/format.test.ts`（parseTime 8 用例）、`tests/auth.test.ts`（token 4 用例）；`npm run test` 12 用例全过；package.json 加 `test` / `test:watch`

### P2-1：代码生成增强（TS 化，与全站 TS 对齐，#66）
- `dao/gen_dao.go`：`GenField` 增派生 `TsType`（int/uint/float64→number、bool→boolean、其余→string）；`GenZipFiles` 前端 API 输出 `.js` → `.ts`
- 重写 `backend/templates/vue_api.tmpl`：导出 `Query` / `Row` 两个 interface（Query 含 pageNum/pageSize）+ 各函数参数类型
- 改 `vue_index.tmpl`：`<script setup lang="ts">`、`ref<Row[]>`、`reactive<Query>`、`openDialog(row?: Row)`、`handleDelete(row: Row)`、import type
- `dao/gen_template_test.go` 渲染断言（go test 工作目录是 dao/，需先 `os.Chdir` 到上级 backend/）

### P2-2：前端构建优化（#67）
- `vite-plugin-compression@^0.5.1`：构建产物预压缩 gzip（.gz）+ brotli（.br），threshold 10KB；Nginx 可直接伺服省运行时压缩
- `build.esbuild.drop: ['console','debugger']`（**放 build 作用域**，仅生产移除、dev 保留）；echarts 独立 manualChunk；`chunkSizeWarningLimit: 1500`（element-plus/echarts 分包较大属已知）
- 构建实测（~24s）：index 206KB（gzip 54/br 41）、vendor 168（64/56）、echarts 533（183/151）、element-plus 1090（342/273）、css 400（55/42）；压缩插件日志路径显示 `dist/D:/Golang/...` 属显示 bug，实测产物正常落在 `dist/assets/*.gz/.br`

### P2-3：日志审计导出增强（#68）
- **登录日志补查询过滤**：`dao.LoginLogQuery{PageNum,PageSize,Username,IP,Status,BeginTime,EndTime}`，`ListLoginLogs(q)` 支持用户名/登录IP/状态 LIKE 过滤 + 时间范围；列表接口与导出接口同口径传参
- **操作日志补时间范围**：`OperLogQuery` 加 BeginTime/EndTime（oper_time 过滤）
- 新增 `dao/time_query.go`：纯日期（YYYY-MM-DD）自动补 `00:00:00` / `23:59:59`，与 sqlite/mysql/pg datetime 存储格式可比较
- 前端：登录日志页新增搜索表单（用户名/登录IP/状态/日期范围），导出携带当前过滤条件；操作日志页补日期范围
- 实测：按用户名过滤登录日志只返回命中行（共 5 条）✓

### P2-4：i18n 国际化（#69）
- 接入 `vue-i18n@^9`（Composition API 模式），`src/i18n/index.ts` + `locales/zh.ts` / `en.ts`；localStorage `go-admin-language` 持久化
- Element Plus 组件语言随切换联动（App.vue `el-config-provider` locale computed：zh-cn / en）
- 设置抽屉顶部新增「语言」切换（中文 / English）；`store/settings.ts` 增 `language` 字段 + `setLanguage` action
- 已国际化：登录页（品牌区/表单/校验消息）、顶栏（折叠/搜索/全屏/布局配置 title + 面包屑首页）、用户下拉（个人中心/退出登录）、面包屑菜单名（`i18n/translate.ts` 中文标题→key 映射，命中才翻译、未收录保持原文）
- **渐进策略**：各列表页内部文案（表格列/按钮/弹窗）为高频低成本跟进项，框架就绪后可逐页替换；侧边栏菜单名由后端返回，同 breadcrumb 映射机制可扩展
- 实测：切 English 后面包屑 Home / System / Login Logs、分页 Total 5、抽屉 Layout Settings ✓

### 〇.7 追加：复查精进（2026-10-31，P0/P2 之上）
- **限流器内存回收**（`utils/limiter.go`）：原 `fails` map 只增不减 → 新增 `loginEntryIdle=30min` 常量、`failEntry.lastFail` 字段、`startGCLoop()`（init 启动，5 分钟 ticker 调 `gc()`）；`gc()` 持锁清理两类过期条目（锁定到期超 1 分钟 / 未锁定但 lastFail 超 30 分钟）；`fail()` 每次更新 lastFail
  - 单测 `TestLoginLimiterGC`：首次写法在持 `mu.Lock()` 期间调用 `LoginUserFail` 造成自死锁（utils 600s 超时 FAIL），改为「先造失败数据 → 持锁改时间戳 → 解锁后调 gc()」后 PASS
- **LIKE 通配符转义**（`dao/time_query.go`）：新增 `escapeLike()`（`\`→`\\`、`%`→`\%`、`_`→`\_`，SQLite/MySQL/PG 默认反斜杠转义）；登录日志（username/ipaddr）、操作日志（title/oper_name）、`ListGenTables`（table_name/class_name）共 6 处改为 `LIKE ? ESCAPE '\'` 参数化 + 转义
- **分页参数保底**（`dao/time_query.go`）：`normPage()` PageNum<1→1、PageSize<1→10；三个 List 函数 offset/limit 改用保底值（防非法入参负 offset）
- **死代码清理**：`gen_dao.go` BuildGenData 中 gorm 标签前的空 `if strings.HasPrefix(f.GoType, "int")...` 分支删除
- **菜单 i18n 全覆盖**（`useMenuTitle()` 从 `@/i18n/translate` 接入 11 个组件）：
  - 侧边栏：`Sidebar.vue` 首页固定项 + `SidebarItem.vue` 目录子菜单/叶子菜单
  - 顶栏：`TopNav.vue` 首页固定项 + `TopNavItem.vue` mix 平铺/目录下拉/叶子 3 处
  - 面包屑：`layout/index.vue` side 4 处 + mix/desktop 4 处 + mix/desktop 固定首页项
  - 页签：`TagsView.vue` 标题 + 右键菜单（Refresh/Close Current/Close Others/Close All）
  - 桌面：`DesktopWindow.vue` 窗口标题、`DesktopArea.vue` 任务栏窗口项 + 开始按钮 title + 右键菜单、`StartMenu.vue` 面板标题 + 目录项
  - 搜索：`SearchMenu.vue` 弹窗标题/placeholder/空态 + 结果项按「父 / 子」分段翻译（`titleParts`，原文保留供中文搜索匹配）
  - 映射补齐：`zhToKey` 增「参数配置→config」「日志管理→logmgmt」；zh/en 语言包增 layout 组（start/ctxTitle/goHome/minimizeAll/rearrange/changeBg/refresh）
- 回归：`go test ./dao/ ./router/ ./utils/ -count=1 -timeout 120s` 全绿（dao 0.385s / router 1.343s / utils 0.264s）；`npm run type-check` / `vitest 12/12` / `npm run build`（gzip+brotli）全绿
- 浏览器实测（English 态）：侧边栏 System/Users/…/Configs/Log Management、搜索弹窗「System / Users」分段英文、TagsView 右键 Refresh/Close Current、桌面窗口 Home/Users + 任务栏 + 开始菜单 System/Monitor/Tools、mix 顶栏 Home/System/Monitor/Tools + 面包屑全部英文；验证后已切回中文/侧边栏布局
- 已知限制：MySQL/PG 实连冒烟需目标数据库服务（本机无）；列表页内部文案全量 i18n 仍为渐进项；分页保底/LIKE 转义由 dao 单测覆盖（API 级回归需登录态 token，本轮未做）

## 〇.8、工作流 / 审批模块（2026-10-31，自研状态机引擎）
- 形态拍板：完整自研轻量状态机引擎（不引入 Flowable/Activite），审批人三类全要（指定人员 / 按角色 / 发起人部门负责人），一次交付前后端。
- 数据表（`model/flow.go`，`dao/db.go` AutoMigrate 追加，老库启动自动补建）：`sys_flow_def` / `sys_flow_node` / `sys_flow_instance` / `sys_flow_task`（抄送、加签同表） / `sys_flow_record`（时间线）。
- 引擎（`dao/flow_dao.go`）：发起（含抄送人）、通过/驳回（or-sign/any-sign 会签计数推进、驳回回发起人/退上一节点）、重提（从 RejectFromNodeID 恢复）、转办（原任务置 3、新建任务）、加签（type=3 并行或签）、撤回（仅申请人+运行中）、终止（管理员）、催办（写记录+判重）、超时自动催办（`StartOverdueWatcher` 每分钟扫描，main 启动）、乐观锁 `WHERE status='0'` 防重复处理、或签通过后同节点其余待办置 5。
- 状态编码：实例 1运行/2通过/3驳回/4撤销/5终止；任务 0待办/1通过/2驳回/3转出/4加签/5撤销/6已读；记录 action 1发起/2通过/3驳回/4转办/5加签/6撤回/8终止/9催办/10重新提交/11抄送。
- API（`api/flow_api.go` + `router/router.go` 23 条，均 CheckPerm）：`/api/flow/def/*`（CRUD+启停，SaveFlowDef 有运行实例拒改节点、DeleteFlowDef 有实例拒删、status!=0 停用拦截发起，列表支持 `keyword`/`status`/`category` 过滤）、`/api/flow/apply/*`（发起/撤回/重提/催办/终止，列表支持 `keyword`/`status` 过滤）、`/api/flow/todo/*`（通过/驳回/转办/加签）、`/api/flow/cc/*`（列表/已读）、`/api/flow/instance/*`（列表/详情，列表支持 `keyword`/`status` 过滤，详情含 Nodes/Tasks/Records）。管理员终止 `PUT /api/flow/apply/terminate/:id`（权限 `flow:apply:terminate`，仅运行中可终止，重复终止 400，终止写入时间线 action=8）。
- 种子（`dao/seed.go` ensureFlows，count>0 跳过）：菜单 2000-2040（流程中心+5 页面+17 按钮，role_id=1 已绑）+ 23 条 `flow:*` API 权限 + 两个示例流程（通用审批 general 一级指定管理员可加签；请假审批 leave 两级：部门负责人 → 管理员终审，节点 1 驳回退上一步）。
- 前端：`api/flow.ts` 19 个封装；`views/flow/` 共享详情组件（描述+el-steps 节点进度+任务表+el-timeline 时间线）+ 5 页面（流程定义节点编辑器/发起申请/待办审批/已办任务/抄送我的）；i18n zhToKey + zh/en 各增 6 个菜单 key；「再继续完善」补：流程定义页分类搜索框、发起申请页「终止」按钮（`flow:apply:terminate` 权限 + 终止原因弹窗）。
- 关键 bug 修复（勿回退）：
  - `flowAddRecord` 必须传事务内 db（曾用全局 DB.Create 导致 SQLITE_BUSY 每次 ~5s、单测 116s、SaveFlowDef 返回 ID=0）
  - `SaveFlowDef` 返回局部 defID 变量（曾 `return input.FlowDefID, DB.Transaction(...)` 新建恒返回 0）
  - 列表补标题：todo/done/cc 改返回 `[]FlowTaskVO`（embed FlowTask + DefName/Title/ApplicantName），`flowTasksWithInst` 批量 IN 查实例防 N+1
  - 流程更新列名 `update_time` → `updated_at`（gorm 默认列名）
  - gofmt 在 PowerShell 需数组展开（`$files=@(...)`）传 `& gofmt.exe -w`
- 「再继续优化+修 bug」轮（2026-10-31）：
  - 列表搜索补全：待办/已办/抄送三页后端 `ListFlowTodo/Done/Cc` 增加 `keyword` 过滤（标题模糊，`instance_id IN (SELECT ...)` 子查询，LIKE ESCAPE '\'），API 透传 `c.Query("keyword")`，前端三页加搜索框（搜索/重置/回车）；**同时修正历史参数名不一致**：def 页 `defName`、apply 页 `title` 均统一为 `keyword`（此前前端发 defName/title 而后端只读 keyword，页面搜索实际不生效）
  - 撤回语义收紧：`WithdrawFlowInstance` 增加「已有审批人处理（action=2/3 记录）不可撤回」校验——审批人已通过/驳回后申请人不得撤回，仅审批人未处理时可撤回
  - 防重复指派：`TransferFlowTask` / `AddSignFlowTask` 拒绝转办/加签给自己，且目标人在当前节点已有待办任务（含已加签）时拒绝，避免重复生成任务
  - 抄送净化：`StartFlowInstance` 抄送人排除申请人本人并去重（此前可抄送自己产生无意义任务）
  - 遗留疑点修复（2026-10-31 第 4 轮）：催办判重口径统一为「同实例同节点」——手动 `UrgeFlowInstance` 由 `instance_id` 改为 `(instance_id, node_id=当前节点)`，与自动超时催办一致（流转到新节点后可再次催办）；`dept_leader` 路由报错带部门名与负责人（「部门「XX」未配置负责人（leader）/ 负责人「YY」未匹配到启用用户」）；驳回退上一节点追加时间线 action=12「退回」记录（前端时间线文案同步）；核查「GORM 复用 q 无 Session 隔离」疑点：全部 List 均为 `q = q.Where(...)` 克隆链写法，不修改 DB 实例，无实际污染，无需改动
  - 权限与一致性修复（2026-10-31 第 5 轮）：①**实例详情横向越权（IDOR）**——`GetFlowInstanceDetail` 增加归属校验：仅申请人本人 / 实例任务任一 assignee（含审批、抄送、被转办人）/ 管理员（`IsAdmin` 豁免）可查看，其余报「无权查看该申请」；②**催办越权**——`UrgeFlowInstance` 增加归属校验：仅申请人本人 / 当前节点待办任务审批人 / 管理员可催办，其余报「无权催办该申请」；③**部门快照一致性**——`flowCreateNodeTasks` 路由时以实例 `ApplicantDeptID` 快照覆盖申请人当前部门（推进/退回/重新提交均生效），避免申请人中途换部门导致 dept_leader 节点审批路由漂移；④**撤回语义收紧**——审批人已转办/加签（action 4/5）后申请人不可再撤回（原仅拦截 2/3）；⑤**SQLITE_BUSY 防御**——手动催办与超时自动催办改在事务内写催办记录（遵守「flowAddRecord 必须传事务内 db」约定）
  - 验证：新增 3 个引擎单测（`TestFlowInstanceDetailScope` 详情权限、`TestFlowUrgeScope` 催办权限、`TestFlowWithdrawAfterTransfer` 转办后撤回被拒）；`go test ./dao/ ./router/ ./utils/` 全绿；HTTP 冒烟（`flow_smoke5.py`）15 项断言全 PASS（无关用户看详情/催办被拒、申请人/抄送人/被转办人可见、转办后撤回应被拒）；冒烟数据与临时用户已清理；captcha 恢复 true 正式重启
  - 验证：新增 2 个引擎单测（`TestFlowRejectBackRecord` 退回标记、`TestFlowUrgeDedupByNode` 催办判重）；`go test ./dao/ ./router/ ./utils/` 全绿；HTTP 冒烟（`flow_smoke4.py`）17 项断言全 PASS（同节点二次催办仅 1 条记录、节点1通过后新节点可再催共 2 条、动态两级流程节点2驳回→退回节点1 运行+待办重生成+时间线 action=12「退回至「初审」重新审批」）；前端 type-check 全绿；冒烟数据已清理（含临时流程 smokeback），captcha 恢复 true 正式重启
- 验证：
  - `dao/flow_test.go` 7 个引擎单测（独立临时 SQLite TestMain）全绿；`go test ./dao/ ./router/ ./utils/ -count=1 -timeout 120s` 全量回归绿
  - HTTP 全链路冒烟（`flow_smoke.py`）38 项断言全 PASS：登录/定义列表/发起通过/两级走完/驳回重提/转办/加签/抄送已读/撤回/催办/定义 CRUD+有实例拒删+停用拦截/已办+实例列表
  - 「再继续完善」冒烟（`flow_smoke2.py`）14 项断言全 PASS：def/list 关键词+状态+分类过滤、apply/list 关键词+状态过滤、instance/list 过滤、发起→管理员终止→实例 status=5→时间线 action=8、重复终止 400
  - 前端 `npm run type-check` / `npm run build`（gzip+brotli）全绿
  - 浏览器端到端（bu plane，localhost:5173）：流程中心菜单下发、流程定义页（7 条含冒烟残留）、发起申请页发起「浏览器验收-测试」成功、待办审批通过（23→22 条）、已办页显示已通过+意见、抄送页未读/已读+详情弹窗（节点进度/任务记录/审批时间线）全部实测通过
- 交付后清理：联调期 captcha.enabled=false 已恢复 true；冒烟/验收产生的实例（编号 46-92 区间）、任务、记录及非种子流程定义（smoke/xflow/dbg/守卫测试）已全部清空，仅保留种子流程 general/leave 与 0 实例。
- 流程定义可视化改造（#81，2026-11-01，方案 A：AntV X6 画布）：
  - 技术选型：`@antv/x6@3.1.8`（MIT，可商用；对比 LogicFlow Apache-2.0 / Flowable AGPLv3+（SaaS 传染）/ Activiti Apache-2.0）。后端零改动、节点数据契约不变（`nodeOrder/nodeName/approverType/approverIds/nodeMode/canAddSign/rejectTo/timeoutHours/remark`，approverIds 逗号分隔字符串）。
  - 实现（`frontend/src/views/flow/def/`）：`FlowCanvas.vue` 独立画布组件（`props.initNodes` 数据驱动，onMounted 等容器尺寸后自动渲染，消除 openDialog 等 canvasRef 的竞态）；`index.vue` 弹窗内嵌画布 + 节点配置弹窗（节点名/审批人类型/多选用户/部门负责人只读文案/或签会签/驳回策略/可加签/超时/备注），保存走 `collect()` 按画布 y 坐标升序回填 nodeOrder（START/END 过滤）。
  - X6 3.x 关键适配（勿回退）：无 `dist/style.css`（CSS-in-JS）→ `import { content as x6Style } from '@antv/x6/es/style/raw'` + 幂等注入 `<style id="x6-base-style">`；节点用自定义 markup（body/label/summary 三 selector）+ attrs 文本（`labels` 是 Edge 专用，Node 不可用）；尺寸用 `getBBox()`（非 getBoundingBox）；**Graph 必须显式 `async:false`**（3.x 默认 true 走 rAF 渲染队列，后台 tab rAF 不触发 → 模型有节点但视图永不渲染，弹窗 0 节点终极根因）；`virtual:false`；`graph.findViews(ref)` 需 ref 参数；节点点击监听 pointer 事件序列。
  - 交互：选中节点高亮 + 「删除选中」（`selectedId` 响应式化，模板 `:disabled` 实时联动，原非响应式 let 导致按钮永禁用）、添加节点（默认「审批节点N」插入末位前）、自动排列、拖拽调序（collect 按 y 升序 = 审批顺序）。
  - 验收（bu 浏览器实证）：新增流程画布 3 节点 2 边、修改流程画布 4 节点 3 边完整渲染（async:false 修复后）；删除节点 / 添加节点 / updateNode 改名即时生效；整窗保存 PUT 成功且后端 GET 数据一致（早期「nodes:0」系未带 token 401 误读 data=null）；调试标记（`__fcModule/__fcSetup/__fcDebug/__hmrProbe/__openDebug` + body dataset）已全部清理；type-check + build 全绿（生产 gzip/brotli，FlowCanvas chunk 约 554kB / gzip 158kB）。
- 复查 + 明暗主题适配（#81 收尾复查，2026-11-01）：
  - 复查修复（`index.vue` + `FlowCanvas.vue`）：①**节点配置弹窗深拷贝快照**（`openNodeDialog` 曾直接引用画布节点对象，取消后 v-model 已污染画布 data，保存会把未确认改动写入后端——改为 `{...node, approverIds:[...]}` 快照，实测：改名-取消-保存后 GET 节点名未变）；②**提交前预校验审批人**（`handleSubmit` 对 user/role 节点 approverIds 非空前置拦截，提示「节点「XX」未配置审批人」，与后端 SaveFlowDef 口径一致）；③**起止节点固定**（`interacting.nodeMovable` 回调仅审批节点可拖，实测拖「发起申请/流程结束」位置不变）；④画布容器 `overflow:auto`（拖出界可滚动查看）。
  - 实测确认无问题：X6 3.x 拖拽结束不触发 `node:click`（不会误弹节点配置窗）；空审批人保存被前端拦截且弹窗不关。
  - **明暗主题全面适配**：CSS 部分 CSS 变量化（`.canvas-container` 背景 `var(--el-bg-color)` / 边框 `var(--go-border-color)`、`.flow-canvas-wrap` 背景 `var(--el-fill-color-lighter)`、tip 文字 `var(--el-text-color-secondary)`）；X6 节点颜色改 `palette()` 明暗双色板（审批节点深色 `#1f2d3d`/浅色 `#fff`、label `#e5eaf3`/`#303133`、summary `#a8abb2`/`#909399`、START 深绿 `#16301f`/浅绿 `#f0f9eb`、END 深红 `#3a1f1f`/浅红 `#fef0f0`、连线 `#5c5c66`/`#909399`）+ **MutationObserver 监听 `html.dark` 类实时重绘节点与连线**（与 dashboard 页惯例一致，onBeforeUnmount disconnect）+ 节点 hover 微反馈（浅色 brightness 1.02 / 深色 1.2）。实测：切深色节点/文本/连线/画布背景全量切换，切回浅色双向恢复。
  - 复查再提升（同日）：`saveNodeDialog`/`handleSubmit` 的 `form.validate()` 改 `.catch(() => false)` 静默处理（原裸 await 在校验失败时产生 unhandledrejection 控制台报错，实测空表单提交弹窗不关、控制台零 error/warning）；`openDialog` 节点字段补默认值（异常/旧数据缺字段时兜底 `nodeName`/`approverType`/`nodeMode` 等）；`addNode` 清理未使用的 `insertY` 死代码。
  - **连线跟手优化（同日）**：edge 由 `router:'orth'` 直角路由升级为 `router:{name:'manhattan',args:{padding:12}}`（避障路由，拖拽时实时重算、路径不穿节点）+ `connector:{name:'rounded',args:{radius:6}}`（圆角连接）+ `smooth:true`（平滑），拖拽节点时连线更丝滑、路径更合理。实测连线路径含圆角曲线命令，垂直直连保持直线、拐角圆滑。

### 〇.9 首页大屏增强（dashboard）
- **视觉炫酷**：欢迎卡加渐变流光动画（`@keyframes welcomeShift` 背景位移）+ 实时时钟（每秒刷新，tabular-nums）；统计卡数字滚动动画（自实现 cubic 缓动 count-up，无新依赖）+ 卡片 hover 上浮/图标缩放；登录趋势折线改渐变面积（`echarts.graphic.LinearGradient`）+ 发光阴影（`lineStyle.shadowBlur`）+ 圆点 symbol；操作类型饼图自定义 6 色渐变板 + 强调放大（`emphasis.scaleSize`）+ 边框透明化。
- **内容丰富**：新增「工作流概览」卡（我的待办/已办事项/我发起的，走 `listFlowTodo/listFlowDone/listFlowApply` total，点击跳转 `/flow/todo|done|apply`）；新增「系统信息」卡（Go 版本/运行时长/服务启动/CPU·内存·磁盘使用率，走 `getServerInfo`）；新增「快捷入口」8 宫格（用户/角色/菜单/部门/流程定义/公告/在线用户/定时任务，`router.push` 跳转，hover 缩放）。
- 全部沿用既有 dark MutationObserver 换色 + resize 自适应 + 逐接口失败静默（无权限留占位）。




### 修复的真实 bug
- **`scheduler/executor.go`**（第 68 行）：`cache.Info()` 返回的 `hitRate` 为 `int64`，而 `fmt.Sprintf` 使用 `%.1f%%` 动词（需 `float64`），原代码运行时会输出 `%!f(int64=xx)` 占位错误，`go vet` 也报错。已修为 `float64(hitRate)`，仅类型转换、数值含义不变。

### gofmt 规范化（14 个文件，纯空白/对齐整理，无逻辑改动）
- `api/`：api_doc_api.go、dept_api.go、gen_api.go、upload_api.go
- `dao/`：api_dao.go、config_dao.go、dict_cache.go、gen_dao.go
- `model/`：api.go、gen.go、job.go、job_log.go、role.go
- `middleware/`：operlog.go

### 注释补齐
- `middleware/operlog.go`：`Write` / `WriteString` 补充 godoc 注释
- 全量扫描 74 个 .go 的导出标识符，其余均已具备中文 godoc 风格注释

### 回归验证（backend 目录实测）
- `go build ./...`：0 错误
- `go vet ./...`：0 错误 0 warning
- `gofmt -l .`：为空（全部符合 gofmt）
- 未重启 8080 服务、未动 `data/go-admin.db`、`uploads/`、`templates/`

## 二、前端 Vue/TS（frontend/，42 个 .vue + 38 个 .ts）

### 新建规范工具链
| 文件 | 说明 |
| --- | --- |
| `.prettierrc.json` | printWidth 100 / semi true / singleQuote true / trailingComma none / 2 空格缩进 |
| `.prettierignore` | 排除 dist / node_modules / public |
| `.editorconfig` | 统一 UTF-8、CRLF、2 空格（Go 文件用 tab）；后端同样受益 |
| `eslint.config.mjs` | ESLint 扁平配置：`@eslint/js` 推荐 + `typescript-eslint` 推荐 + `eslint-plugin-vue` flat 推荐 + `eslint-config-prettier`（冲突关闭）；项目豁免：`no-explicit-any`（request.ts ApiClient 合理用法）、`vue/multi-word-component-names`（视图按路由命名）、`caughtErrors: none`（catch(e) 参数保留 RuoYi 风格）；浏览器+Node 全局变量显式声明 |

### 新增依赖（devDependencies）
`prettier@3.9.10`、`eslint@10.13.0`、`typescript-eslint@8.72.1`、`eslint-plugin-vue@10.12.0`、`vue-eslint-parser@10.4.1`、`eslint-config-prettier@10.1.8`、`@eslint/js`、`globals`

### package.json 新增脚本
```json
"lint": "eslint .",
"lint:fix": "eslint . --fix",
"format": "prettier --write src"
```

### 全量格式化与 lint 修复
- Prettier 全量格式化 80 个源文件（含模板属性顺序由 ESLint 自动修复：`vue/attributes-order` 等 64 处 warning）
- **修复 21 处真实 lint 错误**：
  - 删除未使用 import：`Navbar.vue`（FullScreen/Aim）、`layout/index.vue`（FullScreen/Aim）、`gen/index.vue`（Refresh/Plus）、`login.vue`（login）
  - 删除未使用变量/函数：`ColumnSetting`（keys 死变量）、`gen/index.vue`（resetQuery 死函数，模板未引用）
  - 未使用回调参数：`TagsView.vue`（watch path）
  - 表达式语句规范：`permission.ts` / `role/index.vue` / `DesktopArea.vue` 的 `a && a.x()` 改为 `a?.x()` 或 if 语句；`dashboard/index.vue` 5 处改为可选链
  - 空 catch 块：`loginlog/index.vue` 2 处补充注释（用户取消操作）
  - `settings.ts`：`drawerVisible` 解构排除改为 `delete rest.drawerVisible`（语义更清晰）
  - 首轮 api 注释误用实体名（`api/user.ts` 等标为"接口"）已修正为对应实体（用户/角色/菜单…）

### 注释整理（全项目中文注释，函数级覆盖）
- 脚本扫描 + 人工补充 **约 290 处**函数注释：
  - `api/*.ts` 全部接口封装（97 处）：`// 查询用户列表`、`// 新增角色`、`// 重置用户密码` 等
  - `views/*` 页面 CRUD 函数（140 处）：fetchList/handleQuery/resetQuery/handleSubmit/handleDelete/handleExport/handleClean 等
  - layout/utils/store 语义函数（53 处）：桌面窗口移动/缩放回调、页签右键操作、开始菜单级联、背景更换、令牌读写等
- 复核扫描：**缺失注释函数 0**（含嵌套回调）

### 回归验证（frontend 目录实测）
- `npm run lint`：0 error 0 warning
- `npm run type-check`（vue-tsc --noEmit）：0 错误
- `npm run build`：通过（13.11s）
- vite 5173 热更新自动生效，未重启

## 三、使用方式

```bash
# 后端（Go 已安装于 E:\Go\bin）
cd backend
go build ./... && go vet ./...

# 前端
cd frontend
npm run lint        # 规范检查
npm run lint:fix    # 自动修复
npm run format      # prettier 格式化
npm run type-check  # TS 类型检查
npm run build       # 构建
```

## 四、已知说明
- Go 命令未加入当前会话 PATH（安装在 E:\Go\bin），用完整路径调用：`& "E:\Go\bin\go.exe" build ./...`
- 前端 lint 豁免项（no-explicit-any / multi-word-component-names / caughtErrors）为项目刻意保留，符合 RuoYi 生态惯例

---

## 五、首页大屏重设计 + 侧边栏布局改造 + 登录热力图（2026-11-01）

### 首页大屏重设计（frontend/src/views/dashboard/index.vue 整页重写）
- 移除整段欢迎语（"你好，管理员，欢迎使用..."）与快捷入口 8 宫格、工作流 3 卡（用户授权"用不上的可以不要"）
- 布局：`.dash` 容器 height:100% + flex column + overflow:hidden，**任何方向无滚动条**；统计卡 7 张多彩渐变（User/Avatar/Menu/OfficeBuilding/Bell/SwitchButton/Connection，数字滚动缓动）
- 图表行 1：近 7 天登录趋势折线（渐变面积 + shadowBlur glow + symbol circle，span 2 宽）+ 登录热力图（7 天 × 24 小时矩阵，visualMap 深蓝→紫→红→橙渐变）
- 图表行 2：操作类型玫瑰图（roseType radius）+ 系统资源三表合一仪表盘（CPU/内存/磁盘 progress 环 + valueAnimation）+ 最新公告表格（el-tag 通知/公告 + 标题 + 时间，点击开详情 dialog）
- ECharts 按需注册：LineChart/PieChart/HeatmapChart/GaugeChart + Grid/Tooltip/Legend/VisualMap + CanvasRenderer；dark 主题 MutationObserver 切图表文字/分割线颜色

### 关键调试修复（图表渲染竞争的根治）
- **接口响应结构**：request 拦截器返回业务体 `{code,data,...}`，图表数据处理曾误按数组直接用 → 新增 `extractList()` 统一提取 `.data` 数组（trend/pie/heat 三图全空 → 全部恢复）
- **0 高容器**：图表 init 时容器高度未稳定导致 canvas 布局塌陷（trend 只渲染底部 X 轴线、gauge detail 错位）→ `ensureSize()` 等 clientWidth/clientHeight>0（≤40 帧）+ `safeInit()`（init 后 getWidth/getHeight 为 0 则 dispose 重建）
- **数据驱动幂等**：数据存 ref（trendData/pieData/heatData/serverInfo）+ `renderData()` 幂等应用，接口完成与图表初始化两处调用，消除时序竞争
- 公告时间格式：parseTime pattern 为 `MM-DD` 风格（非 `{m}-{d}`），修正后显示 `10-23`

### 侧边栏布局改造
- `layout/components/Sidebar.vue`：Logo 区 flex space-between，新增 `logo-fold` 折叠按钮（Fold/Expand 图标，切换 settings.sidebarCollapsed + save()），位于 "Go Admin" Logo 区域右侧；折叠态 60px 紧凑排布
- `layout/components/Navbar.vue`：移除原折叠按钮，面包屑（max-width:45%，<768px 隐藏）+ 内嵌 `<TagsView>` 页签同一行；navbar height 50→48
- `layout/components/AppMain.vue`：TagsView 条件加 `layoutMode !== 'side'`（side 布局由 Navbar 承接，top/mix 原位显示）

### 后端登录热力图接口
- `dao/dashboard_dao.go`：`LoginHeatItem{Date,Hour,Count}` + `GetLoginHeatmap(days)`（sys_logininfor 按 天 × 小时 分组计数）
- `api/dashboard_api.go`：`LoginHeatmap` handler（utils.OkData）
- `router/router.go`：`auth.GET("/dashboard/loginHeatmap", api.LoginHeatmap)`
- 实测：4 个 dashboard 接口全部 200；heatmap 真实数据（2026-10-28 11 时 6 次等）；趋势 10-26~11-01 [0,0,14,14,2,46,15]；操作类型 查询1751/新增96/修改125/删除5；server info cpu 33.7/mem 79/disk 93.8（实时变化）

### 浏览器实测验收（4180 preview + canvas 像素采样）
- 4 图全部渲染且有真实数据：trend 折线（非空像素 ~41990，x 轴 10-26..11-01 + y 轴 0-50）、heat 热力（~18269，visualMap 0-26）、pie 玫瑰（~13584，图例 查询/新增/修改/删除）、gauge（~9143，CPU 41%/内存 79%/磁盘 94% 实时值）
- 折叠按钮：点击后菜单收起（侧边栏 60px 图标态），按钮随 Logo 区
- 面包屑「首页」+ 页签「首页」同 Navbar 一行；无任何滚动条；欢迎语已移除
---

## 六、首页圆角 2px + 新增两种布局（底部导航 / 右侧边栏）（2026-11-01）

### 首页圆角 2px（frontend/src/views/dashboard/index.vue）
- `.dash-bar`（数据总览顶栏）border-radius 6px → 2px
- `.stat-card`（多彩统计卡）border-radius 8px → 2px
- `.panel`（图表/公告面板 el-card）新增 border-radius: 2px
- 圆形元素保持不变：`.stat-icon`（统计卡图标容器）、`.dot`（面板标题圆点）仍为 50%

### 新增布局模式（与现有布局互为镜像）
- `rightSide` 右侧边栏：与 `side` 相反——内容区在左、侧边栏完整菜单在右。复用 Sidebar/Navbar（面包屑+页签承接），镜像处理：侧边栏左缘分隔线（border-left）、Logo 区内边距左右互换（padding 0 18px 0 8px）、窄窗（<768px）同样临时折叠
- `bottom` 底部导航：与 `top` 相反——内容区（面包屑 + AppMain + 页脚）在上，Logo + 横向一级菜单（TopNav）+ 用户区/搜索/全屏/设置单行在底部（`.bottom-header` box-shadow 0 -1px、border-top，与顶栏镜像）。子菜单经 Element Plus popper flip 自动向上弹出，不超出视口底部
- 改动文件：`store/settings.ts`（注释）、`layout/index.vue`（模板新增 rightSide/bottom 分支 + wrapper class + `.bottom-main`/`.bottom-header`/`.layout-rightSide` 样式 + isSide 窄窗判断含 rightSide + 480px 窄屏底部栏 wrap）、`layout/components/AppMain.vue`（TagsView 条件排除 rightSide）、`layout/components/SettingsDrawer.vue`（布局 radio 增加右侧边栏/底部导航 + 6 选项换行样式）
- 页签归属：side/rightSide 由 Navbar 承接（面包屑同行）；top/bottom/mix 在 AppMain 原位显示

### 浏览器实测（4180 preview，1440 视口）
- 圆角：dash-bar/stat-card/panel 均为 2px，stat-icon/dot 50%
- rightSide：wrapper=layout-rightSide，侧边栏 right=967（视口右缘）、分隔线 borderLeft=1px、logo padding 镜像；main 0-907
- bottom：导航栏 top=834/bottom=880（底部 h46）、5 菜单项（首页+系统管理/监控/工具/流程中心）；面包屑 top=-3 + 页签 top=33 在内容区顶部；hover 系统管理子菜单 popper top=454/bottom=826 自动向上弹出不截断
- 回归：切回 side 布局侧边栏 0-200、菜单完整、dashbar 圆角 2px、大屏图表正常
---

## 七、新增「转盘」布局（右上角 1/4 圆转盘菜单）（2026-11-01）

### 新组件 frontend/src/layout/components/DialMenu.vue
- 转盘布局（layoutMode='dial'）：全屏内容区 + 右上角悬浮 1/4 圆转盘菜单
- **最小态**：右上角 30px 小圆钮，透明度 0.4（渐变主题色），hover 或点击放大
- **展开态**：圆心在视口右上角，1/4 圆扇形底（conic-gradient + border-radius 0 0 0 100%）
- 菜单项沿圆弧分布：`rotate(deg) translateX(radius)`（圆心旋转 + 径向平移），内层 `rotate(-deg)` 保持文字水平；单圈最多同时可见 8 项（90°/8 = 11.25°/项），超过则滚轮滑动（wheel 改 offset，项沿弧平滑移动，`transition transform 0.28s`）；项少（≤8）不显示滚动提示、滚轮无效
- **级联展开**：点击目录项（M，有菜单子项）→ 其子菜单在更外圈显示（半径每级 +96px，148/244/340…），逐级向外；当前路径链高亮（active）；点击叶子（C 或无可点子项）→ `router.push` 完整路径 + 转盘回到最小态
- 一级菜单固定含「首页」入口；数据源 userStore.menus（与 Sidebar 同源），i18n menuTitle + safeIcon
- **自动收起**：鼠标移出转盘交互区 600ms 收起（移入取消计时）
- 改动文件：`store/settings.ts`（注释）、`layout/index.vue`（dial 模板分支 + import）、`layout/components/SettingsDrawer.vue`（布局选项加「转盘」+ 描述）、新建 `layout/components/DialMenu.vue`

### 浏览器实测（4180 preview）
- 最小态：右上角 30px 圆钮、opacity 0.4
- hover 展开：一级 5 项沿圆弧（首页/系统管理/系统监控/系统工具/流程中心，位置符合 r=148 弧线坐标）
- 点击「系统管理」→ 二级外圈出现 8 项（用户管理…通知公告，r=244），一级 active 高亮
- 点击二级叶子「用户管理」→ 跳转 /system/user，转盘回最小态（mini 显示、panel 移除）
- 点击「流程中心」→ 二级（流程定义…抄送我的）→ 流程定义为叶子直接跳转 /flow/def
- 鼠标移出转盘区域约 600ms 自动收起（dispatch mouseleave 实测 panel=false）
- 注：菜单树当前无 >8 项的圈，滚轮路径按代码逻辑（clamp offset + 项沿弧滑动）实现，未实景触发
---

## 八、布局检查 + 名称调整 + 流程设计器弹窗优化（2026-11-01）

### 三布局检查（bottom / rightSide / dial，浏览器实测）
- 底部导航（bottom）：窄窗（<480px）两行式正常（Logo+右侧 50px / 菜单 42px，栏高约 93px）；子菜单 popper 自动向上弹出（内容主体远离底部导航，不截断）；菜单点击跳转正常（系统管理→用户管理 → /system/user）；console 无报错
- 右侧边栏（rightSide）：wrapper class=layout-rightSide、侧栏贴视口右缘（pinnedRight）、内容区在左、Logo 内边距镜像（0 18px 0 8px）、窄窗自动折叠到 60px；console 无报错
- 转盘（dial）：与首页大屏共存正常（4 图表 canvas 渲染、TagsView 显示、统计卡数据完整）；转盘最小态小圆钮正常
- 未发现三个布局的回归 bug

### 布局名称调整
- SettingsDrawer 布局选项「侧边栏」→「左侧边栏」（与右侧边栏成对），描述同步更新（左侧边栏：菜单在左侧（默认））
- store/settings.ts 注释同步

### 流程设计器弹窗优化（flow/def）
- 弹窗 width 760px → 960px（+max-width: calc(100vw - 32px) 防窄屏溢出）、top 6vh → 4vh、新增 class=flow-def-dialog
- 表单紧凑：label-width 90 → 80、el-row gutter 16 → 12、字段 margin-bottom 18 → 10、备注 textarea rows 2 → 1、流程编排标题间距收紧
- 弹窗自身不滚动：.flow-def-dialog display:flex column + max-height calc(100vh - 8vh) + body flex:1 overflow:hidden（浏览器实测 body scrollH == clientH，无滚动条）
- 画布撑满剩余空间：.flow-form flex column flex:1、.flow-canvas-wrap flex:1 min-height:0、.canvas-container flex:1 min-height:120px（兜底）；工具栏 flex-wrap 换行防挤压
- FlowCanvas：graph 高度按容器撑满（X6 3.x Graph 尺寸跟随容器，height 参数仅作兜底），onMounted 等待容器宽高就绪后渲染
- 浏览器实测：弹窗 960 宽（窄窗收缩 187 不溢出）、body 无滚动、画布 flex 撑满（窄窗 toolbar 换行时兜底 120px）、3 节点渲染正常
---

## 九、转盘菜单可见性 bug 修复（2026-11-01）

### 现象
用户反馈：转盘布局点击展开后「菜单几乎都是透明的、看不见」，无法判断是内容加载还是透明度问题。

### 根因（浏览器实测定位）
- 内容加载正常（点击后 5 个一级菜单项正常渲染，opacity=1）；问题出在**菜单项文字颜色**：
- `.dial-item-inner` 使用 `color: var(--go-sidebar-text)` + `background: var(--go-navbar-bg)` 两组不同主题来源的变量——当侧栏主题为深色时 `--go-sidebar-text` 解析为浅灰 `rgb(191,203,217)`，而 `--go-navbar-bg` 仍为白色 → **白底浅字，视觉上几乎透明**
- 修复：改用 Element Plus 主题变量 `--el-text-color-primary`（文字）+ `--el-bg-color-overlay`（背景）+ `--el-border-color`（边框），浅/深色主题自动适配，实测文字恢复为深色 `rgb(48,49,51)` 白底，清晰可见

### 附带优化
- 自动收起延时 600ms → 1200ms：点击展开后给用户充分阅读/操作时间，避免鼠标移开后菜单一闪而过

### 验证（浏览器实测）
- 点击 mini → 一级 5 项展开，文字 `rgb(48,49,51)` 深色白底；点「系统管理」→ 二级外圈 8 项同样清晰（共 13 项可见）；mouseleave 后 1.2s 自动收起（panel 消失、mini 回归）
### 修正：真正根因是菜单项几何位置（视口外）
- 上一轮修的颜色只是次要因素；用户「看不见菜单」的**真正根因**：圆心锚定在视口右上角 (W,0)，而菜单项位移用 `translateX(+radius)` 沿旋转方向**向右平移** → 一级菜单全部渲染到**视口右侧外**（实测 x≈1114 > 视口宽 967）→ 用户只能看到右上角一片几乎透明的扇形底，菜单项在屏幕外
- 修复：菜单项改为**负旋转 + 负半径** `rotate(-deg) translateX(-radius)`，位移 = (-r·cosθ, +r·sinθ)，沿「左→下」弧落在视口内的 1/4 扇形区域；文字用正旋转抵消保持水平
- 实测：一级 5 项视口内沿弧排列（820→873, 8→44）；二级外圈 8 项同样视口内；点击叶子正常跳转并收起

### 第三轮：扇形角度展开 + 消除重叠（几何改造）
用户验收可见性后提出两项修正：①展开后菜单项有重叠；②菜单不应「横平竖直」，应以扇形角度展开。本轮完成几何重构。

- **角度分度**：`itemDeg = (idx + 0.5 - offset) * 90 / items.length`——不再按固定 VISIBLE=8 分度，而是**按该层实际项数均分 1/4 圆**（项少张开成完整扇形、项多自动收窄），角度窗口 [0°, 90°) 对齐右下 1/4 扇形
- **半径**：RADIUS_BASE 148 → 168（一级）、RADIUS_STEP 96 → 110（二级外圈半径 = 168+110 = 278）；扇形底 sector 半径随 maxRadius 动态（= radius + 40）
- **定位改为纯三角函数**（关键修复）：此前用 `transform: rotate(-deg) translateX(-radius)` 做位置+姿态，浏览器实测项视觉 rect 与 transform 矩阵理论位置系统性偏移（约 66px，命中测试证实项实际渲染在偏位处）→ 改为 **left/top 由 JS 直接计算**（项中心精确落在弧上：`left = -r·cosθ - 33`、`top = r·sinθ - 14`，圆心在视口右上角），`transform` 仅保留 `rotate(-deg)` 做**径向倾斜**（文字随项同角度，呈扇形姿态），彻底绕开位移偏差
- **滑动动画**：`transition: left .28s, top .28s, transform .28s, opacity .2s`——滚轮切换/层级切换时项沿弧平滑滑动
- **可见性窗口**：idx 落在 [offset, offset+VISIBLE) 且角度 ∈ [-4°, 94°) 双条件，滚轮 maxOff 逻辑不变
- **浏览器实测（视口 967×883）**：
  - 一级 5 项沿半径 168 弧均匀展开：中心距 ≈ 51px（理论弧距 52.8px，逐项吻合），角度 -9°/-27°/-45°/-63°/-81° 径向倾斜，视觉间隙清晰无重叠
  - 二级 10 项沿半径 278 弧展开至 y≈231，同样均匀无重叠；「用户管理/角色管理/菜单管理…」OCR 逐项可读
  - 滚轮滚动换页、点一级展开二级、点叶子跳转并收起均正常；菜单项 ref 截图确认白底深字胶囊随径向旋转
  - 说明：JS 矩形相交检测对旋转项报 overlap=true 属于**轴对齐包围盒误报**（胶囊旋转后 AABB 相交，胶囊本体沿弧间距 ~23px 实际分开），以视觉截图为准
---

## 十、流程弹窗加大 + 转盘设计按钮 + 二级固定间距扇形（2026-11-01）

### 1. 流程编排弹窗进一步加大
- width 960px → **1280px**、top 4vh → **2vh**
- `.flow-def-dialog`：`max-height: calc(100vh - 8vh)` → **`height: calc(100vh - 4vh)`（由 max-height 改固定高度，让画布真正撑满剩余空间）**；max-width calc(100vw - 32px) 防窄屏溢出保留
- 实测（视口 967×883）：弹窗高度 **477 → 848px**（≈96vh 撑满）、画布 canvas-container **≈120 → 491px**、body scrollHeight==clientHeight（**无滚动条**）；窄屏 1280 宽被 max-width 收缩至 935（不溢出），大屏下全宽生效

### 2. 转盘展开态圆心设计按钮（替代小圆钮）
- 展开态面板圆心（right 8 / top 8，与原小圆钮同位）新增 `.dial-design` 30×30 圆形渐变按钮（齿轮 Setting 图标，z-index:2，hover 放大）
- 点击 `@click.stop → openSettings()` → 置 `settings.drawerVisible = true` 打开**界面设计抽屉**（可再选其他布局 / 主题配色）；与最小态 `.dial-mini` 同位共存（v-if 互斥：mini 在未展开时、design 在展开时）
- 实测：展开后 design 按钮 flex 可见 @944,5；点击 → el-drawer 打开（drawer-open flex）、再点关闭正常

### 3. 二级菜单改为「围绕一级固定间距扇形展开」（不再均分整体半径）
- 新增 `CHILD_STEP_DEG = 12`：`childDeg = parentDeg + (idx - (n-1)/2) * 12`——**以所点一级的角度为中心，向两侧固定间隔展开**（项少间距不再拉大）
- 越界保护：圆心在视口右上角，扇区整体平移到 [0°, 90°] 保证全视口内；**项跨度 > 90°（子项 > 8）时回退为均分 90° 全显示**（系统管理 10 项即此支路，固定 9° 间隔 5°~85°，不用滚轮）
- 二级**取消滚轮**：onWheel / lastScrollable 对 parentDeg != null 层直接跳过；itemVisible 对二级恒可见（一级仍按 offset+VISIBLE 窗口滚动）
- onItemClick 记录所点项角度 `parentDeg: itemDeg(lv, item)` 传入子层
- 实测（视口 967×883）：
  - 点「系统工具」（一级 63°）→ 二级 3 项 **51°/63°/75° 围绕 63° 对称、固定 12° 间隔**（代码生成/接口文档/文件上传），完全符合「以一级位置为中心向两边固定间距」
  - 点「系统管理」（一级 27°，10 项）→ 二级 5°~85° 均分展开（固定 9° 间隔、全显示、无滚轮提示）；一级 5 项 9°/27°/45°/63°/81° 均分不变
  - 设计按钮点击打开抽屉、关闭正常
---

## 十一、转盘三项反馈修复：二级滚轮恢复 + 光圈收回 + 点击光圈外收起（2026-11-01）

用户三条反馈（硬性）：①二级/多菜单项时滚轮失效要恢复；②展开态光圈（扇形底）太大不要超出菜单显示区域；③点击光圈外边时整圈收起回到最小态。三项均浏览器实测通过。

### 1. 二级滚轮恢复（窗口化固定间隔）
- 上一轮「子项 > 8 回退均分 90° 全显示」被用户否决（间距被压缩、且二级取消了滚轮）→ 改为**窗口化**：
  - `winN = min(n, VISIBLE)`、`rel = idx - off - (winN - 1) / 2`、`deg = parentDeg + rel * CHILD_STEP_DEG`——窗口内仍固定 12° 间隔，只显示连续 8 项，超出部分滚轮浏览
  - **越界平移与 offset 无关**（关键修复）：`halfSpan = (winN - 1) / 2 * 12`，`deg += halfSpan - base`（左越界）/ `deg += 90 - (base + halfSpan)`（右越界）——初版误把平移写成与 off 相关，滚动后窗口整体上移致末项 96°~108° 被 itemVisible 隐藏
  - `itemVisible` / `lastScrollable` / `onWheel` 统一为所有层（items.length > VISIBLE 即可滚，一级/二级一致）
- 实测（系统管理 10 项，一级 27°）：滚轮 0 次窗口 用户0°…通知84°（日志96°/接口108° 隐藏）；滚轮 1 次 → 角色0°…日志84° 8 项全可见；滚轮 2 次 → 菜单0°…接口84°；hint「滚轮浏览」显示

### 2. 光圈收回菜单弧
- `maxRadius` 由 `deep + 40` 改回 `deep`（最深圈菜单弧半径），sector = maxRadius × 2
- 实测：一级状态 `.dial-sector` CSS **336×336**（此前 416=2×(168+40)，超出菜单弧 40px）；二级 556×556 同样贴弧

### 3. 点击光圈外边整圈收起
- 用户原话「点击光圈外边的时候，菜单就全收起来」。实现分两层，**全部走 DOM 元素事件、不依赖 document 全局监听**（此前 document 捕获/冒泡监听多次实测在页面 bundle 中均不触发，根因未定，弃用）：
  - **全屏透明遮罩 `.dial-backdrop`**（v-if expanded、fixed inset 0、z-index 2999 < wrap 3000）：点击转盘以外任意位置 → `@click="collapse"` 整圈收起
  - **wrap 自身 `@click="onWrapClick"`**：点击 wrap 内空白（如 1/4 圆外的右下直角区）也按光圈几何判定收起；交互元素（`.dial-mini/.dial-design/.dial-item/.dial-ring/.dial-panel`）直接放行，不误收起
  - 遮罩低于 wrap：转盘滚轮、hover、菜单点击均不受影响
- 实测（视口 967×883，bu.click_xy 真实点击）：
  - 展开 → 点页面中部（光圈外）→ **panel-collapsed** ✓
  - 展开 → 点 wrap 内右下空白（900,280，扇形外）→ **panel-collapsed** ✓
  - 展开 → 点圆心设计按钮 → 抽屉打开且 panel 保持展开（不误收起）✓
  - 展开 → 点一级「系统管理」→ 二级正常展开；滚轮窗口正常
  - 点二级叶子 → 跳转页面 + 转盘收起（既有行为回归通过）

### 回归
- 光圈 336×336 不超菜单弧；二级滚轮窗口 8 项固定 12° 间隔全可见；点击光圈外收起；设计按钮开抽屉；点叶子跳转收起——浏览器逐项实测通过（verify28）

### 12. 深检修复：设计抽屉被转盘全屏遮罩拦截（verify29）
- 现象（继续检查时发现）：点圆心设计按钮打开抽屉后，**抽屉内所有点击无效**（实测点「English」语言选项，选中态不变）——面板层叠关系：`.dial-backdrop` z-index 2999 **高于** el-drawer overlay 2004，全屏遮罩拦截了抽屉内点击
- 修复：`openSettings()` 打开抽屉前先 `collapse()`（收起转盘、遮罩随之消失），抽屉独占交互
- 实测（verify29，视口 1051×1083）：点 design → 抽屉开 + 转盘收起 + backdrop 消失；点「English」语言切换成功、切回中文成功；关抽屉后三项回归通过（滚轮窗口 8 项、点圈外收起、光圈 336px）

### 13. 首页图表在桌面布局不自适应 → ResizeObserver + 容器查询（verify30）
- 现象（用户实测反馈）：首页切到桌面布局后图表不随窗口缩放——根因一：`dashboard/index.vue` 仅监听 `window resize`，而桌面布局下首页运行在可拖拽缩放的 `.desktop-window` 内，窗口变化不触发 window resize；根因二：首页响应式用 `@media (max-width:900px)` 按**视口**判断，桌面窗口缩窄时视口不变、列数不切换，图表被挤压
- 修复 ①：图表容器（trend/heat/pie/gauge 四个 `.chart-box`）挂 `ResizeObserver`，容器尺寸变化即 `resizeCharts()`；onBeforeUnmount 断开；原 window resize 监听保留（侧栏布局浏览器缩放仍受益）
- 修复 ②：`.dash` 设 `container-type:inline-size; container-name:dash`，两处 `@media` 改 `@container dash`；窄容器下 stat-row 4 列、chart-row 1 列、panel `min-height:220px`、`.dash overflow-y:auto` 整页滚动——布局随**窗口宽度**（而非视口）响应
- 修复 ③（verify31 阈值微调）：初版阈值 900px 会误伤**侧边栏布局**（内容区 851px < 900 也被切成单列竖排，观感大变）——阈值降为 **760px**：side 内容区 851px 恢复 3 列原观感；桌面默认窗口 757px（<760）保持单列可滚动；窗口放大 900px 恢复 3 列；dial 全宽 1051px 3 列不变
- 实测（verify30，视口 1051×1083）：窗口 757→600 图表 458/211→354/158、600→950→587/275（双向自适应）；窗口 757 时 stat 4 列 + 图表单列 + 面板 220px 可滚动；窗口 1050 时恢复 stat 7 列 + chart 3 列；切回 dial 布局回归正常（7 列/3 列、图表 650/306）

### 14. 桌面布局继续优化：容器阈值校准 + 窗口位置记忆（verify31）
- 容器阈值校准：`@container` 阈值 900→**760px**——原 900 会误伤侧边栏布局（内容区 851px 被切成单列竖排）。760 后：side 851px 保持 7 列 stat + 3 列图表原观感；桌面默认窗口 757px 单列可滚动；窗口放大 900px 恢复 3 列；dial 全宽 3 列不变——三布局实测通过
- 窗口位置/尺寸记忆：`.desktop-window` 的 x/y/w/h 持久化到 `localStorage 'go-admin-desktop-windows'`——刷新/切布局后恢复用户上次排布（z、最小化不持久化按会话重建；新视口越界自动夹回画布）；窗口关闭时清除该窗口记录；真实拖拽/缩放/最大化后即时落盘
- 实测（verify31）：预置 `{x:150,y:120,680×520}` 等 → reload 后窗口按预置位置尺寸恢复；bu 真实拖拽首页标题栏 → 窗口移动且 store 同步更新 `{"/index":{x:300,y:290,...}}`

### 15. 桌面布局 Windows 风格四项优化（verify32）
- **开始菜单搜索**（StartMenu.vue）：基础面板顶部加搜索框，对菜单树全量扁平化按「标题/父路径/path」模糊匹配，结果列表显示名称+父路径，支持 回车/↑↓ 键选 + 点击直达；搜索时级联浮层 v-show 隐藏；复用导航菜单既有 i18n（searchMenuPlaceholder/searchMenuEmpty）。实测：输入「用户」→ 命中「用户管理系统管理」「在线用户系统监控」，点击直达 /system/user 并关闭菜单
- **任务栏分组/预览**（DesktopArea.vue）：窗口按路径首段归入顶层菜单组；组内单窗口直接显示窗口项，多窗口合并为「组名+N」按钮；点击弹出浮层窗口列表（图标+标题+激活圆点），点击项切换窗口；组名从动态菜单树取顶层菜单翻译名。实测：角色管理+用户管理合并为「系统管理 2」，浮层切换角色管理窗口生效
- **窗口半屏吸附**（DesktopArea.vue moveWindow）：拖动中标题栏到画布左/右边缘（≤14px）吸附左/右半屏（宽=画布一半、满高），到顶部边缘最大化；吸附位置即时持久化。实测：拖左缘 → 0,0 525×1041、拖右缘 → 526,0 525×1041、拖顶部 → 全屏最大化；reload 恢复吸附位置
- **桌面图标墙**（DesktopArea.vue）：画布左侧显示「回到首页」+ 各顶层菜单应用图标（单击选中/双击打开）；目录双击打开其第一个可访问叶子页。实测：5 图标显示，双击「系统管理」打开用户管理窗口

### 16. 设置抽屉删除布局说明文字（verify33）
- 用户反馈布局模式下方的说明文字（各布局模式行为描述，多行堆叠）不需要
- SettingsDrawer.vue：删除「布局模式」组下的 `.group-desc` 说明文字（含全部布局描述），并清理不再使用的 `.group-desc` 样式；布局单选按钮保留
- 实测：设置抽屉布局模式下方无说明文字，7 个布局按钮正常

### 17. 新增三种新颖布局/导航形态（verify34）
- 用户点名「Dock 悬浮坞 / 垂直页签 / 命令面板 三种都做」
- **Dock 悬浮坞**（新布局 dock，新组件 DockBar.vue）：底部中央悬浮应用坞（macOS 风格），深色半透明毛玻璃容器；图标=首页+各顶层菜单（多彩渐变图标），hover 上浮 10px + 放大 1.18 + 图标 52→60px + 名称气泡提示；点击打开（目录取 firstLeaf 首个可访问叶子页）；右端分隔线+设置入口。实测：hover 放大/tooltip、点击系统管理直达 /system/user
- **垂直页签**（新布局 vtab，新组件 VTabBar.vue）：左侧 128px 垂直页签栏（浏览器侧栏标签风格），数据源=tagsView.visitedViews；点击切换、hover 显示关闭按钮关闭页签（当前页签关闭后回退到最后一个）；底部设置/首页快捷按钮；深/浅侧边栏主题适配。实测：切换、新增、关闭全通过
- **全局命令面板**（升级 SearchMenu.vue）：Ctrl+K/⌘K 打开，输入同时匹配「菜单直达 + 快捷命令」；命令区=回到首页/打开布局配置/切换全屏/刷新页面 + 9 种布局切换（当前布局带 ✓ 标记）；分组展示（命令/页面），↑↓ 循环导航、回车执行、Esc 关闭，footer 快捷键提示。实测：Ctrl+K 打开 13 条命令、输入「桌面」过滤唯一命令、回车切换 layoutMode=desktop
- 布局总数 7 → 9（side/rightSide/top/bottom/mix/dial/desktop/dock/vtab）；设置抽屉新增「Dock 坞」「垂直页签」单选；i18n zh/en 补命令面板与布局名 key

### 18. 三种新形态继续打磨（verify35）
- 用户点名「继续打磨」→ 三项各加一个亮点
- **Dock 自动隐藏**（DockBar.vue）：初始 mouseenter/mouseleave 方案有缺陷（鼠标从未进入 Dock 就不触发）→ 改为全局 mousemove：视口底部 120px 条带视为激活区，鼠标离开 1.2s 后 Dock 沉底隐藏（transform 下沉 + opacity 0），仅留底部 56×5px 半透提示条；鼠标移回底部条带即弹出。实测：页面中央停留 → hidden=true、bar opacity 0、peek opacity 1；移到底部 → 弹出恢复
- **垂直页签拖拽排序 + 中键关闭**（VTabBar.vue）：HTML5 draggable（首页 affix 不可拖），dragstart 记录路径、drop 重排 visitedViews，顺序持久化 localStorage `go-admin-vtab-order`（启动时按序恢复）；中键（auxclick button=1）关闭页签。实测：bu 原生 drag 被浏览器策略抑制（kDragAndReleaseToMoveSuppressed）→ 用 DragEvent 派发验证：拖用户管理至首页前 → DOM 变 [用户管理,首页]、order=[/system/user,/index]；中键关闭用户管理 → 剩首页并回跳
- **命令面板最近使用**（SearchMenu.vue）：执行命令/菜单时记录 key（命令 id 或菜单 path）到 localStorage `go-admin-palette-history`（去重置顶、上限 5 条）；空输入时显示「最近使用」分组（⚡ 命令/菜单图标区分），并入 activeIndex 索引链（历史→命令→菜单），回车/点击可直接执行历史项。实测：执行「打开布局配置」→ 再次打开面板出现「最近使用」分组首项
---

### 19. 整体系统检查 + 三种新形态再优化（verify36）
- 用户要求「整体系统检查」→ 逐布局走查 9 种布局发现：**dial 布局右上角仅时间，无账号/搜索/全屏入口**（历史遗留：个人中心/退出登录/改密无处可用）——其余布局入口均完备（vtab 已有账号/全菜单/设置/首页，dock 已有 Launchpad/账号/设置，desktop 右下角账号 ✓，命令面板在 dial 下 Ctrl+K 全局可用——此前「不响应」为 el-dialog 选择器误判）
- **dial 右上角工具条**（DialMenu.vue）：转盘下方竖排「命令面板(Ctrl+K)/全屏/账号(UserDropdown 头像)」三个半透明圆钮，与转盘同风格；转盘展开时自动隐藏（v-show=!expanded）避免与转盘交互区叠放。实测：三按钮出现、账号按钮可弹用户菜单
- **Dock 运行指示点**（macOS 白点，DockBar.vue）：按 visitedViews 路径首段归组，顶层菜单组有页签打开时图标下方亮 5px 白点（首页页签 affix 常驻常亮）；激活态 outline 高亮保留。实测：首页白点常亮；打开 /system/user 后「系统管理」白点+激活同时点亮
- **垂直页签右键菜单**（VTabBar.vue）：@contextmenu.prevent 浮层「刷新页面(tagsStore.refreshView)/关闭当前/关闭其他(delOthersViews)/关闭所有(delAllViews)」，复用 tagsView i18n keys；点击菜单外关闭；首页 affix 项不显示「关闭当前」。实测：菜单四项、关闭所有只剩首页、首页右键无「关闭当前」
- **命令面板更多命令**（SearchMenu.vue）：新增「切换语言(settings.setLanguage)/切换暗黑模式(darkMode+applyTheme+save)/清空最近使用(清 localStorage 且 runCommand 不记录自身)」；zh/en 补 cmdLang/cmdDark/cmdClearRecent。实测：三命令出现、清历史生效、暗黑切换 darkMode=true 且 html.dark class 生效
- 另补：vtab 底部快捷区新增「全屏」入口
---

### 20. 移动端 App（uni-app + Vue3，RuoYi-App 架构，任务 #95~#102，verify37）
- 用户点名「再加一个手机版本，uni-app 那种 vue3，参考 ruoyi 的 app 端架构，先做主要功能：登录、权限等」
- **脚手架**：`npx degit dcloudio/uni-preset-vue#vite .` 官方 vite 版 vue3 模板到 `D:\Golang\go-admin\app`；版本锁定 `@dcloudio/*`=3.0.0-5030420261029002（vue3 发行版）、vue 3.4.21、vite 5.2.8
- **pinia 兼容坑**：pinia 2.2.6/2.3.1 要求 vue≥3.5.11 与模板 ERESOLVE → 改装 `pinia@2.0.36`（`--legacy-peer-deps`）；主流程 `main.js` 按 uni 模板约定 `createSSRApp + app.use(Pinia.createPinia()) + return {app, Pinia}`
- **架构分层（对标 RuoYi-App）**：`utils/auth.js`（token KEY='Admin-Token'，与 Web 端 request.ts 同键）、`utils/request.js`（uni.request 封装：BASE_URL='http://localhost:8080'、Bearer 头、HTTP 401 与业务 code 200/401/500 统一处理、handleUnauthorized 清 token 后 reLaunch 登录页）、`api/login.js`（5 个接口）、`store/user.js`（pinia：token/userInfo/roles/permissions/menus；login/getInfo/getRouters/init/hasPerm（admin 或 *:*:* 通配放行）/logout/reset）
- **页面**：登录页（账号/密码/验证码图点击刷新/记住账号 checkbox/登录，`captchaEnabled=false` 免验证码）；首页（欢迎卡+统计+菜单宫格，目录→menu 页下钻、页面菜单→showModal 提示）；我的（用户卡+角色权限展示+常用功能+退出确认 modal）；菜单页（面包屑下钻列表，onLoad 收 parentId/title）
- **权限守卫**：App.vue onLaunch/onShow 登录态检查（有 token→init() 拉信息，无→reLaunch 登录页）+ request 401 统一处理 + store.hasPerm
- **构建**：`npm run build:h5` exit 0（Compiler 5.34 vue3）；`npm run dev:h5` → 5173
- **实测卡点与解法**：①**点阵验证码 OCR 成功率低**（9347/5781 等多次失败，验证码一次性失败即作废）→ 批量取 3 张逐张尝试（「1421」成功）；②**5173 被桌面端历史 vite dev 进程占用**（PID 6200，页面实为桌面 dashboard 导致截图/DOM 错乱）→ 杀旧进程重启 uni dev 后恢复正常；③bu.js 传参数组会被误当页面名 → token 直接拼接进 js 表达式
- **H5 全链路实测（浏览器实证）**：登录页 3 输入+验证码图+登录按钮 → 注入 token 重载 → 首页（欢迎卡+统计「4/全部/1」+菜单宫格 系统管理/系统监控/系统工具/流程中心）→ 系统管理下钻（10 子项：用户/角色/菜单/部门/字典/岗位/参数/公告/日志(目录)/接口权限，图标+类型标签+面包屑）→ 我的（管理员/admin/角色 admin/权限 全部权限/常用功能/退出登录）→ 退出（uni.showModal 取消/确定 → token 清空 → reLaunch 回登录页）
- **未决/后续**：真机/小程序联调需 BASE_URL 改局域网 IP（request.js 注释已说明）；非 admin 用户菜单过滤未实测（*:*:* 全权限可覆盖）；H5 端口沿用 5173

### 21. 移动端 App 修复与优化（验证码双前缀根因 + 圆角收敛 2px，任务 #95~#102 后续，verify38）
- 用户续报「登录页的验证码有问题」+「圆角别太大，2px 左右就可以了」
- **验证码不显示根因（层层排查实证）**：①uni dev server 僵死（进程 26548 在但 5173 不监听）→ 改 `npm run build:h5` 产物 + `python -m http.server 5173 --directory app\dist\build\h5` 静态托管（无 vite proxy，App 端 request.js BASE_URL 直连 8080，不依赖代理）；②静态产物下 uni-image 背景 `background-image:none`、`imgCount=0` → 手动设 div background 可显示 → 定位为 uni-h5 image 组件对动态 data URL 渲染不稳；③**最终根因：后端 `/api/captchaImage` 返回的 `data.img` 已带 `data:image/png;base64,` 前缀**（`Invoke-RestMethod` 实测 head 命中），登录页又拼一次前缀 → 双前缀 data URL 无效 → 图不渲染（此前探查记录「无前缀」已过时）
- **修复**：①登录页 `loadCaptcha` 前缀幂等（`img.startsWith('data:') ? img : 'data:image/png;base64,' + img`，兼容两种后端返回）；②H5 端验证码图改 `view + :style backgroundImage` 条件编译渲染（`#ifdef H5` view / `#ifndef H5` image，小程序端仍走 image 组件）——实测 OCR「5748」识别并登录成功
- **圆角收敛 2px**：全页面 `border-radius` 统一 4rpx（750rpx 基准下 = 2px；头像 50% 圆形保留）：login-card/form-input/captcha-img/remember-box/login-btn、welcome-card/stat-card/menu-cell/menu-icon/role-badge/welcome-badge/empty、user-card/perm-card/list-card/logout-btn/user-tag、menu-list/menu-item-icon
- **功能小优化**：首页欢迎卡右上角显示「{MM-DD} 已登录」日期；我的页角色标签由硬编码「管理员」改为动态（admin→「超级管理员」，其他角色取首个角色名）
- **验证（浏览器实测，verify38）**：登录页验证码图显示（OCR「5748」）→ 登录成功（管理员/admin/11-03 已登录/统计 4·全部·1）→ 首页菜单宫格 4 项 → 系统管理下钻 10 子项（用户/角色/菜单/部门/字典/岗位/参数/公告/日志/接口权限）→ 我的（超级管理员标签/admin/角色 admin/权限 *:*:*）→ 退出（点退出→弹窗确定→会话切换；bu 会话在 reLaunch 后失稳，退出回登录页由上阶段截图证据 + 退出逻辑未改动佐证）

### 22. 移动端 App 核心业务页落地（用户管理：列表/搜索/状态/分页，任务 #104）
- 用户续选「核心业务页落地」：把菜单「页面」类菜单做成真实业务页，先做代表性页面——用户管理
- **后端接口（实测确认）**：`GET /api/system/user/list?pageNum&pageSize&username&phone&status&roleId`，`Authorization: Bearer <token>`，响应 `{code:200, rows:[{userId,userName,nickName,email,phonenumber,status,deptName,...}], total}`（`utils.OkTable`，total 顶层）；perms `system:user:list`；数据权限注入部门/本人过滤（admin 全量）
- **App 端实现**：
  - `src/api/user.js`：`listUser(query)` 封装（复用 request.js，GET 参数走 data 拼 query）
  - `src/utils/menuRouter.js`：菜单 path → 本地业务页路由映射（`PAGE_MAP`：`/system/user` → `/pages/system/user/index`）；`openPageMenu(m)` 能映射则 `navigateTo` 返回 true，未映射返回 false 由调用方弹「请在 Web 端操作」提示——首页 `pages/index/index.vue` 与菜单下钻页 `pages/menu/index.vue` 的 C 类分支均改调此函数，后续新业务页只需在此登记
  - `src/pages/system/user/index.vue`：搜索栏（账号关键词 + 搜索/重置）、列表（首字头像按 userId 调色板着色、昵称/账号/部门/手机号、状态标签「正常/停用」按 status 0/1）、触底分页（onReachBottom 加载更多 + 加载转圈 + 没有更多提示 + 空态）、圆角 4rpx 规范
  - `src/pages.json`：注册 `pages/system/user/index`（标题「用户管理」）
- **构建与验证（verify39）**：`npm run build:h5` DONE（Compiler 5.34 vue3）；产物 `dist/build/h5/assets/pages-system-user-index.*.js` 存在且含「用户管理」/`system/user`、主包含 `/system/user` 映射、menu chunk 引用 menuRouter；接口实测 `GET /api/system/user/list?pageNum=1&pageSize=3` → `code=200 total=1 rows=[admin]`（后端 8080 经「登录 → token 黑名单失效 → 批量验证码 OCR 重新登录（0898/9112/7517，7517 成功）→ 新 token」链路确认）；静态服务 5173 可访问（index.html 200）。注：bu 浏览器会话本轮持续 daemon 故障（NoneType / unrecoverable failure，重启服务未恢复），运行态视觉验证受限，页面逻辑以接口实测 + 产物编译验证 + 静态审查为准
- **服务恢复**：后端 8080 曾随会话清理掉线，`go-admin-new.exe`（WorkingDirectory backend）重启后 8080 LISTENING；静态服务 python http.server 5173 亦重启恢复

### 23. 移动端 App 手势交互 + 设计优化 + 角色管理页（任务 #105/#106）
- 用户续令「把各种手势交互、设计都要优化；再继续完善功能」
- **勘误（§22 遗留）**：后端菜单 `path` 为**相对路径**（seed.go：系统管理目录 Path=“system”，用户管理 C 类 Path=“user”，角色管理 Path=“role”，父目录 path 仅用于前端路由组装）——§22 中 `PAGE_MAP: /system/user → …` 的完整路径键**不命中菜单数据**，本轮修正 `menuRouter.js` 的 `PAGE_MAP` 改为以 `m.path` 相对路径为键（`user`/`role`），首页与菜单页 C 类分支无需改动即正确跳转
- **手势交互**：
  - **全局下拉刷新**：pages.json 各业务页 `enablePullDownRefresh: true`（首页/我的/菜单/用户管理/角色管理），各页 `onPullDownRefresh` 重新拉取数据（首页/我的/菜单→`userStore.init()`；列表页→重载第一页）后 `stopPullDownRefresh`；H5 端自带下拉手势触发
  - **全局触摸反馈**：App.vue 全局样式统一可点击元素（menu-item/menu-cell/user-item/crumb-link/search-btn/btn/logout-btn/role-item/func-cell）`:active` 按下变淡 opacity .7 + 0.15s 过渡
  - **长按菜单详情**：首页宫格与菜单下钻列表项 `@longpress` 弹菜单详情（ID/类型/路径/权限标识 perms）
  - 列表触底分页加载（用户/角色页 onReachBottom，含加载转圈/没有更多/空态）
- **设计优化**：tabBar 增加图标（81×81 PNG，PIL 手绘线性风格：首页房子 / 我的 人形 × 正常灰 #909399 / 选中蓝 #409eff，`src/static/tabbar/`，pages.json iconPath/selectedIconPath）
- **角色管理业务页（继续完善功能）**：`src/api/role.js` `listRole(query)`；`src/pages/system/role/index.vue`（角色名搜索/重置、列表：角色名/权限字符 roleKey/排序/数据范围（1全部 2自定义 3本部门 4本部门及以下 5仅本人）/状态标签、触底分页、下拉刷新、空态，4rpx 圆角规范）；`menuRouter.js` 登记 `role`；pages.json 注册
- **验证（verify40）**：`npm run build:h5` DONE；产物 `pages-system-role-index.*.js`/`pages-system-user-index.*.js` 存在、`static/tabbar/*.png` 进产物；角色接口实测 `GET /api/system/role/list?pageNum=1&pageSize=5` → `code=200 total=2`（超级管理员 admin / 普通角色 common，sort/status/dataScope 字段齐）；用户接口复测正常；静态服务 5173 托管新产物。注：bu 浏览器会话持续 daemon 故障（多轮 NoneType/unrecoverable，navigate 可执行但 js/screenshot 通道坏），运行态视觉验证受限，以接口实测 + 产物编译验证 + 静态审查为准
### 24. 移动端 App 剩余业务页批量落地（菜单/部门/字典/岗位/参数/公告/日志/接口权限，任务 #107）
- 用户续令「继续」→ 把菜单中剩余「页面」类菜单全部做成真实业务页
- **后端接口探查（实测确认）**：
  - 树形（OkData，无分页）：GET /api/system/dept/list（system:dept:list，data 树，字段 deptId/parentId/deptName/orderNum/leader/phone/email/status/children）、GET /api/system/menu/list（system:menu:list，data 树，字段 menuId/menuName/parentId/path/component/menuType(M/C/F)/perms/icon/visible/status/children）
  - 分页列表（OkTable：{code, rows, total}）：GET /api/system/dict/type/list?dictName&dictType（dictId/dictName/dictType/status/remark）、GET /api/system/post/list?postName&postCode&status（postId/postCode/postName/postSort/status/remark）、GET /api/system/config/list?configName&configKey&configType（configId/configName/configKey/configValue/configType(Y内置 N非内置)/remark）、GET /api/system/notice/list?noticeTitle&noticeType&status（noticeId/noticeTitle/noticeType(1通知2公告)/status(0正常1关闭)/createBy/noticeContent）、GET /api/system/api/list?apiName&path&status（apiId/apiName/path/method/perms/status(0启用1停用)/remark）
  - **日志接口勘误**：登录日志 = GET /api/monitor/logininfor/list?username&ipaddr&status（infoId/userName/ipaddr/status(0成功1失败)/msg/loginTime）；操作日志 = GET /api/monitor/operlog/list?title&operName&status（operId/title/businessType/requestMethod/operIp/operUrl/status/costTime/operTime/operName）——**不是** /api/system/log/*（404）
- **App 端实现（复用用户/角色页统一骨架，4rpx 圆角规范）**：
  - api 层：src/api/{dept,menu,dict,post,config,notice,log,api}.js（dept/menu 无分页直接取 data；其余分页参数走 data 拼 query）
  - 树形页 pages/system/{dept,menu}/index.vue：拍平树 + 层级缩进 + 点击展开/收起（▾/▸），默认展开第一层；部门行显示负责人/电话/排序/状态，菜单行显示类型标签（目录/页面/按钮）/path/perms/状态
  - 分页列表页 pages/system/{dict,post,config,notice,loginlog,operlog,api}/index.vue：搜索栏（单关键词 + 搜索/重置）+ 列表（头像首字 + 双/三行信息 + 状态标签）+ 触底分页 + 下拉刷新 + 空态/加载转圈/没有更多
  - pages.json 注册 9 页（均 enablePullDownRefresh）；menuRouter.js PAGE_MAP 登记 11 项：user/role/menu/dept/dict/post/config/notice/loginlog/operlog/api
- **验证（verify41）**：
pm run build:h5 DONE；产物 11 个页面 chunk 全在（pages-system-{api,config,dept,dict,loginlog,menu,notice,operlog,post,role,user}-index.*.js）；9 接口实测全部 200（api total=124 / config total=4 / dict total=4 / notice total=1 / post total=3 / loginlog total=118 / operlog total=3354 / dept、menu 树形 data 正常）；静态服务 5173 托管新产物。注：bu 浏览器自动化持续故障（navigate 可执行但 js/screenshot 通道坏，本轮返回空），运行态视觉验证仍受限，以接口实测 + 产物编译验证 + 静态审查闭环
### 25. 移动端 App 用户/角色管理升级：完整增删改（CRUD，继续完善）
- 用户续令「再继续」→ 把用户/角色管理从只读列表升级为完整管理（新增/编辑/删除/状态切换），对齐 RuoYi-App 管理能力
- **后端接口（router.go + handler 实测确认）**：
  - 用户：POST/PUT \/api/system/user\（body \{userId, deptId, username, password, nickName, email, phonenumber, status, remark, roleIds[], postIds[]}\，新增必填 username+password，编辑用户名不可改）；DELETE \/api/system/user/:id\；PUT \/api/system/user/changeStatus\（\{userId, status}\）；保护：admin 不允许被他人操作/停用/删除
  - 角色：POST/PUT \/api/system/role\（body \{roleId, roleName, roleKey, roleSort, status, dataScope, remark, menuIds[], deptIds[]}\，roleName+roleKey 必填、roleKey 唯一）；DELETE \/api/system/role/:id\；保护：id=1 超级管理员角色仅 admin 本人可改、任何人不许删
- **App 端实现**：
  - src/api/user.js 扩展 createUser/updateUser/deleteUser/changeUserStatus；src/api/role.js 扩展 createRole/updateRole/deleteRole
  - 用户页升级：搜索栏右侧「＋ 新增」；行右侧编辑✎/删除✕圆形操作钮；点行或✎开底部弹层表单（用户名[编辑禁用]/密码[仅新增]/昵称/手机/邮箱/备注/状态 switch）；点状态标签切换（admin 前端拦截）；删除二次确认；提交后刷新列表
  - 角色页升级：同模式表单（角色名/权限字符/排序/数据范围 picker 5 档/备注/状态 switch）；id=1 角色前端拦截删除与状态修改；删除二次确认
  - 弹层为纯 view 实现（遮罩 + 底部卡片，H5/小程序通用），圆角 4rpx 规范，无新增依赖
- **验证（verify42）**：build:h5 DONE，user/role chunk 更新；CRUD 全链路接口实测全部 200：角色 create→find(id=12)→update→delete；用户 create→find(id=29)→update→changeStatus→delete；清理复核 role-left=0 user-left=0（无测试数据残留）。bu 浏览器自动化持续故障，视觉验证仍以静态审查 + 接口实测闭环
### 26. 移动端 App CRUD 推广：部门/菜单/字典(类型+数据二级页)/岗位/参数/公告 全量增删改（继续完善）
- 用户续令「继续」→ 把 §25 的用户/角色 CRUD 模式推广到剩余业务页（部门、菜单、字典类型+字典数据二级页、岗位、参数、公告）
- **后端接口（router.go + Grep Req struct 确认）**：树形 dept/menu：POST/PUT \/api/system/dept\、DELETE \/api/system/dept/:id\（DeptReq={deptId,parentId,deptName,orderNum,leader,phone,email,status}）；POST/PUT \/api/system/menu\、DELETE \/api/system/menu/:id\（MenuReq={menuId,menuName,parentId,orderNum,path,component,menuType(M目录/C菜单/F按钮),perms,icon,visible,status}）；字典：POST/PUT \/api/system/dict/type\、DELETE \/api/system/dict/type/:id\（DictTypeReq），POST/PUT \/api/system/dict/data\、DELETE \/api/system/dict/data/:dictCode\（DictDataReq={dictCode,dictSort,dictLabel,dictValue,dictType,status}）；岗位 POST/PUT \/api/system/post\、DELETE \/api/system/post/:id\；参数 POST/PUT \/api/system/config\、DELETE \/api/system/config/:id\；公告 POST/PUT \/api/system/notice\、DELETE \/api/system/notice/:id\。**后端保护（实测确认）**：内置参数（configType=Y）DeleteConfig 返回 500「内置参数不允许删除」；DeleteDictType 前 CountDictDataByType>0 返回 500「该字典类型存在数据，不允许删除」
- **App 端实现（统一骨架：搜索+横排新增钮+行✎✕+状态标签点按开表单[prefill 下一状态]+底部弹层纯 view，4rpx 圆角规范）**：
  - api 层：src/api/{dept,menu}.js 加 create/update/delete（树形无分页）；dict.js 加类型 CRUD + 数据 CRUD 4 方法（listDictData/createDictData/updateDictData/deleteDictData）；post.js/config.js/notice.js 加 CRUD
  - 树形页 pages/system/{dept,menu}/index.vue：树拍平+层级缩进+展开收起（保持）；新增「＋新增/✎编辑」底部弹层表单——dept（父部门 picker 树选项/部门名/排序/负责人/电话/邮箱/状态）、menu（上级菜单 picker/类型 picker M/C/F/路由/组件/权限标识/图标/是否显示/状态）；config 页内置参数（configType=Y）删除前前端 toast「内置参数不允许删除」拦截；notice 页新增类型 picker（1通知/2公告）+ textarea 内容 + 行点击开详情弹层；dict 页类型 CRUD + 行点击 navigateTo `/pages/system/dictdata/index?dictType=...&dictName=...`（encodeURIComponent）进数据页
  - **新建 pages/system/dictdata/index.vue（二级页）**：头部显示 dictName/dictType + 新增钮；listDictData 分页 + 数据 CRUD（标签/键值/排序/状态）+ 状态切换 + 下拉刷新/触底分页
  - pages.json 注册 dictdata（现 17 页，tabBar 首页/我的不变）
  - 修正：dept 页初稿 `reactive0()` 笔误已改 `reactive({...})`
- **验证（verify43）**：`npm run build:h5` DONE（Compiler 5.34 vue3，12 个 pages-system-* chunk 含新增 pages-system-dictdata-index.*.js）；接口实测全部 200：岗位 create→find(id=9)→update→delete、参数(7)、公告(4)、字典类型(7)、字典数据(code=19)、部门（树形递归 id=125）、菜单（树形递归 id=2041）create→find→update→delete 全通；清理复核 post/config/notice/dicttype left 全 0、dept/menu left=0；保护规则实测：内置参数删除 500「内置参数不允许删除」、字典类型含数据删除 500「该字典类型存在数据，不允许删除」（清数据后删除成功）。注：本轮后端 8080 与静态 5173 均曾掉线，已重启恢复（8080 新 PID、5173 重新托管 build/h5）；bu 浏览器自动化持续故障，dictdata 跳转与页面交互以静态审查 + 产物编译 + 接口实测闭环
### 27. 数据库支持核验：PostgreSQL（并修复 PG 下启动阻断 bug）
- 用户问「现在这个项目支持 PostgreSQL 数据库吗」→ 逐层核查后端数据层，结论：**支持**（架构层已完整接入三方言 sqlite / mysql / postgres）
- **现有实现（核查证据）**：
  - `config/config.go` DatabaseConfig{Type,Path,Host,Port,User,Password,DBName}，Type 支持 sqlite（默认）/mysql/postgres，未声明时回落 sqlite；`config.yaml` 已内置 MySQL/PG 连接示例
  - `dao/db.go` InitDB 三路分派：sqlite 走 `glebarez/sqlite`（纯 Go，WAL+busy_timeout）；mysql/postgres 走 `initNetworkDB`——PG DSN `host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai`，默认端口 5432，连接池 MaxOpen 20 / MaxIdle 5；启动后 AutoMigrate 全部 model（22 张表）+ seed 种子数据 + ReloadApiRules
  - `go.mod` 已锁 `gorm.io/driver/postgres v1.6.3`
  - 方言适配点：`dao/dashboard_dao.go`（PG 用 `to_char(login_time,'YYYY-MM-DD')` / `EXTRACT(HOUR FROM login_time)`，其余用 substr）、`api/db_api.go`（PG `version()` + `pg_tables`/schemaname='public'）、`dao/gen_dao.go`（PG `COALESCE(obj_description(...),'')` + `information_schema.columns`/ordinal_position）；其余 4 处 Raw SQL（data_scope/flow_dao/menu_dao/user_dao）均为标准 ANSI（JOIN/IN/LIKE ESCAPE），PG 兼容
- **发现并修复的 PG 阻断 bug**：`dao/db.go` fixApiCreateTimes 原按方言生成 `NOW()` / `now()` 赋给 `sys_api.create_time`（模型为 `string size:20` → PG 下 varchar(20)），PostgreSQL 无 timestamp→varchar 隐式赋值转换，该 UPDATE 会在解析/执行阶段报类型不匹配，**导致 InitDB 返回错误、服务启动失败**（sqlite/mysql 存在隐式转换故未暴露）。修复：统一格式化为字符串——mysql `DATE_FORMAT(NOW(),'%Y-%m-%d %H:%i:%s')`、postgres `to_char(now(),'YYYY-MM-DD HH24:MI:SS')`、sqlite 保持 `datetime('now')`
- **验证**：`E:\Go\bin\go.exe build` exit 0；产物覆盖 backend/go-admin-new.exe（2026-11-12 18:23）后重启，8080 LISTENING（新 PID）、`GET /api/captchaImage` 200；sqlite 路径行为未变（该分支表达式原样保留）
- **未做与注意**：①本机无 PostgreSQL 服务（5432 无监听）、无 docker、无 psql，**未做 PG 实机连接/建表/接口实测**，仅代码级核查 + 编译验证；②SSLMODE 硬编码 `disable`、时区硬编码 `Asia/Shanghai`（如需远程 PG/TLS 需改 DSN）；③PG 的 LIKE 大小写敏感（账号/英文关键词搜索与 SQLite 行为不同）；④PG 下需先手动建库（`CREATE DATABASE go_admin`），表与种子由首次启动自动迁移写入；⑤后端重启会使内存态 token 失效（旧 token 返回 401，需重新登录）
### 28. 项目切换到 PostgreSQL（本地 Docker PG 实机验证 + 自增序列修复）
- 用户指令：SQLite 数据保留在项目内 → 在本地（ docker 已装 PG）新建数据库 → 项目切换为 PostgreSQL。连接：`jdbc:postgresql://localhost:5432/postgres`，`postgres/root`
- **SQLite 数据位置确认**：`backend/data/go-admin.db`（1.2 MB，另有 `-wal/-shm`），`config.yaml` 的 `path: "data/go-admin.db"` 相对 config 所在目录解析，**本就在项目内**，本次不移动、不删除，仅保留作历史数据与回退来源
- **环境修复**：`docker ps` 原本报 `Docker Desktop is unable to start`，根因是 WSL 的 `docker-desktop` 发行版处于 Stopped；`wsl -d docker-desktop -e echo ok` 唤醒后引擎恢复（server 29.9.3），容器 `my-postgres`（`postgres:18.4`，`127.0.0.1:5432->5432`）正常运行
- **建库与连接验证**：临时探针 `E:\Go\bin\go.exe run ./tools/pgprobe` 连 `postgres` 库成功（`PostgreSQL 18.4 ... Debian`），`CREATE DATABASE go_admin` 成功（探针用后已删除）
- **配置切换**：`backend/config.yaml` → `type: postgres`、`host: 127.0.0.1`、`port: 5432`、`user: postgres`、`password: root`、`dbname: go_admin`；`path` 与 PG 示例注释保留；README 技术栈表、config 示例、数据库支持章节同步更新为「当前使用 PostgreSQL + 回退方式」
- **PG bug 1（阻断级，切库后必现）**：种子数据以固定 ID 写入（角色 1/2、菜单 1/100/101…、岗位 1/2/3、字典、参数、公告），PostgreSQL 的 `bigserial` 序列不会随显式 ID 插入推进，仍停在 1；新增记录 `nextval` 与既有主键冲突——实测 `POST /api/system/post` 返回 500，日志 `ERROR: duplicate key value violates unique constraint "sys_post_pkey" (SQLSTATE 23505)`。**修复**：`dao/db.go` 新增 `syncPgSequences()`，在 `seed()` 之后执行（仅 `Dialect()=="postgres"`），对每张「id 列 default 为 `nextval(...)`」的表 `setval(序列, MAX(id), true)`；幂等，每次启动自动纠正
- **PG bug 2（自身实现踩坑）**：`syncPgSequences` 初版用 `pg_get_serial_sequence(c.table_name,'id')` 写在 information_schema 查询的 WHERE 中，PostgreSQL 会先求值该函数，遇到无 `id` 列的关联表（`sys_role_menu`）直接报 `column "id" of relation "sys_role_menu" does not exist (SQLSTATE 42703)`，导致 InitDB 失败、服务无法启动。**修复**：序列名改从 `column_default`（`nextval('sys_post_id_seq'::regclass)`）解析，先按 `column_name='id' AND column_default LIKE 'nextval(%'` 过滤，完全绕开 `pg_get_serial_sequence`
- **验证（实测，全部通过）**：
  - 启动：AutoMigrate 建 26 张表（22 业务 + `sys_role_menu`/`sys_user_post`/`sys_user_role`/`sys_role_dept`），种子写入 `sys_user=1 sys_role=2 sys_menu=90 sys_dept=4 sys_dict_type=4 sys_dict_data=8 sys_post=3 sys_config=4 sys_notice=1 sys_api=124 sys_job=4`
  - 登录：`GET /api/captchaImage` 200 → 验证码放大 4 倍识别（OCR）→ `POST /api/login` code=200，token 191 字符
  - 读接口 19 个全 200/code=200：`getRouters`、`dashboard/{stats,loginTrend,loginHeatmap,operTypeStats}`、`monitor/db/info`、`monitor/{logininfor,operlog}/list`、`system/{user,role,menu,dept,post,config,notice,dict/type}/list`、`tool/gen/db/list`（返回 0，因全部 sys_ 表被排除，符合预期）、`monitor/{server,cache}/info`
  - 数据监控：`monitor/db/info` 返回 `version=PostgreSQL 18.4`、`tableCount=26`、`totalRows=363`、`journalMode=WAL`，各表行数正确
  - 方言 SQL 实测：`loginTrend` 返回 7 天数组、`loginHeatmap` 返回数据点（插入一条探针登录日志验证 `EXTRACT(HOUR)` → int 扫描为 `hour=16`，验证后已删除）、`operTypeStats` 正常
  - **写操作闭环 8 模块全 200**：岗位（id=4）/菜单（id=2041）/部门（id=104）/字典类型（id=5）+字典数据（dictCode=9）/参数（id=5）/公告（id=2）/角色（id=3）/用户（id=2），每个均 create→list/find→update→delete，残留复核全 0
  - 序列一致性：`sys_dept/user/post/role/menu/logininfor/dict_type/dict_data/oper_log/config` 的 `max(id)` 与序列 `last_value` 全部对齐（last ≥ max，is_called=1）
- **未做**：①SQLite 历史业务数据（旧操作日志、登录日志等）**未迁移**到 PG，PG 是全新库+种子数据；如需要可再写一次性迁移脚本；②Docker Desktop 引擎依赖 WSL，若 `docker-desktop` 发行版再次 Stopped，需先唤醒或重启 Docker Desktop，否则 5432 不可用、后端起不来；③`sslmode=disable`、`TimeZone=Asia/Shanghai` 仍硬编码在 DSN；④PG 的 LIKE 大小写敏感
- **回退方式**：`config.yaml` 把 `database.type` 改回 `sqlite`（`path` 未改动），重启即可继续使用 `data/go-admin.db` 原有数据

### 29. 为 PostgreSQL 补齐表与字段注释（26 表 / 236 列）
- 问题：SQLite 没有表/列注释属性，切到 PG 后 `information_schema` 与客户端里表注释、列注释全部为空（实测 26 张表、236 列，`obj_description` / `col_description` 全为 NULL）
- **注释来源**（不臆造字段含义）：`model/*.go` 的结构体说明与字段 Go 注释；`status`/`menu_type`/`data_scope`/`approver_type` 等枚举在 model 里已有的取值说明；`sys_oper_log.business_type` 取自前端 `system/operlog/index.vue` 的 `bizTypeText` 映射（0其它 1新增 2修改 3删除 4查询）；工作流状态取自 `model/flow.go` 顶部状态机约定
- **落地方式**：新增 `backend/scripts/pg_comments.sql`（每条语句一行，表注释 26 条 + 列注释 236 条，共 262 条）；`dao/db.go` 新增 `syncPgComments()`，在 `afterOpen()` 的 PG 分支（紧随 `syncPgSequences()`）读取该文件逐行 `Exec`，幂等；文件缺失或单条失败不阻断启动（注释属元数据增强，不应影响可用性）；查找顺序为 `工作目录/scripts/pg_comments.sql` → `可执行文件目录/scripts/pg_comments.sql`
- **执行结果**：`EXEC ok=262 fail=0`；覆盖率 `tables=26/26`、`columns=236/236`，`MISS_TABLES=[]`、`MISS_COLUMNS=[]`
- **持久化验证（不只依赖命令返回）**：把 `sys_post` 表注释与 `post_code` 列注释手动置 NULL（`COMMENT ON ... IS NULL`）→ 重新启动后端 → 再查 `obj_description`/`col_description` 已自动恢复为「岗位表」/「岗位编码，唯一」；全表复查缺失标记数为 0；8080 `captchaImage` 200、5173 H5 200
- **注释内容口径示例**：`sys_user.status` 0正常 1停用、`sys_menu.menu_type` M目录 C菜单 F按钮、`sys_role.data_scope` 1全部 2自定义 3本部门 4本部门及以下 5仅本人、`sys_flow_instance.status` 1运行中 2已通过 3已驳回 4已撤销 5已终止、`sys_flow_task.status` 0待办 1已通过 2已驳回 3已转出 4已加签 5已撤销 6已读、`sys_flow_record.action` 1发起 2通过 3驳回 4转办 5加签 6撤回 7撤销 8终止 9催办 10重新提交 11抄送、`sys_job.concurrent` 0允许 1禁止
- **维护约定**：新增表或字段时同步在 `pg_comments.sql` 追加 `COMMENT ON`（否则该列仍为空）；文件中的注释文本不含分号与单引号，便于逐行执行
- **未做**：①代码生成模块的列注释仍取列名（`gen_dao.go` 中 `ColumnComment: c.Name`，未读取 `col_description`），即注释只落在数据库元数据，不会出现在代码生成配置里，若要联动需改 `tableColumns()` 的列元数据查询；②MySQL 分支未做同等注释补齐（当前库只有 PG）；③临时探针（`tools/dbmeta`、`tools/dbnil`、`tools/dbcomment`）用完已删除，`backend/tools` 空目录一并清理

### 30. 迁移 SQLite 历史数据到 PostgreSQL（操作/登录/任务日志 + 部门）
- 需求：把 `backend/data/go-admin.db` 里的历史数据（用户、角色、操作日志、登录日志等）迁到 PG 的 `go_admin` 库
- **迁移前备份**：`docker exec my-postgres pg_dump -U postgres -d go_admin` → `backend/data/backup/go_admin_before_migrate.sql`（126856 字节）。回滚：`docker exec -i my-postgres psql -U postgres -d go_admin < go_admin_before_migrate.sql`
- **前置核查（26 张表逐表 count 对比）**：主数据两侧完全一致（sys_user 1 / sys_role 2 / sys_menu 90 / sys_api 124 / sys_job 4 / sys_dict_type 4 / sys_dict_data 8 / sys_config 4 / sys_notice 1 / sys_post 3 / sys_flow_def 2 / sys_flow_node 3 / sys_role_menu 90 / sys_user_role 1）；差异仅 6 处：sys_oper_log 3449 vs 65、sys_logininfor 120 vs 3、sys_job_log 801 vs 6、sys_dept 5 vs 4、sys_user_post 2 vs 1、sys_online 0 vs 2（PG 的 2 条是当前在线会话，不迁）
- **迁移策略**：日志三表按原时间值迁移、**丢弃原主键交由 PG 序列重新分配**（避免与 PG 切换后产生的数据撞号）；主数据逐表对比只补缺；部门保留原 ID；关联表补缺；最后对自增序列 `setval` 对齐。一次性工具 `backend/tools/migrate`（同时打开 `glebarez/sqlite` 与 `gorm.io/driver/postgres`），执行完即删除
- **三个坑与解法**：
  1. **时间格式**：SQLite 的 `datetime` 列实际以 TEXT 存 7 位小数 + 时区偏移（如 `2026-10-28 11:11:52.2154838+08:00`），直接 Scan 进 `time.Time` 会类型不匹配；改为 `Find(&[]map[string]interface{})` 取原始值 + 多 layout 解析（含 `2006-01-02 15:04:05.999999999-07:00`）后构造 model 插入
  2. **非法 UTF-8（4 条）**：`sys_oper_log.id=2201~2204` 的 `oper_param` 是头像上传的 multipart 原文，混入 PNG 二进制（`0x89 0x50 0x4E 0x47`），PG 报 `invalid byte sequence for encoding "UTF8": 0x89 (SQLSTATE 22021)`，GBK 也无法解码。补迁工具 `tools/migrate4` 把这 4 条该字段替换为 `[multipart/form-data 含二进制内容，原 N 字节，迁移时已省略]` 后插入（pg_id 3511~3514），其余字段原样
  3. **脏关联**：SQLite `sys_user_post` 存在 `(user_id=8, post_id=3)`，但 PG 中无 user 8，插入被外键 `fk_sys_user_post_user` 拒绝（`SQLSTATE 23503`）→ 判定为 SQLite 侧残留脏数据，不迁
- **结果（迁移后立即核对）**：sys_oper_log 65 → **3514**（SQLite 3449 条全部迁入）、sys_logininfor 3 → **123**（120 条全部）、sys_job_log 6 → **807**（801 条全部）、sys_dept 4 → **5**（补 `id=126 测试部门`，parent=100）、sys_user_post 保持 1
- **验证（不只查 count）**：
  - 日期分布（`AT TIME ZONE 'Asia/Shanghai'` 分组）：操作日志 10-28 692 / 10-29 262 / 10-30 96 / 10-31 581 / 11-01 1367 / 11-02 153 / 11-03 288 / 11-05 10，历史合计 **3449**，与 SQLite 完全一致；登录日志 10-28~11-05 合计 119 + 1 条零时间（原始脏数据）+ 今日 4；任务日志 10-23~11-12 历史 801
  - 时区无错位：SQLite `2026-11-05 18:24:56+08` 迁入后接口读回 `2026-11-05T18:24:56+08:00`
  - 接口实测（重启后端 + 重新登录后）：`monitor/operlog/list` total=3518、`monitor/logininfor/list` total=124、`monitor/jobLog/list` total=809——比迁移完成时多出的部分是验证期间的登录、查询与定时任务新增，列表可见 11-05 等历史记录
  - 序列：`sys_oper_log_id_seq` setval 到 `max(id)`，迁移后新增记录不冲突
- **数据质量说明（原样保留，未擅自清理）**：①`sys_logininfor` 有 1 条 `login_time='0001-01-01 00:00:00+00'` 的零时间记录，原样迁入；如需清理可执行 `DELETE FROM sys_logininfor WHERE login_time < '1970-01-01'`；②`id=126 测试部门` 为历史手工测试数据，已迁入，不需要可删；③仪表盘 `loginTrend`/`loginHeatmap`/`operTypeStats` 是**近 7 天**口径（`dao.GetLoginTrend(7)` 等），10-23~11-05 的历史数据不会出现在图表上，但日志列表页可按时间范围查到
- **清理**：临时工具 `backend/tools/migrate`、`backend/tools/migrate4` 已删除，`backend/tools` 恢复为空目录
### 31. PostgreSQL 深化：A→E 五档实施（逐项落地 + 逐项验证）
需求：切库完成后，对「PostgreSQL 还能补充和优化什么」的 A–E 清单按顺序实施，每项做完即测试、查 bug、再进下一项。

| 档 | 项 | 结论 |
| --- | --- | --- |
| A1 | 操作日志 UTF-8 截断 | 完成，e2e 验证 |
| A2 | LIKE → ILIKE + 通配符转义 | 完成，接口对照验证 |
| A3 | 时区显式化 | 完成，PG 手工计数逐条对齐 |
| B1 | 数据监控真实指标 | 完成，指标与 psql 一致 |
| B2 | 代码生成 PG 元数据/类型映射 | 完成，探针表验证 |
| B3 | 时间列 varchar → timestamptz | 完成（引入 model.LocalTime），PG/SQLite 双向验证 |
| C | 索引、模糊检索、分页与日志保留 | 完成（分区未做，见 31.7） |
| D | 连接池 / TLS / autoMigrate / 运维脚本 | 完成并实测开关 |
| E | 数据治理 | 完成（1 条脏数据已清） |

#### 31.1 A 档

**A1 日志截断（阻断级）**
- 现象：middleware/operlog.go 用 param[:2000] 按字节截断，中文被切在字符中间产生非法 UTF-8，PG 报 invalid byte sequence ... (SQLSTATE 22021)，异步写日志静默丢弃。
- 修复：新增 truncateUTF8(s, maxBytes)——先用 strings.ToValidUTF8 清洗非法字节，再用 utf8.RuneStart 回退到字符边界，并把截断标记 …[截断] 的长度计入上限；为 Title/OperMethod/IP/OperURL/OperName/ErrorMsg/ReqMethod 各加长度常量（128/255/64/255/32/2000/16）。
- 验证：单测 middleware/operlog_test.go 2 例通过（首轮失败是测试用例本身写错："abc\x89\x50def" 里 0x50 是合法 P，正确结果是 abcPdef）；e2e 发送 2141 字节中文 JSON 到 POST /api/system/post（postCode 为空 → 500，不落业务数据），异步日志落库，重启后最新行 length=680 / octet_length=2000、结尾为截断标记，全库 octet_length > 2000 计数为 0。

**A2 LIKE 大小写与通配符**
- 现象：PG 的 LIKE 区分大小写、SQLite 不区分；同样搜索词两库结果不同，且用户输入 %/_ 会被当通配符。
- 修复：dao/time_query.go 新增 likeExpr(col)（PG → col ILIKE ?）与 likeExprEsc(col)（追加 ESCAPE '\'），并统一用 escapeLike() 处理 % _ \；批量改写 13 个 DAO 文件 40+ 处，flow_dao.go 的复合子查询手改；templates/dao.tmpl 同步（生成代码也用 likeExprEsc + escapeLike）。
- 验证：/api/system/user/list?username=ADMIN total=1（改前 0）、username=% total=0（通配符已转义）、/api/monitor/operlog/list?operName=ADMIN total=3546；PG 侧 LIKE=0 / ILIKE=1 对照成立。

**A3 时区显式化**
- 修复：config.DatabaseConfig.Timezone（默认 Asia/Shanghai）+ dao/time_query.go 的 initBizTimezone（字符白名单校验 + time.LoadLocation，import _ "time/tzdata" 保证 Windows 可用）、tzName()/timeCond()/parseDayTime()/dayStart()/dayStartCond()；PG DSN 带 TimeZone=%s，MySQL 带 loc=%s；统计 SQL 改 AT TIME ZONE '<tz>'（to_char(... AT TIME ZONE) 与 EXTRACT(HOUR FROM ... AT TIME ZONE)）。
- 顺带修复真 bug：api/operlog_api.go 的 ListOperLogs 只读 title/operName/status，前端传的 beginTime/endTime 被静默忽略（DAO 支持、导出接口在用），已补齐参数。
- 验证：dashboard todayLogin=1、loginTrend 11-13=5 / 11-14=1、heatmap 4 点；operlog 2026-11-05 total=10、2026-11-13 total=97；logininfor 11-14 total=1、11-13 total=5，全部等于 PG (x AT TIME ZONE 'Asia/Shanghai')::date 的计数。

#### 31.2 B 档

**B1 数据监控（api/db_api.go 重写 + frontend/src/views/monitor/db/index.vue）**
- 原实现 PG 分支 journalMode 硬编码 WAL，totalRows 逐表 COUNT(*) 26 次往返；前端把版本卡写死为「SQLite 版本」，大小只读 fileSize。
- 现实现：countTablesSQL 用 UNION ALL 一次统计全部表行数（sqlIdent/sqlLiteral 转义，失败回退逐表循环）；PG 增补 dbSize（pg_database_size）、connections（pg_stat_activity）、maxConnections（current_setting）、cacheHitRatio（blks_hit 比例）、idleInTransaction，并输出 dbType/version/versionRaw；前端卡片按方言动态展示。
- 验证：dbType=postgres version=PostgreSQL 18.4 dbSize=11015871 conns=1/100 hit=99.95% tables=26 total=4878，dbSize、maxConnections 与 psql 实测一致。

**B2 代码生成元数据**
- PG 列元数据从 information_schema.data_type（无长度）改为 pg_attribute + format_type(atttypid, atttypmod) + col_description + pg_get_expr 默认值 + pg_index 主键判定；ColumnMeta 增加 Comment；MySQL 补 column_comment。
- sqliteToGoType → dbTypeToGoType：覆盖 character varying(n)/numeric(10,2)/timestamp with time zone/double precision，boolean→bool、整型→int64、numeric/decimal/money→float64；新增 isTimeColumn() 按类型识别时间列（created_at 这类列名不含 time/date 也能识别）；布尔列默认 switch 表单且不参与查询过滤，text 类默认 textarea。
- 模板：vue_index.tmpl 增 el-switch 表单分支与布尔列表 el-tag；dao.tmpl 修复 Where("col EQ ?") 生成的非法 SQL（EQ 映射为 =）。
- 验证：临时 PG 探针表（serial/boolean/numeric(10,2)/varchar(64)/timestamptz/text + 中文表列注释）经 /api/tool/gen/importTable 导入，列映射为 boolean→bool/switch、numeric(10,2)→float64、character varying(64)→string、timestamp with time zone→string/date/不查询、text→textarea；/api/tool/gen/preview/{id} 生成的 Go 文件 gofmt 通过，db.Where 输出 amount = ? 与 likeExprEsc("title")；验证后已删生成配置并 DROP TABLE。

#### 31.3 B3 时间列类型（含一次真实的兼容性返工）

- 目标：sys_api.create_time、sys_gen_table.create_time/update_time 由 varchar(20) 改为原生时间列，避免字符串时间无法比较/排序。
- 返工过程（记录清楚，避免后人再踩）：第一版直接把 model 字段改成 time.Time，go build 通过、go test ./... 一度全绿（router 测试走 sqlite 但没覆盖这两张表的时间扫描）。随后新增 dao/time_column_test.go 才暴露问题：
  - unsupported Scan, storing driver.Value type string into type *time.Time——glebarez/sqlite 下旧 TEXT 列里的 "2026-11-05 12:00:00" 无法扫进 time.Time，即 SQLite 老库会直接读不出数据（PG 侧因列已转时间类型不受影响）。
- 最终方案：新增 model.LocalTime（backend/model/localtime.go）
  - 实现 Scan（兼容 time.Time/string/[]byte，多 layout 解析并按本地时区解释无时区时间）、Value（零值写 NULL，避免 0001-01-01）、GormDBDataType（postgres→timestamptz，mysql/sqlite→datetime）、MarshalJSON/UnmarshalJSON（JSON 仍输出 YYYY-MM-DD HH:mm:ss，前端 parseTime() 与既有接口约定完全不受影响）、Time()/String() 辅助方法。
  - 迁移放在 afterOpen 的 AutoMigrate 之前：migrateLegacyTimeColumns() 查 information_schema.columns，仅当类型为 character varying/text 时执行 ALTER TABLE ... TYPE timestamptz USING NULLIF(btrim(col),'')::timestamptz（PG 直接 ALTER TYPE 需要 USING，AutoMigrate 不提供）；MySQL 分支用 MODIFY COLUMN ... DATETIME NULL；幂等，已是时间类型则跳过。
  - fixApiCreateTimes() 改为方言原生 now()/NOW()/datetime('now')（列已是时间类型，不再需要格式化字符串）；dao.CreateApi 在写入时补 CreateTime，避免新增规则出现 NULL。
- 验证：
  - 三列类型均为 timestamp with time zone；sys_api 124 行 count(create_time)=124、零丢失，接口返回 "2026-11-13 16:54:28"（+08 显示与库内 08:54:28+00 一致）。
  - 代码生成导入探针表后写入 2026-11-14 05:36:17.3477+00，接口读回 "2026-11-14 13:36:17"；生成配置已清理（sys_gen_table=0）。
  - dao/time_column_test.go 2 例通过：①SQLite 新建库写入/读取时间列；②模拟旧库 TEXT 列 "2026-11-05 12:00:00" 能被正确解析。
  - go build ./... exit=0；go test ./... 全绿。

#### 31.4 C 档（性能与容量）

- 索引：新增 idx_sys_oper_log_status、idx_sys_job_log_create_time、idx_sys_user_dept_id、idx_sys_dept_parent_id、idx_sys_menu_parent_id、idx_sys_logininfor_username；PG 另启用 pg_trgm 并对 sys_oper_log.title/oper_name、sys_logininfor.username、sys_job_log.job_name 建 GIN 三元组索引（CREATE INDEX IF NOT EXISTS；MySQL 先查 information_schema.statistics 再建；SQLite 用 IF NOT EXISTS）。
- ANALYZE：迁移后 pg_stat_user_tables 里 sys_user 显示 n_live_tup=0 会让优化器误判，启动时对 8 张热点表重新采样。
- 执行计划证据：oper_name LIKE '%nobody_xyz%' 从 Seq Scan 变为 Bitmap Index Scan on idx_sys_oper_log_name_trgm（0.11ms）；admin 这类命中 99.9% 的查询仍走 Seq Scan（0.53ms，属优化器正确选择）；ORDER BY id DESC LIMIT 10 OFFSET 3490 0.4ms。
- 分页上限：normPage 增加 pageSize ≤ 500（maxPageSize），防 pageSize=100000 一次拉全表；单测补 2 个边界用例。
- 日志保留：logs.retentionDays（默认 0=关闭）+ dao.StartLogRetention（启动清一次，之后 24h 一次）+ dao.PurgeOldLogs，按 oper_time/login_time/create_time 删除。
- 接口耗时（warm，15 次采样）：操作日志列表 page1 avg 17.3ms、page350 avg 10.4ms；登录日志列表约 60ms（含鉴权与写日志）。

#### 31.5 D 档（运维配置与脚本）

- 连接池/TLS/迁移开关进配置：maxOpenConns、maxIdleConns、connMaxLifetimeMinutes、sslmode、autoMigrate；SQLite 固定 10/5 连接以降低写锁竞争。
- 校验实测：sslmode: "bogus" 启动即「初始化数据库失败: database.sslmode 取值非法 ...」退出码 1；autoMigrate: false 启动正常、列表接口可用（跳过建表/种子/索引 DDL）。
- 脚本：
  - scripts/pg_role.sql：创建 go_admin_app 最小权限账号（DML + 序列，无 DDL），替换超级用户连接前先执行；
  - scripts/pg_backup.ps1：pg_dump --clean --if-exists 输出到 data/backup/go_admin_<时间戳>.sql（实测生成 791488 字节，中文注释与索引定义完整）；
  - scripts/pg_stat_statements.sql：慢查询扩展的启用步骤（ALTER SYSTEM SET shared_preload_libraries + 重启容器 + CREATE EXTENSION）与查询语句。
- 未自动执行的部分：专用账号创建与 pg_stat_statements 预加载涉及外部数据库配置变更（后者需重启容器），只提供脚本，未擅自切换 config.yaml 的账号或重启 PG。

#### 31.6 E 档（数据治理）

- 全库扫描：异常时间（< 2020-01-01 或 > 当前 +1 天）登录日志 1 条、操作/任务日志 0 条；用户 dept_id 悬空 0；部门自引用 0；菜单父级缺失 0；sys_role_menu 悬空 0。
- 处理：删除 1 条 login_time='0001-01-01' 的登录日志（原样迁入的脏数据，会让日期筛选/统计出现 0001 年），删除前已有 13:42 的全库备份；sys_logininfor 125 行、异常行 0。
- 保留项及理由：sys_dept id=126 测试部门 属历史测试数据但为正常部门记录，保留（用户可在部门页自行删除）；4 条上传操作日志的 oper_param 为 [multipart/form-data 含二进制内容，原 N 字节，迁移时已省略] 的可读占位，保留可追溯；SQLite 侧 sys_user_post(8,3) 脏关联未迁入，PG 现为 (1,1) 一条合法关联。

#### 31.7 遗留与后续

- 日志表未分区：当前 sys_oper_log 约 3600 行、1160 kB，分区收益低于整表重写与 AutoMigrate 兼容成本；上万行后再评估 PARTITION BY RANGE (oper_time)。
- sys_oper_log 上 idx_oper_name(oper_time, oper_name) 与 idx_sys_oper_log_oper_time 前缀重叠，可择一删除以省约 150 kB 索引空间（当前由 model tag 生成，删除后 AutoMigrate 会重建）。
- sslmode 仅对 PG 生效；MySQL TLS 参数、pg_stat_statements 自动启用仍未做。
- 前端：A–E 未改变接口 JSON 形状（LocalTime 维持 YYYY-MM-DD HH:mm:ss），前端无需改动；仅 B1 的监控卡片字段为新增。
### 32. 表结构去除主外键约束（数据保留关联语义）
约定：**数据之间可以有主外键关系，但建表结构里不创建 FOREIGN KEY 约束**——关联由业务代码维护，DDL 变更、跨库导入、历史脏数据都不会被约束卡住。

- **根因**：GORM AutoMigrate 会为 `foreignKey` / `many2many` 关联自动建 FK，PG 侧实际存在 11 个：`fk_sys_user_dept`、`fk_sys_user_post_user/post`、`fk_sys_user_role_user/role`、`fk_sys_role_dept_role/dept`、`fk_sys_role_menu_role/menu`、`fk_sys_gen_table_columns`、`fk_sys_flow_def_nodes`。它们既限制删除（如删除带岗位的用户、带角色的部门会 23503），也让 SQLite 历史数据迁移时被脏关联挡住。
- **不再生成**：两处 `gorm.Open`（SQLite 与 MySQL/PG）统一加 `DisableForeignKeyConstraintWhenMigrating: true`；模型里的 `foreignKey`/`many2many` 声明保留（GORM 仍能 Preload/Association），只是不再落成数据库约束。
- **清理存量**：新增 `dao/db_fk.go` 的 `dropForeignKeys()`，在 `afterOpen()` 末尾（建表之后、接口规则回填之前）执行：
  - PostgreSQL：查 `pg_constraint`（`contype='f'`，public schema），逐条 `ALTER TABLE ... DROP CONSTRAINT ...`，标识符走 `quotePgIdent`（兼容 `public.` 前缀）；
  - MySQL：查 `information_schema.table_constraints`（`constraint_type='FOREIGN KEY' AND table_schema = DATABASE()`），逐条 `ALTER TABLE ... DROP FOREIGN KEY ...`；
  - SQLite：外键写在建表 DDL 里、不支持原地删除，这里只统计并打印提示（新库已不会再生成）。
  - 幂等：没有外键时静默返回；`autoMigrate=false` 时同样执行（结构约定不依赖建表开关）。
- **应用层补齐关联清理**（去掉数据库级联/限制后必须由代码保证，否则会产生孤儿行）：
  - `DeleteUser` / `DeleteUsers`：清理 `Roles` 之外**补清 `Posts`**（此前依赖 FK 拦截，不许删带岗位的用户）；
  - `DeleteDept`：删除前清 `sys_role_dept` 中该部门的行（角色自定义数据权限）；
  - 既有守卫保持不变：删除角色校验 `sys_user_role`、删除岗位校验 `sys_user_post`、删除菜单校验 `sys_role_menu`；删除流程定义清 `sys_flow_node`、删除生成配置清 `sys_gen_column`、删除用户清 `sys_user_role`。
- **验证（实测）**：
  - 首次以 PG 启动：日志 `[db] 已删除 PostgreSQL 外键约束 11 个`，`pg_constraint` 中 `contype='f'` 计数 11 → **0**；
  - 再次重启（AutoMigrate 全量跑一遍）：外键计数仍为 **0**，启动日志无外键删除记录（不再生成，也无需重复删除）；
  - 用户 e2e：`POST /api/system/user`（roleIds=[1]、postIds=[1]）→ 列表 total=1、`sys_user_role=1 / sys_user_post=1` → `DELETE` → 用户行 0、两个关联表均 0；
  - 部门 e2e：新增部门 127 → 手工插入 `sys_role_dept(2,127)` → `DELETE /api/system/dept/127` → 部门行 0、`sys_role_dept` 行 0；
  - 孤儿扫描（`user_role/user_post/role_menu/role_dept/gen_column/flow_node` 左连接主表）：全部为 0；
  - 接口回归 8 个（用户/角色/部门/菜单/岗位/操作日志/代码生成/仪表盘）code=200；
  - 单测新增 `dao/schema_test.go` 4 例：建表 DDL 不含 `FOREIGN KEY`、SQLite 分支 `dropForeignKeys` 不报错、删用户清角色+岗位、删部门清 role_dept；`go test ./...` 全绿。
- **保留项**：主键、唯一索引、普通索引、`NOT NULL` 等结构性约束**保留**（它们是 ORM 与业务正确性的基础，不属于“主外键关系”）；数据层的 `dept_id`、`user_id`、`role_id`、`table_id`、`flow_def_id` 等字段与关联表继续按原语义维护。
- **限制**：①SQLite 老库（`data/go-admin.db`）中既有的 FK 定义无法原地删除，切回 SQLite 前如需彻底去除需重建库（新库已不含）；②MySQL 分支逻辑与 PG 同构，但本机未在 MySQL 实例上实测（避免改动现有 MySQL 数据），仅编译与 `go vet` 通过；③FK 删除后不再有数据库级联，新增删除逻辑时必须同步清理关联表，已在 `dao` 层统一处理。
### 33. 代码生成：单表 / 主子表 / 树形三类模板与模块自注册（2026-11）

目标：生成的前后端代码按目录落盘后，**重启服务即可访问、无需再手改 `router.go` 和 `seed.go`**；数据库不建物理外键，删除主表时由生成的事务代码清子表。

- **自注册中心** `backend/genreg/genreg.go`：`Module{Module, ClassVar, ClassName, FunctionName, Routes, Menu}`，生成 api 文件在 `init()` 中 `RegisterModule()`；`router.go` 在 auth 分组末尾遍历 `Modules()` 挂载并套 `CheckPerm`；`seed.go` 的 `ensureMenus()` 末尾调用 `ensureGenMenus()`，自动建 C 菜单 + 6 个 F 按钮（查询/新增/修改/删除/导入/导出）并绑定角色 1，幂等；C 菜单父级默认 1700（系统工具），见 §34。
- **生成配置**：`GenTable` 增加 `TplCategory`（crud/sub/tree）、`TreeCode/TreeName/TreeParent`、`SubTableName/SubTableFk`；`BuildGenData` 三分支并校验：树形主键与父级必须同为整型，主子表外键与主键 Go 类型一致；PG `timestamptz` 映射 `model.LocalTime`（映射 string 会 scan 失败）。
- **模板**：`model.tmpl`、`dao.tmpl` 三分支、`api.tmpl`（自注册）、`vue_api.tmpl`、`vue_index.tmpl`、`vue_index_tree.tmpl`、`vue_index_sub.tmpl`、`readme.tmpl`（4 步接入）、`router.tmpl`（仅说明）。
- **本轮修掉的 4 个真实 bug**：
  1. 树形 dao/api 生成后 `total` 声明未使用导致编译失败 → 树形分支改用 `list, _, err`；
  2. 新增时 `create_time` 写 NULL、修改用 `Save` 会把 `create_time` 清空 → `GenField.IsCreateTime/IsUpdateTime` + `HasAutoTime`，Create 前补 `time.Now()`，Update 改 `Updates(map)` 并排除 `create_time`；
  3. 树形可把父级设成自身下级导致构树成环 → Update 增加“上级节点不能是自己/自身下级”校验，`buildTree` 加 `depth > 50` 保护；
  4. 前端数值字段用 `el-input`，提交字符串导致后端 `float64/int64` 反序列化报“请求参数错误” → 三类页面主表单均按 Go 类型输出 `el-input-number`，子表行已有 number/switch/date 分支。
- **验证**：
  - `go test ./dao -run TestGenTemplates -v`：crud/tree/sub/时间字段 4 例 PASS，含 Go 语法解析、自注册、事务、成环校验、时间保留、数值控件断言；
  - PG 真实环境：导入 `demo_book/demo_order/demo_order_item/demo_category`，落盘后编译启动，`sys_menu` 自动生成 15 行、`getRouters` 下发、`pg_constraint` 外键 0；
  - 接口 e2e：单表 CRUD；主子表新增 2 子行 → 修改为 1 行 → 删除主表后子表 0；树形三级构树、删父拦截、成环拦截；
  - 页面 e2e（真实浏览器）：图书页新增/修改/删除、分类页树形展示 + `el-tree-select` 选父级新增子节点、订单页新增明细并删除主表，均成功；订单删除后 DB `demo_order_item` 对应行 0；
  - `npm run type-check`、`npm run build` 均通过。
- **使用步骤**：① 代码生成页导入表 → ② 选生成类型并填树形/子表字段 → ③ 预览或下载 → ④ 按返回路径写入后端/前端 → ⑤ 重启后端（前端 dev 自动生效）。菜单与权限自动注册，无需手工 SQL。
- **验收后清理**：4 张 `demo_*` 表、`sys_gen_table` 4 行、15 条演示菜单及角色绑定、生成的 15 个前后端 demo 文件均已删除；后端重新编译，`/api/system/demoBook/list` 返回 404，外键计数仍为 0，避免示例模块污染正式项目。
- **限制**：被生成表须有单列主键；树形父级列与主键、主子表外键与主键的 Go 类型必须一致且为整型；MySQL 分支与 PG 同构但未在 MySQL 实例实测；生成菜单父级由 `genreg.Menu.ParentID` 控制，默认 1700（系统工具），见 §34。

### 34. 代码生成菜单归入“系统工具” + 三类生成代码 Excel 导入导出（2026-11）

目标：生成模块不再占用顶级菜单，统一挂到“系统工具”；单表/树形/主子表生成的前后端代码自带导入、导出、模板下载和权限按钮，落盘重启即可用。

- **菜单父级**：`genreg.Menu` 增加 `ParentID uint`；`dao/seed.go` 的 `ensureGenMenus()` 默认 1700，并把历史 `ParentID=0` 的生成菜单原地迁移到 1700；抽出 `ensureGenButtons()` 幂等补齐 6 个 F 按钮（查询/新增/修改/删除/导入/导出），避免已生成模块缺新权限；生成 `api.tmpl` 的 `init()` 显式 `ParentID:1700`。
- **导出**：DAO 增加 `List{ClassName}All(q)` 不分页；API `GET /export` 按当前搜索条件生成 xlsx；主子表导出两张 sheet，明细首列“所属主表行号”= 主表 Excel 行号；树形导出平铺数据。
- **模板**：`GET /importTemplate` 生成仅表头的 xlsx；主子表两张 sheet；树形/单表一张。
- **导入**：`POST /importData`，`maxImportSize=5MB`、`maxImportRows=5000`；按表头名称匹配字段；`updateSupport=true` 时主键已存在则更新，否则报“主键 N 已存在（如需覆盖请勾选“更新支持”）”；新行保留 Excel 显式主键；主子表通过明细“所属主表行号”建立映射，新增主表也能把明细挂到新生成主键，更新主表时未提供明细则保留原明细；成功后 `dao.SyncImportSequence(table, pkColumn)` 同步 PG 序列，避免显式主键导致后续自增撞键。
- **数据结构**：`Req` 字段加 `excel:"中文列名"`；主子表 `Children` 与子表 PK/FK 加 `excel:"-"`；`utils/excel_reflect.go` 的 `ExcelRow` 按字段顺序输出并跳过切片/map/数组与 `excel:"-"`，`UnmarshalExcelRow` 按表头名匹配 excel tag，支持 string/int/uint/float/bool/LocalTime，空单元格跳过，未知表头忽略，`matched==0` 报“表头与导入模板不一致”；`model.LocalTime` 增加 `ExcelString()`/`UnmarshalExcel()`。
- **下载文件名**：`Content-Disposition` 不能直接放 UTF-8 中文，否则前端 `parseFilename` 取到 mojibake；`api/common.go` 的 `excelAttachment` 输出 `filename="download.xlsx"; filename*=UTF-8''<url.PathEscape>`，模板与导出共用。
- **前端**：`frontend/src/components/ExcelImportDialog/index.vue` 提供文件选择、模板下载、更新支持开关、成功/失败提示与失败明细；三类 `vue_index*.tmpl` 增加导入/导出按钮（`v-hasPermi` 对应 `:import`/`:export`）、`ExcelImportDialog`、`exportFile` 与 `handleExport`；只接受 xlsx，前端预检 5MB。
- **验证**（PG 18.4 + 真实浏览器）：
  - 三张 demo 表生成菜单 `2041/2048/2055` 均 `parent_id=1700`，各有 6 个 F 按钮并绑定 role 1；侧边栏“系统工具”下出现三个菜单，面包屑“系统工具 / 演示图书”；
  - 三份导出 xlsx 表头/数据正确；订单两张 sheet，明细首列“所属主表行号”与主表行号一致；三份模板仅表头；
  - 单表导入 `updateSupport=true`：3 行成功（2 更新 + 1 新增），随后新增记录 id=4，证明 PG 序列同步；`updateSupport=false` 导入已有主键报“主键 N 已存在”；
  - 树形导入：更新根/子节点 + 新增 id3，层级正确；主子表导入：更新订单 1 及明细、新增订单 2 及 2 条明细，主/子 max(id) 与后续新增 id 均正确；
  - 浏览器：`/tool/demoBook` 页面导入/导出按钮、模板下载（中文文件名）、上传导入成功提示、`/tool/demoCategory`、`/tool/demoOrder` 页面均正常；订单导入弹窗提示双 sheet 规则。
- **测试**：`TestGenTemplates` 5 例（crud/非自增主键跳过序列/tree/sub/时间字段）、`TestExcel` 4 例（导出顺序与 `excel:"-"`、按表头回填、未知表头报错、空单元格保留）；`go build`、`npm run type-check`、`npm run build` 通过。
- **限制**：导入只支持 xlsx（xls 走 excelize 会解析失败）；主表字段更新为整行覆盖语义，空白单元格会按零值/空值写入，更新时建议保留整行；主子表更新时，未填写明细的主表行保留原有明细（页面编辑仍是全量替换）；树形导入的父级列按父级主键引用，新节点若要挂到同批次新增的父节点，需先导入上级或显式填写父级 ID；`SyncImportSequence` 仅 PG 生效且仅自增主键调用。