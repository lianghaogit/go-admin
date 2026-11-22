import { describe, it, expect } from 'vitest';
import { parseTime } from '@/utils/format';

// 时间格式化（parseTime）单元测试
// 注：断言基于本地时区构造的 Date，getFullYear/getHours 等均为本地时区方法，结果自洽稳定

describe('parseTime', () => {
  it('空值返回空串', () => {
    expect(parseTime(null)).toBe('');
    expect(parseTime(undefined)).toBe('');
  });

  it('Date 对象格式化', () => {
    const d = new Date(2026, 0, 2, 3, 4, 5); // 2026-01-02 03:04:05 本地时区
    expect(parseTime(d)).toBe('2026-01-02 03:04:05');
  });

  it('毫秒时间戳格式化', () => {
    const d = new Date(2026, 5, 15, 12, 30, 45);
    expect(parseTime(d.getTime())).toBe('2026-06-15 12:30:45');
  });

  it('数字字符串时间戳格式化', () => {
    const d = new Date(2026, 11, 31, 23, 59, 59);
    expect(parseTime(String(d.getTime()))).toBe('2026-12-31 23:59:59');
  });

  it('Go RFC3339 字符串解析（含时区偏移）', () => {
    // 构造本地时区对应的 RFC3339 字符串
    const d = new Date(2026, 3, 1, 8, 0, 0);
    const iso = d.toISOString();
    expect(parseTime(iso)).toBe('2026-04-01 08:00:00');
  });

  it('自定义 pattern', () => {
    const d = new Date(2026, 0, 2, 3, 4, 5);
    expect(parseTime(d, 'YYYY-MM-DD')).toBe('2026-01-02');
    expect(parseTime(d, 'HH:mm')).toBe('03:04');
  });

  it('无法解析的输入原样返回', () => {
    expect(parseTime('not-a-date')).toBe('not-a-date');
    expect(parseTime('2026-13-99')).toBe('2026-13-99');
  });

  it('timestamp 0 视为有效时间', () => {
    expect(parseTime(0)).toBe('1970-01-01 08:00:00'); // UTC+8 本地时区
  });
});
