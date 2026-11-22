// 通用 API 响应类型
// 后端统一响应体：{ code, msg, ... }，列表带 rows/total，详情带 data

// 分页列表响应（对应后端 ListXxx 接口）
export interface PageResult<T> {
  code: number;
  msg: string;
  rows: T[];
  total: number;
}

// 单对象响应（对应后端 GetXxx / 增删改返回）
export interface DataResult<T> {
  code: number;
  msg: string;
  data: T;
}

// 通用响应（仅 code/msg，如删除、状态切换）
export interface PlainResult {
  code: number;
  msg: string;
}
