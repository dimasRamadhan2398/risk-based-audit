<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        <TableEntities
          :data="store.filteredDataF05"
          :columns="store.columnsF05"
          :empty-state="store.hasAssignmentLetter ? { icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' } : { icon: 'i-heroicons-document-magnifying-glass', label: t('workingPaper.index.selectLetterFirst') }"
          :ui="{ td: '!whitespace-normal' }"
        >
            <template #recommendation-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.recommendation || '-' }}
              </div>
            </template>

            <template #response-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.response || '-' }}
              </div>
            </template>

            <template #pic-cell="{ row }">
              <span class="text-sm text-gray-600 dark:text-gray-300 whitespace-nowrap">
                {{ store.formatEmployeeName(row.original.pic) }}
              </span>
            </template>

            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                  <UTooltip v-if="canWrite" :text="t('workingPaper.common.edit')">
                    <UButton 
                        size="md" 
                        color="warning" 
                        variant="ghost" 
                        icon="i-lucide-edit" 
                        @click="store.handleEditF05(row.original)" 
                    />
                  </UTooltip>
                  <UTooltip v-if="canWrite" :text="t('workingPaper.common.delete')">
                    <UButton 
                        size="md" 
                        color="error" 
                        variant="ghost" 
                        icon="i-lucide-trash-2" 
                        @click="store.handleDeleteF05(row.original.id)" 
                    />
                  </UTooltip>
                </div>
            </template>
        </TableEntities>
        </UCard>
    </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRbac } from '~/composables/useRbac'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useI18n } from '~/composables/useI18n'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const { t } = useI18n()
const { canWrite } = useRbac()
const store = useWorkingPaperStore()

onMounted(() => {
  store.fetchEmployees?.()
})
</script>