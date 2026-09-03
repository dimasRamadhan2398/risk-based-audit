<template>
  <div class="space-y-4">
    <!-- Header with Statistics and Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">{{ t('auditFieldwork.testControls.title') }}</h3>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.testControls.subtitle') }}</p>
      </div>
      <div class="flex gap-2">
        <div v-if="store.testControls.length > 0" class="flex gap-2 mr-2">
          <UBadge color="success" variant="solid">{{ t('auditFieldwork.testControls.effectiveCount', { count: store.effectiveControls }) }}</UBadge>
          <UBadge color="error" variant="solid">{{ t('auditFieldwork.testControls.ineffectiveCount', { count: store.ineffectiveControls }) }}</UBadge>
        </div>
        <UButton color="primary" icon="i-heroicons-plus" :label="t('auditFieldwork.testControls.addBtn')" @click="store.openTestControlModal()" />
      </div>
    </div>

    <!-- Test Controls List via TableEntities -->
    <TableEntities
      :data="store.testControls"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-shield-check',
        label: t('auditFieldwork.testControls.empty')
      }"
      class="w-full"
    >
      <template #controlName-cell="{ row }">
        <span class="font-medium">{{ row.original.controlName }}</span>
      </template>
      <template #controlType-cell="{ row }">
        <UBadge :color="getControlTypeColor(row.original.controlType)" variant="subtle">{{ row.original.controlType }}</UBadge>
      </template>
      <template #testResult-cell="{ row }">
        <UBadge :color="getResultColor(row.original.testResult)" variant="solid">{{ row.original.testResult }}</UBadge>
      </template>
      <template #finding-cell="{ row }">
        <span class="text-sm text-gray-600 line-clamp-2">{{ row.original.finding || '-' }}</span>
      </template>
      <template #mitigationPlan-cell="{ row }">
        <span class="text-sm text-gray-600 line-clamp-2">{{ row.original.mitigationPlan || '-' }}</span>
      </template>
      <template #dueDate-cell="{ row }">
        <span class="text-sm text-gray-600 whitespace-nowrap">{{ formatDate(row.original.dueDate) || '-' }}</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-lucide-eye" color="neutral" variant="ghost" size="md" @click="store.viewTestControl(row.original)" />
          <UButton icon="i-lucide-edit" color="warning" variant="ghost" size="md" @click="store.editTestControl(row.original)" />
          <UButton icon="i-lucide-trash-2" color="error" variant="ghost" size="md" @click="store.deleteTestControl(row.index)" />
        </div>
      </template>
    </TableEntities>

    <!-- Test Control Modal -->
    <AuditFieldworkTestControlsModal />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'
import { formatDate } from '~/utils/dateConverter'
import AuditFieldworkTestControlsModal from '~/components/audit-fieldwork/AuditFieldworkTestControlsModal.vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { accessorKey: 'controlName', header: t('auditFieldwork.testControls.columns.name') },
  { accessorKey: 'controlType', header: t('auditFieldwork.testControls.columns.type') },
  { accessorKey: 'testResult', header: t('auditFieldwork.testControls.columns.result') },
  { accessorKey: 'finding', header: t('auditFieldwork.testControls.columns.finding') },
  { accessorKey: 'mitigationPlan', header: t('auditFieldwork.testControls.columns.mitigation') },
  { accessorKey: 'dueDate', header: t('auditFieldwork.testControls.columns.dueDate') || t('auditFieldwork.testControls.dueDate') },
  { accessorKey: 'actions', header: t('auditFieldwork.testControls.columns.actions') }
])

const getControlTypeColor = (type: string) => {
  const colors: Record<string, "success" | "warning" | "info" | "neutral" | "primary"> = {
    'Preventive': 'success',
    'Detective': 'warning',
    'Corrective': 'info',
    'Manual': 'neutral',
    'Automated': 'primary'
  }
  return colors[type] || 'neutral'
}

const getResultColor = (result: string) => {
  const colors: Record<string, "success" | "error" | "warning" | "neutral"> = {
    'Effective': 'success',
    'Ineffective': 'error',
    'Partially Effective': 'warning',
    'Not Tested': 'neutral'
  }
  return colors[result] || 'neutral'
}
</script>
