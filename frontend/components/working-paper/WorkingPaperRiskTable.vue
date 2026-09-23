<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        
        <TableEntities
          :data="store.filteredDataF02"
          :columns="store.columnsF02"
          :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'No data saved yet.' }"
          :ui="{ td: '!whitespace-normal' }"
        >
            <template #risk-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.risk || '-' }}
              </div>
            </template>

            <template #controlDescription-cell="{ row }">
              <div
                class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
                style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
              >
                {{ row.original.controlDescription || '-' }}
              </div>
            </template>

            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                  <UTooltip text="Edit Risk">
                    <UButton 
                        size="md" 
                        color="warning" 
                        variant="ghost" 
                        icon="i-lucide-edit" 
                        @click="store.handleEditF02(row.original)" 
                    />
                  </UTooltip>
                  <UTooltip text="Delete Risk">
                    <UButton 
                        size="md" 
                        color="error" 
                        variant="ghost" 
                        icon="i-lucide-trash-2" 
                        @click="store.handleDeleteF02(row.original.id)" 
                    />
                  </UTooltip>
                </div>
            </template>
        </TableEntities>
    </UCard>
    </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>