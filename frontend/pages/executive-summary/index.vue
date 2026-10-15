<template>
  <div class="space-y-6 min-w-0">
    <!-- Header -->
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon name="i-lucide-file-text" class="size-6 sm:size-7 text-primary-500 shrink-0" />
          {{ t('executiveSummary.individual.title') }}
        </h1>
        <p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('executiveSummary.individual.subtitle') }}
        </p>
      </div>
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full md:w-auto">
        <UButton
          color="info"
          variant="ghost"
          icon="i-lucide-upload"
          :label="t('executiveSummary.individual.import')"
          to="/executive-summary/upload"
          class="font-bold shadow w-full sm:w-auto justify-center"
        />
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummary.individual.create')"
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.individual.stats.total') }}</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.individual.stats.approved') }}</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.individual.stats.draft') }}</div>
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
            <div class="text-md text-gray-500 dark:text-gray-400">{{ t('executiveSummary.individual.stats.handling') }}</div>
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
              v-model="selectedAssignmentLetter"
              :items="assignmentLetterOptions"
              value-key="value"
              :placeholder="t('executiveSummary.individual.filters.assignmentLetter')"
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
              :placeholder="t('executiveSummary.individual.filters.search')"
              class="w-full"
            />
          </div>
        </div>
        <div class="text-sm text-gray-500 dark:text-gray-400 shrink-0">
          {{ t('executiveSummary.individual.filters.showing', { count: filteredSummaries.length }) }}
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
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ item.nomorDokumen || t('executiveSummary.individual.card.draftLha') }}</h3>
                  <UBadge :color="getStatusColor(item.status)" variant="soft" class="font-semibold uppercase tracking-wider text-[10px]">
                    {{ t('executiveSummary.individual.card.status', { status: getStatusLabel(item.status) }) }}
                  </UBadge>
                  <UBadge v-if="item.assignmentLetterId" color="info" variant="subtle" class="font-mono text-[11px]">
                    {{ t('executiveSummary.individual.card.assignmentLetter', { number: item.assignmentLetterId }) }}
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
              <UTooltip :text="t('executiveSummary.individual.actions.view')">
                <UButton 
                  color="neutral" 
                  variant="ghost" 
                  icon="i-lucide-eye" 
                  size="md" 
                  :title="t('executiveSummary.individual.actions.view')"
                  :aria-label="t('executiveSummary.individual.actions.view')"
                  @click="store.openView(item)" 
                />
              </UTooltip>
              <UTooltip :text="t('executiveSummary.individual.actions.edit')">
                <UButton 
                  v-if="item.status !== 'Approved'" 
                  color="warning" 
                  variant="ghost" 
                  icon="i-lucide-edit" 
                  size="md" 
                  @click="store.openEditForm(item as any)" 
                />
              </UTooltip>
              <UTooltip :text="t('executiveSummary.individual.actions.delete')">
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
              <span class="text-md font-bold uppercase tracking-wider text-gray-400">{{ t('executiveSummary.individual.card.mainSummary') }}</span>
              <p class="text-sm text-gray-600 dark:text-gray-300 line-clamp-3 leading-relaxed italic">
                "{{ item.narrative || t('executiveSummary.individual.card.noNarrative') }}"
              </p>
            </div>

            <!-- Metrics Overview -->
            <div class="bg-gray-50 dark:bg-gray-800/60 p-4 rounded-xl space-y-3 border border-gray-100 dark:border-gray-700/50">
              <div class="text-md font-bold uppercase tracking-wider text-gray-400">{{ t('executiveSummary.individual.card.statistics') }}</div>
              <div class="grid grid-cols-2 gap-2 text-center">
                <div class="bg-white dark:bg-gray-800 p-2 rounded border">
                  <div class="text-md text-gray-400">{{ t('executiveSummary.individual.card.highRisk') }}</div>
                  <div class="text-lg font-bold text-error-600">{{ item.risikoTinggi || 0 }}</div>
                </div>
                <div class="bg-white dark:bg-gray-800 p-2 rounded border">
                  <div class="text-md text-gray-400">{{ t('executiveSummary.individual.card.totalRecommendations') }}</div>
                  <div class="text-lg font-bold text-primary-600">{{ item.jumlahRekomendasi || 0 }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12">
        <UIcon name="i-lucide-file-x" class="size-16 text-gray-300 mx-auto mb-3" />
        <h3 class="text-lg font-bold text-gray-700 dark:text-gray-300">{{ t('executiveSummary.individual.empty.title') }}</h3>
        <p class="text-sm text-gray-500 mt-1 max-w-md mx-auto">
          {{ t('executiveSummary.individual.empty.description') }}
        </p>
        <UButton
          color="primary"
          icon="i-lucide-plus"
          :label="t('executiveSummary.individual.create')"
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
                  {{ store.isViewing ? t('executiveSummary.individual.modal.detail') : store.isEditing ? t('executiveSummary.individual.modal.edit') : t('executiveSummary.individual.modal.create') }}
                </h2>
                <p class="text-xs sm:text-sm text-gray-500">{{ t('executiveSummary.individual.modal.lhaId', { id: store.form.nomorDokumen || t('executiveSummary.individual.modal.draft') }) }}</p>
              </div>
              <UBadge :color="getStatusColor(store.form.status)" variant="soft" class="font-bold uppercase">
                {{ t('executiveSummary.individual.modal.status', { status: getStatusLabel(store.form.status) }) }}
              </UBadge>
            </div>

            <div class="flex flex-wrap items-center gap-2 sm:gap-3 w-full sm:w-auto justify-end">
              <UButton
                v-if="!store.isViewing && store.form.status !== 'Approved'"
                color="primary"
                icon="i-lucide-save"
                :label="t('executiveSummary.individual.modal.save')"
                class="font-bold flex-1 sm:flex-none justify-center" 
                :loading="store.loading"
                @click="store.saveForm"
              />
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-lucide-x"
                :label="t('executiveSummary.individual.modal.close')"
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
import { useExecutiveSummaryStore, type ExecutiveSummary } from '~/stores/executive-summary'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import ExecutiveSummaryIndividualForm from '~/components/audit-result-report/ExecutiveSummaryIndividualForm.vue'

definePageMeta({
  middleware: 'auth'
})

const store = useExecutiveSummaryStore()
const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()
const { t, locale } = useI18n()
const searchQuery = ref('')
const selectedAssignmentLetter = ref('')
const ALL_ASSIGNMENT_LETTERS = '__all__'

const totalHandling = computed(() => {
  return store.summaryList.reduce((total, summary) => {
    return total + (summary.followUpTable || []).reduce((count, row) => count + Number(row.jumlah || 0), 0)
  }, 0)
})

const assignmentLetterOptions = computed(() => {
  const lettersFromStore = assignmentLetterStore.assignmentLetterList.map((st: any) => st.letterNumber)
  const lettersFromReports = auditReportStore.reportList.map((r: any) => r.assignmentLetterId).filter(Boolean)
  const combined = Array.from(new Set([...lettersFromStore, ...lettersFromReports, 'ST-001/SKAI/2026', 'ST-002/SKAI/2026', 'ST-003/SKAI/2026', '020/ST/01/KSIAD/2023']))
  return [
    { label: t('executiveSummary.individual.filters.allAssignmentLetters'), value: ALL_ASSIGNMENT_LETTERS },
    ...combined.map(value => ({ label: value, value }))
  ]
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
  if (!Number.isInteger(m) || m < 1 || m > 12) return new Date(2000, 3, 1).toLocaleString(locale.value === 'id' ? 'id-ID' : 'en-US', { month: 'long' })
  return new Date(2000, m - 1, 1).toLocaleString(locale.value === 'id' ? 'id-ID' : 'en-US', { month: 'long' })
}

// Synced list combining Executive Summaries and Result Reports LHA items
const syncedSummaries = computed(() => {
  const summaries: (ExecutiveSummary & { assignmentLetterId?: string })[] = JSON.parse(JSON.stringify(store.summaryList))

  auditReportStore.reportList.forEach(lha => {
    const lhaNum = lha.reportNumber || (lha as any).report_number
    if (!lhaNum) return

    const existingIndex = summaries.findIndex(s => s.nomorDokumen === (lhaNum || ''))
    if (existingIndex >= 0) {
      const existing = summaries[existingIndex]
      if (!existing) return
      existing.assignmentLetterId = lha.assignmentLetterId
      // Keep narrative and findingsCount in sync with LHA
      if (lha.executiveSummary && (!existing.narrative || existing.narrative.startsWith('Executive Summary Individual untuk'))) {
        existing.narrative = lha.executiveSummary
      }
      if (lha.findingsCount && existing.jumlahRekomendasi === 0 && lha.findingsCount > 0) {
        existing.jumlahRekomendasi = lha.findingsCount
      }
    } else {
      if (!store.deletedDocNumbers.includes(lhaNum)) {
        // Automatically add synced Executive Summary card for this LHA ID
        const dateParts = lha.reportDate ? lha.reportDate.split('-') : ['2026', '04', '15']
        const yr = parseInt(dateParts[0] || '2026') || 2026
        const monthNumber = parseInt(dateParts[1] || '04') || 4
        const mo = getMonthName(String(monthNumber))

        summaries.push({
          id: `ES-${lha.id}`,
          assignmentLetterId: lha.assignmentLetterId,
          quarter: Math.ceil(monthNumber / 3),
          periodeBulan: `${mo} ${yr}`,
          tahun: yr,
          nomorDokumen: lhaNum,
          dokumenPath: `Executive_Summary_${lhaNum.replace(/[\/\s]/g, '_')}.pdf`,
          status: lha.status === 'Final' ? 'Approved' : 'Draft',
          narrative: lha.executiveSummary || `Executive Summary untuk ${lha.reportTitle} (${lhaNum}).`,
          jumlahLaporan: 1,
          risikoTinggi: (lha.findings || []).filter(f => ['Very Significant', 'Significant'].includes(f.category)).length || 1,
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

  // Map missing assignmentLetterId for initial items
  summaries.forEach(s => {
    if (!s.assignmentLetterId) {
      const matchLha = auditReportStore.reportList.find(l => l.reportNumber === s.nomorDokumen)
      if (matchLha) {
        s.assignmentLetterId = matchLha.assignmentLetterId
      } else if (s.nomorDokumen.includes('021')) {
        s.assignmentLetterId = 'ST-001/SKAI/2026'
      } else if (s.nomorDokumen.includes('020')) {
        s.assignmentLetterId = '020/ST/01/KSIAD/2023'
      } else if (s.nomorDokumen.includes('022')) {
        s.assignmentLetterId = 'ST-002/SKAI/2026'
      } else if (s.nomorDokumen.includes('023')) {
        s.assignmentLetterId = 'ST-003/SKAI/2026'
      }
    }
  })

  return summaries.filter(s => !store.deletedDocNumbers.includes(s.nomorDokumen))
})

const filteredSummaries = computed(() => {
  let list = syncedSummaries.value

  if (selectedAssignmentLetter.value && selectedAssignmentLetter.value !== ALL_ASSIGNMENT_LETTERS) {
    list = list.filter(s => s.assignmentLetterId === selectedAssignmentLetter.value)
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

const getStatusLabel = (status: string) => {
  const key = status.toLowerCase()
  return ['approved', 'draft', 'rejected'].includes(key)
    ? t(`executiveSummary.individual.statuses.${key}`)
    : status
}
</script>
