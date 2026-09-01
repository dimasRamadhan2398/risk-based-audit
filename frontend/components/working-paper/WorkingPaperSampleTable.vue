<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        
        <TableEntities :data="store.filteredDataF03" :columns="store.columnsF03" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
            <template #samples-cell="{ row }">
            <div class="flex flex-col gap-2 max-w-md">
                <div 
                v-for="s in row.original.samples" 
                :key="s.id" 
                class="text-sm p-2 border rounded bg-gray-50  border-gray-200 "
                >
                <div class="flex justify-between font-bold mb-1">
                    <span>{{ s.document || 'Tanpa Nama Dokumen' }}</span>
                    <UBadge 
                    :color="store.checkSampleStatus(s) ? 'success' : 'error'" 
                    size="md" 
                    variant="subtle"
                    >
                    {{ store.checkSampleStatus(s) ? 'Efektif' : 'Tidak Efektif' }}
                    </UBadge>
                </div>
                <div class="text-sm text-gray-500 italic">
                    L1: {{ s.l1 || '-' }} | L2: {{ s.l2 || '-' }} | L3: {{ s.l3 || '-' }}
                </div>
                </div>
                
                <span v-if="!row.original.samples?.length" class="text-gray-400 italic text-md">
                Tidak ada data sampel
                </span>
            </div>
            </template>

            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                    <UButton 
                        size="md" 
                        color="warning" 
                        variant="ghost" 
                        icon="i-lucide-edit" 
                        @click="store.handleEditF03(row.original)" 
                        title="Edit"
                    />
                    <UButton 
                        size="md" 
                        color="error" 
                        variant="ghost" 
                        icon="i-lucide-trash-2" 
                        @click="store.handleDeleteF03(row.original.id)" 
                        title="Hapus"
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