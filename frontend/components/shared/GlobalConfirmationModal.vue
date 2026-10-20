<template>
  <UModal
    :open="store.isOpen"
    @update:open="onOpenChange"
    :ui="{
      content: 'z-[10001] sm:max-w-lg w-[calc(100vw-2rem)] sm:w-full max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-4 sm:p-6 space-y-4 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'z-[10000] bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #header>
      <div class="flex items-start justify-between gap-3 w-full">
        <div class="flex items-start gap-3 min-w-0">
          <div v-if="store.options?.type === 'submit'" class="flex-shrink-0 w-8 h-8 rounded-full bg-primary-50 dark:bg-primary-900/20 flex items-center justify-center text-primary-500 font-black text-lg leading-none">
            ?
          </div>
          <div v-else class="flex-shrink-0 w-8 h-8 rounded-full bg-error-50 dark:bg-error-900/20 flex items-center justify-center text-error-500 font-black text-lg leading-none">
            !
          </div>
          <!-- Wraps instead of truncating: the header is the whole question, so it must stay readable. -->
          <h3 class="min-w-0 pt-0.5 text-base sm:text-lg font-bold leading-snug text-gray-900 dark:text-white break-words">
            {{ description }}
          </h3>
        </div>
        <UButton color="neutral" variant="ghost" icon="i-heroicons-x-mark-20-solid" class="-my-1 shrink-0" @click="store.resolve(false)" />
      </div>
    </template>

    <template #body>
      <div v-if="store.options?.body?.length">
        <p v-for="(line, index) in store.options.body" :key="index" class="text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">{{ line }}</p>
      </div>
      <div v-else-if="store.options?.type === 'submit'">
        <p class="text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">{{ t('globalConfirmation.submit.question') }}</p>
        <p class="text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">{{ t('globalConfirmation.submit.note') }}</p>
      </div>
      <div v-else>
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">{{ t('globalConfirmation.delete.question') }}</p>
        <p class="text-sm sm:text-md text-gray-600 dark:text-gray-400 mt-1 mb-1">{{ t('globalConfirmation.delete.note') }}</p>
      </div>
    </template>

    <template #footer>
      <div class="flex justify-end gap-3 w-full">
        <UButton :label="t('common.cancel')" color="neutral" variant="ghost" @click="store.resolve(false)" />
        <UButton v-if="store.options?.type === 'submit'" :label="store.options?.confirmLabel || t('globalConfirmation.submit.confirmButton')" color="primary" @click="store.resolve(true)" />
        <UButton v-else :label="store.options?.confirmLabel || t('globalConfirmation.delete.confirmButton')" color="error" @click="store.resolve(true)" />
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useGlobalModalStore } from '~/stores/global-modal'
import { useI18n } from '~/composables/useI18n'

const store = useGlobalModalStore()
const { t } = useI18n()

// Closing from outside the buttons (overlay click, Escape) counts as Cancel, so the caller is never left waiting.
const onOpenChange = (open: boolean) => {
  if (!open) store.resolve(false)
}

const title = computed(() => {
  if (store.options?.title) return store.options.title
  if (store.options?.type === 'submit') {
    if (store.options?.itemName) return t('globalConfirmation.submit.titleWithItem', { itemName: store.options.itemName })
    return t('globalConfirmation.submit.title')
  }
  if (store.options?.itemName) return t('globalConfirmation.delete.titleWithItem', { itemName: store.options.itemName })
  return t('globalConfirmation.delete.title')
})

const description = computed(() => {
  if (store.options?.description) return store.options.description
  if (store.options?.type === 'submit') return t('globalConfirmation.submit.question')
  return t('globalConfirmation.delete.question')
})
</script>
