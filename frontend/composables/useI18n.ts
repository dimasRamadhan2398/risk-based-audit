import en from '~/locales/en/common.json'
import id from '~/locales/id/common.json'

type DeepPartial<T> = {
  [P in keyof T]?: DeepPartial<T[P]>
}

type Translations = any

const translations: Record<string, Translations> = {
  en,
  id
}

export const useI18n = () => {
  const locale = useCookie('locale', { default: () => 'en' })

  const setLocale = (newLocale: string) => {
    if (translations[newLocale]) {
      locale.value = newLocale
    }
  }

  const t = (key: string, params?: Record<string, string | number>): string => {
    const keys = key.split('.')
    let value: any = translations[locale.value] || translations.en

    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k]
      } else {
        return key
      }
    }

    if (typeof value !== 'string') {
      return key
    }

    // Replace parameters in translation
    if (params) {
      return value.replace(/\{(\w+)\}/g, (match, paramKey) => {
        return params[paramKey]?.toString() || match
      })
    }

    return value
  }

  return {
    locale: computed(() => locale.value),
    setLocale,
    t
  }
}
