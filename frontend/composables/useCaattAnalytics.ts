import { ref } from 'vue'
import {
  useCAATTFullPopulationData,
  useCAATTDuplicateGapData,
  useCAATTBenfordData,
  useCAATTStratificationData,
  useCAATTReconciliationData,
  useCAATTPolicyData,
  useCAATTDataQualityData
} from '~/composables/useAnalyticsData'

const initialFPT = useCAATTFullPopulationData()
const initialDupGap = useCAATTDuplicateGapData()
const initialBenford = useCAATTBenfordData()
const initialStrat = useCAATTStratificationData()
const initialRecon = useCAATTReconciliationData()
const initialPolicy = useCAATTPolicyData()
const initialDQ = useCAATTDataQualityData()

const loading = ref(false)
const error = ref('')
const hasFetchedOnce = ref(false)

const fptState = ref<any>({
  records: [...initialFPT.records],
  summary: { ...initialFPT.summary }
})

const dupGapState = ref<any>({
  records: [...initialDupGap.records],
  summary: { ...initialDupGap.summary }
})

const benfordState = ref<any>({
  records: [...initialBenford.records],
  summary: { ...initialBenford.summary }
})

const stratState = ref<any>({
  records: [...initialStrat.records],
  summary: { ...initialStrat.summary }
})

const reconState = ref<any>({
  records: [...initialRecon.records],
  summary: { ...initialRecon.summary }
})

const policyState = ref<any>({
  records: [...initialPolicy.records],
  summary: { ...initialPolicy.summary }
})

const dqState = ref<any>({
  records: [...initialDQ.records],
  summary: { ...initialDQ.summary },
  overallQualityScore: initialDQ.overallQualityScore
})

export const useCaattAnalytics = () => {
  const config = useRuntimeConfig()
  const getAnalyticsUrl = () => config.public.analyticsApiBase || '/api/analytics'

  const formatIDR = (val: number): string => {
    return 'Rp ' + Number(val || 0).toLocaleString('id-ID')
  }

  const safeApiFetch = async (endpoint: string, options: any = {}): Promise<any> => {
    const analyticsUrl = getAnalyticsUrl()
    try {
      return await $fetch(`${analyticsUrl}${endpoint}`, options)
    } catch (e) {
      return null
    }
  }

  const fetchCaattAnalytics = async (force = false) => {
    if (hasFetchedOnce.value && !force) return

    try {
      loading.value = true
      const [resFPT, resDupGap, resBenford, resStrat, resRecon, resPolicy, resDQ]: any[] = await Promise.all([
        safeApiFetch('/caatt/full-population').catch(() => null),
        safeApiFetch('/caatt/duplicate-gap').catch(() => null),
        safeApiFetch('/caatt/benford').catch(() => null),
        safeApiFetch('/caatt/stratification').catch(() => null),
        safeApiFetch('/caatt/reconciliation').catch(() => null),
        safeApiFetch('/caatt/policy-violations').catch(() => null),
        safeApiFetch('/caatt/data-quality').catch(() => null)
      ])

      if (resFPT?.data && resFPT.data.length > 0) {
        fptState.value.records = resFPT.data.map((r: any) => ({
          id: r.id || 'FPT',
          testDate: r.test_date || r.testDate || '2026-06-01',
          branchName: r.branch_name || r.branchName || '-',
          accountId: r.account_id || r.accountId || '-',
          customerName: r.customer_name || r.customerName || '-',
          category: r.category || 'General',
          transactionAmount: r.transaction_amount || r.transactionAmount || 0,
          thresholdLimit: r.threshold_limit || r.thresholdLimit || 0,
          excessAmount: r.excess_amount || r.excessAmount || 0,
          violationType: r.violation_type || r.violationType || 'Plafond Exceeded',
          status: r.status || 'FLAGGED'
        }))
        if (resFPT.summary) fptState.value.summary = {
          totalTested: resFPT.summary.total_tested ?? fptState.value.summary.totalTested,
          totalViolations: resFPT.summary.total_violations ?? fptState.value.summary.totalViolations,
          avgViolationRate: resFPT.summary.avg_violation_rate ?? fptState.value.summary.avgViolationRate,
          branchesTested: resFPT.summary.branches_tested ?? fptState.value.summary.branchesTested,
          categoriesTested: resFPT.summary.categories_tested ?? fptState.value.summary.categoriesTested
        }
      }

      if (resDupGap?.data && resDupGap.data.length > 0) {
        dupGapState.value.records = resDupGap.data.map((r: any) => ({
          id: r.id || 'DG',
          testDate: r.test_date || r.testDate || '2026-06-01',
          resultType: r.result_type || r.resultType || 'DUPLICATE',
          branchName: r.branch_name || r.branchName || '-',
          referenceNo: r.reference_no || r.referenceNo || '-',
          accountId: r.account_id || r.accountId || '-',
          amount: r.amount ?? 0,
          description: r.description || '',
          occurrences: r.occurrences || 2,
          status: r.status || 'OPEN'
        }))
        if (resDupGap.summary) dupGapState.value.summary = {
          totalDuplicates: resDupGap.summary.total_duplicates ?? dupGapState.value.summary.totalDuplicates,
          totalGaps: resDupGap.summary.total_gaps ?? dupGapState.value.summary.totalGaps,
          branchesAffected: resDupGap.summary.branches_affected ?? dupGapState.value.summary.branchesAffected
        }
      }

      if (resBenford?.data && resBenford.data.length > 0) {
        benfordState.value.records = resBenford.data.map((r: any) => ({
          digit: r.digit,
          actualCount: r.actual_count ?? r.actualCount ?? 0,
          actualPct: r.actual_pct ?? r.actualPct ?? 0,
          expectedPct: r.expected_pct ?? r.expectedPct ?? 0,
          deviationPct: r.deviation_pct ?? r.deviationPct ?? 0,
          isSignificant: r.is_significant ?? r.isSignificant ?? false
        }))
        if (resBenford.summary) benfordState.value.summary = {
          totalDigitsAnalyzed: resBenford.summary.total_digits_analyzed ?? 9,
          significantDeviations: resBenford.summary.significant_deviations ?? 0,
          conclusion: resBenford.summary.conclusion ?? ''
        }
      }

      if (resStrat?.data && resStrat.data.length > 0) {
        stratState.value.records = resStrat.data.map((r: any) => ({
          stratumLabel: r.stratum_label || r.stratumLabel || 'Strata',
          category: r.category || 'Retail',
          trxCount: r.trx_count ?? r.trxCount ?? 0,
          totalAmount: r.total_amount ?? r.totalAmount ?? 0,
          pctCount: r.pct_count ?? r.pctCount ?? 0,
          pctAmount: r.pct_amount ?? r.pctAmount ?? 0
        }))
        if (resStrat.summary) stratState.value.summary = {
          totalTransactions: resStrat.summary.total_transactions ?? stratState.value.summary.totalTransactions,
          totalAmount: resStrat.summary.total_amount ?? stratState.value.summary.totalAmount,
          totalStrata: resStrat.summary.total_strata ?? stratState.value.summary.totalStrata
        }
      }

      if (resRecon?.data && resRecon.data.length > 0) {
        reconState.value.records = resRecon.data.map((r: any) => ({
          module: r.module || 'Module',
          systemA: r.system_a || r.systemA || 'System A',
          systemB: r.system_b || r.systemB || 'System B',
          totalRecordsA: r.total_records_a ?? r.totalRecordsA ?? 0,
          totalRecordsB: r.total_records_b ?? r.totalRecordsB ?? 0,
          matchedRecords: r.matched_records ?? r.matchedRecords ?? 0,
          matchRatePct: r.match_rate_pct ?? r.matchRatePct ?? 100,
          totalDifference: r.total_difference ?? r.totalDifference ?? 0,
          status: r.status || 'BALANCED'
        }))
        if (resRecon.summary) reconState.value.summary = {
          avgMatchRate: resRecon.summary.avg_match_rate ?? reconState.value.summary.avgMatchRate,
          totalUnmatched: resRecon.summary.total_unmatched ?? reconState.value.summary.totalUnmatched,
          totalDifference: resRecon.summary.total_difference ?? reconState.value.summary.totalDifference
        }
      }

      if (resPolicy?.data && resPolicy.data.length > 0) {
        policyState.value.records = resPolicy.data.map((r: any) => ({
          id: r.id || 'POL',
          testDate: r.test_date || r.testDate || '2026-06-01',
          ruleName: r.rule_name || r.ruleName || 'Policy Rule',
          branchName: r.branch_name || r.branchName || '-',
          customerName: r.customer_name || r.customerName || '-',
          severity: r.severity || 'Medium',
          description: r.description || '',
          status: r.status || 'OPEN'
        }))
        if (resPolicy.summary) policyState.value.summary = {
          totalViolations: resPolicy.summary.total_violations ?? policyState.value.summary.totalViolations,
          critical: resPolicy.summary.critical ?? policyState.value.summary.critical,
          high: resPolicy.summary.high ?? policyState.value.summary.high,
          medium: resPolicy.summary.medium ?? policyState.value.summary.medium,
          branchesAffected: resPolicy.summary.branches_affected ?? policyState.value.summary.branchesAffected
        }
      }

      if (resDQ?.data && resDQ.data.length > 0) {
        dqState.value.records = resDQ.data.map((r: any) => ({
          tableName: r.table_name || r.tableName || 'table',
          zone: r.zone || 'Gold',
          totalRows: r.total_rows ?? r.totalRows ?? 0,
          completenessPct: r.completeness_pct ?? r.completenessPct ?? 100,
          accuracyPct: r.accuracy_pct ?? r.accuracyPct ?? 100,
          timelinessDays: r.timeliness_days ?? r.timelinessDays ?? 0,
          duplicateCount: r.duplicate_count ?? r.duplicateCount ?? 0,
          status: r.status || 'GOOD'
        }))
        if (resDQ.summary) dqState.value.summary = {
          avgCompleteness: resDQ.summary.avg_completeness ?? dqState.value.summary.avgCompleteness,
          avgAccuracy: resDQ.summary.avg_accuracy ?? dqState.value.summary.avgAccuracy,
          avgTimelinessDays: resDQ.summary.avg_timeliness_days ?? dqState.value.summary.avgTimelinessDays,
          tablesProfiled: resDQ.summary.tables_profiled ?? dqState.value.summary.tablesProfiled,
          totalRowsProfiled: resDQ.summary.total_rows_profiled ?? dqState.value.summary.totalRowsProfiled
        }
        if (resDQ.overall_quality_score) dqState.value.overallQualityScore = resDQ.overall_quality_score
      }

      hasFetchedOnce.value = true
    } catch (err: any) {
      error.value = err.message || 'Error loading CAATT analytics'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    fptState,
    dupGapState,
    benfordState,
    stratState,
    reconState,
    policyState,
    dqState,
    formatIDR,
    fetchCaattAnalytics
  }
}
