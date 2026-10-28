/**
 * Write rules (UI side): master data ADMIN only; QA and consulting writes for ADMIN/EXECUTIVE/CAE/AUDIT_MANAGER;
 * audit universe edits for ADMIN/EXECUTIVE/CAE, and DEPARTMENT_HEAD for its own department only.
 * Rights are the union of roles and VIEWER adds nothing. The backend is authoritative.
 */
import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import fs from 'node:fs'
import path from 'node:path'
import { useAuthStore } from '../../stores/auth'
import { useRbac } from '../../composables/useRbac'
import { isQaConsultingWritePath, isOwnDepartment } from '../../utils/viewerPolicy'
import { matrixData } from '../../utils/rbacMatrix'

const read = (rel: string) => fs.readFileSync(path.resolve(__dirname, '../../', rel), 'utf8')
const as = (roles: string[], extra: Record<string, unknown> = {}) => {
  useAuthStore().user = { id: '1', username: 'u', email: 'u@t.com', fullName: 'U', roles, ...extra } as never
  const r = useRbac()
  return {
    master: r.canWriteMasterData.value,
    qa: r.canWriteQaConsulting.value,
    universe: r.canWriteAuditUniverse.value,
    viewer: r.isViewer.value,
    scoped: r.canWriteAuditUniverseFor
  }
}

describe('write flags per role', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it.each([
    ['ADMIN', true, true, true],
    ['EXECUTIVE', false, true, true],
    ['CHIEF_AUDIT_EXECUTIVE', false, true, true],
    ['AUDIT_MANAGER', false, true, false],
    ['AUDITOR', false, false, false],
    ['AUDITEE', false, false, false],
    ['DEPARTMENT_HEAD', false, false, false],
    ['VIEWER', false, false, false],
    ['viewer', false, false, false],
    ['admin', true, true, true]
  ])('%s => master %s, qa/consulting %s, universe (all) %s', (role, master, qa, universe) => {
    const f = as([role])
    expect(f.master).toBe(master)
    expect(f.qa).toBe(qa)
    expect(f.universe).toBe(universe)
  })

  it('AUDITOR+VIEWER is treated as AUDITOR: not viewer-only, no write rights', () => {
    const f = as(['AUDITOR', 'VIEWER'])
    expect(f.viewer).toBe(false)
    expect(f.master).toBe(false)
    expect(f.qa).toBe(false)
    expect(f.universe).toBe(false)
    expect(f.scoped({ department_id: 'd1' })).toBe(false)
  })

  it('VIEWER contributes nothing to a union (ADMIN+VIEWER keeps ADMIN rights)', () => {
    const f = as(['VIEWER', 'ADMIN'])
    expect(f.viewer).toBe(false)
    expect(f.master && f.qa && f.universe).toBe(true)
  })
})

describe('scoped audit universe flag (DEPARTMENT_HEAD, own department only)', () => {
  beforeEach(() => setActivePinia(createPinia()))
  const dh = { departmentId: 'dep-1', department: 'Finance' }

  it('DEPARTMENT_HEAD edits rows of its own department by id or by name', () => {
    const f = as(['DEPARTMENT_HEAD'], dh)
    expect(f.scoped({ department_id: 'DEP-1' })).toBe(true)
    expect(f.scoped({ departmentId: 'dep-1' })).toBe(true)
    expect(f.scoped({ department: { id: 'dep-1' } })).toBe(true)
    expect(f.scoped({ department: '  finance ' })).toBe(true)
    expect(f.scoped('dep-1')).toBe(true)
  })

  it('DEPARTMENT_HEAD cannot edit another department or a row with no department', () => {
    const f = as(['DEPARTMENT_HEAD'], dh)
    expect(f.scoped({ department_id: 'dep-2', department: 'HR' })).toBe(false)
    expect(f.scoped({})).toBe(false)
    expect(f.scoped(null)).toBe(false)
    expect(f.scoped(undefined)).toBe(false)
    expect(f.universe).toBe(false)
  })

  it('DEPARTMENT_HEAD without a department of its own matches nothing', () => {
    const f = as(['DEPARTMENT_HEAD'])
    expect(f.scoped({ department_id: 'dep-1', department: 'Finance' })).toBe(false)
    expect(f.scoped({})).toBe(false)
  })

  it('full writers pass for any row, even without department data', () => {
    for (const role of ['ADMIN', 'EXECUTIVE', 'CHIEF_AUDIT_EXECUTIVE']) {
      expect(as([role]).scoped({})).toBe(true)
      expect(as([role]).scoped({ department_id: 'other' })).toBe(true)
    }
  })

  it('AUDITOR, AUDITEE and VIEWER never pass, even for their own department', () => {
    for (const role of ['AUDITOR', 'AUDITEE', 'VIEWER']) {
      expect(as([role], dh).scoped({ department_id: 'dep-1', department: 'Finance' })).toBe(false)
    }
  })

  it('isOwnDepartment ignores blanks on both sides', () => {
    expect(isOwnDepartment({ departmentId: ' ', department: '' }, { department: '' })).toBe(false)
    expect(isOwnDepartment(null, 'x')).toBe(false)
  })

  it('the audit universe page wires row controls to the scoped flag and keeps sync global', () => {
    const src = read('pages/risk-profile/audit-universe/index.vue')
    expect(src).toContain('canEditNode(node)')
    expect(src).toContain('canEditNode(sub)')
    // sync-from-master keeps the global flag (hidden for DEPARTMENT_HEAD)
    expect(src).toMatch(/<UButton[^>]*v-if="canEditAuditUniverse"[^>]*@click="handleSyncMasterData"/)
  })
})

describe('write-rule paths and templates', () => {
  it('QA import and consulting upload are write paths', () => {
    expect(isQaConsultingWritePath('/consulting-service/upload')).toBe(true)
    expect(isQaConsultingWritePath('/quality-assurance/import-saiv')).toBe(true)
    expect(isQaConsultingWritePath('/quality-assurance')).toBe(false)
    expect(isQaConsultingWritePath('/consulting-service')).toBe(false)
  })

  it.each([
    ['components/master/CompanyTable.vue', 'canWriteMasterData'],
    ['components/master/DepartmentTable.vue', 'canWriteMasterData'],
    ['components/master/LocationTable.vue', 'canWriteMasterData'],
    ['components/master/EmployeeTable.vue', 'canWriteMasterData'],
    ['components/strategic-audit-plan/VisionMissionCard.vue', 'canWriteMasterData'],
    ['pages/quality-assurance/index.vue', 'canWriteQaConsulting'],
    ['components/quality-assurance/QATable.vue', 'canWriteQaConsulting'],
    ['pages/consulting-service/index.vue', 'canWriteQaConsulting'],
    ['pages/risk-profile/audit-universe/index.vue', 'canEditAuditUniverse']
  ])('%s is gated by %s', (file, flag) => {
    expect(read(file)).toMatch(new RegExp(`v-if="[^"]*${flag}`))
  })
})

describe('grid matches the write rules', () => {
  const row = (id: string) => matrixData.find(m => m.id === id)!
  it('master data: only ADMIN writes, nobody else above READ, VIEWER none', () => {
    for (const id of ['16-1', '16-2']) {
      const m = row(id)
      expect(m.admin).toBe('FULL')
      for (const c of [m.executive, m.auditor, m.auditee, m.departmentHead]) expect(['READ', 'NONE']).toContain(c)
      expect(m.viewer).toBe('NONE')
    }
  })
  it('consulting and QA: only ADMIN and EXECUTIVE write', () => {
    for (const id of ['12', '13']) {
      const m = row(id)
      expect([m.admin, m.executive]).toEqual(['FULL', 'FULL'])
      for (const c of [m.auditor, m.auditee, m.departmentHead, m.viewer]) expect(['READ', 'NONE']).toContain(c)
    }
  })
  it('audit universe: ADMIN and EXECUTIVE full, DEPARTMENT_HEAD own department only, others at most READ', () => {
    const m = row('3-4')
    expect([m.admin, m.executive]).toEqual(['FULL', 'FULL'])
    expect(m.departmentHead).toBe('EDIT — Own Dept.')
    for (const c of [m.auditor, m.auditee, m.viewer]) expect(['READ', 'NONE']).toContain(c)
  })
  it('keeps the backend-is-authoritative note', () => {
    expect(read('utils/rbacMatrix.ts')).toContain('The backend is authoritative')
  })
})
