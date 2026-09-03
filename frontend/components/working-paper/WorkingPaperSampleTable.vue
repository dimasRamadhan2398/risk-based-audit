<template>
    <div class="mt-4 space-y-6">
        <UCard class="shadow-sm mt-10">
        
        <TableEntities
          :data="store.filteredDataF03"
          :columns="store.columnsF03"
          :empty-state="{ icon: 'i-heroicons-circle-stack', label: 'Belum ada data tersimpan.' }"
          :ui="{ td: '!whitespace-normal' }"
        >
            <template #samples-cell="{ row }">
              <div class="flex flex-col gap-2 max-w-md py-1">
                <UCard
                  v-for="s in row.original.samples"
                  :key="s.id"
                  :ui="{
                    root: 'border border-gray-200 dark:border-gray-700 rounded-lg shadow-none',
                    body: 'p-2.5 sm:p-2.5',
                  }"
                >
                  <div class="flex items-start justify-between gap-2 mb-1.5">
                    <span class="text-sm font-semibold text-gray-800 dark:text-gray-100 leading-snug">
                      {{ s.document || 'Tanpa Nama Dokumen' }}
                    </span>
                    <UBadge
                      :color="store.checkSampleStatus(s) ? 'success' : 'error'"
                      size="sm"
                      variant="subtle"
                      class="shrink-0"
                    >
                      {{ store.checkSampleStatus(s) ? 'Efektif' : 'Tidak Efektif' }}
                    </UBadge>
                  </div>
                  <USeparator class="my-1.5" />
                  <div class="text-xs text-gray-500 dark:text-gray-400 italic">
                    L1: {{ s.l1 || '-' }} &nbsp;|&nbsp; L2: {{ s.l2 || '-' }} &nbsp;|&nbsp; L3: {{ s.l3 || '-' }}
                  </div>
                </UCard>

                <p v-if="!row.original.samples?.length" class="text-sm text-gray-400 dark:text-gray-500 italic">
                  Tidak ada data sampel
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