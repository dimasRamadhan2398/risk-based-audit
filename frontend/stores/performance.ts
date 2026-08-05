import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface KPIAchievement {
  id: string;
  year: number;
  period?: string;
  report_id?: string;
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

  const mockKpis: KPIAchievement[] = [
    { id: '1', year: 2026, period: 'Tahunan', kpi_name: 'Penyelesaian Program Kerja Audit Tahunan (PKAT)', target: 100, actual: 92, achievement_rate: 92, notes: '1 audit operasional ditunda ke Q1 2027 karena restrukturisasi unit bisnis' },
    { id: '2', year: 2026, period: 'Tahunan', kpi_name: 'Persentase Tindak Lanjut Rekomendasi Audit', target: 85, actual: 88, achievement_rate: 103.5, notes: 'Melebihi target karena implementasi sistem monitoring otomatis baru' },
    { id: '3', year: 2026, period: 'Tahunan', kpi_name: 'Indeks Kepuasan Auditee terhadap Layanan Audit', target: 80, actual: 82, achievement_rate: 102.5, notes: 'Survei akhir tahun menunjukkan kepuasan tinggi terhadap kejelasan rekomendasi' },
    { id: '4', year: 2026, period: 'Tahunan', kpi_name: 'Rata-rata Waktu Penyampaian Laporan Hasil Audit (LHA)', target: 14, actual: 15, achievement_rate: 93.3, notes: 'Target 14 hari kerja setelah exit meeting, rata-rata aktual 15 hari kerja' }
  ];

  const mockRealizations: WorkPlanRealization[] = [
    { id: '1', year: 2026, audit_annual_plan_id: 'AP-2026-01', planned_activities: 12, executed_activities: 11, realization_rate: 91.67, annual_plan: { title: 'Rencana Audit Tahunan 2026' } },
    { id: '2', year: 2026, audit_annual_plan_id: 'AP-2026-02', planned_activities: 8, executed_activities: 8, realization_rate: 100.00, annual_plan: { title: 'Audit Investigatif & Khusus 2026' } }
  ];

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig();
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1';
  };

  const fetchKPIAchievements = async (year: number = 2026, period?: string) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const params: Record<string, any> = { year };
      if (period && period !== 'Semua') {
        params.period = period;
      }
      const response: any = await $fetch(`${baseUrl}/performance/kpi`, {
        params
      });
      let fetched: KPIAchievement[] = [];
      if (response && Array.isArray(response.data) && response.data.length > 0) {
        fetched = response.data;
      } else if (Array.isArray(response) && response.length > 0) {
        fetched = response;
      } else {
        fetched = [...mockKpis];
      }
      if (period && period !== 'Semua') {
        kpiAchievements.value = fetched.filter(item => !item.period || item.period === period);
      } else {
        kpiAchievements.value = fetched;
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch KPI achievements';
      kpiAchievements.value = [...mockKpis];
    } finally {
      loading.value = false;
    }
  };

  const fetchWorkPlanRealizations = async (year: number = 2026) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/realization`, {
        params: { year }
      });
      if (response && Array.isArray(response.data) && response.data.length > 0) {
        workPlanRealizations.value = response.data;
      } else if (Array.isArray(response) && response.length > 0) {
        workPlanRealizations.value = response;
      } else {
        workPlanRealizations.value = [...mockRealizations];
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch Work Plan realizations';
      workPlanRealizations.value = [...mockRealizations];
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
