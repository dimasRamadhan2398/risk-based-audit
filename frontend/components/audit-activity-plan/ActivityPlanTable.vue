<template>
  <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
    <UTable
      :rows="store.filteredPlans"    
      :columns="store.columns"
      :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }"
      class="w-full text-sm text-left"
    >
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-2">
          <UButton
            icon="i-heroicons-eye"
            color="primary"
            variant="ghost"
            size="sm"
            @click="store.openViewModal(row.original)"
          />
          <UButton
            icon="i-heroicons-pencil-square"
            color="primary"
            variant="ghost"
            size="sm"
            @click="store.handleEdit(row.original)"
          />
        </div>
      </template>
    </UTable>
  </UCard>
</template>

<script setup lang="ts">
import { useActivityPlanStore } from '~/stores/activity-plan'

const store = useActivityPlanStore()
</script>