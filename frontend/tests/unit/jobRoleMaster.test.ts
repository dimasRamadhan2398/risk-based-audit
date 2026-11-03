// @ts-nocheck
/**
 * Job role master page: helper utils, API composable, store behaviour (access-role
 * suggestion, partial update payload, master-options refresh) and wiring (nav, RBAC, i18n parity).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'
import * as vue from 'vue'
import { setActivePinia, createPinia, defineStore } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import {
  JOB_POSITION_TYPES,
  FALLBACK_ACCESS_ROLES,
  suggestAccessRole,
  normalizeAccessRoles,
  labelOrRaw
} from '~/utils/jobRole'
import { useJobRoleApi } from '~/composables/useJobRoleApi'

const root = resolve(__dirname, '../..')
const read = (p: string) => readFileSync(resolve(root, p), 'utf8')

const fetchAll = vi.fn().mockResolvedValue(undefined)
vi.mock('~/stores/master-options', () => ({ useMasterOptionsStore: () => ({ fetchAll }) }))
const confirmDelete = vi.fn()
vi.mock('~/stores/global-modal', () => ({ useGlobalModalStore: () => ({ confirmDelete }) }))
vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({ showSuccess: vi.fn(), showError: vi.fn() })
}))

describe('jobRole utils', () => {
  it('covers every backend position type', () => {
    for (const t of ['COMMISSIONER', 'AUDIT_COMMITTEE', 'CHIEF_AUDIT_EXECUTIVE', 'EXTERNAL_AUDITOR', 'STAFF', 'DIRECTOR']) {
      expect(JOB_POSITION_TYPES).toContain(t)
    }
  })

  it('suggests an access role from the position type', () => {
    for (const t of ['DIRECTOR', 'GENERAL_MANAGER', 'SENIOR_MANAGER', 'COMMISSIONER', 'CHIEF_AUDIT_EXECUTIVE']) {
      expect(suggestAccessRole(t)).toBe('EXECUTIVE')
    }
    expect(suggestAccessRole('STAFF')).toBe('AUDITEE')
    expect(suggestAccessRole('EXTERNAL_AUDITOR')).toBe('')
    expect(suggestAccessRole('')).toBe('')
    expect(suggestAccessRole('UNKNOWN')).toBe('')
  })

  it('normalises the access-roles helper response shapes', () => {
    expect(normalizeAccessRoles({ data: ['A', 'B', 'A'] })).toEqual(['A', 'B'])
    expect(normalizeAccessRoles({ data: { access_roles: [{ name: 'X' }, ' Y '] } })).toEqual(['X', 'Y'])
    expect(normalizeAccessRoles({ roles: ['Z'] })).toEqual(['Z'])
    expect(normalizeAccessRoles(null)).toEqual([])
    expect(normalizeAccessRoles({ data: 'nope' })).toEqual([])
  })

  it('labelOrRaw falls back to the raw value when there is no translation', () => {
    expect(labelOrRaw(k => k, 'a.b', 'RAW')).toBe('RAW')
    expect(labelOrRaw(() => 'Label', 'a.b', 'RAW')).toBe('Label')
  })
})

describe('useJobRoleApi writes', () => {
  beforeEach(() => {
    global.$fetch = vi.fn().mockResolvedValue({ data: { id: '1' } })
  })

  it('creates, updates and deletes against /job-roles without hand-rolled auth headers', async () => {
    const api = useJobRoleApi()
    await api.createJobRole({ job_role_code: 'C', job_role_name: 'N', job_position_type: 'STAFF', is_active: true })
    await api.updateJobRole('1', { default_access_role: '' })
    await api.deleteJobRole('1')
    const calls = global.$fetch.mock.calls
    expect(calls[0][0]).toMatch(/\/job-roles$/)
    expect(calls[0][1].method).toBe('POST')
    expect(calls[1][0]).toMatch(/\/job-roles\/1$/)
    expect(calls[1][1]).toMatchObject({ method: 'PUT', body: { default_access_role: '' } })
    expect(calls[2][1].method).toBe('DELETE')
    for (const c of calls) expect(c[1].headers).toBeUndefined()
  })

  it('reads access roles from the helper and falls back to the constant list on failure or empty', async () => {
    global.$fetch = vi.fn().mockResolvedValue({ data: ['EXECUTIVE', 'AUDITOR'] })
    expect(await useJobRoleApi().getAccessRoles()).toEqual(['EXECUTIVE', 'AUDITOR'])
    expect(global.$fetch.mock.calls[0][0]).toMatch(/\/job-roles\/access-roles$/)

    global.$fetch = vi.fn().mockRejectedValue(new Error('404'))
    expect(await useJobRoleApi().getAccessRoles()).toEqual(FALLBACK_ACCESS_ROLES)

    global.$fetch = vi.fn().mockResolvedValue({ data: [] })
    expect(await useJobRoleApi().getAccessRoles()).toEqual(FALLBACK_ACCESS_ROLES)
  })
})

describe('job-role store', () => {
  let store: any

  beforeEach(async () => {
    Object.assign(global, { ref: vue.ref, reactive: vue.reactive, computed: vue.computed, defineStore, useJobRoleApi })
    setActivePinia(createPinia())
    vi.resetModules()
    fetchAll.mockClear()
    confirmDelete.mockReset()
    global.$fetch = vi.fn().mockResolvedValue({ data: [] })
    const { useJobRoleStore } = await import('~/stores/job-role')
    store = useJobRoleStore()
  })

  it('suggests an access role until the user picks one', () => {
    store.openCreateModal()
    store.onPositionTypeChange('DIRECTOR')
    expect(store.form.default_access_role).toBe('EXECUTIVE')
    store.onPositionTypeChange('STAFF')
    expect(store.form.default_access_role).toBe('AUDITEE')

    store.onAccessRoleChange('AUDITOR')
    store.onPositionTypeChange('DIRECTOR')
    expect(store.form.default_access_role).toBe('AUDITOR')
  })

  it('does not overwrite the stored access role of an existing job role', () => {
    store.handleEdit({ id: '1', job_role_code: 'C', job_role_name: 'N', job_position_type: 'STAFF', default_access_role: '', is_active: true })
    store.onPositionTypeChange('DIRECTOR')
    expect(store.form.default_access_role).toBe('')
  })

  it('validates required fields before calling the API', async () => {
    store.openCreateModal()
    await store.handleSubmit()
    expect(store.errorMsg).toBeTruthy()
    expect(global.$fetch).not.toHaveBeenCalled()
  })

  it('sends the unchanged code on update (access role cleared with empty string) and refreshes options', async () => {
    store.handleEdit({ id: '9', job_role_code: 'C', job_role_name: 'Old', job_position_type: 'STAFF', default_access_role: 'AUDITEE', is_active: true })
    store.form.job_role_name = 'New'
    store.onAccessRoleChange(undefined)
    await store.handleSubmit()

    const put = global.$fetch.mock.calls.find(c => c[1]?.method === 'PUT')
    expect(put[0]).toMatch(/\/job-roles\/9$/)
    expect(put[1].body).toEqual({
      job_role_code: 'C',
      job_role_name: 'New',
      job_role_description: '',
      job_position_type: 'STAFF',
      default_access_role: '',
      is_active: true
    })
    expect(fetchAll).toHaveBeenCalledWith(true)
    expect(store.showModal).toBe(false)
  })

  it('refreshes master options after create and delete, not after a cancelled delete', async () => {
    store.openCreateModal()
    Object.assign(store.form, { job_role_code: 'C', job_role_name: 'N' })
    store.onPositionTypeChange('MANAGER')
    await store.handleSubmit()
    const post = global.$fetch.mock.calls.find(c => c[1]?.method === 'POST')
    expect(post[1].body).toMatchObject({ job_role_code: 'C', default_access_role: 'AUDITEE' })
    expect(fetchAll).toHaveBeenCalledTimes(1)

    confirmDelete.mockResolvedValueOnce(false)
    await store.handleDelete({ id: '1', job_role_name: 'N' })
    expect(fetchAll).toHaveBeenCalledTimes(1)

    confirmDelete.mockResolvedValueOnce(true)
    await store.handleDelete({ id: '1', job_role_name: 'N' })
    expect(global.$fetch.mock.calls.some(c => c[1]?.method === 'DELETE')).toBe(true)
    expect(fetchAll).toHaveBeenCalledTimes(2)
  })

  it('filters and pages client-side', () => {
    store.jobRoles = Array.from({ length: 25 }, (_, i) => ({
      id: String(i), job_role_code: `C${i}`, job_role_name: i === 3 ? 'Special' : `Role ${i}`, job_position_type: 'STAFF', is_active: true
    }))
    expect(store.paged).toHaveLength(10)
    store.setPage(3)
    expect(store.paged).toHaveLength(5)
    store.setSearch('special')
    expect(store.filtered).toHaveLength(1)
    expect(store.page).toBe(1)
  })
})

describe('wiring', () => {
  it('table gates writes behind canWriteMasterData', () => {
    const src = read('components/master/JobRoleTable.vue')
    expect(src).toContain('canWriteMasterData')
    expect(src.match(/v-if="canWriteMasterData"/g)).toHaveLength(3)
  })

  it('has a nav entry and a page rendering table and form', () => {
    expect(read('layouts/default.vue')).toContain('to: \'/master/job-role\'')
    const page = read('pages/master/job-role/index.vue')
    expect(page).toContain('<MasterJobRoleTable')
    expect(page).toContain('<MasterJobRoleForm')
  })

  it('en and id locales have the same jobRole keys', () => {
    const flat = (o: any, p = ''): string[] =>
      Object.entries(o).flatMap(([k, v]) => typeof v === 'object' && v ? flat(v, `${p}${k}.`) : [`${p}${k}`])
    const en = JSON.parse(read('locales/en/common.json'))
    const id = JSON.parse(read('locales/id/common.json'))
    expect(flat(en.masterData.jobRole).sort()).toEqual(flat(id.masterData.jobRole).sort())
    expect(en.navigation.jobRole).toBeTruthy()
    expect(id.navigation.jobRole).toBeTruthy()
    for (const t of JOB_POSITION_TYPES) expect(en.masterData.jobRole.positionTypes[t]).toBeTruthy()
    for (const k of ['fetchJobRoles', 'createJobRole', 'updateJobRole', 'deleteJobRole']) {
      expect(en.masterData.errors[k]).toBeTruthy()
      expect(id.masterData.errors[k]).toBeTruthy()
    }
  })
})
