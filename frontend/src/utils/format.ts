// 时间格式化工具（对标 RuoYi parseTime）
// 支持 Date 对象 / 时间戳 / 可解析的时间字符串（含 Go RFC3339: 2026-10-23T08:56:48+08:00）

export function parseTime(
  time: string | number | Date | null | undefined,
  pattern = 'YYYY-MM-DD HH:mm:ss'
) {
  if (!time && time !== 0) {
    return '';
  }
  let date;
  if (time instanceof Date) {
    date = time;
  } else if (typeof time === 'number' || (typeof time === 'string' && /^\d+$/.test(time))) {
    date = new Date(parseInt(String(time), 10));
  } else {
    date = new Date(time);
  }
  // 无法解析的时间原样返回，避免显示 Invalid Date
  if (isNaN(date.getTime())) {
    return String(time);
  }
  const formatObj = {
    YYYY: date.getFullYear(),
    MM: String(date.getMonth() + 1).padStart(2, '0'),
    DD: String(date.getDate()).padStart(2, '0'),
    HH: String(date.getHours()).padStart(2, '0'),
    mm: String(date.getMinutes()).padStart(2, '0'),
    ss: String(date.getSeconds()).padStart(2, '0')
  };
  return pattern.replace(/YYYY|MM|DD|HH|mm|ss/g, (key) => formatObj[key]);
}
