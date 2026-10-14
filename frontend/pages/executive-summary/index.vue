<template>
  <div class="space-y-6 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-file-text" class="size-6 sm:size-7 text-primary-500 shrink-0" />
          {{ t('navigation.executiveSummaryIndividual') }}
        </h1>
        <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('executiveSummaryIndividual.page.description') }}
        </p>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full md:w-auto">
        <UButton
          color="info"
          variant="ghost"
          icon="i-lucide-upload"
          :label="t('executiveSummaryIndividual.page.importButton')"
          to="/executive-summary/upload"
          class="font-bold shadow w-full sm:w-auto justify-center"
        />
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummaryIndividual.page.createButton')"
          class="font-bold w-full sm:w-auto justify-center"
          @click="store.openNewForm(1)"
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummaryIndividual.stats.total') }}</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">{{ syncedSummaries.length }}</div>
          </div>
        </div>
      </UCard>
      
      <UCard :ui="{ body: 'p-4' }">
        <div class="flex items-center gap-3">
          <div class="p-3 bg-success-50 dark:bg-success-950 rounded-lg text-success-600">
            <UIcon name="i-lucide-check-circle" class="size-6" />
          </div>
          <div>
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummaryIndividual.stats.approved') }}</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ syncedSummaries.filter(s => s.status === 'Approved').length }}
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummaryIndividual.stats.draft') }}</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ syncedSummaries.filter(s => s.status === 'Draft').length }}
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummaryIndividual.stats.handling') }}</div>
            <div class="text-2xl font-bold text-gray-800 dark:text-white">
              {{ totalHandling }}
            </div>
          </div>
        </div>
      </UCard>
    </div>

    <!-- Main List & Filter Card -->
    <UCard class="overflow-hidden">
      <!-- Search and Filters -->
      <div class="mb-4 flex flex-col md:flex-row gap-4 items-center justify-between">
        <div class="flex flex-col md:flex-row gap-3 w-full md:w-auto flex-1">
          <div class="w-full md:w-80">
            <USelectMenu
              v-model="selectedLhaNumber"
              :items="lhaNumberOptions"
              placeholder="Filter berdasarkan Nomor LHA..."
              class="w-full"
            >
              <template #leading>
                <UIcon name="i-lucide-file-text" class="size-4 text-primary-500" />
              </template>
            </USelectMenu>
          </div>
          <div class="w-full md:w-80">
            <UInput
              v-model="searchQuery"
              icon="i-lucide-search"
              :placeholder="t('executiveSummaryIndividual.filters.search')"
              class="w-full"
            />
          </div>
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400 shrink-0">
          {{ t('executiveSummaryIndividual.filters.showingPrefix') }} <span class="font-semibold">{{ filteredSummaries.length }}</span> {{ t('executiveSummaryIndividual.filters.showingSuffix') }}
        </div>
      </div>

      <!-- Document Cards List -->
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
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ item.nomorDokumen || t('executiveSummaryIndividual.card.draftLha') }}</h3>
                  <UBadge :color="getStatusColor(item.status)" variant="soft" class="font-semibold uppercase tracking-wider text-[10px]">
                    {{ t('executiveSummaryIndividual.card.status', { status: item.status }) }}
                  </UBadge>
                  <UBadge v-if="item.assignmentLetterId" color="info" variant="subtle" class="font-mono text-[11px]">
                    {{ t('executiveSummaryForm.upload.assignmentLetter', { number: item.assignmentLetterId }) }}
                  </UBadge>
                </div>
                <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  <span v-if="item.dokumenPath" class="inline-flex items-center gap-1 text-primary-600 dark:text-primary-400">
                    <UIcon name="i-lucide-paperclip" class="size-3" />
                    {{ item.dokumenPath }}
                  </span>
                </p>
              </div>
            </div>
            
            <div class="flex gap-2 shrink-0">
              <UTooltip :text="t('executiveSummaryIndividual.card.viewDetail')">
                <UButton 
                  color="neutral" 
                  variant="ghost" 
                  icon="i-lucide-eye" 
                  size="md" 
                  :title="t('executiveSummaryIndividual.card.viewDetail')"
                  :aria-label="t('executiveSummaryIndividual.card.viewDetail')"
                  @click="store.openView(item)" 
                />
              </UTooltip>
              <UTooltip :text="t('executiveSummaryIndividual.card.edit')">
                <UButton 
                  v-if="item.status !== 'Approved'" 
                  color="warning" 
                  variant="ghost" 
                  icon="i-lucide-edit" 
                  size="md" 
                  @click="store.openEditForm(item as any)" 
                />
              </UTooltip>
              <UTooltip :text="t('executiveSummaryIndividual.card.delete')">
                <UButton 
                  color="error" 
                  variant="ghost" 
                  icon="i-lucide-trash-2" 
                  size="md" 
                  @click="store.deleteSummary(item.id, item.nomorDokumen)" 
                />
              </UTooltip>

            </div>
          </div>

          <!-- Body Info & Metrics -->
          <div class="p-5 grid grid-cols-1 md:grid-cols-3 gap-6">
            <!-- Narrative Preview -->
            <div class="md:col-span-2 space-y-2">
              <span class="text-md font-bold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryIndividual.card.narrativeSummary') }}</span>
              <p class="text-sm text-gray-600 dark:text-gray-300 line-clamp-3 leading-relaxed italic">
                "{{ item.narrative || t('executiveSummaryIndividual.card.noNarrative') }}"
              </p>
            </div>

            <!-- Metrics Overview -->
            <div class="bg-gray-50 dark:bg-gray-800/60 p-4 rounded-xl space-y-3 border border-gray-100 dark:border-gray-700/50">
              <div class="text-md font-bold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryIndividual.card.metrics') }}</div>
              <div class="grid grid-cols-2 gap-2 text-center">
                <div class="bg-white dark:bg-gray-800 p-2 rounded border">
                  <div class="text-md text-gray-400">{{ t('executiveSummaryIndividual.card.highRisk') }}</div>
                  <div class="text-lg font-bold text-error-600">{{ item.risikoTinggi || 0 }}</div>
                </div>
                <div class="bg-white dark:bg-gray-800 p-2 rounded border">
                  <div class="text-md text-gray-400">{{ t('executiveSummaryIndividual.card.totalRecommendations') }}</div>
                  <div class="text-lg font-bold text-primary-600">{{ item.jumlahRekomendasi || 0 }}</div>
                </div>
              </div>
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
                  <span>{{ t('executiveSummaryForm.notes.title') }}</span>
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
      <div v-else class="text-center py-12">
        <UIcon name="i-lucide-file-x" class="size-16 text-gray-300 mx-auto mb-3" />
        <h3 class="text-lg font-bold text-gray-700 dark:text-gray-300">{{ t('executiveSummaryIndividual.empty.title') }}</h3>
        <p class="text-sm text-gray-500 mt-1 max-w-md mx-auto">
          {{ t('executiveSummaryIndividual.empty.description') }}
        </p>
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummaryIndividual.page.createButton')"
          class="mt-4 font-bold"
          @click="store.openNewForm(1)"
        />
      </div>
    </UCard>

    <!-- Fullscreen Builder Modal -->
    <UModal v-model:open="store.showModal" fullscreen>
      <template #content>
        <div class="flex flex-col h-screen bg-white dark:bg-gray-900">
          <!-- Modal Header Bar -->
          <div class="px-4 sm:px-6 py-3 sm:py-4 border-b border-gray-200 dark:border-gray-800 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3 bg-gray-50 dark:bg-gray-800 shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-primary-100 dark:bg-primary-900 rounded-lg text-primary-600 shrink-0">
                <UIcon name="i-lucide-file-signature" class="size-6" />
              </div>
              <div>
                <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">
                  {{ store.isViewing ? t('executiveSummaryIndividual.modal.view') : store.isEditing ? t('executiveSummaryIndividual.modal.edit') : t('executiveSummaryIndividual.modal.create') }}
                </h2>
                <p class="text-xs sm:text-sm text-gray-500">{{ t('executiveSummaryIndividual.modal.lhaId', { id: store.form.nomorDokumen || t('executiveSummaryIndividual.modal.draft') }) }}</p>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto justify-end">
              <UButton
                v-if="!store.isViewing && store.form.status !== 'Approved'"
                color="primary"
                icon="i-lucide-save"
                :label="t('executiveSummaryForm.actions.saveReport')"
                class="font-bold flex-1 sm:flex-none justify-center" 
                :loading="store.loading"
                @click="store.saveForm"
              />
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-x"
                :label="t('executiveSummaryForm.actions.close')"
                class="flex-1 sm:flex-none justify-center"
                @click="() => { store.showModal = false }"
              />
            </div>
          </div>

          <!-- Form Component Body -->
          <div class="flex-1 overflow-hidden">
            <ExecutiveSummaryIndividualForm />
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useExecutiveSummaryStore, loadPersistedExecutiveSummaryOverrides, type ExecutiveSummary } from '~/stores/executive-summary'
import { useI18n } from '~/composables/useI18n'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import ExecutiveSummaryIndividualForm from '~/components/audit-result-report/ExecutiveSummaryIndividualForm.vue'

const { t } = useI18n()

definePageMeta({
  middleware: 'auth'
})

const store = useExecutiveSummaryStore()
const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()
const searchQuery = ref('')
const selectedLhaNumber = ref('')

const totalHandling = computed(() => {
  return syncedSummaries.value.reduce((total, summary) => {
    return total + (summary.followUpTable || []).reduce((count, row) => count + Number(row.jumlah || 0), 0)
  }, 0)
})

const lhaNumberOptions = computed(() => {
  const isExcluded = (num: string) =>
    num === 'DOC-EXSUM-Q1-2026' ||
    num === '020/LHA/01/KS IAD/2023' ||
    num === '019/LHA/01/KS IAD/2025' ||
    num === 'LHA-020/SKAI/2023' ||
    num === 'LHA-019/SKAI/2025' ||
    num.startsWith('ESC-')

  const numsFromReports = auditReportStore.reportList
    .map((r: any) => r.reportNumber || r.report_number)
    .filter(Boolean)
  const numsFromSummaries = store.summaryList
    .map((s: any) => s.nomorDokumen)
    .filter(Boolean)
  const numsFromSynced = syncedSummaries.value
    .map((s: any) => s.nomorDokumen)
    .filter(Boolean)

  const combined = Array.from(new Set([
    ...numsFromReports, 
    ...numsFromSummaries, 
    ...numsFromSynced
  ]))
    .filter(num => !isExcluded(num))
    .sort()

  return ['Semua Nomor LHA', ...combined]
})

// Load stores on mount
if (!store.loading) {
  store.fetchSummaries()
}
if (!auditReportStore.loading) {
  auditReportStore.fetchReports()
}

// Map month string
const getMonthName = (mStr: string) => {
  const m = parseInt(mStr)
  const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  return months[m - 1] || 'Januari'
}

// Synced list combining Executive Summaries and Result Reports LHA items
const syncedSummaries = computed(() => {
  const overrides = loadPersistedExecutiveSummaryOverrides()
  const rawList: (ExecutiveSummary & { assignmentLetterId?: string })[] = JSON.parse(JSON.stringify(store.summaryList))

  // Filter out any dummy / non-standard legacy documents
  const filteredRaw = rawList.filter(s => 
    s.nomorDokumen !== 'DOC-EXSUM-Q1-2026' &&
    s.nomorDokumen !== '020/LHA/01/KS IAD/2023' &&
    s.nomorDokumen !== '019/LHA/01/KS IAD/2025' &&
    s.nomorDokumen !== 'LHA-020/SKAI/2023' &&
    s.nomorDokumen !== 'LHA-019/SKAI/2025'
  )

  // Map to deduplicate strictly by nomorDokumen
  const docMap = new Map<string, ExecutiveSummary & { assignmentLetterId?: string }>()

  filteredRaw.forEach(s => {
    const override = overrides[s.nomorDokumen] || overrides[s.id]
    if (override) {
      if (override.status) s.status = override.status as any
      if (override.executiveNote !== undefined) s.executiveNote = override.executiveNote
      if (override.assignmentLetterId) s.assignmentLetterId = override.assignmentLetterId
    }
    if (s.nomorDokumen) {
      docMap.set(s.nomorDokumen, s)
    }
  })

  // Synchronize with LHA items from auditReportStore
  auditReportStore.reportList.forEach(lha => {
    const lhaNum = lha.reportNumber || (lha as any).report_number
    if (!lhaNum ||
      lhaNum === 'DOC-EXSUM-Q1-2026' ||
      lhaNum === '020/LHA/01/KS IAD/2023' ||
      lhaNum === '019/LHA/01/KS IAD/2025' ||
      lhaNum === 'LHA-020/SKAI/2023' ||
      lhaNum === 'LHA-019/SKAI/2025') return

    if (docMap.has(lhaNum)) {
      const existing = docMap.get(lhaNum)!
      if (lha.assignmentLetterId && (!existing.assignmentLetterId || !existing.assignmentLetterId.startsWith('ST-'))) {
        existing.assignmentLetterId = lha.assignmentLetterId
      }
      if (lha.executiveSummary && (!existing.narrative || existing.narrative.startsWith('Executive Summary Individual untuk'))) {
        existing.narrative = lha.executiveSummary
      }
      if (lha.findingsCount && existing.jumlahRekomendasi === 0 && lha.findingsCount > 0) {
        existing.jumlahRekomendasi = lha.findingsCount
      }
      const override = overrides[lhaNum] || overrides[existing.id]
      if (override?.executiveNote !== undefined) {
        existing.executiveNote = override.executiveNote
      }
      if (override?.status) {
        existing.status = override.status as any
      }
    } else {
      if (!store.deletedDocNumbers.includes(lhaNum)) {
        const dateParts = lha.reportDate ? lha.reportDate.split('-') : ['2026', '01', '15']
        const yr = parseInt(dateParts[0] || '2026') || 2026
        const mo = getMonthName(dateParts[1] || '01')

        const override = overrides[lhaNum] || overrides[`ES-${lha.id}`]
        let itemStatus: 'Draft' | 'Approved' | 'Rejected' = 
          (lha.status === 'Final' || (lha.status as any) === 'Approved' || (lha.status as any) === 'APPROVED') ? 'Approved' : 'Draft'
        let itemNote = (lha as any).executiveNote || (lha as any).executive_note || ''

        if (override) {
          if (override.status) itemStatus = override.status as any
          if (override.executiveNote !== undefined) itemNote = override.executiveNote
        }

        const fallbackSt = lha.assignmentLetterId || ''

        docMap.set(lhaNum, {
          id: `ES-${lha.id}`,
          assignmentLetterId: fallbackSt,
          quarter: mo === 'Januari' || mo === 'Februari' || mo === 'Maret' ? 1 : mo === 'April' || mo === 'Mei' || mo === 'Juni' ? 2 : 3,
          periodeBulan: `${mo} ${yr}`,
          tahun: yr,
          nomorDokumen: lhaNum,
          dokumenPath: `Executive_Summary_${lhaNum.replace(/[\/\s]/g, '_')}.pdf`,
          status: itemStatus,
          executiveNote: itemNote,
          narrative: lha.executiveSummary || `Executive Summary untuk ${lha.reportTitle} (${lhaNum}).`,
          jumlahLaporan: 1,
          risikoTinggi: (lha.findings || []).filter(f => ['Very Significant', 'Significant'].includes(f.category)).length || 0,
          risikoSedang: (lha.findings || []).filter(f => f.category === 'Quite Significant').length || 0,
          risikoRendah: (lha.findings || []).filter(f => f.category === 'Not Significant').length || 0,
          jumlahRekomendasi: lha.findingsCount || (lha.findings?.length || 0),
          followUpTable: [],
          topFindings: (lha.findings || []).map(f => ({
            unitDivision: lha.reportTitle.includes('Keuangan') ? 'Finance' : 'Operasi',
            judulTemuan: f.title,
            risiko: f.category === 'Very Significant' || f.category === 'Significant' ? 'Tinggi' : f.category === 'Quite Significant' ? 'Sedang' : 'Rendah',
            statusTL: 'In Progress',
            usulan: 'Rekomendasi Perbaikan'
          })),
          matriksKompilasi: [],
          akarMasalah: 'Penguatan sistem pengendalian internal dan efektivitas otomatisasi SOP.',
          kesimpulan: 'Tata kelola dan pengendalian internal berjalan baik dengan rekomendasi perbaikan berkala.',
          signatureTempat: 'Jakarta' as string,
          signatureTanggal: lha.reportDate || '2026-04-15',
          signatureNamaKepala: 'Head of SKAI',
          signatureNIK: 'NIK-100240'
        })
      }
    }
  })

  const result: (ExecutiveSummary & { assignmentLetterId?: string })[] = []
  docMap.forEach(s => {
    if (!store.deletedDocNumbers.includes(s.nomorDokumen)) {
      result.push(s)
    }
  })

  return result
})

const filteredSummaries = computed(() => {
  let list = syncedSummaries.value

  if (selectedLhaNumber.value && selectedLhaNumber.value !== 'Semua Nomor LHA' && selectedLhaNumber.value !== 'All LHA Numbers') {
    list = list.filter(s => s.nomorDokumen === selectedLhaNumber.value)
  }

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(s => 
      s.nomorDokumen.toLowerCase().includes(q) ||
      (s.assignmentLetterId && s.assignmentLetterId.toLowerCase().includes(q)) ||
      (s.narrative && s.narrative.toLowerCase().includes(q))
    )
  }

  return list
})

const getStatusColor = (status: string) => {
  switch (status) {
    case 'Approved': return 'success'
    case 'Draft': return 'warning'
    case 'Rejected': return 'error'
    default: return 'neutral'
  }
}
</script>
