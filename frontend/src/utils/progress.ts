// 全局路由进度条（对标 RuoYi nprogress，轻量自实现）
let barEl = null;
let timer = null;
let done = false;

// 创建并返回进度条实例
function ensureBar() {
  if (barEl) return barEl;
  barEl = document.createElement('div');
  barEl.id = 'go-progress-bar';
  document.body.appendChild(barEl);
  return barEl;
}

// 开始：宽度动画到 90%
export function startProgress() {
  const bar = ensureBar();
  done = false;
  bar.classList.add('go-progress-active');
  bar.style.width = '0%';
  // 强制 reflow 后过渡
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      if (!done) bar.style.width = '90%';
    });
  });
}

// 完成：到 100% 后淡出
export function finishProgress() {
  const bar = ensureBar();
  done = true;
  bar.style.width = '100%';
  clearTimeout(timer);
  timer = setTimeout(() => {
    bar.classList.remove('go-progress-active');
    bar.style.width = '0%';
  }, 300);
}

// 路由切换失败也要收尾
export function failProgress() {
  finishProgress();
}
