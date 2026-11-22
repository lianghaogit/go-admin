package dao

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderVueTemplates 代码生成前端模板 TS 化渲染验证
func TestRenderVueTemplates(t *testing.T) {
	// go test 的工作目录是包目录（dao/），模板在 backend/templates → 切到上级
	wd, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(wd, "..")); err != nil {
		t.Fatalf("切换目录失败: %v", err)
	}
	defer os.Chdir(wd)

	data := &GenData{
		ModuleName:   "system",
		ClassName:    "TestItem",
		ClassVar:     "testItem",
		ClassSnake:   "test_item",
		TableName:    "sys_test_item",
		TableComment: "测试项",
		FunctionName: "测试项",
		IDJsonField:  "id",
		Fields: []GenField{
			{JsonField: "id", GoType: "int64", TsType: "number", Comment: "编号", IsPk: true, IsQuery: true},
			{JsonField: "name", GoType: "string", TsType: "string", Comment: "名称", IsList: true, IsQuery: true, IsForm: true},
			{JsonField: "count", GoType: "int", TsType: "number", Comment: "数量", IsList: true, IsForm: true},
			{JsonField: "active", GoType: "bool", TsType: "boolean", Comment: "启用", IsForm: true},
			{JsonField: "createTime", GoType: "time.Time", TsType: "string", Comment: "创建时间", IsList: true},
		},
	}

	api, err := renderTemplate("vue_api.tmpl", data)
	if err != nil {
		t.Fatalf("渲染 vue_api.tmpl 失败: %v", err)
	}
	for _, want := range []string{
		"export interface TestItemQuery",
		"export interface TestItemRow",
		"pageNum?: number",
		"name?: string",
		"active?: boolean",
		"export function listTestItem(params: TestItemQuery)",
		"delTestItem(id: number | string)",
	} {
		if !strings.Contains(api, want) {
			t.Errorf("api.ts 缺少 %q\n---\n%s", want, api)
		}
	}

	index, err := renderTemplate("vue_index.tmpl", data)
	if err != nil {
		t.Fatalf("渲染 vue_index.tmpl 失败: %v", err)
	}
	for _, want := range []string{
		`<script setup lang="ts">`,
		"type TestItemQuery",
		"type TestItemRow",
		"reactive<TestItemQuery>",
		"reactive<TestItemRow>",
		"function openDialog(row?: TestItemRow)",
		"function handleDelete(row: TestItemRow)",
		"ref<TestItemRow[]>([])",
	} {
		if !strings.Contains(index, want) {
			t.Errorf("index.vue 缺少 %q\n---\n%s", want, index)
		}
	}
}
