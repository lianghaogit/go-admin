package dao

import (
	"log"
	"time"

	"go-admin/model"
)

// OnlineQuery 在线用户查询条件
type OnlineQuery struct {
	PageNum  int
	PageSize int
	UserName string
}

// CreateOnline 写入在线会话（已存在则覆盖登录时间）
func CreateOnline(o *model.Online) error {
	var cnt int64
	if err := DB.Model(&model.Online{}).Where("token_hash = ?", o.TokenHash).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return DB.Model(&model.Online{}).Where("token_hash = ?", o.TokenHash).
			Updates(map[string]interface{}{"user_id": o.UserID, "user_name": o.UserName, "ip": o.IP, "browser": o.Browser, "os": o.OS, "login_time": o.LoginTime, "expire_at": o.ExpireAt}).Error
	}
	return DB.Create(o).Error
}

// DeleteOnlineByTokenHash 删除会话（退出/强退）
func DeleteOnlineByTokenHash(tokenHash string) error {
	return DB.Where("token_hash = ?", tokenHash).Delete(&model.Online{}).Error
}

// OnlineExists 会话是否存在且未过期（鉴权中间件使用）。
// 用 SELECT 1 ... LIMIT 1 替代 COUNT(*)，并排除过期/无 expire_at 的历史会话。
func OnlineExists(tokenHash string) (bool, error) {
	var one int
	err := DB.Raw(`SELECT 1 FROM sys_online WHERE token_hash = ? AND expire_at IS NOT NULL AND expire_at > ? LIMIT 1`,
		tokenHash, time.Now()).Scan(&one).Error
	return one > 0, err
}

// DeleteExpiredOnlines 清理过期会话（含历史遗留无过期时间的会话），返回清理条数
func DeleteExpiredOnlines() (int64, error) {
	res := DB.Exec("DELETE FROM sys_online WHERE expire_at IS NULL OR expire_at < ?", time.Now())
	return res.RowsAffected, res.Error
}

// ListOnlines 分页查询在线用户（先清理过期会话）
func ListOnlines(q OnlineQuery) ([]model.Online, int64, error) {
	// 清理已过期会话（含历史遗留无过期时间的会话），避免在线列表只增不减
	if _, err := DeleteExpiredOnlines(); err != nil {
		return nil, 0, err
	}
	var list []model.Online
	var total int64
	db := DB.Model(&model.Online{})
	if q.UserName != "" {
		db = db.Where(likeExprEsc("user_name"), "%"+escapeLike(q.UserName)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("login_time desc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&list).Error
	return list, total, err
}

// StartOnlineCleanup 周期清理过期在线会话（启动先清一次，之后每 interval 一次）。
// 与 StartLogRetention 共用 stop channel，退出时由 WaitBackground 等待。
func StartOnlineCleanup(stop <-chan struct{}, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	cleanup := func() {
		n, err := DeleteExpiredOnlines()
		if err != nil {
			log.Printf("清理过期在线会话失败: %v", err)
			return
		}
		if n > 0 {
			log.Printf("清理过期在线会话 %d 条", n)
		}
	}
	backgroundWG.Add(1)
	go func() {
		defer backgroundWG.Done()
		cleanup()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				cleanup()
			}
		}
	}()
}

// DeleteAllOnlines 清空全部在线会话
func DeleteAllOnlines() error {
	return DB.Exec("DELETE FROM sys_online").Error
}
