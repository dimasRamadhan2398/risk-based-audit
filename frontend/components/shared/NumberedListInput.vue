<template>
  <div
    class="space-y-2"
    :data-testid="testId"
  >
    <ol class="space-y-2">
      <li
        v-for="(item, index) in items"
        :key="index"
        class="flex items-start gap-2"
      >
        <span class="w-6 shrink-0 pt-1.5 text-right text-sm font-semibold text-[var(--text-muted)] select-none">{{ index + 1 }}.</span>
        <div class="min-w-0 flex-1">
          <UTextarea
            :model-value="item"
            :rows="1"
            autoresize
            :disabled="disabled"
            :placeholder="t('auditActivityPlan.list.itemPlaceholder', { num: index + 1 })"
            :aria-label="`${label} ${index + 1}`"
            :color="item.length > maxLength ? 'error' : undefined"
            :highlight="item.length > maxLength"
            class="w-full"
            @update:model-value="(value: string) => updateItem(index, value)"
            @paste="(event: ClipboardEvent) => onPaste(index, event)"
          />
          <p
            v-if="item.length > maxLength"
            class="mt-0.5 text-xs text-error-600 dark:text-error-400"
          >
            {{ item.length }}/{{ maxLength }}
          </p>
        </div>
        <div
          v-if="!disabled"
          class="flex shrink-0 items-center"
        >
          <UButton
            icon="i-heroicons-chevron-up"
            color="neutral"
            variant="ghost"
            size="xs"
            :disabled="index === 0"
            :aria-label="t('auditActivityPlan.rows.moveUp')"
            @click="move(index, -1)"
          />
          <UButton
            icon="i-heroicons-chevron-down"
            color="neutral"
            variant="ghost"
            size="xs"
            :disabled="index === items.length - 1"
            :aria-label="t('auditActivityPlan.rows.moveDown')"
            @click="move(index, 1)"
          />
          <UButton
            icon="i-heroicons-x-mark"
            color="error"
            variant="ghost"
            size="xs"
            :aria-label="t('auditActivityPlan.rows.remove')"
            @click="remove(index)"
          />
        </div>
      </li>
    </ol>

    <div class="flex flex-wrap items-center justify-between gap-2 pl-8">
      <UButton
        v-if="!disabled"
        icon="i-heroicons-plus"
        color="primary"
        variant="soft"
        size="xs"
        :label="t('auditActivityPlan.list.addItem')"
        :disabled="filledCount >= maxItems"
        @click="add"
      />
      <span class="text-xs text-[var(--text-muted)]">{{ t('auditActivityPlan.list.pasteHint') }}</span>
      <span
        class="ml-auto text-xs font-medium"
        :class="filledCount > maxItems ? 'text-error-600 dark:text-error-400' : 'text-[var(--text-muted)]'"
      >{{ filledCount }}/{{ maxItems }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import {
  NUMBERED_LIST_MAX_ITEMS,
  NUMBERED_LIST_MAX_ITEM_LENGTH,
  applyPastedText,
  normalizeListItems
} from '~/utils/numberedList'

/**
 * Editor for a numbered list (string[]): one auto-resizing textarea per item, with add, remove and
 * move. Pasting a multi-line numbered list splits it into items and strips the "1." / "1)".
 */
const props = withDefaults(defineProps<{
  label?: string
  disabled?: boolean
  maxItems?: number
  maxLength?: number
  testId?: string
}>(), {
  label: '',
  disabled: false,
  maxItems: NUMBERED_LIST_MAX_ITEMS,
  maxLength: NUMBERED_LIST_MAX_ITEM_LENGTH,
  testId: undefined
})

const model = defineModel<string[]>({ default: () => [] })
const { t } = useI18n()

// Always at least one item to type into.
const items = computed(() => (model.value.length ? model.value : ['']))
const filledCount = computed(() => normalizeListItems(model.value).length)

const updateItem = (index: number, value: string) => {
  const next = [...items.value]
  next[index] = value ?? ''
  model.value = next
}

const add = () => {
  model.value = [...items.value, '']
}

const remove = (index: number) => {
  const next = items.value.filter((_, i) => i !== index)
  model.value = next.length ? next : ['']
}

const move = (index: number, direction: -1 | 1) => {
  const target = index + direction
  if (target < 0 || target >= items.value.length) return
  const next = [...items.value]
  const [item] = next.splice(index, 1)
  next.splice(target, 0, item!)
  model.value = next
}

const onPaste = (index: number, event: ClipboardEvent) => {
  if (props.disabled) return
  const text = event.clipboardData?.getData('text/plain') ?? ''
  const next = applyPastedText(items.value, index, text)
  if (!next) return
  event.preventDefault()
  model.value = next
}
</script>
