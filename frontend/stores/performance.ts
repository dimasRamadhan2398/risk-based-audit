import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface KPIAchievement {
  id: string;
  year: number;
  kpi_name: string;
  target: number;
  actual: number;
  achievement_rate: number;
  notes: string;
}

export interface WorkPlanRealization {
  id: string;
  year: number;
  audit_annual_plan_id: string;
  annual_plan?: {
    title: string;
  };
  planned_activities: number;
  executed_activities: number;
  realization_rate: number;
}

export const usePerformanceStore = defineStore('performance', () => {
  const kpiAchievements = ref<KPIAchievement[]>([]);
  const workPlanRealizations = ref<WorkPlanRealization[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig();
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1';
  };

  const fetchKPIAchievements = async (year: number = 2024) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/kpi`, {
        params: { year }
      });
      kpiAchievements.value = response.data || [];
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch KPI achievements';
    } finally {
      loading.value = false;
    }
  };

  const fetchWorkPlanRealizations = async (year: number = 2024) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/realization`, {
        params: { year }
      });
      workPlanRealizations.value = response.data || [];
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch Work Plan realizations';
    } finally {
      loading.value = false;
    }
  };

  return {
    kpiAchievements,
    workPlanRealizations,
    loading,
    error,
    fetchKPIAchievements,
    fetchWorkPlanRealizations
  };
});
