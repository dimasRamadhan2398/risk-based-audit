<template>
  <div class="space-y-4">
    <!-- Header with Add Button -->
    <div class="flex justify-between items-center">
      <div>
        <h3 class="text-lg font-semibold">{{ t('auditFieldwork.sample.title') }}</h3>
        <p class="text-sm text-gray-500">{{ t('auditFieldwork.sample.subtitle') }}</p>
      </div>
      <UButton color="primary" icon="i-heroicons-plus" :label="t('auditFieldwork.sample.addBtn')" @click="store.openSampleModal()" />
    </div>

    <!-- Sample List via TableEntities -->
    <TableEntities
      :data="store.samples"
      :columns="columns"
      :empty-state="{
        icon: 'i-heroicons-table-cells',
        label: t('auditFieldwork.sample.empty')
      }"
      :ui="{ td: '!whitespace-normal' }"
      class="w-full"
    >
      <template #documentName-cell="{ row }">
        <span class="font-medium text-gray-900 dark:text-white">{{ row.original.documentName }}</span>
      </template>
      <template #documentNumber-cell="{ row }">
        <UBadge color="neutral" variant="subtle">{{ row.original.documentNumber }}</UBadge>
      </template>
      <template #date-cell="{ row }">
        <span class="text-sm text-[var(--text-main)]">{{ formatDate(row.original.date) }}</span>
      </template>
      <template #description-cell="{ row }">
        <div
          class="w-full min-w-0 whitespace-normal break-words leading-relaxed text-sm text-gray-600 dark:text-gray-300"
          style="white-space: normal !important; word-break: break-word !important; overflow-wrap: anywhere !important;"
        >
          {{ row.original.description || '-' }}
        </div>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center justify-center gap-1">
          <UTooltip text="Lihat Sample">
          <UButton 
            icon="i-heroicons-eye" 
            color="neutral" 
            variant="ghost" 
            size="sm" 
            :title="t('common.actions.view') || 'Lihat'" 
            @click="store.viewSample(row.original)" />
          </UTooltip>
          <UTooltip text="Ubah Sample">
          <UButton 
            icon="i-heroicons-pencil-square" 
            color="primary" 
            variant="ghost" 
            size="sm" 
            :title="t('common.actions.edit') || 'Ubah'" 
            @click="store.editSample(row.original)" />
          </UTooltip>
          <UTooltip text="Hapus Sample">
          <UButton 
            icon="i-heroicons-trash" 
            color="error" 
            variant="ghost" 
            size="sm" 
            :title="t('common.actions.delete') || 'Hapus'" 
            @click="store.deleteSample(row.index)" />
          </UTooltip>
        </div>
      </template>
    </TableEntities>

    <!-- Sample Modal -->
    <AuditFieldworkSampleModal />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'
import { formatDate } from '~/utils/dateConverter'
import AuditFieldworkSampleModal from '~/components/audit-fieldwork/AuditFieldworkSampleModal.vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { key: 'documentName', accessorKey: 'documentName', header: t('auditFieldwork.sample.columns.name'), class: 'w-56 min-w-[180px]' },
  { key: 'documentNumber', accessorKey: 'documentNumber', header: t('auditFieldwork.sample.columns.number'), class: 'w-44 min-w-[140px] whitespace-nowrap' },
  { key: 'date', accessorKey: 'date', header: t('auditFieldwork.sample.columns.date'), class: 'w-36 min-w-[120px] whitespace-nowrap' },
  {
    key: 'description',
    accessorKey: 'description',
    header: t('auditFieldwork.sample.columns.description'),
    class: 'w-80 min-w-[240px] max-w-sm !whitespace-normal break-words',
    tdClass: '!whitespace-normal break-words'
  },
  { key: 'actions', accessorKey: 'actions', header: t('auditFieldwork.sample.columns.actions'), class: 'w-24 min-w-[90px] whitespace-nowrap text-center' }
])
</script>
