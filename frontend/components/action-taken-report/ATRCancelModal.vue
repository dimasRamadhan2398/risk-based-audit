<template>
  <AtrModalShell
    v-model:open="store.showCancel"
    :title="t('actionTakenReport.cancelModal.title')"
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
      <UAlert
        color="error"
        variant="subtle"
        icon="i-lucide-triangle-alert"
        :title="t('actionTakenReport.cancelModal.warningTitle')"
        :description="t('actionTakenReport.cancelModal.warning')"
      />

      <AppFormField
        :label="t('actionTakenReport.cancelModal.note')"
        :description="t('actionTakenReport.cancelModal.noteHelp')"
        name="note"
        required
        counter
        :model-value="note"
        :max-count="ATR_NOTE_MAX"
        :error="errors.note ? t(errors.note) : undefined"
      >
        <template #default="{ id }">
          <UTextarea
            :id="id"
            v-model="note"
            :rows="4"
            :maxlength="ATR_NOTE_MAX"
            :disabled="!canCancel || store.saving"
            :placeholder="t('actionTakenReport.cancelModal.notePlaceholder')"
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
          :label="t('actionTakenReport.cancelModal.confirm')"
          icon="i-lucide-ban"
          color="error"
          class="justify-center"
          :disabled="!canCancel"
          :loading="store.saving"
          @click="save"
        />
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import { ATR_NOTE_MAX, atrCancelSchema, atrFieldErrors } from '~/utils/actionTakenReport'
import AppFormField from '~/components/shared/AppFormField.vue'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'

const store = useActionTakenReportStore()
const { t } = useI18n()
const { availableActions } = useAtrActions()

const item = computed(() => store.current)
const canCancel = computed(() => availableActions(item.value).cancel)

const note = ref('')
const errors = ref<Record<string, string>>({})

watch(
  () => [store.showCancel, item.value?.id] as const,
  ([open]) => {
    if (!open) return
    note.value = ''
    errors.value = {}
  },
  { immediate: true }
)

const save = async () => {
  if (!item.value) return
  const result = atrCancelSchema.safeParse({ note: note.value })
  errors.value = atrFieldErrors(result)
  if (!result.success) return
  if (await store.cancel(item.value.id, result.data.note)) store.closeModal()
}
</script>
