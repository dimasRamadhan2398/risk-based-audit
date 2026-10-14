import { defineStore } from "pinia";
import type {
  AuditMainStats,
  AuditCoverage,
  AuditDataState,
  AuditCategoryRisk,
  AtrReport,
  AuditExecutionStatusItem,
  RecentFinding,
} from "~/types/audit";

export const useAuditDataStore = defineStore("audit-data", {
  state: (): AuditDataState => ({
    auditMainStats: {
      plannedAudit: 0,
      openFinding: 0,
      executionStatus: 0,
      atrCompliance: 0,
    },
    auditMainStatsLastMonth: {
      plannedAudit: 0,
      openFinding: 0,
      executionStatus: 0,
      atrCompliance: 0,
    },
    auditCoverage: {
      plannedAudits: 0,
      completedAudits: 0,
      remainingAudits: 0,
    },
    auditCategories: [],
    auditExecutionStatus: [],
    atrReports: [],
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
    getAuditExecutionStatus: (state): AuditExecutionStatusItem[] => state.auditExecutionStatus
  },

  actions: {
    fetchAuditData() {
      // TODO: Implement API call to fetch audit data
    },
  },
});
