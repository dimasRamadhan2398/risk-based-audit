<template>
  <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
    <UTable
      :data="store.filteredPlans"    
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

          <UButton
            icon="i-heroicons-trash"
            color="error"
            variant="ghost"
            size="sm"
            @click="store.handleDelete(row.original.id)"
          />
        </div>
      </template>
      
    </UTable>
    <ActivityPlanViewModal />
  </UCard>
  
</template>


<script setup lang="ts">
import { useActivityPlanStore } from '~/stores/activity-plan'
import ActivityPlanViewModal from '~/components/audit-activity-plan/ActivityPlanViewModal.vue'

const store = useActivityPlanStore()
</script>