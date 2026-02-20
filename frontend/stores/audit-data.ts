import { defineStore } from "pinia";
import type {
  AuditMainStats,
  AuditCoverage,
  AuditDataState,
  AuditCategoryRisk,
  AtrReport,
  AuditExecutionStatus,
  RecentFinding,
} from "~/types/audit";

export const useAuditDataStore = defineStore("audit-data", {
  state: (): AuditDataState => ({
    auditMainStats: {
      plannedAudit: 50,
      openFinding: 15,
      executionStatus: 0.8,
      atrCompliance: 0.9,
    },
    auditMainStatsLastMonth: {
      plannedAudit: 45,
      openFinding: 18,
      executionStatus: 0.75,
      atrCompliance: 0.85,
    },
    auditCoverage: {
      plannedAudits: 30,
      completedAudits: 15,
      remainingAudits: 15,
    },
    auditCategories: [
      { id: 1, name: "Operational", inherentRisk: 16, residualRisk: 8 },
      { id: 2, name: "Finance", inherentRisk: 12, residualRisk: 6 },
      { id: 3, name: "IT", inherentRisk: 8, residualRisk: 4 },
      { id: 4, name: "Legal", inherentRisk: 10, residualRisk: 7 },
      { id: 5, name: "HR", inherentRisk: 14, residualRisk: 3 },
    ],
    auditExecutionStatus: [
      { id: 1, name: "Financial Operations", percentage: 1 },
      { id: 2, name: "IT Security Compliance", percentage: 0.5 },
      { id: 3, name: "HR Policies", percentage: 0 },
    ],
    atrReports: [
      { id: "AUD-2023-009", name: "Staff Training on GDPR", owner: "HR Dept",  date: "2023-12-10", status: "Pending" },
      { id: "AUD-2023-002", name: "Revise Procurement Policy", owner: "Sarah M.",  date: "2023-11-10", status: "In Progress" },
      { id: "AUD-2023-001", name: "Update Firewall Ruleset", owner: "John Doe",  date: "2023-11-10", status: "Completed"},
    ],
    recentFindings: [
      {id: 1, audit_finding: "Vendor Contract Missing", audit_category: "Procurement Audit", severity: "Medium"},
      {id: 2, audit_finding: "Unpatched Server", audit_category: "IT Security Audit", severity: "High"},
      {id: 3, audit_finding: "Cash Discrepancy", audit_category: "Finance Operations", severity: "Low"}
    ]
  }),

  getters: {
    getAuditMainStats: (state): AuditMainStats => state.auditMainStats,
    getAuditMainStatsLastMonth: (state): AuditMainStats =>
      state.auditMainStatsLastMonth,
    getAuditCoverage: (state): AuditCoverage => state.auditCoverage,
    getRiskData: (state): AuditCategoryRisk[] => state.auditCategories,
    getDropdownYear: (state): number[] => {
      const currentYear = new Date().getFullYear();
      return Array.from({ length: 3 }, (_, index) => currentYear - index);
    },
    getAtrReports: (state): AtrReport[] => state.atrReports,
    getAuditExecutionStatus: (state): AuditExecutionStatus[] => state.auditExecutionStatus,
    getRecentFindings: (state): RecentFinding[] => state.recentFindings
  },

  actions: {
    fetchAuditData() {
      // TODO: Implement API call to fetch audit data
    },
  },
});
