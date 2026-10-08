<template>
  <div class="space-y-6 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-presentation" class="size-6 sm:size-7 text-primary-500 shrink-0" />
          {{ t('executiveSummary.title') }}
        </h1>
        <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('executiveSummary.subtitle') }}
        </p>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full md:w-auto">
        <UButton
          color="neutral"
          variant="outline"
          icon="i-lucide-upload"
          :label="t('executiveSummary.importButton')"
          to="/executive-summary-compilation/upload"
          class="font-bold shadow w-full sm:w-auto justify-center"
        />
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummary.createNew')"
          class="font-bold w-full sm:w-auto justify-center"
          @click="store.openNewForm(activeQuarter)"
        />
      </div>
    </div>

    <!-- Stats Overview Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-primary-50 dark:bg-primary-950 rounded-lg text-primary-600">
            <UIcon name="i-lucide-file-text" class="size-6" />
          </div>
          <div>
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.stats.total') }}</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.stats.approved') }}</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.stats.draft') }}</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ store.summaryList.filter(s => s.status === 'Draft').length }}
            </div>
          </div>
        </div>
      </UCard>

      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-info-50 dark:bg-info-950 rounded-lg text-info-600">
            <UIcon name="i-lucide-list-checks" class="size-6" />
          </div>
          <div>
            <div class="text-md text-gray-500 dark:text-gray-400">Jumlah Penanganan / Handling</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ totalHandling }}
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
        <div class="flex flex-col md:flex-row gap-3 w-full md:w-auto flex-1">
          <div class="w-full md:w-80">
            <USelectMenu
              v-model="selectedAssignmentLetter"
              :items="assignmentLetterOptions"
              placeholder="Filter berdasarkan Surat Tugas..."
              class="w-full"
            >
              <template #leading>
                <UIcon name="i-heroicons-document-text" class="size-4 text-primary-500" />
              </template>
            </USelectMenu>
          </div>
          <div class="w-full md:w-80">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              :placeholder="t('executiveSummary.searchPlaceholder')"
              class="w-full"
            />
          </div>
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400 shrink-0">
          {{ t('executiveSummary.showingDocuments', { count: filteredSummaries.length }) }}
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
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ item.nomorDokumen || t('executiveSummary.draftReport') }}</h3>
                  <UBadge :color="getStatusColor(item.status)" variant="soft" class="font-semibold uppercase tracking-wider text-[10px]">
                    Status: {{ item.status }}
                  </UBadge>
                  <UBadge v-if="item.assignmentLetterId" color="info" variant="subtle" class="font-mono text-[11px]">
                    <UIcon name="i-lucide-file-signature" class="size-3 mr-1" />
                    {{ item.assignmentLetterId }}
                  </UBadge>
                </div>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  Periode: <span class="font-semibold text-gray-700 dark:text-gray-300">Kuartal {{ item.quarter }} {{ item.tahun }}</span>
                  <span v-if="item.dokumenPath" class="mx-2">•</span>
                  <span v-if="item.dokumenPath" class="inline-flex items-center gap-1 text-primary-600 dark:text-primary-400">
                    <UIcon name="i-lucide-paperclip" class="size-3" />
                    {{ item.dokumenPath }}
                  </span>
                </p>
              </div>
            </div>
            
            <div class="flex gap-2 shrink-0 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity duration-200">
              <UButton color="neutral" variant="ghost" icon="i-lucide-eye" size="sm" @click="store.openView(item)" :title="t('executiveSummary.actions.view')" />
              <UButton v-if="item.status !== 'Approved' || isHigherAuthority" color="primary" variant="ghost" icon="i-lucide-edit" size="sm" @click="store.openEditForm(item as any)" :title="t('executiveSummary.actions.edit')" />
              <UButton v-if="item.status !== 'Approved' || isHigherAuthority" color="error" variant="ghost" icon="i-lucide-trash-2" size="sm" @click="store.deleteSummary(item.id)" :title="t('executiveSummary.actions.delete')" />
              
            </div>
          </div>

          <!-- Stats Overview -->
          <div class="p-5 grid grid-cols-2 md:grid-cols-4 gap-4 divide-y md:divide-y-0 md:divide-x divide-gray-100 dark:divide-gray-800 text-center">
            <div>
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">{{ t('executiveSummary.cardStats.reportsCount') }}</span>
              <span class="block text-2xl font-bold text-gray-800 dark:text-white">{{ item.jumlahLaporan }}</span>
            </div>
            <div class="pt-4 md:pt-0">
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">{{ t('executiveSummary.cardStats.totalFindings') }}</span>
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
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-1">{{ t('executiveSummary.cardStats.totalRecommendations') }}</span>
              <span class="block text-2xl font-bold text-gray-800 dark:text-white">{{ item.jumlahRekomendasi }}</span>
            </div>
            <div class="pt-4 md:pt-0 flex flex-col justify-center items-center">
              <span class="block text-md font-semibold text-gray-400 uppercase tracking-wider mb-2">{{ t('executiveSummary.cardStats.completionClosed') }}</span>
              <div class="w-full max-w-[120px]">
                <div class="flex justify-between text-md font-medium text-gray-700 dark:text-gray-300 mb-1">
                  <span>{{ t('executiveSummary.cardStats.progress') }}</span>
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

          <!-- Note dari Executive untuk Auditor (Hanya muncul jika diisi oleh Admin/CAE/Audit Manager) -->
          <div v-if="item.executiveNote && item.executiveNote.trim()" class="px-5 pb-5 pt-0">
            <div class="bg-amber-50/70 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/60 rounded-xl p-3.5 flex items-start gap-3">
              <div class="p-1.5 bg-amber-100 dark:bg-amber-900/60 rounded-lg text-amber-700 dark:text-amber-400 shrink-0 mt-0.5">
                <UIcon name="i-lucide-message-square-text" class="size-4" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-xs font-bold uppercase tracking-wider text-amber-800 dark:text-amber-300 mb-1 flex items-center gap-1.5">
                  <span>Noted dari Executive untuk Auditor</span>
                </div>
                <p class="text-sm text-amber-900 dark:text-amber-200/90 whitespace-pre-line leading-relaxed">
                  {{ item.executiveNote }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-16 bg-gray-50 dark:bg-gray-900 rounded-lg border border-dashed border-gray-200 dark:border-gray-800">
        <UIcon name="i-lucide-presentation" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300">{{ t('executiveSummary.empty.title') }}</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-1 max-w-sm mx-auto mb-6">
          {{ t('executiveSummary.empty.desc', { quarter: activeQuarter, year: 2026 }) }}
        </p>
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummary.empty.button')"
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
                {{ store.isViewing ? t('executiveSummary.modal.viewTitle') : (store.isEditing ? t('executiveSummary.modal.editTitle') : t('executiveSummary.modal.createTitle')) }}
                <UBadge :color="getStatusColor(store.form.status)" variant="soft" class="ml-2 font-bold uppercase">
                  Status: {{ store.form.status }}
                </UBadge>
              </h3>
              <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="() => { store.showModal = false }" />
            </div>
          </template>

          <!-- Form content handles all sections -->
          <ExecutiveSummaryCompilationForm />
        </UCard>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExecutiveSummaryStore, loadPersistedExecutiveSummaryOverrides } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useI18n } from '~/composables/useI18n'
import { UserRole } from '~/types/auth'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import ExecutiveSummaryCompilationForm from '~/components/audit-result-report/ExecutiveSummaryCompilationForm.vue'

definePageMeta({
  middleware: 'auth'
})

const { t } = useI18n()
const store = useExecutiveSummaryStore()
const authStore = useAuthStore()
const assignmentLetterStore = useAssignmentLetterStore()
const auditReportStore = useAuditResultReportStore()

const activeQuarter = ref(1)
const searchQuery = ref('')
const selectedAssignmentLetter = ref('')

const assignmentLetterOptions = computed(() => {
  const lettersFromStore = assignmentLetterStore.assignmentLetterList
    .map((st: any) => st.letterNumber)
    .filter((num: string) => num && num.startsWith('ST-') && num.includes('2026'))
  const lettersFromReports = auditReportStore.reportList
    .map((r: any) => r.assignmentLetterId)
    .filter((num: string) => num && num.startsWith('ST-') && num.includes('2026'))
  const lettersFromSummaries = store.summaryList
    .map((s: any) => s.assignmentLetterId)
    .filter((num: string) => num && num.startsWith('ST-') && num.includes('2026'))

  const combined = Array.from(new Set([
    'ST-001/SKAI/2026',
    ...lettersFromStore, 
    ...lettersFromReports, 
    ...lettersFromSummaries,
    'ST-002/SKAI/2026', 
    'ST-003/SKAI/2026', 
    'ST-004/SKAI/2026', 
    'ST-005/SKAI/2026'
  ]))
  return ['All Assignment Letters', ...combined]
})

const totalHandling = computed(() => {
  return store.summaryList.reduce((total, summary) => {
    return total + (summary.followUpTable || []).reduce((count, row) => count + Number(row.jumlah || 0), 0)
  }, 0)
})

const quarters = computed(() => [
  { num: 1, label: t('executiveSummary.quarters.q1') },
  { num: 2, label: t('executiveSummary.quarters.q2') },
  { num: 3, label: t('executiveSummary.quarters.q3') },
  { num: 4, label: t('executiveSummary.quarters.q4') }
])

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
  const overrides = loadPersistedExecutiveSummaryOverrides()
  const rawSummaries = store.summaryList.filter(s => 
    s.nomorDokumen !== 'DOC-EXSUM-Q1-2026' &&
    s.nomorDokumen !== '020/LHA/01/KS IAD/2023' &&
    s.nomorDokumen !== '019/LHA/01/KS IAD/2025' &&
    s.nomorDokumen !== 'LHA-020/SKAI/2023' &&
    s.nomorDokumen !== 'LHA-019/SKAI/2025'
  )

  const docMap = new Map<string, ExecutiveSummary>()
  rawSummaries.forEach(s => {
    const override = overrides[s.nomorDokumen] || overrides[s.id]
    const updated = override ? { ...s, ...override } : { ...s }
    if (updated.nomorDokumen) {
      docMap.set(updated.nomorDokumen, updated)
    }
  })

  const summaries = Array.from(docMap.values()).map(updated => {
    // Synchronize jumlahLaporan with individual LHAs count matching quarter or assignment letter
    const matchingIndividualLhas = auditReportStore.reportList.filter(r => {
      if (r.reportNumber === '020/LHA/01/KS IAD/2023' ||
          r.reportNumber === '019/LHA/01/KS IAD/2025' ||
          r.reportNumber === 'LHA-020/SKAI/2023' ||
          r.reportNumber === 'LHA-019/SKAI/2025') return false
      if (updated.assignmentLetterId && r.assignmentLetterId === updated.assignmentLetterId) return true
      const dateParts = r.reportDate ? r.reportDate.split('-') : []
      const m = parseInt(dateParts[1] || '0')
      const q = m <= 3 ? 1 : m <= 6 ? 2 : m <= 9 ? 3 : 4
      return q === updated.quarter
    })

    if (matchingIndividualLhas.length > 0 && (!updated.jumlahLaporan || updated.jumlahLaporan <= 1)) {
      updated.jumlahLaporan = matchingIndividualLhas.length
      const totalRecs = matchingIndividualLhas.reduce((sum, r) => sum + (r.findingsCount || r.findings?.length || 1), 0)
      if (!updated.jumlahRekomendasi) updated.jumlahRekomendasi = totalRecs
    }

    return updated
  })

  return summaries.filter(s => {
    const qMatches = s.quarter === activeQuarter.value
    const stMatches = !selectedAssignmentLetter.value ||
      selectedAssignmentLetter.value === 'All Assignment Letters' ||
      s.assignmentLetterId === selectedAssignmentLetter.value

    const searchLower = searchQuery.value.toLowerCase()
    const matchesSearch = !searchQuery.value ||
      s.nomorDokumen.toLowerCase().includes(searchLower) ||
      (s.assignmentLetterId && s.assignmentLetterId.toLowerCase().includes(searchLower)) ||
      `kuartal ${s.quarter}`.includes(searchLower) ||
      String(s.tahun).includes(searchLower)
    
    return qMatches && stMatches && matchesSearch
  })
})
</script>
