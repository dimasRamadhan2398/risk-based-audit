<script setup lang="ts">
import { useAuditExecutionStore } from '~/stores/audit-execution'
import AuditExecutionDetailModal from '~/components/audit-execution/AuditExecutionDetailModal.vue'
import { AuditCategory, AuditStatus, EXECUTION_PHASES, getExecutionPhase } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useAuditExecutionStore()
store.fetchAuditExecutions()
const { auditExecutions, getSummary } = storeToRefs(store)

const search = ref('')
const quarter = ref('')
const status = ref<AuditStatus | undefined>(undefined)
const category = ref<AuditCategory | undefined>(undefined)

const isHelpModalOpen = ref(false)

const quarters = ['Quarter I', 'Quarter II', 'Quarter III', 'Quarter IV']

const resetFilters = () => {
  search.value = ''
  quarter.value = ''
  category.value = undefined
  status.value = undefined
}

const columns = computed(() => [
  { accessorKey: 'name', header: t('auditExecution.columns.name') },
  { accessorKey: 'phase', header: t('auditExecution.columns.phase') },
  { accessorKey: 'progress', header: t('auditExecution.columns.progress') },
  { accessorKey: 'lead_auditor', header: t('auditExecution.columns.leadAuditor') },
  { accessorKey: 'actions', header: t('auditExecution.columns.actions') }
])

const filteredAudits = computed(() => {
  return auditExecutions.value.filter(audit => {
    const matchesSearch = !search.value || 
                          (audit.name && audit.name.toLowerCase().includes(search.value.toLowerCase())) || 
                          (audit.category && audit.category.toLowerCase().includes(search.value.toLowerCase())) ||
                          (audit.ref && audit.ref.toLowerCase().includes(search.value.toLowerCase()))
    const matchesCategory = !category.value || audit.category === category.value
    const matchesStatus = !status.value || audit.status === status.value
    return matchesSearch && matchesCategory && matchesStatus
  })
})

const phaseBreakdown = computed(() => {
  const counts: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0 }
  auditExecutions.value.forEach(audit => {
    const phase = getExecutionPhase(audit.progress)
    counts[phase.step] = (counts[phase.step] || 0) + 1
  })
  return counts
})

const getProgressValue = (row: any): number => {
  const item = row?.original || row
  const val = item?.progress
  if (typeof val === 'number') return val
  if (typeof val === 'string') return parseFloat(val) || 0
  return 0
}

const getProgressColor = (val: number) => {
  if (val >= 100) return 'bg-secondary-600 shadow-sm shadow-secondary-600/40'
  if (val >= 76) return 'bg-indigo-500 shadow-sm shadow-indigo-500/40'
  if (val >= 51) return 'bg-purple-500 shadow-sm shadow-purple-500/40'
  if (val >= 26) return 'bg-violet-500 shadow-sm shadow-violet-500/40'
  if (val >= 1) return 'bg-blue-500 shadow-sm shadow-blue-500/40'
  return 'bg-sky-400 shadow-sm shadow-sky-400/40'
}

const isDetailOpen = ref(false)
const selectedAudit = ref<any>(undefined)

const openDetail = (row: any) => {
  selectedAudit.value = row?.original || row
  isDetailOpen.value = true
}

const handleRemind = (audit: any) => {
  useToast().add({
    title: t('auditExecution.detailModal.reminderSent'),
    description: t('auditExecution.detailModal.reminderDescription', {
      name: audit?.name || audit?.nama_audit || '',
      auditor: audit?.lead_auditor || ''
    }),
    color: 'success'
  })
}
</script>

<template>
  <div class="p-6 space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('auditExecution.title') }}</h1>
        <p class="text-md text-gray-500 dark:text-gray-400 mt-1">
          {{ t('auditExecution.subtitle') }}
        </p>
      </div>

      <UButton
        icon="i-lucide-help-circle"
        :label="t('auditExecution.cycleGuideButton')"
        color="primary"
        variant="soft"
        size="md"
        class="font-bold"
        @click="() => { isHelpModalOpen = true }"
      />
    </div>

    <!-- Summary Section -->
    <div class="bg-white dark:bg-gray-900 rounded-xl p-4 border border-gray-200 dark:border-gray-800 shadow-sm flex flex-wrap items-center justify-between gap-4">
      <div class="flex flex-wrap items-center gap-6">
        <span class="text-md font-bold text-gray-500 uppercase tracking-wider">{{ t('auditExecution.summary.title') }}</span>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-emerald-500"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.completed', { count: getSummary.completed }) }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-sky-500"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.inProgress', { count: getSummary.inProgress }) }}</span>
        </div>
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-gray-300 dark:bg-gray-600"></span>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('auditExecution.summary.planned', { count: getSummary.planned }) }}</span>
        </div>
      </div>

      <UButton
        :label="t('auditExecution.summary.viewPhaseGuide')"
        icon="i-lucide-book-open"
        color="neutral"
        variant="ghost"
        size="md"
        @click="() => { isHelpModalOpen = true }"
      />
    </div>

    <!-- Filters Section -->
    <div class="flex flex-wrap items-center gap-4">
      <UInput
        v-model="search"
        icon="i-lucide-search"
        :placeholder="t('auditExecution.filters.searchPlaceholder')"
        class="w-48"
      />
      <USelectMenu
        v-model="quarter"
        :items="quarters"
        :placeholder="t('auditExecution.filters.quarterPlaceholder')"
        class="w-48"
      />
      <USelectMenu
        v-model="category"
        :items="Object.values(AuditCategory)"
        :placeholder="t('auditExecution.filters.categoryPlaceholder')"
        class="w-48"
      />
      <USelectMenu
        v-model="status"
        :items="Object.values(AuditStatus)"
        :placeholder="t('auditExecution.filters.statusPlaceholder')"
        class="w-48"
      />
      <UButton
        label="Reset Filter"
        icon="i-lucide-rotate-ccw"
        color="neutral"
        variant="outline"
        @click="resetFilters"
      />
    </div>

    <!-- Table Section -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <TableEntities :columns="columns" :data="filteredAudits">
        <template #name-cell="{ row }">
          <div class="flex flex-col">
            <span class="font-bold text-gray-900 dark:text-white">{{ (row.original || row).name }}</span>
            <span class="text-md text-gray-500 font-medium">{{ t('auditExecution.table.ref') }} {{ (row.original || row).ref || '-' }} | ({{ (row.original || row).category }})</span>
          </div>
        </template>

        <template #phase-cell="{ row }">
          <UBadge :color="getExecutionPhase(getProgressValue(row)).badgeColor" variant="subtle" size="md">
            <UIcon :name="getExecutionPhase(getProgressValue(row)).icon" class="mr-1.5 inline-block text-md" />
            {{ t(`auditExecution.phases.${getExecutionPhase(getProgressValue(row)).step}.shortLabel`) }}
          </UBadge>
        </template>

        <template #progress-cell="{ row }">
          <div class="flex items-center gap-3 min-w-[200px]">
            <div class="flex-1 h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden border border-gray-200 dark:border-gray-700 shadow-inner">
              <div
                class="h-full rounded-full transition-all duration-500 ease-out"
                :class="getProgressColor(getProgressValue(row))"
                :style="{ width: `${Math.min(100, Math.max(0, getProgressValue(row)))}%` }"
              ></div>
            </div>
            <span class="text-md font-bold text-gray-700 dark:text-white w-10 text-right">
              {{ getProgressValue(row) }}%
            </span>
          </div>
        </template>

        <template #lead_auditor-cell="{ row }">
          <span class="text-sm text-gray-700 dark:text-white">{{ (row.original || row).lead_auditor || '-' }}</span>
        </template>

        <template #actions-cell="{ row }">
          <UButton
            icon="i-lucide-eye"
            variant="ghost"
            color="neutral"
            @click="openDetail(row)"
          />
        </template>
      </TableEntities>
    </div>

    <!-- Detail Modal -->
    <AuditExecutionDetailModal
      v-model:open="isDetailOpen"
      :audit="selectedAudit"
      @remind="handleRemind"
    />

    <!-- Lifecycle Phase Guide Help Modal -->
    <UModal v-model:open="isHelpModalOpen" :ui="{ content: 'sm:max-w-3xl' }" scrollable>
      <template #content>
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2.5">
                <div class="p-2 rounded-lg bg-secondary-100 dark:bg-secondary-950/60 text-secondary-600 dark:text-secondary-400">
                  <UIcon name="i-lucide-route" class="text-xl" />
                </div>
                <div>
                  <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('auditExecution.guideModal.title') }}</h3>
                  <p class="text-md text-gray-500">{{ t('auditExecution.guideModal.subtitle') }}</p>
                </div>
              </div>
              <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="() => { isHelpModalOpen = false }" />
            </div>
          </template>

          <div class="space-y-4">
            <p class="text-md text-gray-600 dark:text-gray-300 leading-relaxed">
              {{ t('auditExecution.guideModal.description') }}
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div
                v-for="phase in EXECUTION_PHASES"
                :key="phase.step"
                class="p-4 rounded-xl border space-y-2 transition-all duration-200 hover:shadow-md"
                :class="phase.cardClass"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div
                      class="w-6 h-6 rounded-full flex items-center justify-center text-md font-bold shrink-0 shadow-md"
                      :class="phase.numBgClass"
                    >
                      {{ phase.step }}
                    </div>
                    <span class="text-sm font-bold text-gray-900 dark:text-white">{{ t(`auditExecution.phases.${phase.step}.title`) }}</span>
                  </div>
                  <span
                    class="text-md font-bold px-2 py-0.5 rounded-full shrink-0"
                    :class="phase.badgeClass"
                  >
                    {{ phase.minProgress === phase.maxProgress ? `${phase.minProgress}%` : `${phase.minProgress}–${phase.maxProgress}%` }}
                  </span>
                </div>
                <p class="text-[11px] text-gray-600 dark:text-gray-400 pl-8 leading-relaxed">
                  {{ t(`auditExecution.phases.${phase.step}.description`) }}
                </p>
              </div>
            </div>
          </div>

          <template #footer>
            <div class="flex justify-end">
              <UButton
                :label="t('auditExecution.guideModal.close')"
                color="primary"
                class="px-5 font-bold"
                @click="() => { isHelpModalOpen = false }"
              />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </div>
</template>

