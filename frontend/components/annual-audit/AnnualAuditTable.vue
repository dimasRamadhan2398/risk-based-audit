<template>
    <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
      <TableEntities
        :data="store.filteredPlans"    
        :columns="store.columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: t('annualAudit.table.empty') }"
        class="w-full text-sm text-left"
        :pagination="store.pagination"
      >
        <template #activity-cell="{ row }">
          <div class="py-2">
            <div class="font-bold text-gray-600 dark:text-gray-300">{{ (row.original || row)?.code || '-' }}</div>
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
          <div class="flex flex-col gap-1">
            <div 
              v-for="(act, idx) in ((row.original || row)?.activities || [])" 
              :key="idx"
              class="text-md font-semibold text-gray-700 dark:text-white truncate max-w-[200px]"
              :title="store.activityRiskName(act)"
            >
              {{ store.activityRiskName(act) || '-' }}
            </div>
          </div>
        </template>

        <template #riskLevel-cell="{ row }">
          <div class="flex flex-col gap-1 items-start">
            <span
              v-for="(act, idx) in ((row.original || row)?.activities || [])"
              :key="idx"
              class="inline-flex items-center px-2 py-1 rounded-md text-xs font-semibold"
              :class="getRiskLevelColorClass(act.riskLevel)"
            >
              {{ riskLevelLabel(act.riskLevel) }}
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
            <span class="size-1.5 rounded-full" :class="store.getStatusColor((row.original || row)?.status)" />
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
            <UTooltip v-if="canWrite" :text="t('annualAudit.table.editTooltip')">
            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              :title="t('common.edit')"
              @click="store.handleEdit(row.original || row)"
            />
            </UTooltip>
            <UTooltip v-if="canWrite" :text="t('annualAudit.table.deleteTooltip')">
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
      
    </UCard>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useI18n } from '~/composables/useI18n'
import { getRiskLevelColorClass, normalizeRiskLevel, RISK_LEVEL_LABEL_KEYS } from '~/utils/riskLevelBadge'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()
const { t } = useI18n()
const { canWrite } = useRbac()

// Risk level text from the backend (e.g. "Medium to High") shown in the active locale.
const riskLevelLabel = (level?: string | null) => {
  const key = normalizeRiskLevel(level)
  return key ? t(RISK_LEVEL_LABEL_KEYS[key]) : (level || '-')
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