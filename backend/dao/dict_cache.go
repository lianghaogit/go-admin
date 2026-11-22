package dao

import (
	"sync"
	"time"

	"go-admin/model"
)

// 字典数据内存缓存：按 dictType 缓存下拉数据，TTL 60s，
// 字典数据增删改时主动失效，降低高频下拉接口的数据库压力。
var (
	dictCache    sync.Map // key: dictType, value: dictCacheItem
	dictCacheTTL = 60 * time.Second
)

type dictCacheItem struct {
	data     []model.DictData
	expireAt time.Time
}

// GetDictDataCache 读取字典缓存（命中且未过期返回 true）
func GetDictDataCache(dictType string) ([]model.DictData, bool) {
	if v, ok := dictCache.Load(dictType); ok {
		item := v.(dictCacheItem)
		if time.Now().Before(item.expireAt) {
			return item.data, true
		}
		dictCache.Delete(dictType)
	}
	return nil, false
}

// SetDictDataCache 写入字典缓存
func SetDictDataCache(dictType string, data []model.DictData) {
	dictCache.Store(dictType, dictCacheItem{data: data, expireAt: time.Now().Add(dictCacheTTL)})
}

// ClearDictCache 清除指定字典缓存（dictType 为空时清空全部）
func ClearDictCache(dictType string) {
	if dictType == "" {
		dictCache.Range(func(k, _ interface{}) bool {
			dictCache.Delete(k)
			return true
		})
		return
	}
	dictCache.Delete(dictType)
}
