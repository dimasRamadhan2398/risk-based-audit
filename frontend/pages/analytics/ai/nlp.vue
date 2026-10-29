<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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

const mobilePage = ref(1)
const mobilePageSize = ref(10)

const paginatedDocuments = computed<any[]>(() => {
  const start = (mobilePage.value - 1) * mobilePageSize.value
  return nlpState.value.documents.slice(start, start + mobilePageSize.value)
})
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
        <!-- Desktop View (>= md): Full Table -->
        <div class="hidden md:block overflow-x-auto">
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
                    :class="getRiskConfig(doc.riskLevel).badgeClass"
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

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Empty State -->
          <div
            v-if="!nlpState.documents.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-heroicons-table-cells" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs text-[var(--text-muted)]">{{ t('analytics.nlp.empty') || 'Tidak ada data dokumen' }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(doc, index) in paginatedDocuments"
              :key="doc.docId || index"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: Doc ID & Source + Risk Level Badge -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-slate-700 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-md border border-slate-200/80 dark:border-slate-700">
                    <UIcon name="i-heroicons-document-text" class="size-3 text-violet-500" />
                    {{ doc.docId }}
                  </span>
                  <UBadge :color="doc.source === 'Working Paper' ? 'primary' : 'warning'" variant="subtle" size="sm">
                    {{ doc.source }}
                  </UBadge>
                </div>
                <UBadge
                  :class="getRiskConfig(doc.riskLevel).badgeClass"
                  variant="solid"
                  size="sm"
                  class="font-bold shrink-0"
                >
                  {{ getRiskConfig(doc.riskLevel).label }}
                </UBadge>
              </div>

              <!-- Card Body: Title & Excerpt -->
              <div class="space-y-1.5">
                <div class="font-bold text-sm text-gray-900 dark:text-white leading-snug">
                  {{ doc.title }}
                </div>
                <p v-if="doc.excerpt" class="text-xs text-gray-500 dark:text-gray-400 italic bg-gray-50 dark:bg-gray-800/40 p-2.5 rounded-lg border border-gray-100 dark:border-gray-800 leading-relaxed">
                  "{{ doc.excerpt }}"
                </p>
              </div>

              <!-- Badges Row: Category & Sentiment -->
              <div class="flex items-center gap-2 flex-wrap pt-0.5">
                <div class="flex items-center gap-1.5 text-xs">
                  <span class="text-[11px] text-gray-400 uppercase font-semibold tracking-wider">{{ t('analytics.nlp.colCategory') }}:</span>
                  <UBadge :color="riskCategoryColor(doc.autoCategory)" variant="subtle" size="sm">
                    {{ doc.autoCategory }}
                  </UBadge>
                </div>
                <div class="flex items-center gap-1.5 text-xs">
                  <span class="text-[11px] text-gray-400 uppercase font-semibold tracking-wider">{{ t('analytics.nlp.colSentiment') }}:</span>
                  <UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="sm">
                    {{ doc.sentiment }}
                  </UBadge>
                </div>
              </div>

              <!-- Card Footer: Index & Source -->
              <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800/80 text-xs text-slate-500 dark:text-slate-400">
                <span class="text-xs text-slate-400 font-mono">
                  #{{ (mobilePage - 1) * mobilePageSize + Number(index) + 1 }}
                </span>
                <span class="text-[11px] text-gray-400">{{ doc.source }}</span>
              </div>
            </div>

            <!-- Mobile Pagination Controls -->
            <div
              v-if="nlpState.documents.length > mobilePageSize"
              class="flex flex-col sm:flex-row items-center justify-between gap-3 pt-3 border-t border-slate-100 dark:border-slate-800"
            >
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {{ (mobilePage - 1) * mobilePageSize + 1 }}-{{ Math.min(mobilePage * mobilePageSize, nlpState.documents.length) }} of {{ nlpState.documents.length }}
              </span>
              <UPagination
                v-model:page="mobilePage"
                :items-per-page="mobilePageSize"
                :total="nlpState.documents.length"
                size="md"
                active-color="primary"
                color="neutral"
                variant="outline"
                class="max-w-full overflow-x-auto"
              />
            </div>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>
