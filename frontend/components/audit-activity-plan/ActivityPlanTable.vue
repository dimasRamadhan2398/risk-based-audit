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

      <!-- Period: start and end on separate lines so long dates fit the fixed (frozen) column width -->
      <template #period-cell="{ row }">
        <div class="flex flex-col leading-snug">
          <template v-if="periodParts(getOriginal(row).period).length > 1">
            <span class="whitespace-nowrap">{{ periodParts(getOriginal(row).period)[0] }} –</span>
            <span class="whitespace-nowrap">{{ periodParts(getOriginal(row).period)[1] }}</span>
          </template>
          <span v-else>{{ getOriginal(row).period || '-' }}</span>
        </div>
      </template>

      <!-- Activity ID per planned activity, in the same order as Risk Name / Risk Level -->
      <template #activityCode-cell="{ row }">
        <div class="flex flex-col gap-1">
          <span
            v-for="(act, idx) in getOriginal(row).plannedActivities"
            :key="idx"
            class="whitespace-nowrap"
            :class="act.activityCode ? 'font-mono text-xs font-semibold text-gray-800 dark:text-gray-100' : 'text-gray-400 dark:text-gray-500'"
          >
            {{ act.activityCode || '-' }}
          </span>
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
          <span
            v-for="(act, idx) in getOriginal(row).plannedActivities"
            :key="idx"
            class="inline-flex items-center px-2 py-1 rounded-md text-xs font-semibold"
            :class="getRiskLevelColorClass(act.riskLevel)"
          >
            {{ act.riskLevel || '-' }}
          </span>
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
          <UTooltip text="View Plan">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              @click="store.openViewModal(getOriginal(row))"
            />
          </UTooltip>
          <UTooltip text="Edit Plan">
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              @click="store.handleEdit(getOriginal(row))"
            />
          </UTooltip>
          <UTooltip text="Delete Plan">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              @click="store.handleDelete(getOriginal(row).id)"
            />
          </UTooltip>
        </div>
      </template>
    </TableEntities>
    <ActivityPlanViewModal />
  </UCard>
  
</template>


<script setup lang="ts">
import { computed } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import ActivityPlanViewModal from '~/components/audit-activity-plan/ActivityPlanViewModal.vue'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const store = useActivityPlanStore()
const getOriginal = (row: any) => row.original as any

// formatPeriod joins start/end with " - "; split so each date gets its own line.
const periodParts = (period?: string): string[] =>
  (period || '').split(' - ').map(p => p.trim()).filter(Boolean)

const columns = computed(() => [
  // First two columns are frozen (sticky) so the rest scrolls horizontally. Fixed widths keep the
  // second column's left offset exact; same pattern as StrategicPlanTable.
  {
    accessorKey: 'planTitle',
    header: t('auditActivityPlan.table.title'),
    class: 'w-[160px] min-w-[160px] max-w-[160px] sm:w-[280px] sm:min-w-[280px] sm:max-w-[280px] whitespace-normal break-words font-medium sticky left-0',
    thClass: 'z-20 !bg-[var(--bg-surface)]',
    tdClass: 'z-10 bg-[var(--bg-main)]'
  },
  {
    accessorKey: 'period',
    header: t('auditActivityPlan.table.period'),
    class: 'w-[150px] min-w-[150px] max-w-[150px] sm:w-[220px] sm:min-w-[220px] sm:max-w-[220px] whitespace-normal break-words sticky left-[160px] sm:left-[280px] border-r border-[var(--border-main)] shadow-[2px_0_4px_-2px_rgba(0,0,0,0.15)]',
    thClass: 'z-20 !bg-[var(--bg-surface)]',
    tdClass: 'z-10 bg-[var(--bg-main)]'
  },
  { accessorKey: 'department', header: t('auditActivityPlan.table.department'), class: 'w-36' },
  { accessorKey: 'activityCode', header: t('auditActivityPlan.table.activityId'), class: 'w-32 whitespace-nowrap' },
  { accessorKey: 'riskName', header: t('auditActivityPlan.table.riskName'), class: 'w-48' },
  { accessorKey: 'riskLevel', header: t('auditActivityPlan.table.riskLevel'), class: 'w-28 whitespace-nowrap' },
  { accessorKey: 'attachments', header: t('auditActivityPlan.table.attachment'), class: 'w-28 whitespace-nowrap text-center' },
  { accessorKey: 'actions', header: t('auditActivityPlan.table.actions'), class: 'w-28 whitespace-nowrap text-center' }
])
</script>