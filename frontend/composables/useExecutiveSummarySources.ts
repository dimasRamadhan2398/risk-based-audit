import { ref } from 'vue'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { parseCompilationStats, type CompilationStats } from '~/utils/executiveSummaryCompilation'
import { extractErrorMessage } from '~/utils/error'

/**
 * Read-only statistics of the executive summary compilation form, counted by the backend:
 * GET /executive-summary/compilation-stats?year=&quarter=  (quarter 1-4)
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
      const response = await $fetch<unknown>(`${getAuditServiceBaseUrl()}/executive-summary/compilation-stats`, {
        method: 'GET',
        params: { year: input.year, quarter: input.quarter }
      })
      if (id !== requestId) return null
      return parseCompilationStats(response)
    } catch (err) {
      if (id !== requestId) return null
      console.error('Failed to load executive summary compilation statistics:', err)
      error.value = extractErrorMessage(err, 'Failed to load compilation statistics.')
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
