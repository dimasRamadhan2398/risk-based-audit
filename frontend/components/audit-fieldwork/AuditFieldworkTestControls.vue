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
      :ui="{ td: '!whitespace-normal' }"
      class="w-full"
    >
      <template #controlName-cell="{ row }">
        <span class="font-medium text-gray-900 dark:text-white">{{ row.original.controlName }}</span>
      </template>
      <template #controlType-cell="{ row }">
        <UBadge :color="getControlTypeColor(row.original.controlType)" variant="subtle">{{ row.original.controlType }}</UBadge>
      </template>
      <template #testResult-cell="{ row }">
        <UBadge :color="getResultColor(row.original.testResult)" variant="solid">{{ row.original.testResult }}</UBadge>
      </template>
      <template #finding-cell="{ row }">
        <div
          class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
          style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
        >
          {{ row.original.finding || '-' }}
        </div>
      </template>
      <template #mitigationPlan-cell="{ row }">
        <div
          class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
          style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
        >
          {{ row.original.mitigationPlan || '-' }}
        </div>
      </template>
      <template #dueDate-cell="{ row }">
        <span class="text-sm text-gray-600 dark:text-gray-300 whitespace-nowrap">{{ formatDate(row.original.dueDate) || '-' }}</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center justify-center gap-1">
          <UButton icon="i-heroicons-eye" color="neutral" variant="ghost" size="sm" :title="t('common.actions.view') || 'Lihat'" @click="store.viewTestControl(row.original)" />
          <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" :title="t('common.actions.edit') || 'Ubah'" @click="store.editTestControl(row.original)" />
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" :title="t('common.actions.delete') || 'Hapus'" @click="store.deleteTestControl(row.index)" />
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
  { key: 'controlName', accessorKey: 'controlName', header: t('auditFieldwork.testControls.columns.name'), class: 'w-56 min-w-[180px]' },
  { key: 'controlType', accessorKey: 'controlType', header: t('auditFieldwork.testControls.columns.type'), class: 'w-36 min-w-[120px] whitespace-nowrap' },
  { key: 'testResult', accessorKey: 'testResult', header: t('auditFieldwork.testControls.columns.result'), class: 'w-36 min-w-[130px] whitespace-nowrap' },
  {
    key: 'finding',
    accessorKey: 'finding',
    header: t('auditFieldwork.testControls.columns.finding'),
    class: 'w-80 min-w-[240px] max-w-sm !whitespace-normal break-words',
    tdClass: '!whitespace-normal break-words'
  },
  {
    key: 'mitigationPlan',
    accessorKey: 'mitigationPlan',
    header: t('auditFieldwork.testControls.columns.mitigation'),
    class: 'w-80 min-w-[240px] max-w-sm !whitespace-normal break-words',
    tdClass: '!whitespace-normal break-words'
  },
  { key: 'dueDate', accessorKey: 'dueDate', header: t('auditFieldwork.testControls.columns.dueDate') || t('auditFieldwork.testControls.dueDate'), class: 'w-36 min-w-[120px] whitespace-nowrap' },
  { key: 'actions', accessorKey: 'actions', header: t('auditFieldwork.testControls.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
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
