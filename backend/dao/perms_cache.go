package dao

import (
	"sync"
	"time"
)

// 用户按钮权限的进程内缓存。
// 角色/菜单/用户授权变更时调用 InvalidatePermCache 立即失效；未覆盖到的变更最多陈旧 30 秒。
// 注意：这是进程内缓存，多实例部署时其他实例的变更不会主动失效（属于 B 路径改造范围）。
const permsCacheTTL = 30 * time.Second

type permsCacheEntry struct {
	perms    []string
	expireAt time.Time
}

var (
	permsCacheMu sync.RWMutex
	permsCache   = map[uint]permsCacheEntry{}
)

func getCachedPerms(userID uint) ([]string, bool) {
	permsCacheMu.RLock()
	e, ok := permsCache[userID]
	permsCacheMu.RUnlock()
	if !ok || time.Now().After(e.expireAt) {
		return nil, false
	}
	return cloneStrings(e.perms), true
}

func cachePerms(userID uint, perms []string) {
	permsCacheMu.Lock()
	defer permsCacheMu.Unlock()
	permsCache[userID] = permsCacheEntry{perms: cloneStrings(perms), expireAt: time.Now().Add(permsCacheTTL)}
}

// InvalidatePermCache 清空权限缓存（角色、菜单、用户授权变更后调用）。
func InvalidatePermCache() {
	permsCacheMu.Lock()
	permsCache = map[uint]permsCacheEntry{}
	permsCacheMu.Unlock()
}

func cloneStrings(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
