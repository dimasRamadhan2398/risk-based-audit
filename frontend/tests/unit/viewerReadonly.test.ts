/**
 * VIEWER read-only policy (UI): role flags in useRbac, Account Settings tab filtering, blocked routes,
 * and the static RBAC grid. The backend is authoritative; these tests only pin what the UI shows.
 */
import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'node:fs'
import path from 'node:path'
import { useAuthStore } from '../../stores/auth'
import { useRbac } from '../../composables/useRbac'
import {
  getSettingsTabs,
  resolveSettingsTab,
  isViewerBlockedPath,
  isMasterDataPath,
  isViewerOnly
} from '../../utils/viewerPolicy'
import { matrixData, getPermissionsForUserRoles } from '../../utils/rbacMatrix'

const read = (rel: string) => fs.readFileSync(path.resolve(__dirname, '../../', rel), 'utf8')

const setRoles = (roles: string[]) => {
  const auth = useAuthStore()
  auth.user = { id: '1', username: 'u', email: 'u@test.com', fullName: 'U', roles } as never
}

describe('useRbac VIEWER flags', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it.each(['VIEWER', 'viewer', 'Viewer'])('treats %s as read-only (case-insensitive)', (role) => {
    setRoles([role])
    const r = useRbac()
    expect(r.isViewer.value).toBe(true)
    expect(r.canWrite.value).toBe(false)
    expect(r.canAccessFiles.value).toBe(false)
    expect(r.canExport.value).toBe(false)
    expect(r.canAccessMasterData.value).toBe(false)
  })

  it.each(['ADMIN', 'EXECUTIVE', 'AUDITOR', 'AUDITEE', 'DEPARTMENT_HEAD'])('keeps %s fully enabled', (role) => {
    setRoles([role])
    const r = useRbac()
    expect(r.isViewer.value).toBe(false)
    expect(r.canWrite.value).toBe(true)
    expect(r.canAccessFiles.value).toBe(true)
    expect(r.canExport.value).toBe(true)
    expect(r.canAccessMasterData.value).toBe(true)
  })

  it('a VIEWER no longer passes privileged checks (it used to alias to EXECUTIVE)', () => {
    setRoles(['VIEWER'])
    const r = useRbac()
    expect(r.canManageAnnualPlan.value).toBe(false)
    expect(r.canManageCharter.value).toBe(false)
    expect(r.canImportPlanDocs.value).toBe(false)
    expect(r.canReviewExecutiveSummary.value).toBe(false)
    expect(r.canReviewAtr.value).toBe(false)
    expect(r.canCancelAtr.value).toBe(false)
    expect(r.canAssignAtr.value).toBe(false)
  })

  it('an EXECUTIVE is not a viewer and keeps its privileges', () => {
    setRoles(['EXECUTIVE'])
    const r = useRbac()
    expect(r.hasRole('viewer')).toBe(false)
    expect(r.canManageAnnualPlan.value).toBe(true)
  })

  it('keeps the existing ATR list scope for VIEWER', () => {
    setRoles(['VIEWER'])
    expect(useRbac().canViewAllAtr.value).toBe(true)
  })

  it('a user holding VIEWER plus a business role is not viewer-only', () => {
    expect(isViewerOnly(['VIEWER', 'AUDITOR'])).toBe(false)
    expect(isViewerOnly(['VIEWER'])).toBe(true)
    expect(isViewerOnly([])).toBe(false)
  })
})

describe('Account Settings tab filtering', () => {
  it('VIEWER only gets profile, activity and faq', () => {
    expect(getSettingsTabs(true)).toEqual(['profile', 'activity', 'faq'])
  })

  it('other roles keep their tabs and the email tab is gone for everyone', () => {
    const tabs = getSettingsTabs(false)
    expect(tabs).toEqual(['profile', 'settings', 'mfa', 'activity', 'permissions', 'datasource', 'faq'])
    expect(tabs).not.toContain('email')
  })

  it('falls back to profile when a VIEWER opens a hidden tab by query', () => {
    for (const tab of ['mfa', 'settings', 'permissions', 'datasource', 'email', 'nope', undefined]) {
      expect(resolveSettingsTab(tab, true)).toBe('profile')
    }
    expect(resolveSettingsTab('activity', true)).toBe('activity')
    expect(resolveSettingsTab('faq', true)).toBe('faq')
  })

  it('does not change the tab for other roles, except the removed email tab', () => {
    expect(resolveSettingsTab('mfa', false)).toBe('mfa')
    expect(resolveSettingsTab('email', false)).toBe('profile')
    expect(resolveSettingsTab(['permissions'], false)).toBe('permissions')
  })

  it('settings page uses the filtered list for sidebar, mobile bar and has no email tab', () => {
    const src = read('pages/settings/index.vue')
    expect(src).toContain('getSettingsTabs(isViewer.value)')
    expect(src).toContain('allLinks.value.filter')
    expect(src).not.toContain('\'email\'')
    expect(src).not.toContain('SettingsEmail')
    expect(src).not.toContain('titles.email')
    expect(fs.existsSync(path.resolve(__dirname, '../../components/settings/email.vue'))).toBe(true)
  })
})

describe('blocked routes and menu', () => {
  it('blocks master data and import/upload screens for VIEWER', () => {
    for (const p of ['/master/company', '/master/department/', '/master/location', '/master/employee', '/master',
      '/annual-audit/upload', '/working-paper/upload', '/quality-assurance/import', '/quality-assurance/import-saiv',
      '/import-working-paper', '/audit-assignment/upload']) {
      expect(isViewerBlockedPath(p)).toBe(true)
    }
    expect(isMasterDataPath('/master/company')).toBe(true)
  })

  it('leaves business pages open', () => {
    for (const p of ['/dashboard', '/annual-audit', '/audit-charter', '/settings', '/action-taken-report',
      '/quality-assurance', '/risk-profile/audit-universe', '/masterpiece']) {
      expect(isViewerBlockedPath(p)).toBe(false)
    }
  })

  it('route middleware and menu use the shared policy', () => {
    expect(read('middleware/02-viewer-readonly.global.ts')).toContain('navigateTo(\'/dashboard\')')
    expect(read('layouts/default.vue')).toContain('isViewerBlockedPath(item.to)')
  })
})

describe('static RBAC grid, VIEWER column', () => {
  const find = (module: string, sub?: string) =>
    matrixData.find(m => m.module === module && (sub === undefined || m.submodule === sub))!

  it('is READ for business modules and NONE for import, master data and system settings', () => {
    expect(find('Dashboard').viewer).toBe('READ')
    expect(find('Annual Audit Plan', 'Create / Manage Plan').viewer).toBe('READ')
    expect(find('Annual Audit Plan', 'Import Plan Document').viewer).toBe('NONE')
    expect(find('Master Data', 'Company, Department, Location').viewer).toBe('NONE')
    expect(find('Master Data', 'Employee Management').viewer).toBe('NONE')
    expect(find('Settings & User Management').viewer).toBe('NONE')
    expect(find('Account Settings').viewer).toBe('READ')
    expect(find('Files & Exports').viewer).toBe('NONE')
  })

  it('never grants a VIEWER anything above READ', () => {
    for (const m of matrixData) expect(['READ', 'NONE']).toContain(m.viewer)
  })

  it('"your permissions" for VIEWER excludes master data, files and exports', () => {
    const perms = getPermissionsForUserRoles(['VIEWER'], 'en')
    expect(perms.length).toBeGreaterThan(0)
    expect(perms.every(p => p.accessLevel === 'READ')).toBe(true)
    expect(perms.some(p => p.module === 'Master Data')).toBe(false)
    expect(perms.some(p => p.module === 'Files & Exports')).toBe(false)
    expect(perms.some(p => p.module === 'Settings & User Management')).toBe(false)
  })

  it('documents that the backend is authoritative', () => {
    expect(read('utils/rbacMatrix.ts')).toContain('The backend is authoritative')
  })
})

describe('gated actions in templates', () => {
  it.each([
    ['components/audit-fieldwork/AuditFieldworkDocument.vue', 'canAccessFiles'],
    ['components/audit-fieldwork/AuditFieldworkInterviewModal.vue', 'canAccessFiles'],
    ['components/audit-charter/AuditCharterCard.vue', 'canAccessFiles'],
    ['components/action-taken-report/AtrEvidenceList.vue', 'canAccessFiles'],
    ['pages/kpi-performance/index.vue', 'canExport'],
    ['pages/audit-result-report/index.vue', 'canExport'],
    ['components/settings/permission.vue', 'canExport'],
    ['pages/executive-summary/index.vue', 'canWrite'],
    ['components/quality-assurance/QATable.vue', 'canWrite']
  ])('%s uses %s', (file, flag) => {
    expect(read(file)).toContain(`${flag}`)
    expect(read(file)).toMatch(new RegExp(`v-if="[^"]*${flag}`))
  })
})
