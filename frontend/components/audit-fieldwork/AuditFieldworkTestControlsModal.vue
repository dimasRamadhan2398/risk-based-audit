<template>
  <!-- Test Control Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showTestControlModal"
    :ui="{ content: 'sm:max-w-4xl w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden' }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <h3 class="text-lg font-bold text-[var(--text-main)]">
            {{ store.isReadOnlyTestControl ? t('auditFieldwork.testControls.modalView') : (store.isEditingTestControl ? t('auditFieldwork.testControls.modalEdit') : t('auditFieldwork.testControls.modalAdd')) }}
          </h3>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="store.showTestControlModal = false" />
        </div>

        <div class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
            <!-- Control Information -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionControl') }}</h4>
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
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionProcedure') }}</h4>
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
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionFinding') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.finding')">
                <UTextarea v-model="store.testControlForm.finding" :placeholder="t('auditFieldwork.testControls.findingPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <UFormField :label="t('auditFieldwork.testControls.recommendation')">
                <UTextarea v-model="store.testControlForm.recommendation" :placeholder="t('auditFieldwork.testControls.recommendationPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
            </div>

            <!-- Mitigation Plan -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionMitigation') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.mitigation')">
                <UTextarea v-model="store.testControlForm.mitigationPlan" :placeholder="t('auditFieldwork.testControls.mitigationPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.pic')">
                  <UInput v-model="store.testControlForm.pic" :placeholder="t('auditFieldwork.testControls.picPlaceholder')" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.dueDate')">
                  <AppDatePicker v-model="store.testControlForm.dueDate" class="w-full" :disabled="store.isReadOnlyTestControl" />
                </UFormField>
              </div>
            </div>
          </UForm>
        </div>

        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <UButton color="neutral" variant="soft" :label="t('common.close')" v-if="store.isReadOnlyTestControl" @click="store.showTestControlModal = false" />
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="store.showTestControlModal = false" />
            <UButton color="primary" :label="store.isEditingTestControl ? t('common.edit') : t('common.submit')" @click="store.saveTestControl()" />
          </template>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useI18n } from '~/composables/useI18n'

const store = useAuditFieldworkStore()
const { t } = useI18n()
</script>
