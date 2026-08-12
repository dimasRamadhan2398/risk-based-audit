<template>
  <component
    :is="to ? NuxtLink : 'button'"
    :to="to"
    :type="to ? undefined : type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center font-bold transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-error-500/50 select-none disabled:cursor-not-allowed disabled:opacity-50"
    :class="[
      sizeClasses[size],
      variantClasses[variant],
      block ? 'w-full' : '',
      roundedClass,
    ]"
    v-bind="$attrs"
  >
    <!-- Loading spinner icon -->
    <UIcon
      v-if="loading"
      name="i-heroicons-arrow-path"
      class="animate-spin shrink-0"
      :class="iconSizeClasses[size]"
    />

    <!-- Left icon -->
    <UIcon
      v-else-if="icon"
      :name="icon"
      class="shrink-0"
      :class="iconSizeClasses[size]"
    />

    <!-- Button label or slot -->
    <span v-if="label || $slots.default" class="whitespace-nowrap">
      <slot>{{ label }}</slot>
    </span>

    <!-- Right icon -->
    <UIcon
      v-if="iconRight && !loading"
      :name="iconRight"
      class="shrink-0"
      :class="iconSizeClasses[size]"
    />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NuxtLink } from '#components'

interface Props {
  label?: string
  icon?: string
  iconRight?: string
  variant?: 'solid' | 'soft' | 'outline' | 'ghost'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  disabled?: boolean
  loading?: boolean
  block?: boolean
  rounded?: 'normal' | 'full' | 'xl' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'soft',
  size: 'md',
  disabled: false,
  loading: false,
  block: false,
  rounded: 'xl',
  type: 'button',
})

const sizeClasses = {
  xs: 'text-xs px-2.5 py-1 gap-1',
  sm: 'text-xs px-3 py-1.5 gap-1.5',
  md: 'text-sm px-4 py-2 gap-2',
  lg: 'text-sm px-5 py-3 gap-2',
  xl: 'text-base px-6 py-3.5 gap-2.5',
}

const iconSizeClasses = {
  xs: 'w-3.5 h-3.5',
  sm: 'w-4 h-4',
  md: 'w-4 h-4',
  lg: 'w-5 h-5',
  xl: 'w-5 h-5',
}

const roundedClass = computed(() => {
  switch (props.rounded) {
    case 'full': return 'rounded-full'
    case 'lg': return 'rounded-lg'
    case 'normal': return 'rounded-md'
    case 'xl':
    default: return 'rounded-xl'
  }
})

const variantClasses = {
  solid: 'bg-error-500 hover:bg-error-600 active:bg-error-700 text-white shadow-md shadow-error-500/20 border border-transparent',
  soft: 'bg-error-500/10 hover:bg-error-500/20 active:bg-error-500/30 text-error-600 dark:text-error-400 border border-error-500/30',
  outline: 'border border-error-500 text-error-600 dark:text-error-400 hover:bg-error-500/10 active:bg-error-500/20 bg-transparent',
  ghost: 'text-error-600 dark:text-error-400 hover:bg-error-500/10 active:bg-error-500/20 border border-transparent',
}
</script>
