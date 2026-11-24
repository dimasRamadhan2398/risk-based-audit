<template>
  <div>
    <TableEntities
      :data="store.filteredPlans"
      :columns="store.columns"
      :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: t('annualAudit.table.empty') }"
      class="w-full text-sm text-left"
      :pagination="store.pagination"
    >
      <template #activity-cell="{ row }">
        <div class="py-2">
          <div class="font-bold text-gray-600 dark:text-gray-300">
            {{ (row.original || row)?.code || '-' }}
          </div>
        </div>
      </template>

      <template #department-cell="{ row }">
        <div class="flex gap-1 flex-wrap">
          <UBadge
            v-for="dept in getRowDepartments(row.original || row)"
            :key="dept"
            color="neutral"
            variant="soft"
            size="md"
          >
            {{ dept }}
          </UBadge>
        </div>
      </template>

      <template #riskName-cell="{ row }">
        <div
          v-if="summaryOf(row.original || row).activityCount"
          class="flex flex-col gap-0.5 text-sm"
        >
          <span class="font-semibold text-gray-700 dark:text-white">
            {{ t('annualAudit.table.summary.activities', { count: summaryOf(row.original || row).activityCount }) }}
          </span>
          <span class="text-gray-500 dark:text-gray-400">
            {{ t('annualAudit.table.summary.mandays', { count: summaryOf(row.original || row).totalMandays }) }}
            · {{ t('annualAudit.table.summary.auditors', { count: summaryOf(row.original || row).totalAuditors }) }}
          </span>
          <span
            v-if="monthRange(summaryOf(row.original || row).months)"
            class="text-gray-500 dark:text-gray-400"
          >
            {{ monthRange(summaryOf(row.original || row).months) }}
          </span>
        </div>
        <span
          v-else
          class="text-gray-400 dark:text-gray-500"
        >{{ t('annualAudit.table.summary.none') }}</span>
      </template>

      <template #riskLevel-cell="{ row }">
        <div
          v-if="summaryOf(row.original || row).activityCount"
          class="flex flex-col gap-1 items-start"
        >
          <div
            v-if="summaryOf(row.original || row).riskLevels.length"
            class="flex flex-wrap gap-1"
          >
            <span
              v-for="item in summaryOf(row.original || row).riskLevels"
              :key="item.level"
              class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold"
              :class="getRiskLevelColorClass(item.level)"
            >
              {{ t(RISK_LEVEL_LABEL_KEYS[item.level]) }} · {{ item.count }}
            </span>
          </div>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('annualAudit.table.summary.riskLinked', { linked: summaryOf(row.original || row).withRisk, total: summaryOf(row.original || row).activityCount }) }}
          </span>
        </div>
      </template>

      <template #status-cell="{ row }">
        <UBadge
          :color="store.getStatusBadgeColor((row.original || row)?.status)"
          variant="soft"
          size="sm"
          class="whitespace-nowrap inline-flex items-center gap-1.5"
        >
          <span
            class="size-1.5 rounded-full"
            :class="store.getStatusColor((row.original || row)?.status)"
          />
          {{ store.getStatusLabel((row.original || row)?.status) }}
        </UBadge>
      </template>

      <template #actions-cell="{ row }">
        <div class="flex items-center">
          <UTooltip :text="t('annualAudit.table.viewTooltip')">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              :title="t('common.view')"
              @click="store.openViewModal(row.original || row)"
            />
          </UTooltip>
          <UTooltip
            v-if="canWrite"
            :text="t('annualAudit.table.editTooltip')"
          >
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              :title="t('common.edit')"
              @click="store.handleEdit(row.original || row)"
            />
          </UTooltip>
          <UTooltip
            v-if="canWrite"
            :text="t('annualAudit.table.deleteTooltip')"
          >
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              :title="t('common.delete')"
              @click="store.handleDelete((row.original || row)?.id)"
            />
          </UTooltip>
        </div>
      </template>
    </TableEntities>
  </div>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useI18n } from '~/composables/useI18n'
import { getRiskLevelColorClass, RISK_LEVEL_LABEL_KEYS } from '~/utils/riskLevelBadge'
import { summarizeAnnualPlan } from '~/utils/annualPlanSummary'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
const { t } = useI18n()
const { canWrite } = useRbac()

// Per-plan roll-up (count, mandays, auditors, months, risk coverage); the list never renders one element per activity.
const summaryOf = (plan: Parameters<typeof summarizeAnnualPlan>[0]) => summarizeAnnualPlan(plan, act => store.activityRiskName(act as Parameters<typeof store.activityRiskName>[0]))

// "Jan - Dec" for a contiguous run, otherwise the month names joined.
const monthRange = (months: number[]): string => {
  if (!months.length) return ''
  const name = (m: number) => store.monthsList[m] || String(m + 1)
  if (months.length === 1) return name(months[0]!)
  const contiguous = months[months.length - 1]! - months[0]! + 1 === months.length
  return contiguous ? `${name(months[0]!)} - ${name(months[months.length - 1]!)}` : months.map(name).join(', ')
}

const getRowDepartments = (plan: any): string[] => {
  const depts = new Set<string>()
  if (plan.department?.name) depts.add(plan.department.name)
  if (plan.department && typeof plan.department === 'string') depts.add(plan.department)

  if (Array.isArray(plan.activities)) {
    plan.activities.forEach((act: any) => {
      // Server snapshot, else the departmentId resolved against master data; a row has one department now.
      const name = store.activityDepartmentName(act)
      if (name) depts.add(name)
    })
  }
  return Array.from(depts).filter(Boolean)
}
</script>
