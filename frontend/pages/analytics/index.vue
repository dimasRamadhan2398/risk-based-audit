<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import { onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

// Backward-compatibility: Redirect legacy query params ?tab=... to the new dedicated pages
onMounted(() => {
  const tab = route.query.tab as string | undefined
  if (tab) {
    const routeMap: Record<string, string> = {
      xgboost: '/analytics/ai/risk-scoring',
      isolation: '/analytics/ai/anomaly-detection',
      nlp: '/analytics/ai/nlp',
      'kpi-forecast': '/analytics/ai/kpi-forecast',
      timeseries: '/analytics/ai/kpi-forecast',
      caatt_full_population: '/analytics/caatt/full-population',
      caatt_duplicate_gap: '/analytics/caatt/duplicate-gap',
      caatt_benford: '/analytics/caatt/benford',
      caatt_stratification: '/analytics/caatt/stratification',
      caatt_reconciliation: '/analytics/caatt/reconciliation',
      caatt_policy: '/analytics/caatt/policy',
      caatt_data_quality: '/analytics/caatt/data-quality'
    }
    const target = routeMap[tab]
    if (target) {
      router.replace(target)
    }
  }
})

const aiModules = [
  {
    title: t('navigation.riskScoringPrediction'),
    description: 'Prediksi kuantitatif probabilitas dan dampak risiko departemen/cabang menggunakan model machine learning XGBoost.',
    icon: 'i-lucide-binary',
    badge: 'XGBoost ML',
    badgeColor: 'primary' as const,
    to: '/analytics/ai/risk-scoring'
  },
  {
    title: t('navigation.anomalyDetection'),
    description: 'Deteksi deviasi pola transaksi dan anomali operasional perbankan menggunakan algoritma Isolation Forest.',
    icon: 'i-lucide-shield-alert',
    badge: 'Isolation Forest',
    badgeColor: 'error' as const,
    to: '/analytics/ai/anomaly-detection'
  },
  {
    title: t('navigation.nlpAnalysis'),
    description: 'Ekstraksi sentimen dan klasifikasi otomatis taksonomi risiko dari kertas kerja audit berbasis IndoBERT deep learning.',
    icon: 'i-lucide-file-search',
    badge: 'IndoBERT NLP',
    badgeColor: 'warning' as const,
    to: '/analytics/ai/nlp'
  },
  {
    title: t('navigation.kpiForecast'),
    description: 'Prakiraan lintasan kinerja KPI audit kuartal mendatang disertai rentang keyakinan prediktif berbasis PyTorch LSTM.',
    icon: 'i-lucide-trending-up',
    badge: 'PyTorch LSTM',
    badgeColor: 'info' as const,
    to: '/analytics/ai/kpi-forecast'
  }
]

const caattModules = [
  {
    title: t('navigation.caattFullPopulation'),
    description: 'Pengujian 100% populasi transaksi kredit, valas, dan tunai tanpa sampling risk untuk mendeteksi pelanggaran plafon.',
    icon: 'i-lucide-check-check',
    badge: 'Full Population',
    badgeColor: 'primary' as const,
    to: '/analytics/caatt/full-population'
  },
  {
    title: t('navigation.caattDuplicateGap'),
    description: 'Deteksi transaksi duplikat bernilai identik dan identifikasi celah nomor warkat/voucher General Ledger yang hilang.',
    icon: 'i-lucide-copy-x',
    badge: 'Integrity Check',
    badgeColor: 'warning' as const,
    to: '/analytics/caatt/duplicate-gap'
  },
  {
    title: t('navigation.caattBenford'),
    description: 'Uji forensik sebaran angka pertama (First-Digit Law) untuk mendeteksi potensi manipulasi data atau transaksi terstruktur.',
    icon: 'i-lucide-calculator',
    badge: 'Forensic Math',
    badgeColor: 'info' as const,
    to: '/analytics/caatt/benford'
  },
  {
    title: t('navigation.caattStratification'),
    description: 'Stratifikasi nominal transaksi dan portofolio untuk evaluasi konsentrasi risiko dan Monetary Unit Sampling.',
    icon: 'i-lucide-layers',
    badge: 'Stratification',
    badgeColor: 'neutral' as const,
    to: '/analytics/caatt/stratification'
  },
  {
    title: t('navigation.caattReconciliation'),
    description: '3-Way Reconciliation otomatis antara Core Banking System (CBS), General Ledger, LOS, dan ATM Switch.',
    icon: 'i-lucide-git-compare',
    badge: '3-Way Recon',
    badgeColor: 'success' as const,
    to: '/analytics/caatt/reconciliation'
  },
  {
    title: t('navigation.caattPolicy'),
    description: 'Pengujian otomatis kepatuhan operasional perbankan terhadap regulasi OJK/BI dan SOP internal (BMPK, LPS, Dual Control).',
    icon: 'i-lucide-alert-triangle',
    badge: 'Compliance',
    badgeColor: 'error' as const,
    to: '/analytics/caatt/policy'
  },
  {
    title: t('navigation.caattDataQuality'),
    description: 'Profiling kualitas multi-zona data lake (Bronze, Silver, Gold) untuk kelengkapan, akurasi, latensi, dan duplikasi.',
    icon: 'i-lucide-gauge',
    badge: 'Data Lake',
    badgeColor: 'success' as const,
    to: '/analytics/caatt/data-quality'
  }
]
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8 space-y-10">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-6 border-b border-gray-200 dark:border-gray-800">
      <div>
        <h1 class="text-3xl font-black tracking-tight text-gray-900 dark:text-white flex items-center gap-3">
          <UIcon name="i-lucide-pie-chart" class="text-indigo-600 dark:text-indigo-400 w-8 h-8" />
          {{ t('analytics.title') }}
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          {{ t('analytics.subtitle') }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <UBadge color="primary" variant="subtle" size="lg" class="px-3.5 py-1.5 font-bold">
          <UIcon name="i-heroicons-squares-2x2" class="w-4 h-4 mr-1.5" />
          11 Modul Analitik Terintegrasi
        </UBadge>
      </div>
    </div>

    <!-- Section 1: AI Insights -->
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-indigo-500/10 flex items-center justify-center text-indigo-500">
            <UIcon name="i-heroicons-cpu-chip" class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">AI Insights</h2>
            <p class="text-xs text-gray-500">Kecerdasan buatan, predictive modelling, dan deteksi anomali mutakhir</p>
          </div>
        </div>
        <UButton to="/analytics/ai" variant="ghost" color="primary" size="sm" icon="i-heroicons-arrow-right" trailing>
          Dashboard AI
        </UButton>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <NuxtLink
          v-for="mod in aiModules"
          :key="mod.to"
          :to="mod.to"
          class="block group"
        >
          <UCard class="h-full border border-gray-100 dark:border-gray-800 hover:border-indigo-400 dark:hover:border-indigo-600 hover:shadow-md transition-all duration-200 flex flex-col justify-between">
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <div class="w-10 h-10 rounded-lg bg-indigo-500/10 text-indigo-500 flex items-center justify-center group-hover:scale-105 transition-transform">
                  <UIcon :name="mod.icon" class="w-5 h-5" />
                </div>
                <UBadge :color="mod.badgeColor" variant="subtle" size="xs">
                  {{ mod.badge }}
                </UBadge>
              </div>
              <h3 class="font-bold text-base text-gray-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                {{ mod.title }}
              </h3>
              <p class="text-xs text-gray-500 dark:text-gray-400 leading-relaxed line-clamp-3">
                {{ mod.description }}
              </p>
            </div>
            <div class="mt-4 pt-3 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs font-semibold text-indigo-600 dark:text-indigo-400">
              <span>Buka Modul</span>
              <UIcon name="i-heroicons-arrow-right" class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </UCard>
        </NuxtLink>
      </div>
    </div>

    <!-- Section 2: CAATT Insight -->
    <div class="space-y-4 pt-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-500">
            <UIcon name="i-heroicons-chart-bar" class="w-5 h-5" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">CAATT Insight</h2>
            <p class="text-xs text-gray-500">Computer-Assisted Audit Tools & Techniques untuk pengujian 100% populasi data</p>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        <NuxtLink
          v-for="mod in caattModules"
          :key="mod.to"
          :to="mod.to"
          class="block group"
        >
          <UCard class="h-full border border-gray-100 dark:border-gray-800 hover:border-emerald-400 dark:hover:border-emerald-600 hover:shadow-md transition-all duration-200 flex flex-col justify-between">
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <div class="w-10 h-10 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center group-hover:scale-105 transition-transform">
                  <UIcon :name="mod.icon" class="w-5 h-5" />
                </div>
                <UBadge :color="mod.badgeColor" variant="subtle" size="xs">
                  {{ mod.badge }}
                </UBadge>
              </div>
              <h3 class="font-bold text-base text-gray-900 dark:text-white group-hover:text-emerald-600 dark:group-hover:text-emerald-400 transition-colors">
                {{ mod.title }}
              </h3>
              <p class="text-xs text-gray-500 dark:text-gray-400 leading-relaxed line-clamp-3">
                {{ mod.description }}
              </p>
            </div>
            <div class="mt-4 pt-3 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs font-semibold text-emerald-600 dark:text-emerald-400">
              <span>Buka Modul</span>
              <UIcon name="i-heroicons-arrow-right" class="w-4 h-4 group-hover:translate-x-1 transition-transform" />
            </div>
          </UCard>
        </NuxtLink>
      </div>
    </div>
  </div>
</template>
