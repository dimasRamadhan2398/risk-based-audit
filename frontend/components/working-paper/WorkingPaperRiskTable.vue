<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        
        <UTable :data="store.filteredDataF02" :columns="store.columnsF02" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'No data saved yet.' }">
            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                    <UButton 
                        size="md" 
                        color="warning" 
                        variant="soft" 
                        icon="i-heroicons-pencil-square" 
                        @click="store.handleEditF02(row.original)" 
                        title="Edit"
                    />
                    <UButton 
                        size="md" 
                        color="error" 
                        variant="soft" 
                        icon="i-heroicons-trash" 
                        @click="store.handleDeleteF02(row.original.id)" 
                        title="Hapus"
                    />
                </div>
            </template>
        </UTable>
    </UCard>
    </div>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>