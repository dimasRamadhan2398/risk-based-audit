<template>
  <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
    <UAlert
      v-if="store.errorMsg && !store.loading"
      color="error"
      variant="subtle"
      icon="i-heroicons-exclamation-triangle"
      :description="store.errorMsg"
      :actions="[{ label: t('auditActivityPlan.retry'), color: 'error', variant: 'outline', onClick: () => store.fetchPlans() }]"
      class="mb-4"
      data-testid="activity-plan-list-error"
    />

    <TableEntities
      :data="store.plans"
      :columns="columns"
      :loading="store.loading"
      :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: t('auditActivityPlan.emptyState') }"
      :ui="frozenTableUi"
      class="w-full text-sm text-left"
    >
      <template #year-cell="{ row }">
        <span class="font-semibold text-gray-900 dark:text-white">{{ getOriginal(row).year || '-' }}</span>
      </template>

      <template #title-cell="{ row }">
        <div class="flex flex-col gap-1 whitespace-normal break-words font-medium text-gray-900 dark:text-white">
          <span>{{ getOriginal(row).title || '-' }}</span>
          <UBadge
            v-if="getOriginal(row).isLegacy"
            color="neutral"
            variant="subtle"
            size="sm"
            class="w-fit"
            data-testid="legacy-badge"
          >
            {{ t('auditActivityPlan.legacy.badge') }}
          </UBadge>
        </div>
      </template>

      <template #activityType-cell="{ row }">
        <span>{{ store.activityTypeLabel(getOriginal(row).activityType) || '-' }}</span>
      </template>

      <template #pkat-cell="{ row }">
        <div
          v-if="getOriginal(row).annualPlan"
          class="flex flex-col"
        >
          <span class="font-mono text-xs font-semibold">{{ getOriginal(row).annualPlan?.code || '-' }}</span>
          <span
            v-if="getOriginal(row).annualPlan?.isActive === false"
            class="text-xs text-[var(--text-muted)]"
          >{{ t('auditActivityPlan.table.pkatInactive') }}</span>
        </div>
        <span v-else class="text-gray-400">-</span>
      </template>

      <template #rowCount-cell="{ row }">
        <span>{{ getOriginal(row).totals.rowCount }}</span>
      </template>

      <template #totalMandays-cell="{ row }">
        <span>{{ getOriginal(row).isLegacy ? '-' : getOriginal(row).totals.totalMandays }}</span>
      </template>

      <template #status-cell="{ row }">
        <UBadge
          v-if="getOriginal(row).status"
          color="neutral"
          variant="soft"
          size="sm"
        >
          {{ formatActivityStatus(getOriginal(row).status) }}
        </UBadge>
        <span v-else class="text-gray-400">-</span>
      </template>

      <template #attachments-cell="{ row }">
        <div class="flex items-center justify-center">
          <UButton
            v-if="getOriginal(row).attachmentCount"
            icon="i-heroicons-paper-clip"
            :label="String(getOriginal(row).attachmentCount)"
            color="primary"
            variant="ghost"
            size="sm"
            :title="t('auditActivityPlan.table.viewAttachments')"
            @click="store.openViewModal(getOriginal(row))"
          />
          <span v-else class="text-gray-400 text-md">-</span>
        </div>
      </template>

      <template #actions-cell="{ row }">
        <div class="flex items-center justify-center gap-2">
          <UTooltip :text="t('auditActivityPlan.table.viewTooltip')">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              :aria-label="t('auditActivityPlan.table.viewTooltip')"
              data-testid="action-view"
              @click="store.openViewModal(getOriginal(row))"
            />
          </UTooltip>
          <UTooltip
            v-if="canWrite && (!getOriginal(row).isLegacy)"
            :text="t('auditActivityPlan.table.editTooltip')"
          >
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              :aria-label="t('auditActivityPlan.table.editTooltip')"
              data-testid="action-edit"
              @click="store.handleEdit(getOriginal(row))"
            />
          </UTooltip>
          <UTooltip v-if="canWrite" :text="t('auditActivityPlan.table.deleteTooltip')">
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              :aria-label="t('auditActivityPlan.table.deleteTooltip')"
              data-testid="action-delete"
              @click="store.handleDelete(getOriginal(row))"
            />
          </UTooltip>
        </div>
      </template>
    </TableEntities>
    <ActivityPlanViewModal />
  </UCard>
</template>


<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { computed } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import ActivityPlanViewModal from '~/components/audit-activity-plan/ActivityPlanViewModal.vue'
import { useI18n } from '~/composables/useI18n'
import { formatActivityStatus } from '~/utils/activityPlanForm'
import type { ActivityPlanSummary } from '~/types/audit'

const { t } = useI18n()
const { canWrite } = useRbac()
const store = useActivityPlanStore()
const getOriginal = (row: any) => (row.original ?? row) as ActivityPlanSummary

// UTable's root is `overflow-auto` but TableEntities makes it `min-w-max`, so it never scrolls
// (TableEntities' wrapper does) and sticky cells would bind to the non-scrolling root. Make the
// root overflow-visible so the scrolling wrapper is the sticky container. Separate borders keep
// cell borders painted by the sticky cells themselves; collapsed borders stay behind on scroll.
const frozenTableUi = {
  root: 'overflow-visible [&>table]:border-separate [&>table]:border-spacing-0'
}

const columns = computed(() => [
  // First two columns are frozen (sticky) so the rest scrolls horizontally. Fixed widths keep the
  // second column's left offset exact; same pattern as StrategicPlanTable.
  {
    accessorKey: 'year',
    header: t('auditActivityPlan.table.year'),
    class: 'w-[80px] min-w-[80px] max-w-[80px] whitespace-nowrap sticky left-0',
    thClass: 'z-20 !bg-[var(--bg-surface)]',
    tdClass: 'z-10 bg-[var(--bg-main)]'
  },
  {
    accessorKey: 'title',
    header: t('auditActivityPlan.table.title'),
    class: 'w-[200px] min-w-[200px] max-w-[200px] sm:w-[320px] sm:min-w-[320px] sm:max-w-[320px] whitespace-normal break-words sticky left-[80px] border-r border-r-[var(--border-main)] shadow-[2px_0_4px_-2px_rgba(0,0,0,0.15)]',
    thClass: 'z-20 !bg-[var(--bg-surface)]',
    tdClass: 'z-10 bg-[var(--bg-main)]'
  },
  { accessorKey: 'activityType', header: t('auditActivityPlan.table.activityType'), class: 'w-44 whitespace-nowrap' },
  { accessorKey: 'pkat', header: t('auditActivityPlan.table.pkat'), class: 'w-40 whitespace-nowrap' },
  { accessorKey: 'rowCount', header: t('auditActivityPlan.table.rowCount'), class: 'w-28 whitespace-nowrap text-center' },
  { accessorKey: 'totalMandays', header: t('auditActivityPlan.table.totalMandays'), class: 'w-32 whitespace-nowrap text-center' },
  { accessorKey: 'status', header: t('auditActivityPlan.table.status'), class: 'w-32 whitespace-nowrap' },
  { accessorKey: 'attachments', header: t('auditActivityPlan.table.attachment'), class: 'w-28 whitespace-nowrap text-center' },
  { accessorKey: 'actions', header: t('auditActivityPlan.table.actions'), class: 'w-32 whitespace-nowrap text-center' }
])
</script>
