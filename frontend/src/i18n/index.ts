import { createI18n } from 'vue-i18n';
import zh from './locales/zh';
import en from './locales/en';

export const SUPPORT_LANGS = ['zh', 'en'] as const;
export type Lang = (typeof SUPPORT_LANGS)[number];

const LANG_KEY = 'go-admin-language';

export function getSavedLanguage(): Lang {
  const saved = localStorage.getItem(LANG_KEY);
  return saved === 'en' ? 'en' : 'zh';
}

export function saveLanguage(lang: Lang) {
  localStorage.setItem(LANG_KEY, lang);
}

const i18n = createI18n({
  legacy: false, // Composition API 模式
  globalInjection: true,
  locale: getSavedLanguage(),
  fallbackLocale: 'zh',
  messages: { zh, en }
});

export default i18n;
