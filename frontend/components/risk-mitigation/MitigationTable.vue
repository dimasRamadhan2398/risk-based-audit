<template>
  <div class="space-y-4">
    <div class="flex justify-between items-center">
      <div class="flex items-center gap-3">
        <UIcon name="i-heroicons-clipboard-document-check" class="text-primary-600 text-2xl" />
        <h3 class="text-lg font-bold text-gray-800 dark:text-white">Daftar Rencana Mitigasi</h3>
      </div>
      <UButton 
        label="Tambah Mitigasi" 
        icon="i-heroicons-plus" 
        color="primary" 
        class="font-bold shadow-md"
        @click="store.openForm()"
      />
    </div>

    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
      <UTable :data="filteredMitigations" :columns="columns">
        <!-- <template #status-data="{ row }">
          <UBadge :color="getStatusColor(row.original.status)" variant="soft" class="font-bold">
            {{ row.original.status }}
          </UBadge>
        </template> -->

        <template #period-cell="{ row }">
          <div class="flex flex-col text-xs">
            <span class="text-gray-500 font-medium">Mulai: {{ row.original.start_date }}</span>
            <span class="text-primary-600 font-bold">Selesai: {{ row.original.end_date }}</span>
          </div>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex gap-2 justify-center">
            <UButton icon="i-heroicons-pencil-square" size="sm" color="warning" variant="ghost" @click="store.openForm(row.original)" />
            <UButton icon="i-heroicons-trash" size="sm" color="error" variant="ghost" @click="store.deleteMitigation(row.original.id)" />
          </div>
        </template>
        
        <template #empty-state>
          <div class="flex flex-col items-center justify-center py-10 text-gray-400">
            <UIcon name="i-heroicons-inbox" class="text-4xl mb-2" />
            <p>Belum ada rencana mitigasi yang ditambahkan.</p>
          </div>
        </template>

      </UTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useMitigationStore } from '~/stores/mitigation-risk'

const store = useMitigationStore()

const props = defineProps<{
  currentRiskId: string
}>()

const filteredMitigations = computed(() => {
  return store.getMitigationsByRiskId(props.currentRiskId)
})

// Definisi Kolom Tabel
const columns = [
  { accessorKey: 'actionPlan', header: 'Aktivitas' },
  { accessorKey: 'supervisor', header: 'Supervisor' },
  { accessorKey: 'pic', header: 'PIC' },
  { accessorKey: 'period', header: 'Jadwal' },
  // { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'notes', header: 'Catatan' },
  { accessorKey: 'actions', header: 'Aksi', meta: { class: { td: 'text-center', th: 'text-center' } } },
]

// Helper fungsi untuk warna status
// const getStatusColor = (status: string) => {
//   switch (status) {
//     case 'Completed': return 'success'
//     case 'In Progress': return 'primary'
//     case 'Delayed': return 'error'
//     default: return 'neutral' // Open
//   }
// }
</script>