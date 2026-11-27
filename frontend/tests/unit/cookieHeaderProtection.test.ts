import { describe, expect, it } from 'vitest'
import { isProxiedApiPath, purgeLegacyDashboardCookies } from '~/utils/cookieCleanup'

describe('purgeLegacyDashboardCookies', () => {
  it('expires legacy dashboard-sidebar and dashboard-panel cookies while keeping auth and app cookies', () => {
    const writes: string[] = []
    const fakeDoc = {
      get cookie() {
        return [
          'auth-token=jwt.token.here',
          'auth-user=%7B%22username%22%3A%22admin%22%7D',
          'locale=id',
          'dashboard-sidebar-v-3-0=%7B%22size%22%3A15%2C%22collapsed%22%3Afalse%7D',
          'dashboard-sidebar-main-sidebar-final=%7B%22size%22%3A21%2C%22collapsed%22%3Afalse%7D',
          'dashboard-panel-v-0-1=%7B%22size%22%3A79%7D',
          'frontend_lang=en_US',
        ].join('; ')
      },
      set cookie(val: string) {
        writes.push(val)
      },
    }

    const removed = purgeLegacyDashboardCookies(fakeDoc)

    expect(removed).toEqual([
      'dashboard-sidebar-v-3-0',
      'dashboard-sidebar-main-sidebar-final',
      'dashboard-panel-v-0-1',
    ])
    expect(writes).toHaveLength(3)
    expect(writes[0]).toContain('dashboard-sidebar-v-3-0=;')
    expect(writes[0]).toContain('Max-Age=0')
    expect(writes.some(w => w.startsWith('auth-token='))).toBe(false)
    expect(writes.some(w => w.startsWith('auth-user='))).toBe(false)
    expect(writes.some(w => w.startsWith('locale='))).toBe(false)
    expect(writes.some(w => w.startsWith('frontend_lang='))).toBe(false)
  })

  it('returns an empty array safely when document or cookies are absent', () => {
    expect(purgeLegacyDashboardCookies(undefined)).toEqual([])
    expect(purgeLegacyDashboardCookies({ cookie: '' })).toEqual([])
  })
})

describe('isProxiedApiPath', () => {
  it('matches all proxied backend API and upload prefixes', () => {
    expect(isProxiedApiPath('/api/v1/companies?page=1&page_size=10')).toBe(true)
    expect(isProxiedApiPath('/api/v1/risks')).toBe(true)
    expect(isProxiedApiPath('/api/analytics/report')).toBe(true)
    expect(isProxiedApiPath('/api/python-ai/predict/anomaly/batch')).toBe(true)
    expect(isProxiedApiPath('/uploads/charter.pdf')).toBe(true)
  })

  it('does not match frontend pages or Nuxt icon routes', () => {
    expect(isProxiedApiPath('/dashboard')).toBe(false)
    expect(isProxiedApiPath('/risk-profile')).toBe(false)
    expect(isProxiedApiPath('/_nuxt_icon/lucide.json')).toBe(false)
    expect(isProxiedApiPath('/api/v10/other')).toBe(false)
    expect(isProxiedApiPath(undefined)).toBe(false)
  })
})
