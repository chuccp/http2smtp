import { createI18n } from 'vue-i18n'
import en from './en'
import zhCN from './zh-cn'
import zhTW from './zh-tw'
import ja from './ja'

const messages = {
  en,
  'zh-cn': zhCN,
  'zh-tw': zhTW,
  ja
}

// Get browser language or default to zh-cn
function getDefaultLanguage(): string {
  const savedLang = localStorage.getItem('http2smtp-lang')
  if (savedLang && messages[savedLang]) {
    return savedLang
  }
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) {
    return 'zh-cn'
  }
  return 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLanguage(),
  fallbackLocale: 'en',
  messages
})

export default i18n
