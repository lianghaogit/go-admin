package cache

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// Item 缓存项
type Item struct {
	Value    string    `json:"value"`
	CreateAt time.Time `json:"createAt"`
	ExpireAt time.Time `json:"expireAt"`
}

var (
	mu        sync.Mutex
	items     = map[string]Item{}
	hits      int64
	misses    int64
	startTime = time.Now()
)

// Set 写入缓存
func Set(key, value string, ttl time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	items[key] = Item{Value: value, CreateAt: time.Now(), ExpireAt: time.Now().Add(ttl)}
}

// Get 读取缓存（统计命中/未命中；过期自动淘汰）
func Get(key string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	it, ok := items[key]
	if !ok {
		misses++
		return "", false
	}
	if time.Now().After(it.ExpireAt) {
		delete(items, key)
		misses++
		return "", false
	}
	hits++
	return it.Value, true
}

// GetItem 读取缓存项详情（含过期时间）
func GetItem(key string) (Item, bool) {
	mu.Lock()
	defer mu.Unlock()
	it, ok := items[key]
	if !ok {
		return Item{}, false
	}
	if time.Now().After(it.ExpireAt) {
		delete(items, key)
		return Item{}, false
	}
	return it, true
}

// Delete 删除单个键
func Delete(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(items, key)
}

// Flush 清空全部缓存
func Flush() {
	mu.Lock()
	defer mu.Unlock()
	items = map[string]Item{}
	hits = 0
	misses = 0
}

// Keys 全部有效键
func Keys() []string {
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	keys := make([]string, 0, len(items))
	for k, it := range items {
		if now.After(it.ExpireAt) {
			delete(items, k)
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Info 缓存统计信息
func Info() (keyCount, hitCount, missCount, hitRate, memBytes int64) {
	mu.Lock()
	defer mu.Unlock()
	for _, it := range items {
		memBytes += int64(len(it.Value)) + int64(len(it.CreateAt.String()))
	}
	keyCount = int64(len(items))
	hitCount = hits
	missCount = misses
	total := hits + misses
	if total > 0 {
		hitRate = hits * 100 / total
	}
	return
}

// StartTime 缓存服务启动时间
func StartTime() time.Time {
	return startTime
}

// GroupKeys 按前缀分组键名
func GroupKeys(keys []string) map[string][]string {
	groups := map[string][]string{}
	for _, k := range keys {
		prefix := "other"
		if strings.HasPrefix(k, "captcha:") {
			prefix = "captcha"
		} else if strings.HasPrefix(k, "online:") {
			prefix = "online"
		}
		groups[prefix] = append(groups[prefix], k)
	}
	for _, v := range groups {
		sort.Strings(v)
	}
	return groups
}

// CaptchaStore base64Captcha 自定义存储：验证码接入内存缓存（可被缓存监控统计）
type CaptchaStore struct{}

// Set 保存验证码，5 分钟过期
func (s *CaptchaStore) Set(id, value string) error {
	Set("captcha:"+id, value, 5*time.Minute)
	return nil
}

// Get 获取验证码，clear=true 时取后删除
func (s *CaptchaStore) Get(id string, clear bool) string {
	v, ok := Get("captcha:" + id)
	if ok && clear {
		Delete("captcha:" + id)
	}
	return v
}

// Verify 校验验证码
func (s *CaptchaStore) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v != "" && v == answer
}
