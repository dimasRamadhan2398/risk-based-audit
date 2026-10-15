<template>
  <AtrModalShell
    v-model:open="store.showActionPlan"
    size="lg"
    :title="t('actionTakenReport.actionPlanForm.title')"
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
      <span class="text-xs text-gray-500 dark:text-gray-400">
        {{ t('actionTakenReport.actionPlanForm.dueMeta', { date: formatAtrDate(item.due_date, locale) }) }}
      </span>
    </template>

    <template v-if="item">
      <UAlert
        v-if="returnedForRevision"
        color="warning"
        variant="subtle"
        icon="i-lucide-undo-2"
        :title="t('actionTakenReport.actionPlanForm.returnedTitle')"
        :description="item.review_note"
      />
      <UAlert
        v-if="!editable"
        color="neutral"
        variant="subtle"
        icon="i-lucide-lock"
        :title="t('actionTakenReport.actionPlanForm.readOnly')"
      />

      <AtrFindingSummary :item="item" />

      <AppFormField
        :label="t('actionTakenReport.actionPlanForm.actionPlan')"
        :description="t('actionTakenReport.actionPlanForm.actionPlanHelp')"
        name="action_plan"
        required
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
            :disabled="!editable || busy"
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
        required
        :error="errors.progress ? t(errors.progress) : undefined"
      >
        <template #default="{ id }">
          <div class="flex items-center gap-4">
            <USlider
              v-model="form.progress"
              :min="0"
              :max="100"
              :step="5"
              :disabled="!editable || busy"
              :aria-label="t('actionTakenReport.actionPlanForm.progress')"
              class="flex-1"
            />
            <UInputNumber
              :id="id"
              v-model="form.progress"
              :min="0"
              :max="100"
              :step="1"
              :disabled="!editable || busy"
              class="w-28"
            />
            <span class="text-sm font-semibold text-gray-500">%</span>
          </div>
        </template>
      </AppFormField>

      <AppFormField
        :label="t('actionTakenReport.evidence.title')"
        :description="t('actionTakenReport.actionPlanForm.evidenceHelp')"
        name="evidence"
        optional
      >
        <template #actions>
          <UButton
            :label="t('actionTakenReport.evidence.upload')"
            icon="i-lucide-upload"
            color="neutral"
            variant="outline"
            size="sm"
            :loading="store.uploading"
            :disabled="!editable || busy"
            @click="fileInput?.click()"
          />
        </template>
        <input
          ref="fileInput"
          type="file"
          multiple
          class="hidden"
          :aria-label="t('actionTakenReport.evidence.upload')"
          @change="onFilesSelected"
        >
        <AtrEvidenceList
          :atr-id="item.id"
          :evidence="item.evidence"
        />
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
          :label="t('actionTakenReport.actionPlanForm.saveDraft')"
          icon="i-lucide-save"
          color="neutral"
          variant="outline"
          class="justify-center"
          :disabled="!editable || busy"
          :loading="pending === 'save'"
          @click="saveDraft"
        />
        <UButton
          :label="t('actionTakenReport.actionPlanForm.submit')"
          icon="i-lucide-send"
          color="primary"
          class="justify-center"
          :disabled="!editable || busy"
          :loading="pending === 'submit'"
          @click="submitForReview"
        />
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useGlobalModalStore } from '~/stores/global-modal'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import {
  ATR_ACTION_PLAN_MAX,
  atrActionPlanSchema,
  atrFieldErrors,
  formatAtrDate,
  normalizeAtrStatus
} from '~/utils/actionTakenReport'
import AppFormField from '~/components/shared/AppFormField.vue'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'
import AtrFindingSummary from '~/components/action-taken-report/AtrFindingSummary.vue'
import AtrEvidenceList from '~/components/action-taken-report/AtrEvidenceList.vue'

const store = useActionTakenReportStore()
const { t, locale } = useI18n()
const { availableActions } = useAtrActions()

const item = computed(() => store.current)
const editable = computed(() => availableActions(item.value).actionPlan)
const returnedForRevision = computed(() => normalizeAtrStatus(item.value?.status) === 'IN_PROGRESS' && !!item.value?.review_note)

const form = reactive({ action_plan: '', progress: 0 })
const loaded = reactive({ action_plan: '', progress: 0, id: '' })
const errors = ref<Record<string, string>>({})
const pending = ref<'save' | 'submit' | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const busy = computed(() => !!pending.value || store.saving || store.uploading)
const dirty = computed(() => form.action_plan !== loaded.action_plan || form.progress !== loaded.progress)

const loadFrom = (source: { id: string, action_plan: string, progress: number }) => {
  form.action_plan = source.action_plan
  form.progress = source.progress
  Object.assign(loaded, { id: source.id, action_plan: source.action_plan, progress: source.progress })
  errors.value = {}
}

// Fill the form when the modal opens, and again when the server copy arrives, unless the user already typed.
watch(
  () => [store.showActionPlan, item.value?.id, item.value?.updated_at] as const,
  ([open]) => {
    if (!open || !item.value) return
    if (loaded.id !== item.value.id || !dirty.value) loadFrom(item.value)
  },
  { immediate: true }
)

const validate = () => {
  const result = atrActionPlanSchema.safeParse({ action_plan: form.action_plan, progress: Number(form.progress) })
  errors.value = atrFieldErrors(result)
  return result.success ? result.data : null
}

const save = async (): Promise<boolean> => {
  if (!item.value) return false
  const data = validate()
  if (!data) return false
  const ok = await store.saveActionPlan(item.value.id, data)
  if (ok && item.value) loadFrom({ id: item.value.id, action_plan: data.action_plan, progress: data.progress })
  return ok
}

const saveDraft = async () => {
  pending.value = 'save'
  try {
    await save()
  } finally {
    pending.value = null
  }
}

// Submit saves pending edits first (and moves a PLANNED item to IN_PROGRESS), then sends it for review.
const submitForReview = async () => {
  if (!item.value || !validate()) return
  const confirmed = await useGlobalModalStore().confirmSubmit({
    description: t('actionTakenReport.actionPlanForm.confirmSubmitTitle'),
    body: [t('actionTakenReport.actionPlanForm.confirmSubmitBody')],
    confirmLabel: t('actionTakenReport.actionPlanForm.submit')
  })
  if (!confirmed || !item.value) return
  pending.value = 'submit'
  try {
    const id = item.value.id
    const needsSave = dirty.value || normalizeAtrStatus(item.value.status) === 'PLANNED'
    if (needsSave && !(await save())) return
    if (await store.submit(id)) store.closeModal()
  } finally {
    pending.value = null
  }
}

const onFilesSelected = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files ?? [])
  input.value = ''
  if (!files.length || !item.value) return
  await store.uploadEvidence(item.value.id, files)
}
</script>
