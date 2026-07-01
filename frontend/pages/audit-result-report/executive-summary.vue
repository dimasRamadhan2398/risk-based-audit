<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-presentation" class="size-7 text-primary-500" />
          Executive Summary (LHA Kompilasi)
        </h1>
        <p class="text-gray-500 dark:text-gray-400 mt-1">
          Kompilasi Laporan Hasil Audit triwulanan untuk BOD & Komite Audit.
        </p>
      </div>
      <UButton
        color="primary"
        icon="i-lucide-plus"
        label="Buat Laporan Kompilasi Baru"
        class="font-bold"
        @click="store.openNewForm(activeQuarter)"
      />
    </div>

    <!-- Stats Overview Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-primary-50 dark:bg-primary-950 rounded-lg text-primary-600">
            <UIcon name="i-lucide-file-text" class="size-6" />
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">Total Kompilasi</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">{{ store.summaryList.length }}</div>
          </div>
        </div>
      </UCard>
      
      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-success-50 dark:bg-success-950 rounded-lg text-success-600">
            <UIcon name="i-lucide-check-circle" class="size-6" />
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">Disetujui (Approved)</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ store.summaryList.filter(s => s.status === 'Approved').length }}
            </div>
          </div>
        </div>
      </UCard>

      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-warning-50 dark:bg-warning-950 rounded-lg text-warning-600">
            <UIcon name="i-lucide-edit-3" class="size-6" />
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">Draft</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ store.summaryList.filter(s => s.status === 'Draft').length }}
            </div>
          </div>
        </div>
      </UCard>

      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-error-50 dark:bg-error-950 rounded-lg text-error-600">
            <UIcon name="i-lucide-alert-triangle" class="size-6" />
          </div>
          <div>
            <div class="text-xs text-gray-500 dark:text-gray-400">Ditolak (Rejected)</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ store.summaryList.filter(s => s.status === 'Rejected').length }}
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Quarter Tabs -->
    <UCard class="overflow-hidden">
      <template #header>
        <div class="flex justify-between items-center">
          <div class="flex border-b border-gray-200 dark:border-gray-800 w-full">
            <button
              v-for="q in quarters"
              :key="q.num"
              @click="activeQuarter = q.num"
              class="px-6 py-3 font-semibold text-sm transition-all border-b-2"
              :class="activeQuarter === q.num 
                ? 'border-primary-500 text-primary-600 dark:text-primary-400 font-bold' 
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'"
            >
              {{ q.label }}
            </button>
          </div>
        </div>
      </template>

      <!-- Search and Filters -->
      <div class="mb-4 flex flex-col md:flex-row gap-4 items-center justify-between">
        <div class="w-full md:w-80">
          <UInput
            v-model="searchQuery"
            icon="i-lucide-search"
            placeholder="Cari nomor dokumen atau bulan..."
            class="w-full"
          />
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400">
          Menampilkan <span class="font-semibold">{{ filteredSummaries.length }}</span> dokumen
        </div>
      </div>

      <!-- Document History Table -->
      <div v-if="filteredSummaries.length > 0" class="overflow-x-auto">
        <UTable :data="filteredSummaries" :columns="columns" class="min-w-full">
          <template #periodeBulan-data="{ row }">
            <span class="font-medium text-gray-800 dark:text-gray-200">{{ row.original.periodeBulan }} {{ row.original.tahun }}</span>
          </template>

          <template #nomorDokumen-data="{ row }">
            <div class="flex flex-col">
              <span class="font-mono text-xs font-semibold text-primary-600 dark:text-primary-400">{{ row.original.nomorDokumen }}</span>
              <span v-if="row.original.dokumenPath" class="text-xs text-gray-400 flex items-center gap-1 mt-0.5">
                <UIcon name="i-lucide-paperclip" class="size-3" />
                {{ row.original.dokumenPath }}
              </span>
            </div>
          </template>

          <template #status-data="{ row }">
            <UBadge :color="getStatusColor(row.original.status)" variant="subtle" class="font-semibold capitalize">
              {{ row.original.status }}
            </UBadge>
          </template>

          <template #metrics-data="{ row }">
            <div class="flex flex-col text-xs text-gray-500 gap-0.5">
              <span>Laporan: <strong class="text-gray-700 dark:text-gray-300">{{ row.original.jumlahLaporan }}</strong></span>
              <span>Temuan: <strong class="text-gray-700 dark:text-gray-300">{{ row.original.risikoTinggi + row.original.risikoSedang + row.original.risikoRendah }}</strong></span>
              <span>Rekomendasi: <strong class="text-gray-700 dark:text-gray-300">{{ row.original.jumlahRekomendasi }}</strong></span>
            </div>
          </template>

          <template #actions-data="{ row }">
            <div class="flex items-center gap-2">
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-eye"
                size="sm"
                @click="store.openView(row.original as any)"
                title="Lihat Detail"
              />
              <UButton
                v-if="row.original.status !== 'Approved' || isHigherAuthority"
                color="primary"
                variant="ghost"
                icon="i-lucide-edit"
                size="sm"
                @click="store.openEditForm(row.original as any)"
                title="Edit Laporan"
              />
              <UButton
                v-if="row.original.status !== 'Approved' || isHigherAuthority"
                color="error"
                variant="ghost"
                icon="i-lucide-trash-2"
                size="sm"
                @click="store.deleteSummary(row.original.id)"
                title="Hapus"
              />
              
              <!-- Quick Workflow Actions for Chief SPI (CAE) -->
              <span v-if="row.original.status === 'Draft' && isChiefAuditExecutive" class="border-l border-gray-200 pl-2 ml-1 flex gap-1">
                <UButton
                  color="success"
                  variant="soft"
                  icon="i-lucide-check"
                  size="xs"
                  label="Approve"
                  @click="store.updateStatus(row.original.id, 'Approved')"
                />
              </span>
              <span v-if="row.original.status === 'Approved' && isHigherAuthority" class="border-l border-gray-200 pl-2 ml-1 flex gap-1">
                <UButton
                  color="warning"
                  variant="soft"
                  icon="i-lucide-unlock"
                  size="xs"
                  label="Revert Draft"
                  @click="store.updateStatus(row.original.id, 'Draft')"
                />
              </span>
            </div>
          </template>
        </UTable>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-16 bg-gray-50 dark:bg-gray-900 rounded-lg border border-dashed border-gray-200 dark:border-gray-800">
        <UIcon name="i-lucide-presentation" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300">Belum ada Laporan Kompilasi</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-1 max-w-sm mx-auto mb-6">
          Belum ada berkas Executive Summary yang diunggah untuk Triwulan {{ activeQuarter }} (2026).
        </p>
        <UButton
          color="primary"
          icon="i-lucide-plus"
          label="Mulai Laporan Pertama"
          @click="store.openNewForm(activeQuarter)"
        />
      </div>
    </UCard>

    <!-- Modal Form (Create / Edit / View) -->
    <UModal v-model="store.showModal" fullscreen :prevent-close="store.loading">
      <UCard class="flex flex-col h-full" :ui="{ body: 'flex-1 overflow-y-auto p-0', footer: 'border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900' }">
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-presentation" class="size-5 text-primary-500" />
              {{ store.isViewing ? 'Detail Laporan Kompilasi' : (store.isEditing ? 'Edit Laporan Kompilasi' : 'Buat Laporan Kompilasi Baru') }}
              <UBadge v-if="store.currentSummary" :color="getStatusColor(store.form.status)" variant="soft" class="ml-2 font-bold uppercase">
                {{ store.form.status }}
              </UBadge>
            </h3>
            <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="store.showModal = false" />
          </div>
        </template>

        <!-- Form content handles all sections -->
        <ExecutiveSummaryForm />
      </UCard>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { UserRole } from '~/types/auth'
import ExecutiveSummaryForm from '~/components/audit-result-report/ExecutiveSummaryForm.vue'

definePageMeta({
  middleware: 'auth'
})

const store = useExecutiveSummaryStore()
const authStore = useAuthStore()

const activeQuarter = ref(1)
const searchQuery = ref('')

const quarters = [
  { num: 1, label: 'Triwulan 1 (Q1)' },
  { num: 2, label: 'Triwulan 2 (Q2)' },
  { num: 3, label: 'Triwulan 3 (Q3)' },
  { num: 4, label: 'Triwulan 4 (Q4)' }
]

const columns = [
  { accessorKey: 'periodeBulan', header: 'Periode Bulan' },
  { accessorKey: 'nomorDokumen', header: 'Nomor Dokumen' },
  { accessorKey: 'metrics', header: 'Statistik Laporan' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'actions', header: 'Aksi' }
]

// Role checks
const isChiefAuditExecutive = computed(() => {
  return authStore.user?.roles.includes(UserRole.CHIEF_AUDIT_EXECUTIVE) || authStore.user?.roles.includes(UserRole.ADMIN)
})

const isHigherAuthority = computed(() => {
  // Komite audit mapped as admin or explicit audit_committee role
  return authStore.user?.roles.includes(UserRole.ADMIN) || authStore.user?.roles.includes('audit_committee')
})

const getStatusColor = (status: string) => {
  switch (status) {
    case 'Approved': return 'success'
    case 'Rejected': return 'error'
    case 'Draft': return 'warning'
    default: return 'neutral'
  }
}

const filteredSummaries = computed(() => {
  return store.summaryList.filter(s => {
    const qMatches = s.quarter === activeQuarter.value
    const searchLower = searchQuery.value.toLowerCase()
    const matchesSearch = !searchQuery.value ||
      s.nomorDokumen.toLowerCase().includes(searchLower) ||
      s.periodeBulan.toLowerCase().includes(searchLower)
    
    return qMatches && matchesSearch
  })
})
</script>
