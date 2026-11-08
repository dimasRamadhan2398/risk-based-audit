import { ref, shallowRef } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { describeStatsError } from '~/composables/useExecutiveSummarySources'
import type { FollowUpStatusQuery, FollowUpStatusResult } from '~/utils/executiveSummaryFollowUp'

/**
 * Loads the ATR follow-up status of an executive summary (one instance per use: the draft form fill,
 * or the read-only "current status" panel of an approved summary).
 * `load` returns null on failure or when a newer request has replaced this one; `error` holds the message.
 */
export const useExecutiveSummaryFollowUp = () => {
  const store = useExecutiveSummaryStore()
  const loading = ref(false)
  const error = ref('')
  const data = shallowRef<FollowUpStatusResult | null>(null)
  let requestId = 0

  const load = async (query: FollowUpStatusQuery): Promise<FollowUpStatusResult | null> => {
    const id = ++requestId
    loading.value = true
    error.value = ''
    try {
      const result = await store.fetchFollowUpStatus(query)
      if (id !== requestId) return null
      data.value = result
      return result
    } catch (err) {
      if (id !== requestId) return null
      console.error('Failed to load executive summary follow-up status:', err)
      error.value = describeStatsError(err, 'Failed to load the follow-up status.')
      return null
    } finally {
      if (id === requestId) loading.value = false
    }
  }

  /** Drops any in-flight request and clears the state. */
  const reset = () => {
    requestId++
    loading.value = false
    error.value = ''
    data.value = null
  }

  return { loading, error, data, load, reset }
}
