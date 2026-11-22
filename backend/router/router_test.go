package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-admin/config"
	"go-admin/dao"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// 测试环境说明：
//   - 使用独立临时 SQLite 文件（每轮测试全新建表 + 种子，互不影响）
//   - 关闭验证码（Captcha.Enabled=false）便于直接测登录
//   - 限流测试使用专属账号与独立 IP，避免污染其他用例

var (
	testEngine *gin.Engine
	testDBPath string
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	testDBPath = filepath.Join(os.TempDir(), fmt.Sprintf("go-admin-test-%d.db", time.Now().UnixNano()))
	config.Cfg = config.Config{
		Server:  config.ServerConfig{Port: 8080},
		JWT:     config.JWTConfig{Secret: "test-secret", Expire: 30},
		Captcha: config.CaptchaConfig{Enabled: false},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: testDBPath,
		},
	}
	if err := dao.InitDB(config.Cfg.Database); err != nil {
		log.Fatalf("初始化测试数据库失败: %v", err)
	}
	testEngine = InitRouter()
	code := m.Run()
	_ = os.Remove(testDBPath)
	_ = os.Remove(testDBPath + "-wal")
	_ = os.Remove(testDBPath + "-shm")
	os.Exit(code)
}

// doJSON 发起 JSON 请求；token 为空则不带认证头
func doJSON(t *testing.T, method, path, token string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	// 独立来源 IP：默认用 198.51.100.x，避免与限流用例（10.99.x）交叉
	req.RemoteAddr = "198.51.100.7:12345"
	w := httptest.NewRecorder()
	testEngine.ServeHTTP(w, req)
	var m map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	return w.Code, m
}

// login 登录并返回 token（失败时 t.Fatal）
func login(t *testing.T) string {
	t.Helper()
	_, m := doJSON(t, http.MethodPost, "/api/login", "", map[string]interface{}{
		"username": "admin", "password": "admin123",
	})
	if m["code"] != float64(200) {
		t.Fatalf("登录失败: %v", m)
	}
	data := m["data"].(map[string]interface{})
	tok, _ := data["token"].(string)
	if tok == "" {
		t.Fatalf("token 为空: %v", m)
	}
	return tok
}

func TestCaptchaImage(t *testing.T) {
	_, m := doJSON(t, http.MethodGet, "/api/captchaImage", "", nil)
	if m["code"] != float64(200) {
		t.Fatalf("验证码接口异常: %v", m)
	}
	data := m["data"].(map[string]interface{})
	if data["captchaEnabled"] != false {
		t.Fatalf("测试配置应关闭验证码: %v", data)
	}
	if data["uuid"] == "" {
		t.Fatalf("uuid 为空: %v", data)
	}
}

func TestLoginSuccess(t *testing.T) {
	_, m := doJSON(t, http.MethodPost, "/api/login", "", map[string]interface{}{
		"username": "admin", "password": "admin123",
	})
	if m["code"] != float64(200) {
		t.Fatalf("正确凭据登录失败: %v", m)
	}
	data := m["data"].(map[string]interface{})
	if data["token"] == "" {
		t.Fatalf("token 为空: %v", m)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	_, m := doJSON(t, http.MethodPost, "/api/login", "", map[string]interface{}{
		"username": "admin", "password": "wrong-pass",
	})
	if m["code"] != float64(500) {
		t.Fatalf("错误密码应返回失败: %v", m)
	}
}

// TestLoginLockAfterFails 连续失败 5 次后第 6 次（即使凭据正确）被锁定
func TestLoginLockAfterFails(t *testing.T) {
	const lockUser = "lockuser"
	payload := map[string]interface{}{"username": lockUser, "password": "x"}
	req := func(ip string) map[string]interface{} {
		var buf bytes.Buffer
		_ = json.NewEncoder(&buf).Encode(payload)
		r := httptest.NewRequest(http.MethodPost, "/api/login", &buf)
		r.Header.Set("Content-Type", "application/json")
		r.RemoteAddr = ip
		w := httptest.NewRecorder()
		testEngine.ServeHTTP(w, r)
		var m map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &m)
		return m
	}
	// 前 5 次均因账号不存在失败
	for i := 0; i < 5; i++ {
		m := req("10.99.0.1:12345")
		if m["code"] != float64(500) {
			t.Fatalf("第 %d 次失败应返回 500: %v", i+1, m)
		}
	}
	// 第 6 次：即使密码正确（账号仍不存在则必然失败），应被限流拦截
	m := req("10.99.0.1:12345")
	if m["code"] != float64(500) {
		t.Fatalf("锁定后应返回 500: %v", m)
	}
	if msg, _ := m["msg"].(string); msg != "登录失败次数过多，请稍后再试" {
		t.Fatalf("锁定提示不符: %v", m)
	}
}

func TestAuthRequired(t *testing.T) {
	_, m := doJSON(t, http.MethodGet, "/api/system/user/list", "", nil)
	if m["code"] != float64(401) {
		t.Fatalf("无 token 应返回 401: %v", m)
	}
}

func TestAuthInvalidToken(t *testing.T) {
	_, m := doJSON(t, http.MethodGet, "/api/system/user/list", "fake-token", nil)
	if m["code"] != float64(401) {
		t.Fatalf("伪造 token 应返回 401: %v", m)
	}
}

func TestUserListAsAdmin(t *testing.T) {
	tok := login(t)
	_, m := doJSON(t, http.MethodGet, "/api/system/user/list?pageNum=1&pageSize=10", tok, nil)
	if m["code"] != float64(200) {
		t.Fatalf("用户列表失败: %v", m)
	}
	rows, ok := m["rows"].([]interface{})
	if !ok || len(rows) == 0 {
		t.Fatalf("用户列表为空: %v", m)
	}
	first := rows[0].(map[string]interface{})
	if first["userName"] != "admin" {
		t.Fatalf("首行应为 admin: %v", first)
	}
}

func TestGetInfoAndRouters(t *testing.T) {
	tok := login(t)
	_, m := doJSON(t, http.MethodGet, "/api/getInfo", tok, nil)
	if m["code"] != float64(200) {
		t.Fatalf("getInfo 失败: %v", m)
	}
	data := m["data"].(map[string]interface{})
	roles := data["roles"].([]interface{})
	if len(roles) == 0 || roles[0] != "admin" {
		t.Fatalf("admin 角色缺失: %v", data)
	}
	_, m2 := doJSON(t, http.MethodGet, "/api/getRouters", tok, nil)
	if m2["code"] != float64(200) {
		t.Fatalf("getRouters 失败: %v", m2)
	}
	menus := m2["data"].([]interface{})
	if len(menus) == 0 {
		t.Fatalf("动态路由为空: %v", m2)
	}
}

func TestRoleCreateAndList(t *testing.T) {
	tok := login(t)
	roleKey := fmt.Sprintf("test_role_%d", time.Now().UnixNano())
	_, m := doJSON(t, http.MethodPost, "/api/system/role", tok, map[string]interface{}{
		"roleName": "测试角色", "roleKey": roleKey, "roleSort": 9, "status": "0", "dataScope": "1",
	})
	if m["code"] != float64(200) {
		t.Fatalf("新增角色失败: %v", m)
	}
	_, m2 := doJSON(t, http.MethodGet, "/api/system/role/list?pageNum=1&pageSize=50", tok, nil)
	if m2["code"] != float64(200) {
		t.Fatalf("角色列表失败: %v", m2)
	}
	rows := m2["rows"].([]interface{})
	found := false
	for _, r := range rows {
		if r.(map[string]interface{})["roleKey"] == roleKey {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("新建角色未出现在列表: %v", rows)
	}
}

// TestGenDbList 代码生成表列表（SQLite 方言分支：sqlite_master）
func TestGenDbList(t *testing.T) {
	tok := login(t)
	_, m := doJSON(t, http.MethodGet, "/api/tool/gen/db/list?pageNum=1&pageSize=10", tok, nil)
	if m["code"] != float64(200) {
		t.Fatalf("代码生成表列表失败: %v", m)
	}
	// 本项目全部业务表以 sys_ 开头，均被排除 → data 为空数组而非 nil
	if _, ok := m["data"].([]interface{}); !ok {
		t.Fatalf("data 应为数组: %v", m)
	}
}

// TestDbInfo 数据监控（SQLite 方言分支：sqlite_version/PRAGMA）
func TestDbInfo(t *testing.T) {
	tok := login(t)
	_, m := doJSON(t, http.MethodGet, "/api/monitor/db/info", tok, nil)
	if m["code"] != float64(200) {
		t.Fatalf("数据监控失败: %v", m)
	}
	data := m["data"].(map[string]interface{})
	if data["version"] == "" {
		t.Fatalf("SQLite 版本为空: %v", data)
	}
	if data["journalMode"] != "wal" {
		t.Fatalf("journal_mode 应为 wal（WAL 已开启）: %v", data)
	}
	tables, ok := data["tables"].([]interface{})
	if !ok || len(tables) < 10 {
		t.Fatalf("表统计异常: %v", data)
	}
}

// TestLoginSuccessResetsLimit 限流状态不影响后续正常登录（不同账号不受账号维度锁影响）
func TestLoginSuccessResetsLimit(t *testing.T) {
	login(t)
	_, m := doJSON(t, http.MethodPost, "/api/login", "", map[string]interface{}{
		"username": "admin", "password": "admin123",
	})
	if m["code"] != float64(200) {
		t.Fatalf("限流后正常账号登录失败: %v", m)
	}
}

// TestUtilsTokenHash 工具函数冒烟：token 哈希稳定且不可逆
func TestUtilsTokenHash(t *testing.T) {
	h1 := utils.TokenHash("abc")
	h2 := utils.TokenHash("abc")
	if h1 != h2 {
		t.Fatalf("同一 token 哈希应一致")
	}
	if len(h1) != 64 {
		t.Fatalf("sha256 长度应为 64: %s", h1)
	}
}

// TestOperLogQueueWrites 验证 operlog 的有界队列 + 批量 worker 真能把访问写入 sys_oper_log。
func TestOperLogQueueWrites(t *testing.T) {
	token := login(t)
	const path = "/api/system/post/list"

	var before int64
	if err := dao.DB.Table("sys_oper_log").Where("oper_url = ?", path).Count(&before).Error; err != nil {
		t.Fatalf("count before: %v", err)
	}

	// GET 默认按 1/10 采样；连续 10 次请求至少命中一次采样，避免测试依赖单次命中。
	for i := 0; i < 10; i++ {
		_, m := doJSON(t, http.MethodGet, path+"?pageNum=1&pageSize=10", token, nil)
		if m["code"] != float64(200) {
			t.Fatalf("岗位列表请求失败: %v", m)
		}
	}

	deadline := time.Now().Add(3 * time.Second)
	var after int64
	for time.Now().Before(deadline) {
		if err := dao.DB.Table("sys_oper_log").Where("oper_url = ?", path).Count(&after).Error; err != nil {
			t.Fatalf("count after: %v", err)
		}
		if after > before {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Cleanup(func() {
		dao.DB.Exec("DELETE FROM sys_oper_log WHERE oper_url = ?", path)
	})
	if after <= before {
		t.Fatalf("操作日志未在 3 秒内落库: before=%d after=%d", before, after)
	}
}

// TestOperLogBinaryResponseNotFail xlsx 导出不是 JSON，不应被 operlog 记成失败。
func TestOperLogBinaryResponseNotFail(t *testing.T) {
	token := login(t)
	const path = "/api/system/user/export"

	code, _ := doJSON(t, http.MethodGet, path, token, nil)
	if code != http.StatusOK {
		t.Fatalf("导出请求 HTTP 状态 = %d，期望 200", code)
	}

	deadline := time.Now().Add(3 * time.Second)
	var total int64
	for time.Now().Before(deadline) {
		if err := dao.DB.Table("sys_oper_log").Where("oper_url = ?", path).Count(&total).Error; err != nil {
			t.Fatalf("count: %v", err)
		}
		if total > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Cleanup(func() {
		dao.DB.Exec("DELETE FROM sys_oper_log WHERE oper_url = ?", path)
	})
	if total == 0 {
		t.Fatalf("导出请求未写入操作日志")
	}

	var failed int64
	if err := dao.DB.Table("sys_oper_log").
		Where("oper_url = ? AND status = ?", path, "1").
		Count(&failed).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if failed > 0 {
		t.Fatalf("xlsx 导出被 operlog 记成失败 %d 条", failed)
	}
}
