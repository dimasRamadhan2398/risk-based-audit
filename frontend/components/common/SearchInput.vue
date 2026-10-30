<template>
  <UInput
    v-model="model"
    v-bind="$attrs"
    icon="i-lucide-search"
  >
    <template
      v-if="model"
      #trailing
    >
      <UButton
        color="neutral"
        variant="link"
        size="sm"
        icon="i-lucide-x"
        :aria-label="t('common.clearSearch')"
        @click="clear"
      />
    </template>
  </UInput>
</template>

<script setup lang="ts">
defineOptions({ inheritAttrs: false })

const emit = defineEmits<{ clear: [] }>()
const model = defineModel<string>({ default: '' })
const { t } = useI18n()

// Clearing goes through v-model, so any watcher/debounce/computed that reacts to typing runs too.
// `clear` lets callers with an explicit Search button refresh results.
const clear = () => {
  model.value = ''
  emit('clear')
}
</script>
