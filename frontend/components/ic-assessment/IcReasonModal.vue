<template>
  <UModal
    :open="open"
    :title="title"
    :description="description"
    :ui="{ content: 'sm:max-w-lg bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-800 rounded-2xl' }"
    @update:open="onOpen"
  >
    <template #body>
      <UFormField :label="props.label || t('icAssessment.reason.label')" :required="props.required" :error="showError ? t('icAssessment.reason.required') : undefined">
        <UTextarea
          v-model="text"
          :rows="4"
          autoresize
          :maxlength="2000"
          :placeholder="t('icAssessment.reason.placeholder')"
          class="w-full"
        />
      </UFormField>
    </template>
    <template #footer>
      <div class="flex justify-end gap-2 w-full">
        <UButton color="neutral" variant="ghost" :label="t('icAssessment.actions.cancel')" @click="emit('update:open', false)" />
        <UButton :color="confirmColor || 'primary'" :loading="loading" :label="confirmLabel || t('icAssessment.reason.confirm')" @click="confirm" />
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'

// Reason prompt for return / reopen / period lock and unlock (the global confirm modal has no input).
const props = defineProps<{
  open: boolean
  title: string
  description?: string
  label?: string
  required?: boolean
  confirmLabel?: string
  confirmColor?: 'primary' | 'error' | 'warning'
  loading?: boolean
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: [reason: string]
}>()

const { t } = useI18n()
const text = ref('')
const showError = ref(false)

watch(() => props.open, (o) => {
  if (o) {
    text.value = ''
    showError.value = false
  }
})

const onOpen = (o: boolean) => emit('update:open', o)

const confirm = () => {
  const reason = text.value.trim()
  if (props.required && !reason) {
    showError.value = true
    return
  }
  emit('confirm', reason)
}
</script>
