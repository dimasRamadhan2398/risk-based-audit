<template>
  <div class="space-y-6 sm:space-y-8 animate-fade-in pb-12">
    <!-- Header: Executive & Board Profile Context -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4 p-5 sm:p-6 bg-gradient-to-r from-indigo-500/10 via-primary-500/5 to-transparent border border-indigo-500/20 rounded-2xl shadow-xs">
      <div class="flex items-start sm:items-center gap-4">
        <div class="size-14 rounded-2xl bg-indigo-600/10 dark:bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0 shadow-xs">
          <UIcon name="i-lucide-line-chart" class="size-7" />
        </div>
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">
              Dashboard Eksekutif & Manajemen (Executive & Board View)
            </h1>
            <UBadge color="primary" variant="subtle" size="sm" class="font-bold uppercase tracking-wider">
              Role: Executive / Viewer
            </UBadge>
            <span class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-0.5 rounded-full bg-indigo-100 dark:bg-indigo-950/50 text-indigo-800 dark:text-indigo-300 border border-indigo-300/40">
              <UIcon name="i-lucide-eye" class="size-3" />
              Mode Read-Only
            </span>
          </div>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mt-1 max-w-3xl">
            Selamat datang, <strong class="text-gray-900 dark:text-gray-200">{{ currentUserFullName }}</strong>. 
            Ringkasan strategis hasil pengawasan audit internal, ketercapaian PKPT tahunan, tren risiko korporasi, dan efektivitas pengendalian internal bagi Direksi dan Komite Audit.
          </p>
        </div>
      </div>
    </div>

    <!-- Notice Card: Prinsip Tampilan Eksekutif (Read-Only) -->
    <div class="flex items-center justify-between gap-4 p-4 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-900 dark:text-indigo-200 text-xs sm:text-sm">
      <div class="flex items-center gap-3">
        <UIcon name="i-lucide-shield-check" class="size-5 text-indigo-600 dark:text-indigo-400 shrink-0" />
        <div>
          <span class="font-bold">Fokus Pengawasan Eksekutif:</span>
          <span class="ml-1 text-indigo-800 dark:text-indigo-300">
            Laporan makro terintegrasi untuk Dewan Direksi dan Komite Audit. Layar kerja bersifat Read-Only untuk memantau tren temuan, kepatuhan tindak lanjut (CAPA), dan realisasi rencana audit tahunan tanpa tombol input teknis.
          </span>
        </div>
      </div>
      <div class="hidden md:flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-md bg-indigo-600/10 text-indigo-700 dark:text-indigo-300 shrink-0">
        <UIcon name="i-lucide-award" class="size-3.5" />
        <span>Matriks RBAC Executive (Row 3, 5, 8, 9, 10)</span>
      </div>
    </div>

    <!-- 1. 4 Executive High-Level Stat Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
      <!-- Card 1: Realisasi Rencana Audit (PKPT) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-indigo-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Realisasi Audit Tahunan
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-indigo-600 dark:text-indigo-400 mt-1">
              {{ pkptCompletionPercent }}%
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-indigo-50 dark:bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">
            <UIcon name="i-lucide-calendar-check-2" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-indigo-600 dark:text-indigo-400 font-semibold">
            {{ completedPlansCount }} dari {{ totalPlansCount }} Selesai
          </span>
          <span class="text-gray-400">Tahun {{ currentYear }}</span>
        </div>
      </div>

      <!-- Card 2: Temuan Kategori Tinggi (High Risk Findings) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-rose-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Temuan Signifikan / Kritis
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-rose-600 dark:text-rose-400 mt-1">
              {{ highRiskFindingsCount }}
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-rose-50 dark:bg-rose-500/10 text-rose-600 dark:text-rose-400">
            <UIcon name="i-lucide-alert-octagon" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-rose-600 dark:text-rose-400 font-semibold">
            {{ overdueAtrCount }} Lewat Jatuh Tempo
          </span>
          <span class="text-gray-400">Perhatian Direksi</span>
        </div>
      </div>

      <!-- Card 3: Rasio Kepatuhan Tindak Lanjut (ATR Compliance) -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-emerald-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Kepatuhan Tindak Lanjut (CAPA)
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-emerald-600 dark:text-emerald-400 mt-1">
              {{ atrComplianceRate }}%
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <UIcon name="i-lucide-check-circle-2" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-emerald-600 dark:text-emerald-400 font-semibold">
            {{ closedAtrCount }} Rekomendasi Selesai
          </span>
          <span class="text-gray-400">{{ openAtrCount }} Terbuka</span>
        </div>
      </div>

      <!-- Card 4: Efektivitas Pengendalian COSO 2013 -->
      <div class="p-5 rounded-2xl bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 shadow-xs flex flex-col justify-between hover:border-sky-500/40 transition-all">
        <div class="flex items-start justify-between">
          <div>
            <p class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
              Skor Efektivitas Kontrol
            </p>
            <h3 class="text-2xl sm:text-3xl font-extrabold text-sky-600 dark:text-sky-400 mt-1">
              {{ cosoEffectivenessScore }}%
            </h3>
          </div>
          <div class="p-2.5 rounded-xl bg-sky-50 dark:bg-sky-500/10 text-sky-600 dark:text-sky-400">
            <UIcon name="i-lucide-shield" class="size-5" />
          </div>
        </div>
        <div class="mt-3 flex items-center justify-between text-xs border-t border-gray-100 dark:border-gray-800/80 pt-2.5">
          <span class="text-sky-600 dark:text-sky-400 font-semibold">
            Status: {{ cosoRatingLabel }}
          </span>
          <span class="text-gray-400">Standar COSO</span>
        </div>
      </div>
    </div>

    <!-- 2. Section 1: Monitoring Rencana Audit Tahunan (PKPT / Annual Audit Plan) -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-calendar-days" class="size-5 text-indigo-500" />
            1. Progres Pelaksanaan Rencana Kerja Audit Tahunan (PKPT {{ currentYear }})
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Tingkat ketercapaian penugasan audit operasional, kepatuhan, dan investigasi terhadap rencana audit tahunan.
          </p>
        </div>

        <NuxtLink
          to="/annual-audit"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Lihat Seluruh Rencana Tahunan</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div
          v-for="plan in displayedPlans"
          :key="plan.id"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 space-y-3"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono font-bold text-indigo-600 dark:text-indigo-400 text-xs">{{ plan.code }}</span>
            <UBadge
              :color="plan.status === 'Completed' ? 'success' : (plan.status === 'In Progress' ? 'info' : 'neutral')"
              variant="subtle"
              size="xs"
              class="font-semibold"
            >
              {{ plan.status }}
            </UBadge>
          </div>
          <h4 class="font-bold text-gray-900 dark:text-white text-xs line-clamp-1">{{ plan.title }}</h4>
          <div class="space-y-1">
            <div class="flex justify-between text-[11px] text-gray-500">
              <span>Progres Audit</span>
              <span class="font-bold text-gray-700 dark:text-gray-300">{{ plan.progress || 0 }}%</span>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 h-1.5 rounded-full overflow-hidden">
              <div
                class="bg-indigo-600 dark:bg-indigo-500 h-full rounded-full transition-all duration-300"
                :style="{ width: `${plan.progress || 0}%` }"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 3. Section 2: Ringkasan Temuan Signifikan & LHA Terbit -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Laporan Hasil Audit (LHA) Final Terbit -->
      <div class="lg:col-span-6 p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-file-text" class="size-5 text-emerald-500" />
            Laporan Hasil Audit (LHA) Final
          </h3>
          <NuxtLink to="/audit-result-report" class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 hover:underline">
            Rincian &rarr;
          </NuxtLink>
        </div>
        <p class="text-xs text-gray-500">Daftar laporan hasil audit yang telah selesai diuji dan diterbitkan kepada auditee dan direksi.</p>

        <div class="space-y-2.5">
          <div
            v-for="rep in publishedReports.slice(0, 4)"
            :key="rep.id"
            class="p-3.5 rounded-xl border border-gray-100 dark:border-gray-800 bg-gray-50/60 dark:bg-gray-800/40 flex items-center justify-between gap-3 text-xs"
          >
            <div class="min-w-0 space-y-0.5">
              <span class="font-mono text-emerald-600 dark:text-emerald-400 font-bold text-[11px] block">{{ rep.reportNumber || (rep as any).report_number || (rep as any).letter_no }}</span>
              <p class="font-bold text-gray-900 dark:text-white truncate">{{ rep.reportTitle || (rep as any).title || (rep as any).audit_title }}</p>
              <span class="text-gray-400 text-[11px] block">Auditee: {{ (rep as any).auditee_name || (rep as any).department || '-' }}</span>
            </div>
            <UBadge color="success" variant="subtle" size="xs" class="font-bold shrink-0">
              Final / Published
            </UBadge>
          </div>

          <div v-if="publishedReports.length === 0" class="py-8 text-center text-gray-400 text-xs">
            Belum ada laporan LHA final yang diterbitkan.
          </div>
        </div>
      </div>

      <!-- Ringkasan Temuan Signifikan (High Risk Findings) -->
      <div class="lg:col-span-6 p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-alert-triangle" class="size-5 text-rose-500" />
            Temuan Signifikan Perlu Perhatian
          </h3>
          <NuxtLink to="/action-taken-report" class="text-xs font-semibold text-rose-600 dark:text-rose-400 hover:underline">
            Semua CAPA &rarr;
          </NuxtLink>
        </div>
        <p class="text-xs text-gray-500">Temuan dengan kategori Very Significant atau Significant yang memerlukan eskalasi manajemen.</p>

        <div class="space-y-2.5">
          <div
            v-for="f in significantFindings.slice(0, 4)"
            :key="f.id"
            class="p-3.5 rounded-xl border border-rose-100 dark:border-rose-900/30 bg-rose-50/30 dark:bg-rose-950/20 flex items-center justify-between gap-3 text-xs"
          >
            <div class="min-w-0 space-y-0.5">
              <div class="flex items-center gap-1.5">
                <UBadge color="error" variant="subtle" size="xs" class="font-bold text-[10px]">
                  {{ f.finding_category || 'Significant' }}
                </UBadge>
                <span class="font-mono text-gray-500 text-[11px]">{{ f.report_number }}</span>
              </div>
              <p class="font-bold text-gray-900 dark:text-white truncate">{{ f.finding_title }}</p>
              <span class="text-gray-400 text-[11px] block">PIC: {{ f.pic_name || 'Unit PIC' }} · Due: {{ f.due_date || '-' }}</span>
            </div>
            <div class="text-right shrink-0">
              <span class="font-bold text-rose-600 text-xs block">{{ f.progress || 0 }}%</span>
              <span class="text-[10px] text-gray-400">Progres CAPA</span>
            </div>
          </div>

          <div v-if="significantFindings.length === 0" class="py-8 text-center text-gray-400 text-xs">
            Tidak ada temuan berisiko tinggi yang aktif saat ini.
          </div>
        </div>
      </div>
    </div>

    <!-- 4. Section 3: Ringkasan Pengendalian Internal COSO 2013 -->
    <div class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-shield-check" class="size-5 text-sky-500" />
            3. Ringkasan Pengendalian Internal Berbasis COSO 2013
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Evaluasi kesehatan sistem pengendalian internal korporasi pada 5 pilar utama COSO Framework.
          </p>
        </div>

        <NuxtLink
          to="/risk-profile/risk-control-matrix"
          class="text-xs font-bold text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Risk Control Matrix Penuh</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-5 gap-3">
        <div
          v-for="dim in cosoDimensionList"
          :key="dim.key"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 text-center space-y-1.5"
        >
          <span class="text-xs font-semibold text-gray-500 block truncate">{{ dim.name }}</span>
          <h4 class="text-xl font-extrabold text-sky-600 dark:text-sky-400">{{ dim.score }}%</h4>
          <div class="w-full bg-gray-200 dark:bg-gray-700 h-1 rounded-full overflow-hidden">
            <div class="bg-sky-500 h-full rounded-full" :style="{ width: `${dim.score}%` }" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useRCMStore } from '~/stores/rcm'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

const authStore = useAuthStore()
const toast = useToastNotification()
const { isAdmin, canManageAudits } = useRbac()

const annualPlanStore = useAnnualPlanStore()
const atrStore = useActionTakenReportStore()
const auditResultStore = useAuditResultReportStore()
const riskProfileStore = useRiskProfileStore()
const rcmStore = useRCMStore()

const isSyncing = ref(false)
let isMounted = false

onBeforeUnmount(() => {
  isMounted = false
})

const currentUserFullName = computed(() => {
  return authStore.user?.fullName || authStore.user?.username || 'Executive Officer'
})



const currentYear = computed(() => new Date().getFullYear())

// --- ANNUAL PLAN METRICS ---
const totalPlansCount = computed(() => annualPlanStore.plans?.length || 0)
const completedPlansCount = computed(() => {
  return (annualPlanStore.plans || []).filter((p: any) => p.status === 'Completed').length
})
const pkptCompletionPercent = computed(() => {
  if (totalPlansCount.value === 0) return 0
  return Math.round((completedPlansCount.value / totalPlansCount.value) * 100)
})

const displayedPlans = computed(() => {
  return (annualPlanStore.plans || []).slice(0, 3).map((p: any) => ({
    id: p.id,
    code: p.code || 'PKPT',
    title: p.title || p.auditTitle || 'Rencana Audit Tahunan',
    status: p.status || 'Planned',
    progress: p.status === 'Completed' ? 100 : (p.status === 'In Progress' ? 50 : 0)
  }))
})

// --- ACTION TAKEN REPORT METRICS ---
const allAtrItems = computed(() => {
  return atrStore.summaryItems.length > 0 ? atrStore.summaryItems : atrStore.items || []
})

const totalAtrCount = computed(() => allAtrItems.value.length)
const closedAtrCount = computed(() => {
  return allAtrItems.value.filter((i: any) => String(i.status || '').toUpperCase() === 'CLOSED').length
})
const openAtrCount = computed(() => totalAtrCount.value - closedAtrCount.value)

const atrComplianceRate = computed(() => {
  if (totalAtrCount.value === 0) return 100
  return Math.round((closedAtrCount.value / totalAtrCount.value) * 100)
})

const overdueAtrCount = computed(() => {
  return allAtrItems.value.filter((i: any) => i.is_overdue || String(i.status || '').toUpperCase() === 'OVERDUE').length
})

const highRiskFindingsCount = computed(() => {
  return allAtrItems.value.filter((i: any) => {
    const cat = String(i.finding_category || '').toLowerCase()
    return cat.includes('very') || cat.includes('significant') || cat.includes('high')
  }).length
})

const significantFindings = computed(() => {
  return allAtrItems.value.filter((i: any) => {
    const cat = String(i.finding_category || '').toLowerCase()
    return cat.includes('very') || cat.includes('significant') || cat.includes('high')
  })
})

// --- AUDIT RESULT REPORTS ---
const publishedReports = computed(() => {
  return (auditResultStore.reportList || []).filter((r: any) => {
    return r.status === 'Final' || r.status === 'Published'
  })
})

// --- COSO 2013 CONTROLS ---
const cosoEffectivenessScore = computed(() => {
  return rcmStore.internalControlEffectiveness || rcmStore.cosoAverages?.totalWeighted || 78
})

const cosoRatingLabel = computed(() => {
  return rcmStore.effectivenessRating?.rating || 'Effective'
})

const cosoDimensionList = computed(() => [
  { key: 'dim1', name: 'Control Environment', score: rcmStore.cosoAverages?.design || 80 },
  { key: 'dim2', name: 'Risk Assessment', score: rcmStore.cosoAverages?.operating || 75 },
  { key: 'dim3', name: 'Control Activities', score: rcmStore.cosoAverages?.coverage || 82 },
  { key: 'dim4', name: 'Info & Communication', score: rcmStore.cosoAverages?.timeliness || 76 },
  { key: 'dim5', name: 'Monitoring Activities', score: rcmStore.cosoAverages?.automation || 74 },
])

const loadInitialExecutiveData = async () => {
  try {
    await Promise.allSettled([
      annualPlanStore.fetchPlans(),
      atrStore.fetchSummary(),
      atrStore.fetchReports(),
      auditResultStore.fetchReports(),
      riskProfileStore.fetchRisks(),
      rcmStore.fetchRCMList()
    ])
  } catch (err) {
    console.warn('Silent executive dashboard data load notice:', err)
  }
}

const manualSyncExecutiveData = async () => {
  isSyncing.value = true
  try {
    await Promise.allSettled([
      annualPlanStore.fetchPlans(),
      atrStore.fetchSummary(),
      atrStore.fetchReports(),
      auditResultStore.fetchReports(),
      riskProfileStore.fetchRisks(),
      rcmStore.fetchRCMList()
    ])
    if (isMounted) toast.success('Data eksekutif berhasil disinkronkan!')
  } catch (err) {
    if (isMounted) toast.error('Gagal memperbarui data eksekutif')
  } finally {
    if (isMounted) isSyncing.value = false
  }
}

onMounted(async () => {
  isMounted = true
  await loadInitialExecutiveData()
})
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
