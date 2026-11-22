// 通用表单校验工具：统一各页面的必填、格式与数值校验规则

export const MOBILE_RE = /^1[3-9]\d{9}$/
export const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
export const CODE_RE = /^[A-Za-z][A-Za-z0-9_-]*$/
export const DICT_TYPE_RE = /^[A-Za-z][A-Za-z0-9_]*$/
export const CONFIG_KEY_RE = /^[A-Za-z][A-Za-z0-9._-]*$/

export function trimStr(value) {
  return value == null ? '' : String(value).trim()
}

export function isMobile(value) {
  return MOBILE_RE.test(trimStr(value))
}

export function isEmail(value) {
  return EMAIL_RE.test(trimStr(value))
}

// 非负整数（允许 0）
export function isNonNegInt(value) {
  const n = Number(value)
  return Number.isInteger(n) && n >= 0
}

// Cron 表达式：6 段（秒 分 时 日 月 周）
export function isCron(value) {
  const parts = trimStr(value).split(/\s+/)
  return parts.length === 6 && parts.every((p) => p.length > 0)
}
