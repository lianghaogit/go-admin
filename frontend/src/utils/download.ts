import { ElMessage } from 'element-plus';

// 下载 Blob 文件（xlsx/zip 等）
export function downloadBlob(blob, filename) {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}

// 从 Content-Disposition 提取文件名（含中文解码）
export function parseFilename(disposition) {
  if (!disposition) return '';
  // filename*=UTF-8''xxx 优先
  const star = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (star) {
    try {
      return decodeURIComponent(star[1]);
    } catch (e) {
      return star[1];
    }
  }
  const plain = disposition.match(/filename="?([^";]+)"?/i);
  return plain ? plain[1] : '';
}

// 通用文件导出：请求 → Blob → 触发下载（失败时尝试解析 JSON 错误信息）
export async function exportFile(requestPromise, defaultFilename) {
  try {
    const response = await requestPromise;
    const disposition = response.headers?.['content-disposition'] || '';
    const filename = parseFilename(disposition) || defaultFilename;
    downloadBlob(response.data, filename);
    return true;
  } catch (e) {
    // 服务端业务错误以 JSON 返回时，尝试提取 msg
    if (e.response && e.response.data instanceof Blob) {
      try {
        const text = await e.response.data.text();
        const json = JSON.parse(text);
        if (json.msg) {
          ElMessage.error(json.msg);
          return false;
        }
      } catch (err) {
        // 非 JSON
      }
    }
    return false;
  }
}
