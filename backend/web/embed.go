// Package web 内嵌前端构建产物，使后端单文件二进制即可同时提供 API 与 Web 界面。
// dist 目录由构建脚本从 frontend/dist 复制而来（不包含 .gz/.br 预压缩文件）。
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
