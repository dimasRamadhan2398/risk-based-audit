import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useQualityAssuranceStore, QAType, QAStatus } from '~/stores/quality-assurance'
import * as companyApiModule from '~/composables/useCompanyApi'
import type { Company } from '~/types/master'
import type { QAReport } from '~/types/quality-assurance'

vi.mock('#app', () => ({
  useRuntimeConfig: () => ({
    public: {
      auditServiceBaseUrl: 'http://localhost:8080/api/v1',
      masterServiceBaseUrl: 'http://localhost:8080/api/v1'
    }
  })
}))

vi.mock('~/components/shared/ToastNotification.vue', () => ({
  useToastNotification: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn(),
    success: vi.fn(),
    error: vi.fn()
  })
}))

;(globalThis as any).$fetch = vi.fn().mockResolvedValue([])

const ADD_MODAL_SFC = resolve(__dirname, '../../components/quality-assurance/QAAddModal.vue')
const TABLE_SFC = resolve(__dirname, '../../components/quality-assurance/QATable.vue')
const DETAIL_MODAL_SFC = resolve(__dirname, '../../components/quality-assurance/QADetailModal.vue')

const mockCompanies: Company[] = [
  {
    id: 'comp-1',
    company_code: 'AIFL-HQ',
    company_name: 'PT AIFL Indonesia',
    company_type: 'HOLDING',
    is_active: true,
    created_at: '2025-01-01',
    updated_at: '2025-01-01'
  },
  {
    id: 'comp-2',
    company_code: 'AIFL-FIN',
    company_name: 'AIFL Financial Services',
    company_type: 'SUBSIDIARY',
    is_active: true,
    created_at: '2025-02-01',
    updated_at: '2025-02-01'
  },
  {
    id: 'comp-3',
    company_code: 'AIFL-SBY',
    company_name: 'AIFL Surabaya',
    company_type: 'BRANCH',
    is_active: true,
    created_at: '2025-03-01',
    updated_at: '2025-03-01'
  },
  {
    id: 'comp-inactive',
    company_code: 'OLD-CO',
    company_name: 'Legacy Holding',
    company_type: 'SUBSIDIARY',
    is_active: false,
    created_at: '2024-01-01',
    updated_at: '2024-01-01'
  }
]

describe('Quality Assurance Review - Conducted By Synchronized with Master Company Management', () => {
  let store: ReturnType<typeof useQualityAssuranceStore>
  let getAllCompaniesMock: Mock

  beforeEach(() => {
    setActivePinia(createPinia())
    getAllCompaniesMock = vi.fn().mockResolvedValue(mockCompanies)
    vi.spyOn(companyApiModule, 'useCompanyApi').mockReturnValue({
      getAllCompanies: getAllCompaniesMock,
      getCompanies: vi.fn(),
      getCompanyById: vi.fn(),
      createCompany: vi.fn(),
      updateCompany: vi.fn(),
      deleteCompany: vi.fn()
    })
    store = useQualityAssuranceStore()
  })

  it('fetches companies from Company Management API and filters out inactive companies', async () => {
    await store.fetchCompanies()

    expect(getAllCompaniesMock).toHaveBeenCalled()
    expect(store.companies).toHaveLength(3)
    expect(store.companies.some(c => c.id === 'comp-inactive')).toBe(false)
    expect(store.companies[0]?.company_name).toBe('AIFL Financial Services')
    expect(store.companies[1]?.company_name).toBe('AIFL Surabaya')
    expect(store.companies[2]?.company_name).toBe('PT AIFL Indonesia')
  })

  it('generates companyOptions with company_name and company_code sorted alphabetically', async () => {
    await store.fetchCompanies()

    expect(store.companyOptions).toEqual([
      'AIFL Financial Services (AIFL-FIN)',
      'AIFL Surabaya (AIFL-SBY)',
      'PT AIFL Indonesia (AIFL-HQ)'
    ])
    expect(store.conductedByOptions).toEqual(store.companyOptions)
  })

  it('resolves company option from ID, code, name or formatted string', async () => {
    await store.fetchCompanies()

    expect(store.resolveCompanyOption('comp-1')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.resolveCompanyOption('AIFL-FIN')).toBe('AIFL Financial Services (AIFL-FIN)')
    expect(store.resolveCompanyOption('PT AIFL Indonesia')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.resolveCompanyOption('PT AIFL Indonesia (AIFL-HQ)')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.resolveCompanyOption('External Consultant X')).toBe('External Consultant X')
    expect(store.resolveCompanyOption('')).toBe('')
    expect(store.resolveCompanyOption(undefined)).toBe('')
  })

  it('formats company name for table and detail display correctly', async () => {
    await store.fetchCompanies()

    expect(store.formatCompanyName('comp-1')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.formatCompanyName('AIFL-FIN')).toBe('AIFL Financial Services (AIFL-FIN)')
    expect(store.formatCompanyName('PT AIFL Indonesia')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.formatCompanyName('PT AIFL Indonesia (AIFL-HQ)')).toBe('PT AIFL Indonesia (AIFL-HQ)')
    expect(store.formatCompanyName('External Consultant X')).toBe('External Consultant X')
    expect(store.formatCompanyName('')).toBe('-')
    expect(store.formatCompanyName(undefined)).toBe('-')
  })

  it('preserves existing conductedBy and resolves it when opening edit form', async () => {
    await store.fetchCompanies()

    const mockReport: QAReport = {
      id: 'qa-1',
      type: QAType.QAR,
      period: '2026',
      reportName: 'External QAR 2026',
      result: 'Generally Conformed',
      status: QAStatus.COMPLETED,
      conductedBy: 'PT AIFL Indonesia',
      assessmentTitle: 'QAR Assessment 2026'
    }

    store.editReport(mockReport)

    expect(store.isEditing).toBe(true)
    expect(store.isFormOpen).toBe(true)
    expect(store.newReport.conductedBy).toBe('PT AIFL Indonesia (AIFL-HQ)')
  })

  it('preserves custom/unmatched conductedBy when editing and includes it in companyOptions', async () => {
    await store.fetchCompanies()

    const mockReport: QAReport = {
      id: 'qa-custom',
      type: QAType.SAIV,
      period: 'Cycle 2025',
      reportName: 'SAIV 2025',
      result: 'Generally Conformed',
      status: QAStatus.COMPLETED,
      conductedBy: 'BPKP / Kementerian BUMN',
      assessmentTitle: 'SAIV Cycle 2025'
    }

    store.editReport(mockReport)

    expect(store.isEditing).toBe(true)
    expect(store.newReport.conductedBy).toBe('BPKP / Kementerian BUMN')
    expect(store.companyOptions).toContain('BPKP / Kementerian BUMN')
  })

  it('resets conductedBy on openForm in create mode and triggers company fetch', async () => {
    await store.fetchCompanies()

    store.newReport.conductedBy = 'PT AIFL Indonesia (AIFL-HQ)'
    store.openForm()

    expect(store.isEditing).toBe(false)
    expect(store.isFormOpen).toBe(true)
    expect(store.newReport.conductedBy).toBe('')
  })

  it('handles API errors gracefully when fetching companies', async () => {
    store.companiesLoaded = false
    getAllCompaniesMock.mockRejectedValueOnce(new Error('Network error'))

    await store.fetchCompanies(true)

    expect(store.companiesLoaded).toBe(false)
    expect(store.companiesLoading).toBe(false)
    expect(store.companiesError).toBe('Network error')
  })

  it('does not re-fetch companies if already loaded unless force=true', async () => {
    await Promise.all([
      store.fetchCompanies(),
      store.fetchCompanies(),
      store.fetchCompanies()
    ])

    expect(getAllCompaniesMock).toHaveBeenCalledTimes(1)
    expect(store.companiesLoaded).toBe(true)
    expect(store.companiesLoading).toBe(false)

    await store.fetchCompanies(true)
    expect(getAllCompaniesMock).toHaveBeenCalledTimes(2)
  })

  it('QAAddModal.vue binds conductedBy to store.companyOptions with USelectMenu, loading, create-item, and empty fallback', () => {
    const sfc = readFileSync(ADD_MODAL_SFC, 'utf-8')

    expect(sfc).toMatch(/<UFormField[^>]*?:label="t\('qualityAssurance\.modal\.conductedBy'\)"/)
    expect(sfc).toContain(':items="store.companyOptions"')
    expect(sfc).toContain('v-model="store.newReport.conductedBy"')
    expect(sfc).toContain(':loading="store.companiesLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('@create=')
    expect(sfc).toContain('masterData.company.empty')
  })

  it('QATable.vue displays formatted conductedBy using store.formatCompanyName and triggers fetchCompanies on mounted', () => {
    const sfc = readFileSync(TABLE_SFC, 'utf-8')

    expect(sfc).toContain('#conductedBy-cell')
    expect(sfc).toContain('store.formatCompanyName(row.original.conductedBy)')
    expect(sfc).toContain('store.fetchCompanies()')
  })

  it('QADetailModal.vue displays formatted conductedBy using store.formatCompanyName', () => {
    const sfc = readFileSync(DETAIL_MODAL_SFC, 'utf-8')

    expect(sfc).toContain('store.formatCompanyName(store.selectedReport?.conductedBy)')
  })
})
