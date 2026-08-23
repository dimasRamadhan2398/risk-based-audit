<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        <TableEntities :data="store.filteredDataF04" :columns="store.columnsF04" :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }">
            <template #evidenceFile-cell="{ row }">
                <div v-if="row.original.evidenceFile" class="flex items-center gap-1 text-blue-600">
                    <UIcon name="i-heroicons-paper-clip" />
                    <span class="text-md truncate max-w-[150px]">{{ row.original.evidenceFile.name }}</span>
                </div>
                <span v-else class="text-gray-400">-</span>
            </template>

            <template #rootCause-cell="{ row }">
                <div class="space-y-2 py-2">
                    <div v-for="rca in row.original.rootCause" :key="rca.id" class="text-[11px] leading-tight border-l-2 border-orange-400 pl-2">
                        <div class="font-bold text-lg text-gray-700">{{ rca.method }}</div>
                    <div class="text-gray-500 text-sm italic">
                        Why 1: {{ rca.w1 || '-' }} <br>
                        Why 2: {{ rca.w2 || '-' }} <br>
                        Why 3: {{ rca.w3 || '-' }}
                    </div>
                    </div>
                </div>
            </template>

            <template #actions-cell="{ row }">
                <div class="flex gap-2">
                <UButton 
                    size="md" 
                    color="warning" 
                    variant="soft" 
                    icon="i-heroicons-pencil-square" 
                    @click="store.handleEditF04(row.original)" 
                    title="Edit"
                />
                <UButton 
                    size="md" 
                    color="error" 
                    variant="soft" 
                    icon="i-heroicons-trash" 
                    @click="store.handleDeleteF04(row.original.id)" 
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

const store = useWorkingPaperStore()

</script>