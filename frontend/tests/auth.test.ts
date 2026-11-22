import { describe, it, expect, beforeEach } from 'vitest';
import { getToken, setToken, removeToken } from '@/utils/auth';

// localStorage 令牌存取单元测试（jsdom 环境）

describe('auth token', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('未登录时 token 为空', () => {
    expect(getToken()).toBeNull();
  });

  it('setToken 后 getToken 可取回', () => {
    setToken('jwt-token-123');
    expect(getToken()).toBe('jwt-token-123');
  });

  it('removeToken 清除令牌', () => {
    setToken('jwt-token-123');
    removeToken();
    expect(getToken()).toBeNull();
  });

  it('不同 token 互不干扰（覆盖写入）', () => {
    setToken('first');
    setToken('second');
    expect(getToken()).toBe('second');
  });
});
