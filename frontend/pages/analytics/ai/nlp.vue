<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend
} from 'chart.js'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()
const {
  loading,
  nlpState,
  getRiskConfig,
  sentimentColor,
  riskCategoryColor,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

const categoryDoughnutData = computed(() => ({
  labels: Object.keys(nlpState.value.categoryDistribution || {}),
  datasets: [{
    data: Object.values(nlpState.value.categoryDistribution || {}) as number[],
    backgroundColor: [
      '#EF4444', '#8B5CF6', '#F59E0B', '#3B82F6',
      '#10B981', '#6366F1', '#EC4899',
    ],
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.1)',
  }]
}))

const getSentimentCount = (key: string): number => {
  const dist = nlpState.value?.sentimentDistribution || {}
  const lower = key.toLowerCase()
  const upper = key.charAt(0).toUpperCase() + key.slice(1).toLowerCase()
  return (dist[lower] ?? dist[upper] ?? dist[key] ?? 0) as number
}

const sentimentDoughnutData = computed(() => ({
  labels: ['Positive', 'Neutral', 'Negative'],
  datasets: [{
    data: [
      getSentimentCount('positive'),
      getSentimentCount('neutral'),
      getSentimentCount('negative')
    ],
    backgroundColor: ['#10B981', '#F59E0B', '#EF4444'],
    borderWidth: 2,
    borderColor: 'rgba(255,255,255,0.1)',
  }]
}))

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' as const }
  }
}
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      :title="t('navigation.nlpAnalysis')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-file-search"
      currentSubModule="nlp"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content -->
    <div v-else class="space-y-6">
      <UAlert
        icon="i-heroicons-document-magnifying-glass"
        color="success"
        variant="subtle"
        :title="t('analytics.nlp.alertTitle')"
        :description="t('analytics.nlp.alertDesc')"
      />

      <!-- Charts Row -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <UCard>
          <template #header>
            <h3 class="font-bold text-sm">{{ t('analytics.nlp.chartCategoryTitle') }}</h3>
          </template>
          <div class="h-60">
            <Doughnut :data="categoryDoughnutData" :options="doughnutOptions" />
          </div>
        </UCard>
        <UCard>
          <template #header>
            <h3 class="font-bold text-sm">{{ t('analytics.nlp.chartSentimentTitle') }}</h3>
          </template>
          <div class="h-60">
            <Doughnut :data="sentimentDoughnutData" :options="doughnutOptions" />
          </div>
        </UCard>
      </div>

      <!-- Documents Table -->
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-document-text" class="text-violet-500" />
            <h3 class="font-bold">{{ t('analytics.nlp.tableTitle') }}</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colDocID') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colTitle') }}</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colSource') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colCategory') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colSentiment') }}</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">{{ t('analytics.nlp.colRiskLevel') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="doc in nlpState.documents" :key="doc.docId" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold text-md">{{ doc.docId }}</td>
                <td class="py-3 px-3">
                  <div class="font-bold text-md">{{ doc.title }}</div>
                  <div class="text-[11px] text-gray-400 mt-0.5 italic truncate max-w-[300px]">"{{ doc.excerpt }}"</div>
                </td>
                <td class="py-3 px-3"><UBadge :color="doc.source === 'Working Paper' ? 'primary' : 'warning'" variant="subtle" size="md">{{ doc.source }}</UBadge></td>
                <td class="text-center py-3 px-3"><UBadge :color="riskCategoryColor(doc.autoCategory)" variant="subtle" size="md">{{ doc.autoCategory }}</UBadge></td>
                <td class="text-center py-3 px-3"><UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="md">{{ doc.sentiment }}</UBadge></td>
                <td class="text-center py-3 px-3">
                  <UBadge
                    :style="{ backgroundColor: getRiskConfig(doc.riskLevel).color, color: 'white' }"
                    variant="solid"
                    size="md"
                    class="font-bold"
                  >
                    {{ getRiskConfig(doc.riskLevel).label }}
                  </UBadge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </UCard>
    </div>
  </div>
</template>
