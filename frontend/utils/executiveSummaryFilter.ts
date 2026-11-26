import { isCompilationDocNumber } from '~/utils/executiveSummaryDocx'
import { normalizeReportNumber } from '~/stores/audit-result-report'

const LEGACY_EXCLUDED_DOCS = new Set([
  'DOC-EXSUM-Q1-2026',
  '020/LHA/01/KS IAD/2023',
  '019/LHA/01/KS IAD/2025',
  'LHA-020/SKAI/2023',
  'LHA-019/SKAI/2025'
])

export const isExcludedLegacyDoc = (doc?: string | null): boolean => {
  const trimmed = String(doc ?? '').trim()
  return !trimmed || LEGACY_EXCLUDED_DOCS.has(trimmed)
}

export const isIndividualSummaryDoc = (doc?: string | null): boolean =>
  !isExcludedLegacyDoc(doc) && !isCompilationDocNumber(doc)

export const isCompilationSummaryDoc = (doc?: string | null): boolean =>
  !isExcludedLegacyDoc(doc) && isCompilationDocNumber(doc)

/**
 * Filters Individual Executive Summaries out of a shared summary list so Compilation
 * documents (ESC-*, EXSUM-KOMP-*, KOMPILASI-*) never appear in Individual views.
 */
export const filterIndividualSummaries = <T extends { nomorDokumen?: string | null }>(
  list: readonly T[]
): T[] => list.filter(s => isIndividualSummaryDoc(s.nomorDokumen))

/**
 * Filters Compilation Executive Summaries out of a shared summary list so Individual
 * documents (LHA-*, etc.) never appear in Compilation views.
 */
export const filterCompilationSummaries = <T extends { nomorDokumen?: string | null }>(
  list: readonly T[]
): T[] => list.filter(s => isCompilationSummaryDoc(s.nomorDokumen))

/**
 * Filters LHA reports for the Individual Executive Summary "Select Assignment Letter / LHA"
 * dropdown, excluding LHAs that have already been selected into an Individual Executive Summary
 * (except the currently edited/viewed summary's own LHA so its selection stays visible).
 */
export const filterAvailableIndividualLhas = <
  R extends { reportNumber?: string | null, report_number?: string | null },
  S extends { nomorDokumen?: string | null }
>(
  reports: readonly R[],
  summaries: readonly S[],
  editingOwnDocNumber?: string | null
): R[] => {
  const ownNorm = editingOwnDocNumber ? normalizeReportNumber(editingOwnDocNumber) : ''
  const usedNorms = new Set<string>()
  for (const s of summaries) {
    if (!isIndividualSummaryDoc(s.nomorDokumen)) continue
    const norm = normalizeReportNumber(s.nomorDokumen)
    if (norm && norm !== ownNorm) {
      usedNorms.add(norm)
    }
  }

  const seenReportNorms = new Set<string>()
  return reports.filter((r) => {
    const num = r.reportNumber || r.report_number || ''
    if (!isIndividualSummaryDoc(num)) return false
    const norm = normalizeReportNumber(num)
    if (!norm || usedNorms.has(norm) || seenReportNorms.has(norm)) return false
    seenReportNorms.add(norm)
    return true
  })
}
