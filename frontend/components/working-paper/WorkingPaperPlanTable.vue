<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        <UTable :data="store.dataF05" :columns="store.columnsF05" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }" >
            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                    <UButton 
                        size="md" 
                        color="warning" 
                        variant="soft" 
                        icon="i-heroicons-pencil-square" 
                        @click="store.handleEditF05(row.original)" 
                        title="Edit"
                    />
                    <UButton 
                        size="md" 
                        color="error" 
                        variant="soft" 
                        icon="i-heroicons-trash" 
                        @click="store.handleDeleteF05(row.original.id)" 
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