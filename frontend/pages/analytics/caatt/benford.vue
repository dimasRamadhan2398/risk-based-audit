<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  BarController,
  LineElement,
  PointElement,
  LineController,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { useCaattAnalytics, type BenfordDigitRecord } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  BarController,
  LineElement,
  PointElement,
  LineController,
  Title,
  Tooltip,
  Legend
)

const { t } = useI18n()
const {
  loading,
  loadingDigitFindings,
  runningCaatt,
  dataHubOffline,
  benfordState,
  formatIDR,
  fetchCaattAnalytics,
  fetchBenfordFindings,
  triggerCaattRun
} = useCaattAnalytics()

const selectedDigit = ref<number | null>(null)
const searchQuery = ref('')
const showMethodology = ref(false)
const showGuide = ref(true)

onMounted(() => {
  fetchCaattAnalytics()
})

const isOffline = computed(() => dataHubOffline.value || benfordState.value.summary.dataHubOffline)
const hasRecords = computed(() => (benfordState.value.records?.length || 0) > 0)
const hasSilverWithoutGold = computed(() =>
  !hasRecords.value &&
  Number(benfordState.value.summary.silverRecordsCount || 0) > 0 &&
  Number(benfordState.value.summary.goldTransactionsCount || 0) === 0
)

const hasStaleBackendSchema = computed(() => {
  const recs = benfordState.value.records || []
  if (recs.length === 0 || Number(benfordState.value.summary.sampleSize || 0) <= 0) return false
  return recs.every((r) => Number(r.expectedCount || 0) === 0)
})

const handleSelectDigit = async (digit: number | null) => {
  const nextDigit = selectedDigit.value === digit ? null : digit
  selectedDigit.value = nextDigit
  await fetchBenfordFindings(nextDigit, 100)
}

const topRiskDigitRecord = computed<BenfordDigitRecord | null>(() => {
  const recs = benfordState.value.records || []
  if (recs.length === 0) return null
  return [...recs].sort((a, b) => b.chiSquare - a.chiSquare)[0] || null
})

const executiveExposureAmount = computed(() => {
  const recs = benfordState.value.records || []
  if (selectedDigit.value !== null) {
    const match = recs.find((r) => r.digit === selectedDigit.value)
    return Number(match?.totalAmount || 0)
  }
  const sig = recs.filter((r) => r.isSignificant)
  const target = sig.length > 0 ? sig : recs
  return target.reduce((sum, r) => sum + Number(r.totalAmount || 0), 0)
})

const executiveRedFlags = computed(() => {
  const recs = benfordState.value.records || []
  const sig = recs.filter((r) => r.isSignificant)
  const target = sig.length > 0 ? sig : recs
  return target.reduce(
    (acc, r) => ({
      round: acc.round + Number(r.roundAmountCount || 0),
      afterHours: acc.afterHours + Number(r.afterHoursCount || 0),
      suspicious: acc.suspicious + Number(r.suspiciousCount || 0)
    }),
    { round: 0, afterHours: 0, suspicious: 0 }
  )
})

const hasIncompleteOriginDetails = computed(() => {
  const recs = benfordState.value.records || []
  if (recs.length === 0) return false
  return recs.every((r) => (!r.findingOrigin || r.findingOrigin === '-') && (r.dominantBranch === '-' || !r.dominantBranch))
})

const verdictTheme = computed(() => {
  const sigCount = benfordState.value.summary.significantDeviations
  const overCrit = benfordState.value.summary.totalChiSquare > benfordState.value.summary.criticalValue
  if (overCrit || sigCount >= 2) {
    return {
      color: 'error' as const,
      borderClass: 'border-l-4 border-l-red-500 border-red-200 dark:border-red-900/60 bg-red-50/40 dark:bg-red-950/20',
      icon: 'i-lucide-shield-alert',
      iconColor: 'text-red-600 dark:text-red-400',
      headline: t('analytics.caattUi.benford.deviatedTitle'),
      narrative: t('analytics.caattUi.benford.execVerdictAnomaly', { count: Math.max(sigCount, 1) })
    }
  }
  if (sigCount === 1) {
    return {
      color: 'warning' as const,
      borderClass: 'border-l-4 border-l-amber-500 border-amber-200 dark:border-amber-900/60 bg-amber-50/40 dark:bg-amber-950/20',
      icon: 'i-lucide-alert-triangle',
      iconColor: 'text-amber-600 dark:text-amber-400',
      headline: t('analytics.caattUi.benford.deviatedTitle'),
      narrative: t('analytics.caattUi.benford.execVerdictMinor')
    }
  }
  return {
    color: 'success' as const,
    borderClass: 'border-l-4 border-l-emerald-500 border-emerald-200 dark:border-emerald-900/60 bg-emerald-50/40 dark:bg-emerald-950/20',
    icon: 'i-lucide-shield-check',
    iconColor: 'text-emerald-600 dark:text-emerald-400',
    headline: t('analytics.caattUi.benford.normalTitle'),
    narrative: t('analytics.caattUi.benford.execVerdictNormal')
  }
})

const filteredOriginRecords = computed(() => {
  const recs = benfordState.value.records || []
  if (selectedDigit.value === null) return recs
  return recs.filter((r) => r.digit === selectedDigit.value)
})

const filteredFindings = computed(() => {
  let list = benfordState.value.findings || []
  if (selectedDigit.value !== null) {
    list = list.filter((f) => Number(f.digit) === selectedDigit.value)
  }
  const q = searchQuery.value.trim().toLowerCase()
  if (q) {
    list = list.filter((f) =>
      [f.trx_ref_number, f.branch_name, f.customer_name, f.category, f.channel]
        .some((v) => String(v || '').toLowerCase().includes(q))
    )
  }
  return list
})

const formatDirectionLabel = (r: BenfordDigitRecord) => {
  const pct = Math.abs(Number(r.deviationPct || 0)).toFixed(1)
  if (r.deviationPct > 0.5) return t('analytics.caattUi.benford.directionOver', { pct })
  if (r.deviationPct < -0.5) return t('analytics.caattUi.benford.directionUnder', { pct: `-${pct}` })
  return t('analytics.caattUi.benford.directionBalanced')
}

const formatCountDiffText = (r: BenfordDigitRecord) => {
  const diff = Math.round(Number(r.countDiff || 0))
  if (diff > 0) return t('analytics.caattUi.benford.diffMore', { count: diff.toLocaleString('id-ID') })
  if (diff < 0) return t('analytics.caattUi.benford.diffLess', { count: Math.abs(diff).toLocaleString('id-ID') })
  return t('analytics.caattUi.benford.diffBalanced')
}

const getOriginDescription = (r: BenfordDigitRecord) => {
  if (r.findingOrigin && r.findingOrigin !== '-') return r.findingOrigin
  if (hasStaleBackendSchema.value || hasIncompleteOriginDetails.value) {
    return `${formatCountDiffText(r)} (${formatDirectionLabel(r)}).`
  }
  return '-'
}

const formatHighlightedAmount = (amount: number) => {
  const formatted = Math.round(Number(amount || 0)).toLocaleString('id-ID')
  if (!formatted || formatted === '0') {
    return { first: '0', rest: '' }
  }
  return {
    first: formatted.charAt(0),
    rest: formatted.slice(1)
  }
}

const benfordChartData = computed(() => ({
  labels: benfordState.value.records.map((r) => `#${r.digit}`),
  datasets: [
    {
      label: t('analytics.caattUi.benford.actualFrequency'),
      type: 'bar' as const,
      backgroundColor: benfordState.value.records.map((r) => {
        if (selectedDigit.value === r.digit) return 'rgba(79, 70, 229, 0.92)'
        return r.isSignificant ? 'rgba(239, 68, 68, 0.85)' : 'rgba(59, 130, 246, 0.80)'
      }),
      borderRadius: 6,
      data: benfordState.value.records.map((r) => r.actualPct),
      order: 2
    },
    {
      label: t('analytics.caattUi.benford.benfordCurve'),
      type: 'line' as const,
      borderColor: '#f59e0b',
      backgroundColor: '#f59e0b',
      borderWidth: 2.5,
      pointRadius: 4,
      pointHoverRadius: 6,
      tension: 0.25,
      fill: false,
      data: benfordState.value.records.map((r) => r.expectedPct),
      order: 1
    } as any
  ]
}))

const benfordChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    mode: 'index' as const,
    intersect: false
  },
  plugins: {
    legend: {
      display: true,
      position: 'top' as const
    },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `${ctx.dataset.label}: ${Number(ctx.parsed.y || 0).toFixed(2)}%`
      }
    }
  },
  scales: {
    y: {
      beginAtZero: true,
      title: {
        display: true,
        text: t('analytics.caattUi.benford.frequencyDistribution')
      }
    },
    x: {
      title: {
        display: true,
        text: t('analytics.caattUi.benford.digit')
      }
    }
  }
}))
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8 space-y-6">
    <CaattHeader
      :title="t('navigation.caattBenford')"
      :subtitle="t('analytics.caattUi.benford.subtitle')"
      icon="i-lucide-calculator"
      :badgeText="t('analytics.caattUi.benford.badge')"
      badgeColor="info"
    />

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-arrow-path" class="w-12 h-12 animate-spin text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Data Hub Offline Recovery Card -->
    <div v-else-if="isOffline">
      <UCard>
        <div class="flex flex-col items-center justify-center py-10 text-center space-y-4">
          <UIcon name="i-lucide-server-off" class="w-12 h-12 text-red-400" />
          <div class="max-w-lg space-y-1.5">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('analytics.caattUi.benford.dataHubOfflineTitle') }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ benfordState.summary.dataHubError || t('analytics.caattUi.benford.dataHubOfflineDesc') }}
            </p>
          </div>
          <div class="flex flex-wrap justify-center gap-3">
            <UButton color="primary" icon="i-lucide-refresh-cw" @click="fetchCaattAnalytics(true)">
              {{ t('analytics.caattUi.common.refresh') }}
            </UButton>
            <UButton to="/settings?tab=datasource" color="neutral" variant="outline" icon="i-lucide-database">
              {{ t('analytics.caattUi.benford.goToDataSource') }}
            </UButton>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Empty State (Differentiates synced Silver with no amounts vs never synced) -->
    <div v-else-if="!hasRecords">
      <UCard>
        <div class="flex flex-col items-center justify-center py-12 text-center space-y-4">
          <UIcon :name="hasSilverWithoutGold ? 'i-lucide-file-warning' : 'i-lucide-database-zap'" class="w-12 h-12 text-amber-500" />
          <div class="max-w-lg space-y-1.5">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">
              {{
                hasSilverWithoutGold
                  ? t('analytics.caattUi.benford.silverOnlyWarningTitle', { silver: benfordState.summary.silverRecordsCount.toLocaleString('id-ID') })
                  : t('analytics.caattUi.benford.noDataTitle')
              }}
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{
                hasSilverWithoutGold
                  ? t('analytics.caattUi.benford.silverOnlyWarningDesc')
                  : t('analytics.caattUi.benford.noDataDesc')
              }}
            </p>
          </div>
          <div class="flex flex-wrap justify-center gap-3">
            <UButton to="/settings?tab=datasource" color="primary" icon="i-lucide-database">
              {{ t('analytics.caattUi.benford.goToDataSource') }}
            </UButton>
            <UButton
              color="neutral"
              variant="outline"
              icon="i-lucide-play"
              :loading="runningCaatt"
              @click="triggerCaattRun()"
            >
              {{ t('analytics.caattUi.benford.runAnalysis') }}
            </UButton>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Plain-Language Executive Forensic View -->
    <div v-else class="space-y-6">
      <!-- Diagnostic Warning if server still returns pre-rebuild 0 expected_count -->
      <UAlert
        v-if="hasStaleBackendSchema"
        icon="i-lucide-alert-circle"
        color="warning"
        variant="subtle"
        :title="t('analytics.caattUi.benford.staleSchemaBanner')"
      />

      <!-- 1. Plain-Language Audit Summary Hero + Beginner Guide -->
      <div
        class="rounded-2xl border p-5 sm:p-6 shadow-xs transition-all"
        :class="verdictTheme.borderClass"
      >
        <div class="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
          <div class="flex items-start gap-3.5">
            <div class="p-2.5 rounded-xl bg-white/90 dark:bg-gray-900/80 shadow-xs shrink-0">
              <UIcon :name="verdictTheme.icon" class="w-6 h-6" :class="verdictTheme.iconColor" />
            </div>
            <div class="space-y-1.5">
              <div class="flex flex-wrap items-center gap-2">
                <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ t('analytics.caattUi.benford.execVerdictTitle') }}
                </span>
                <UBadge :color="verdictTheme.color" variant="solid" size="xs">
                  {{ verdictTheme.headline }}
                </UBadge>
              </div>
              <p class="text-sm sm:text-base font-medium text-gray-900 dark:text-white leading-relaxed">
                {{ verdictTheme.narrative }}
              </p>
            </div>
          </div>

          <div class="flex items-center gap-2 shrink-0">
            <UButton
              size="xs"
              color="neutral"
              variant="ghost"
              :icon="showGuide ? 'i-lucide-chevron-up' : 'i-lucide-help-circle'"
              @click="showGuide = !showGuide"
            >
              {{ t('analytics.caattUi.benford.plainGuideTitle') }}
            </UButton>
            <UBadge v-if="benfordState.summary.lowSampleWarning" color="warning" variant="subtle" size="sm">
              <UIcon name="i-lucide-alert-triangle" class="w-3.5 h-3.5 mr-1" />
              {{ benfordState.summary.sampleSize }} {{ t('analytics.caattUi.benford.trxUnit') }}
            </UBadge>
          </div>
        </div>

        <!-- Quick Everyday Guide for Non-Technical Readers -->
        <div
          v-if="showGuide"
          class="mt-4 rounded-xl bg-white/85 dark:bg-gray-900/75 border border-gray-200/80 dark:border-gray-800 p-3.5 sm:p-4 text-xs sm:text-sm text-gray-700 dark:text-gray-300 space-y-2"
        >
          <p class="leading-relaxed">
            {{ t('analytics.caattUi.benford.plainGuideText') }}
          </p>
          <div class="flex flex-wrap items-center gap-4 pt-1 text-xs font-medium">
            <span class="inline-flex items-center gap-1.5">
              <span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span>
              {{ t('analytics.caattUi.benford.plainGuideBarBlue') }}
            </span>
            <span class="inline-flex items-center gap-1.5">
              <span class="w-2.5 h-2.5 rounded-full bg-red-500"></span>
              {{ t('analytics.caattUi.benford.plainGuideBarRed') }}
            </span>
            <span class="inline-flex items-center gap-1.5">
              <span class="w-4 h-0.5 bg-amber-500"></span>
              {{ t('analytics.caattUi.benford.plainGuideLineOrange') }}
            </span>
          </div>
        </div>

        <!-- 3 Executive Takeaways Strip -->
        <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-3.5">
          <div class="rounded-xl bg-white/90 dark:bg-gray-900/80 border border-gray-200/70 dark:border-gray-800 p-3.5 flex items-center justify-between gap-3">
            <div class="min-w-0">
              <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
                {{ t('analytics.caattUi.benford.execTopDigit') }}
              </div>
              <div v-if="topRiskDigitRecord" class="mt-0.5 text-sm font-bold text-gray-900 dark:text-white truncate">
                {{ t('analytics.caattUi.benford.startsWithDigit', { digit: topRiskDigitRecord.digit }) }}
                <span class="font-normal text-gray-500">
                  · {{ topRiskDigitRecord.chiSquareSharePct.toFixed(0) }}% {{ t('analytics.caattUi.benford.shareOfRisk') }}
                </span>
              </div>
            </div>
            <UButton
              v-if="topRiskDigitRecord"
              size="xs"
              color="primary"
              variant="soft"
              @click="handleSelectDigit(topRiskDigitRecord.digit)"
            >
              #{{ topRiskDigitRecord.digit }}
            </UButton>
          </div>

          <div class="rounded-xl bg-white/90 dark:bg-gray-900/80 border border-gray-200/70 dark:border-gray-800 p-3.5">
            <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('analytics.caattUi.benford.execHotspot') }}
            </div>
            <div v-if="topRiskDigitRecord" class="mt-0.5 text-sm font-bold text-gray-900 dark:text-white truncate">
              {{ topRiskDigitRecord.dominantBranch || '-' }}
              <span v-if="topRiskDigitRecord.dominantCategory && topRiskDigitRecord.dominantCategory !== '-'" class="font-normal text-gray-500">
                · {{ topRiskDigitRecord.dominantCategory }}
              </span>
            </div>
          </div>

          <div class="rounded-xl bg-white/90 dark:bg-gray-900/80 border border-gray-200/70 dark:border-gray-800 p-3.5">
            <div class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('analytics.caattUi.benford.execRedFlags') }}
            </div>
            <div class="mt-1 flex flex-wrap gap-1.5">
              <UBadge color="warning" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.roundAmounts') }}: {{ executiveRedFlags.round.toLocaleString('id-ID') }}
              </UBadge>
              <UBadge color="info" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.afterHours') }}: {{ executiveRedFlags.afterHours.toLocaleString('id-ID') }}
              </UBadge>
              <UBadge :color="executiveRedFlags.suspicious > 0 ? 'error' : 'neutral'" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.flagged') }}: {{ executiveRedFlags.suspicious.toLocaleString('id-ID') }}
              </UBadge>
            </div>
          </div>
        </div>
      </div>

      <UAlert
        v-if="benfordState.summary.lowSampleWarning"
        icon="i-lucide-alert-triangle"
        color="warning"
        variant="subtle"
        :title="t('analytics.caattUi.benford.lowSampleWarning', { count: benfordState.summary.sampleSize })"
      />

      <!-- 2. Clear Non-Technical KPI Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
        <UCard>
          <div class="text-xs font-semibold text-gray-500 dark:text-gray-400">
            {{ t('analytics.caattUi.benford.kpiSampleSize') }}
          </div>
          <div class="mt-1.5 text-2xl font-extrabold text-gray-900 dark:text-white tabular-nums">
            {{ benfordState.summary.sampleSize.toLocaleString('id-ID') }}
          </div>
          <div class="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
            {{ benfordState.summary.sourceName || benfordState.summary.activeSourceId || t('analytics.caattUi.benford.defaultSource') }}
          </div>
        </UCard>

        <UCard>
          <div class="text-xs font-semibold text-gray-500 dark:text-gray-400">
            {{ t('analytics.caattUi.benford.kpiTotalChiSquare') }}
          </div>
          <div class="mt-1.5 flex items-baseline gap-2">
            <span
              class="text-2xl font-extrabold tabular-nums"
              :class="benfordState.summary.totalChiSquare > benfordState.summary.criticalValue ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'"
            >
              {{ benfordState.summary.totalChiSquare.toFixed(1) }}
            </span>
          </div>
          <div class="mt-1 text-xs" :class="benfordState.summary.totalChiSquare > benfordState.summary.criticalValue ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'">
            {{
              benfordState.summary.totalChiSquare > benfordState.summary.criticalValue
                ? t('analytics.caattUi.benford.execExceedsThreshold', { threshold: benfordState.summary.criticalValue.toFixed(1) })
                : t('analytics.caattUi.benford.execPassThreshold', { threshold: benfordState.summary.criticalValue.toFixed(1) })
            }}
          </div>
        </UCard>

        <UCard>
          <div class="text-xs font-semibold text-gray-500 dark:text-gray-400">
            {{ t('analytics.caattUi.benford.kpiSignificantDigits') }}
          </div>
          <div
            class="mt-1.5 text-2xl font-extrabold tabular-nums"
            :class="benfordState.summary.significantDeviations > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'"
          >
            {{ benfordState.summary.significantDeviations }} / 9
          </div>
          <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{
              benfordState.summary.significantDeviations > 0 && topRiskDigitRecord
                ? `${t('analytics.caattUi.benford.execTopDigit')}: #${topRiskDigitRecord.digit}`
                : t('analytics.caattUi.benford.statusNormal')
            }}
          </div>
        </UCard>

        <UCard>
          <div class="text-xs font-semibold text-gray-500 dark:text-gray-400">
            {{ t('analytics.caattUi.benford.execTotalExposure') }}
          </div>
          <div class="mt-1.5 text-xl sm:text-2xl font-extrabold text-gray-900 dark:text-white tabular-nums truncate" :title="formatIDR(executiveExposureAmount)">
            {{ formatIDR(executiveExposureAmount) }}
          </div>
          <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ selectedDigit !== null ? `${t('analytics.caattUi.benford.filterByDigit')}: #${selectedDigit}` : t('analytics.caattUi.benford.allDigits') }}
          </div>
        </UCard>
      </div>

      <!-- 3. Pattern Chart & Plain-Language Leading Digit Summary Table (1-9) -->
      <div class="grid grid-cols-1 xl:grid-cols-12 gap-6">
        <UCard class="xl:col-span-5">
          <template #header>
            <div class="flex items-center justify-between gap-2">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('analytics.caattUi.benford.comparison') }}
              </h3>
              <UButton
                v-if="selectedDigit !== null"
                size="xs"
                color="neutral"
                variant="outline"
                @click="handleSelectDigit(null)"
              >
                {{ t('analytics.caattUi.benford.clearFilter') }}
              </UButton>
            </div>
          </template>
          <div class="h-72 sm:h-80 lg:h-96">
            <Bar :data="benfordChartData" :options="benfordChartOptions" />
          </div>
        </UCard>

        <UCard class="xl:col-span-7">
          <template #header>
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('analytics.caattUi.benford.chiSquare') }}
                </h3>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                  {{ t('analytics.caattUi.benford.clickDigitHint') }}
                </p>
              </div>
              <UButton
                size="xs"
                color="neutral"
                variant="ghost"
                :icon="showMethodology ? 'i-lucide-chevron-up' : 'i-lucide-book-open'"
                @click="showMethodology = !showMethodology"
              >
                {{ showMethodology ? t('analytics.caattUi.benford.hideMethodology') : t('analytics.caattUi.benford.showMethodology') }}
              </UButton>
            </div>
          </template>

          <!-- Optional Statistical Methodology Drawer -->
          <div v-if="showMethodology" class="mb-4 rounded-xl bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700 p-3.5 space-y-2">
            <p class="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
              {{ t('analytics.caattUi.benford.formulaDesc') }}
            </p>
            <div class="flex flex-wrap gap-2 text-xs font-mono">
              <span class="px-2 py-0.5 rounded bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700">
                P(d) = log10(1 + 1/d)
              </span>
              <span class="px-2 py-0.5 rounded bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700">
                E_d = N × P(d)
              </span>
              <span class="px-2 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950/60 text-indigo-700 dark:text-indigo-300 font-semibold">
                χ²_d = (O_d − E_d)² / E_d
              </span>
            </div>
          </div>

          <!-- Desktop Plain-Language Summary Table (1-9) -->
          <div class="hidden sm:block overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-sm">
              <thead class="bg-gray-50 dark:bg-gray-800/60">
                <tr>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.digit') }}
                  </th>
                  <th class="px-3 py-2.5 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.actualCount') }}
                  </th>
                  <th class="px-3 py-2.5 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.expectedCount') }}
                  </th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.deviation') }}
                  </th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.chiSquareShare') }}
                  </th>
                  <th class="px-3 py-2.5 text-center text-xs font-semibold text-gray-600 dark:text-gray-300">
                    {{ t('analytics.caattUi.benford.status') }}
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr
                  v-for="r in benfordState.records"
                  :key="r.digit"
                  class="cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-gray-800/50"
                  :class="selectedDigit === r.digit ? 'bg-indigo-50/80 dark:bg-indigo-950/40' : ''"
                  @click="handleSelectDigit(r.digit)"
                >
                  <td class="px-3 py-2.5 whitespace-nowrap">
                    <UBadge :color="selectedDigit === r.digit ? 'primary' : (r.isSignificant ? 'error' : 'neutral')" variant="solid" size="xs">
                      {{ t('analytics.caattUi.benford.startsWithDigit', { digit: r.digit }) }}
                    </UBadge>
                  </td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <div class="font-bold text-gray-900 dark:text-white tabular-nums">
                      {{ r.actualCount.toLocaleString('id-ID') }}
                    </div>
                    <div class="text-[11px] text-gray-500 tabular-nums">
                      {{ r.actualPct.toFixed(1) }}%
                    </div>
                  </td>
                  <td class="px-3 py-2.5 text-right whitespace-nowrap">
                    <div class="font-medium text-gray-700 dark:text-gray-300 tabular-nums">
                      ~{{ Math.round(r.expectedCount).toLocaleString('id-ID') }}
                    </div>
                    <div class="text-[11px] text-gray-400 tabular-nums">
                      {{ r.expectedPct.toFixed(1) }}%
                    </div>
                  </td>
                  <td class="px-3 py-2.5 whitespace-nowrap">
                    <UBadge
                      :color="r.deviationPct > 0.5 ? 'warning' : (r.deviationPct < -0.5 ? 'info' : 'neutral')"
                      variant="subtle"
                      size="xs"
                    >
                      {{ formatCountDiffText(r) }}
                    </UBadge>
                    <div class="text-[11px] text-gray-400 mt-0.5">
                      {{ formatDirectionLabel(r) }}
                    </div>
                  </td>
                  <td class="px-3 py-2.5 whitespace-nowrap min-w-32">
                    <div class="flex items-center gap-2">
                      <div class="w-16 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
                        <div
                          class="h-full rounded-full"
                          :class="r.isSignificant ? 'bg-red-500' : 'bg-indigo-500'"
                          :style="{ width: `${Math.min(100, Math.max(4, r.chiSquareSharePct))}%` }"
                        />
                      </div>
                      <span
                        class="text-xs font-semibold tabular-nums"
                        :class="r.isSignificant ? 'text-red-600 dark:text-red-400' : 'text-gray-700 dark:text-gray-300'"
                      >
                        {{ r.chiSquareSharePct.toFixed(0) }}%
                      </span>
                    </div>
                    <div class="text-[11px] text-gray-400 tabular-nums mt-0.5">
                      {{ t('analytics.caattUi.benford.chiSquareValue') }}: {{ r.chiSquare.toFixed(1) }}
                    </div>
                  </td>
                  <td class="px-3 py-2.5 text-center whitespace-nowrap">
                    <UBadge :color="r.isSignificant ? 'error' : 'success'" variant="subtle" size="xs">
                      {{ r.isSignificant ? t('analytics.caattUi.benford.statusAnomaly') : t('analytics.caattUi.benford.statusNormal') }}
                    </UBadge>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Mobile Cards for Digit Summary (1-9) -->
          <div class="sm:hidden space-y-2.5">
            <div
              v-for="r in benfordState.records"
              :key="`mob-summary-${r.digit}`"
              class="rounded-xl border border-gray-200 dark:border-gray-800 p-3 space-y-2 cursor-pointer"
              :class="selectedDigit === r.digit ? 'ring-2 ring-indigo-500 bg-indigo-50/30 dark:bg-indigo-950/20' : ''"
              @click="handleSelectDigit(r.digit)"
            >
              <div class="flex items-center justify-between">
                <UBadge :color="r.isSignificant ? 'error' : 'neutral'" variant="solid" size="xs">
                  {{ t('analytics.caattUi.benford.startsWithDigit', { digit: r.digit }) }}
                </UBadge>
                <UBadge :color="r.isSignificant ? 'error' : 'success'" variant="subtle" size="xs">
                  {{ r.isSignificant ? t('analytics.caattUi.benford.statusAnomaly') : t('analytics.caattUi.benford.statusNormal') }}
                </UBadge>
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span class="text-gray-400 block">{{ t('analytics.caattUi.benford.actualCount') }}</span>
                  <span class="font-bold text-gray-900 dark:text-white">{{ r.actualCount.toLocaleString('id-ID') }} ({{ r.actualPct.toFixed(1) }}%)</span>
                </div>
                <div>
                  <span class="text-gray-400 block">{{ t('analytics.caattUi.benford.expectedCount') }}</span>
                  <span class="font-medium text-gray-700 dark:text-gray-300">~{{ Math.round(r.expectedCount).toLocaleString('id-ID') }} ({{ r.expectedPct.toFixed(1) }}%)</span>
                </div>
              </div>
              <div class="flex items-center justify-between text-xs pt-1 border-t border-gray-100 dark:border-gray-800">
                <span class="text-gray-600 dark:text-gray-300">{{ formatCountDiffText(r) }}</span>
                <span class="font-semibold" :class="r.isSignificant ? 'text-red-600 dark:text-red-400' : 'text-gray-500'">
                  {{ r.chiSquareSharePct.toFixed(0) }}% {{ t('analytics.caattUi.benford.shareRiskUnit') }}
                </span>
              </div>
            </div>
          </div>
        </UCard>
      </div>

      <!-- 4. Plain-Language Origin Breakdown (Where Deviations Came From) -->
      <UCard>
        <template #header>
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="space-y-1">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                <UIcon name="i-lucide-microscope" class="w-4 h-4 text-amber-500" />
                {{ t('analytics.caattUi.benford.originTableTitle') }}
              </h3>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('analytics.caattUi.benford.originTableDesc') }}
              </p>
            </div>
            <UBadge v-if="selectedDigit !== null" color="primary" variant="subtle" size="sm">
              {{ t('analytics.caattUi.benford.filterByDigit') }}: #{{ selectedDigit }}
            </UBadge>
          </div>
        </template>

        <UAlert
          v-if="hasIncompleteOriginDetails"
          class="mb-4"
          icon="i-lucide-info"
          color="warning"
          variant="subtle"
          :title="t('analytics.caattUi.benford.originMissingWarning')"
        />

        <!-- Desktop Origin Table -->
        <div class="hidden md:block overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-sm">
            <thead class="bg-gray-50 dark:bg-gray-800/60">
              <tr>
                <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                  {{ t('analytics.caattUi.benford.leadingDigitAndRisk') }}
                </th>
                <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                  {{ t('analytics.caattUi.benford.branchAndCategory') }}
                </th>
                <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">
                  {{ t('analytics.caattUi.benford.anomalyIndicators') }}
                </th>
                <th class="px-3 py-2.5 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">
                  {{ t('analytics.caattUi.benford.totalAmount') }}
                </th>
                <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300 min-w-80">
                  {{ t('analytics.caattUi.benford.findingOrigin') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr
                v-for="r in filteredOriginRecords"
                :key="`origin-${r.digit}`"
                class="cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-gray-800/50"
                :class="selectedDigit === r.digit ? 'bg-indigo-50/70 dark:bg-indigo-950/40' : ''"
                @click="handleSelectDigit(r.digit)"
              >
                <td class="px-3 py-3 align-top whitespace-nowrap">
                  <div class="flex items-center gap-2">
                    <UBadge :color="r.isSignificant ? 'error' : 'neutral'" variant="solid" size="sm">
                      #{{ r.digit }}
                    </UBadge>
                    <span class="text-xs font-medium" :class="r.isSignificant ? 'text-red-600 dark:text-red-400' : 'text-gray-500'">
                      {{ r.chiSquareSharePct.toFixed(0) }}% {{ t('analytics.caattUi.benford.shareRiskUnit') }}
                    </span>
                  </div>
                </td>
                <td class="px-3 py-3 align-top">
                  <div class="font-semibold text-gray-900 dark:text-white">{{ r.dominantBranch || '-' }}</div>
                  <div class="text-xs text-gray-500 mt-0.5">
                    <span v-if="r.dominantBranchCount > 0">{{ r.dominantBranchCount.toLocaleString('id-ID') }} {{ t('analytics.caattUi.benford.trxUnit') }}</span>
                    <span v-if="r.dominantCategory && r.dominantCategory !== '-'"> · {{ r.dominantCategory }}</span>
                  </div>
                </td>
                <td class="px-3 py-3 align-top">
                  <div v-if="r.roundAmountCount > 0 || r.afterHoursCount > 0 || r.suspiciousCount > 0" class="flex flex-wrap gap-1.5">
                    <UBadge v-if="r.roundAmountCount > 0" color="warning" variant="subtle" size="xs">
                      {{ t('analytics.caattUi.benford.roundAmounts') }}: {{ r.roundAmountCount.toLocaleString('id-ID') }}
                    </UBadge>
                    <UBadge v-if="r.afterHoursCount > 0" color="info" variant="subtle" size="xs">
                      {{ t('analytics.caattUi.benford.afterHours') }}: {{ r.afterHoursCount.toLocaleString('id-ID') }}
                    </UBadge>
                    <UBadge v-if="r.suspiciousCount > 0" color="error" variant="subtle" size="xs">
                      {{ t('analytics.caattUi.benford.flagged') }}: {{ r.suspiciousCount.toLocaleString('id-ID') }}
                    </UBadge>
                  </div>
                  <span v-else class="text-xs text-gray-400">
                    {{ t('analytics.caattUi.benford.noSpecialFlags') }}
                  </span>
                </td>
                <td class="px-3 py-3 align-top text-right font-mono whitespace-nowrap">
                  <div class="font-semibold text-gray-900 dark:text-white">{{ formatIDR(r.totalAmount) }}</div>
                  <div class="text-[11px] text-gray-400">{{ t('analytics.caattUi.benford.avgPrefix') }} {{ formatIDR(r.avgAmount) }}</div>
                </td>
                <td class="px-3 py-3 align-top text-xs sm:text-sm text-gray-700 dark:text-gray-300 leading-relaxed">
                  {{ getOriginDescription(r) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Mobile Responsive Origin Cards -->
        <div class="md:hidden space-y-3">
          <div
            v-for="r in filteredOriginRecords"
            :key="`origin-card-${r.digit}`"
            class="rounded-xl border border-gray-200 dark:border-gray-800 p-4 space-y-2.5 cursor-pointer"
            :class="selectedDigit === r.digit ? 'ring-2 ring-indigo-500 bg-indigo-50/30 dark:bg-indigo-950/20' : ''"
            @click="handleSelectDigit(r.digit)"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <UBadge :color="r.isSignificant ? 'error' : 'neutral'" variant="solid" size="sm">
                  {{ t('analytics.caattUi.benford.startsWithDigit', { digit: r.digit }) }}
                </UBadge>
                <span class="text-xs font-medium text-gray-500">
                  {{ r.chiSquareSharePct.toFixed(0) }}% {{ t('analytics.caattUi.benford.shareRiskUnit') }}
                </span>
              </div>
              <span class="text-xs font-mono font-semibold text-gray-900 dark:text-white">
                {{ formatIDR(r.totalAmount) }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs">
              <div>
                <span class="text-gray-400 block">{{ t('analytics.caattUi.benford.dominantBranch') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ r.dominantBranch || '-' }}</span>
              </div>
              <div>
                <span class="text-gray-400 block">{{ t('analytics.caattUi.benford.dominantCategory') }}</span>
                <span class="font-medium text-gray-900 dark:text-white">{{ r.dominantCategory || '-' }}</span>
              </div>
            </div>
            <div v-if="r.roundAmountCount > 0 || r.afterHoursCount > 0 || r.suspiciousCount > 0" class="flex flex-wrap gap-1.5">
              <UBadge v-if="r.roundAmountCount > 0" color="warning" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.roundAmounts') }}: {{ r.roundAmountCount }}
              </UBadge>
              <UBadge v-if="r.afterHoursCount > 0" color="info" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.afterHours') }}: {{ r.afterHoursCount }}
              </UBadge>
              <UBadge v-if="r.suspiciousCount > 0" color="error" variant="subtle" size="xs">
                {{ t('analytics.caattUi.benford.flagged') }}: {{ r.suspiciousCount }}
              </UBadge>
            </div>
            <p class="text-xs text-gray-600 dark:text-gray-300 leading-relaxed">
              {{ getOriginDescription(r) }}
            </p>
          </div>
        </div>
      </UCard>

      <!-- 5. Real Database Transactions Supporting Each Finding -->
      <UCard>
        <template #header>
          <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <h3 class="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                  <UIcon name="i-lucide-file-search" class="w-4 h-4 text-indigo-500" />
                  {{ t('analytics.caattUi.benford.findingsTableTitle') }}
                </h3>
                <UBadge color="neutral" variant="subtle" size="xs">
                  {{ t('analytics.caattUi.benford.filteredCountBadge', { count: filteredFindings.length }) }}
                </UBadge>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                {{ t('analytics.caattUi.benford.findingsTableDesc') }}
              </p>
            </div>

            <div class="flex flex-col sm:flex-row sm:items-center gap-2">
              <UInput
                v-model="searchQuery"
                size="xs"
                icon="i-lucide-search"
                :placeholder="t('analytics.caattUi.benford.searchTransactionsPlaceholder')"
                class="w-full sm:w-60"
              />
              <div class="flex flex-wrap items-center gap-1">
                <UButton
                  size="xs"
                  :color="selectedDigit === null ? 'primary' : 'neutral'"
                  :variant="selectedDigit === null ? 'solid' : 'outline'"
                  @click="handleSelectDigit(null)"
                >
                  {{ t('analytics.caattUi.benford.allDigits') }}
                </UButton>
                <UButton
                  v-for="r in benfordState.records"
                  :key="`btn-digit-${r.digit}`"
                  size="xs"
                  :color="selectedDigit === r.digit ? 'primary' : (r.isSignificant ? 'error' : 'neutral')"
                  :variant="selectedDigit === r.digit ? 'solid' : 'subtle'"
                  @click="handleSelectDigit(r.digit)"
                >
                  {{ r.digit }}
                </UButton>
              </div>
            </div>
          </div>
        </template>

        <div v-if="loadingDigitFindings" class="py-6 flex justify-center items-center gap-2 text-xs text-gray-500">
          <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin text-indigo-500" />
          <span>{{ t('analytics.loading') }}</span>
        </div>

        <div v-else-if="filteredFindings.length === 0" class="py-8 text-center space-y-2 text-sm text-gray-500 dark:text-gray-400">
          <p>
            {{
              selectedDigit === null && !searchQuery
                ? t('analytics.caattUi.benford.noFindingsSyncHint')
                : t('analytics.caattUi.benford.noFindingsForFilter')
            }}
          </p>
          <div v-if="selectedDigit !== null || searchQuery" class="pt-1 flex justify-center gap-2">
            <UButton size="xs" color="neutral" variant="outline" @click="searchQuery = ''; handleSelectDigit(null)">
              {{ t('analytics.caattUi.benford.clearFilter') }}
            </UButton>
          </div>
        </div>

        <div v-else>
          <!-- Desktop Real Transactions Table -->
          <div class="hidden sm:block overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-sm">
              <thead class="bg-gray-50 dark:bg-gray-800/60">
                <tr>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.digit') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.trxRef') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.trxDate') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.branch') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.category') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.customerName') }}</th>
                  <th class="px-3 py-2.5 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.amount') }}</th>
                  <th class="px-3 py-2.5 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">{{ t('analytics.caattUi.benford.indicators') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr
                  v-for="f in filteredFindings"
                  :key="`${f.trx_id}-${f.trx_ref_number}`"
                  class="hover:bg-gray-50 dark:hover:bg-gray-800/40"
                >
                  <td class="px-3 py-2.5 whitespace-nowrap">
                    <UBadge color="primary" variant="subtle" size="xs">#{{ f.digit }}</UBadge>
                  </td>
                  <td class="px-3 py-2.5 font-mono text-xs font-semibold text-gray-900 dark:text-white whitespace-nowrap">
                    {{ f.trx_ref_number }}
                  </td>
                  <td class="px-3 py-2.5 font-mono text-xs text-gray-500 whitespace-nowrap">
                    {{ f.trx_date }}
                  </td>
                  <td class="px-3 py-2.5 text-xs text-gray-800 dark:text-gray-200">
                    {{ f.branch_name }}
                  </td>
                  <td class="px-3 py-2.5 text-xs text-gray-700 dark:text-gray-300">
                    {{ f.category }}
                  </td>
                  <td class="px-3 py-2.5 text-xs text-gray-700 dark:text-gray-300">
                    <div>{{ f.customer_name }}</div>
                    <div v-if="f.channel && f.channel !== '-'" class="text-[11px] text-gray-400">
                      {{ f.channel }}
                    </div>
                  </td>
                  <td class="px-3 py-2.5 text-right font-mono font-semibold whitespace-nowrap">
                    <span class="text-gray-400 mr-0.5">Rp</span>
                    <span class="text-indigo-600 dark:text-indigo-400 underline decoration-2 underline-offset-2 font-extrabold">{{ formatHighlightedAmount(f.amount).first }}</span><span class="text-gray-900 dark:text-white">{{ formatHighlightedAmount(f.amount).rest }}</span>
                  </td>
                  <td class="px-3 py-2.5">
                    <div class="flex flex-wrap gap-1">
                      <UBadge v-if="f.is_suspicious" color="error" variant="subtle" size="xs">
                        {{ t('analytics.caattUi.benford.flagged') }}
                      </UBadge>
                      <UBadge v-if="f.is_round_amount" color="warning" variant="subtle" size="xs">
                        {{ t('analytics.caattUi.benford.roundAmounts') }}
                      </UBadge>
                      <UBadge v-if="f.is_after_hours" color="info" variant="subtle" size="xs">
                        {{ t('analytics.caattUi.benford.afterHours') }}
                      </UBadge>
                      <span
                        v-if="!f.is_suspicious && !f.is_round_amount && !f.is_after_hours"
                        class="text-xs text-gray-400"
                      >
                        {{ t('analytics.caattUi.benford.normalTransaction') }}
                      </span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Mobile Transaction Cards -->
          <div class="sm:hidden space-y-3">
            <div
              v-for="f in filteredFindings"
              :key="`mob-${f.trx_id}-${f.trx_ref_number}`"
              class="rounded-xl border border-gray-200 dark:border-gray-800 p-3.5 space-y-2"
            >
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 min-w-0">
                  <UBadge color="primary" variant="subtle" size="xs">#{{ f.digit }}</UBadge>
                  <span class="font-mono text-xs font-semibold text-gray-900 dark:text-white truncate">
                    {{ f.trx_ref_number }}
                  </span>
                </div>
                <span class="font-mono text-xs font-bold text-gray-900 dark:text-white shrink-0">
                  Rp <span class="text-indigo-600 dark:text-indigo-400 underline">{{ formatHighlightedAmount(f.amount).first }}</span>{{ formatHighlightedAmount(f.amount).rest }}
                </span>
              </div>
              <div class="text-xs text-gray-600 dark:text-gray-300 flex flex-wrap gap-x-3 gap-y-1">
                <span>{{ f.branch_name }}</span>
                <span>·</span>
                <span>{{ f.category }}</span>
                <span>·</span>
                <span class="text-gray-400">{{ f.trx_date }}</span>
              </div>
              <div class="text-xs text-gray-500">
                {{ f.customer_name }}
                <span v-if="f.channel && f.channel !== '-'"> ({{ f.channel }})</span>
              </div>
              <div v-if="f.is_suspicious || f.is_round_amount || f.is_after_hours" class="flex flex-wrap gap-1 pt-1">
                <UBadge v-if="f.is_suspicious" color="error" variant="subtle" size="xs">
                  {{ t('analytics.caattUi.benford.flagged') }}
                </UBadge>
                <UBadge v-if="f.is_round_amount" color="warning" variant="subtle" size="xs">
                  {{ t('analytics.caattUi.benford.roundAmounts') }}
                </UBadge>
                <UBadge v-if="f.is_after_hours" color="info" variant="subtle" size="xs">
                  {{ t('analytics.caattUi.benford.afterHours') }}
                </UBadge>
              </div>
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>
