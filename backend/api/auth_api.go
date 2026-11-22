package api

import (
	"log"
	"strings"
	"time"

	"go-admin/config"
	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// CaptchaImage 生成图形验证码
func CaptchaImage(c *gin.Context) {
	id, b64, _, err := utils.GenerateCaptcha()
	if err != nil {
		utils.Fail(c, 500, "验证码生成失败")
		return
	}
	utils.OkData(c, gin.H{"uuid": id, "img": b64, "captchaEnabled": config.Cfg.Captcha.Enabled})
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
	UUID     string `json:"uuid"`
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	ip := c.ClientIP()

	// 登录防爆破限流：账号/IP 双维度，连续失败达阈值后锁定 15 分钟
	if !utils.LoginUserAllowed(req.Username, ip) {
		utils.Fail(c, 500, "登录失败次数过多，请稍后再试")
		return
	}

	// 验证码校验
	if config.Cfg.Captcha.Enabled {
		if req.UUID == "" || !utils.VerifyCaptcha(req.UUID, req.Code) {
			logLoginLog(req.Username, ip, "1", "验证码错误")
			utils.LoginUserFail(req.Username, ip)
			utils.Fail(c, 500, "验证码错误")
			return
		}
	}

	user, err := dao.GetUserByUsername(req.Username)
	if err != nil || !utils.CheckPassword(user.Password, req.Password) {
		logLoginLog(req.Username, ip, "1", "用户名或密码错误")
		utils.LoginUserFail(req.Username, ip)
		utils.Fail(c, 500, "用户不存在或密码错误")
		return
	}
	if user.Status != "0" {
		logLoginLog(req.Username, ip, "1", "账号已停用")
		utils.LoginUserFail(req.Username, ip)
		utils.Fail(c, 500, "账号已停用，请联系管理员")
		return
	}
	// 登录成功：清除账号维度失败计数
	utils.LoginUserReset(req.Username)

	token, err := utils.GenerateToken(config.Cfg.JWT.Secret, config.Cfg.JWT.Expire, user.ID, user.Username)
	if err != nil {
		utils.Fail(c, 500, "登录失败")
		return
	}
	// 写入在线会话（在线用户/强退依据）
	browser, osName := utils.ParseUA(c.GetHeader("User-Agent"))
	_ = dao.CreateOnline(&model.Online{
		TokenHash: utils.TokenHash(token), UserID: user.ID, UserName: user.Username,
		IP: ip, Browser: browser, OS: osName, LoginTime: time.Now(),
		ExpireAt: time.Now().Add(time.Duration(config.Cfg.JWT.Expire) * time.Minute),
	})
	logLoginLog(req.Username, ip, "0", "登录成功")
	utils.OkData(c, gin.H{"token": token})
}

// logLoginLog 异步写入登录日志（不阻塞登录响应，panic 不影响主流程）
func logLoginLog(username, ip, status, msg string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("登录日志写入 panic 已捕获: %v", r)
			}
		}()
		_ = dao.InsertLoginLog(&model.LoginLog{
			Username: username, IP: ip, Status: status, Msg: msg,
			LoginTime: time.Now(),
		})
	}()
}

// Logout 退出登录（删除在线会话）
func Logout(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		_ = dao.DeleteOnlineByTokenHash(utils.TokenHash(strings.TrimPrefix(header, "Bearer ")))
	}
	utils.Ok(c)
}

// GetInfo 获取当前用户信息、角色与权限
func GetInfo(c *gin.Context) {
	userID := c.GetUint("userId")
	user, err := dao.GetUserByID(userID)
	if err != nil {
		utils.Fail(c, 500, "用户不存在")
		return
	}
	var roles, perms []string
	if dao.IsSuperAdminName(user.Username) {
		roles = []string{"admin"}
		perms = []string{"*:*:*"}
	} else {
		var roleErr, permErr error
		roles, roleErr = dao.GetRoleKeysByUser(userID)
		perms, permErr = dao.GetPermsByUser(userID)
		if roleErr != nil || permErr != nil {
			log.Printf("获取用户 %d 角色/权限失败: %v / %v", userID, roleErr, permErr)
			utils.Fail(c, 500, "获取用户权限失败")
			return
		}
	}
	utils.OkData(c, gin.H{"user": user, "roles": roles, "permissions": perms})
}

// GetRouters 获取当前用户的动态路由菜单
func GetRouters(c *gin.Context) {
	userID := c.GetUint("userId")
	menus, err := dao.GetMenusByUser(userID)
	if err != nil {
		utils.Fail(c, 500, "获取菜单失败")
		return
	}
	utils.OkData(c, dao.BuildMenuTree(menus))
}
