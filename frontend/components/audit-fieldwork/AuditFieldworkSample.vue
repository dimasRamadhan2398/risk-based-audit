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
        <span>{{ row.original.date }}</span>
      </template>
      <template #description-cell="{ row }">
        <span class="text-sm text-gray-600">{{ row.original.description }}</span>
      </template>
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editSample(row.original)" />
          <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteSample(row.index)" />
        </div>
      </template>
    </TableEntities>

    <!-- Sample Modal -->
    <Teleport to="body">
      <div v-if="store.showSampleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">{{ store.isEditingSample ? t('auditFieldwork.sample.modalEdit') : t('auditFieldwork.sample.modalAdd') }}</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showSampleModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveSample()" class="space-y-4">
            <UFormField :label="t('auditFieldwork.sample.name')" required>
              <UInput v-model="store.sampleForm.documentName" :placeholder="t('auditFieldwork.sample.namePlaceholder')" class="w-full" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField :label="t('auditFieldwork.sample.number')" required>
                <UInput v-model="store.sampleForm.documentNumber" :placeholder="t('auditFieldwork.sample.numberPlaceholder')" class="w-full" required />
              </UFormField>
              <UFormField :label="t('auditFieldwork.sample.date')" required>
                <UInput v-model="store.sampleForm.date" type="date" class="w-full" required />
              </UFormField>
            </div>

            <UFormField :label="t('auditFieldwork.sample.description')" required>
              <UTextarea v-model="store.sampleForm.description" :placeholder="t('auditFieldwork.sample.descriptionPlaceholder')" class="w-full" required />
            </UFormField>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showSampleModal = false }" />
              <UButton color="primary" :label="store.isEditingSample ? t('common.edit') : t('common.submit')" @click="store.saveSample()" />
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'

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
