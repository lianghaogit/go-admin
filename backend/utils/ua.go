package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// TokenHash 计算 token 的 SHA-256 摘要（在线会话主键）
func TokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// ParseUA 从 User-Agent 解析浏览器与操作系统
func ParseUA(ua string) (browser, os string) {
	if ua == "" {
		return "未知", "未知"
	}
	switch {
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	case strings.Contains(ua, "MicroMessenger"):
		browser = "微信内置浏览器"
	default:
		browser = "其他"
	}
	switch {
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Mac OS X"):
		os = "macOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad"):
		os = "iOS"
	default:
		os = "其他"
	}
	return browser, os
}
