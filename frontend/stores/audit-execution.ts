import { defineStore } from 'pinia'
import { AuditCategory, AuditStatus, type AuditExecution } from '~/types/audit'
import { extractErrorMessage } from '~/utils/error'
import { getDisplayProgress, progressToStageStatus } from '~/utils/executionSources'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'

/**
 * Canonical stage status for an execution. A recognised status string wins;
 * otherwise (or for a generic "in progress") the stage follows the progress.
 */
export function normalizeAuditStatus(status?: string | null, progress?: number): string {
  if (!status) return progressToStageStatus(progress)
  const clean = String(status || '').toLowerCase().replace(/[\s_-]+/g, ' ').trim()
  if (clean === 'completed' || clean === 'done' || clean === 'finished' || clean === 'audit completed') {
    return 'completed'
  }
  if (clean === 'reporting' || clean === 'reporting & exit meeting' || clean === 'reporting exit meeting') {
    return 'reporting'
  }
  if (clean === 'draft findings' || clean === 'draft findings & recommendations' || clean === 'draft findings recommendations') {
    return 'draft findings'
  }
  if (clean === 'fieldwork' || clean === 'fieldwork & control testing' || clean === 'fieldwork control testing') {
    return 'fieldwork'
  }
  if (clean === 'entry meeting' || clean === 'entry meeting & scope alignment' || clean === 'entry meeting scope alignment') {
    return 'entry meeting'
  }
  if (clean === 'planning' || clean === 'planned' || clean === 'planning & preparation' || clean === 'planning preparation') {
    return 'planning'
  }
  if (clean === 'in progress') {
    return typeof progress === 'number' ? progressToStageStatus(progress) : 'fieldwork'
  }
  return clean || 'planning'
}

/** Outcome of POST /audit-executions/sync. */
export interface AuditExecutionSyncResult {
  /** Executions recomputed by the sync. */
  synced: number
  /** Executions newly created from assignment letters that had none. */
  created: number
  /** Source rows (fieldwork, working paper, ...) whose letter reference matches no assignment letter. */
  unmatched: number
}

/** Body of POST /audit-executions/sync (under `data`). */
interface SyncResponse {
  synced?: number
  created?: number
  items?: unknown[]
  unmatched_source_rows?: Record<string, number>
}

const toCount = (value: unknown): number => {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0
}

export const useAuditExecutionStore = defineStore('audit-execution', {
  state: () => ({
    auditExecutions: [] as AuditExecution[],
    loading: false,
    error: null as string | null,
    syncing: false,
    syncError: null as string | null
  }),

  getters: {
    getSummary: (state) => {
      const completed = state.auditExecutions.filter((e: { status?: string, progress?: number }) =>
        normalizeAuditStatus(e.status, e.progress) === 'completed'
      ).length
      const inProgress = state.auditExecutions.filter((e: { status?: string, progress?: number }) => {
        const norm = normalizeAuditStatus(e.status, e.progress)
        return norm === 'entry meeting' || norm === 'fieldwork' || norm === 'draft findings' || norm === 'reporting' || norm === 'in progress'
      }).length
      const planned = state.auditExecutions.filter((e: { status?: string, progress?: number }) =>
        normalizeAuditStatus(e.status, e.progress) === 'planning'
      ).length
      return { completed, inProgress, planned }
    }
  },

  actions: {
    async fetchAuditExecutions() {
      this.loading = true
      this.error = null

      try {
        const baseUrl = getAuditServiceBaseUrl()
        const response: any = await $fetch(`${baseUrl}/audit-executions`, {
          method: 'GET'
        })
        let items: AuditExecution[] = []
        if (response && response.data && Array.isArray(response.data.items)) {
          items = response.data.items
        } else if (response && Array.isArray(response.items)) {
          items = response.items
        } else if (Array.isArray(response)) {
          items = response
        }

        // Read-only: progress is derived on the server. Here it is only coerced, reconciled with the
        // LHA state (waiting approval is never 100), and the stage follows that progress.
        this.auditExecutions = items.map((item) => {
          const progress = getDisplayProgress(item.progress, item.sources)
          return {
            ...item,
            progress,
            status: progressToStageStatus(progress)
          }
        })
      } catch (err: any) {
        console.error('Failed to fetch audit executions:', err)
        this.error = extractErrorMessage(err, 'Failed to load audit executions.')
        this.auditExecutions = []
      } finally {
        this.loading = false
      }
    },

    async updateAuditExecution(id: string, payload: Partial<AuditExecution>) {
      this.loading = true
      try {
        const baseUrl = getAuditServiceBaseUrl()
        await $fetch(`${baseUrl}/audit-executions/${id}`, {
          method: 'PUT',
          body: payload
        })
        await this.fetchAuditExecutions()
      } catch (err: any) {
        console.error('Failed to update audit execution:', err)
        this.error = extractErrorMessage(err, 'Failed to update audit execution.')
        throw err
      } finally {
        this.loading = false
      }
    },

    /**
     * Creates the missing executions for every live assignment letter on the server,
     * recomputes progress, then reloads the list. Returns null when the sync failed
     * (message in syncError; the current list is left as it is) or one is already running.
     */
    async syncFromAssignmentLetters(): Promise<AuditExecutionSyncResult | null> {
      if (this.syncing) return null
      this.syncing = true
      this.syncError = null

      try {
        const baseUrl = getAuditServiceBaseUrl()
        const response = await $fetch<SyncResponse | { data?: SyncResponse }>(`${baseUrl}/audit-executions/sync`, {
          method: 'POST'
        })
        const data: SyncResponse = (response && 'data' in response ? response.data : response as SyncResponse) ?? {}
        const unmatchedRows = data.unmatched_source_rows
        const result: AuditExecutionSyncResult = {
          synced: toCount(data.synced),
          created: toCount(data.created),
          unmatched: unmatchedRows && typeof unmatchedRows === 'object'
            ? Object.values(unmatchedRows).reduce<number>((sum, n) => sum + toCount(n), 0)
            : 0
        }
        await this.fetchAuditExecutions()
        return result
      } catch (err) {
        console.error('Failed to sync audit executions from assignment letters:', err)
        this.syncError = extractErrorMessage(err, 'Failed to sync audits from assignment letters.')
        return null
      } finally {
        this.syncing = false
      }
    }
  }
})
