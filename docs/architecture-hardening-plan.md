# 架构加固双路径：单机加固与多实例改造

> 基线日期：2026-11-15
> 适用仓库：`D:\Golang\go-admin`
> 结论前提：当前 go-admin 是“单节点、单库、单进程内嵌前端”的中小型后台，不是玩具；但它的协调原语大量在进程内，不能直接横向扩容。
> 本文只落“执行清单 + 验收标准”，不一次性重写代码。两条路径可独立执行，也可按 A → B 顺序推进。

---

## 0. 当前基线（先防止用错数字）

以下数字来自当前仓库代码/数据库，不是 kmerit 的目标规模：

| 项 | 当前实际 |
| --- | --- |
| 路由注册 | 145 处方法注册（含 4 处 genreg 动态注册），运行时 141 条路由 |
| `sys_api` 规则 | 124 条（GET 58 / PUT 29 / DELETE 20 / POST 17） |
| `DB.Transaction` | 26 处，全在 `dao`：flow 12、role 5、user 5、gen 3、dept 1 |
| `interface{}` | `api/*.go` 7 处，全在 `export_api.go` 的 Excel 行数据 |
| 响应无类型 | `gin.H{` 14 处，`utils.OkData(c, gin.H...)` 12 处 |
| 鉴权 DB 读 | 已登录请求 5–7 次读；另 operlog 异步 1 次写 |
| AutoMigrate | `config.yaml` 默认 `true`，22 个模型 |
| 进程内状态 | 验证码、登录限流、cron、字典缓存、`sys_api` 规则 |
| 外部依赖 | 无 Redis、无消息队列、无对象存储、无 trace |
| 上传 | 本地 `backend/uploads/yyyyMMdd/` |
| `sys_api` 规则失效 | 仅本实例 `ReloadApiRules()`，无跨实例广播、无 TTL |
| 字典缓存 | `sync.Map` + 60s TTL，跨实例最多脏 60s |

**结论：** 不要把 471 端点、PPT 缩略图、ApprovalBusinessHandler 当成 go-admin 现状；它们是“复刻 kmerit”的目标，适合用来评估未来容量，不适合用来给当前代码定罪。

---

## A. 单机加固（不动目录结构，先把单节点的性能和运维风险压下去）

**目标：** 单实例继续稳定运行；鉴权查询从 5–7 次降到 2–3 次；日志写入有界；生产迁移/种子可控；路由契约可回归。
**不建议：** 为了“像大厂”先搬 `internal/`、先加空 service、先抽不存在的 Redis/OSS 接口。

### A1. 鉴权查询瘦身（最高优先级）

**问题：**
- `middleware/auth.go:29` 调 `dao.GetUserByID`，后者是 `Preload("Roles").Preload("Posts").Preload("Dept").First(...)`，一次鉴权产生 4 条查询；
- 再加 `dao.OnlineExists` 1 条 = 5 条；
- `middleware/apiperm.go` 命中规则后 `IsAdmin` 1 条；非 admin 再 `GetPermsByUser` 1 条 → 6–7 条。

**改动：**
1. 在 `dao/user_dao.go` 新增只取鉴权字段的查询：
   ```go
   func GetUserAuthInfo(id uint) (*model.User, error) {
       var u model.User
       err := DB.Select("id", "username", "status").First(&u, id).Error
       return &u, err
   }
   ```
   `Auth` 改用它，不再加载 Roles/Posts/Dept；`c.Set("username", user.Username)` 用 DB 返回值，避免 JWT 里的旧用户名。
2. `middleware/apiperm.go` 删除 `dao.IsAdmin(userID)` 查询，改为 `c.GetString("username") == "admin"`。`Auth` 已经查过 username，且 JWT 签名可信。
3. `GetPermsByUser` 增加进程内 TTL 缓存（30–60 秒），key 为 `userID`；角色/菜单/用户授权变更时清空对应缓存或递增全局 `permsVersion`。
4. `OnlineExists` 改为 `SELECT 1 ... LIMIT 1` 语义（GORM `Take` + `Select("1")` 或原生 SQL），不再 `COUNT`。

**验收：**
- 管理员已登录请求：2 条 SELECT（用户 + 在线会话）；
- 非管理员命中权限规则：3 条 SELECT（用户 + 在线 + 权限）；
- 用 `pg_stat_statements` 或 GORM logger 对比改造前后，单请求查询数下降 ≥50%；
- 现有登录、停用、强退、角色变更测试全绿。

**测试文件：**
- `dao/user_dao_test.go`（如新增查询函数）
- `middleware/auth_test.go` / `middleware/apiperm_test.go`（可用 `httptest` + SQLite）
- 回归：`go test ./dao ./middleware ./router -count=1`

---

### A2. 操作日志改成有界队列 + 批量写

**问题：**
- `middleware/operlog.go:236` 每个认证请求一个裸 `go func`，无并发上限；
- 每条一个 `INSERT`，连接池默认 `maxOpen=25`；
- 列表/详情 GET 也记录，高频轮询会放大；
- 白名单实际是 11 个精确路径 + 字典前缀 + profile 前缀，不是 10 个。

**改动：**
1. 包内初始化 `operLogQueue := make(chan *model.OperLog, 2048)`，启动 2–4 个固定 worker。
2. worker 聚合 100 条或 200ms 批量 `CreateInBatches`；DAO 新增 `InsertOperLogs([]*model.OperLog) error`。
3. 入队用非阻塞 `select`；队列满时丢弃并累加 `dropped_total`，每 N 次打一条 warning，不阻塞请求。
4. 默认只记录写操作（POST/PUT/DELETE）和显式审计接口；GET 列表/详情是否记录做成配置 `operLog.logReads`，默认 false。
5. 响应状态优先用 `c.Writer.Status()`；只有 `Content-Type: application/json` 时才解析 `code`，避免 xlsx 导出等二进制响应被误判为失败。
6. 退出时关闭队列并 `WaitGroup.Wait()`，最多等 3 秒。

**验收：**
- 压测 100 QPS 持续 1 分钟，goroutine 数不随请求线性增长；
- 日志队列长度有上限，丢弃有指标；
- 操作日志表写入量下降（按实际启用的路由统计）；
- `go test ./middleware -run TestOperLog` 覆盖：入队、满队列丢弃、批量写、二进制响应状态。

---

### A3. AutoMigrate / seed / dropForeignKeys 解耦

**问题：**
- `config.yaml` 默认 `autoMigrate: true`，22 个模型每次启动自动迁移；
- `dao/db.go:190` 的 `dropForeignKeys()` 在 `autoMigrate` 判断之外，**无条件执行**；
- `seed()` 和 `ensureGenMenus()` 在 `autoMigrate` 分支内，关掉迁移会一并关掉生成模块菜单；
- 两个实例同时启动会并发 DDL + 固定 ID 种子，可能一个实例 `log.Fatalf` 退出。

**改动：**
1. 拆成三个独立配置：
   ```yaml
   database:
     autoMigrate: false      # 仅开发/首次建库
     seedData: false         # 菜单/示例数据/生成模块菜单
     dropForeignKeys: false  # 生产默认不删外键
   ```
2. `afterOpen()` 按三个开关分别执行：`AutoMigrate + ensureIndexes`、`dropForeignKeys`、`seed + ensureGenMenus`。
3. 已新增 `backend/cmd/migrate/main.go`：`migrate` 只跑 `AutoMigrate + ensureIndexes`，`-seed` 才写种子，`-drop-fk` 才删外键；生产由部署 job 单次执行，不在应用启动时跑。
4. 当前没有版本化 SQL migration 文件；`cmd/migrate` 直接复用 GORM AutoMigrate，首个 baseline 由当前结构生成。若后续需要可回滚的版本化迁移，再接 `golang-migrate`。
5. `dropForeignKeys` 只允许在迁移命令里显式执行；应用启动不再删约束。

**验收：**
- 新库：`migrate up` 后表结构完整；`seed` 两次幂等；生成模块菜单正常创建；
- 生产配置 `autoMigrate=false, seedData=false, dropForeignKeys=false` 下应用正常启动，接口可用；
- 两个实例同时启动，不执行 DDL/种子，均健康；
- `go test ./dao -run TestMigrateFlags` 覆盖三个开关组合。

**2026-11-15 落地情况：** A1 第 1–2 项、A2 第 1–3/5 项、A3 第 1–2 项已实现；随后复检又补齐 A1 权限缓存与 `OnlineExists LIMIT 1`、A2 operlog 优雅 drain/请求体上限/递归脱敏/队列观测、日志保留 90 天、后台 watcher 收敛。`go build ./...`、`go vet ./...`、`go test ./... -count=1` 全绿。测试覆盖：`TestGetUserAuthInfoSkipsPreloads`、`TestInsertOperLogs`、`TestOperLogQueueWrites`、`TestOperLogBinaryResponseNotFail`、`TestAfterOpenSeedWithoutAutoMigrate`、`TestAfterOpenWithoutAutoMigrateDoesNotCreateTable`、`TestMaskSensitiveNested`、`TestOnlineExistsExcludesExpired`、`TestOnlineExistsActive`、`TestIsSuperAdminName`、`TestRouteContract`、`TestRequestID`、`TestResolveBusinessType`、`TestOperLogQueueDrop`、`TestResponseShape`、`TestBodyLimit`、`TestShouldRecordGET`、`TestStartOnlineCleanup`。A4 路由/响应 shape golden、A5 request id、独立 `cmd/migrate`、GET 采样、全局请求体上限、在线会话周期清理均已补；尚未做：B 路径（Redis/选主/多实例缓存失效）。

---

### A4. 路由/响应契约快照（最低成本防漂移）

**问题：**
- 当前请求 DTO 多数存在，但响应大量 `gin.H`；
- 运行时 141 条路由没有统一契约快照，后续加端点容易改坏前端。

**改动：**
1. 新增 `router/contract_test.go`：调用 `r.Routes()`，输出 `method + path + handler name`，与 `router/testdata/routes.golden.json` 比较。
2. 对 10–20 个核心接口补响应 shape 快照（`code/msg/data` 的字段集合），不要求全量。
3. 前端 `src/types/api.ts` 已有的 `PageResult<T>` 等类型继续用；新增接口必须带返回类型。
4. 可选：从 `sys_api` + 路由表生成 OpenAPI 3.0，先只读，不接代码生成。

**验收：**
- 新增/删除/改路径必须显式更新 golden 文件，CI 失败可定位；
- `go test ./router -run TestRouteContract` 通过；
- 前端 `npm run type-check` 通过。

---

### A5. 请求 ID / 基础可观测性

**改动：**
- 新增 `middleware/request_id.go`：读取 `X-Request-ID`，没有则生成 UUID；写回响应头；放入 `gin.Context`。
- 访问日志统一输出：`request_id`、method、path、status、latency、user、client_ip。
- 暂不引入 OpenTelemetry；先把“能按一个请求串起来”做到。

**验收：**
- 任意响应带 `X-Request-ID`；
- 同一次请求的访问日志、operlog 可通过 ID 关联；
- `go test ./middleware -run TestRequestID` 通过。

---

### A6. 长任务阈值化（可选，先不做全量队列）

- 导出：小于 5000 行走同步；超过走后台 job（见 B7）。
- 导入：小于 1000 行走同步；超过走后台 job。
- 外部数据源同步、缩略图等天然长任务，直接进 B7。

**验收：**
- 同步接口 p95 不因大导出超过 3 秒；
- 超过阈值的请求返回 `jobId`，前端可轮询进度。

---

## B. 多实例改造（要上第二个实例前必须做）

**目标：** 两个实例共享登录态、限流、调度、权限规则、文件和长任务；任何一个实例故障不影响另一个。
**前置：** Redis（验证码/限流/广播/可选会话）、共享存储（MinIO/S3 或共享卷）、单次迁移 job。
**顺序：** B1 → B2 → B3 → B4 → B5 → B6 → B7；不要先搬目录。

### B1. 验证码 + 登录限流迁移到 Redis

**改动：**
1. 配置 `redis.addr/password/db`，新增 `cache.driver=memory|redis`；生产用 redis，开发可 memory。
2. `cache.CaptchaStore` 抽成接口，新增 `RedisCaptchaStore`（实现 `base64Captcha.Store` 的 `Set/Get/Verify`），key 前缀 `captcha:`，TTL 5 分钟。
3. `utils/limiter.go` 新增 Redis 实现：账号/IP 两个 key，Lua 脚本原子执行 `INCR + EXPIRE + 阈值判断`；锁定 15 分钟。
4. Redis 故障策略要明确：生产 fail-fast（登录暂不可用）优于退回本地内存导致防护失效；开发可退回 memory。

**验收：**
- 两个实例：A 生成验证码，B 校验通过；
- 两个实例：同一账号连续失败 5 次，A/B 都锁定；
- Redis 重启后验证码/限流行为符合配置策略；
- 测试用 `miniredis` 或真实 Redis 容器，覆盖 Lua 原子性。

---

### B2. 定时任务 / 后台 watcher 选主

**问题不止 cron：**
- `scheduler.Start()` 每实例注册全部任务；
- `main.go:80` 的 `StartOverdueWatcher` 每分钟扫描催办；
- `main.go:82` 的 `StartLogRetention` 也会在每个实例跑。

**改动：**
1. 新增 `dao/lock.go`：PG 用 `pg_try_advisory_lock(key)` 持有会话级锁；MySQL 用 `GET_LOCK`；SQLite 单实例直接返回 true。
2. `scheduler.Start()` 只有 leader 注册 cron；follower 每 5 秒重试抢锁；leader 丢失后释放并重新选主。
3. `StartOverdueWatcher`、`StartLogRetention` 同样放进 leader gate；非 leader 不启动。
4. `scheduler.Stop()` / 进程退出释放锁。

**验收：**
- 两个实例同时启动，同一 job 只执行一次；
- kill leader 后，5–10 秒内 follower 接管；
- 催办 watcher 不重复写 `sys_flow_record`；
- 日志保留清理不并发删同一批数据。

---

### B3. 启动迁移/种子单点化

**改动：**
- 生产建议把 `autoMigrate`/`seedData`/`dropForeignKeys` 设为 false，部署前先跑 `go run ./cmd/migrate -config config.yaml`；当前默认仍为 `autoMigrate: true`、`seedData: true`，以便开发环境直接启动。
- 迁移命令自己抢 advisory lock，保证多副本部署只有一个 job 改结构。
- 本地开发保留 `autoMigrate=true` 的便捷模式。

**验收：**
- 空库 + 两个实例同时启动：两个都健康，表只建一次，菜单只种一次；
- 迁移失败时应用不启动，回滚/重试有明确日志；
- 生产 `autoMigrate=false` 下代码生成模块菜单仍能通过 `seed` 命令创建。

---

### B4. `sys_api` 规则 / 字典缓存跨实例失效

**问题：**
- `dao/api_dao.go` 的增删改只调用本实例 `ReloadApiRules()`；
- 没有 TTL，A 停用规则后 B 可以一直放行；
- 字典缓存有 60s TTL，最多脏 60 秒，可接受但不是强一致。

**改动：**
1. API 规则：
   - 写操作成功后发布 Redis channel `go-admin:api-rules:reload`；
   - 所有实例订阅并执行 `ReloadApiRules()`；
   - 增加 30 秒周期兜底 reload（Redis 不可用时仍能收敛）。
2. 字典缓存：
   - 本实例 `ClearDictCache` 后发布 `go-admin:dict:invalidate`，携带 dictType；
   - 其他实例清理对应 key；60s TTL 继续兜底。
3. 多实例权限变更 SLA 写进文档：目标 1 秒内，最差 30 秒。

**验收：**
- 两实例：A 停用某 `sys_api` 规则，B 在 1 秒内拒绝对应请求；
- Redis 断连：B 在 30 秒内通过周期 reload 收敛；
- 字典修改后两实例返回一致（或 ≤60s）。

---

### B5. 上传文件共享存储

**问题：**
- `api/upload_api.go` 写本地 `uploads/yyyyMMdd/`；
- 头像只允许 `/uploads/` 前缀；
- 多实例下 A 上传、B 读不到。

**改动：**
1. 新增 `platform/storage` 接口：`Save(ctx, name, reader) (url, error)`、`Open`、`Delete`、`URL`。
2. 本地实现保留给开发；生产用 S3/MinIO 实现；配置 `storage.driver=local|s3`。
3. `UploadFile`、头像更新、导出文件落盘全部走接口；头像 URL 允许 `/uploads/` 或配置的 CDN 前缀。
4. 存量文件迁移：一次性脚本把 `uploads/` 同步到对象存储；短期可双读（本地 fallback + 对象存储）。

**验收：**
- A 上传图片，B 通过公网/内网 URL 正常读取；
- 删除/替换头像两实例一致；
- 10MB 上限、扩展名/类型校验不回退。

---

### B6. 鉴权缓存多实例化（可选，但 141 条路由规模建议做）

**改动：**
- 用户权限：Redis key `perms:user:<id>:v<version>`，TTL 60s；角色/菜单变更递增 version 并广播。
- 管理员判断：继续用 `username == "admin"`，不查库。
- 在线会话：当前 `sys_online` 已是 DB 共享，可保留；若要 1 次 Redis GET，则 token hash key + TTL，logout 删除，DB 仅作审计/兜底。
- 用户状态：保留每请求 1 次轻量查询，保证停用立即生效；不要为了省 1 次查询把停用延迟到 TTL。

**验收：**
- 两实例：角色变更后 1 秒内新权限生效；
- 停用用户后，旧 token 在两个实例都立即 401；
- 鉴权 p95 从 5–7 次读降到 1–3 次。

---

### B7. 后台任务队列（长任务）

**改动：**
1. 新增 `sys_job_queue`：`id/type/payload/status/attempts/lease_until/result_path/created_at`。
2. Worker 用 `FOR UPDATE SKIP LOCKED`（PG）或等价机制取任务，保证多实例不重复执行；失败重试 + 退避 + 最大次数。
3. 任务类型：大导出、大导入、外部数据源同步、库对比、缩略图、批量通知。
4. HTTP 接口只负责入队，返回 `jobId`；前端轮询状态/下载结果。
5. 小任务仍可同步，避免所有导出都排队。

**验收：**
- A 入队，B 的 worker 执行，任务只成功一次；
- worker 崩溃后 lease 到期可被重新领取；
- 导出 10 万行不占 HTTP worker，进度可查；
- `go test ./jobs` 覆盖：入队、领取、幂等、重试、超时。

---

## C. 推荐执行顺序

| 阶段 | 内容 | 预估 | 退出条件 |
| --- | --- | --- | --- |
| 0 | 基线统计 + 路由 golden + request id | 1–2 天 | 能说清每个请求查了几次库、路由变更能被 CI 拦住 |
| 1 | A1 鉴权瘦身 + A2 日志队列 + A3 迁移/种子解耦 | 3–5 天 | 单机 p95 明显下降；生产配置可安全启动 |
| 2 | B1 Redis 验证码/限流 + B2 选主 + B3 迁移单点 | 1 周 | 两实例登录可用、cron 只跑一份、启动不打架 |
| 3 | B4 规则/字典失效 + B5 共享存储 | 1 周 | 权限变更 1 秒收敛；上传跨实例可读 |
| 4 | B6 鉴权缓存 + B7 任务队列 | 1–2 周 | 141 条路由规模下鉴权/长任务不成为瓶颈 |
| 5 | 垂直切片/ service 层（按模块渐进） | 持续 | 新模块走 `internal/`，旧模块桥接，不一次性搬 |

**工作量是单人粗估，仅用于排序，不作为承诺。**

---

## D. 明确不做的事

- 不一次性把 `api/dao/model` 全部搬进 `internal/`；先做 P0，再按模块绞杀。
- 不给每个 CRUD 加空 service；只给跨表/跨聚合用例加事务编排。
- 不为当前不存在的 Redis/OSS 造空接口；等第二个实现出现再抽。
- 不全表加 `del_flag`；软删按表决定，唯一索引改 partial index。
- 不动 `genreg`、`sys_api`、`data_scope`、UTF-8 截断/脱敏、excelize 这些已经验证过的设计。
- 不把 471 端点当成当前仓库事实；它是目标容量，不是现状。

---

## E. 总体验收

1. `go test ./...`、`go vet ./...`、`go build ./...` 全绿；
2. `npm run type-check`、`npm run build` 全绿；
3. 两实例联调：登录、验证码、限流、权限变更、上传、定时任务、长任务各跑一遍；
4. `pg_stat_statements` 对比鉴权查询数；
5. 路由 golden 无未解释 diff；
6. 生产配置 `autoMigrate=false, seedData=false, dropForeignKeys=false` 下启动成功；
7. 所有改动都有对应测试或可复现的验收命令，不能只靠“代码看起来对”。

---

## 附录 C：2026-11-15 复检修订

复检里大部分判断成立，但有 5 处数字/归类需要修正。已按真实缺口继续修复，下面把“成立并已修”“不准确”“仍未做”分开写，避免把目标容量当成现状。

### C.1 成立且已修复

| 复检结论 | 当前实现 |
| --- | --- |
| `middleware/permission.go` 漏改 | 改为 `isSuperAdmin(c)` 复用 Auth 写入的 DB username；`GetPermsByUser` 加 30s 进程内缓存，角色/菜单/用户授权变更时 `InvalidatePermCache()` 显式失效 |
| operlog 退出不 drain | 新增 `StopOperLog(ctx)`：关闭入队 → 通知 worker 排空队列 → flush 剩余批次 → `WaitGroup` 归零；`main.go` 在 `sqlDB.Close()` 前调用 |
| 丢弃计数无观测出口 | 新增 `OperLogStats()`，并在 `/api/monitor/server/info` 返回 `operLog.queueLen/queueCap/dropped/insertErrors` |
| 请求体无上限 | `io.LimitReader(..., maxOperBody+1)`，日志只取前 1MB，剩余原始流通过 `MultiReader` 完整还给业务 handler |
| 脱敏只做顶层 | `maskSensitive` 改为递归 `maskValue`，覆盖嵌套对象、数组、数组内对象、非字符串密码值 |
| 日志表无限增长 | `logs.retentionDays` 默认改为 90；启动清一次、之后每 24 小时一次 |
| 超管判定两套口径 | 新增 `dao.IsSuperAdminName`，`IsAdmin`、`apiperm`、`permission`、`auth_api` 统一走同一判定 |
| `stopOverdue` 未 close | `main.go` 在关库前 `close(stopOverdue)`，并新增 `dao.WaitBackground(ctx)` 等两个 watcher 退出 |
| 在线会话每请求 COUNT | `OnlineExists` 改 `SELECT 1 ... LIMIT 1`，并排除过期/空 `expire_at` 会话 |
| 测试断言弱 | 新增/加强：`TestAfterOpenWithoutAutoMigrateDoesNotCreateTable`、`TestMaskSensitiveNested`、`TestOnlineExistsExcludesExpired`、`TestOnlineExistsActive`、`TestIsSuperAdminName` |

### C.2 复检中不准确的地方

1. **`CheckPerm` 263 次不成立**：`router/router.go` 实测 **132 处** `middleware.CheckPerm(`，全仓库 **133 处**（含 `permission.go` 定义）。`CheckApiPerm` 只有 1 处注册（另 1 处定义）。263 可能是把前端 `v-permission` 或其他统计混进来了。
2. **471 端点不成立**：当前运行时路由实测 **141 条**（`router.go` 有 145 处 GET/POST/PUT/DELETE 方法注册，其中 4 处是 `genreg` 动态注册，不逐条计路由），`sys_api` **124 条**规则。132 处 `CheckPerm` 注册（含 4 处动态注册）对 141 条运行时路由是合理覆盖。
3. **“7 项都改对了”的归类不准**：`likeExprEsc`（ILIKE + ESCAPE）与 `normPage`（pageSize 1–500）是之前已有的 PG/分页加固，不是本轮 A1–A3 改动；功能确实存在，但不能算作这轮新修的 7 项。
4. **“每请求 4 次”只对非 admin 成立**：修复前 admin 是 5（Auth）+1（CheckPerm）=6 次，非 admin 是 5+1+1=7 次。修复后 admin 是 2 次（Auth 2 + CheckPerm 0），非 admin 首次 3 次（Auth 2 + 第一次 perms 查询 1），30 秒内后续 2 次。缓存带来的撤权延迟最长 30 秒，已用变更显式失效兜底。
5. **“在线会话落 DB”不应当作 P0**：DB 共享反而比进程内会话更利于多实例；真正的问题是每请求 `COUNT` 和过期行不主动清。现在已改成 `LIMIT 1` + 过期过滤，并补了每 10 分钟的周期清理（`dao.StartOnlineCleanup`）。

### C.3 单实例补强（2026-11-15 下午已补）

- **A2 GET 读日志采样**：`OperLog` 对 GET 默认按 1/10 采样；失败、慢请求（≥1000ms）、`/export`、`/download` 始终记录；`GO_ADMIN_GET_LOG_SAMPLE=1` 可关闭采样。新增 `TestShouldRecordGET`。
- **A4 响应 shape 快照**：路由 golden 已补（`router/testdata/routes.golden.json`，当前 141 条）；`router/response_shape_test.go` 对 15 个核心接口做字段形状快照（`router/testdata/response_shapes.golden.json`）。
- **A5 request id 落操作日志**：`model.OperLog` 增加 `request_id` 列并建索引；`OperLog` 从 `gin.Context` 取 `request_id` 写入；`X-Request-ID` 继续写入响应头/访问日志，CORS 允许并暴露。旧库通过下面的 `cmd/migrate` 自动补列。
- **独立迁移入口**：`backend/cmd/migrate/main.go`，`go run ./cmd/migrate -config config.yaml` 只跑表结构/索引，默认不跑种子；`-seed`、`-drop-fk` 显式开启。
- **在线会话周期清理**：`dao.StartOnlineCleanup` 启动先清一次、之后每 10 分钟清 `expire_at IS NULL` 或已过期会话，随 `stop` 退出并由 `WaitBackground` 等待；新增 `TestStartOnlineCleanup`。
- **全局请求体上限**：`middleware.BodyLimit` 普通 JSON 4MB、multipart 64MB；超限返回 413，chunked 由 `MaxBytesReader` 兜底；新增 `TestBodyLimit`。
- **业务类型语义映射**：`resolveBusinessType`：GET/export 归查询，upload/flow/apply 归其它，importData 仍是新增；新增 `TestResolveBusinessType`。
- **operlog 队列满丢弃单测**：`TestOperLogQueueDrop` 用局部队列稳定断言“队列满时 dropped 递增、成功入队不递增”。
- **慢 SQL / pprof**：GORM 默认 `logger.Default` 已以 200ms 为阈值输出 `SLOW SQL`；pprof 通过 `GO_ADMIN_PPROF_ADDR=127.0.0.1:6060` 显式开启，默认不监听。

### C.4 仍未做（单实例范围外）

- **B 路径**：Redis 验证码/限流、调度器与 watcher 选主、跨实例权限/字典失效、共享存储、分布式任务队列。当前权限缓存是进程内缓存，多实例下最长 30 秒脏读；单实例部署不需要。

