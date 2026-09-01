<template>
    <UCard class="rounded-xl shadow overflow-y-auto" variant="soft" color="primary">
      <TableEntities
        :data="store.filteredPlans"    
        :columns="store.columns"
        :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }"
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
              :title="act.riskName"
            >
              {{ act.riskName || '-' }}
            </div>
          </div>
        </template>

        <template #riskLevel-cell="{ row }">
          <div class="flex flex-col gap-1 items-start">
            <UBadge 
              v-for="(act, idx) in ((row.original || row)?.activities || [])" 
              :key="idx"
              :color="store.getRiskLevelColor ? store.getRiskLevelColor(act.riskLevel) : 'neutral'"
              variant="soft"
              size="md"
            >
              {{ act.riskLevel || '-' }}
            </UBadge>
          </div>
        </template>

        <template #progress-cell="{ row }">
          <UProgress v-model="store.progressAudit" color="secondary" status />
        </template>

        <template #status-cell="{ row }">
          <span 
            class="w-2.5 h-2.5 rounded-full inline-block"
            :class="store.getStatusColor((row.original || row)?.status)"
          ></span>
        </template>

        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton
              icon="i-lucide-eye"
              color="neutral"
              variant="ghost"
              size="md"
              title="View"
              @click="store.openViewModal(row.original || row)"
            />

            <UButton
              icon="i-lucide-edit"
              color="warning"
              variant="ghost"
              size="md"
              title="Edit"
              @click="store.handleEdit(row.original || row)"
            />

            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              size="md"
              title="Hapus"
              @click="store.handleDelete((row.original || row)?.id)"
            />
          </div>
        </template>
      
      </TableEntities>
      
    </UCard>
</template>

<script setup lang="ts">
import { useAnnualPlanStore } from '~/stores/annual-audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()

const getRowDepartments = (plan: any): string[] => {
  const depts = new Set<string>()
  if (plan.department?.name) depts.add(plan.department.name)
  if (plan.department && typeof plan.department === 'string') depts.add(plan.department)
  
  if (Array.isArray(plan.activities)) {
    plan.activities.forEach((act: any) => {
      if (act.department) depts.add(typeof act.department === 'string' ? act.department : act.department.name)
      const inv = act.involvedDepartments || act.involved_departments
      if (Array.isArray(inv)) {
        inv.forEach((d: any) => {
          if (typeof d === 'string') depts.add(d)
          else if (d?.name) depts.add(d.name)
          else if (d?.code) depts.add(d.code)
        })
      }
    })
  }
  return Array.from(depts).filter(Boolean)
}
</script>