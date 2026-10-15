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

export const useAuditExecutionStore = defineStore('audit-execution', {
  state: () => ({
    auditExecutions: [] as AuditExecution[],
    loading: false,
    error: null as string | null
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
    }
  }
})
