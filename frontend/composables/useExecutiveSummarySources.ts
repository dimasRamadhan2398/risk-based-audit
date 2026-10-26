import { ref } from 'vue'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { parseCompilationStats, type CompilationStats } from '~/utils/executiveSummaryCompilation'
import { extractErrorMessage } from '~/utils/error'

/**
 * Path of the statistics endpoint on the audit-service. It must sit under a prefix that Kong
 * routes: /executive-summaries is routed, the older /executive-summary/compilation-stats is not
 * (Kong answers it with 404 "no Route matched with those values").
 */
export const COMPILATION_STATS_PATH = '/executive-summaries/compilation-stats'

/** Error text for the form: the server's message plus the HTTP status, so a 404/403/500 is told apart. */
export const describeStatsError = (err: unknown): string => {
  const message = extractErrorMessage(err, 'Failed to load compilation statistics.')
  const e = err as { statusCode?: unknown, status?: unknown, response?: { status?: unknown } } | null
  const status = Number(e?.statusCode ?? e?.status ?? e?.response?.status)
  if (!Number.isInteger(status) || status <= 0 || message.includes(String(status))) return message
  return `${message} (HTTP ${status})`
}

/**
 * Read-only statistics of the executive summary compilation form, counted by the backend:
 * GET /executive-summaries/compilation-stats?year=&quarter=  (quarter 1-4)
 * Only LHA with status FINAL, APPROVED or PUBLISHED are counted by the server.
 *
 * Returns null on failure or when a newer request has replaced this one; `error` holds the message.
 */
export const useExecutiveSummarySources = () => {
  const loading = ref(false)
  const error = ref('')
  let requestId = 0

  const fetchCompilationStats = async (input: { year: number, quarter: number }): Promise<CompilationStats | null> => {
    const id = ++requestId
    loading.value = true
    error.value = ''
    try {
      const response = await $fetch<unknown>(`${getAuditServiceBaseUrl()}${COMPILATION_STATS_PATH}`, {
        method: 'GET',
        params: { year: input.year, quarter: input.quarter }
      })
      if (id !== requestId) return null
      return parseCompilationStats(response)
    } catch (err) {
      if (id !== requestId) return null
      console.error('Failed to load executive summary compilation statistics:', err)
      error.value = describeStatsError(err)
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

  return { loading, error, fetchCompilationStats, cancel }
}
