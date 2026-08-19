<template>
  <component
    :is="to ? NuxtLink : 'button'"
    :to="to"
    :type="to ? undefined : type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center font-bold transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-primary/20 select-none disabled:cursor-not-allowed disabled:opacity-50"
    :class="[
      sizeClasses[size],
      computedVariantClass,
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
  variant?: 'fill' | 'solid' | 'muted' | 'soft' | 'outline' | 'ghost'
  color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success' | 'neutral'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  disabled?: boolean
  loading?: boolean
  block?: boolean
  rounded?: 'normal' | 'full' | 'xl' | 'lg'
  type?: 'button' | 'submit' | 'reset'
  to?: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'fill',
  color: 'primary',
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

const computedVariantClass = computed(() => {
  const c = props.color || 'primary'
  const v = props.variant || 'fill'

  if (c === 'error') {
    if (v === 'solid' || v === 'fill') return 'bg-red-600 hover:bg-red-700 active:bg-red-800 text-white shadow-md border border-transparent'
    if (v === 'muted') return 'bg-red-100 dark:bg-red-950/40 text-red-700 dark:text-red-300 border border-red-200 dark:border-red-900/50'
    if (v === 'soft') return 'bg-red-500/20 hover:bg-red-500/30 active:bg-red-500/40 text-red-700 dark:text-red-300 border border-red-500/30'
    if (v === 'outline') return 'border border-red-500 text-red-600 dark:text-red-400 hover:bg-red-500/10 active:bg-red-500/20 bg-transparent'
    return 'text-red-600 dark:text-red-400 hover:bg-red-500/10 active:bg-red-500/20 border border-transparent'
  }

  if (c === 'secondary') {
    if (v === 'solid' || v === 'fill') return 'bg-secondary hover:opacity-90 active:opacity-95 text-white shadow-md border border-transparent'
    if (v === 'muted') return 'bg-secondary/20 text-secondary border border-secondary/30'
    if (v === 'soft') return 'bg-secondary/20 hover:bg-secondary/30 active:bg-secondary/40 text-secondary border border-secondary/30'
    if (v === 'outline') return 'border border-secondary text-secondary hover:bg-secondary/10 active:bg-secondary/20 bg-transparent'
    return 'text-secondary hover:bg-secondary/10 active:bg-secondary/20 border border-transparent'
  }

  if (c === 'warning') {
    if (v === 'solid' || v === 'fill') return 'bg-amber-500 hover:bg-amber-600 active:bg-amber-700 text-white shadow-md border border-transparent'
    if (v === 'muted') return 'bg-amber-100 dark:bg-amber-950/40 text-amber-800 dark:text-amber-300 border border-amber-200 dark:border-amber-900/50'
    if (v === 'soft') return 'bg-amber-500/20 hover:bg-amber-500/30 active:bg-amber-500/40 text-amber-700 dark:text-amber-300 border border-amber-500/30'
    if (v === 'outline') return 'border border-amber-500 text-amber-600 dark:text-amber-400 hover:bg-amber-500/10 active:bg-amber-500/20 bg-transparent'
    return 'text-amber-600 dark:text-amber-400 hover:bg-amber-500/10 active:bg-amber-500/20 border border-transparent'
  }

  // Default: Primary
  if (v === 'solid' || v === 'fill') return 'bg-primary hover:opacity-90 active:opacity-95 text-white shadow-md shadow-primary/20 border border-transparent'
  if (v === 'muted') return 'bg-primary/15 hover:bg-primary/25 text-primary border border-primary/20 font-bold'
  if (v === 'soft') return 'bg-primary/20 hover:bg-primary/30 active:bg-primary/40 text-primary border border-primary/30'
  if (v === 'outline') return 'border border-primary text-primary hover:bg-primary/10 active:bg-primary/20 bg-transparent'
  return 'text-primary hover:bg-primary/10 active:bg-primary/20 border border-transparent'
})
</script>
