import { ref } from 'vue'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { ATR_MAX_PAGE_SIZE, computeAtrOverdue, normalizeAtrStatus, parseAtrPagination } from '~/utils/actionTakenReport'
import { emptyRiskBreakdown, mapWorkingPaperRiskLevel, type RiskBreakdown } from '~/utils/executiveSummaryCompilation'
import { extractErrorMessage } from '~/utils/error'

/** Safety cap for paging through one LHA's ATRs (ATR_MAX_PAGE_SIZE rows per page). */
const ATR_MAX_PAGES = 50

export type CompilationRiskBreakdown = RiskBreakdown

export interface CompilationFollowUp {
  closed: number
  inProgress: number
  overdue: number
}

export interface CompilationSources {
  /** Recommendations of the LHAs in scope (data.total of /fieldwork/test-controls/recommendations/summary). */
  recommendationCount: number
  /** Working paper risks grouped into the five risk levels. */
  riskBreakdown: CompilationRiskBreakdown
  /** Action taken reports of the LHAs in scope, grouped into the three follow-up statuses. */
  followUp: CompilationFollowUp
}

type RawRow = Record<string, unknown>

/** Only the fields used here. */
interface AtrRow {
  status?: unknown
  due_date?: unknown
}

interface ListPagePayload {
  items?: unknown
  pagination?: Partial<Record<'page' | 'page_size' | 'total' | 'total_pages', unknown>>
}

interface RecommendationSummaryPayload {
  total?: unknown
  items?: { auditResultReportId?: string, assignmentLetterId?: string, count?: number }[]
}

/**
 * ATR rows -> follow-up status counts.
 * Closed = COMPLETED; CANCELLED is excluded; every other status is open.
 * Overdue = open and due_date (date part) before today in Asia/Jakarta. The server's is_overdue flag is not used.
 * In Progress = open and not overdue.
 */
export const summarizeAtrFollowUp = (rows: AtrRow[], now: Date = new Date()): CompilationFollowUp => {
  const followUp: CompilationFollowUp = { closed: 0, inProgress: 0, overdue: 0 }
  for (const row of rows) {
    const status = normalizeAtrStatus(row.status)
    if (status === 'CANCELLED') continue
    if (status === 'COMPLETED') {
      followUp.closed++
    } else if (computeAtrOverdue(row.due_date, status, now).isOverdue) {
      followUp.overdue++
    } else {
      followUp.inProgress++
    }
  }
  return followUp
}

/** Rows from a list response, accepting `{ items }`, `{ data: { items } }`, `{ data: [] }` or a bare array. */
const extractItems = (res: unknown): RawRow[] => {
  if (Array.isArray(res)) return res as RawRow[]
  const body = res as { items?: unknown, data?: unknown } | null | undefined
  if (Array.isArray(body?.items)) return body.items as RawRow[]
  if (Array.isArray(body?.data)) return body.data as RawRow[]
  const payload = body?.data as { items?: unknown } | null | undefined
  if (Array.isArray(payload?.items)) return payload.items as RawRow[]
  return []
}

const fetchAllAtrForLha = async (baseUrl: string, lhaId: string): Promise<AtrRow[]> => {
  const collected: AtrRow[] = []
  let page = 1
  let totalPages = 1
  do {
    const response = await $fetch<{ data?: ListPagePayload }>(`${baseUrl}/action-taken-reports`, {
      method: 'GET',
      params: { audit_result_report_id: lhaId, page, page_size: ATR_MAX_PAGE_SIZE }
    })
    const rows = Array.isArray(response?.data?.items) ? (response.data.items as AtrRow[]) : []
    collected.push(...rows)
    totalPages = parseAtrPagination(response?.data?.pagination, { page, pageSize: ATR_MAX_PAGE_SIZE }).total_pages
    page++
  } while (page <= totalPages && page <= ATR_MAX_PAGES)
  return collected
}

const fetchRecommendationTotal = async (baseUrl: string, lhaIds: string[]): Promise<number> => {
  if (lhaIds.length === 0) return 0
  const response = await $fetch<{ data?: RecommendationSummaryPayload }>(
    `${baseUrl}/fieldwork/test-controls/recommendations/summary`,
    { method: 'GET', params: { audit_result_report_ids: lhaIds.join(',') } }
  )
  const total = Number(response?.data?.total)
  return Number.isFinite(total) && total > 0 ? total : 0
}

/**
 * Read-only data sources for the executive summary compilation statistics (audit-service):
 * - /fieldwork/test-controls/recommendations/summary?audit_result_report_ids= (recommendations, one call)
 * - /working-papers/risks                         (risk breakdown, matched by workingPaperId = assignment letter)
 * - /action-taken-reports?audit_result_report_id= (follow-up status)
 */
export const useExecutiveSummarySources = () => {
  const loading = ref(false)
  const error = ref('')
  let requestId = 0

  const fetchCompilationSources = async (input: {
    assignmentLetterIds: string[]
    lhaIds: string[]
  }): Promise<CompilationSources | null> => {
    const id = ++requestId
    loading.value = true
    error.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const letters = [...new Set(input.assignmentLetterIds.filter(Boolean))]
      const lhaIds = [...new Set(input.lhaIds.filter(Boolean))]

      const [recommendationCount, riskResponse, atrRowsPerLha] = await Promise.all([
        fetchRecommendationTotal(baseUrl, lhaIds),
        letters.length > 0
          ? $fetch<unknown>(`${baseUrl}/working-papers/risks`, { method: 'GET' })
          : Promise.resolve(null),
        Promise.all(lhaIds.map(lhaId => fetchAllAtrForLha(baseUrl, lhaId)))
      ])

      if (id !== requestId) return null

      const riskBreakdown = emptyRiskBreakdown()
      const letterSet = new Set(letters)
      for (const risk of extractItems(riskResponse)) {
        if (!letterSet.has(String(risk.workingPaperId || risk.assignmentLetterId || ''))) continue
        const level = mapWorkingPaperRiskLevel(risk.riskLevel)
        if (level) riskBreakdown[level]++
      }

      const followUp = summarizeAtrFollowUp(atrRowsPerLha.flat())

      return { recommendationCount, riskBreakdown, followUp }
    } catch (err) {
      if (id !== requestId) return null
      console.error('Failed to load executive summary compilation sources:', err)
      error.value = extractErrorMessage(err, 'Failed to load compilation sources.')
      return null
    } finally {
      if (id === requestId) loading.value = false
    }
  }

  /** Drops any in-flight request (its result is ignored) and clears loading/error. */
  const cancel = () => {
    requestId++
    loading.value = false
    error.value = ''
  }

  return { loading, error, fetchCompilationSources, cancel }
}
