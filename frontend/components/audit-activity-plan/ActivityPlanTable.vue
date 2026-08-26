<template>
  <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
    <TableEntities
      :data="store.filteredPlans"    
      :columns="columns"
      :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: t('auditActivityPlan.emptyState') }"
      class="w-full text-sm text-left"
    >
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
        <div class="flex gap-1 flex-wrap">
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
        <div class="flex items-center gap-1">
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
        <div class="flex items-center gap-2">
          <UButton
            icon="i-heroicons-eye"
            color="primary"
            variant="ghost"
            size="sm"
            @click="store.openViewModal(getOriginal(row))"
          />
          
          <UButton
            icon="i-heroicons-pencil-square"
            color="primary"
            variant="ghost"
            size="sm"
            @click="store.handleEdit(getOriginal(row))"
          />

          <UButton
            icon="i-heroicons-trash"
            color="error"
            variant="ghost"
            size="sm"
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
  { accessorKey: 'planTitle', header: t('auditActivityPlan.table.title') },
  { accessorKey: 'planYear', header: t('auditActivityPlan.table.year') },
  { accessorKey: 'period', header: t('auditActivityPlan.table.period') },
  { accessorKey: 'department', header: t('auditActivityPlan.table.department') },
  { accessorKey: 'riskName', header: t('auditActivityPlan.table.riskName') },
  { accessorKey: 'riskLevel', header: t('auditActivityPlan.table.riskLevel') },
  { accessorKey: 'attachments', header: t('auditActivityPlan.table.attachment') },
  { accessorKey: 'actions', header: t('auditActivityPlan.table.actions') }
])
</script>