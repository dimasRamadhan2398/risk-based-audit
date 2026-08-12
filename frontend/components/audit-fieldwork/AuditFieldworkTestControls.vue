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

    <!-- Test Controls List -->
    <UCard v-if="store.testControls.length > 0" :ui="{ body: 'p-4' }">
      <UTable :data="store.testControls" :columns="columns">
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
        <template #actions-cell="{ row }">
          <div class="flex items-center">
            <UButton icon="i-heroicons-eye" color="neutral" variant="ghost" size="sm" @click="store.viewTestControl(row.original)" />
            <UButton icon="i-heroicons-pencil-square" color="primary" variant="ghost" size="sm" @click="store.editTestControl(row.original)" />
            <UButton icon="i-heroicons-trash" color="error" variant="ghost" size="sm" @click="store.deleteTestControl(row.index)" />
          </div>
        </template>
      </UTable>
    </UCard>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <UIcon name="i-heroicons-shield-check" class="size-12 text-gray-300 mx-auto mb-2" />
      <p class="text-gray-500">{{ t('auditFieldwork.testControls.empty') }}</p>
      <UButton color="primary" variant="soft" class="mt-2" :label="t('auditFieldwork.testControls.addBtn')" @click="store.openTestControlModal()" />
    </div>

    <!-- Test Control Modal -->
    <Teleport to="body">
      <div v-if="store.showTestControlModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">
                {{ store.isReadOnlyTestControl ? t('auditFieldwork.testControls.modalView') : (store.isEditingTestControl ? t('auditFieldwork.testControls.modalEdit') : t('auditFieldwork.testControls.modalAdd')) }}
              </h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="() => { store.showTestControlModal = false }" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
            <!-- Control Information -->
            <div class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700 dark:text-gray-200">{{ t('auditFieldwork.testControls.sectionControl') }}</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.name')" required>
                  <UInput v-model="store.testControlForm.controlName" :placeholder="t('auditFieldwork.testControls.namePlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.type')" required>
                  <USelectMenu v-model="store.testControlForm.controlType" :items="store.options.controlTypes" :placeholder="t('auditFieldwork.testControls.typePlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
              </div>
              <UFormField :label="t('auditFieldwork.testControls.description')" required>
                <UTextarea v-model="store.testControlForm.controlDescription" :placeholder="t('auditFieldwork.testControls.descriptionPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
              </UFormField>
            </div>

            <!-- Test Procedure -->
            <div class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700 dark:text-gray-200">{{ t('auditFieldwork.testControls.sectionProcedure') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.steps')" required>
                <UTextarea v-model="store.testControlForm.testProcedure" :placeholder="t('auditFieldwork.testControls.stepsPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.result')" required>
                  <USelectMenu v-model="store.testControlForm.testResult" :items="store.options.testResults" :placeholder="t('auditFieldwork.testControls.resultPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
              </div>
            </div>

            <!-- Finding and Recommendation -->
            <div class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700 dark:text-gray-200">{{ t('auditFieldwork.testControls.sectionFinding') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.finding')">
                <UTextarea v-model="store.testControlForm.finding" :placeholder="t('auditFieldwork.testControls.findingPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <UFormField :label="t('auditFieldwork.testControls.recommendation')">
                <UTextarea v-model="store.testControlForm.recommendation" :placeholder="t('auditFieldwork.testControls.recommendationPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
            </div>

            <!-- Mitigation Plan -->
            <div class="bg-gray-50 dark:bg-gray-800/40 p-4 rounded-lg space-y-4">
              <h4 class="font-medium text-gray-700 dark:text-gray-200">{{ t('auditFieldwork.testControls.sectionMitigation') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.mitigation')">
                <UTextarea v-model="store.testControlForm.mitigationPlan" :placeholder="t('auditFieldwork.testControls.mitigationPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.pic')">
                  <UInput v-model="store.testControlForm.pic" :placeholder="t('auditFieldwork.testControls.picPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.dueDate')">
                  <UInput v-model="store.testControlForm.dueDate" type="date" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
              </div>
            </div>
          </UForm>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" :label="t('common.close')" v-if="store.isReadOnlyTestControl" @click="() => { store.showTestControlModal = false }" />
              <template v-else>
                <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showTestControlModal = false }" />
                <UButton color="primary" :label="store.isEditingTestControl ? t('common.edit') : t('common.submit')" @click="store.saveTestControl()" />
              </template>
            </div>
          </template>
        </UCard>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'
import { computed } from 'vue'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const columns = computed(() => [
  { accessorKey: 'controlName', header: t('auditFieldwork.testControls.columns.name') },
  { accessorKey: 'controlType', header: t('auditFieldwork.testControls.columns.type') },
  { accessorKey: 'testResult', header: t('auditFieldwork.testControls.columns.result') },
  { accessorKey: 'finding', header: t('auditFieldwork.testControls.columns.finding') },
  { accessorKey: 'mitigationPlan', header: t('auditFieldwork.testControls.columns.mitigation') },
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
