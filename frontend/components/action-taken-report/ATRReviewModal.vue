<template>
  <AtrModalShell
    v-model:open="store.showReview"
    size="lg"
    :title="t('actionTakenReport.review.title')"
    :subtitle="item?.finding_title || ''"
    :busy="store.saving"
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
      <AtrFindingSummary :item="item" />

      <!-- What the PIC submitted -->
      <section class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <h4 class="text-base font-bold">
            {{ t('actionTakenReport.review.submission') }}
          </h4>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('actionTakenReport.review.submittedBy', { pic: item.pic_name || '-' }) }}
          </span>
        </div>
        <div class="space-y-1 text-sm">
          <p class="font-semibold text-gray-600 dark:text-gray-300">
            {{ t('actionTakenReport.detail.actionPlan') }}
          </p>
          <p class="whitespace-pre-line break-words">
            {{ item.action_plan || t('actionTakenReport.detail.noActionPlan') }}
          </p>
        </div>
        <div class="flex items-center gap-2 text-sm">
          <span class="font-semibold text-gray-600 dark:text-gray-300">{{ t('actionTakenReport.detail.progress') }}</span>
          <UProgress
            :model-value="item.progress"
            :max="100"
            size="sm"
            class="flex-1 max-w-xs"
          />
          <span class="tabular-nums font-semibold">{{ item.progress }}%</span>
        </div>
        <div class="space-y-2">
          <p class="text-sm font-semibold text-gray-600 dark:text-gray-300">
            {{ t('actionTakenReport.evidence.title') }}
          </p>
          <AtrEvidenceList
            :atr-id="item.id"
            :evidence="item.evidence"
          />
        </div>
      </section>

      <AppFormField
        :label="t('actionTakenReport.review.decision')"
        name="decision"
        required
        :error="errors.decision ? t(errors.decision) : undefined"
      >
        <URadioGroup
          v-model="form.decision"
          :items="decisionItems"
          orientation="horizontal"
          variant="card"
          :disabled="!canReview || store.saving"
        />
      </AppFormField>

      <AppFormField
        :label="t('actionTakenReport.review.note')"
        :description="form.decision === 'reject' ? t('actionTakenReport.review.noteRejectHelp') : t('actionTakenReport.review.noteApproveHelp')"
        name="note"
        :required="form.decision === 'reject'"
        :optional="form.decision === 'reject' ? false : t('actionTakenReport.form.optional')"
        counter
        :model-value="form.note"
        :max-count="ATR_NOTE_MAX"
        :error="errors.note ? t(errors.note) : undefined"
      >
        <template #default="{ id }">
          <UTextarea
            :id="id"
            v-model="form.note"
            :rows="4"
            :maxlength="ATR_NOTE_MAX"
            :disabled="!canReview || store.saving"
            :placeholder="t('actionTakenReport.review.notePlaceholder')"
            :color="errors.note ? 'error' : undefined"
            class="w-full"
          />
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
          :disabled="store.saving"
          @click="store.closeModal()"
        />
        <UButton
          :label="form.decision === 'reject' ? t('actionTakenReport.review.reject') : t('actionTakenReport.review.approve')"
          :icon="form.decision === 'reject' ? 'i-lucide-undo-2' : 'i-lucide-check-circle-2'"
          :color="form.decision === 'reject' ? 'warning' : 'success'"
          class="justify-center"
          :disabled="!canReview || !form.decision"
          :loading="store.saving"
          @click="save"
        />
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import { ATR_NOTE_MAX, atrFieldErrors, atrReviewSchema } from '~/utils/actionTakenReport'
import AppFormField from '~/components/shared/AppFormField.vue'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'
import AtrFindingSummary from '~/components/action-taken-report/AtrFindingSummary.vue'
import AtrEvidenceList from '~/components/action-taken-report/AtrEvidenceList.vue'

const store = useActionTakenReportStore()
const { t } = useI18n()
const { availableActions } = useAtrActions()

const item = computed(() => store.current)
const canReview = computed(() => availableActions(item.value).review)

const form = reactive<{ decision: 'approve' | 'reject' | '', note: string }>({ decision: '', note: '' })
const errors = ref<Record<string, string>>({})

const decisionItems = computed(() => [
  { label: t('actionTakenReport.review.approve'), description: t('actionTakenReport.review.approveHelp'), value: 'approve' },
  { label: t('actionTakenReport.review.reject'), description: t('actionTakenReport.review.rejectHelp'), value: 'reject' }
])

watch(
  () => [store.showReview, item.value?.id] as const,
  ([open]) => {
    if (!open) return
    form.decision = ''
    form.note = ''
    errors.value = {}
  },
  { immediate: true }
)

const save = async () => {
  if (!item.value) return
  const result = atrReviewSchema.safeParse({ decision: form.decision || undefined, note: form.note })
  errors.value = atrFieldErrors(result)
  if (!result.success) return
  if (await store.review(item.value.id, result.data)) store.closeModal()
}
</script>
