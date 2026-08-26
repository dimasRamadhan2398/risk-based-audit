<template>
  <UModal
    v-model:open="isOpen"
    :dismissible="dismissible && !loading"
    :prevent-close="preventClose || loading"
    :ui="{
      content: 'sm:max-w-md w-full bg-[var(--bg-main)] border border-[var(--border-main)] rounded-2xl shadow-2xl overflow-hidden p-0'
    }"
  >
    <template #content>
      <div class="relative p-6 sm:p-7 flex flex-col items-center text-center space-y-4">
        <!-- Close button in top right -->
        <UButton
          v-if="dismissible && !loading"
          color="neutral"
          variant="ghost"
          icon="i-lucide-x"
          size="sm"
          class="absolute top-4 right-4 text-[var(--text-muted)] hover:text-[var(--text-main)]"
          @click="handleCancel"
        />

        <!-- Icon container with soft colored ring -->
        <slot name="icon">
          <div
            class="flex items-center justify-center size-14 rounded-full transition-all duration-300"
            :class="iconContainerClasses"
          >
            <UIcon
              :name="resolvedIcon"
              class="size-7"
            />
          </div>
        </slot>

        <!-- Header: Title & Description -->
        <div class="space-y-2 max-w-sm">
          <slot name="title">
            <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)] tracking-tight">
              {{ resolvedTitle }}
            </h3>
          </slot>

          <slot name="description">
            <p
              v-if="resolvedDescription"
              class="text-sm text-[var(--text-muted)] leading-relaxed"
            >
              {{ resolvedDescription }}
            </p>
          </slot>
        </div>

        <!-- Optional Item Name Highlight Pill -->
        <div
          v-if="itemName"
          class="w-full px-3.5 py-2.5 rounded-xl border text-xs sm:text-sm font-medium flex items-center justify-center gap-2 truncate"
          :class="itemHighlightClasses"
        >
          <UIcon
            :name="isDelete ? 'i-lucide-alert-triangle' : 'i-lucide-info'"
            class="size-4 shrink-0"
          />
          <span class="truncate font-semibold">{{ itemName }}</span>
        </div>

        <!-- Extra body content slot -->
        <div
          v-if="$slots.default"
          class="w-full text-left pt-1"
        >
          <slot />
        </div>

        <!-- Action Buttons Footer -->
        <slot name="footer">
          <div class="flex items-center justify-center gap-3 w-full pt-3">
            <UButton
              v-if="showCancel"
              :label="resolvedCancelText"
              :icon="cancelIcon"
              variant="outline"
              color="neutral"
              size="lg"
              :disabled="loading"
              class="flex-1 justify-center rounded-xl font-semibold"
              @click="handleCancel"
            />
            <UButton
              :label="resolvedConfirmText"
              :icon="resolvedConfirmIcon"
              :color="resolvedConfirmColor"
              variant="solid"
              size="lg"
              :loading="loading"
              class="flex-1 justify-center rounded-xl font-semibold shadow-md"
              @click="handleConfirm"
            />
          </div>
        </slot>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type ModalType = 'confirmation' | 'delete'
type ButtonColor = 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success' | 'neutral'

interface Props {
  open?: boolean
  modelValue?: boolean
  isOpen?: boolean
  type?: ModalType
  variant?: 'confirmation' | 'delete' | 'danger' | 'warning' | 'info' | 'success'
  title?: string
  description?: string
  question?: string
  itemName?: string
  confirmText?: string
  cancelText?: string
  confirmColor?: ButtonColor
  confirmIcon?: string
  cancelIcon?: string
  icon?: string
  loading?: boolean
  showCancel?: boolean
  dismissible?: boolean
  preventClose?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  open: undefined,
  modelValue: undefined,
  isOpen: undefined,
  type: 'confirmation',
  variant: undefined,
  title: undefined,
  description: undefined,
  question: undefined,
  itemName: undefined,
  confirmText: undefined,
  cancelText: undefined,
  confirmColor: undefined,
  confirmIcon: undefined,
  cancelIcon: undefined,
  icon: undefined,
  loading: false,
  showCancel: true,
  dismissible: true,
  preventClose: false
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  'update:modelValue': [value: boolean]
  'update:isOpen': [value: boolean]
  'confirm': []
  'cancel': []
}>()

// Two-way open binding supporting v-model:open, v-model, and v-model:isOpen
const isOpen = computed({
  get: () => {
    if (props.open !== undefined) return props.open
    if (props.modelValue !== undefined) return props.modelValue
    if (props.isOpen !== undefined) return props.isOpen
    return false
  },
  set: (val: boolean) => {
    emit('update:open', val)
    emit('update:modelValue', val)
    emit('update:isOpen', val)
  }
})

// Determine if this is a delete/destructive modal
const isDelete = computed(() => {
  if (props.type === 'delete') return true
  if (props.variant === 'delete' || props.variant === 'danger') return true
  return false
})

// Icon container styling
const iconContainerClasses = computed(() => {
  if (isDelete.value) {
    return 'bg-error-50 dark:bg-error-950/50 text-error-600 dark:text-error-400 ring-8 ring-error-500/10 dark:ring-error-500/20'
  }
  if (props.variant === 'warning') {
    return 'bg-warning-50 dark:bg-warning-950/50 text-warning-600 dark:text-warning-400 ring-8 ring-warning-500/10 dark:ring-warning-500/20'
  }
  if (props.variant === 'info') {
    return 'bg-info-50 dark:bg-info-950/50 text-info-600 dark:text-info-400 ring-8 ring-info-500/10 dark:ring-info-500/20'
  }
  if (props.variant === 'success') {
    return 'bg-success-50 dark:bg-success-950/50 text-success-600 dark:text-success-400 ring-8 ring-success-500/10 dark:ring-success-500/20'
  }
  // Default confirmation: Primary
  return 'bg-primary-50 dark:bg-primary-950/50 text-primary-600 dark:text-primary-400 ring-8 ring-primary-500/10 dark:ring-primary-500/20'
})

// Item highlight container styling
const itemHighlightClasses = computed(() => {
  if (isDelete.value) {
    return 'bg-error-50/50 dark:bg-error-950/30 border-error-200 dark:border-error-800/60 text-error-800 dark:text-error-300'
  }
  return 'bg-primary-50/50 dark:bg-primary-950/30 border-primary-200 dark:border-primary-800/60 text-primary-800 dark:text-primary-300'
})

// Resolved Icon
const resolvedIcon = computed(() => {
  if (props.icon) return props.icon
  if (isDelete.value) return 'i-lucide-trash-2'
  if (props.variant === 'warning') return 'i-lucide-alert-triangle'
  if (props.variant === 'info') return 'i-lucide-info'
  if (props.variant === 'success') return 'i-lucide-check-circle-2'
  return 'i-lucide-help-circle'
})

// Resolved Title
const resolvedTitle = computed(() => {
  if (props.title) return props.title
  if (isDelete.value) return 'Konfirmasi Hapus'
  return 'Konfirmasi Tindakan'
})

// Resolved Description
const resolvedDescription = computed(() => {
  if (props.description) return props.description
  if (props.question) return props.question
  if (isDelete.value) {
    return 'Apakah Anda yakin ingin menghapus data ini? Tindakan ini tidak dapat dibatalkan.'
  }
  return 'Apakah Anda yakin ingin melanjutkan tindakan ini?'
})

// Resolved Confirm Text
const resolvedConfirmText = computed(() => {
  if (props.confirmText) return props.confirmText
  if (isDelete.value) return 'Hapus'
  return 'Lanjutkan'
})

// Resolved Cancel Text
const resolvedCancelText = computed(() => {
  if (props.cancelText) return props.cancelText
  return 'Batal'
})

// Resolved Confirm Button Color
const resolvedConfirmColor = computed<ButtonColor>(() => {
  if (props.confirmColor) return props.confirmColor
  if (isDelete.value) return 'error'
  if (props.variant === 'warning') return 'warning'
  if (props.variant === 'success') return 'success'
  return 'primary'
})

// Resolved Confirm Button Icon
const resolvedConfirmIcon = computed(() => {
  if (props.confirmIcon !== undefined) return props.confirmIcon
  if (isDelete.value) return 'i-lucide-trash-2'
  return 'i-lucide-check'
})

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  isOpen.value = false
  emit('cancel')
}
</script>
