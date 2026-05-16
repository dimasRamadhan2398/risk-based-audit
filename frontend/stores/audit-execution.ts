import { defineStore } from 'pinia'
import { AuditCategory, AuditStatus, type AuditExecution } from '~/types/audit'

export const useAuditExecutionStore = defineStore('audit-execution', {
  state: () => ({
    auditExecutions: [
      {
        id: '1',
        ref: 'AUD-2026-001',
        name: 'Financial Operations',
        category: 'Finance',
        progress: 100,
        lead_auditor: 'Dimas P',
        status: AuditStatus.COMPLETED,
        status_detail: 'On Time',
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
        progress: 50,
        lead_auditor: 'Sarah',
        status: AuditStatus.IN_PROGRESS,
        status_detail: 'On Time',
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
        progress: 100,
        lead_auditor: 'Budi',
        status: AuditStatus.COMPLETED,
        status_detail: 'On Time',
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
        name: 'Vendor Risk Assessment',
        category: AuditCategory.SPECIAL_AUDIT,
        progress: 0,
        lead_auditor: 'Budi',
        status: AuditStatus.PLANNED,
      },
      {
        id: '5',
        ref: 'AUD-2026-005',
        name: 'Inventory Audit',
        category: AuditCategory.ASSURANCE,
        progress: 0,
        lead_auditor: 'Budi',
        status: AuditStatus.PLANNED,
        status_detail: 'On Time',
        sample_data_test_controls: {
          progress: 100,
          description: 'Testing the efficiency of controls & risk mitigation for IT security.'
        },
        working_papers: {
          condition: 'Inventory difference 5% (Warehouse A)',
          criteria: 'OP Inventory No. 12'
        },
        action_plan_improvements: {
          recommendation: 'Re-count stock & double lock.',
          deadline: '2026-02-20',
          pic: 'Logistics Department'
        },
        latest_update_progress: {
          attachment: 'Fiskal_Photo.jpg',
          description: 'Still counting. Issue: Lack of night shift personnel.'
        }
      }
    ] as AuditExecution[],
    loading: false,
    error: null as string | null,
  }),

  getters: {
    getSummary: (state) => {
      const completed = state.auditExecutions.filter((e: { status: AuditStatus }) => e.status === AuditStatus.COMPLETED).length
      const inProgress = state.auditExecutions.filter((e: { status: AuditStatus }) => e.status === AuditStatus.IN_PROGRESS).length
      const planned = state.auditExecutions.filter((e: { status: AuditStatus }) => e.status === AuditStatus.PLANNED).length
      return { completed, inProgress, planned }
    }
  },

  actions: {
    async fetchAuditExecutions() {
      // Logic to fetch data from API
    }
  }
})
