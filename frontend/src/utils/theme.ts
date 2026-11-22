// 主题工具：主题色色阶生成与 CSS 变量应用

// hex 转 rgb
export function hexToRgb(hex) {
  hex = hex.replace('#', '');
  if (hex.length === 3) {
    hex = hex
      .split('')
      .map((c) => c + c)
      .join('');
  }
  const num = parseInt(hex, 16);
  return [(num >> 16) & 255, (num >> 8) & 255, num & 255];
}

// rgb 转 hex
function rgbToHex(r, g, b) {
  const toHex = (n) => {
    const s = n.toString(16);
    return s.length === 1 ? '0' + s : s;
  };
  return '#' + toHex(r) + toHex(g) + toHex(b);
}

// 混色：c1 与 c2 按 weight（0~1，c2 占比）混合
export function mix(c1, c2, weight) {
  const [r1, g1, b1] = hexToRgb(c1);
  const [r2, g2, b2] = hexToRgb(c2);
  const r = Math.round(r1 * (1 - weight) + r2 * weight);
  const g = Math.round(g1 * (1 - weight) + g2 * weight);
  const b = Math.round(b1 * (1 - weight) + b2 * weight);
  return rgbToHex(r, g, b);
}

// 根据主色生成 Element Plus 主题色色阶（覆盖 --el-color-primary 系列）
export function generateThemeColors(primary) {
  return {
    '--el-color-primary': primary,
    '--el-color-primary-light-3': mix(primary, '#ffffff', 0.3),
    '--el-color-primary-light-5': mix(primary, '#ffffff', 0.5),
    '--el-color-primary-light-7': mix(primary, '#ffffff', 0.7),
    '--el-color-primary-light-8': mix(primary, '#ffffff', 0.8),
    '--el-color-primary-light-9': mix(primary, '#ffffff', 0.9),
    '--el-color-primary-dark-2': mix(primary, '#000000', 0.2)
  };
}

// 应用主题色到根元素
export function applyThemeColors(primary) {
  const colors = generateThemeColors(primary);
  const root = document.documentElement;
  Object.entries(colors).forEach(([key, value]) => {
    root.style.setProperty(key, value);
  });
}
