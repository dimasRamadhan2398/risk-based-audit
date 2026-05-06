<template>
  <div class="space-y-4">
    <div class="flex justify-between items-center">
      <div class="flex items-center gap-3">
        <UIcon name="i-heroicons-clipboard-document-check" class="text-primary-600 text-2xl" />
        <h3 class="text-lg font-bold text-gray-800">Risk Mitigation & Controls</h3>
      </div>
      <UButton 
        label="Add Mitigation Plan" 
        icon="i-heroicons-plus" 
        color="primary" 
        class="font-bold shadow-md"
        @click="store.openForm()"
      />
    </div>

    <div class="bg-[var(--bg-main)]  border border-[var(--border-main)] rounded-xl overflow-hidden shadow-sm transition-colors duration-300">
      <UTable :data="filteredMitigations" :columns="columns">
        <template #period-cell="{ row }">
          <div class="flex flex-col text-xs">
            <span class="text-[var(--text-muted)] font-medium">Start: {{ row.original.start_date }}</span>
            <span class="text-primary-600  font-bold">End: {{ row.original.end_date }}</span>
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
            <p>No risk mitigation plans have been added yet.</p>
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
  { accessorKey: 'riskEvent', header: 'Risk Event' },
  { accessorKey: 'mitigationPlan', header: 'Risk Mitigations & Controls' },
  { accessorKey: 'period', header: 'Timeline' },
  { accessorKey: 'pic', header: 'PIC' },
  { accessorKey: 'unitInCharge', header: 'Unit in Charge' },
  { accessorKey: 'supervisor', header: 'Supervisor' },
  { accessorKey: 'notes', header: 'Notes' },
  { accessorKey: 'actions', header: 'Actions', meta: { class: { td: 'text-center', th: 'text-center' } } },
]

</script>