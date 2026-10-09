import { defineStore } from 'pinia';
import { ref } from 'vue';
import { extractErrorMessage } from '~/utils/error';
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl';
import { useI18n } from '~/composables/useI18n';
import {
  buildKpiBreakdownParams,
  emptyKpiBreakdownFilterOptions,
  emptyKpiBreakdownPagination,
  parseKpiBreakdownFilters,
  parseKpiBreakdownPagination,
  KPI_BREAKDOWN_DEFAULT_PAGE_SIZE,
  KPI_BREAKDOWN_MAX_PAGE_SIZE,
  type KpiBreakdownFilterOptions,
  type KpiBreakdownItem,
  type KpiBreakdownPagination,
  type KpiBreakdownQuery
} from '~/utils/kpiBreakdown';

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

export interface SubMetric {
  title: string;
  value: string;
  target: string;
  trend: string;
}

export interface SummaryCardData {
  title: string;
  key: string;
  value: string;
  target: string;
  actual_number: number;
  target_number: number;
  gap: string;
  trend: string;
  trend_up: boolean;
  unit: string;
  sub_metrics?: SubMetric[];
}

export interface MonthlyTrendData {
  labels: string[];
  completion_rate_series: number[];
  timeliness_series: number[];
  csat_series: number[];
}

export const usePerformanceStore = defineStore('performance', () => {
  const { t } = useI18n();
  const kpiAchievements = ref<KPIAchievement[]>([]);
  const workPlanRealizations = ref<WorkPlanRealization[]>([]);
  const dashboardCards = ref<SummaryCardData[]>([]);
  const monthlyTrends = ref<MonthlyTrendData | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  // --- KPI Detailed Breakdown (server-side filtered and paged) ---
  // Kept apart from `loading`/`error` above so the table does not flicker with the cards and charts.
  const kpiBreakdown = ref<KpiBreakdownItem[]>([]);
  const kpiBreakdownPagination = ref<KpiBreakdownPagination>(emptyKpiBreakdownPagination());
  const kpiBreakdownLoading = ref(false);
  const kpiBreakdownError = ref<string | null>(null);
  // `search` here is the applied (debounced) term; the input box holds the live text.
  const kpiBreakdownQuery = ref<KpiBreakdownQuery>({
    year: new Date().getFullYear(),
    page: 1,
    pageSize: KPI_BREAKDOWN_DEFAULT_PAGE_SIZE,
    search: '',
    category: '',
    status: '',
    period: ''
  });
  // Category/period menu options for the year, as the API lists them in `data.filters` (empty with an older backend).
  const kpiBreakdownFilters = ref<KpiBreakdownFilterOptions>(emptyKpiBreakdownFilterOptions());
  let kpiBreakdownRequestId = 0;
  let kpiBreakdownSearchTimer: ReturnType<typeof setTimeout> | null = null;
  const KPI_BREAKDOWN_SEARCH_DEBOUNCE_MS = 300;

  const fetchKpiBreakdown = async () => {
    const requestId = ++kpiBreakdownRequestId;
    const query = { ...kpiBreakdownQuery.value };
    kpiBreakdownLoading.value = true;
    kpiBreakdownError.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/kpi-breakdown`, {
        params: buildKpiBreakdownParams(query)
      });
      // A newer request (filter/page change) was started meanwhile; its result wins.
      if (requestId !== kpiBreakdownRequestId) return;

      const items: KpiBreakdownItem[] = Array.isArray(response?.data?.items) ? response.data.items : [];
      const pagination = parseKpiBreakdownPagination(response?.data?.pagination, query);

      // The requested page no longer exists (e.g. rows were removed): jump to the last one.
      if (items.length === 0 && pagination.total > 0 && pagination.total_pages > 0 && query.page > pagination.total_pages) {
        kpiBreakdownQuery.value.page = pagination.total_pages;
        await fetchKpiBreakdown();
        return;
      }

      kpiBreakdown.value = items;
      kpiBreakdownPagination.value = pagination;
      kpiBreakdownQuery.value.page = pagination.page;
      kpiBreakdownQuery.value.pageSize = pagination.page_size;
      kpiBreakdownFilters.value = parseKpiBreakdownFilters(response?.data?.filters);
    } catch (err: any) {
      if (requestId !== kpiBreakdownRequestId) return;
      console.error('Failed to fetch KPI breakdown:', err);
      kpiBreakdownError.value = extractErrorMessage(err, t('kpiPerformance.store.fetchBreakdownFailed'));
      kpiBreakdown.value = [];
      kpiBreakdownPagination.value = emptyKpiBreakdownPagination(query.pageSize);
    } finally {
      if (requestId === kpiBreakdownRequestId) kpiBreakdownLoading.value = false;
    }
  };

  const cancelKpiBreakdownSearch = () => {
    if (kpiBreakdownSearchTimer) {
      clearTimeout(kpiBreakdownSearchTimer);
      kpiBreakdownSearchTimer = null;
    }
  };

  /** Change year/filters/page size; anything that changes goes back to page 1 and refetches. */
  const setKpiBreakdownFilters = (changes: Partial<Omit<KpiBreakdownQuery, 'page'>>) => {
    const current = kpiBreakdownQuery.value;
    const next: KpiBreakdownQuery = { ...current, ...changes, page: 1 };
    next.pageSize = Math.min(KPI_BREAKDOWN_MAX_PAGE_SIZE, Math.max(1, Number(next.pageSize) || KPI_BREAKDOWN_DEFAULT_PAGE_SIZE));
    next.search = (next.search ?? '').trim();
    next.category = next.category ?? '';
    next.status = next.status ?? '';
    next.period = next.period ?? '';
    const changed = (Object.keys(changes) as Array<keyof KpiBreakdownQuery>).some(k => next[k] !== current[k]);
    if (!changed) return;
    if ('search' in changes) cancelKpiBreakdownSearch();
    if (next.year !== current.year) kpiBreakdownFilters.value = emptyKpiBreakdownFilterOptions();
    kpiBreakdownQuery.value = next;
    return fetchKpiBreakdown();
  };

  /** Debounced search: only the last term typed within the window is sent. */
  const setKpiBreakdownSearch = (term: string) => {
    cancelKpiBreakdownSearch();
    const value = (term ?? '').trim();
    if (value === kpiBreakdownQuery.value.search) return;
    kpiBreakdownSearchTimer = setTimeout(() => {
      kpiBreakdownSearchTimer = null;
      setKpiBreakdownFilters({ search: value });
    }, KPI_BREAKDOWN_SEARCH_DEBOUNCE_MS);
  };

  const setKpiBreakdownPage = (page: number) => {
    const target = Math.max(1, Math.trunc(Number(page)) || 1);
    if (target === kpiBreakdownQuery.value.page) return;
    kpiBreakdownQuery.value.page = target;
    return fetchKpiBreakdown();
  };

  /** First load / year prop change: a new year starts at page 1, the same year reloads the current page. */
  const loadKpiBreakdown = (year: number) => {
    if (year !== kpiBreakdownQuery.value.year) return setKpiBreakdownFilters({ year });
    return fetchKpiBreakdown();
  };

  const resetKpiBreakdownFilters = () => {
    cancelKpiBreakdownSearch();
    return setKpiBreakdownFilters({ search: '', category: '', status: '', period: '' });
  };

  const fetchDashboardSummary = async (year: number = 2026) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/dashboard-summary`, {
        params: { year }
      });
      // Only what the API returned: no cards (or another year's cards) when the response has none.
      dashboardCards.value = Array.isArray(response?.data) ? response.data : [];
    } catch (err: any) {
      console.error('Failed to fetch dashboard summary:', err);
      dashboardCards.value = [];
      error.value = extractErrorMessage(err, t('kpiPerformance.store.fetchDashboardSummaryFailed'));
    } finally {
      loading.value = false;
    }
  };

  const fetchMonthlyTrends = async (year: number = 2026) => {
    loading.value = true;
    error.value = null;
    try {
      const baseUrl = getAuditServiceBaseUrl();
      const response: any = await $fetch(`${baseUrl}/performance/monthly-trends`, {
        params: { year }
      });
      monthlyTrends.value = response?.data && typeof response.data === 'object' ? response.data : null;
    } catch (err: any) {
      console.error('Failed to fetch monthly trends:', err);
      monthlyTrends.value = null;
      error.value = extractErrorMessage(err, t('kpiPerformance.store.fetchMonthlyTrendsFailed'));
    } finally {
      loading.value = false;
    }
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
      // Only what the API returned: an empty response stays empty.
      let fetched: KPIAchievement[] = [];
      if (response && Array.isArray(response.data)) {
        fetched = response.data;
      } else if (Array.isArray(response)) {
        fetched = response;
      }
      if (period && period !== 'Semua') {
        kpiAchievements.value = fetched.filter(item => !item.period || item.period === period);
      } else {
        kpiAchievements.value = fetched;
      }
    } catch (err: any) {
      error.value = extractErrorMessage(err, t('kpiPerformance.store.fetchKpiFailed'));
      kpiAchievements.value = [];
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
      if (response && Array.isArray(response.data)) {
        workPlanRealizations.value = response.data;
      } else if (Array.isArray(response)) {
        workPlanRealizations.value = response;
      } else {
        workPlanRealizations.value = [];
      }
    } catch (err: any) {
      error.value = extractErrorMessage(err, t('kpiPerformance.store.fetchRealizationFailed'));
      workPlanRealizations.value = [];
    } finally {
      loading.value = false;
    }
  };

  return {
    kpiAchievements,
    workPlanRealizations,
    dashboardCards,
    monthlyTrends,
    loading,
    error,
    fetchDashboardSummary,
    fetchMonthlyTrends,
    fetchKPIAchievements,
    fetchWorkPlanRealizations,
    kpiBreakdown,
    kpiBreakdownPagination,
    kpiBreakdownLoading,
    kpiBreakdownError,
    kpiBreakdownQuery,
    kpiBreakdownFilters,
    fetchKpiBreakdown,
    loadKpiBreakdown,
    setKpiBreakdownFilters,
    setKpiBreakdownSearch,
    setKpiBreakdownPage,
    resetKpiBreakdownFilters
  };
});
