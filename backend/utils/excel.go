package utils

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ExcelSheet 一个工作表：名称 + 表头 + 数据行
type ExcelSheet struct {
	Name    string
	Headers []string
	Rows    [][]interface{}
}

// BuildExcel 由表头 + 数据行生成 xlsx 字节流（浅色表头样式）
func BuildExcel(sheet string, headers []string, rows [][]interface{}) ([]byte, error) {
	return BuildExcelSheets([]ExcelSheet{{Name: sheet, Headers: headers, Rows: rows}})
}

// BuildExcelSheets 生成包含多个工作表的 xlsx（主子表导出/模板使用）
func BuildExcelSheets(sheets []ExcelSheet) ([]byte, error) {
	if len(sheets) == 0 {
		return nil, fmt.Errorf("至少需要一个工作表")
	}
	f := excelize.NewFile()
	defer f.Close()

	// 表头样式：加粗 + 背景色 #409EFF + 白字
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF", Size: 11},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#409EFF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	// 数据单元格样式：垂直居中
	cellStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	defaultSheet := f.GetSheetName(0)
	for s, sh := range sheets {
		name := sh.Name
		if name == "" {
			name = fmt.Sprintf("Sheet%d", s+1)
		}
		if s == 0 {
			// 第一个工作表复用默认 Sheet，避免留下空白页
			if err := f.SetSheetName(defaultSheet, name); err != nil {
				return nil, err
			}
		} else {
			if _, err := f.NewSheet(name); err != nil {
				return nil, err
			}
		}

		for i, h := range sh.Headers {
			colName, err := excelize.ColumnNumberToName(i + 1)
			if err != nil {
				return nil, err
			}
			cell := fmt.Sprintf("%s%d", colName, 1)
			if err := f.SetCellValue(name, cell, h); err != nil {
				return nil, err
			}
			if err := f.SetCellStyle(name, cell, cell, headerStyle); err != nil {
				return nil, err
			}
		}

		for r, row := range sh.Rows {
			// 行数据可能比表头长，超出部分直接忽略，避免列名/宽度越界
			for i, v := range row {
				if i >= len(sh.Headers) {
					break
				}
				colName, err := excelize.ColumnNumberToName(i + 1)
				if err != nil {
					return nil, err
				}
				cell := fmt.Sprintf("%s%d", colName, r+2)
				if err := f.SetCellValue(name, cell, v); err != nil {
					return nil, err
				}
				if err := f.SetCellStyle(name, cell, cell, cellStyle); err != nil {
					return nil, err
				}
			}
		}

		// 列宽自适应（按内容估算）
		widths := make([]float64, len(sh.Headers))
		for _, row := range sh.Rows {
			for i, v := range row {
				if i >= len(widths) {
					break
				}
				l := len(fmt.Sprintf("%v", v)) + 2
				if l > 0 && float64(l) > widths[i] {
					widths[i] = float64(l)
				}
			}
		}
		for i := range sh.Headers {
			if len(sh.Headers[i])+2 > int(widths[i]) {
				widths[i] = float64(len(sh.Headers[i]) + 2)
			}
			if widths[i] < 10 {
				widths[i] = 10
			}
			colName, err := excelize.ColumnNumberToName(i + 1)
			if err != nil {
				return nil, err
			}
			if err := f.SetColWidth(name, colName, colName, widths[i]); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
