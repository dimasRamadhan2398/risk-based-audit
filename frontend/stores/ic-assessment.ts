// Internal Control Assessment (COSO 2013) store
//
// Talks to audit-service /api/v1/ic-assessments (plan section 8.2). All response parsing goes
// through utils/icAssessmentParse.ts. Actions never throw: they return null/false and leave the
// error in `lastError` (409 also sets `conflict`), so pages decide how to show it.
import { defineStore } from 'pinia'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import type {
  IcAssessment,
  IcAssessmentListItem,
  IcDrift,
  IcEvent,
  IcEvidence,
  IcPeriod,
  IcSignature,
  IcThresholds
} from '~/utils/icAssessment'
import type { IcSummary } from '~/utils/icAssessmentParse'
import {
  buildApprovePayload,
  buildThresholdsPayload,
  buildUpdatePayload,
  parseAssessment,
  parseAssessmentList,
  parseDrift,
  parseEvents,
  parseEvidenceList,
  parseSummary,
  parsePeriods,
  parseThresholds,
  unwrap
} from '~/utils/icAssessmentParse'

export interface IcListFilters {
  year: number
  /** 0 = annual */
  quarter: number
  department_id?: string
  location_id?: string
  status?: string
  page?: number
  page_size?: number
}

const statusOf = (e: unknown): number | undefined => {
  const err = e as { statusCode?: number, status?: number, response?: { status?: number } } | null
  return err?.statusCode ?? err?.status ?? err?.response?.status
}

export const icAssessmentBaseUrl = () => `${getAuditServiceBaseUrl()}/ic-assessments`

export const useIcAssessmentStore = defineStore('icAssessment', {
  state: () => ({
    items: [] as IcAssessmentListItem[],
    total: 0,
    listLoading: false,

    current: null as IcAssessment | null,
    detailLoading: false,
    saving: false,

    events: [] as IcEvent[],
    eventsLoading: false,

    /** Evidence per proposal run, cached for the drawer's run selector. */
    evidenceByRun: {} as Record<number, { items: IcEvidence[], thresholds: IcThresholds | null }>,
    evidenceLoading: false,

    periods: [] as IcPeriod[],

    drift: null as IcDrift | null,
    /** The drift check endpoint is missing or failed: the indicator shows "not available". */
    driftUnavailable: false,
    driftLoading: false,

    thresholds: null as IcThresholds | null,
    thresholdsLoading: false,

    /** Dashboard roll-up (GET /summary). null = not loaded or failed; check summaryFailed. */
    summary: null as IcSummary | null,
    summaryLoading: false,
    summaryFailed: false,

    lastError: null as unknown,
    conflict: false
  }),

  actions: {
    fail(e: unknown) {
      console.error('[ic-assessment]', e)
      this.lastError = e
      if (statusOf(e) === 409) this.conflict = true
    },

    clearError() {
      this.lastError = null
      this.conflict = false
    },

    async fetchList(filters: IcListFilters): Promise<boolean> {
      this.listLoading = true
      this.clearError()
      try {
        const res = await $fetch(icAssessmentBaseUrl(), {
          query: {
            year: filters.year,
            quarter: filters.quarter,
            department_id: filters.department_id || undefined,
            location_id: filters.location_id || undefined,
            status: filters.status || undefined,
            page: filters.page ?? 1,
            page_size: filters.page_size ?? 500
          }
        })
        const parsed = parseAssessmentList(res)
        this.items = parsed.items
        this.total = parsed.total
        return true
      } catch (e) {
        this.fail(e)
        this.items = []
        this.total = 0
        return false
      } finally {
        this.listLoading = false
      }
    },

    /** Approved-assessment counts per component for the dashboards. Never throws. */
    async fetchSummary(year: number, quarter = 0): Promise<IcSummary | null> {
      this.summaryLoading = true
      this.summaryFailed = false
      try {
        this.summary = parseSummary(await $fetch(`${icAssessmentBaseUrl()}/summary`, { query: { year, quarter } }))
        return this.summary
      } catch (e) {
        console.error('[ic-assessment] summary', e)
        this.summary = null
        this.summaryFailed = true
        return null
      } finally {
        this.summaryLoading = false
      }
    },

    async fetchPeriods(year: number): Promise<void> {
      try {
        this.periods = parsePeriods(await $fetch(`${icAssessmentBaseUrl()}/periods`, { query: { year } }))
      } catch (e) {
        // Lock state is advisory in the UI (the backend refuses writes on a locked period).
        console.error('[ic-assessment] periods', e)
        this.periods = []
      }
    },

    async lockPeriod(year: number, quarter: number, note: string): Promise<boolean> {
      return this.periodAction('lock', { year, quarter, note: note || undefined })
    },

    async unlockPeriod(year: number, quarter: number, reason: string): Promise<boolean> {
      return this.periodAction('unlock', { year, quarter, reason })
    },

    async periodAction(kind: 'lock' | 'unlock', body: Record<string, unknown>): Promise<boolean> {
      this.saving = true
      this.clearError()
      try {
        await $fetch(`${icAssessmentBaseUrl()}/periods/${kind}`, { method: 'POST', body })
        await this.fetchPeriods(Number(body.year))
        return true
      } catch (e) {
        this.fail(e)
        return false
      } finally {
        this.saving = false
      }
    },

    async create(input: { department_id: string, year: number, quarter: number, location_id?: string | null }): Promise<IcAssessment | null> {
      this.saving = true
      this.clearError()
      try {
        const body: Record<string, unknown> = { department_id: input.department_id, year: input.year, quarter: input.quarter }
        if (input.location_id) body.location_id = input.location_id
        const created = parseAssessment(await $fetch(icAssessmentBaseUrl(), { method: 'POST', body }))
        return created
      } catch (e) {
        this.fail(e)
        return null
      } finally {
        this.saving = false
      }
    },

    async fetchDetail(id: string): Promise<IcAssessment | null> {
      this.detailLoading = true
      this.clearError()
      this.drift = null
      this.driftUnavailable = false
      this.evidenceByRun = {}
      try {
        const a = parseAssessment(await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}`))
        this.setCurrent(a)
        return a
      } catch (e) {
        this.fail(e)
        this.current = null
        return null
      } finally {
        this.detailLoading = false
      }
    },

    setCurrent(a: IcAssessment | null) {
      this.current = a
      if (a && a.proposalRun > 0 && a.evidence.length) {
        this.evidenceByRun[a.proposalRun] = { items: a.evidence, thresholds: a.thresholdsUsed }
      }
    },

    /** Replaces `current` with the server reply of a write; refetches when the reply carries no body. */
    async applyWriteResult(id: string, res: unknown): Promise<IcAssessment | null> {
      const parsed = parseAssessment(res)
      if (parsed) {
        this.setCurrent(parsed)
        return parsed
      }
      return this.fetchDetail(id)
    },

    /** Saves only what changed between `original` and `draft`. Returns the saved assessment, or the original when nothing changed. */
    async save(original: IcAssessment, draft: IcAssessment): Promise<IcAssessment | null> {
      const body = buildUpdatePayload(original, draft)
      if (!body) return original
      this.saving = true
      this.clearError()
      try {
        const res = await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(original.id)}`, { method: 'PUT', body })
        return await this.applyWriteResult(original.id, res)
      } catch (e) {
        this.fail(e)
        return null
      } finally {
        this.saving = false
      }
    },

    async transition(id: string, action: 'propose' | 'submit' | 'return' | 'reopen' | 'approve', body: Record<string, unknown>): Promise<IcAssessment | null> {
      this.saving = true
      this.clearError()
      try {
        const res = await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}/${action}`, { method: 'POST', body })
        return await this.applyWriteResult(id, res)
      } catch (e) {
        this.fail(e)
        return null
      } finally {
        this.saving = false
      }
    },

    propose(a: IcAssessment) {
      return this.transition(a.id, 'propose', { version: a.version })
    },
    submit(a: IcAssessment) {
      return this.transition(a.id, 'submit', { version: a.version })
    },
    returnToDraft(a: IcAssessment, reason: string) {
      return this.transition(a.id, 'return', { version: a.version, reason })
    },
    reopen(a: IcAssessment, reason: string) {
      return this.transition(a.id, 'reopen', { version: a.version, reason })
    },
    approve(a: IcAssessment, input: { operateTogether: boolean, operateTogetherNote?: string, approvalNote?: string, signature?: IcSignature | null }) {
      return this.transition(a.id, 'approve', { ...buildApprovePayload({ version: a.version, ...input }) })
    },

    async remove(id: string): Promise<boolean> {
      this.saving = true
      this.clearError()
      try {
        await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}`, { method: 'DELETE' })
        this.items = this.items.filter(i => i.id !== id)
        if (this.current?.id === id) this.current = null
        return true
      } catch (e) {
        this.fail(e)
        return false
      } finally {
        this.saving = false
      }
    },

    async fetchEvidence(id: string, run: number): Promise<{ items: IcEvidence[], thresholds: IcThresholds | null } | null> {
      const cached = this.evidenceByRun[run]
      if (cached) return cached
      this.evidenceLoading = true
      try {
        const parsed = parseEvidenceList(await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}/evidence`, { query: { run } }))
        const entry = { items: parsed.items, thresholds: parsed.thresholds }
        this.evidenceByRun[run] = entry
        return entry
      } catch (e) {
        this.fail(e)
        return null
      } finally {
        this.evidenceLoading = false
      }
    },

    async fetchEvents(id: string): Promise<void> {
      this.eventsLoading = true
      try {
        this.events = parseEvents(await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}/events`))
      } catch (e) {
        console.error('[ic-assessment] events', e)
        this.events = []
      } finally {
        this.eventsLoading = false
      }
    },

    /**
     * Read-only re-run of the collectors for an approved assessment (plan 7.3, decision Q5).
     * GET /ic-assessments/:id/drift. 409 NO_PROPOSAL (no proposal yet) and 403 (scoped readers) show "not available".
     */
    async checkDrift(id: string): Promise<IcDrift | null> {
      this.driftLoading = true
      this.driftUnavailable = false
      try {
        this.drift = parseDrift(await $fetch(`${icAssessmentBaseUrl()}/${encodeURIComponent(id)}/drift`))
        if (!this.drift) this.driftUnavailable = true
        return this.drift
      } catch (e) {
        console.error('[ic-assessment] drift', e)
        this.drift = null
        this.driftUnavailable = true
        return null
      } finally {
        this.driftLoading = false
      }
    },

    /** Tenant thresholds (decision Q4). GET/PUT /ic-assessments/settings, ADMIN only. */
    async fetchThresholds(): Promise<IcThresholds | null> {
      this.thresholdsLoading = true
      this.clearError()
      try {
        this.thresholds = parseThresholds(unwrap(await $fetch(`${icAssessmentBaseUrl()}/settings`)))
        return this.thresholds
      } catch (e) {
        this.fail(e)
        return null
      } finally {
        this.thresholdsLoading = false
      }
    },

    async saveThresholds(next: IcThresholds): Promise<boolean> {
      const body = buildThresholdsPayload(this.thresholds, next)
      if (!body) return true
      this.saving = true
      this.clearError()
      try {
        const res = await $fetch(`${icAssessmentBaseUrl()}/settings`, { method: 'PUT', body })
        this.thresholds = parseThresholds(unwrap(res)) ?? { ...next }
        return true
      } catch (e) {
        this.fail(e)
        return false
      } finally {
        this.saving = false
      }
    }
  }
})
