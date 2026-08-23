<template>
    <div class="mt-4 space-y-6">
    <UCard class="shadow-sm mt-10">
        
        <TableEntities :data="store.filteredDataF01" :columns="store.columnsF01" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'No data saved yet.' }" >
        
        <template #teamMembers-cell="{ row }">
            <div class="flex flex-wrap gap-1">
            <UBadge 
                v-for="member in row.original.teamMembers"
                :key="member.id"
                color="neutral" 
                variant="subtle" 
                size="lg"
                class="flex flex-col items-start px-2 py-1"
            >
                <span class="font-bold text-primary-700">{{ member.name }}</span>
                <span class="text-[10px] opacity-70 italic">{{ member.role }}</span>
            </UBadge>
            <span v-if="!row.original.teamMembers?.length" class="text-gray-400">-</span>
            </div>
        </template>

        <template #actions-cell="{ row }">
            <div class="flex gap-2">
                <UButton 
                    size="md" 
                    color="warning" 
                    variant="soft" 
                    icon="i-heroicons-pencil-square" 
                    @click="store.handleEditF01(row.original)" 
                    
                />
                <UButton 
                    size="md" 
                    color="error" 
                    variant="soft" 
                    icon="i-heroicons-trash" 
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