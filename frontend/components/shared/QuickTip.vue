<template>
  <UAlert
    v-if="isVisible"
    :icon="icon"
    :color="color"
    :variant="variant"
    :title="computedTitle"
    :description="computedDescription"
    :close="isDismissible ? { icon: 'i-heroicons-x-mark', color: 'neutral', variant: 'link' } : undefined"
    v-bind="$attrs"
    @close="handleClose"
  >
    <!-- Slot overrides -->
    <template
      v-if="$slots.icon"
      #icon="slotProps"
    >
      <slot
        name="icon"
        v-bind="slotProps || {}"
      />
    </template>

    <template
      v-if="$slots.title"
      #title="slotProps"
    >
      <slot
        name="title"
        v-bind="slotProps || {}"
      >
        {{ computedTitle }}
      </slot>
    </template>

    <template
      v-if="$slots.description || $slots.default"
      #description="slotProps"
    >
      <slot
        name="description"
        v-bind="slotProps || {}"
      >
        <slot v-bind="slotProps || {}">
          {{ computedDescription }}
        </slot>
      </slot>
    </template>

    <template
      v-if="$slots.actions"
      #actions="slotProps"
    >
      <slot
        name="actions"
        v-bind="slotProps || {}"
      />
    </template>
  </UAlert>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface Props {
  title?: string
  description?: string
  desc?: string
  icon?: string
  color?: 'primary' | 'secondary' | 'neutral' | 'info' | 'success' | 'warning' | 'error'
  variant?: 'subtle' | 'solid' | 'outline' | 'soft' | 'none'
  dismissible?: boolean
  closable?: boolean
  visible?: boolean
  modelValue?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: undefined,
  description: undefined,
  desc: undefined,
  icon: 'i-heroicons-light-bulb',
  color: 'primary',
  variant: 'subtle',
  dismissible: false,
  closable: false,
  visible: true,
  modelValue: undefined
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:visible': [value: boolean]
  'close': []
  'dismiss': []
}>()

const { t } = useI18n()

const isVisibleState = ref(props.modelValue ?? props.visible)

watch(() => props.modelValue, (val) => {
  if (val !== undefined) isVisibleState.value = val
})

watch(() => props.visible, (val) => {
  if (val !== undefined && props.modelValue === undefined) isVisibleState.value = val
})

const isVisible = computed({
  get: () => isVisibleState.value,
  set: (val: boolean) => {
    isVisibleState.value = val
    emit('update:modelValue', val)
    emit('update:visible', val)
  }
})

const computedTitle = computed(() => {
  if (props.title) return props.title
  const commonTitle = t('common.quickTip.title')
  if (commonTitle !== 'common.quickTip.title') return commonTitle
  const riskTitle = t('riskProfile.quickTip.title')
  if (riskTitle !== 'riskProfile.quickTip.title') return riskTitle
  return 'Quick Tip'
})

const computedDescription = computed(() => {
  if (props.description) return props.description
  if (props.desc) return props.desc
  const commonDesc = t('common.quickTip.desc')
  if (commonDesc !== 'common.quickTip.desc') return commonDesc
  const riskDesc = t('riskProfile.quickTip.desc')
  if (riskDesc !== 'riskProfile.quickTip.desc') return riskDesc
  return ''
})

const isDismissible = computed(() => props.dismissible || props.closable)

function handleClose() {
  isVisible.value = false
  emit('close')
  emit('dismiss')
}
</script>
