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
      class="w-full"
    >
      <template #documentName-cell="{ row }">
        <span class="font-medium">{{ row.original.documentName }}</span>
      </template>
      <template #documentNumber-cell="{ row }">
        <UBadge color="neutral" variant="subtle">{{ row.original.documentNumber }}</UBadge>
      </template>
      <template #date-cell="{ row }">
        <span>{{ formatDate(row.original.date) }}</span>
      </template>
      <template #description-cell="{ row }">
        <span class="text-sm text-gray-600">{{ row.original.description }}</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-lucide-edit" color="warning" variant="ghost" size="md" @click="store.editSample(row.original)" />
          <UButton icon="i-lucide-trash-2" color="error" variant="ghost" size="md" @click="store.deleteSample(row.index)" />
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
  { accessorKey: 'documentName', header: t('auditFieldwork.sample.columns.name') },
  { accessorKey: 'documentNumber', header: t('auditFieldwork.sample.columns.number') },
  { accessorKey: 'date', header: t('auditFieldwork.sample.columns.date') },
  { accessorKey: 'description', header: t('auditFieldwork.sample.columns.description') },
  { accessorKey: 'actions', header: t('auditFieldwork.sample.columns.actions') }
])
</script>
