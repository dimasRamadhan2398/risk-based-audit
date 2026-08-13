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
import { useI18n } from 'vue-i18n'

interface Props {
  title?: string
  description?: string
  desc?: string
  icon?: string
  color?: 'primary' | 'secondary' | 'neutral' | 'info' | 'success' | 'warning' | 'error'
  variant?: 'subtle' | 'solid' | 'outline' | 'soft'
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

// Handle i18n safely
let t: (key: string) => string = k => k
let te: (key: string) => boolean = () => false

try {
  const i18n = useI18n()
  t = i18n.t
  te = i18n.te
} catch {
  // i18n optional fallback
}

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
  if (te('common.quickTip.title')) return t('common.quickTip.title')
  if (te('riskProfile.quickTip.title')) return t('riskProfile.quickTip.title')
  return 'Quick Tip'
})

const computedDescription = computed(() => {
  if (props.description) return props.description
  if (props.desc) return props.desc
  if (te('common.quickTip.desc')) return t('common.quickTip.desc')
  if (te('riskProfile.quickTip.desc')) return t('riskProfile.quickTip.desc')
  return ''
})

const isDismissible = computed(() => props.dismissible || props.closable)

function handleClose() {
  isVisible.value = false
  emit('close')
  emit('dismiss')
}
</script>
