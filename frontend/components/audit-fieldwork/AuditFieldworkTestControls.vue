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
      <template #actions-cell="{ row }">
        <div class="flex items-center gap-1">
          <UButton icon="i-lucide-eye" color="neutral" variant="ghost" size="md" @click="store.viewTestControl(row.original)" />
          <UButton icon="i-lucide-edit" color="warning" variant="ghost" size="md" @click="store.editTestControl(row.original)" />
          <UButton icon="i-lucide-trash-2" color="error" variant="ghost" size="md" @click="store.deleteTestControl(row.index)" />
        </div>
      </template>
    </TableEntities>

    <!-- Test Control Modal -->
    <UModal v-model:open="store.showTestControlModal" scrollable :ui="{ content: 'sm:max-w-4xl bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }">
      <template #content>
        <div class="relative flex flex-col max-h-[90vh]">
          <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlyTestControl ? t('auditFieldwork.testControls.modalView') : (store.isEditingTestControl ? t('auditFieldwork.testControls.modalEdit') : t('auditFieldwork.testControls.modalAdd')) }}
            </h3>
            <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="() => { store.showTestControlModal = false }" />
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
              <!-- Control Information -->
              <div class="bg-[var(--bg-surface)] p-4 rounded-lg space-y-4 border border-[var(--border-main)]">
                <h4 class="font-medium text-[var(--text-main)]">{{ t('auditFieldwork.testControls.sectionControl') }}</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditFieldwork.testControls.name')" required>
                    <UInput v-model="store.testControlForm.controlName" :placeholder="t('auditFieldwork.testControls.namePlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                  </UFormField>
                  <UFormField :label="t('auditFieldwork.testControls.type')" required>
                    <ReusableSelectMenu v-model="store.testControlForm.controlType" :items="store.options.controlTypes" :placeholder="t('auditFieldwork.testControls.typePlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                  </UFormField>
                </div>
                <UFormField :label="t('auditFieldwork.testControls.description')" required>
                  <UTextarea v-model="store.testControlForm.controlDescription" :placeholder="t('auditFieldwork.testControls.descriptionPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
              </div>

              <!-- Test Procedure -->
              <div class="bg-[var(--bg-surface)] p-4 rounded-lg space-y-4 border border-[var(--border-main)]">
                <h4 class="font-medium text-[var(--text-main)]">{{ t('auditFieldwork.testControls.sectionProcedure') }}</h4>
                <UFormField :label="t('auditFieldwork.testControls.steps')" required>
                  <UTextarea v-model="store.testControlForm.testProcedure" :placeholder="t('auditFieldwork.testControls.stepsPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                </UFormField>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <UFormField :label="t('auditFieldwork.testControls.result')" required>
                    <ReusableSelectMenu v-model="store.testControlForm.testResult" :items="store.options.testResults" :placeholder="t('auditFieldwork.testControls.resultPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" required />
                  </UFormField>
                </div>
              </div>

              <!-- Finding and Recommendation -->
              <div class="bg-[var(--bg-surface)] p-4 rounded-lg space-y-4 border border-[var(--border-main)]">
                <h4 class="font-medium text-[var(--text-main)]">{{ t('auditFieldwork.testControls.sectionFinding') }}</h4>
                <UFormField :label="t('auditFieldwork.testControls.finding')">
                  <UTextarea v-model="store.testControlForm.finding" :placeholder="t('auditFieldwork.testControls.findingPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.recommendation')">
                  <UTextarea v-model="store.testControlForm.recommendation" :placeholder="t('auditFieldwork.testControls.recommendationPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
              </div>

              <!-- Mitigation Plan -->
              <div class="bg-[var(--bg-surface)] p-4 rounded-lg space-y-4 border border-[var(--border-main)]">
                <h4 class="font-medium text-[var(--text-main)]">{{ t('auditFieldwork.testControls.sectionMitigation') }}</h4>
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
          </div>

          <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
            <UButton color="neutral" variant="soft" :label="t('common.close')" v-if="store.isReadOnlyTestControl" @click="() => { store.showTestControlModal = false }" />
            <template v-else>
              <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => { store.showTestControlModal = false }" />
              <UButton color="primary" :label="store.isEditingTestControl ? t('common.edit') : t('common.submit')" @click="store.saveTestControl()" />
            </template>
          </div>
        </div>
      </template>
    </UModal>
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
