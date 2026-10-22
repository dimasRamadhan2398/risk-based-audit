<template>
  <div>
    <div
      v-if="state === 'loading'"
      class="space-y-2.5 animate-pulse"
      role="status"
      aria-live="polite"
    >
      <div class="h-3 w-2/3 rounded bg-gray-200 dark:bg-gray-800" />
      <div class="h-3 w-full rounded bg-gray-200 dark:bg-gray-800" />
      <div class="h-3 w-5/6 rounded bg-gray-200 dark:bg-gray-800" />
    </div>

    <div
      v-else-if="state === 'error'"
      role="alert"
      class="flex flex-wrap items-center justify-between gap-3 p-3 rounded-xl bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900/40 text-xs text-red-700 dark:text-red-300"
    >
      <span>{{ errorText }}</span>
      <UButton
        v-if="retryLabel"
        size="xs"
        variant="outline"
        color="error"
        class="cursor-pointer"
        @click="emit('retry')"
      >
        {{ retryLabel }}
      </UButton>
    </div>

    <div
      v-else-if="state === 'empty'"
      class="py-6 text-center text-xs text-gray-500 dark:text-gray-400"
    >
      {{ emptyText }}
    </div>

    <slot v-else />
  </div>
</template>

<script setup lang="ts">
// Shared loading / error / empty wrapper for the auditor dashboard widgets.
// Each widget passes its own state so a failed source never blanks the whole page.
type AuditorWidgetStateValue = 'loading' | 'error' | 'empty' | 'ready'

defineProps<{
  state: AuditorWidgetStateValue
  emptyText?: string
  errorText?: string
  retryLabel?: string
}>()

const emit = defineEmits<{
  retry: []
}>()
</script>
