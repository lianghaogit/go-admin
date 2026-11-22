package utils

import (
	"go-admin/cache"

	"github.com/mojocn/base64Captcha"
)

var store base64Captcha.Store = &cache.CaptchaStore{}

// GenerateCaptcha 生成图形验证码，返回 id、base64 图片与答案
func GenerateCaptcha() (id, b64s, answer string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	return captcha.Generate()
}

// VerifyCaptcha 校验验证码（一次性，校验后即失效）
func VerifyCaptcha(id, code string) bool {
	return store.Verify(id, code, true)
}
