<template>
    <div class="mt-4 space-y-6">
    <UCard class="shadow-sm mt-10">
        
        <TableEntities
          :data="store.filteredDataF01"
          :columns="store.columnsF01"
          :empty-state="{ icon: 'i-heroicons-circle-stack', label: t('workingPaper.common.empty') }"
          :ui="{ td: '!whitespace-normal' }"
        >
        
        <template #businessProcess-cell="{ row }">
          <div
            class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
            style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
          >
            {{ row.original.businessProcess || '-' }}
          </div>
        </template>

        <template #teamMembers-cell="{ row }">
          <TeamMembersBadge :members="row.original.teamMembers" />
        </template>

        <template #actions-cell="{ row }">
            <div class="flex gap-2">
              <UTooltip :text="t('workingPaper.common.edit')">
                <UButton 
                    size="md" 
                    color="warning" 
                    variant="ghost" 
                    icon="i-lucide-edit" 
                    @click="store.handleEditF01(row.original)" 
                />
              </UTooltip>
              <UTooltip :text="t('workingPaper.common.delete')">
                <UButton 
                    size="md" 
                    color="error" 
                    variant="ghost" 
                    icon="i-lucide-trash-2" 
                    @click="store.handleDeleteF01(row.original.id)" 
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
const { t } = useI18n()
</script>
