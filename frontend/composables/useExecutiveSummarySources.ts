import { ref } from 'vue'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { ATR_MAX_PAGE_SIZE, computeAtrStats, parseAtrPagination } from '~/utils/actionTakenReport'
import { extractErrorMessage } from '~/utils/error'

/** Safety cap for paging through one LHA's ATRs (ATR_MAX_PAGE_SIZE rows per page). */
const ATR_MAX_PAGES = 50

export interface CompilationRiskBreakdown {
  high: number
  medium: number
  low: number
}

export interface CompilationFollowUp {
  closed: number
  inProgress: number
  overdue: number
}

export interface CompilationSources {
  /** Test controls (fieldwork) with a non-empty recommendation, across the assignment letters in scope. */
  recommendationCount: number
  /** Working paper risks grouped into the three executive-summary levels. */
  riskBreakdown: CompilationRiskBreakdown
  /** Action taken reports of the LHAs in scope, grouped into the three follow-up statuses. */
  followUp: CompilationFollowUp
}

/**
 * Working paper risk level -> executive summary level.
 * Working paper uses five levels (low, low-moderate, moderate, moderate-high, high);
 * the executive summary has three (Tinggi / Sedang / Rendah).
 * Returns null for unknown values so they are not counted.
 */
export const mapWorkingPaperRiskLevel = (raw: unknown): keyof CompilationRiskBreakdown | null => {
  const key = String(raw ?? '').trim().toLowerCase().replace(/[\s_]+/g, '-')
  if (key === 'high' || key === 'moderate-high') return 'high'
  if (key === 'moderate') return 'medium'
  if (key === 'low' || key === 'low-moderate') return 'low'
  return null
}

export const countTestControlRecommendations = (rows: Array<{ recommendation?: unknown }>): number =>
  rows.filter(row => String(row.recommendation ?? '').trim() !== '').length

const extractItems = (res: any): any[] => {
  if (res?.items && Array.isArray(res.items)) return res.items
  if (res?.data?.items && Array.isArray(res.data.items)) return res.data.items
  if (Array.isArray(res?.data)) return res.data
  if (Array.isArray(res)) return res
  return []
}

const fetchAllAtrForLha = async (baseUrl: string, lhaId: string): Promise<Record<string, unknown>[]> => {
  const collected: Record<string, unknown>[] = []
  let page = 1
  let totalPages = 1
  do {
    const response = await $fetch<any>(`${baseUrl}/action-taken-reports`, {
      method: 'GET',
      params: { audit_result_report_id: lhaId, page, page_size: ATR_MAX_PAGE_SIZE }
    })
    const rows = Array.isArray(response?.data?.items) ? response.data.items : []
    collected.push(...rows)
    totalPages = parseAtrPagination(response?.data?.pagination, { page, pageSize: ATR_MAX_PAGE_SIZE }).total_pages
    page++
  } while (page <= totalPages && page <= ATR_MAX_PAGES)
  return collected
}

/**
 * Read-only data sources for the executive summary compilation statistics.
 * Uses the existing audit-service endpoints:
 * - /fieldwork/test-controls?assignmentLetterId=  (recommendations)
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

      const [controlResponses, riskResponse, atrRows] = await Promise.all([
        Promise.all(letters.map(letter =>
          $fetch<any>(`${baseUrl}/fieldwork/test-controls`, { method: 'GET', params: { assignmentLetterId: letter } })
        )),
        letters.length > 0
          ? $fetch<any>(`${baseUrl}/working-papers/risks`, { method: 'GET' })
          : Promise.resolve(null),
        Promise.all(lhaIds.map(lhaId => fetchAllAtrForLha(baseUrl, lhaId)))
      ])

      if (id !== requestId) return null

      const recommendationCount = controlResponses.reduce(
        (sum, res) => sum + countTestControlRecommendations(extractItems(res)),
        0
      )

      const riskBreakdown: CompilationRiskBreakdown = { high: 0, medium: 0, low: 0 }
      const letterSet = new Set(letters)
      for (const risk of extractItems(riskResponse)) {
        if (!letterSet.has(risk?.workingPaperId || risk?.assignmentLetterId)) continue
        const level = mapWorkingPaperRiskLevel(risk.riskLevel)
        if (level) riskBreakdown[level]++
      }

      const stats = computeAtrStats(atrRows as { status?: unknown, is_overdue?: boolean }[])
      const followUp: CompilationFollowUp = {
        closed: stats.counts.completed,
        overdue: stats.counts.overdue,
        inProgress: stats.counts.open - stats.counts.overdue
      }

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

  return { loading, error, fetchCompilationSources }
}
