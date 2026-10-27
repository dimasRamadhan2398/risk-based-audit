<script setup lang="ts">
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useI18n } from '~/composables/useI18n'

const props = defineProps<{
  title: string
  subtitle?: string
  icon?: string
  currentSubModule?: 'risk-scoring' | 'anomaly-detection' | 'nlp' | 'kpi-forecast' | 'overview'
}>()

const { t } = useI18n()
const {
  isAiConnected,
  usingCachedRealData,
  lastSyncedTime,
  xgboostState,
  isolationState,
  nlpState,
  timeseriesState,
  summary
} = useAiAnalytics()
</script>

<template>
  <div class="space-y-6 mb-8">
    <!-- Breadcrumb & Header Row -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <!-- Breadcrumb -->
        <nav class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 mb-2">
          <NuxtLink to="/analytics" class="hover:text-primary transition-colors font-medium">
            {{ t('navigation.analytics') }}
          </NuxtLink>
          <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
          <span class="font-medium text-gray-600 dark:text-gray-300">{{ t('analytics.hub.aiTitle') }}</span>
          <template v-if="props.title">
            <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5 text-gray-400" />
            <span class="text-indigo-600 dark:text-indigo-400 font-bold">{{ props.title }}</span>
          </template>
        </nav>

        <h1 class="text-2xl md:text-3xl font-black tracking-tight text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon :name="props.icon || 'i-heroicons-cpu-chip'" class="text-indigo-600 dark:text-indigo-400 w-8 h-8 shrink-0" />
          <span>{{ props.title }}</span>
        </h1>
        <p v-if="props.subtitle" class="text-sm md:text-base text-gray-500 dark:text-gray-400 mt-1">
          {{ props.subtitle }}
        </p>
      </div>

      <!-- AI Status Badge (Transferred from index.vue:L1174-L1179) -->
      <div class="flex items-center gap-2 shrink-0">
        <UBadge :color="isAiConnected ? 'primary' : 'warning'" variant="subtle" size="lg" class="px-3.5 py-1.5 font-bold shadow-xs">
          <UIcon
            :name="isAiConnected ? 'i-heroicons-check-circle' : 'i-heroicons-exclamation-triangle'"
            :class="isAiConnected ? 'text-emerald-500' : 'text-amber-500'"
            class="w-4 h-4 mr-1.5"
          />
          {{ isAiConnected ? t('analytics.statusActive') : (usingCachedRealData ? t('analytics.statusOffline') : t('analytics.statusDisconnected')) }}
        </UBadge>
      </div>
    </div>

    <!-- Connection Warning Alert Banner (Transferred from index.vue:L1183-L1193) -->
    <UAlert
      v-if="!isAiConnected"
      icon="i-heroicons-exclamation-triangle"
      color="warning"
      variant="solid"
      class="border border-amber-500 shadow-md font-medium"
      :title="usingCachedRealData ? t('analytics.warningTitleCached') : t('analytics.warningTitleDisconnected')"
      :description="usingCachedRealData
        ? t('analytics.warningDescCached', { time: lastSyncedTime || '' })
        : t('analytics.warningDescDisconnected')"
    />

    <!-- AI Summary Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Card 1: Risk Scoring -->
      <div
        class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] transition-all duration-200 dashboard-card"
        :class="props.currentSubModule === 'risk-scoring'
          ? 'ring-2 ring-indigo-500/30 border-indigo-500 dark:border-indigo-400 bg-indigo-50/40 dark:bg-indigo-950/20 shadow-sm'
          : ''"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('analytics.summary.entitiesScored') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ xgboostState.predictions.length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
            <UIcon name="i-heroicons-chart-bar-square" class="text-indigo-600 size-5 dark:text-indigo-400" />
          </div>
        </div>
      </div>

      <!-- Card 2: Anomaly Detection -->
      <div
        class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] transition-all duration-200 dashboard-card"
        :class="props.currentSubModule === 'anomaly-detection'
          ? 'ring-2 ring-rose-500/30 border-rose-500 dark:border-rose-400 bg-rose-50/40 dark:bg-rose-950/20 shadow-sm'
          : ''"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('analytics.summary.anomaliesDetected') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ isolationState.summary?.anomaliesFound || summary.anomaliesDetected || 150 }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-rose-50 p-2.5 flex items-center justify-center dark:bg-rose-500/10">
            <UIcon name="i-heroicons-shield-exclamation" class="text-rose-600 size-5 dark:text-rose-400" />
          </div>
        </div>
      </div>

      <!-- Card 3: NLP Analysis -->
      <div
        class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] transition-all duration-200 dashboard-card"
        :class="props.currentSubModule === 'nlp'
          ? 'ring-2 ring-amber-500/30 border-amber-500 dark:border-amber-400 bg-amber-50/40 dark:bg-amber-950/20 shadow-sm'
          : ''"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('analytics.summary.docsAnalyzed') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ nlpState.documents.length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-amber-50 p-2.5 flex items-center justify-center dark:bg-amber-500/10">
            <UIcon name="i-heroicons-document-magnifying-glass" class="text-amber-600 size-5 dark:text-amber-400" />
          </div>
        </div>
      </div>

      <!-- Card 4: KPI Forecast -->
      <div
        class="rounded-xl p-5 flex flex-col justify-between min-h-[100px] transition-all duration-200 dashboard-card"
        :class="props.currentSubModule === 'kpi-forecast'
          ? 'ring-2 ring-violet-500/30 border-violet-500 dark:border-violet-400 bg-violet-50/40 dark:bg-violet-950/20 shadow-sm'
          : ''"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0">
            <span class="text-sm font-semibold text-slate-500 dark:text-slate-200 block leading-tight">
              {{ t('analytics.summary.kpiForecasts') }}
            </span>
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white mt-1.5">
              {{ timeseriesState.kpiForecasts.length }}
            </h4>
          </div>
          <div class="shrink-0 rounded-xl bg-violet-50 p-2.5 flex items-center justify-center dark:bg-violet-500/10">
            <UIcon name="i-heroicons-bell-alert" class="text-violet-600 size-5 dark:text-violet-400" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
