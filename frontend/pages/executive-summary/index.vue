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
      <div class="flex items-center gap-2">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-upload"
          label="Import Document"
          to="/audit-result-report/executive-summary-upload"
          class="font-bold shadow"
        />
        <UButton
          color="primary"
          icon="i-lucide-plus"
          label="Buat Laporan Kompilasi Baru"
          class="font-bold"
          @click="store.openNewForm(activeQuarter)"
        />
      </div>
    </div>

    <!-- Stats Overview Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-primary-50 dark:bg-primary-950 rounded-lg text-primary-600">
            <UIcon name="i-lucide-file-text" class="size-6" />
          </div>
          <div>
            <div class="text-md text-gray-500 dark:text-gray-400">Total Kompilasi</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">Disetujui (Approved)</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">Draft</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">Ditolak (Rejected)</div>
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

      <!-- Document History Cards -->
      <div v-if="filteredSummaries.length > 0" class="space-y-4">
        <div 
          v-for="item in filteredSummaries" 
          :key="item.id" 
          class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all duration-300 group"
        >
          <!-- Header -->
          <div class="p-5 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 bg-gray-50/50 dark:bg-gray-800/50 border-b border-gray-100 dark:border-gray-800">
            <div class="flex items-start gap-4">
              <div class="p-3 bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 text-primary-500">
                <UIcon name="i-lucide-file-text" class="size-6" />
              </div>
              <div>
                <div class="flex items-center gap-2 flex-wrap">
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ item.nomorDokumen || 'Draft Laporan' }}</h3>
                  <UBadge :color="getStatusColor(item.status)" variant="soft" class="font-semibold uppercase tracking-wider text-[10px]">
                    {{ item.status }}
                  </UBadge>
                </div>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  Periode: <span class="font-semibold text-gray-700 dark:text-gray-300">{{ item.periodeBulan }} {{ item.tahun }}</span>
                  <span v-if="item.dokumenPath" class="mx-2">•</span>
                  <span v-if="item.dokumenPath" class="inline-flex items-center gap-1 text-primary-600 dark:text-primary-400">
                    <UIcon name="i-lucide-paperclip" class="size-3" />
                    {{ item.dokumenPath }}
                  </span>
                </p>
              </div>
            </div>
            
            <div class="flex gap-2 shrink-0 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity duration-200">
              <UButton color="neutral" variant="ghost" icon="i-lucide-eye" size="sm" @click="store.openView(item)" title="Lihat Detail" />
              <UButton v-if="item.status !== 'Approved' || isHigherAuthority" color="primary" variant="ghost" icon="i-lucide-edit" size="sm" @click="store.openEditForm(item as any)" title="Edit Laporan" />
              <UButton v-if="item.status !== 'Approved' || isHigherAuthority" color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" @click="store.deleteSummary(item.id)" title="Hapus" />
              
              <!-- Quick Workflow Actions -->
              <div v-if="(item.status === 'Draft' && isChiefAuditExecutive) || (item.status === 'Approved' && isHigherAuthority)" class="border-l border-gray-200 dark:border-gray-700 pl-2 ml-1 flex gap-1">
                <UButton v-if="item.status === 'Draft' && isChiefAuditExecutive" color="success" variant="soft" icon="i-lucide-check" size="sm" label="Approve" @click="store.updateStatus(item.id, 'Approved')" />
                <UButton v-if="item.status === 'Approved' && isHigherAuthority" color="warning" variant="soft" icon="i-lucide-unlock" size="sm" label="Revert Draft" @click="store.updateStatus(item.id, 'Draft')" />
              </div>
            </div>
          </div>

          <!-- Stats Overview -->
          <div class="p-5 grid grid-cols-2 md:grid-cols-4 gap-4 divide-y md:divide-y-0 md:divide-x divide-gray-100 dark:divide-gray-800 text-center">
            <div>
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">Jumlah LHA</span>
              <span class="block text-2xl font-bold text-gray-800 dark:text-white">{{ item.jumlahLaporan }}</span>
            </div>
            <div class="pt-4 md:pt-0">
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">Total Temuan</span>
              <span class="block text-2xl font-bold text-gray-800 dark:text-white">
                {{ Number(item.risikoTinggi) + Number(item.risikoSedang) + Number(item.risikoRendah) }}
              </span>
              <div class="flex justify-center gap-2 mt-1.5 text-md">
                <span class="text-error-600 font-semibold bg-error-50 dark:bg-error-950/50 px-1.5 rounded">{{ item.risikoTinggi }} H</span>
                <span class="text-warning-600 font-semibold bg-warning-50 dark:bg-warning-950/50 px-1.5 rounded">{{ item.risikoSedang }} M</span>
                <span class="text-success-600 font-semibold bg-success-50 dark:bg-success-950/50 px-1.5 rounded">{{ item.risikoRendah }} L</span>
              </div>
            </div>
            <div class="pt-4 md:pt-0">
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">Total Rekomendasi</span>
              <span class="block text-2xl font-bold text-gray-800 dark:text-white">{{ item.jumlahRekomendasi }}</span>
            </div>
            <div class="pt-4 md:pt-0 flex flex-col justify-center items-center">
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-2">Penyelesaian (Closed)</span>
              <div class="w-full max-w-[120px]">
                <div class="flex justify-between text-md font-medium text-gray-700 dark:text-gray-300 mb-1">
                  <span>Progress</span>
                  <span class="text-success-600">{{ item.followUpTable?.[0]?.jumlah ? Math.round((Number(item.followUpTable[0].jumlah) / ((Number(item.followUpTable[0].jumlah) || 0) + (Number(item.followUpTable[1]?.jumlah) || 0) + (Number(item.followUpTable[2]?.jumlah) || 0))) * 100) : 0 }}%</span>
                </div>
                <div class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-2">
                  <div class="bg-success-500 h-2 rounded-full transition-all" :style="{ width: `${item.followUpTable?.[0]?.jumlah ? Math.round((Number(item.followUpTable[0].jumlah) / ((Number(item.followUpTable[0].jumlah) || 0) + (Number(item.followUpTable[1]?.jumlah) || 0) + (Number(item.followUpTable[2]?.jumlah) || 0))) * 100) : 0}%` }"></div>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Narrative Snippet -->
          <div v-if="item.narrative" class="px-5 pb-5 pt-0">
            <div class="bg-gray-50 dark:bg-gray-800/30 rounded-lg p-3 text-sm text-gray-600 dark:text-gray-400 border border-gray-100 dark:border-gray-800">
              <UIcon name="i-lucide-quote" class="size-4 text-gray-300 mr-2 inline-block -mt-1" />
              <span class="italic line-clamp-2">{{ item.narrative }}</span>
            </div>
          </div>
        </div>
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
    <UModal v-model:open="store.showModal" fullscreen :prevent-close="store.loading">
      <template #content>
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
              <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="() => { store.showModal = false }" />
            </div>
          </template>

          <!-- Form content handles all sections -->
          <ExecutiveSummaryForm />
        </UCard>
      </template>
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
