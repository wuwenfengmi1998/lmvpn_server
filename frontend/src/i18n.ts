import { createI18n } from 'vue-i18n'
import { messages, type Locale } from './locales'

const STORAGE_KEY = 'lang'

function detectLocale(): Locale {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved === 'zh' || saved === 'en') return saved
  const navLang = navigator.language?.toLowerCase() ?? ''
  return navLang.startsWith('zh') ? 'zh' : 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages,
})

export function setLocale(locale: Locale) {
  i18n.global.locale.value = locale
  localStorage.setItem(STORAGE_KEY, locale)
  document.documentElement.lang = locale === 'zh' ? 'zh-CN' : 'en'
}

export function toggleLocale() {
  setLocale(i18n.global.locale.value === 'zh' ? 'en' : 'zh')
}

document.documentElement.lang = i18n.global.locale.value === 'zh' ? 'zh-CN' : 'en'

export default i18n
