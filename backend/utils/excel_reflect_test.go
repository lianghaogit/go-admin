package utils

import (
	"reflect"
	"testing"
	"time"

	"go-admin/model"
)

type excelReflectRow struct {
	Name     string          `excel:"名称"`
	Hidden   string          `excel:"-"`
	Enabled  bool            `excel:"启用"`
	Amount   float64         `excel:"金额"`
	CreateAt model.LocalTime `excel:"创建时间"`
	Ignore   string          // 未打 tag，不参与导入
}

func TestExcelRowHonorsExcludeTagAndOrder(t *testing.T) {
	row := excelReflectRow{
		Name:     "测试",
		Hidden:   "不应导出",
		Enabled:  true,
		Amount:   12.5,
		CreateAt: model.LocalTime(time.Date(2026, 11, 15, 9, 30, 0, 0, time.Local)),
	}
	cells := ExcelRow(row)
	want := []interface{}{"测试", "是", 12.5, "2026-11-15 09:30:00", ""}
	if !reflect.DeepEqual(cells, want) {
		t.Fatalf("ExcelRow = %#v, want %#v", cells, want)
	}
}

func TestUnmarshalExcelRowByNameAndSkip(t *testing.T) {
	var dst excelReflectRow
	err := UnmarshalExcelRow(&dst, []string{"忽略列", "名称", "启用", "金额", "创建时间", "未知"}, []string{"x", "测试", "否", "12.5", "2026-11-15 09:30:00", "y"})
	if err != nil {
		t.Fatalf("UnmarshalExcelRow 失败: %v", err)
	}
	if dst.Name != "测试" || dst.Enabled || dst.Amount != 12.5 {
		t.Fatalf("字段回填不正确: %+v", dst)
	}
	if got := time.Time(dst.CreateAt).Format("2006-01-02 15:04:05"); got != "2026-11-15 09:30:00" {
		t.Fatalf("时间回填 = %s", got)
	}
	if dst.Hidden != "" || dst.Ignore != "" {
		t.Fatalf("excel:\"-\" 或无 tag 字段被错误回填: %+v", dst)
	}
}

func TestUnmarshalExcelRowRejectsUnknownHeader(t *testing.T) {
	var dst excelReflectRow
	if err := UnmarshalExcelRow(&dst, []string{"未知"}, []string{"x"}); err == nil {
		t.Fatal("表头与模板完全不一致时应报错")
	}
}

func TestUnmarshalExcelRowEmptyCellKeepsValue(t *testing.T) {
	dst := excelReflectRow{Name: "保留", Enabled: true, Amount: 2.5}
	err := UnmarshalExcelRow(&dst, []string{"名称", "启用", "金额"}, []string{"", "", ""})
	if err != nil {
		t.Fatalf("空单元格不应报错: %v", err)
	}
	if dst.Name != "保留" || !dst.Enabled || dst.Amount != 2.5 {
		t.Fatalf("空单元格不应覆盖原值: %+v", dst)
	}
}
