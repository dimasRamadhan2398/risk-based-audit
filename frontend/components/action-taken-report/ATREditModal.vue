<template>
  <AtrModalShell
    v-model:open="store.showEdit"
    size="lg"
    :title="t('actionTakenReport.editForm.title')"
    :subtitle="item?.finding_title || ''"
    :busy="busy"
  >
    <template
      v-if="item"
      #meta
    >
      <AtrStatusBadge
        :status="item.status"
        :is-overdue="item.is_overdue"
        :overdue-days="item.overdue_days"
        size="md"
      />
    </template>

    <template v-if="item">
      <UAlert
        color="neutral"
        variant="subtle"
        icon="i-lucide-info"
        :title="t('actionTakenReport.editForm.partialHint')"
      />
      <UAlert
        v-if="!fields.dueDate"
        color="neutral"
        variant="subtle"
        icon="i-lucide-lock"
        :title="t('actionTakenReport.editForm.readOnlyDueDate')"
      />
      <UAlert
        v-if="!fields.actionPlan"
        color="neutral"
        variant="subtle"
        icon="i-lucide-lock"
        :title="t('actionTakenReport.editForm.readOnlyPlan')"
      />

      <AppFormField
        :label="t('actionTakenReport.assignment.dueDate')"
        name="due_date"
        required
        :error="errors.due_date ? t(errors.due_date) : undefined"
      >
        <template #default="{ id }">
          <AppDatePicker
            :id="id"
            v-model="form.due_date"
            :disabled="!fields.dueDate || busy"
            :placeholder="t('actionTakenReport.assignment.dueDatePlaceholder')"
          />
        </template>
      </AppFormField>

      <AppFormField
        :label="t('actionTakenReport.actionPlanForm.actionPlan')"
        :description="t('actionTakenReport.actionPlanForm.actionPlanHelp')"
        name="action_plan"
        counter
        :model-value="form.action_plan"
        :max-count="ATR_ACTION_PLAN_MAX"
        :error="errors.action_plan ? t(errors.action_plan) : undefined"
      >
        <template #default="{ id }">
          <UTextarea
            :id="id"
            v-model="form.action_plan"
            :rows="6"
            autoresize
            :maxrows="14"
            :maxlength="ATR_ACTION_PLAN_MAX"
            :disabled="!fields.actionPlan || busy"
            :placeholder="t('actionTakenReport.actionPlanForm.actionPlanPlaceholder')"
            :color="errors.action_plan ? 'error' : undefined"
            class="w-full"
          />
        </template>
      </AppFormField>

      <AppFormField
        :label="t('actionTakenReport.actionPlanForm.progress')"
        :description="t('actionTakenReport.actionPlanForm.progressHelp')"
        name="progress"
        :error="errors.progress ? t(errors.progress) : undefined"
      >
        <template #default="{ id }">
          <div class="flex items-center gap-4">
            <USlider
              v-model="form.progress"
              :min="0"
              :max="100"
              :step="5"
              :disabled="!fields.progress || busy"
              :aria-label="t('actionTakenReport.actionPlanForm.progress')"
              class="flex-1"
            />
            <UInputNumber
              :id="id"
              v-model="form.progress"
              :min="0"
              :max="100"
              :step="1"
              :disabled="!fields.progress || busy"
              class="w-28"
            />
            <span class="text-sm font-semibold text-gray-500">%</span>
          </div>
        </template>
      </AppFormField>
    </template>

    <template #footer>
      <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-2">
        <UButton
          :label="t('actionTakenReport.modal.close')"
          color="neutral"
          variant="ghost"
          class="justify-center"
          :disabled="busy"
          @click="store.closeModal()"
        />
        <UButton
          :label="t('actionTakenReport.editForm.save')"
          icon="i-lucide-save"
          color="primary"
          class="justify-center"
          :disabled="!dirty || busy"
          :loading="store.saving"
          @click="save"
        />
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { ActionTakenReport } from '~/types/audit'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import {
  ATR_ACTION_PLAN_MAX,
  atrActionPlanSchema,
  atrDateOnly,
  atrDueDateSchema,
  atrFieldErrors
} from '~/utils/actionTakenReport'
import AppFormField from '~/components/shared/AppFormField.vue'
import AppDatePicker from '~/components/shared/AppDatePicker.vue'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'

const store = useActionTakenReportStore()
const { t } = useI18n()
const { editableFields } = useAtrActions()

const item = computed(() => store.current)
// Which fields the caller may change; the others stay disabled and are never sent.
const fields = computed(() => editableFields(item.value))

const form = reactive({ action_plan: '', progress: 0, due_date: '' })
const loaded = reactive({ action_plan: '', progress: 0, due_date: '' })
const errors = ref<Record<string, string>>({})

const busy = computed(() => store.saving)
const dirty = computed(() => form.action_plan !== loaded.action_plan
  || Number(form.progress) !== loaded.progress
  || form.due_date !== loaded.due_date)

const loadFrom = (source: Pick<ActionTakenReport, 'action_plan' | 'progress' | 'due_date'>) => {
  form.action_plan = source.action_plan
  form.progress = source.progress
  form.due_date = atrDateOnly(source.due_date) ?? ''
  Object.assign(loaded, { action_plan: source.action_plan, progress: source.progress, due_date: form.due_date })
  errors.value = {}
}

// Fill the form each time the modal opens, and when the server copy arrives, unless the user already typed.
watch(
  () => [store.showEdit, item.value?.id, item.value?.updated_at] as const,
  ([open]) => {
    if (!open || !item.value) return
    if (!dirty.value) loadFrom(item.value)
  },
  { immediate: true }
)

/**
 * Builds the partial payload: only changed fields the caller may edit, each validated with its schema.
 * Returns null (and shows field errors) when something is invalid.
 */
const buildPayload = () => {
  const payload: { action_plan?: string, progress?: number, due_date?: string } = {}
  const nextErrors: Record<string, string> = {}

  if (fields.value.actionPlan && (form.action_plan !== loaded.action_plan || Number(form.progress) !== loaded.progress)) {
    const result = atrActionPlanSchema.safeParse({ action_plan: form.action_plan, progress: Number(form.progress) })
    Object.assign(nextErrors, atrFieldErrors(result))
    if (result.success) {
      if (form.action_plan !== loaded.action_plan) payload.action_plan = result.data.action_plan
      if (Number(form.progress) !== loaded.progress) payload.progress = result.data.progress
    }
  }

  if (fields.value.dueDate && form.due_date !== loaded.due_date) {
    const result = atrDueDateSchema.safeParse({ due_date: form.due_date })
    Object.assign(nextErrors, atrFieldErrors(result))
    if (result.success) payload.due_date = result.data.due_date
  }

  errors.value = nextErrors
  return Object.keys(nextErrors).length > 0 ? null : payload
}

const save = async () => {
  if (!item.value) return
  const payload = buildPayload()
  if (!payload) return
  if (Object.keys(payload).length === 0) return
  const ok = await store.updateReport(item.value.id, payload)
  if (ok) store.closeModal()
}
</script>
