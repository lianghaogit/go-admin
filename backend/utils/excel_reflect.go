package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// excelUnmarshaler 由字段类型自行把单元格文本转成自身值（如 model.LocalTime）
type excelUnmarshaler interface {
	UnmarshalExcel(string) error
}

// excelMarshaler 由字段类型自行输出单元格文本
type excelMarshaler interface {
	ExcelString() string
}

// ExcelRow 按结构体字段声明顺序输出一行单元格。
// 切片字段（如 Children）不落 Excel；实现了 excelMarshaler 的类型优先使用其文本。
func ExcelRow(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return []interface{}{fmt.Sprintf("%v", v)}
	}
	rt := rv.Type()
	cells := make([]interface{}, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		f := rv.Field(i)
		if f.Kind() == reflect.Slice || f.Kind() == reflect.Map || f.Kind() == reflect.Array {
			continue
		}
		if rt.Field(i).Tag.Get("excel") == "-" {
			// 与导入保持一致：显式标记不参与 Excel 的字段不输出
			continue
		}
		if !f.CanInterface() {
			continue
		}
		if m, ok := f.Interface().(excelMarshaler); ok {
			cells = append(cells, m.ExcelString())
			continue
		}
		switch f.Kind() {
		case reflect.Bool:
			if f.Bool() {
				cells = append(cells, "是")
			} else {
				cells = append(cells, "否")
			}
		case reflect.String:
			cells = append(cells, f.String())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			cells = append(cells, f.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			cells = append(cells, f.Uint())
		case reflect.Float32, reflect.Float64:
			cells = append(cells, f.Float())
		default:
			cells = append(cells, fmt.Sprintf("%v", f.Interface()))
		}
	}
	return cells
}

// UnmarshalExcelRow 按表头名称把一行文本填进结构体（依据字段的 excel 标签）。
// 空单元格保持字段零值；未知表头忽略；excel:"-" 的字段不参与导入。
func UnmarshalExcelRow(dst interface{}, headers []string, row []string) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("导入目标必须是非空指针")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("导入目标必须是结构体")
	}
	rt := rv.Type()

	index := make(map[string]int, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		tag := rt.Field(i).Tag.Get("excel")
		if tag == "" || tag == "-" {
			continue
		}
		name, _, _ := strings.Cut(tag, ",")
		if name != "" {
			index[name] = i
		}
	}

	matched := 0
	for col, h := range headers {
		name := strings.TrimSpace(h)
		fi, ok := index[name]
		if !ok {
			continue
		}
		matched++
		cell := ""
		if col < len(row) {
			cell = strings.TrimSpace(row[col])
		}
		if cell == "" {
			continue
		}
		f := rv.Field(fi)
		if f.CanAddr() {
			if u, ok := f.Addr().Interface().(excelUnmarshaler); ok {
				if err := u.UnmarshalExcel(cell); err != nil {
					return fmt.Errorf("列[%s]：%v", name, err)
				}
				continue
			}
		}
		switch f.Kind() {
		case reflect.String:
			f.SetString(cell)
		case reflect.Bool:
			b, err := parseExcelBool(cell)
			if err != nil {
				return fmt.Errorf("列[%s]：%v", name, err)
			}
			f.SetBool(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(cell, 10, 64)
			if err != nil {
				return fmt.Errorf("列[%s]：必须是整数", name)
			}
			f.SetInt(n)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			n, err := strconv.ParseUint(cell, 10, 64)
			if err != nil {
				return fmt.Errorf("列[%s]：必须是非负整数", name)
			}
			f.SetUint(n)
		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(cell, 64)
			if err != nil {
				return fmt.Errorf("列[%s]：必须是数字", name)
			}
			f.SetFloat(n)
		default:
			return fmt.Errorf("列[%s]：不支持自动导入该字段类型", name)
		}
	}
	if matched == 0 {
		return fmt.Errorf("表头与导入模板不一致")
	}
	return nil
}

func parseExcelBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "是", "启用", "正常":
		return true, nil
	case "0", "false", "no", "n", "否", "停用":
		return false, nil
	default:
		return false, fmt.Errorf("无法识别为布尔值（可用 是/否、1/0、true/false）")
	}
}
