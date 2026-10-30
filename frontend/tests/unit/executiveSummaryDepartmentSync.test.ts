import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import * as departmentApiModule from '~/composables/useDepartmentApi'
import type { Department } from '~/types/master'
import type { ExecutiveSummary } from '~/stores/executive-summary'

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

;(globalThis as any).$fetch = vi.fn().mockResolvedValue({ data: [] })

const FORM_SFC = resolve(__dirname, '../../components/audit-result-report/ExecutiveSummaryIndividualForm.vue')
const PAGE_SFC = resolve(__dirname, '../../pages/executive-summary/index.vue')

const mockDepartments: Department[] = [
  {
    id: 'dept-1',
    department_code: 'FIN',
    department_name: 'Financial',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-01-01',
    updated_at: '2025-01-01'
  },
  {
    id: 'dept-2',
    department_code: 'OPS',
    department_name: 'Operasional',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-02-01',
    updated_at: '2025-02-01'
  },
  {
    id: 'dept-3',
    department_code: 'DEV',
    department_name: 'Product Development',
    level: 1,
    company_id: 'comp-1',
    is_active: true,
    created_at: '2025-03-01',
    updated_at: '2025-03-01'
  },
  {
    id: 'dept-inactive',
    department_code: 'OLD',
    department_name: 'Legacy Department',
    level: 1,
    company_id: 'comp-1',
    is_active: false,
    created_at: '2024-01-01',
    updated_at: '2024-01-01'
  }
]

describe('Executive Summary (Individual) - Unit Division Synchronized with Master Department Management', () => {
  let store: ReturnType<typeof useExecutiveSummaryStore>
  let getAllDepartmentsMock: Mock

  beforeEach(() => {
    setActivePinia(createPinia())
    getAllDepartmentsMock = vi.fn().mockResolvedValue(mockDepartments)
    vi.spyOn(departmentApiModule, 'useDepartmentApi').mockReturnValue({
      getAllDepartments: getAllDepartmentsMock,
      getDepartments: vi.fn(),
      getDepartmentById: vi.fn(),
      createDepartment: vi.fn(),
      updateDepartment: vi.fn(),
      deleteDepartment: vi.fn()
    })
    store = useExecutiveSummaryStore()
  })

  it('fetches departments from Department Management API and filters out inactive departments', async () => {
    await store.fetchDepartments()

    expect(getAllDepartmentsMock).toHaveBeenCalled()
    expect(store.departments).toHaveLength(3)
    expect(store.departments.some(d => d.id === 'dept-inactive')).toBe(false)
    expect(store.departments[0]?.department_name).toBe('Financial')
    expect(store.departments[1]?.department_name).toBe('Operasional')
    expect(store.departments[2]?.department_name).toBe('Product Development')
  })

  it('generates unitDivisionOptions with department_name and department_code sorted alphabetically', async () => {
    await store.fetchDepartments()

    expect(store.unitDivisionOptions).toEqual([
      'Financial (FIN)',
      'Operasional (OPS)',
      'Product Development (DEV)'
    ])
    expect(store.divisionOptions).toEqual(store.unitDivisionOptions)
  })

  it('resolves department option from ID, code, name or formatted string', async () => {
    await store.fetchDepartments()

    expect(store.resolveDepartmentOption('dept-1')).toBe('Financial (FIN)')
    expect(store.resolveDepartmentOption('OPS')).toBe('Operasional (OPS)')
    expect(store.resolveDepartmentOption('Product Development')).toBe('Product Development (DEV)')
    expect(store.resolveDepartmentOption('Financial (FIN)')).toBe('Financial (FIN)')
    expect(store.resolveDepartmentOption('Custom Division')).toBe('Custom Division')
    expect(store.resolveDepartmentOption('')).toBe('')
    expect(store.resolveDepartmentOption(undefined)).toBe('')
  })

  it('formats department name for display correctly', async () => {
    await store.fetchDepartments()

    expect(store.formatDepartmentName('dept-1')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('OPS')).toBe('Operasional (OPS)')
    expect(store.formatDepartmentName('Product Development')).toBe('Product Development (DEV)')
    expect(store.formatDepartmentName('Financial (FIN)')).toBe('Financial (FIN)')
    expect(store.formatDepartmentName('Custom Division')).toBe('Custom Division')
    expect(store.formatDepartmentName('')).toBe('-')
    expect(store.formatDepartmentName(undefined)).toBe('-')
  })

  it('preserves existing unitDivision values in topFindings and resolves them when opening edit modal', async () => {
    await store.fetchDepartments()

    const mockSummary: ExecutiveSummary = {
      id: 'es-1',
      nomorDokumen: 'LHA-001/SKAI/2026',
      dokumenPath: 'report.pdf',
      assignmentLetterId: 'ST-001/SKAI/2026',
      quarter: 1,
      periodeBulan: 'Januari',
      tahun: 2026,
      status: 'Draft',
      narrative: 'Executive Summary Narrative',
      jumlahLaporan: 1,
      risikoHigh: 1,
      risikoModerateHigh: 0,
      risikoModerate: 0,
      risikoLowModerate: 0,
      risikoLow: 0,
      jumlahRekomendasi: 1,
      followUpTable: [],
      topFindings: [
        {
          unitDivision: 'Operasional',
          judulTemuan: 'Keterlambatan rekonsiliasi',
          risiko: 'Tinggi',
          statusTL: 'In Progress',
          usulan: 'Revisi SOP'
        }
      ],
      matriksKompilasi: [],
      akarMasalah: '',
      kesimpulan: '',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-01-01',
      signatureNamaKepala: 'Head of SKAI',
      signatureNIK: '123'
    }

    store.openEditForm(mockSummary)

    expect(store.isEditing).toBe(true)
    expect(store.showModal).toBe(true)
    expect(store.form.topFindings[0]?.unitDivision).toBe('Operasional (OPS)')
  })

  it('preserves custom/unmatched unitDivision when editing and includes it in unitDivisionOptions', async () => {
    await store.fetchDepartments()

    const mockSummary: ExecutiveSummary = {
      id: 'es-custom',
      nomorDokumen: 'LHA-002/SKAI/2026',
      dokumenPath: 'custom.pdf',
      assignmentLetterId: 'ST-002/SKAI/2026',
      quarter: 1,
      periodeBulan: 'Januari',
      tahun: 2026,
      status: 'Draft',
      narrative: 'Custom Narrative',
      jumlahLaporan: 1,
      risikoHigh: 1,
      risikoModerateHigh: 0,
      risikoModerate: 0,
      risikoLowModerate: 0,
      risikoLow: 0,
      jumlahRekomendasi: 1,
      followUpTable: [],
      topFindings: [
        {
          unitDivision: 'Satgas Khusus Adhoc',
          judulTemuan: 'Audit Investigasi',
          risiko: 'Tinggi',
          statusTL: 'In Progress',
          usulan: 'Tindak lanjut'
        }
      ],
      matriksKompilasi: [],
      akarMasalah: '',
      kesimpulan: '',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-01-01',
      signatureNamaKepala: 'Head of SKAI',
      signatureNIK: '123'
    }

    store.openEditForm(mockSummary)

    expect(store.form.topFindings[0]?.unitDivision).toBe('Satgas Khusus Adhoc')
    expect(store.unitDivisionOptions).toContain('Satgas Khusus Adhoc')
  })

  it('resets topFindings on openNewForm in create mode and triggers department fetch', async () => {
    await store.fetchDepartments()
    store.form.topFindings = [
      {
        unitDivision: 'Operasional (OPS)',
        judulTemuan: 'Test',
        risiko: 'Tinggi',
        statusTL: 'In Progress',
        usulan: 'Usulan'
      }
    ]

    store.openNewForm(1)

    expect(store.isEditing).toBe(false)
    expect(store.form.topFindings).toEqual([])
  })

  it('deduplicates fetchDepartments calls to prevent redundant network requests', async () => {
    await Promise.all([
      store.fetchDepartments(),
      store.fetchDepartments(),
      store.fetchDepartments()
    ])

    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(1)
    expect(store.departmentsLoaded).toBe(true)
    expect(store.departmentsLoading).toBe(false)

    // Force fetch should re-query
    await store.fetchDepartments(true)
    expect(getAllDepartmentsMock).toHaveBeenCalledTimes(2)
  })

  it('handles API errors gracefully when fetching departments', async () => {
    getAllDepartmentsMock.mockRejectedValueOnce(new Error('Network error'))

    await store.fetchDepartments()

    expect(store.departmentsLoaded).toBe(false)
    expect(store.departmentsLoading).toBe(false)
    expect(store.departmentsError).toBe('Network error')
  })

  it('ExecutiveSummaryIndividualForm.vue binds unitDivision to store.unitDivisionOptions with loading, create-item and empty fallback', () => {
    const sfc = readFileSync(FORM_SFC, 'utf-8')

    // Table Header
    expect(sfc).toContain("t('executiveSummaryForm.table.unitDivision')")

    // Unit Division field binding
    expect(sfc).toContain(':items="store.unitDivisionOptions"')
    expect(sfc).toContain('v-model="finding.unitDivision"')
    expect(sfc).toContain(':loading="store.departmentsLoading"')
    expect(sfc).toContain('create-item')
    expect(sfc).toContain('@create=')
    expect(sfc).toContain('masterData.department.empty')

    // Formatted department display in context card
    expect(sfc).toContain('store.formatDepartmentName(linkedLhaReport.department)')
  })

  it('ExecutiveSummaryIndividualForm.vue and index.vue pre-fetch departments on load', () => {
    const formSfc = readFileSync(FORM_SFC, 'utf-8')
    const pageSfc = readFileSync(PAGE_SFC, 'utf-8')

    expect(formSfc).toContain('store.fetchDepartments')
    expect(pageSfc).toContain('store.fetchDepartments')
  })

  it('synchronizes Significance Level column header and options with Result Reports (LHA)', () => {
    const formSfc = readFileSync(FORM_SFC, 'utf-8')
    const idLocale = JSON.parse(readFileSync(resolve(__dirname, '../../locales/id/common.json'), 'utf-8'))
    const enLocale = JSON.parse(readFileSync(resolve(__dirname, '../../locales/en/common.json'), 'utf-8'))

    // Column header translations changed to Significance Level / Tingkat Signifikansi
    expect(enLocale.executiveSummaryForm.table.riskValue).toBe('Significance Level')
    expect(idLocale.executiveSummaryForm.table.riskValue).toBe('Tingkat Signifikansi')
    expect(enLocale.executiveSummaryForm.placeholders.risk).toBe('Significance Level')
    expect(idLocale.executiveSummaryForm.placeholders.risk).toBe('Tingkat Signifikansi')

    // Table header binds t('executiveSummaryForm.table.riskValue')
    expect(formSfc).toContain("t('executiveSummaryForm.table.riskValue')")

    // USelectMenu binds finding.risiko with significanceLevelOptions
    expect(formSfc).toContain('v-model="finding.risiko"')
    expect(formSfc).toContain(':items="significanceLevelOptions"')
    expect(formSfc).toContain('value-key="value"')

    // Form defines 4 significance level options matching Result Reports (LHA)
    expect(formSfc).toContain('auditResultReport.dashboard.categories.verySignificant')
    expect(formSfc).toContain('auditResultReport.dashboard.categories.significant')
    expect(formSfc).toContain('auditResultReport.dashboard.categories.quiteSignificant')
    expect(formSfc).toContain('auditResultReport.dashboard.categories.notSignificant')
  })

  it('normalizes legacy risk rating (Tinggi/Sedang/Rendah) to LHA Significance Level when editing', () => {
    const mockSummary: ExecutiveSummary = {
      id: 'es-legacy',
      nomorDokumen: 'LHA-003/SKAI/2026',
      dokumenPath: 'legacy.pdf',
      assignmentLetterId: 'ST-003/SKAI/2026',
      quarter: 1,
      periodeBulan: 'Januari',
      tahun: 2026,
      status: 'Draft',
      narrative: 'Legacy',
      jumlahLaporan: 1,
      risikoHigh: 1,
      risikoModerateHigh: 0,
      risikoModerate: 0,
      risikoLowModerate: 0,
      risikoLow: 0,
      jumlahRekomendasi: 1,
      followUpTable: [],
      topFindings: [
        {
          unitDivision: 'Finance',
          judulTemuan: 'Finding 1',
          risiko: 'Tinggi',
          statusTL: 'In Progress',
          usulan: 'Action 1'
        },
        {
          unitDivision: 'Operasional',
          judulTemuan: 'Finding 2',
          risiko: 'Sedang',
          statusTL: 'In Progress',
          usulan: 'Action 2'
        },
        {
          unitDivision: 'IT',
          judulTemuan: 'Finding 3',
          risiko: 'Rendah',
          statusTL: 'Closed',
          usulan: 'Action 3'
        }
      ],
      matriksKompilasi: [],
      akarMasalah: '',
      kesimpulan: '',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-01-01',
      signatureNamaKepala: 'Head of SKAI',
      signatureNIK: '123'
    }

    store.openEditForm(mockSummary)

    expect(store.form.topFindings[0]?.risiko).toBe('Significant')
    expect(store.form.topFindings[1]?.risiko).toBe('Quite Significant')
    expect(store.form.topFindings[2]?.risiko).toBe('Not Significant')
  })
})

