<template>
  <USelectMenu
    v-model="model"
    :items="items"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    :size="size"
    :color="color"
    :variant="variant"
    :icon="icon"
    :trailing-icon="trailingIcon"
    :loading="loading"
    :multiple="multiple"
    :searchable="searchable"
    :searchable-placeholder="searchablePlaceholder"
    :value-key="valueKey"
    :label-key="labelKey"
    :portal="portal"
    :ui="computedUi"
    class="w-full"
    v-bind="$attrs"
  >
    <template
      v-for="(_, slotName) in $slots"
      #[slotName]="slotProps"
    >
      <slot
        :name="slotName"
        v-bind="slotProps || {}"
      />
    </template>
  </USelectMenu>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type SelectValue = string | number | boolean | Record<string, unknown> | Array<unknown> | null | undefined

interface Props {
  modelValue?: SelectValue
  items?: Array<string | number | Record<string, unknown> | unknown>
  placeholder?: string
  disabled?: boolean
  required?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success' | 'neutral'
  variant?: 'outline' | 'soft' | 'subtle' | 'ghost' | 'none'
  icon?: string
  trailingIcon?: string
  loading?: boolean
  multiple?: boolean
  searchable?: boolean
  searchablePlaceholder?: string
  valueKey?: string
  labelKey?: string
  portal?: boolean | string | HTMLElement
  ui?: Record<string, unknown>
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: undefined,
  items: () => [],
  placeholder: '',
  disabled: false,
  required: false,
  size: 'md',
  color: undefined,
  variant: undefined,
  icon: undefined,
  trailingIcon: undefined,
  loading: false,
  multiple: false,
  searchable: false,
  searchablePlaceholder: undefined,
  valueKey: undefined,
  labelKey: undefined,
  portal: true,
  ui: undefined
})

const emit = defineEmits<{
  'update:modelValue': [value: SelectValue]
  'change': [value: SelectValue]
}>()

const model = computed({
  get: () => props.modelValue,
  set: (val: SelectValue) => {
    emit('update:modelValue', val)
    emit('change', val)
  }
})

const computedUi = computed(() => {
  const customUi = props.ui || {}
  const contentClass = ['z-[9999]', customUi.content].filter(Boolean).join(' ')
  return {
    ...customUi,
    content: contentClass
  }
})
</script>
