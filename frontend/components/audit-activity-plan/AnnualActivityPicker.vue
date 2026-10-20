<template>
  <UModal
    v-model:open="open"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-3xl max-h-[85vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-sm'
    }"
  >
    <template #content>
      <div
        class="flex max-h-[85vh] flex-col"
        data-testid="annual-activity-picker"
      >
        <div class="flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4">
          <div>
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ t('auditActivityPlan.picker.title') }}
            </h3>
            <p class="text-xs text-[var(--text-muted)]">
              {{ typeLabel }}
            </p>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            :aria-label="t('common.close')"
            @click="open = false"
          />
        </div>

        <div class="flex-1 overflow-y-auto p-4 space-y-3">
          <!-- Loading -->
          <div
            v-if="store.optionsLoading"
            class="flex items-center justify-center gap-2 py-10 text-sm text-[var(--text-muted)]"
          >
            <UIcon
              name="i-heroicons-arrow-path"
              class="h-5 w-5 animate-spin"
            />
            {{ t('common.loading') }}
          </div>

          <!-- Error -->
          <UAlert
            v-else-if="store.optionsError"
            color="error"
            variant="subtle"
            icon="i-heroicons-exclamation-triangle"
            :description="store.optionsError"
            :actions="[{ label: t('auditActivityPlan.retry'), color: 'error', variant: 'outline', onClick: () => store.fetchAnnualActivityOptions() }]"
          />

          <!-- Empty -->
          <div
            v-else-if="!store.annualActivityOptions.length"
            class="rounded-xl border border-dashed border-[var(--border-main)] py-10 text-center text-sm italic text-[var(--text-muted)]"
            data-testid="picker-empty"
          >
            {{ t('auditActivityPlan.picker.noMatch', { type: typeLabel }) }}
          </div>

          <!-- Options -->
          <template v-else>
            <label
              v-for="option in store.annualActivityOptions"
              :key="option.annualActivityId"
              class="flex items-start gap-3 rounded-lg border p-3 transition"
              :class="isDisabled(option)
                ? 'cursor-not-allowed border-[var(--border-main)] bg-gray-50 dark:bg-gray-900/40 opacity-70'
                : 'cursor-pointer border-[var(--border-main)] hover:border-primary-400'"
              data-testid="picker-option"
              :data-disabled="isDisabled(option) ? 'true' : 'false'"
            >
              <UCheckbox
                :model-value="selected.has(option.annualActivityId)"
                :disabled="isDisabled(option)"
                class="mt-0.5"
                @update:model-value="(value: boolean | 'indeterminate') => toggle(option, value === true)"
              />
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span
                    v-if="option.itemNumber"
                    class="text-xs font-semibold text-[var(--text-muted)]"
                  >{{ option.itemNumber }}.</span>
                  <span class="font-semibold text-[var(--text-main)] break-words">{{ option.auditObject }}</span>
                  <span
                    v-if="option.auditUniverse?.riskLevel"
                    class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-semibold"
                    :class="getRiskLevelColorClass(option.auditUniverse?.riskLevel)"
                  >{{ option.auditUniverse?.riskLevel }}</span>
                </div>
                <p class="text-xs text-[var(--text-muted)] break-words">
                  {{ [option.auditUniverse?.name, option.department?.name, option.schedule].filter(Boolean).join(' · ') }}
                </p>
                <p class="text-xs text-[var(--text-muted)]">
                  {{ t('auditActivityPlan.columns.teamSize') }}: {{ option.teamSize }} · {{ t('auditActivityPlan.columns.totalMandays') }}: {{ option.totalMandays }}
                </p>
                <p
                  v-if="option.usedBy"
                  class="text-xs font-medium text-warning-600 dark:text-warning-400"
                  data-testid="picker-used-by"
                >
                  {{ t('auditActivityPlan.picker.usedIn', { title: option.usedBy.activityPlanTitle || option.usedBy.assignmentCode || '-' }) }}
                </p>
                <p
                  v-else-if="inForm.has(option.annualActivityId)"
                  class="text-xs font-medium text-[var(--text-muted)]"
                >
                  {{ t('auditActivityPlan.picker.alreadyAdded') }}
                </p>
              </div>
            </label>
          </template>
        </div>

        <div class="flex flex-col-reverse gap-2 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:flex-row sm:justify-end">
          <UButton
            :label="t('common.cancel')"
            color="neutral"
            variant="ghost"
            @click="open = false"
          />
          <UButton
            icon="i-heroicons-plus"
            :label="t('auditActivityPlan.picker.confirm', { count: selected.size })"
            color="primary"
            :disabled="!selected.size"
            data-testid="picker-confirm"
            @click="confirm"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { useI18n } from '~/composables/useI18n'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'
import type { AnnualActivityOption } from '~/types/audit'

/**
 * Multi-select over the PKAT rows that match the plan's Jenis Kegiatan Audit
 * (GET /activity-plans/annual-activity-options). Rows used by another plan are disabled.
 */
const open = defineModel<boolean>('open', { default: false })

const { t } = useI18n()
const store = useActivityPlanStore()

const selected = ref(new Set<string>())

const inForm = computed(() => new Set(store.formState.rows.map(r => r.annualActivityId)))
const typeLabel = computed(() => store.activityTypeLabel(store.formState.activityType))

const isDisabled = (option: AnnualActivityOption) => !option.available || inForm.value.has(option.annualActivityId)

const toggle = (option: AnnualActivityOption, checked: boolean) => {
  if (isDisabled(option)) return
  const next = new Set(selected.value)
  if (checked) next.add(option.annualActivityId)
  else next.delete(option.annualActivityId)
  selected.value = next
}

const confirm = () => {
  // PKAT order, not click order.
  store.addRowsFromOptions(store.annualActivityOptions.filter(o => selected.value.has(o.annualActivityId)))
  open.value = false
}

watch(open, (isOpen) => {
  if (!isOpen) return
  selected.value = new Set()
  store.fetchAnnualActivityOptions()
}, { immediate: true })
</script>
