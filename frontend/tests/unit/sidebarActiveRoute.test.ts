import { describe, it, expect } from 'vitest'

describe('Sidebar Active Route Logic', () => {
  // Simulate menu registration and isPathActive logic from default.vue
  const registeredPaths = [
    '/dashboard',
    '/audit-charter',
    '/risk-profile',
    '/risk-appetite',
    '/risk-profile/risk-factors',
    '/risk-profile/audit-universe',
    '/risk-profile/audit-universe?tab=priority',
    '/risk-profile/risk-control-matrix',
    '/strategic-audit-plan',
    '/strategic-audit-plan/upload',
    '/audit-result-report',
    '/audit-result-report/upload',
    '/executive-summary',
    '/executive-summary/upload',
    '/executive-summary-compilation',
    '/executive-summary-compilation/upload',
    '/action-taken-report',
    '/analytics?tab=xgboost',
    '/analytics?tab=isolation',
    '/settings',
    '/settings/mfa'
  ]

  const allMenuPaths = new Set(registeredPaths.map(p => (p.split('?')[0] || '').replace(/\/$/, '') || '/'))

  const createIsPathActive = (currentRoute: { path: string; fullPath: string; query: Record<string, string> }) => {
    return (targetPath?: string): boolean => {
      if (!targetPath) return false

      if (targetPath.includes('?')) {
        const [targetBase, targetQueryStr] = targetPath.split('?')
        const currentBase = currentRoute.path.replace(/\/$/, '') || '/'
        const normTargetBase = (targetBase || '').replace(/\/$/, '') || '/'

        if (currentBase !== normTargetBase) return false

        const targetParams = new URLSearchParams(targetQueryStr)
        for (const [key, value] of targetParams.entries()) {
          if (currentRoute.query[key] !== value) {
            return false
          }
        }
        return true
      }

      if (currentRoute.query.tab) {
        return false
      }

      const currentPath = currentRoute.path.replace(/\/$/, '') || '/'
      const targetBase = targetPath.replace(/\/$/, '') || '/'

      if (currentPath === targetBase) {
        return true
      }

      if (!allMenuPaths.has(currentPath) && currentPath.startsWith(targetBase + '/')) {
        return true
      }

      return false
    }
  }

  it('only highlights Import Executive Summary Individual when on /executive-summary/upload', () => {
    const isPathActive = createIsPathActive({
      path: '/executive-summary/upload',
      fullPath: '/executive-summary/upload',
      query: {}
    })

    expect(isPathActive('/executive-summary/upload')).toBe(true)
    expect(isPathActive('/executive-summary')).toBe(false)
    expect(isPathActive('/executive-summary-compilation')).toBe(false)
    expect(isPathActive('/executive-summary-compilation/upload')).toBe(false)
  })

  it('only highlights Executive Summary Individual when on /executive-summary', () => {
    const isPathActive = createIsPathActive({
      path: '/executive-summary',
      fullPath: '/executive-summary',
      query: {}
    })

    expect(isPathActive('/executive-summary')).toBe(true)
    expect(isPathActive('/executive-summary/upload')).toBe(false)
    expect(isPathActive('/executive-summary-compilation')).toBe(false)
    expect(isPathActive('/executive-summary-compilation/upload')).toBe(false)
  })

  it('only highlights Import Executive Summary Compilation when on /executive-summary-compilation/upload', () => {
    const isPathActive = createIsPathActive({
      path: '/executive-summary-compilation/upload',
      fullPath: '/executive-summary-compilation/upload',
      query: {}
    })

    expect(isPathActive('/executive-summary-compilation/upload')).toBe(true)
    expect(isPathActive('/executive-summary-compilation')).toBe(false)
    expect(isPathActive('/executive-summary')).toBe(false)
    expect(isPathActive('/executive-summary/upload')).toBe(false)
  })

  it('only highlights Executive Summary Compilation when on /executive-summary-compilation', () => {
    const isPathActive = createIsPathActive({
      path: '/executive-summary-compilation',
      fullPath: '/executive-summary-compilation',
      query: {}
    })

    expect(isPathActive('/executive-summary-compilation')).toBe(true)
    expect(isPathActive('/executive-summary-compilation/upload')).toBe(false)
    expect(isPathActive('/executive-summary')).toBe(false)
    expect(isPathActive('/executive-summary/upload')).toBe(false)
  })

  it('only highlights Import LHA Document when on /audit-result-report/upload', () => {
    const isPathActive = createIsPathActive({
      path: '/audit-result-report/upload',
      fullPath: '/audit-result-report/upload',
      query: {}
    })

    expect(isPathActive('/audit-result-report/upload')).toBe(true)
    expect(isPathActive('/audit-result-report')).toBe(false)
  })

  it('correctly isolates query tabs such as Audit Priority', () => {
    const isPathActivePriority = createIsPathActive({
      path: '/risk-profile/audit-universe',
      fullPath: '/risk-profile/audit-universe?tab=priority',
      query: { tab: 'priority' }
    })

    expect(isPathActivePriority('/risk-profile/audit-universe?tab=priority')).toBe(true)
    expect(isPathActivePriority('/risk-profile/audit-universe')).toBe(false)

    const isPathActiveUniverse = createIsPathActive({
      path: '/risk-profile/audit-universe',
      fullPath: '/risk-profile/audit-universe',
      query: {}
    })

    expect(isPathActiveUniverse('/risk-profile/audit-universe')).toBe(true)
    expect(isPathActiveUniverse('/risk-profile/audit-universe?tab=priority')).toBe(false)
  })

  it('allows prefix matching fallback only for unlisted detail routes', () => {
    const isPathActiveDetail = createIsPathActive({
      path: '/strategic-audit-plan/details/custom-item-123',
      fullPath: '/strategic-audit-plan/details/custom-item-123',
      query: {}
    })

    expect(isPathActiveDetail('/strategic-audit-plan')).toBe(true)
    expect(isPathActiveDetail('/strategic-audit-plan/upload')).toBe(false)
  })
})
