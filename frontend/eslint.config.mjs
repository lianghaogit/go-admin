// ESLint 扁平配置：TypeScript + Vue3 + Prettier
// 规则说明：
//  - 关闭与 Prettier 冲突的格式化规则（交给 prettier 统一处理）
//  - 关闭 no-explicit-any：项目 utils/request.ts 的 ApiClient 等场景需合理使用 any
//  - 关闭 vue/multi-word-component-names：视图文件按路由命名（如 index.vue / login.vue）
//  - caughtErrors 关闭：catch(e) 中的参数未使用时按项目惯例保留（RuoYi 风格）
import eslint from '@eslint/js'
import tseslint from 'typescript-eslint'
import pluginVue from 'eslint-plugin-vue'
import eslintConfigPrettier from 'eslint-config-prettier'
import globals from 'globals'

export default tseslint.config(
  {
    ignores: ['**/dist/**', '**/node_modules/**', '**/*.d.ts'],
  },
  {
    // 浏览器与 Node 全局变量（window/document/localStorage/FormData/Blob 等）
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  // ESLint 内置推荐规则
  eslint.configs.recommended,
  // TypeScript 推荐规则（覆盖 .ts / .vue 内 script）
  ...tseslint.configs.recommended,
  // Vue3 推荐规则（自动为 .vue 配置 vue-eslint-parser）
  ...pluginVue.configs['flat/recommended'],
  {
    // .vue 文件内 script 使用 TypeScript 解析
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
        sourceType: 'module',
        ecmaVersion: 'latest',
      },
    },
  },
  {
    // 项目级规则覆盖
    rules: {
      ...eslintConfigPrettier.rules,
      '@typescript-eslint/no-explicit-any': 'off',
      'vue/multi-word-component-names': 'off',
      'vue/no-v-html': 'off',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          argsIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          caughtErrors: 'none',
        },
      ],
    },
  }
)
