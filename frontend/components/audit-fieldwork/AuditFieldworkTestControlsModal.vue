<template>
  <!-- Test Control Modal (View / Add / Edit) -->
  <UModal 
    v-model:open="store.showTestControlModal"
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <div class="relative flex flex-col max-h-[90vh] transition-colors duration-300">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-5 border-b border-[var(--border-main)] bg-[var(--bg-surface)]">
          <div class="flex items-center gap-2">
            <div
              class="p-2 rounded-lg"
              :class="store.isReadOnlyTestControl ? 'bg-blue-500/10 text-blue-500' : 'bg-primary-500/10 text-primary-500'"
            >
              <UIcon
                :name="store.isReadOnlyTestControl ? 'i-heroicons-eye' : (store.isEditingTestControl ? 'i-heroicons-pencil-square' : 'i-heroicons-plus-circle')"
                class="w-5 h-5"
              />
            </div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isReadOnlyTestControl ? (t('auditFieldwork.testControls.modalView') || 'Detail Pengujian Pengendalian') : (store.isEditingTestControl ? t('auditFieldwork.testControls.modalEdit') : t('auditFieldwork.testControls.modalAdd')) }}
            </h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" class="-my-1" @click="()=> { store.showTestControlModal = false; }" />
        </div>

        <!-- Read-Only Detail View (using UCard for each section) -->
        <div v-if="store.isReadOnlyTestControl" class="p-6 overflow-y-auto space-y-4">
          <!-- Header Info Card: Assignment Letter, Due Date, and Test Result -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <div class="p-2.5 rounded-lg bg-primary-500/10 text-primary-500">
                  <UIcon name="i-heroicons-document-text" class="w-6 h-6 text-primary-500" />
                </div>
                <div>
                  <p class="text-xs font-semibold tracking-wider text-[var(--text-muted)]">Surat Tugas</p>
                  <p class="text-sm font-bold text-[var(--text-main)]">{{ store.selectedAssignmentLetter || '-' }}</p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <UBadge v-if="store.testControlForm.testResult" :color="getResultColor(store.testControlForm.testResult)" variant="solid" size="md" class="font-semibold">
                  {{ store.testControlForm.testResult }}
                </UBadge>
                <UBadge color="warning" variant="subtle" size="md" class="font-semibold">
                  <UIcon name="i-heroicons-calendar" class="w-4 h-4 mr-1.5" />
                  {{ formatDate(store.testControlForm.dueDate) || '-' }}
                </UBadge>
              </div>
            </div>
          </UCard>

          <!-- Control Information Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-shield-check" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.sectionControl') }}</span>
                </div>
                <UBadge :color="getControlTypeColor(store.testControlForm.controlType)" variant="subtle" size="sm" class="font-medium">
                  {{ store.testControlForm.controlType || '-' }}
                </UBadge>
              </div>
            </template>
            <div class="space-y-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.name') }}</p>
                <p class="text-base font-bold text-[var(--text-main)]">{{ store.testControlForm.controlName || '-' }}</p>
              </div>
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.description') }}</p>
                <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                  {{ store.testControlForm.controlDescription || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Test Procedure Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                <UIcon name="i-heroicons-clipboard-document-check" class="w-4 h-4 text-primary-500" />
                <span>{{ t('auditFieldwork.testControls.sectionProcedure') }}</span>
              </div>
            </template>
            <div class="space-y-3">
              <div>
                <p class="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)] mb-1">{{ t('auditFieldwork.testControls.steps') }}</p>
                <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                  {{ store.testControlForm.testProcedure || '-' }}
                </p>
              </div>
              <div class="flex items-center gap-2 pt-1">
                <span class="text-xs font-semibold text-[var(--text-muted)]">{{ t('auditFieldwork.testControls.result') }}:</span>
                <UBadge :color="getResultColor(store.testControlForm.testResult)" variant="solid" size="sm" class="font-medium">
                  {{ store.testControlForm.testResult || '-' }}
                </UBadge>
              </div>
            </div>
          </UCard>

          <!-- Finding and Recommendation Card -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Finding -->
            <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-amber-500">
                  <UIcon name="i-heroicons-exclamation-triangle" class="w-4 h-4 text-amber-500" />
                  <span>{{ t('auditFieldwork.testControls.finding') }}</span>
                </div>
              </template>
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.finding || '-' }}
              </p>
            </UCard>

            <!-- Recommendation -->
            <UCard color="primary" variant="outline" class="border border-primary-500/20 shadow-xs">
              <template #header>
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-light-bulb" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.recommendation') }}</span>
                </div>
              </template>
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.recommendation || '-' }}
              </p>
            </UCard>
          </div>

          <!-- Mitigation Plan Card -->
          <UCard color="primary" variant="subtle" class="border border-primary-500/20 shadow-xs">
            <template #header>
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary-500">
                  <UIcon name="i-heroicons-wrench-screwdriver" class="w-4 h-4 text-primary-500" />
                  <span>{{ t('auditFieldwork.testControls.sectionMitigation') }}</span>
                </div>
                <div v-if="store.testControlForm.pic" class="flex items-center gap-1.5 text-xs text-[var(--text-muted)]">
                  <UIcon name="i-heroicons-user-circle" class="w-4 h-4" />
                  <span class="font-medium text-[var(--text-main)]">PIC: {{ store.testControlForm.pic }}</span>
                </div>
              </div>
            </template>
            <div class="space-y-3">
              <p class="text-sm text-[var(--text-main)] leading-relaxed bg-[var(--bg-main)] p-3 rounded-lg border border-[var(--border-main)] whitespace-pre-line">
                {{ store.testControlForm.mitigationPlan || '-' }}
              </p>
              <div class="flex items-center gap-4 text-xs text-[var(--text-muted)] pt-1">
                <span>PIC: <strong class="text-[var(--text-main)]">{{ store.testControlForm.pic || '-' }}</strong></span>
                <span>Due Date: <strong class="text-[var(--text-main)]">{{ formatDate(store.testControlForm.dueDate) || '-' }}</strong></span>
              </div>
            </div>
          </UCard>
        </div>

        <!-- Add / Edit Form View -->
        <div v-else class="p-6 overflow-y-auto space-y-5">
          <UForm @submit.prevent="store.saveTestControl()" class="space-y-4">
            <!-- Control Information -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionControl') }}</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.name')" required>
                  <UInput v-model="store.testControlForm.controlName" :placeholder="t('auditFieldwork.testControls.namePlaceholder')" class="w-full" required />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.type')" required>
                  <ReusableSelectMenu v-model="store.testControlForm.controlType" :items="store.options.controlTypes" :placeholder="t('auditFieldwork.testControls.typePlaceholder')" class="w-full" required />
                </UFormField>
              </div>
              <UFormField :label="t('auditFieldwork.testControls.description')" required>
                <UTextarea v-model="store.testControlForm.controlDescription" :placeholder="t('auditFieldwork.testControls.descriptionPlaceholder')" class="w-full" required />
              </UFormField>
            </div>

            <!-- Test Procedure -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionProcedure') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.steps')" required>
                <UTextarea v-model="store.testControlForm.testProcedure" :placeholder="t('auditFieldwork.testControls.stepsPlaceholder')" class="w-full" required />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.result')" required>
                  <ReusableSelectMenu v-model="store.testControlForm.testResult" :items="store.options.testResults" :placeholder="t('auditFieldwork.testControls.resultPlaceholder')" class="w-full" required />
                </UFormField>
              </div>
            </div>

            <!-- Finding and Recommendation -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionFinding') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.finding')">
                <UTextarea v-model="store.testControlForm.finding" :placeholder="t('auditFieldwork.testControls.findingPlaceholder')" class="w-full" />
              </UFormField>
              <UFormField :label="t('auditFieldwork.testControls.recommendation')">
                <UTextarea v-model="store.testControlForm.recommendation" :placeholder="t('auditFieldwork.testControls.recommendationPlaceholder')" class="w-full" />
              </UFormField>
            </div>

            <!-- Mitigation Plan -->
            <div class="bg-[var(--bg-surface)] p-4 rounded-xl border border-[var(--border-main)] space-y-4">
              <h4 class="font-bold text-sm text-[var(--text-main)] uppercase tracking-wider">{{ t('auditFieldwork.testControls.sectionMitigation') }}</h4>
              <UFormField :label="t('auditFieldwork.testControls.mitigation')">
                <UTextarea v-model="store.testControlForm.mitigationPlan" :placeholder="t('auditFieldwork.testControls.mitigationPlaceholder')" class="w-full" />
              </UFormField>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField :label="t('auditFieldwork.testControls.pic')">
                  <UInput v-model="store.testControlForm.pic" :placeholder="t('auditFieldwork.testControls.picPlaceholder')" class="w-full" />
                </UFormField>
                <UFormField :label="t('auditFieldwork.testControls.dueDate')">
                  <AppDatePicker v-model="store.testControlForm.dueDate" class="w-full" />
                </UFormField>
              </div>
            </div>
          </UForm>
        </div>

        <!-- Modal Footer -->
        <div class="p-4 border-t border-[var(--border-main)] bg-[var(--bg-surface)] flex justify-end gap-2">
          <template v-if="store.isReadOnlyTestControl">
            <UButton color="neutral" variant="soft" :label="t('common.close') || 'Tutup'" @click="()=> { store.showTestControlModal = false }" />
            <UButton
              color="primary"
              icon="i-heroicons-pencil-square"
              :label="t('common.edit') || 'Ubah'"
              @click="switchToEditMode()"
            />
          </template>
          <template v-else>
            <UButton color="neutral" variant="soft" :label="t('common.cancel')" @click="() => {store.showTestControlModal = false}" />
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
import { formatDate } from '~/utils/dateConverter'

const store = useAuditFieldworkStore()
const { t } = useI18n()

const switchToEditMode = () => {
  store.isReadOnlyTestControl = false
  store.isEditingTestControl = true
}

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
