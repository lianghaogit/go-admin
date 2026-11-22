// 表格列配置项（ColumnSetting 组件与各页面共享）
export interface ColumnItem {
  prop: string;
  label: string;
  visible?: boolean;
  fixed?: boolean;
  disabled?: boolean;
  width?: number | string;
  [key: string]: unknown;
}
