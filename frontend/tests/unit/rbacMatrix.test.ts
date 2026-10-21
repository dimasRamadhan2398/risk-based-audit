import { describe, it, expect } from 'vitest'
import {
  matrixData,
  normalizeRoleToColumn,
  getBadgeClass,
  getPermissionsForUserRoles,
} from '../../utils/rbacMatrix'
import { UserRole } from '../../types/auth'

describe('RBAC Permission Matrix & Onboarding Alignment', () => {
  it('contains the complete official matrix matching Account Settings permission tab', () => {
    expect(matrixData.length).toBe(27)
    const dashboard = matrixData.find(m => m.module === 'Dashboard')
    expect(dashboard).toBeDefined()
    expect(dashboard?.admin).toBe('FULL')
    expect(dashboard?.auditee).toBe('READ')

    const atr = matrixData.find(m => m.module === 'Action Taken Report (ATR)')
    expect(atr).toBeDefined()
    expect(atr?.auditee).toBe('UPDATE CAPA')

    const fieldwork = matrixData.find(m => m.submodule.includes('Fieldwork — Interviews'))
    expect(fieldwork).toBeDefined()
    expect(fieldwork?.auditee).toBe('RESPOND')

    const survey = matrixData.find(m => m.submodule === 'Client Satisfaction Survey')
    expect(survey).toBeDefined()
    expect(survey?.auditee).toBe('FILL SURVEY')
  })

  it('normalizes roles correctly to matrix column keys', () => {
    expect(normalizeRoleToColumn('AUDITEE')).toBe('auditee')
    expect(normalizeRoleToColumn('auditee')).toBe('auditee')
    expect(normalizeRoleToColumn('ADMIN')).toBe('admin')
    expect(normalizeRoleToColumn('AUDITOR')).toBe('auditor')
    expect(normalizeRoleToColumn('EXECUTIVE')).toBe('executive')
    expect(normalizeRoleToColumn('DEPARTMENT_HEAD')).toBe('departmentHead')
    expect(normalizeRoleToColumn('VIEWER')).toBe('viewer')
  })

  it('derives correct active permissions for AUDITEE role in ID locale', () => {
    const perms = getPermissionsForUserRoles(['AUDITEE'], 'id')

    // Auditee has 14 features with active access
    expect(perms.length).toBe(14)

    // Check specific module permissions matching the Account Settings Permission matrix
    const dashboardPerm = perms.find(p => p.module === 'Dashboard')
    expect(dashboardPerm).toBeDefined()
    expect(dashboardPerm?.accessLevel).toBe('READ')
    expect(dashboardPerm?.badgeClass).toContain('bg-sky-600')

    const rcmPerm = perms.find(p => p.submodule === 'Risk Control Matrix (RCM)')
    expect(rcmPerm).toBeDefined()
    expect(rcmPerm?.accessLevel).toBe('READ')

    const riskProfilePerm = perms.find(p => p.submodule === 'Corporate Risk Profile')
    expect(riskProfilePerm).toBeDefined()
    expect(riskProfilePerm?.accessLevel).toBe('EDIT — Own Dept.')
    expect(riskProfilePerm?.badgeClass).toContain('bg-amber-500')

    const fieldworkPerm = perms.find(p => p.module === 'Audit Fieldwork')
    expect(fieldworkPerm).toBeDefined()
    expect(fieldworkPerm?.accessLevel).toBe('RESPOND')
    expect(fieldworkPerm?.badgeClass).toContain('bg-amber-500')

    const surveyPerm = perms.find(p => p.submodule === 'Client Satisfaction Survey')
    expect(surveyPerm).toBeDefined()
    expect(surveyPerm?.accessLevel).toBe('FILL SURVEY')
    expect(surveyPerm?.badgeClass).toContain('bg-amber-500')

    const atrPerm = perms.find(p => p.module === 'Action Taken Report (ATR)')
    expect(atrPerm).toBeDefined()
    expect(atrPerm?.accessLevel).toBe('UPDATE CAPA')
    expect(atrPerm?.badgeClass).toContain('bg-amber-500')

    const consultingPerm = perms.find(p => p.module === 'Consulting Service')
    expect(consultingPerm).toBeDefined()
    expect(consultingPerm?.accessLevel).toBe('REQUEST')
    expect(consultingPerm?.badgeClass).toContain('bg-amber-500')

    // Features with NONE must NOT be in the permissions list
    expect(perms.some(p => p.module === 'Settings & User Management')).toBe(false)
    expect(perms.some(p => p.submodule === 'Working Papers — F-01 to F-05')).toBe(false)
    expect(perms.some(p => p.module === 'Strategic Audit Plan')).toBe(false)
    expect(perms.some(p => p.module === 'KPI Performance')).toBe(false)
    expect(perms.some(p => p.module === 'Quality Assurance (QA)')).toBe(false)
  })

  it('provides descriptions in Indonesian and English according to locale', () => {
    const permsId = getPermissionsForUserRoles([UserRole.AUDITEE], 'id')
    const atrId = permsId.find(p => p.module === 'Action Taken Report (ATR)')
    expect(atrId?.description).toContain('rencana tindak lanjut')

    const permsEn = getPermissionsForUserRoles([UserRole.AUDITEE], 'en')
    const atrEn = permsEn.find(p => p.module === 'Action Taken Report (ATR)')
    expect(atrEn?.description).toContain('action plan')
  })

  it('supports object role formats like { name: "auditee" } and { role: "AUDITEE" }', () => {
    const permsObject = getPermissionsForUserRoles([{ role: 'AUDITEE' }])
    expect(permsObject.length).toBe(14)
    expect(permsObject.some(p => p.module === 'Action Taken Report (ATR)')).toBe(true)
  })

  it('returns empty array when user has no matching roles', () => {
    const emptyPerms = getPermissionsForUserRoles([])
    expect(emptyPerms).toEqual([])
  })

  it('getBadgeClass returns correct CSS classes for each access type', () => {
    expect(getBadgeClass('FULL')).toContain('bg-emerald-600')
    expect(getBadgeClass('READ')).toContain('bg-sky-600')
    expect(getBadgeClass('NONE')).toContain('bg-slate-500')
    expect(getBadgeClass('RESPOND')).toContain('bg-amber-500')
    expect(getBadgeClass('UPDATE CAPA')).toContain('bg-amber-500')
  })
})
