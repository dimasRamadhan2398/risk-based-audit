<template>
  <UModal
    :open="open"
    :title="t('icAssessment.approve.title')"
    :description="subtitle"
    :ui="{
      content: 'sm:max-w-2xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl overflow-hidden',
      body: 'space-y-5 overflow-y-auto flex-1'
    }"
    @update:open="emit('update:open', $event)"
  >
    <template #body>
      <div class="p-4 rounded-xl bg-amber-50/90 dark:bg-amber-950/40 border border-amber-200 dark:border-amber-900/60 flex items-start gap-3">
        <UIcon name="i-lucide-shield-alert" class="size-5 shrink-0 text-amber-700 dark:text-amber-400 mt-0.5" />
        <p class="text-xs sm:text-sm text-amber-900 dark:text-amber-200 leading-relaxed">
          {{ t('icAssessment.approve.lockNotice') }}
        </p>
      </div>

      <UFormField
        :label="t('icAssessment.approve.operateTogether')"
        required
        :error="showErrors && operateTogether === null ? t('icAssessment.approve.operateTogetherRequired') : undefined"
      >
        <URadioGroup
          v-model="operateTogetherChoice"
          orientation="horizontal"
          :items="[
            { label: t('icAssessment.approve.yes'), value: 'yes' },
            { label: t('icAssessment.approve.no'), value: 'no' }
          ]"
        />
      </UFormField>
      <p v-if="operateTogether === false" class="text-sm text-red-700 dark:text-red-400 flex items-center gap-1.5">
        <UIcon name="i-lucide-circle-alert" class="size-4 shrink-0" />
        {{ t('icAssessment.approve.falseForcesNotEffective') }}
      </p>

      <UFormField :label="t('icAssessment.approve.operateTogetherNote')">
        <UTextarea v-model="operateTogetherNote" :rows="2" autoresize :maxlength="2000" class="w-full" />
      </UFormField>

      <UFormField :label="t('icAssessment.approve.note')">
        <UTextarea v-model="approvalNote" :rows="2" autoresize :maxlength="2000" class="w-full" />
      </UFormField>

      <div class="space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium text-gray-900 dark:text-white">{{ t('icAssessment.approve.signature') }}</span>
          <UButton
            size="xs"
            color="neutral"
            variant="outline"
            :icon="withSignature ? 'i-lucide-x' : 'i-lucide-pen-line'"
            :label="withSignature ? t('icAssessment.approve.removeSignature') : t('icAssessment.approve.addSignature')"
            @click="toggleSignature"
          />
        </div>
        <SignatureBox v-if="withSignature" v-model="signature" :height="150" />
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton color="neutral" variant="ghost" :label="t('icAssessment.actions.cancel')" @click="emit('update:open', false)" />
        <UButton color="primary" icon="i-lucide-badge-check" :loading="loading" :label="t('icAssessment.approve.confirm')" @click="confirm" />
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import SignatureBox from '~/components/shared/SignatureBox.vue'
import type { SignatureBoxValue } from '~/types/signature'
import type { IcSignature } from '~/utils/icAssessment'

const props = defineProps<{
  open: boolean
  subtitle?: string
  signerName?: string
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [value: { operateTogether: boolean, operateTogetherNote: string, approvalNote: string, signature: IcSignature | null }]
}>()

const { t } = useI18n()

const operateTogetherChoice = ref<'yes' | 'no' | undefined>(undefined)
const operateTogether = computed<boolean | null>(() => operateTogetherChoice.value === 'yes' ? true : operateTogetherChoice.value === 'no' ? false : null)
const operateTogetherNote = ref('')
const approvalNote = ref('')
const withSignature = ref(false)
const signature = ref<SignatureBoxValue>({ image: '', name: '', title: '', signedAt: '' })
const showErrors = ref(false)

watch(() => props.open, (o) => {
  if (!o) return
  operateTogetherChoice.value = undefined
  operateTogetherNote.value = ''
  approvalNote.value = ''
  withSignature.value = false
  signature.value = { image: '', name: props.signerName || '', title: '', signedAt: '' }
  showErrors.value = false
})

const toggleSignature = () => {
  withSignature.value = !withSignature.value
  if (!withSignature.value) signature.value = { image: '', name: props.signerName || '', title: '', signedAt: '' }
}

const confirm = () => {
  showErrors.value = true
  if (operateTogether.value === null) return
  emit('confirm', {
    operateTogether: operateTogether.value,
    operateTogetherNote: operateTogetherNote.value,
    approvalNote: approvalNote.value,
    // Optional (decision Q9): only sent when something was drawn
    signature: withSignature.value && signature.value.image ? { ...signature.value } : null
  })
}
</script>
