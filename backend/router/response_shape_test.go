package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

type responseShapeEndpoint struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// TestResponseShape 对核心接口的顶层/数据结构做 golden 快照，只校验字段形状，不校验业务值。
// 更新方式：UPDATE_GOLDEN=1 go test ./router -run TestResponseShape
func TestResponseShape(t *testing.T) {
	tok := login(t)
	checks := []responseShapeEndpoint{
		{"getInfo", "/api/getInfo"},
		{"getRouters", "/api/getRouters"},
		{"dashboardStats", "/api/dashboard/stats"},
		{"userList", "/api/system/user/list?pageNum=1&pageSize=10"},
		{"roleList", "/api/system/role/list?pageNum=1&pageSize=10"},
		{"menuList", "/api/system/menu/list"},
		{"deptList", "/api/system/dept/list"},
		{"dictTypeList", "/api/system/dict/type/list?pageNum=1&pageSize=10"},
		{"postList", "/api/system/post/list?pageNum=1&pageSize=10"},
		{"configList", "/api/system/config/list?pageNum=1&pageSize=10"},
		{"noticeList", "/api/system/notice/list?pageNum=1&pageSize=10"},
		{"jobList", "/api/monitor/job/list?pageNum=1&pageSize=10"},
		{"apiList", "/api/system/api/list?pageNum=1&pageSize=10"},
		{"genList", "/api/tool/gen/list?pageNum=1&pageSize=10"},
		{"flowDefList", "/api/flow/def/list?pageNum=1&pageSize=10"},
	}

	shapes := make(map[string]interface{}, len(checks))
	for _, c := range checks {
		_, m := doJSON(t, http.MethodGet, c.Path, tok, nil)
		if m["code"] != float64(200) {
			t.Fatalf("%s 返回失败: %v", c.Name, m)
		}
		shapes[c.Name] = shapeOf(m)
	}
	data, err := json.MarshalIndent(shapes, "", "  ")
	if err != nil {
		t.Fatalf("序列化响应形状失败: %v", err)
	}
	data = append(data, '\n')

	golden := filepath.Join("testdata", "response_shapes.golden.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatalf("创建 testdata 失败: %v", err)
		}
		if err := os.WriteFile(golden, data, 0o644); err != nil {
			t.Fatalf("写入 golden 失败: %v", err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("读取 %s 失败（可先运行 UPDATE_GOLDEN=1 go test ./router -run TestResponseShape）: %v", golden, err)
	}
	if !bytes.Equal(want, data) {
		t.Fatalf("核心接口响应形状发生变化，请检查并更新 %s", golden)
	}
}

// shapeOf 保留 map 键、数组元素形状和标量类型，丢弃具体值。
func shapeOf(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]interface{}, len(keys))
		for _, k := range keys {
			out[k] = shapeOf(x[k])
		}
		return out
	case []interface{}:
		if len(x) == 0 {
			return []interface{}{}
		}
		return []interface{}{shapeOf(x[0])}
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("%T", x)
	}
}
