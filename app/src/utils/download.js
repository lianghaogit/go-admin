// 文件导出（当前仅 H5 支持直接下载）
// 后端导出接口返回 xlsx 二进制，需要携带 Token，无法直接用浏览器地址打开
import { getBaseUrl } from './config'
import { getToken } from './auth'

function dateStamp() {
  const d = new Date()
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}${p(d.getMonth() + 1)}${p(d.getDate())}`
}

export function exportFile(path, params, filename, mime) {
  const query = Object.entries(params || {})
    .filter(([, v]) => v !== undefined && v !== null && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join('&')
  const url = getBaseUrl() + path + (query ? '?' + query : '')
  const token = getToken()

  uni.request({
    url,
    method: 'GET',
    responseType: 'arraybuffer',
    header: token ? { Authorization: 'Bearer ' + token } : {},
    success: (res) => {
      if (res.statusCode !== 200) {
        uni.showToast({ title: `下载失败（${res.statusCode}）`, icon: 'none' })
        return
      }
      // #ifdef H5
      const blob = new Blob([res.data], { type: mime || 'application/octet-stream' })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = filename || `导出_${dateStamp()}.xlsx`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      setTimeout(() => URL.revokeObjectURL(link.href), 1000)
      uni.showToast({ title: '已开始下载', icon: 'none' })
      // #endif
      // #ifndef H5
      uni.showModal({ title: '提示', content: '文件下载请在浏览器/Web 端使用', showCancel: false })
      // #endif
    },
    fail: () => uni.showToast({ title: '下载失败，请检查网络', icon: 'none' })
  })
}

export function exportExcel(path, params, filename) {
  exportFile(path, params, filename, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet')
}

export function fileStamp() {
  return dateStamp()
}
