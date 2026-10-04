<template>
  <div class="mt-4 space-y-6">
    <UCard class="shadow-sm mt-10">
      <TableEntities
        :data="store.filteredDataF03"
        :columns="columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack', label: t('workingPaper.sampleTable.emptyState') }"
        :ui="{ td: '!whitespace-normal' }"
      >
        <template #population-cell="{ row }">
          <div class="font-medium text-gray-800 dark:text-gray-200 whitespace-normal break-words">
            {{ row.original.population ?? '-' }}
          </div>
        </template>

        <template #samples-cell="{ row }">
          <div class="flex flex-col gap-2 max-w-md py-1">
            <UCard
              v-for="(s, sIdx) in getSamples(row.original)"
              :key="s.id || sIdx"
              :ui="{
                root: 'border border-gray-200 dark:border-gray-700 rounded-lg shadow-none',
                body: 'p-2.5 sm:p-2.5',
              }"
            >
              <div class="flex items-start justify-between gap-2 mb-1.5">
                <div class="min-w-0">
                  <div v-if="s.fieldworkDocument" class="text-xs font-semibold text-primary-600 dark:text-primary-400 mb-0.5 truncate">
                    {{ s.fieldworkDocument }}
                  </div>
                  <span class="text-sm font-semibold text-gray-800 dark:text-gray-100 leading-snug">
                    {{ s.document || t('workingPaper.sampleTable.noDocumentName') }}
                  </span>
                </div>
                <UBadge
                  :color="store.checkSampleStatus(s) ? 'success' : 'error'"
                  size="sm"
                  variant="subtle"
                  class="shrink-0"
                >
                  {{ store.checkSampleStatus(s) ? t('workingPaper.sampleTable.effective') : t('workingPaper.sampleTable.ineffective') }}
                </UBadge>
              </div>
              <USeparator class="my-1.5" />
              <div v-if="s.step1 || s.step2 || s.step3" class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
                <div v-if="s.step1" class="flex justify-between gap-2">
                  <span class="truncate">L1: {{ s.step1 }}</span>
                  <span class="font-semibold shrink-0" :class="formatResult(s.l1) === 'Fail' ? 'text-red-500' : (formatResult(s.l1) === 'Pass' ? 'text-green-600 dark:text-green-400' : 'text-gray-400')">{{ formatResult(s.l1) }}</span>
                </div>
                <div v-if="s.step2" class="flex justify-between gap-2">
                  <span class="truncate">L2: {{ s.step2 }}</span>
                  <span class="font-semibold shrink-0" :class="formatResult(s.l2) === 'Fail' ? 'text-red-500' : (formatResult(s.l2) === 'Pass' ? 'text-green-600 dark:text-green-400' : 'text-gray-400')">{{ formatResult(s.l2) }}</span>
                </div>
                <div v-if="s.step3" class="flex justify-between gap-2">
                  <span class="truncate">L3: {{ s.step3 }}</span>
                  <span class="font-semibold shrink-0" :class="formatResult(s.l3) === 'Fail' ? 'text-red-500' : (formatResult(s.l3) === 'Pass' ? 'text-green-600 dark:text-green-400' : 'text-gray-400')">{{ formatResult(s.l3) }}</span>
                </div>
              </div>
              <div v-else class="text-xs text-gray-500 dark:text-gray-400 italic">
                L1: {{ formatResult(s.l1) }} &nbsp;|&nbsp; L2: {{ formatResult(s.l2) }} &nbsp;|&nbsp; L3: {{ formatResult(s.l3) }}
              </div>
            </UCard>

            <p v-if="!getSamples(row.original).length" class="text-sm text-gray-400 dark:text-gray-500 italic">
              {{ t('workingPaper.sampleTable.noSampleData') }}
            </p>
          </div>
        </template>

        <template #conclusion-cell="{ row }">
          <div
            class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
            style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
          >
            {{ row.original.conclusion || '-' }}
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2">
            <UTooltip :text="t('workingPaper.sampleTable.editSample')">
              <UButton 
                size="md" 
                color="warning" 
                variant="ghost" 
                icon="i-lucide-edit" 
                @click="store.handleEditF03(row.original)" 
              />
            </UTooltip>
            <UTooltip :text="t('workingPaper.sampleTable.deleteSample')">
              <UButton 
                size="md" 
                color="error" 
                variant="ghost" 
                icon="i-lucide-trash-2" 
                @click="store.handleDeleteF03(row.original.id)" 
              />
            </UTooltip>
          </div>
        </template>
      </TableEntities>
    </UCard>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useWorkingPaperStore } from '~/stores/working-paper'

const { t } = useI18n()
const store = useWorkingPaperStore()

const columns = computed(() => [
  { key: 'population', accessorKey: 'population', header: t('workingPaper.sampleTable.columns.population'), class: 'w-48 min-w-[140px] whitespace-normal break-words', tdClass: '!whitespace-normal break-words' },
  { key: 'sampleSize', accessorKey: 'sampleSize', header: t('workingPaper.sampleTable.columns.sampleSize'), class: 'w-36 min-w-[120px]' },
  { key: 'samples', accessorKey: 'samples', header: t('workingPaper.sampleTable.columns.samples'), class: 'min-w-[260px]' },
  {
    key: 'conclusion',
    accessorKey: 'conclusion',
    header: t('workingPaper.sampleTable.columns.conclusion'),
    class: 'w-80 min-w-[240px] max-w-sm !whitespace-normal break-words',
    tdClass: '!whitespace-normal break-words'
  },
  { key: 'actions', accessorKey: 'actions', header: t('workingPaper.sampleTable.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
])

const formatResult = (val: any) => {
  if (val === true || val === 'Pass' || (typeof val === 'string' && val.toLowerCase() === 'pass')) return 'Pass'
  if (val === false || val === 'Fail' || (typeof val === 'string' && val.toLowerCase() === 'fail')) return 'Fail'
  return '-'
}

const getSamples = (rowOriginal: any): any[] => {
  if (!rowOriginal?.samples) return []
  if (Array.isArray(rowOriginal.samples)) return rowOriginal.samples
  if (typeof rowOriginal.samples === 'string') {
    try {
      const parsed = JSON.parse(rowOriginal.samples)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}
</script>