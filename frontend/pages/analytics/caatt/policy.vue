<script setup lang="ts">
import { onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, policyState, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattPolicy')"
      subtitle="Menguji secara otomatis seluruh data operasional terhadap regulasi OJK/BI dan SOP internal (Suku Bunga LPS, BMPK, otorisasi dual control)."
      icon="i-lucide-alert-triangle"
      badgeText="Compliance Automated Testing"
      badgeColor="warning"
    />

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-arrow-path" class="w-12 h-12 animate-spin text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Content -->
    <div v-else class="space-y-6">
      <UAlert
        icon="i-heroicons-shield-check"
        color="primary"
        variant="subtle"
        title="CAATT: Kepatuhan Aturan & Kebijakan Produk Perbankan"
        description="Menguji secara otomatis seluruh data operasional terhadap regulasi OJK/BI dan SOP internal (Suku Bunga LPS, BMPK, otorisasi dual control, dan verifikasi rekening dormant)."
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Pelanggaran</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ policyState.summary.totalViolations }} Kasus</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Aturan terdeteksi dilanggar</div>
        </UCard>
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Tingkat Critical</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ policyState.summary.critical }} Kasus</div>
          <div class="text-[10px] text-rose-400 mt-0.5">Pelanggaran BMPK & LPS</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Tingkat High</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ policyState.summary.high }} Kasus</div>
          <div class="text-[10px] text-amber-400 mt-0.5">Dual Control & Limit Kas</div>
        </UCard>
        <UCard class="border border-emerald-100 dark:border-emerald-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Cabang Terdampak</div>
          <div class="text-2xl font-black text-emerald-600 dark:text-emerald-400 mt-1">{{ policyState.summary.branchesAffected }} Cabang</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Perlu perbaikan kontrol internal</div>
        </UCard>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
            <h3 class="font-bold">Daftar Pelanggaran Kebijakan & Regulasi Produk</h3>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Tanggal</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Nama Kebijakan / SOP</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Cabang</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Nasabah / Entitas</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Severity</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">Uraian Pelanggaran</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in policyState.records" :key="r.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                <td class="py-3 px-3 font-mono font-bold">{{ r.id }}</td>
                <td class="py-3 px-3 text-gray-500 font-mono text-xs">{{ r.testDate }}</td>
                <td class="py-3 px-3 font-bold text-indigo-600 dark:text-indigo-400">{{ r.ruleName }}</td>
                <td class="py-3 px-3">{{ r.branchName }}</td>
                <td class="py-3 px-3 font-semibold">{{ r.customerName }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.severity === 'Critical' ? 'error' : r.severity === 'High' ? 'warning' : 'info'" variant="solid" size="sm">
                    {{ r.severity }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300 leading-relaxed">{{ r.description }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'RESOLVED' ? 'success' : 'warning'" variant="subtle" size="sm">
                    {{ r.status }}
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
