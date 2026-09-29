import { describe, it, expect } from 'vitest'
import fs from 'fs'
import path from 'path'

describe('Quality Assurance Table and Summary Cards Synchronization', () => {
  const tablePath = path.resolve(__dirname, '../../components/quality-assurance/QATable.vue')
  const summaryPath = path.resolve(__dirname, '../../components/quality-assurance/QASummaryCards.vue')
  const storePath = path.resolve(__dirname, '../../stores/quality-assurance.ts')
  const detailModalPath = path.resolve(__dirname, '../../components/quality-assurance/QADetailModal.vue')

  const tableContent = fs.readFileSync(tablePath, 'utf8')
  const summaryContent = fs.readFileSync(summaryPath, 'utf8')
  const storeContent = fs.readFileSync(storePath, 'utf8')
  const detailModalContent = fs.readFileSync(detailModalPath, 'utf8')

  it('QATable binds store.filteredReports so all reports are visible and paginated', () => {
    expect(tableContent).toContain(':data="store.filteredReports"')
  })

  it('QATable formats IACM result to match Summary Cards', () => {
    expect(tableContent).toContain('formatResult(row.original)')
    expect(tableContent).toContain('formatIacmResult')
    expect(tableContent).toContain("${trimmed} / 5")
  })

  it('QASummaryCards formats IACM capability level safely', () => {
    expect(summaryContent).toContain('formatIacmResult((store.summary.iacm as any).result)')
    expect(summaryContent).toContain("${trimmed} / 5")
  })

  it('QASummaryCards uses actual status for QAR matching table status', () => {
    expect(summaryContent).toContain('(store.summary.qar as any).status')
    expect(summaryContent).not.toContain("? 'Verified' : (store.summary.qar as any).status")
  })

  it('QASummaryCards supports fallback for SAIV validator', () => {
    expect(summaryContent).toContain('(store.summary.saiv as any).validator || (store.summary.saiv as any).conductedBy ||')
  })

  it('Store summary filters out imported reports so only table reports appear in summary cards', () => {
    expect(storeContent).toContain('reports.value.filter(r => !r.isImported)')
    expect(storeContent).toContain('getLatestForType')
    expect(storeContent).toContain('compareReportsDesc')
  })

  it('Store filteredReports sorts reports by period weight descending', () => {
    expect(storeContent).toContain('.sort(compareReportsDesc)')
    expect(storeContent).toContain('parsePeriodWeight')
  })

  it('QADetailModal displays consistent IACM formatting', () => {
    expect(detailModalContent).toContain('formatDetailResult')
    expect(detailModalContent).toContain('formatIacmResult')
  })
})
