<script setup lang="ts">
import { computed } from 'vue'
import { EXECUTION_PHASES, getExecutionPhase, type AuditExecution } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()

const props = defineProps<{
  open: boolean
  audit?: AuditExecution
}>()

const emit = defineEmits(['update:open', 'remind'])

const isOpen = computed({
  get: () => props.open,
  set: (value) => emit('update:open', value)
})

const auditProgress = computed(() => {
  return Number(props.audit?.progress) || 0
})

const currentPhase = computed(() => {
  return getExecutionPhase(auditProgress.value)
})

const getProgressColor = (val: number) => {
  if (val >= 100) return 'bg-secondary-600 shadow-sm shadow-secondary-600/40'
  if (val >= 76) return 'bg-indigo-500 shadow-sm shadow-indigo-500/40'
  if (val >= 51) return 'bg-purple-500 shadow-sm shadow-purple-500/40'
  if (val >= 26) return 'bg-violet-500 shadow-sm shadow-violet-500/40'
  if (val >= 1) return 'bg-blue-500 shadow-sm shadow-blue-500/40'
  return 'bg-sky-400 shadow-sm shadow-sky-400/40'
}

const sendReminder = () => {
  emit('remind', props.audit)
  isOpen.value = false
}
</script>

<template>
  <UModal v-model:open="isOpen" :ui="{ content: 'sm:max-w-4xl' }" scrollable>
    <template #content>
      <UCard :ui="{ root: 'divide-y divide-gray-100 dark:divide-gray-800' }">
        <template #header>
          <div class="flex items-center justify-between">
            <div class="space-y-1">
              <div class="flex items-center gap-3">
                <h3 class="text-xl font-bold text-gray-900 dark:text-white">{{ audit?.name || t('auditExecution.detailModal.title') }}</h3>
                <UBadge :color="currentPhase.badgeColor" variant="subtle" size="md">
                  {{ t(`auditExecution.phases.${currentPhase.step}.title`) }}
                </UBadge>
              </div>
              <p class="text-sm text-gray-500">
                {{ t('auditExecution.table.ref') }} <span class="font-semibold text-gray-700 dark:text-gray-300">{{ audit?.ref }}</span> | {{ t('auditExecution.filters.categoryPlaceholder') }}: <span class="font-semibold text-gray-700 dark:text-gray-300">{{ audit?.category }}</span> | {{ t('auditExecution.columns.leadAuditor') }}: <span class="font-semibold text-gray-700 dark:text-gray-300">{{ audit?.lead_auditor || '-' }}</span>
              </p>
            </div>
            <UButton color="neutral" variant="ghost" icon="i-lucide-x" class="-my-1" @click="() => { isOpen = false }" />
          </div>
        </template>

        <div class="space-y-6">
          <!-- Lifecycle Stepper Component -->
          <div class="p-5 border border-primary-200 dark:border-primary-800/60 rounded-xl bg-gradient-to-b from-primary-50/40 to-transparent dark:from-primary-950/20 dark:to-transparent space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-route" class="text-primary-600 dark:text-primary-400 text-lg" />
                <h4 class="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wider">{{ t('auditExecution.detailModal.lifecyclePhase') }}</h4>
              </div>
              <span class="text-sm font-extrabold text-primary-600 dark:text-primary-400">
                {{ t('auditExecution.detailModal.overallProgress', { progress: auditProgress }) }}
              </span>
            </div>

            <!-- Stepper Steps Grid -->
            <div class="grid grid-cols-6 gap-2 pt-2">
              <div
                v-for="phase in EXECUTION_PHASES"
                :key="phase.step"
                class="flex flex-col items-center text-center p-2.5 rounded-lg border transition-all duration-300"
                :class="[
                  phase.step === currentPhase.step
                    ? 'border-primary-500 bg-primary-100/60 dark:bg-primary-900/40 shadow-sm ring-2 ring-primary-500/20'
                    : phase.step < currentPhase.step
                      ? 'border-emerald-300 dark:border-emerald-800/60 bg-emerald-50/50 dark:bg-emerald-950/20 text-emerald-700 dark:text-emerald-400'
                      : 'border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 opacity-60'
                ]"
              >
                <div
                  class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold mb-1.5 transition-colors"
                  :class="[
                    phase.step === currentPhase.step
                      ? 'bg-primary-600 text-white shadow-md'
                      : phase.step < currentPhase.step
                        ? 'bg-emerald-500 text-white'
                        : 'bg-gray-200 dark:bg-gray-700 text-gray-500 dark:text-gray-400'
                  ]"
                >
                  <UIcon v-if="phase.step < currentPhase.step" name="i-lucide-check" class="text-sm" />
                  <span v-else>{{ phase.step }}</span>
                </div>
                <span class="text-xs font-bold line-clamp-1" :class="phase.step === currentPhase.step ? 'text-primary-700 dark:text-primary-300' : 'text-gray-700 dark:text-gray-300'">
                  {{ t(`auditExecution.phases.${phase.step}.shortLabel`) }}
                </span>
              </div>
            </div>

            <!-- Current Active Phase Card -->
            <div class="p-3.5 bg-white dark:bg-gray-900 border border-primary-200/80 dark:border-primary-800/80 rounded-lg flex items-start gap-3 shadow-xs">
              <UIcon :name="currentPhase.icon" class="text-primary-600 dark:text-primary-400 text-2xl shrink-0 mt-0.5" />
              <div class="space-y-0.5 flex-1">
                <div class="flex items-center justify-between">
                  <span class="text-sm font-bold text-gray-900 dark:text-white">{{ t('auditExecution.detailModal.activePhase', { step: currentPhase.step, title: t(`auditExecution.phases.${currentPhase.step}.title`) }) }}</span>
                  <span class="text-xs font-semibold text-gray-500">{{ currentPhase.minProgress === currentPhase.maxProgress ? `${currentPhase.minProgress}%` : t('auditExecution.detailModal.range', { min: currentPhase.minProgress, max: currentPhase.maxProgress }) }}</span>
                </div>
                <p class="text-[11px] text-gray-500 dark:text-gray-400 leading-normal">{{ t(`auditExecution.phases.${currentPhase.step}.description`) }}</p>
              </div>
            </div>
          </div>

          <!-- Section 1: Sample Data & Test Controls -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">{{ t('auditExecution.detailModal.sampleDataTitle') }}</h4>
              <div class="grid grid-cols-4 gap-4 items-center">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.progress') }}</span>
                <div class="col-span-3 flex items-center gap-3">
                  <div class="flex-1 h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden border border-gray-200 dark:border-gray-700 shadow-inner">
                    <div
                      class="h-full rounded-full transition-all duration-500 ease-out"
                      :class="getProgressColor(audit?.sample_data_test_controls?.progress || 0)"
                      :style="{ width: `${Math.min(100, Math.max(0, audit?.sample_data_test_controls?.progress || 0))}%` }"
                    ></div>
                  </div>
                  <span class="text-sm font-bold text-gray-700 dark:text-gray-300 w-12 text-right">{{ audit?.sample_data_test_controls?.progress || 0 }} %</span>
                </div>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.description') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.sample_data_test_controls?.description || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Section 2: Working Papers -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">{{ t('auditExecution.detailModal.workingPapersTitle') }}</h4>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.condition') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.working_papers?.condition || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.criteria') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.working_papers?.criteria || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <!-- Section 3: Action Plan Improvements -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">{{ t('auditExecution.detailModal.actionPlanTitle') }}</h4>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.recommendation') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.recommendation || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.deadline') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.deadline || '-' }}
                </p>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.pic') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ audit?.action_plan_improvements?.pic || '-' }}
                </p>
              </div>
            </div>
          </UCard>

          <hr class="border-gray-200 dark:border-gray-700" />

          <!-- Section 4: Latest Update Progress -->
          <UCard variant="subtle" class="bg-gray-50/50 dark:bg-gray-800/50">
            <div class="space-y-4">
              <h4 class="font-bold text-gray-900 dark:text-white">{{ t('auditExecution.detailModal.latestUpdateTitle') }}</h4>
              <div class="grid grid-cols-4 gap-4 items-center">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.attachment') }}</span>
                <div class="col-span-3">
                  <div v-if="audit?.latest_update_progress?.attachment" class="flex items-center gap-2 text-sm text-blue-600 dark:text-blue-400">
                    <UIcon name="i-lucide-file-text" />
                    <span>[ {{ audit?.latest_update_progress?.attachment }} ]</span>
                  </div>
                  <span v-else class="text-sm text-gray-400">{{ t('auditExecution.detailModal.noAttachment') }}</span>
                </div>
              </div>
              <div class="grid grid-cols-4 gap-4">
                <span class="text-sm font-medium text-gray-500">{{ t('auditExecution.detailModal.description') }}</span>
                <p class="col-span-3 text-sm text-gray-700 dark:text-gray-300 italic">
                  {{ audit?.latest_update_progress?.description || '-' }}
                </p>
              </div>
            </div>
          </UCard>
        </div>

        <template #footer>
          <div class="flex justify-end">
            <UButton
              :label="t('auditExecution.detailModal.sendReminder')"
              color="primary"
              icon="i-lucide-bell"
              @click="sendReminder"
            />
          </div>
        </template>
      </UCard>
    </template>
  </UModal>
</template>

