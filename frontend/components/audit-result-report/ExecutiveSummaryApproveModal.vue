<template>
  <UModal
    :open="open"
    :title="t('executiveSummaryForm.approval.title')"
    :description="t('executiveSummaryForm.approval.description', { report: reportNumber || '-' })"
    dismissible
    :ui="{
      content: 'sm:max-w-2xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
      body: 'space-y-5 overflow-y-auto flex-1'
    }"
    @update:open="emit('update:open', $event)"
  >
    <template #body>
      <div class="p-4 rounded-xl bg-amber-50/90 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900/60 flex items-start gap-3">
        <UIcon
          name="i-lucide-shield-alert"
          class="size-5 shrink-0 text-amber-700 dark:text-amber-400 mt-0.5"
        />
        <p class="text-xs sm:text-sm text-amber-900 dark:text-amber-200 leading-relaxed">
          {{ t('executiveSummaryForm.approval.lockNotice') }}
        </p>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <UFormField
          :label="t('executiveSummaryForm.signature.place')"
          required
        >
          <UInput
            v-model="place"
            :placeholder="t('executiveSummaryForm.signature.placePlaceholder')"
            maxlength="100"
            class="w-full"
          />
        </UFormField>
        <UFormField
          :label="t('executiveSummaryForm.signature.date')"
          required
        >
          <AppDatePicker
            v-model="date"
            class="w-full"
          />
        </UFormField>
      </div>

      <SignatureBox
        v-model="signature"
        :height="160"
      />
    </template>

    <template #footer>
      <div class="flex justify-between items-center gap-3 w-full">
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ canSubmit ? '' : t('executiveSummaryForm.approval.incomplete') }}
        </span>
        <div class="flex items-center gap-2">
          <UButton
            :label="t('executiveSummaryForm.actions.close')"
            color="neutral"
            variant="ghost"
            :disabled="store.loading"
            @click="emit('update:open', false)"
          />
          <UButton
            :label="t('executiveSummaryForm.approval.confirm')"
            color="success"
            icon="i-lucide-shield-check"
            :disabled="!canSubmit"
            :loading="store.loading"
            @click="submit"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useExecutiveSummaryStore } from '~/stores/executive-summary'
import { useGlobalModalStore } from '~/stores/global-modal'
import SignatureBox from '~/components/shared/SignatureBox.vue'
import type { ExecutiveSummaryApprovePayload, SignatureBoxValue } from '~/types/signature'

const { t } = useI18n()
const store = useExecutiveSummaryStore()

const props = defineProps<{
  open: boolean
  reportNumber?: string
  initialPlace?: string
  initialDate?: string
  initialSignerName?: string
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  'approved': []
}>()

const place = ref('')
const date = ref('')
const signature = ref<SignatureBoxValue>({ image: '', name: '', title: '', signedAt: '' })

// Local calendar date as YYYY-MM-DD (toISOString would give the UTC date)
const todayIso = (): string => {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

// Every opening starts from the summary's own signing details; the drawn signature is never reused
const reset = () => {
  place.value = props.initialPlace || ''
  date.value = props.initialDate || todayIso()
  signature.value = { image: '', name: props.initialSignerName || '', title: '', signedAt: '' }
}

watch(() => props.open, (isOpen) => {
  if (isOpen) reset()
}, { immediate: true })

const canSubmit = computed(() => Boolean(
  place.value.trim() && date.value && signature.value.image && signature.value.name.trim()
))

const submit = async () => {
  if (!canSubmit.value) return

  const confirmed = await useGlobalModalStore().confirmSubmit({
    description: t('executiveSummaryForm.approval.confirmDescription'),
    body: [t('executiveSummaryForm.approval.confirmBody')],
    confirmLabel: t('executiveSummaryForm.approval.confirm')
  })
  if (!confirmed) return

  const payload: ExecutiveSummaryApprovePayload = {
    signaturePlace: place.value.trim(),
    signatureDate: date.value,
    signatureImage: signature.value.image,
    signerName: signature.value.name.trim(),
    signerRole: signature.value.title.trim()
  }

  const ok = await store.approveWithSignature(payload)
  if (ok) {
    emit('approved')
    emit('update:open', false)
  }
}
</script>
