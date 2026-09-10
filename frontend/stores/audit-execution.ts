import { defineStore } from 'pinia'
import { AuditCategory, AuditStatus, type AuditExecution } from '~/types/audit'

export function normalizeAuditStatus(status?: string | null, progress?: number): string {
  if (!status && typeof progress === 'number') {
    if (progress >= 100) return 'completed'
    if (progress >= 76) return 'reporting'
    if (progress >= 51) return 'draft findings'
    if (progress >= 26) return 'fieldwork'
    if (progress >= 1) return 'entry meeting'
    return 'planning'
  }
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
    if (typeof progress === 'number') {
      if (progress >= 76) return 'reporting'
      if (progress >= 51) return 'draft findings'
      if (progress >= 26) return 'fieldwork'
      if (progress >= 1) return 'entry meeting'
    }
    return 'fieldwork'
  }
  return clean || 'planning'
}

export const useAuditExecutionStore = defineStore('audit-execution', {
  state: () => ({
    auditExecutions: [
      {
        id: '1',
        ref: 'AUD-2026-001',
        name: 'Financial Operations',
        category: AuditCategory.ASSURANCE,
        progress: 100,
        lead_auditor: 'Dimas P',
        status: 'completed',
        status_detail: 'Audit Completed',
        created_at: '2026-01-15T10:00:00Z',
        sample_data_test_controls: {
          progress: 100,
          description: 'Testing the efficiency of financial controls.'
        },
        working_papers: {
          condition: 'All data is in order.',
          criteria: 'Financial SOP No. 1'
        },
        action_plan_improvements: {
          recommendation: '-',
          deadline: '-',
          pic: '-'
        },
        latest_update_progress: {
          attachment: 'Financial Report.pdf',
          description: 'Audit completed without material findings.'
        }
      },
      {
        id: '2',
        ref: 'AUD-2026-002',
        name: 'Financial Operations',
        category: AuditCategory.ASSURANCE,
        progress: 40,
        lead_auditor: 'Sarah',
        status: 'fieldwork',
        status_detail: 'Fieldwork & Control Testing',
        created_at: '2026-04-10T10:00:00Z',
        sample_data_test_controls: {
          progress: 50,
          description: 'Testing the efficiency of controls & risk mitigation for IT security.'
        },
        working_papers: {
          condition: 'Several ports are still open.',
          criteria: 'ISO 27001'
        },
        action_plan_improvements: {
          recommendation: 'Close unused ports.',
          deadline: '2026-06-15',
          pic: 'IT Security'
        },
        latest_update_progress: {
          attachment: 'Scan_Result.jpg',
          description: 'In progress for hardening.'
        }
      },
      {
        id: '3',
        ref: 'AUD-2026-003',
        name: 'Vendor Risk Assessment',
        category: AuditCategory.SPECIAL_AUDIT,
        progress: 80,
        lead_auditor: 'Budi',
        status: 'reporting',
        status_detail: 'Reporting & Exit Meeting',
        created_at: '2026-07-20T10:00:00Z',
        sample_data_test_controls: {
          progress: 100,
          description: 'Vendor risk assessment'
        },
        working_papers: {
          condition: 'Vendor meets requirements.',
          criteria: 'Vendor Procurement Policy V.2'
        },
        action_plan_improvements: {
          recommendation: '-',
          deadline: '-',
          pic: '-'
        },
        latest_update_progress: {
          attachment: 'Vendor_Doc.pdf',
          description: 'Vendor is approved.'
        }
      },
      {
        id: '4',
        ref: 'AUD-2026-004',
        name: 'Human Capital & Payroll',
        category: AuditCategory.SPECIAL_AUDIT,
        progress: 15,
        lead_auditor: 'Budi Santoso',
        status: 'entry meeting',
        status_detail: 'Entry Meeting & Scope Alignment',
        created_at: '2026-10-05T10:00:00Z',
      },
      {
        id: '5',
        ref: 'AUD-2026-005',
        name: 'Logistics & Warehouse Audit',
        category: AuditCategory.ASSURANCE,
        progress: 60,
        lead_auditor: 'Sarah Putri',
        status: 'draft findings',
        status_detail: 'Draft Findings & Recommendations',
        created_at: '2026-02-20T10:00:00Z',
      },
      {
        id: '6',
        ref: 'AUD-2026-006',
        name: 'Procurement Unit Planning',
        category: AuditCategory.ASSURANCE,
        progress: 0,
        lead_auditor: 'Rina Wulandari',
        status: 'planning',
        status_detail: 'Planning & Preparation',
        created_at: '2026-03-01T10:00:00Z',
      }
    ] as AuditExecution[],
    loading: false,
    error: null as string | null,
  }),

  getters: {
    getSummary: (state) => {
      const completed = state.auditExecutions.filter((e: { status?: string; progress?: number }) => 
        normalizeAuditStatus(e.status, e.progress) === 'completed'
      ).length
      const inProgress = state.auditExecutions.filter((e: { status?: string; progress?: number }) => {
        const norm = normalizeAuditStatus(e.status, e.progress)
        return norm === 'entry meeting' || norm === 'fieldwork' || norm === 'draft findings' || norm === 'reporting' || norm === 'in progress'
      }).length
      const planned = state.auditExecutions.filter((e: { status?: string; progress?: number }) => 
        normalizeAuditStatus(e.status, e.progress) === 'planning'
      ).length
      return { completed, inProgress, planned }
    }
  },

  actions: {
    async fetchAuditExecutions() {
      this.loading = true
      this.error = null
      const mockList = [
        {
          id: '1',
          ref: 'AUD-2026-001',
          name: 'Financial Operations',
          category: 'Finance',
          progress: 100,
          lead_auditor: 'Dimas P',
          status: 'completed',
          status_detail: 'Audit Completed',
          created_at: '2026-01-15T10:00:00Z',
          sample_data_test_controls: {
            progress: 100,
            description: 'Testing the efficiency of financial controls.'
          },
          working_papers: {
            condition: 'All data is in order.',
            criteria: 'Financial SOP No. 1'
          },
          action_plan_improvements: {
            recommendation: '-',
            deadline: '-',
            pic: '-'
          },
          latest_update_progress: {
            attachment: 'Financial Report.pdf',
            description: 'Audit completed without material findings.'
          }
        },
        {
          id: '2',
          ref: 'AUD-2026-002',
          name: 'Financial Operations',
          category: AuditCategory.ASSURANCE,
          progress: 40,
          lead_auditor: 'Sarah',
          status: 'fieldwork',
          status_detail: 'Fieldwork & Control Testing',
          created_at: '2026-04-10T10:00:00Z',
          sample_data_test_controls: {
            progress: 50,
            description: 'Testing the efficiency of controls & risk mitigation for IT security.'
          },
          working_papers: {
            condition: 'Several ports are still open.',
            criteria: 'ISO 27001'
          },
          action_plan_improvements: {
            recommendation: 'Close unused ports.',
            deadline: '2026-06-15',
            pic: 'IT Security'
          },
          latest_update_progress: {
            attachment: 'Scan_Result.jpg',
            description: 'In progress for hardening.'
          }
        },
        {
          id: '3',
          ref: 'AUD-2026-003',
          name: 'Vendor Risk Assessment',
          category: AuditCategory.SPECIAL_AUDIT,
          progress: 80,
          lead_auditor: 'Budi',
          status: 'reporting',
          status_detail: 'Reporting & Exit Meeting',
          created_at: '2026-07-20T10:00:00Z',
          sample_data_test_controls: {
            progress: 100,
            description: 'Vendor risk assessment'
          },
          working_papers: {
            condition: 'Vendor meets requirements.',
            criteria: 'Vendor Procurement Policy V.2'
          },
          action_plan_improvements: {
            recommendation: '-',
            deadline: '-',
            pic: '-'
          },
          latest_update_progress: {
            attachment: 'Vendor_Doc.pdf',
            description: 'Vendor is approved.'
          }
        },
        {
          id: '4',
          ref: 'AUD-2026-004',
          name: 'Human Capital & Payroll',
          category: AuditCategory.SPECIAL_AUDIT,
          progress: 15,
          lead_auditor: 'Budi Santoso',
          status: 'entry meeting',
          status_detail: 'Entry Meeting & Scope Alignment',
          created_at: '2026-10-05T10:00:00Z',
        },
        {
          id: '5',
          ref: 'AUD-2026-005',
          name: 'Logistics & Warehouse Audit',
          category: AuditCategory.ASSURANCE,
          progress: 60,
          lead_auditor: 'Sarah Putri',
          status: 'draft findings',
          status_detail: 'Draft Findings & Recommendations',
          created_at: '2026-02-20T10:00:00Z',
        },
        {
          id: '6',
          ref: 'AUD-2026-006',
          name: 'Procurement Unit Planning',
          category: AuditCategory.ASSURANCE,
          progress: 0,
          lead_auditor: 'Rina Wulandari',
          status: 'planning',
          status_detail: 'Planning & Preparation',
          created_at: '2026-03-01T10:00:00Z',
        }
      ] as AuditExecution[]

      try {
        const config = useRuntimeConfig()
        const baseUrl = config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
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

        if (items.length > 0) {
          this.auditExecutions = items.map(item => ({
            ...item,
            status: normalizeAuditStatus(item.status, item.progress)
          }))
        } else {
          this.auditExecutions = mockList
        }
      } catch (err: any) {
        console.error('Failed to fetch audit executions:', err)
        this.error = err.message
        this.auditExecutions = mockList
      } finally {
        this.loading = false
      }
    },

    async updateAuditExecution(id: string, payload: Partial<AuditExecution>) {
      this.loading = true
      try {
        const config = useRuntimeConfig()
        const baseUrl = config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
        await $fetch(`${baseUrl}/audit-executions/${id}`, {
          method: 'PUT',
          body: payload
        })
        await this.fetchAuditExecutions()
      } catch (err: any) {
        console.error('Failed to update audit execution:', err)
      } finally {
        this.loading = false
      }
    }
  }
})
