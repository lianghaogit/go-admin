package router

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

type routeContractRow struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// TestRouteContract 对路由表做快照，新增/删除/改路径必须显式更新 golden 文件。
// 更新方式：UPDATE_GOLDEN=1 go test ./router -run TestRouteContract
func TestRouteContract(t *testing.T) {
	routes := testEngine.Routes()
	rows := make([]routeContractRow, 0, len(routes))
	for _, r := range routes {
		rows = append(rows, routeContractRow{Method: r.Method, Path: r.Path, Handler: r.Handler})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Method != rows[j].Method {
			return rows[i].Method < rows[j].Method
		}
		return rows[i].Path < rows[j].Path
	})
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatalf("序列化路由快照失败: %v", err)
	}
	data = append(data, '\n')

	golden := filepath.Join("testdata", "routes.golden.json")
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
		t.Fatalf("读取 %s 失败（可先运行 UPDATE_GOLDEN=1 go test ./router -run TestRouteContract）: %v", golden, err)
	}
	if !bytes.Equal(want, data) {
		t.Fatalf("路由契约发生变化，请检查并更新 %s", golden)
	}
}
