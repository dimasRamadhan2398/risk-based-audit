<template>
  <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
    <TableEntities
      :data="store.filteredPlans"    
      :columns="columns"
      :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: t('auditActivityPlan.emptyState') }"
      class="w-full text-sm text-left"
    >
      <template #planTitle-cell="{ row }">
        <div class="max-w-[280px] whitespace-normal break-words font-medium text-gray-900 dark:text-white">
          {{ getOriginal(row).planTitle || '-' }}
        </div>
      </template>

      <template #riskName-cell="{ row }">
        <div class="flex flex-col gap-1">
          <div 
            v-for="(act, idx) in getOriginal(row).plannedActivities" 
            :key="idx"
            class="text-md font-semibold text-gray-700 dark:text-white truncate max-w-[200px]"
            :title="act.riskName"
          >
            {{ act.riskName || '-' }}
          </div>
        </div>
      </template>

      <template #riskLevel-cell="{ row }">
        <div class="flex flex-col gap-1 items-start">
          <UBadge 
            v-for="(act, idx) in getOriginal(row).plannedActivities" 
            :key="idx"
            :color="store.getRiskLevelColor ? store.getRiskLevelColor(act.riskLevel) : 'neutral'"
            variant="soft"
            size="md"
          >
            {{ act.riskLevel || '-' }}
          </UBadge>
        </div>
      </template>

      <template #attachments-cell="{ row }">
        <div class="flex items-center justify-center gap-1">
          <UButton
            v-if="getOriginal(row).attachments?.length"
            v-for="(file, idx) in getOriginal(row).attachments"
            :key="idx"
            :to="file.url"
            target="_blank"
            icon="i-heroicons-document-arrow-down"
            color="primary"
            variant="ghost"
            size="sm"
            :title="file.name"
          />
          <span v-else class="text-gray-400 text-md">-</span>
        </div>
      </template>

      <template #actions-cell="{ row }">
        <div class="flex items-center justify-center gap-2">
          <UButton
            icon="i-lucide-eye"
            color="neutral"
            variant="ghost"
            size="md"
            @click="store.openViewModal(getOriginal(row))"
          />
          
          <UButton
            icon="i-lucide-edit"
            color="warning"
            variant="ghost"
            size="md"
            @click="store.handleEdit(getOriginal(row))"
          />

          <UButton
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            size="md"
            @click="store.handleDelete(getOriginal(row).id)"
          />
        </div>
      </template>
    </TableEntities>
    <ActivityPlanViewModal />
  </UCard>
  
</template>


<script setup lang="ts">
import { computed } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import ActivityPlanViewModal from '~/components/audit-activity-plan/ActivityPlanViewModal.vue'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useActivityPlanStore()
const getOriginal = (row: any) => row.original as any

const columns = computed(() => [
  { accessorKey: 'planTitle', header: t('auditActivityPlan.table.title'), class: 'max-w-[280px] whitespace-normal break-words font-medium' },
  { accessorKey: 'period', header: t('auditActivityPlan.table.period'), class: 'min-w-[220px] whitespace-nowrap' },
  { accessorKey: 'department', header: t('auditActivityPlan.table.department'), class: 'w-36' },
  { accessorKey: 'riskName', header: t('auditActivityPlan.table.riskName'), class: 'w-48' },
  { accessorKey: 'riskLevel', header: t('auditActivityPlan.table.riskLevel'), class: 'w-28 whitespace-nowrap' },
  { accessorKey: 'attachments', header: t('auditActivityPlan.table.attachment'), class: 'w-28 whitespace-nowrap text-center' },
  { accessorKey: 'actions', header: t('auditActivityPlan.table.actions'), class: 'w-28 whitespace-nowrap text-center' }
])
</script>