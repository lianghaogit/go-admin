package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-admin/config"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 上传类型白名单（扩展名 → 说明）
var uploadExtAllow = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif",
	".bmp": "image/bmp", ".webp": "image/webp", ".svg": "image/svg+xml",
	".pdf":  "application/pdf",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".xls": "application/vnd.ms-excel",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document", ".doc": "application/msword",
	".txt": "text/plain", ".md": "text/markdown",
	".zip": "application/zip", ".rar": "application/x-rar-compressed", ".tar": "application/x-tar", ".gz": "application/gzip",
	".mp3": "audio/mpeg", ".mp4": "video/mp4",
}

const maxUploadSize = 10 * 1024 * 1024 // 10MB

// UploadFile 通用文件上传：保存到 uploads/yyyyMMdd/ 目录，返回访问 URL
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Fail(c, 500, "请选择要上传的文件")
		return
	}
	if file.Size > maxUploadSize {
		utils.Fail(c, 500, "文件大小不能超过 10MB")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := uploadExtAllow[ext]; !ok {
		utils.Fail(c, 500, "不支持的文件类型，允许：图片/PDF/Office 文档/文本/压缩包/音视频")
		return
	}

	baseDir := config.Cfg.Server.UploadDir
	if baseDir == "" {
		baseDir = "uploads"
	}
	sub := time.Now().Format("20060102")
	dir := filepath.Join(baseDir, sub)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		utils.Fail(c, 500, "创建上传目录失败")
		return
	}

	newFileName := uuid.NewString() + ext
	dst := filepath.Join(dir, newFileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.Fail(c, 500, "文件保存失败")
		return
	}

	utils.OkData(c, gin.H{
		"fileName":     newFileName,
		"originalName": file.Filename,
		"url":          fmt.Sprintf("/uploads/%s/%s", sub, newFileName),
		"size":         file.Size,
	})
}
