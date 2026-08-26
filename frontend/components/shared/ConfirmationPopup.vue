<template>
  <ConfirmationModal
    v-model:open="isOpen"
    :title="title"
    :description="description || question"
    :item-name="itemName"
    :confirm-text="confirmText"
    :cancel-text="cancelText"
    :show-cancel="showCancel"
    :variant="variant"
    :type="variant === 'danger' ? 'delete' : 'confirmation'"
    :loading="loading"
    @confirm="emit('confirm')"
    @cancel="emit('cancel')"
  >
    <template
      v-if="$slots.default"
      #default
    >
      <slot />
    </template>
  </ConfirmationModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type ConfirmationVariant = 'danger' | 'warning' | 'info' | 'success'

const props = withDefaults(defineProps<{
  isOpen?: boolean
  open?: boolean
  modelValue?: boolean
  title?: string
  question?: string
  description?: string
  itemName?: string
  confirmText?: string
  cancelText?: string
  showCancel?: boolean
  variant?: ConfirmationVariant
  loading?: boolean
}>(), {
  isOpen: undefined,
  open: undefined,
  modelValue: undefined,
  title: undefined,
  question: undefined,
  description: undefined,
  itemName: undefined,
  confirmText: undefined,
  cancelText: undefined,
  showCancel: true,
  variant: 'danger',
  loading: false
})

const emit = defineEmits<{
  'update:isOpen': [value: boolean]
  'update:open': [value: boolean]
  'update:modelValue': [value: boolean]
  'confirm': []
  'cancel': []
}>()

const isOpen = computed({
  get: () => {
    if (props.open !== undefined) return props.open
    if (props.modelValue !== undefined) return props.modelValue
    if (props.isOpen !== undefined) return props.isOpen
    return false
  },
  set: (value) => {
    emit('update:isOpen', value)
    emit('update:open', value)
    emit('update:modelValue', value)
  }
})
</script>
