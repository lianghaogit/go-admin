package api

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// pageNum 获取当前页码
func pageNum(c *gin.Context) int {
	n, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	if n < 1 {
		n = 1
	}
	return n
}

// pageSize 获取每页条数
func pageSize(c *gin.Context) int {
	n, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if n < 1 || n > 100 {
		n = 10
	}
	return n
}

// idParam 解析路径中的 id 参数
func idParam(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}

// paramUint 解析路径中指定名称的参数
func paramUint(c *gin.Context, name string) uint {
	v, _ := strconv.ParseUint(c.Param(name), 10, 64)
	return uint(v)
}

// parseUint 解析字符串为 uint（非法返回 0）
func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

// mustInt64 解析 int64 查询参数（非法值返回 0），生成代码共用，避免各模块重复声明
func mustInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// mustFloat64 解析 float64 查询参数（非法值返回 0）
func mustFloat64(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// excelAttachment 设置 Excel 下载响应头：ASCII 兜底名 + RFC 5987 中文名，
// 避免浏览器把 UTF-8 原文当 Latin-1 显示成乱码。
func excelAttachment(c *gin.Context, filename string) {
	c.Header("Content-Disposition", "attachment; filename=\"download.xlsx\"; filename*=UTF-8''"+url.PathEscape(filename))
}

// excelCell 安全读取一行中的第 idx 列（越界返回空字符串）
func excelCell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

// excelRowEmpty 判断整行是否为空（用于跳过 Excel 末尾空行）
func excelRowEmpty(row []string) bool {
	for _, v := range row {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}
