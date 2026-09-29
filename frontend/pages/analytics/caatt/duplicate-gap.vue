<script setup lang="ts">
import { onMounted } from 'vue'
import { useCaattAnalytics } from '~/composables/useCaattAnalytics'
import { useI18n } from '~/composables/useI18n'
import CaattHeader from '~/components/analytics/CaattHeader.vue'

const { t } = useI18n()
const { loading, dupGapState, formatIDR, fetchCaattAnalytics } = useCaattAnalytics()

onMounted(() => {
  fetchCaattAnalytics()
})
</script>

<template>
  <div class="max-w-full mx-auto py-8 px-4 sm:px-6 lg:px-8">
    <CaattHeader
      :title="t('navigation.caattDuplicateGap')"
      subtitle="Menemukan transaksi bernilai identik yang diproses berulang serta mendeteksi celah nomor warkat/voucher yang hilang (gap)."
      icon="i-lucide-copy-x"
      badgeText="Audit Integritas Data"
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
        icon="i-heroicons-document-duplicate"
        color="primary"
        variant="subtle"
        title="CAATT: Integritas Data — Deteksi Duplikasi & Celah Nomor Dokumen"
        description="Menemukan transaksi bernilai identik yang diproses berulang dalam selang waktu mencurigakan serta mengidentifikasi nomor warkat/voucher General Ledger yang hilang berurutan (gap detection)."
      />

      <!-- Summary Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <UCard class="border border-rose-100 dark:border-rose-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Total Transaksi Duplikat</div>
          <div class="text-2xl font-black text-rose-600 dark:text-rose-400 mt-1">{{ dupGapState.summary.totalDuplicates }} Kasus</div>
          <div class="text-[10px] text-rose-400 mt-0.5">Potensi transaksi ganda / split bill</div>
        </UCard>
        <UCard class="border border-amber-100 dark:border-amber-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Celah Nomor Dokumen (Gap)</div>
          <div class="text-2xl font-black text-amber-600 dark:text-amber-400 mt-1">{{ dupGapState.summary.totalGaps }} Nomor Hilang</div>
          <div class="text-[10px] text-amber-400 mt-0.5">Voucher GL & Warkat Kliring tidak urut</div>
        </UCard>
        <UCard class="border border-indigo-100 dark:border-indigo-900/50">
          <div class="text-[10px] font-bold uppercase tracking-widest text-gray-400">Cabang Terdampak</div>
          <div class="text-2xl font-black text-indigo-600 dark:text-indigo-400 mt-1">{{ dupGapState.summary.branchesAffected }} Cabang</div>
          <div class="text-[10px] text-gray-400 mt-0.5">Perlu investigasi operasional</div>
        </UCard>
      </div>

      <!-- Table -->
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <UIcon name="i-heroicons-table-cells" class="text-indigo-500" />
              <h3 class="font-bold">Daftar Temuan Duplikasi & Celah Nomor Urut (Gap)</h3>
            </div>
          </div>
        </template>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">ID</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Tipe</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Cabang</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">No. Referensi / Voucher</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">No. Akun</th>
                <th class="text-right py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Nominal</th>
                <th class="text-left py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400 min-w-[260px]">Uraian Temuan</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Frekuensi</th>
                <th class="text-center py-3 px-3 font-bold text-[10px] uppercase tracking-widest text-gray-400">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in dupGapState.records" :key="r.id" class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="py-3 px-3 font-mono font-bold">{{ r.id }}</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.resultType === 'DUPLICATE' ? 'error' : 'warning'" variant="solid" size="sm">
                    {{ r.resultType }}
                  </UBadge>
                </td>
                <td class="py-3 px-3 font-semibold">{{ r.branchName }}</td>
                <td class="py-3 px-3 font-mono font-bold">{{ r.referenceNo }}</td>
                <td class="py-3 px-3 font-mono text-xs text-gray-500">{{ r.accountId }}</td>
                <td class="py-3 px-3 text-right font-mono font-bold">{{ r.amount > 0 ? formatIDR(r.amount) : '-' }}</td>
                <td class="py-3 px-3 text-gray-600 dark:text-gray-300 leading-relaxed">{{ r.description }}</td>
                <td class="text-center py-3 px-3 font-mono font-bold">{{ r.occurrences }}x</td>
                <td class="text-center py-3 px-3">
                  <UBadge :color="r.status === 'OPEN' ? 'error' : r.status === 'INVESTIGATING' ? 'warning' : 'success'" variant="subtle" size="sm">
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
