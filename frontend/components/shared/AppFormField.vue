<template>
  <div
    :class="[
      'w-full transition-all duration-150',
      orientation === 'horizontal' ? 'flex flex-col sm:flex-row sm:items-start sm:gap-4' : 'flex flex-col gap-1.5',
      uiClasses.root,
      $attrs.class
    ]"
  >
    <!-- Header: Label, Required/Optional Badges, Tooltip, Actions/Counter, Hint -->
    <div
      v-if="hasHeader"
      :class="[
        'flex items-center justify-between gap-2 min-h-[1.5rem]',
        orientation === 'horizontal' ? 'sm:w-1/3 sm:pt-1.5 shrink-0' : 'w-full',
        uiClasses.labelWrapper
      ]"
    >
      <!-- Left side: Label + Badges + Info Tooltip -->
      <div class="flex items-center flex-wrap gap-1.5 min-w-0">
        <!-- Label / Title -->
        <label
          v-if="effectiveTitle || $slots.label || $slots.title"
          :for="fieldId"
          :title="isTitleTruncated ? rawTitle : undefined"
          :class="[
            'font-medium select-none flex items-center gap-1 truncate text-gray-800 dark:text-gray-200 transition-colors',
            sizeClasses.label,
            isTitleTruncated ? 'cursor-help' : '',
            uiClasses.label
          ]"
        >
          <slot
            name="label"
            :label="displayTitle"
            :raw-label="rawTitle"
            :is-truncated="isTitleTruncated"
            :is-over-limit="isTitleOverLimit"
          >
            <slot
              name="title"
              :title="displayTitle"
              :raw-title="rawTitle"
              :is-truncated="isTitleTruncated"
              :is-over-limit="isTitleOverLimit"
            >
              <span>{{ displayTitle }}</span>
            </slot>
          </slot>
        </label>

        <!-- Required Asterisk / Badge -->
        <span
          v-if="required"
          class="text-red-500 font-bold text-sm leading-none shrink-0 select-none"
          title="Field is required"
          aria-hidden="true"
        >
          *
        </span>

        <!-- Optional Badge -->
        <span
          v-else-if="optional"
          class="text-[10px] font-normal tracking-wide px-1.5 py-0.5 rounded text-gray-400 dark:text-gray-500 bg-gray-100 dark:bg-gray-800/80 shrink-0 select-none"
        >
          {{ typeof optional === 'string' ? optional : 'Optional' }}
        </span>

        <!-- Info Tooltip with highest z-index & fixed popper strategy -->
        <div
          v-if="tooltip || info || $slots.tooltip || $slots.info"
          class="relative inline-flex items-center shrink-0 z-[100]"
        >
          <slot name="tooltip">
            <slot name="info">
              <UTooltip
                v-if="tooltip || info"
                :text="tooltip || info"
                :popper="{ placement: 'top', strategy: 'fixed' }"
                :ui="{ base: 'z-[99999] shadow-2xl text-xs max-w-xs font-normal' }"
              >
                <button
                  type="button"
                  tabindex="-1"
                  class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors focus:outline-none flex items-center p-0.5 rounded"
                  :title="tooltip || info"
                  aria-label="More information"
                >
                  <UIcon
                    name="i-lucide-info"
                    class="w-3.5 h-3.5"
                  />
                </button>
              </UTooltip>
            </slot>
          </slot>
        </div>

        <!-- Title Overlimit Warning Badge -->
        <span
          v-if="showTitleWarning && isTitleOverLimit"
          class="text-[10px] font-medium px-1.5 py-0.5 rounded bg-amber-50 text-amber-600 dark:bg-amber-950/40 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50 shrink-0"
          :title="`Title is ${rawTitle.length} characters (maximum is ${maxTitleLength})`"
        >
          {{ rawTitle.length }}/{{ maxTitleLength }} max
        </span>
      </div>

      <!-- Right side: Hint, Actions Slot, Input Character Counter -->
      <div class="flex items-center gap-2 shrink-0">
        <!-- Custom Actions Slot -->
        <div
          v-if="$slots.actions || $slots.extra"
          class="flex items-center"
        >
          <slot name="actions">
            <slot name="extra" />
          </slot>
        </div>

        <!-- Input Character Counter (e.g., 45 / 100) -->
        <div
          v-if="showCounter"
          :class="[
            'text-xs tabular-nums font-mono select-none transition-colors',
            counterColorClass,
            uiClasses.counter
          ]"
          :title="`Characters: ${currentCount} of ${maxCount}`"
        >
          <span class="font-medium">{{ currentCount }}</span>
          <span class="opacity-60">/{{ maxCount }}</span>
        </div>

        <!-- Hint Text or Slot -->
        <span
          v-if="hint || $slots.hint"
          :class="['text-xs text-gray-400 dark:text-gray-500 select-none', uiClasses.hint]"
        >
          <slot
            name="hint"
            :hint="hint"
          >
            {{ hint }}
          </slot>
        </span>
      </div>
    </div>

    <!-- Description (Optional, rendered below header / above input) -->
    <p
      v-if="description || $slots.description"
      :id="descriptionId"
      :class="[
        'text-xs text-gray-500 dark:text-gray-400 leading-relaxed',
        orientation === 'horizontal' ? 'sm:pl-1' : '',
        uiClasses.description
      ]"
    >
      <slot
        name="description"
        :description="description"
      >
        {{ description }}
      </slot>
    </p>

    <!-- Main Input Container Slot -->
    <div
      :class="[
        'relative w-full',
        orientation === 'horizontal' ? 'sm:flex-1' : '',
        uiClasses.container
      ]"
    >
      <slot
        :id="fieldId"
        :field-name="name"
        :error="error"
        :is-invalid="hasError"
        :aria-describedby="ariaDescribedBy"
      />

      <!-- Error Message -->
      <transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="transform -translate-y-1 opacity-0"
        enter-to-class="transform translate-y-0 opacity-100"
        leave-active-class="transition duration-100 ease-in"
        leave-from-class="transform translate-y-0 opacity-100"
        leave-to-class="transform -translate-y-1 opacity-0"
      >
        <div
          v-if="hasError"
          :id="errorId"
          role="alert"
          :class="[
            'flex items-center gap-1.5 mt-1.5 text-xs text-red-500 dark:text-red-400 font-medium',
            uiClasses.error
          ]"
        >
          <UIcon
            name="i-lucide-alert-circle"
            class="w-3.5 h-3.5 shrink-0"
          />
          <slot
            name="error"
            :error="errorMessage"
          >
            <span>{{ errorMessage }}</span>
          </slot>
        </div>
      </transition>

      <!-- Help Text (Only shown if there is no active error) -->
      <p
        v-if="!hasError && (help || $slots.help)"
        :id="helpId"
        :class="[
          'mt-1.5 text-xs text-gray-400 dark:text-gray-500 leading-normal',
          uiClasses.help
        ]"
      >
        <slot
          name="help"
          :help="help"
        >
          {{ help }}
        </slot>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, useId } from 'vue'

export interface AppFormFieldProps {
  /**
   * Field label / title text
   */
  label?: string
  /**
   * Alias for label
   */
  title?: string
  /**
   * Name of the form field for form integration
   */
  name?: string
  /**
   * Explicit element ID. Auto-generated if omitted.
   */
  id?: string
  /**
   * Maximum allowed characters for the title/label (Default: 100)
   */
  maxTitleLength?: number
  /**
   * Whether to truncate the title when it exceeds maxTitleLength (Default: true)
   */
  truncateTitle?: boolean
  /**
   * Whether to display a warning badge when title exceeds maxTitleLength
   */
  showTitleWarning?: boolean
  /**
   * Mark field as required with a red asterisk
   */
  required?: boolean
  /**
   * Mark field as optional (boolean or custom label string e.g. "Optional")
   */
  optional?: boolean | string
  /**
   * Informational tooltip text
   */
  tooltip?: string
  /**
   * Alias for tooltip
   */
  info?: string
  /**
   * Top-right hint text
   */
  hint?: string
  /**
   * Sub-label description text
   */
  description?: string
  /**
   * Helper text below input
   */
  help?: string
  /**
   * Error message (string) or boolean indicating invalid state
   */
  error?: string | boolean | null
  /**
   * Component sizing
   */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  /**
   * Layout orientation
   */
  orientation?: 'vertical' | 'horizontal'
  /**
   * Show character counter (e.g., 42/100)
   */
  counter?: boolean
  /**
   * Alias for counter
   */
  showCount?: boolean
  /**
   * Current character count or input value string to count characters from
   */
  modelValue?: string | number | null
  /**
   * Explicit character count number
   */
  count?: number
  /**
   * Maximum characters for the input counter (Default: 100)
   */
  maxCount?: number
  /**
   * Fine-grained CSS class overrides
   */
  ui?: {
    root?: string
    labelWrapper?: string
    label?: string
    container?: string
    description?: string
    error?: string
    help?: string
    hint?: string
    counter?: string
  }
}

const props = withDefaults(defineProps<AppFormFieldProps>(), {
  label: '',
  title: '',
  name: undefined,
  id: undefined,
  maxTitleLength: 100,
  truncateTitle: true,
  showTitleWarning: true,
  required: false,
  optional: false,
  tooltip: undefined,
  info: undefined,
  hint: undefined,
  description: undefined,
  help: undefined,
  error: undefined,
  size: 'md',
  orientation: 'vertical',
  counter: false,
  showCount: false,
  modelValue: undefined,
  count: undefined,
  maxCount: 100,
  ui: () => ({})
})

const slots = defineSlots<{
  default?: (props: { id: string, name?: string, error?: string | boolean | null, isInvalid: boolean, ariaDescribedBy?: string }) => unknown
  label?: (props: { label: string, rawLabel: string, isTruncated: boolean, isOverLimit: boolean }) => unknown
  title?: (props: { title: string, rawTitle: string, isTruncated: boolean, isOverLimit: boolean }) => unknown
  hint?: (props: { hint?: string }) => unknown
  tooltip?: () => unknown
  info?: () => unknown
  actions?: () => unknown
  extra?: () => unknown
  description?: (props: { description?: string }) => unknown
  help?: (props: { help?: string }) => unknown
  error?: (props: { error?: string }) => unknown
}>()

// Generate consistent unique ID if not provided
const autoId = useId()
const fieldId = computed(() => props.id || props.name || `app-form-field-${autoId}`)
const errorId = computed(() => `${fieldId.value}-error`)
const descriptionId = computed(() => `${fieldId.value}-description`)
const helpId = computed(() => `${fieldId.value}-help`)

// Title & Maximum 100 Character Management
const rawTitle = computed(() => props.title || props.label || '')
const effectiveTitle = computed(() => rawTitle.value.trim())

const isTitleOverLimit = computed(() => {
  return rawTitle.value.length > props.maxTitleLength
})

const isTitleTruncated = computed(() => {
  return props.truncateTitle && isTitleOverLimit.value
})

const displayTitle = computed(() => {
  if (!rawTitle.value) return ''
  if (isTitleTruncated.value) {
    // Truncate cleanly to maxTitleLength
    return rawTitle.value.slice(0, props.maxTitleLength) + '...'
  }
  return rawTitle.value
})

// Header presence check
const hasHeader = computed(() => {
  return !!(
    effectiveTitle.value
    || props.required
    || props.optional
    || props.tooltip
    || props.info
    || props.hint
    || showCounter.value
    || slots.label
    || slots.title
    || slots.hint
    || slots.tooltip
    || slots.info
    || slots.actions
    || slots.extra
  )
})

// Counter computation
const showCounter = computed(() => props.counter || props.showCount)

const currentCount = computed(() => {
  if (props.count !== undefined) return props.count
  if (typeof props.modelValue === 'string') return props.modelValue.length
  if (typeof props.modelValue === 'number') return String(props.modelValue).length
  return 0
})

const counterColorClass = computed(() => {
  const ratio = currentCount.value / (props.maxCount || 100)
  if (ratio >= 1) return 'text-red-500 font-semibold'
  if (ratio >= 0.9) return 'text-amber-500 dark:text-amber-400 font-medium'
  return 'text-gray-400 dark:text-gray-500'
})

// Error computation
const hasError = computed(() => {
  if (typeof props.error === 'boolean') return props.error
  return typeof props.error === 'string' && props.error.trim().length > 0
})

const errorMessage = computed(() => {
  if (typeof props.error === 'string') return props.error
  return ''
})

// Accessibility describedby
const ariaDescribedBy = computed(() => {
  const ids: string[] = []
  if (hasError.value) ids.push(errorId.value)
  if (props.description || slots.description) ids.push(descriptionId.value)
  if (!hasError.value && (props.help || slots.help)) ids.push(helpId.value)
  return ids.length > 0 ? ids.join(' ') : undefined
})

// Size mappings
const sizeClasses = computed(() => {
  switch (props.size) {
    case 'xs':
      return { label: 'text-xs' }
    case 'sm':
      return { label: 'text-xs' }
    case 'lg':
      return { label: 'text-base' }
    case 'xl':
      return { label: 'text-lg' }
    case 'md':
    default:
      return { label: 'text-sm' }
  }
})

// UI classes safe merging
const uiClasses = computed(() => props.ui || {})
</script>
