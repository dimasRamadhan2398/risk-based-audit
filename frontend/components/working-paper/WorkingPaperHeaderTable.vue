<template>
    <div class="mt-4 space-y-6">
    <UCard class="shadow-sm mt-10">
        
        <TableEntities :data="store.filteredDataF01" :columns="store.columnsF01" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'No data saved yet.' }" >
        
        <template #teamMembers-cell="{ row }">
          <TeamMembersBadge :members="row.original.teamMembers" />
        </template>

        <template #actions-cell="{ row }">
            <div class="flex gap-2">
                <UButton 
                    size="md" 
                    color="warning" 
                    variant="ghost" 
                    icon="i-lucide-edit" 
                    @click="store.handleEditF01(row.original)" 
                    
                />
                <UButton 
                    size="md" 
                    color="error" 
                    variant="ghost" 
                    icon="i-lucide-trash-2" 
                    @click="store.handleDeleteF01(row.original.id)" 
                    
                />
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