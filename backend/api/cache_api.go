package api

import (
	"go-admin/cache"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// CacheInfo 缓存监控：统计信息 + 键分组
func CacheInfo(c *gin.Context) {
	keyCount, hitCount, missCount, hitRate, memBytes := cache.Info()
	utils.OkData(c, gin.H{
		"type":       "内置内存缓存",
		"keyCount":   keyCount,
		"hitCount":   hitCount,
		"missCount":  missCount,
		"hitRate":    hitRate,
		"memBytes":   memBytes,
		"startTime":  cache.StartTime().Format("2006-01-02 15:04:05"),
		"expireUnit": "秒",
		"groups":     cache.GroupKeys(cache.Keys()),
	})
}

// CacheKeyDetail 查询单个缓存键详情
func CacheKeyDetail(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		utils.Fail(c, 500, "请指定缓存键名")
		return
	}
	it, ok := cache.GetItem(key)
	if !ok {
		utils.Fail(c, 500, "缓存键不存在或已过期")
		return
	}
	utils.OkData(c, gin.H{
		"key":      key,
		"value":    it.Value,
		"createAt": it.CreateAt.Format("2006-01-02 15:04:05"),
		"expireAt": it.ExpireAt.Format("2006-01-02 15:04:05"),
		"ttl":      int64(it.ExpireAt.Sub(it.CreateAt).Seconds()),
	})
}

// CacheClear 清空全部缓存
func CacheClear(c *gin.Context) {
	cache.Flush()
	utils.Ok(c)
}
