package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"go-admin/dao"
	"go-admin/model"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter 捕获响应体，用于判断操作结果；仅保留前 limit 字节，避免大响应（导出文件）占满内存
type bodyLogWriter struct {
	gin.ResponseWriter
	body  *bytes.Buffer
	limit int
}

// Write 先缓存响应体再透传到底层 ResponseWriter（实现 io.Writer）
func (w bodyLogWriter) Write(b []byte) (int, error) {
	if w.limit > 0 && w.body.Len() < w.limit {
		n := len(b)
		if w.body.Len()+n > w.limit {
			n = w.limit - w.body.Len()
		}
		w.body.Write(b[:n])
	}
	return w.ResponseWriter.Write(b)
}

// WriteString 先缓存响应字符串再透传（实现 io.StringWriter）
func (w bodyLogWriter) WriteString(s string) (int, error) {
	if w.limit > 0 && w.body.Len() < w.limit {
		n := len(s)
		if w.body.Len()+n > w.limit {
			n = w.limit - w.body.Len()
		}
		w.body.WriteString(s[:n])
	}
	return w.ResponseWriter.WriteString(s)
}

// 跳过无需记录的操作（含高频轮询接口）
var operLogSkip = map[string]bool{
	"/api/captchaImage":           true,
	"/api/login":                  true,
	"/api/logout":                 true,
	"/api/getInfo":                true,
	"/api/getRouters":             true,
	"/api/monitor/server/info":    true, // 服务监控页 30s 轮询，避免日志刷屏
	"/api/monitor/cache/info":     true,
	"/api/monitor/db/info":        true,
	"/api/system/menu/treeselect": true,
	"/api/system/dept/treeselect": true,
	"/api/common/upload":          true, // multipart 大文件，读取请求体代价高
}

// LogMaxRespLen 响应体缓存上限（仅需判定 code/msg，无需完整响应）
const LogMaxRespLen = 4096

// 按前缀跳过的接口（带动态参数）
var operLogSkipPrefix = []string{
	"/api/system/dict/data/type/", // 下拉字典高频
}

// 业务类型（对标 RuoYi）：0其它 1新增 2修改 3删除 4查询
const (
	BizTypeOther   = 0
	BizTypeInsert  = 1
	BizTypeUpdate  = 2
	BizTypeDelete  = 3
	BizTypeQuery   = 4
	StatusSuccess  = "0"
	StatusFail     = "1"
	LogMaxParamLen = 2000
)

// 与 model.OperLog 的 gorm size 约束保持一致。
// PostgreSQL/MySQL 对 varchar 长度与字节合法性是严格校验（SQLite 不校验），
// 超长或非法 UTF-8 会直接抛错（22001 / 22021），而日志是异步静默写入，表现为"日志丢失"。
const (
	maxOperTitle  = 128
	maxOperMethod = 255
	maxOperIP     = 64
	maxOperURL    = 255
	maxOperName   = 32
	maxOperMsg    = 2000
	maxReqMethod  = 16
	// maxOperBody 操作日志读取请求体的上限；超出部分不进入日志，但会完整还给业务 handler。
	maxOperBody = 1 << 20
)

// truncMarker 截断标记，便于排查时区分"原文就这么短"与"被截断"
const truncMarker = "…[截断]"

// truncateUTF8 按 UTF-8 字符边界截断到 maxBytes 以内，并清洗非法字节。
// 直接按字节切片会切断多字节字符（中文 3 字节），产生非法 UTF-8，
// PostgreSQL 会拒绝写入（SQLSTATE 22021）导致整条操作日志丢失。
func truncateUTF8(s string, maxBytes int) string {
	if s == "" || maxBytes <= 0 {
		return ""
	}
	// 先清洗来源中的非法字节（如上传请求体混入的二进制）
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	if len(s) <= maxBytes {
		return s
	}
	// 预留标记长度，保证结果（含标记）不超过字段上限
	limit := maxBytes - len(truncMarker)
	if limit <= 0 {
		limit = maxBytes
		cut := limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		return s[:cut]
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + truncMarker
}

// maskSensitive 对请求体中的密码类字段脱敏，避免明文写入操作日志。
// 递归处理嵌套对象和数组，避免 {"user":{"password":"xxx"}} / [{"password":"xxx"}] 漏网。
func maskSensitive(raw string) string {
	if raw == "" {
		return raw
	}
	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}
	if !maskValue(v) {
		return raw
	}
	out, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return string(out)
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "password") || strings.Contains(lower, "passwd") || lower == "pwd" || lower == "newpwd"
}

// maskValue 递归脱敏；返回是否发生替换。
func maskValue(v interface{}) bool {
	switch x := v.(type) {
	case map[string]interface{}:
		changed := false
		for k, child := range x {
			if isSensitiveKey(k) {
				if s, ok := child.(string); ok {
					if s != "" {
						x[k] = "******"
						changed = true
					}
				} else if child != nil {
					x[k] = "******"
					changed = true
				}
				continue
			}
			if maskValue(child) {
				changed = true
			}
		}
		return changed
	case []interface{}:
		changed := false
		for _, child := range x {
			if maskValue(child) {
				changed = true
			}
		}
		return changed
	default:
		return false
	}
}

// resolveBusinessType 按 HTTP 方法 + 路径语义映射业务类型。
// GET 统一查询；上传、流程提交不是新增；导入仍是新增；DELETE 批量清理仍是删除。
func resolveBusinessType(method, path string) int {
	lower := strings.ToLower(path)
	switch method {
	case http.MethodGet:
		return BizTypeQuery
	case http.MethodPost:
		if strings.Contains(lower, "/upload") || strings.Contains(lower, "/flow/apply") {
			return BizTypeOther
		}
		if strings.Contains(lower, "/export") {
			return BizTypeQuery
		}
		return BizTypeInsert
	case http.MethodPut:
		return BizTypeUpdate
	case http.MethodDelete:
		return BizTypeDelete
	default:
		return BizTypeOther
	}
}

// 把数据库连接池打满。队列满时直接丢弃并计数，宁可少记日志也不阻塞业务请求。
const (
	operLogQueueSize     = 2048
	operLogBatchSize     = 100
	operLogFlushInterval = 200 * time.Millisecond
	operLogWorkerCount   = 2
	operLogSlowGETMS     = 1000 // 慢 GET 始终记录
)

var (
	operLogQueue       = make(chan *model.OperLog, operLogQueueSize)
	operLogWorkerOnce  sync.Once
	operLogDropped     uint64
	operLogInsertError uint64
	operLogStop        = make(chan struct{})
	operLogStopOnce    sync.Once
	operLogWG          sync.WaitGroup
	operLogClosed      uint32
	// A2 读日志采样：默认 10（约 1/10），GO_ADMIN_GET_LOG_SAMPLE=1 表示全部记录。
	operLogGetSampleEvery = loadGetLogSampleEvery()
	operLogGetSampled     uint64
)

func loadGetLogSampleEvery() uint64 {
	v := strings.TrimSpace(os.Getenv("GO_ADMIN_GET_LOG_SAMPLE"))
	if v == "" {
		return 10
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 10
	}
	if n == 0 {
		return 1
	}
	return n
}

// shouldRecordGET 对 GET 做降噪：错误、慢请求、导出/下载始终记录，其余按 1/N 采样。
// 非 GET 不受影响。
func shouldRecordGET(method, path, status string, cost int64) bool {
	if method != http.MethodGet {
		return true
	}
	if status == StatusFail || cost >= operLogSlowGETMS {
		return true
	}
	lower := strings.ToLower(path)
	if strings.Contains(lower, "/export") || strings.Contains(lower, "/download") {
		return true
	}
	if operLogGetSampleEvery <= 1 {
		return true
	}
	return atomic.AddUint64(&operLogGetSampled, 1)%operLogGetSampleEvery == 0
}

// OperLogQueueStats 操作日志队列的只读观测数据（供服务监控页展示）。
type OperLogQueueStats struct {
	QueueLen     int    `json:"queueLen"`
	QueueCap     int    `json:"queueCap"`
	Dropped      uint64 `json:"dropped"`
	InsertErrors uint64 `json:"insertErrors"`
}

// OperLogStats 返回当前队列积压、累计丢弃与批量写入失败次数。
func OperLogStats() OperLogQueueStats {
	return OperLogQueueStats{
		QueueLen:     len(operLogQueue),
		QueueCap:     cap(operLogQueue),
		Dropped:      atomic.LoadUint64(&operLogDropped),
		InsertErrors: atomic.LoadUint64(&operLogInsertError),
	}
}

// tryEnqueueOperLog 非阻塞入队；队列满时递增 dropped 并返回 false。
// 独立出来是为了用局部队列稳定测试“队列满丢弃”路径，不依赖全局 worker。
func tryEnqueueOperLog(queue chan *model.OperLog, log *model.OperLog, dropped *uint64) bool {
	select {
	case queue <- log:
		return true
	default:
		atomic.AddUint64(dropped, 1)
		return false
	}
}

// enqueueOperLog 非阻塞入队；队列满时丢弃，避免日志反压拖慢接口。
func enqueueOperLog(log *model.OperLog) {
	if atomic.LoadUint32(&operLogClosed) == 1 {
		return
	}
	operLogWorkerOnce.Do(func() {
		for i := 0; i < operLogWorkerCount; i++ {
			operLogWG.Add(1)
			go func() {
				defer operLogWG.Done()
				operLogWorker()
			}()
		}
	})
	if !tryEnqueueOperLog(operLogQueue, log, &operLogDropped) {
		n := atomic.LoadUint64(&operLogDropped)
		if n%100 == 0 {
			stdlog.Printf("操作日志队列已满，累计丢弃 %d 条", n)
		}
	}
}

// StopOperLog 停止接收新日志，通知 worker 排空队列并 flush 剩余批次。
// 必须在 dao.DB.DB().Close() 之前调用；ctx 控制等待上限。
func StopOperLog(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	operLogStopOnce.Do(func() {
		atomic.StoreUint32(&operLogClosed, 1)
		close(operLogStop)
	})
	done := make(chan struct{})
	go func() {
		operLogWG.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// operLogWorker 聚合 100 条或 200ms 后批量写库。
func operLogWorker() {
	ticker := time.NewTicker(operLogFlushInterval)
	defer ticker.Stop()

	batch := make([]*model.OperLog, 0, operLogBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		logs := batch
		batch = make([]*model.OperLog, 0, operLogBatchSize)
		defer func() {
			if r := recover(); r != nil {
				stdlog.Printf("操作日志批量写入 panic 已捕获: %v", r)
			}
		}()
		if err := dao.InsertOperLogs(logs); err != nil {
			n := atomic.AddUint64(&operLogInsertError, 1)
			if n%20 == 1 {
				stdlog.Printf("操作日志批量写入失败（累计 %d 次）: %v", n, err)
			}
		}
	}

	drain := func() {
		for {
			select {
			case log := <-operLogQueue:
				batch = append(batch, log)
				if len(batch) >= operLogBatchSize {
					flush()
				}
			default:
				flush()
				return
			}
		}
	}

	for {
		select {
		case log := <-operLogQueue:
			batch = append(batch, log)
			if len(batch) >= operLogBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-operLogStop:
			drain()
			return
		}
	}
}

// OperLog 操作日志中间件（记录增删改查、操作人、IP、耗时与结果；异步写库不阻塞业务响应）
func OperLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		skip := operLogSkip[path] || strings.HasPrefix(path, "/api/system/user/profile")
		if !skip {
			for _, p := range operLogSkipPrefix {
				if strings.HasPrefix(path, p) {
					skip = true
					break
				}
			}
		}
		if skip {
			c.Next()
			return
		}
		start := time.Now()

		// 读取请求体（读后还原，不影响业务处理）；multipart 上传不读取，避免把整个文件流读入内存。
		// 只读前 maxOperBody+1 字节：超限时日志只取前 maxOperBody，剩余原始流通过 MultiReader 还给业务。
		var bodyBytes []byte
		if c.Request.Body != nil && !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
			originalBody := c.Request.Body
			raw, _ := io.ReadAll(io.LimitReader(originalBody, maxOperBody+1))
			c.Request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(raw), originalBody))
			if len(raw) > maxOperBody {
				raw = raw[:maxOperBody]
			}
			bodyBytes = raw
		}
		param := maskSensitive(string(bodyBytes))
		if param == "" {
			param = c.Request.URL.RawQuery
		}
		param = truncateUTF8(param, LogMaxParamLen)

		blw := &bodyLogWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString(""), limit: LogMaxRespLen}
		c.Writer = blw

		c.Next()

		cost := time.Since(start).Milliseconds()
		status := StatusSuccess
		errMsg := ""
		// 先看 HTTP 状态；再看 JSON 业务 code。导出 xlsx 等二进制响应不是 JSON，
		// 不能再靠 json.Unmarshal 的 code 判断，否则会把成功导出记成失败。
		if c.Writer.Status() >= http.StatusBadRequest {
			status = StatusFail
			errMsg = http.StatusText(c.Writer.Status())
		}
		if strings.Contains(c.Writer.Header().Get("Content-Type"), "application/json") {
			var resp struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			if err := json.Unmarshal(blw.body.Bytes(), &resp); err == nil {
				if resp.Code != 200 {
					status = StatusFail
				}
				errMsg = resp.Msg
			}
		}

		// A2：GET 列表/详情按采样降噪；错误、慢请求、导出下载仍全记。
		if !shouldRecordGET(c.Request.Method, path, status, cost) {
			return
		}

		// 业务类型映射：GET 查询；上传/流程提交归其它；导入仍是新增。
		bizType := resolveBusinessType(c.Request.Method, path)

		log := &model.OperLog{
			Title:         truncateUTF8(c.FullPath(), maxOperTitle),
			BusinessType:  bizType,
			Method:        truncateUTF8(c.HandlerName(), maxOperMethod),
			RequestMethod: truncateUTF8(c.Request.Method, maxReqMethod),
			OperName:      truncateUTF8(c.GetString("username"), maxOperName),
			OperIP:        truncateUTF8(c.ClientIP(), maxOperIP),
			OperURL:       truncateUTF8(path, maxOperURL),
			OperParam:     param,
			Status:        status,
			ErrorMsg:      truncateUTF8(errMsg, maxOperMsg),
			CostTime:      cost,
			OperTime:      time.Now(),
			RequestID:     truncateUTF8(c.GetString(RequestIDKey), 64),
		}
		// 入有界队列，由固定 worker 批量写库；不阻塞业务响应。
		enqueueOperLog(log)
	}
}
