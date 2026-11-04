<template>
  <UInput
    ref="inputRef"
    :model-value="display"
    inputmode="decimal"
    autocomplete="off"
    :placeholder="placeholder"
    :disabled="disabled"
    :size="size"
    :class="$attrs.class as string"
    :ui="ui"
    @update:model-value="onInput"
    @blur="onBlur"
  >
    <template
      v-if="prefix"
      #leading
    >
      <span class="text-xs font-semibold text-gray-500 dark:text-gray-400">{{ prefix }}</span>
    </template>
  </UInput>
</template>

<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { caretAfterFormat, formatNumberInput, parseNumberInput } from '~/utils/numberInput'

defineOptions({ inheritAttrs: false })

/**
 * Text input that shows thousands separators for the current UI locale while typing
 * (en "1,234,567.5", id "1.234.567,5") but keeps the raw canonical string ("1234567.5") in v-model.
 */
const props = withDefaults(defineProps<{
  modelValue?: string | number | null
  placeholder?: string
  disabled?: boolean
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  allowNegative?: boolean
  prefix?: string
  ui?: Record<string, string>
}>(), { modelValue: '', placeholder: '', disabled: false, size: 'md', allowNegative: true, prefix: '', ui: undefined })

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const { locale } = useI18n()
const inputRef = ref<{ inputRef?: HTMLInputElement | null } | null>(null)

const display = computed(() => formatNumberInput(props.modelValue, locale.value))

const onInput = (text: string | number) => {
  const typed = String(text ?? '')
  const el = inputRef.value?.inputRef ?? null
  const caret = el?.selectionStart ?? typed.length
  const raw = parseNumberInput(typed, locale.value, { allowNegative: props.allowNegative })
  const formatted = formatNumberInput(raw, locale.value)
  // The model may already equal `raw` (e.g. a dropped letter), so the display would not re-render:
  // write the formatted text straight into the element and put the caret back.
  if (el && el.value !== formatted) {
    el.value = formatted
    const pos = caretAfterFormat(typed, caret, formatted, locale.value)
    nextTick(() => el.setSelectionRange?.(pos, pos))
  }
  if (raw !== String(props.modelValue ?? '')) emit('update:modelValue', raw)
}

// A lone "-" or trailing decimal point is only an intermediate state.
const onBlur = () => {
  const current = String(props.modelValue ?? '')
  const cleaned = current === '-' ? '' : current.replace(/\.$/, '')
  if (cleaned !== current) emit('update:modelValue', cleaned)
}
</script>
