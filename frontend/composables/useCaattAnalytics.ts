import { ref, watch } from 'vue'
import { useActiveDataSource } from '~/composables/useActiveDataSource'
import { useDataSourceStore } from '~/stores/data-source'

export interface BenfordDigitRecord {
  digit: number
  actualCount: number
  expectedCount: number
  countDiff: number
  actualPct: number
  expectedPct: number
  deviationPct: number
  chiSquare: number
  chiSquareSharePct: number
  isSignificant: boolean
  dominantBranch: string
  dominantBranchCount: number
  dominantCategory: string
  dominantCategoryCount: number
  totalAmount: number
  avgAmount: number
  roundAmountCount: number
  afterHoursCount: number
  suspiciousCount: number
  direction: string
  findingOrigin: string
}

export interface BenfordTransactionFinding {
  trx_id: number
  trx_ref_number: string
  trx_date: string
  digit: number
  amount: number
  branch_name: string
  category: string
  customer_name: string
  channel: string
  is_round_amount: boolean
  is_after_hours: boolean
  is_suspicious: boolean
  authorization_status: string
}

const loading = ref(false)
const loadingDigitFindings = ref(false)
const runningCaatt = ref(false)
const error = ref('')
const dataHubOffline = ref(false)
const hasFetchedOnce = ref(false)
const lastFetchedAt = ref(0)
const lastFetchedSourceKey = ref('')
const activeSourceMeta = ref<{ sourceId: string | null; sourceName: string | null }>({
  sourceId: null,
  sourceName: null
})
let sourceWatcherRegistered = false

const fptState = ref<any>({
  records: [],
  summary: {
    totalTested: 0,
    totalViolations: 0,
    avgViolationRate: 0,
    branchesTested: 0,
    categoriesTested: 0
  }
})

const dupGapState = ref<any>({
  records: [],
  summary: {
    totalDuplicates: 0,
    totalGaps: 0,
    branchesAffected: 0
  }
})

const benfordState = ref<{
  records: BenfordDigitRecord[]
  findings: BenfordTransactionFinding[]
  summary: {
    totalDigitsAnalyzed: number
    sampleSize: number
    totalChiSquare: number
    criticalValue: number
    perDigitCriticalValue: number
    significantDeviations: number
    conclusion: string
    activeSourceId: string | null
    sourceName: string | null
    lowSampleWarning: boolean
    goldTransactionsCount: number
    silverRecordsCount: number
    dataHubOffline: boolean
    dataHubError: string
  }
}>({
  records: [],
  findings: [],
  summary: {
    totalDigitsAnalyzed: 0,
    sampleSize: 0,
    totalChiSquare: 0,
    criticalValue: 15.507,
    perDigitCriticalValue: 3.841,
    significantDeviations: 0,
    conclusion: '',
    activeSourceId: null,
    sourceName: null,
    lowSampleWarning: false,
    goldTransactionsCount: 0,
    silverRecordsCount: 0,
    dataHubOffline: false,
    dataHubError: ''
  }
})

const stratState = ref<any>({
  records: [],
  summary: {
    totalTransactions: 0,
    totalAmount: 0,
    totalStrata: 0
  }
})

const reconState = ref<any>({
  records: [],
  summary: {
    avgMatchRate: 0,
    totalUnmatched: 0,
    totalDifference: 0
  }
})

const policyState = ref<any>({
  records: [],
  summary: {
    totalViolations: 0,
    critical: 0,
    high: 0,
    medium: 0,
    branchesAffected: 0
  }
})

const dqState = ref<any>({
  records: [],
  summary: {
    avgCompleteness: 0,
    avgAccuracy: 0,
    avgTimelinessDays: 0,
    tablesProfiled: 0,
    totalRowsProfiled: 0
  },
  overallQualityScore: 0
})

export const useCaattAnalytics = () => {
  const config = useRuntimeConfig()
  const store = useDataSourceStore()
  const { selectedSourceId, sources } = useActiveDataSource()
  const getAnalyticsUrl = () => config.public.analyticsApiBase || '/api/analytics'

  const setSelectedSourceId = (id: string) => {
    store.setSelectedSourceId(id || 'all')
  }

  const refreshSources = async () => {
    await store.fetchSources()
  }

  const formatIDR = (val: number): string => {
    return 'Rp ' + Number(val || 0).toLocaleString('id-ID')
  }

  const buildEndpoint = (path: string): string => {
    const sid = selectedSourceId.value
    if (sid && sid !== 'all') {
      const sep = path.includes('?') ? '&' : '?'
      return `${path}${sep}source_id=${encodeURIComponent(sid)}`
    }
    return path
  }

  const safeApiFetch = async (endpoint: string, options: any = {}): Promise<any> => {
    const analyticsUrl = getAnalyticsUrl()
    try {
      return await $fetch(`${analyticsUrl}${buildEndpoint(endpoint)}`, options)
    } catch {
      return null
    }
  }

  const fetchCaattAnalytics = async (force = false) => {
    const srcKey = selectedSourceId.value || 'all'
    const isFresh =
      hasFetchedOnce.value &&
      !dataHubOffline.value &&
      lastFetchedSourceKey.value === srcKey &&
      Date.now() - lastFetchedAt.value < 2500
    if (!force && isFresh) return

    if (lastFetchedSourceKey.value !== srcKey) {
      benfordState.value.records = []
      benfordState.value.findings = []
    }

    if (store.sources.length === 0) {
      store.fetchSources().catch(() => {})
    }

    try {
      loading.value = true
      error.value = ''
      const responses: any[] = await Promise.all([
        safeApiFetch('/caatt/full-population'),
        safeApiFetch('/caatt/duplicate-gap'),
        safeApiFetch('/caatt/benford'),
        safeApiFetch('/caatt/stratification'),
        safeApiFetch('/caatt/reconciliation'),
        safeApiFetch('/caatt/policy-violations'),
        safeApiFetch('/caatt/data-quality')
      ])
      const [resFPT, resDupGap, resBenford, resStrat, resRecon, resPolicy, resDQ] = responses

      const allFailed = responses.every((r) => r === null)
      const anyOffline = allFailed || responses.some((r) => r?.summary?.data_hub_offline === true)
      dataHubOffline.value = anyOffline

      const firstMeta = responses.find((r) => r?.summary?.active_source_id)
      if (firstMeta?.summary) {
        activeSourceMeta.value = {
          sourceId: firstMeta.summary.active_source_id || null,
          sourceName: firstMeta.summary.source_name || null
        }
      } else if (srcKey !== 'all') {
        const matched = sources.value.find((s) => (s.id || s.source_id) === srcKey)
        activeSourceMeta.value = {
          sourceId: srcKey,
          sourceName: matched?.name || srcKey
        }
      } else {
        activeSourceMeta.value = { sourceId: null, sourceName: null }
      }

      if (Array.isArray(resFPT?.data)) {
        fptState.value.records = resFPT.data.map((r: any) => ({
          id: r.id || 'FPT',
          testDate: r.test_date || r.testDate || '-',
          branchName: r.branch_name || r.branchName || '-',
          accountId: r.account_id || r.accountId || '-',
          customerName: r.customer_name || r.customerName || '-',
          category: r.category || 'General',
          transactionAmount: r.transaction_amount ?? r.transactionAmount ?? 0,
          thresholdLimit: r.threshold_limit ?? r.thresholdLimit ?? 0,
          excessAmount: r.excess_amount ?? r.excessAmount ?? 0,
          violationType: r.violation_type || r.violationType || 'Rule Exception',
          status: r.status || 'FLAGGED'
        }))
        if (resFPT.summary) {
          fptState.value.summary = {
            totalTested: Number(resFPT.summary.total_tested ?? 0),
            totalViolations: Number(resFPT.summary.total_violations ?? 0),
            avgViolationRate: Number(resFPT.summary.avg_violation_rate ?? 0),
            branchesTested: Number(resFPT.summary.branches_tested ?? 0),
            categoriesTested: Number(resFPT.summary.categories_tested ?? 0)
          }
        }
      }

      if (Array.isArray(resDupGap?.data)) {
        dupGapState.value.records = resDupGap.data.map((r: any) => ({
          id: r.id || 'DG',
          testDate: r.test_date || r.testDate || '-',
          resultType: r.result_type || r.resultType || 'DUPLICATE',
          branchName: r.branch_name || r.branchName || '-',
          referenceNo: r.reference_no || r.referenceNo || '-',
          accountId: r.account_id || r.accountId || '-',
          amount: Number(r.amount ?? 0),
          description: r.description || '',
          occurrences: Number(r.occurrences ?? 1),
          status: r.status || 'OPEN'
        }))
        if (resDupGap.summary) {
          dupGapState.value.summary = {
            totalDuplicates: Number(resDupGap.summary.total_duplicates ?? 0),
            totalGaps: Number(resDupGap.summary.total_gaps ?? 0),
            branchesAffected: Number(resDupGap.summary.branches_affected ?? 0)
          }
        }
      }

      if (Array.isArray(resBenford?.data)) {
        benfordState.value.records = resBenford.data.map((r: any) => ({
          digit: Number(r.digit ?? 0),
          actualCount: Number(r.actual_count ?? r.actualCount ?? 0),
          expectedCount: Number(r.expected_count ?? r.expectedCount ?? 0),
          countDiff: Number(r.count_diff ?? r.countDiff ?? 0),
          actualPct: Number(r.actual_pct ?? r.actualPct ?? 0),
          expectedPct: Number(r.expected_pct ?? r.expectedPct ?? 0),
          deviationPct: Number(r.deviation_pct ?? r.deviationPct ?? 0),
          chiSquare: Number(r.chi_square ?? r.chiSquare ?? 0),
          chiSquareSharePct: Number(r.chi_square_share_pct ?? r.chiSquareSharePct ?? 0),
          isSignificant: Boolean(r.is_significant ?? r.isSignificant ?? false),
          dominantBranch: r.dominant_branch || r.dominantBranch || '-',
          dominantBranchCount: Number(r.dominant_branch_count ?? r.dominantBranchCount ?? 0),
          dominantCategory: r.dominant_category || r.dominantCategory || '-',
          dominantCategoryCount: Number(r.dominant_category_count ?? r.dominantCategoryCount ?? 0),
          totalAmount: Number(r.total_amount ?? r.totalAmount ?? 0),
          avgAmount: Number(r.avg_amount ?? r.avgAmount ?? 0),
          roundAmountCount: Number(r.round_amount_count ?? r.roundAmountCount ?? 0),
          afterHoursCount: Number(r.after_hours_count ?? r.afterHoursCount ?? 0),
          suspiciousCount: Number(r.suspicious_count ?? r.suspiciousCount ?? 0),
          direction: r.direction || 'balanced',
          findingOrigin: r.finding_origin || r.findingOrigin || ''
        }))
        benfordState.value.findings = Array.isArray(resBenford.findings) ? resBenford.findings : []
        if (resBenford.summary) {
          benfordState.value.summary = {
            totalDigitsAnalyzed: Number(resBenford.summary.total_digits_analyzed ?? 0),
            sampleSize: Number(resBenford.summary.sample_size ?? 0),
            totalChiSquare: Number(resBenford.summary.total_chi_square ?? 0),
            criticalValue: Number(resBenford.summary.critical_value ?? 15.507),
            perDigitCriticalValue: Number(resBenford.summary.per_digit_critical_value ?? 3.841),
            significantDeviations: Number(resBenford.summary.significant_deviations ?? 0),
            conclusion: resBenford.summary.conclusion || '',
            activeSourceId: resBenford.summary.active_source_id || null,
            sourceName: resBenford.summary.source_name || null,
            lowSampleWarning: Boolean(resBenford.summary.low_sample_warning),
            goldTransactionsCount: Number(resBenford.summary.gold_transactions_count ?? 0),
            silverRecordsCount: Number(resBenford.summary.silver_records_count ?? 0),
            dataHubOffline: Boolean(resBenford.summary.data_hub_offline),
            dataHubError: resBenford.summary.data_hub_error || ''
          }
        }
      } else {
        benfordState.value.records = []
        benfordState.value.findings = []
        benfordState.value.summary = {
          ...benfordState.value.summary,
          totalDigitsAnalyzed: 0,
          sampleSize: 0,
          totalChiSquare: 0,
          significantDeviations: 0,
          conclusion: anyOffline ? 'DATA_HUB_OFFLINE' : 'NO_DATA',
          goldTransactionsCount: 0,
          silverRecordsCount: 0,
          dataHubOffline: anyOffline
        }
      }

      if (Array.isArray(resStrat?.data)) {
        stratState.value.records = resStrat.data.map((r: any) => ({
          stratumLabel: r.stratum_label || r.stratumLabel || 'Strata',
          category: r.category || 'General',
          trxCount: Number(r.trx_count ?? r.trxCount ?? 0),
          totalAmount: Number(r.total_amount ?? r.totalAmount ?? 0),
          pctCount: Number(r.pct_count ?? r.pctCount ?? 0),
          pctAmount: Number(r.pct_amount ?? r.pctAmount ?? 0)
        }))
        if (resStrat.summary) {
          stratState.value.summary = {
            totalTransactions: Number(resStrat.summary.total_transactions ?? 0),
            totalAmount: Number(resStrat.summary.total_amount ?? 0),
            totalStrata: Number(resStrat.summary.total_strata ?? 0)
          }
        }
      }

      if (Array.isArray(resRecon?.data)) {
        reconState.value.records = resRecon.data.map((r: any) => ({
          module: r.module || 'Module',
          systemA: r.system_a || r.systemA || 'System A',
          systemB: r.system_b || r.systemB || 'System B',
          totalRecordsA: Number(r.total_records_a ?? r.totalRecordsA ?? 0),
          totalRecordsB: Number(r.total_records_b ?? r.totalRecordsB ?? 0),
          matchedRecords: Number(r.matched_records ?? r.matchedRecords ?? 0),
          matchRatePct: Number(r.match_rate_pct ?? r.matchRatePct ?? 0),
          totalDifference: Number(r.total_difference ?? r.totalDifference ?? 0),
          status: r.status || 'BALANCED'
        }))
        if (resRecon.summary) {
          reconState.value.summary = {
            avgMatchRate: Number(resRecon.summary.avg_match_rate ?? 0),
            totalUnmatched: Number(resRecon.summary.total_unmatched ?? 0),
            totalDifference: Number(resRecon.summary.total_difference ?? 0)
          }
        }
      }

      if (Array.isArray(resPolicy?.data)) {
        policyState.value.records = resPolicy.data.map((r: any) => ({
          id: r.id || 'POL',
          testDate: r.test_date || r.testDate || '-',
          ruleName: r.rule_name || r.ruleName || 'Policy Rule',
          branchName: r.branch_name || r.branchName || '-',
          customerName: r.customer_name || r.customerName || '-',
          severity: r.severity || 'Medium',
          description: r.description || '',
          status: r.status || 'OPEN'
        }))
        if (resPolicy.summary) {
          policyState.value.summary = {
            totalViolations: Number(resPolicy.summary.total_violations ?? 0),
            critical: Number(resPolicy.summary.critical ?? 0),
            high: Number(resPolicy.summary.high ?? 0),
            medium: Number(resPolicy.summary.medium ?? 0),
            branchesAffected: Number(resPolicy.summary.branches_affected ?? 0)
          }
        }
      }

      if (Array.isArray(resDQ?.data)) {
        dqState.value.records = resDQ.data.map((r: any) => ({
          tableName: r.table_name || r.tableName || 'table',
          zone: r.zone || 'Gold',
          totalRows: Number(r.total_rows ?? r.totalRows ?? 0),
          completenessPct: Number(r.completeness_pct ?? r.completenessPct ?? 0),
          accuracyPct: Number(r.accuracy_pct ?? r.accuracyPct ?? 0),
          timelinessDays: Number(r.timeliness_days ?? r.timelinessDays ?? 0),
          duplicateCount: Number(r.duplicate_count ?? r.duplicateCount ?? 0),
          status: r.status || 'GOOD'
        }))
        if (resDQ.summary) {
          dqState.value.summary = {
            avgCompleteness: Number(resDQ.summary.avg_completeness ?? 0),
            avgAccuracy: Number(resDQ.summary.avg_accuracy ?? 0),
            avgTimelinessDays: Number(resDQ.summary.avg_timeliness_days ?? 0),
            tablesProfiled: Number(resDQ.summary.tables_profiled ?? 0),
            totalRowsProfiled: Number(resDQ.summary.total_rows_profiled ?? 0)
          }
        }
        dqState.value.overallQualityScore = Number(resDQ.overall_quality_score ?? 0)
      }

      lastFetchedAt.value = Date.now()
      lastFetchedSourceKey.value = srcKey
      hasFetchedOnce.value = !anyOffline
    } catch (err: any) {
      error.value = err.message || 'Error loading CAATT analytics'
    } finally {
      loading.value = false
    }
  }

  const fetchBenfordFindings = async (digit: number | null = null, findingsLimit = 100) => {
    loadingDigitFindings.value = true
    try {
      const params = new URLSearchParams()
      params.set('findings_limit', String(findingsLimit))
      if (digit !== null && digit >= 1 && digit <= 9) {
        params.set('digit', String(digit))
      }
      const res: any = await safeApiFetch(`/caatt/benford?${params.toString()}`)
      if (res && Array.isArray(res.findings)) {
        benfordState.value.findings = res.findings
        if (res.summary) {
          benfordState.value.summary.goldTransactionsCount = Number(
            res.summary.gold_transactions_count ?? benfordState.value.summary.goldTransactionsCount ?? 0
          )
          benfordState.value.summary.silverRecordsCount = Number(
            res.summary.silver_records_count ?? benfordState.value.summary.silverRecordsCount ?? 0
          )
        }
      }
    } finally {
      loadingDigitFindings.value = false
    }
  }

  const triggerCaattRun = async () => {
    runningCaatt.value = true
    try {
      const analyticsUrl = getAnalyticsUrl()
      await $fetch(`${analyticsUrl}${buildEndpoint('/caatt/run')}`, { method: 'POST' })
      await fetchCaattAnalytics(true)
    } catch (err: any) {
      error.value = err?.data?.error || err?.message || 'Failed to trigger CAATT Engine'
    } finally {
      runningCaatt.value = false
    }
  }

  if (!sourceWatcherRegistered) {
    sourceWatcherRegistered = true
    watch(selectedSourceId, () => {
      fetchCaattAnalytics(true)
    })
  }

  return {
    loading,
    loadingDigitFindings,
    runningCaatt,
    error,
    dataHubOffline,
    selectedSourceId,
    activeSourceMeta,
    sources,
    setSelectedSourceId,
    refreshSources,
    fptState,
    dupGapState,
    benfordState,
    stratState,
    reconState,
    policyState,
    dqState,
    formatIDR,
    fetchCaattAnalytics,
    fetchBenfordFindings,
    triggerCaattRun
  }
}
