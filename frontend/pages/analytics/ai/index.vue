<script setup lang="ts">
import { onMounted } from 'vue'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'

const { t } = useI18n()
const {
  loading,
  xgboostState,
  isolationState,
  nlpState,
  timeseriesState,
  fetchAiAnalytics
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <AiInsightHeader
      title="AI Insights Suite"
      :subtitle="t('analytics.subtitle')"
      icon="i-heroicons-cpu-chip"
      currentSubModule="overview"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <div v-else class="space-y-8">
      <!-- Welcome Callout Banner -->
      <UAlert
        icon="i-heroicons-sparkles"
        color="primary"
        variant="subtle"
        title="Modul Kecerdasan Buatan Terintegrasi (AI Insights)"
        description="Fitur Analitik AI menyediakan 4 engine prediktif audit internal: XGBoost Risk Scoring, Isolation Forest Anomaly Detection, IndoBERT NLP Analysis, dan PyTorch LSTM KPI Forecast. Pilih salah satu modul di bawah atau gunakan kartu navigasi di atas."
      />

      <!-- Grid of 4 Module Quick-Access Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- 1. Risk Scoring -->
        <UCard class="hover:border-indigo-500 hover:shadow-md transition-all duration-200">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-indigo-500/10 flex items-center justify-center text-indigo-500">
                  <UIcon name="i-lucide-binary" class="w-6 h-6" />
                </div>
                <div>
                  <h3 class="font-bold text-lg text-gray-900 dark:text-white">
                    {{ t('navigation.riskScoringPrediction') }}
                  </h3>
                  <p class="text-xs text-gray-500">XGBoost Regressor • 8 Fitur Matriks Risiko</p>
                </div>
              </div>
              <UBadge color="primary" variant="subtle">ML Model</UBadge>
            </div>
          </template>
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4 leading-relaxed">
            Prediksi skor risiko kuantitatif cabang & departemen secara proaktif berdasarkan data historis audit, deviasi KPI, dan insiden operasional.
          </p>
          <div class="flex items-center justify-between pt-3 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs font-semibold text-indigo-600 dark:text-indigo-400">
              {{ xgboostState.predictions.length }} Entitas Dinilai
            </span>
            <UButton to="/analytics/ai/risk-scoring" color="primary" variant="solid" size="sm" icon="i-heroicons-arrow-right" trailing>
              Buka Analitik
            </UButton>
          </div>
        </UCard>

        <!-- 2. Anomaly Detection -->
        <UCard class="hover:border-rose-500 hover:shadow-md transition-all duration-200">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-rose-500/10 flex items-center justify-center text-rose-500">
                  <UIcon name="i-lucide-shield-alert" class="w-6 h-6" />
                </div>
                <div>
                  <h3 class="font-bold text-lg text-gray-900 dark:text-white">
                    {{ t('navigation.anomalyDetection') }}
                  </h3>
                  <p class="text-xs text-gray-500">Isolation Forest • Multi-Variate Clustering</p>
                </div>
              </div>
              <UBadge color="error" variant="subtle">Unsupervised</UBadge>
            </div>
          </template>
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4 leading-relaxed">
            Mendeteksi anomali operasional perbankan (Funding, Lending, Treasury, KYC, IT Control) tanpa rule statis untuk mitigasi fraud dini.
          </p>
          <div class="flex items-center justify-between pt-3 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs font-semibold text-rose-600 dark:text-rose-400">
              {{ isolationState.anomalies.length }} Anomali Terdeteksi
            </span>
            <UButton to="/analytics/ai/anomaly-detection" color="error" variant="solid" size="sm" icon="i-heroicons-arrow-right" trailing>
              Buka Analitik
            </UButton>
          </div>
        </UCard>

        <!-- 3. NLP Document Processing -->
        <UCard class="hover:border-amber-500 hover:shadow-md transition-all duration-200">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-amber-500/10 flex items-center justify-center text-amber-500">
                  <UIcon name="i-lucide-file-search" class="w-6 h-6" />
                </div>
                <div>
                  <h3 class="font-bold text-lg text-gray-900 dark:text-white">
                    {{ t('navigation.nlpAnalysis') }}
                  </h3>
                  <p class="text-xs text-gray-500">IndoBERT Fine-Tuned • Klasifikasi Teks Audit</p>
                </div>
              </div>
              <UBadge color="warning" variant="subtle">NLP Deep Learning</UBadge>
            </div>
          </template>
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4 leading-relaxed">
            Analisis sentimen dan kategorisasi otomatis temuan kertas kerja audit (Warkat, Memo, Working Paper) ke dalam taksonomi risiko baku.
          </p>
          <div class="flex items-center justify-between pt-3 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs font-semibold text-amber-600 dark:text-amber-400">
              {{ nlpState.documents.length }} Dokumen Dianalisis
            </span>
            <UButton to="/analytics/ai/nlp" color="warning" variant="solid" size="sm" icon="i-heroicons-arrow-right" trailing>
              Buka Analitik
            </UButton>
          </div>
        </UCard>

        <!-- 4. KPI Forecast -->
        <UCard class="hover:border-violet-500 hover:shadow-md transition-all duration-200">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-violet-500/10 flex items-center justify-center text-violet-500">
                  <UIcon name="i-lucide-trending-up" class="w-6 h-6" />
                </div>
                <div>
                  <h3 class="font-bold text-lg text-gray-900 dark:text-white">
                    {{ t('navigation.kpiForecast') }}
                  </h3>
                  <p class="text-xs text-gray-500">PyTorch LSTM • Time-Series Horizon</p>
                </div>
              </div>
              <UBadge color="info" variant="subtle">Predictive Horizon</UBadge>
            </div>
          </template>
          <p class="text-sm text-gray-600 dark:text-gray-300 mb-4 leading-relaxed">
            Prakiraan lintasan KPI audit kuartal mendatang disertai batas keyakinan (Upper/Lower Bound) dan rekomendasi perbaikan proaktif.
          </p>
          <div class="flex items-center justify-between pt-3 border-t border-gray-100 dark:border-gray-800">
            <span class="text-xs font-semibold text-violet-600 dark:text-violet-400">
              {{ timeseriesState.kpiForecasts.length }} Prakiraan KPI Aktif
            </span>
            <UButton to="/analytics/ai/kpi-forecast" color="primary" variant="solid" size="sm" icon="i-heroicons-arrow-right" trailing>
              Buka Analitik
            </UButton>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
