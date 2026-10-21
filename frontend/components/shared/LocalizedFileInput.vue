<template>
  <div class="flex items-center gap-3 w-full min-w-0">
    <!-- The native control renders "Choose file" / "No file chosen" in the browser's language,
         so it stays visually hidden and a translated button + label are shown instead. -->
    <input
      ref="inputRef"
      type="file"
      class="sr-only"
      tabindex="-1"
      :accept="accept"
      :required="required"
      @change="onChange"
    />
    <UButton
      type="button"
      color="primary"
      variant="soft"
      size="md"
      icon="i-lucide-paperclip"
      class="shrink-0"
      :label="t('common.fileInput.chooseFile')"
      @click="inputRef?.click()"
    />
    <span class="text-sm text-gray-500 dark:text-gray-400 truncate" :title="pickedName || undefined">
      {{ pickedName || t('common.fileInput.noFileChosen') }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { useI18n } from '~/composables/useI18n'

defineProps<{
  accept?: string
  required?: boolean
}>()

const emit = defineEmits<{
  (e: 'change', event: Event): void
}>()

const { t } = useI18n()
const inputRef = ref<HTMLInputElement | null>(null)
const pickedName = ref('')

const onChange = (event: Event) => {
  emit('change', event)
  // The parent may reject the file and clear the input, so read the name after its handler ran.
  nextTick(() => {
    pickedName.value = inputRef.value?.files?.[0]?.name || ''
  })
}
</script>
