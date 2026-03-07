<template>
  <UModal :open="props.isOpen" :ui="{ content: 'max-w-md', header: 'border-none block', body: 'shrink', footer: 'items-end text-right'  }" :title="props.title" :description="props.question"     
    scrollable
    dismissible>
      <template #header>
        <div class="flex gap-4 flex-col justify-center items-center text-center self-center">
          <div class="flex flex-row items-center justify-center gap-4">
            <div :class="iconWrapperClass">
              <UIcon :name="icon" :class="iconClass" />
            </div>
            <h3 class="text-lg font-semibold text-gray-900">{{ title }}</h3>
          </div>

          <div class="space-y-2 text-center block">
            <p v-if="question" class="text-sm text-gray-500">{{ question }}</p>
          </div>
        </div>
      </template>
      <div></div>
      <template #footer>
        <div class="flex flex-row justify-items-end justify-center gap-3 items-end text-right w-full">
          <UButton
            v-if="showCancel"
            :label="cancelText"
            variant="outline"
            color="neutral"
            @click="handleCancel"
          />
          <UButton
            :label="confirmText"
            :color="confirmColor"
            @click="handleConfirm"
          />
        </div>
      </template>
  </UModal>
</template>

<script setup lang="ts">
type ConfirmationVariant = 'danger' | 'warning' | 'info' | 'success'

type ButtonVariant = "warning" | "info" | "success" | "error" | "primary" | "secondary" | "neutral" | undefined

const props = withDefaults(defineProps<{
  isOpen?: boolean
  title?: string
  question?: string
  description?: string
  confirmText?: string
  cancelText?: string
  showCancel?: boolean
  variant?: ConfirmationVariant
}>(), {
  isOpen: false,
  title: 'Confirmation',
  confirmText: 'Yes',
  cancelText: 'No',
  showCancel: true,
  variant: 'danger',
})

const emit = defineEmits<{
  "update:isOpen": [value: boolean];
  "confirm": [];
  "cancel": [];
}>();

const isOpen = computed({
  get: () => props.isOpen,
  set: (value) => emit('update:isOpen', value)
})

const variantConfig = computed(() => {
  switch (props.variant) {
    case 'danger':
      return {
        icon: 'i-lucide-alert-triangle',
        iconWrapper: 'bg-error-100',
        iconClass: 'text-error-600',
        confirmColor: 'error'
      }
    case 'warning':
      return {
        icon: 'i-lucide-alert-circle',
        iconWrapper: 'bg-warning-100',
        iconClass: 'text-warning-600',
        confirmColor: 'warning'
      }
    case 'info':
      return {
        icon: 'i-lucide-info',
        iconWrapper: 'bg-info-100',
        iconClass: 'text-info-600',
        confirmColor: 'info'
      }
    case 'success':
      return {
        icon: 'i-lucide-check-circle',
        iconWrapper: 'bg-success-100',
        iconClass: 'text-success-600',
        confirmColor: 'success'
      }
    default:
      return {
        icon: 'i-lucide-alert-triangle',
        iconWrapper: 'bg-error-100',
        iconClass: 'text-error-600',
        confirmColor: 'error'
      }
  }
})

const icon = computed(() => variantConfig.value.icon)
const iconWrapperClass = computed(() => `rounded-full p-2 ${variantConfig.value.iconWrapper}`)
const iconClass = computed(() => `size-5 ${variantConfig.value.iconClass}`)
const confirmColor = computed(() => variantConfig.value.confirmColor as ButtonVariant)


function handleConfirm() {
  emit('update:isOpen', false)
  emit('confirm')
}

function handleCancel() {
  emit('update:isOpen', false)
}
</script>
