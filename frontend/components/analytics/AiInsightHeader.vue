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
          <span class="font-medium text-gray-600 dark:text-gray-300">AI Insights</span>
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

    <!-- AI Summary Stats Cards (Transferred from index.vue:L1205-L1250) -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <!-- Card 1: Risk Scoring -->
      <NuxtLink to="/analytics/ai/risk-scoring" class="block group">
        <UCard
          class="border transition-all duration-200 h-full"
          :class="props.currentSubModule === 'risk-scoring'
            ? 'border-indigo-500 dark:border-indigo-400 ring-2 ring-indigo-500/20 shadow-md bg-indigo-50/30 dark:bg-indigo-950/20'
            : 'border-indigo-100 dark:border-indigo-900/50 hover:border-indigo-300 dark:hover:border-indigo-700 hover:shadow-sm'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-indigo-500/10 flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <UIcon name="i-heroicons-chart-bar-square" class="w-5 h-5 text-indigo-500" />
            </div>
            <div class="min-w-0">
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400 truncate">
                {{ t('analytics.summary.entitiesScored') }}
              </div>
              <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400">
                {{ xgboostState.predictions.length }}
              </div>
            </div>
          </div>
        </UCard>
      </NuxtLink>

      <!-- Card 2: Anomaly Detection -->
      <NuxtLink to="/analytics/ai/anomaly-detection" class="block group">
        <UCard
          class="border transition-all duration-200 h-full"
          :class="props.currentSubModule === 'anomaly-detection'
            ? 'border-rose-500 dark:border-rose-400 ring-2 ring-rose-500/20 shadow-md bg-rose-50/30 dark:bg-rose-950/20'
            : 'border-rose-100 dark:border-rose-900/50 hover:border-rose-300 dark:hover:border-rose-700 hover:shadow-sm'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-rose-500/10 flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <UIcon name="i-heroicons-shield-exclamation" class="w-5 h-5 text-rose-500" />
            </div>
            <div class="min-w-0">
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400 truncate">
                {{ t('analytics.summary.anomaliesDetected') }}
              </div>
              <div class="text-2xl font-black text-rose-600 dark:text-rose-400">
                {{ isolationState.summary?.anomaliesFound || summary.anomaliesDetected || 150 }}
              </div>
            </div>
          </div>
        </UCard>
      </NuxtLink>

      <!-- Card 3: NLP Analysis -->
      <NuxtLink to="/analytics/ai/nlp" class="block group">
        <UCard
          class="border transition-all duration-200 h-full"
          :class="props.currentSubModule === 'nlp'
            ? 'border-amber-500 dark:border-amber-400 ring-2 ring-amber-500/20 shadow-md bg-amber-50/30 dark:bg-amber-950/20'
            : 'border-amber-100 dark:border-amber-900/50 hover:border-amber-300 dark:hover:border-amber-700 hover:shadow-sm'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-amber-500/10 flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <UIcon name="i-heroicons-document-magnifying-glass" class="w-5 h-5 text-amber-500" />
            </div>
            <div class="min-w-0">
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400 truncate">
                {{ t('analytics.summary.docsAnalyzed') }}
              </div>
              <div class="text-2xl font-black text-amber-600 dark:text-amber-400">
                {{ nlpState.documents.length }}
              </div>
            </div>
          </div>
        </UCard>
      </NuxtLink>

      <!-- Card 4: KPI Forecast -->
      <NuxtLink to="/analytics/ai/kpi-forecast" class="block group">
        <UCard
          class="border transition-all duration-200 h-full"
          :class="props.currentSubModule === 'kpi-forecast'
            ? 'border-violet-500 dark:border-violet-400 ring-2 ring-violet-500/20 shadow-md bg-violet-50/30 dark:bg-violet-950/20'
            : 'border-violet-100 dark:border-violet-900/50 hover:border-violet-300 dark:hover:border-violet-700 hover:shadow-sm'"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-violet-500/10 flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <UIcon name="i-heroicons-bell-alert" class="w-5 h-5 text-violet-500" />
            </div>
            <div class="min-w-0">
              <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400 truncate">
                {{ t('analytics.summary.kpiForecasts') }}
              </div>
              <div class="text-2xl font-black text-violet-600 dark:text-violet-400">
                {{ timeseriesState.kpiForecasts.length }}
              </div>
            </div>
          </div>
        </UCard>
      </NuxtLink>
    </div>
  </div>
</template>
