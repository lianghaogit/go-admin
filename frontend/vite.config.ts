import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import compression from 'vite-plugin-compression';

export default defineConfig({
  plugins: [
    vue(),
    // 构建产物预压缩：gzip + brotli，Nginx 可直接伺服（省去运行时压缩开销）
    compression({ algorithm: 'gzip', ext: '.gz', threshold: 10240 }),
    compression({ algorithm: 'brotliCompress', ext: '.br', threshold: 10240 })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      // 上传文件静态访问（头像/附件在 dev 环境走后端静态服务）
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    esbuild: {
      // 生产构建移除 console/debugger（仅生产生效，dev 保留便于调试）
      drop: ['console', 'debugger']
    },
    // element-plus/echarts 分包体积已知较大（已按依赖拆 chunk），阈值上调避免噪音告警
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus', '@element-plus/icons-vue'],
          vendor: ['vue', 'vue-router', 'pinia', 'axios'],
          echarts: ['echarts']
        }
      }
    }
  }
});
