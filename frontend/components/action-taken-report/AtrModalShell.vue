<template>
  <!-- Same structure as the Risk Control Matrix modal: fixed header and footer, only the body scrolls -->
  <UModal
    v-model:open="open"
    :title="title"
    :description="subtitle || title"
    :dismissible="!busy"
    :ui="{
      content: `w-[calc(100vw-2rem)] sm:w-full ${widthClass} max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden`,
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <div class="flex flex-col max-h-[85vh] overflow-hidden bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100">
        <div class="shrink-0 flex items-start justify-between gap-4 border-b border-slate-100 dark:border-slate-800 px-4 sm:px-6 py-4">
          <div class="min-w-0 space-y-1">
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ title }}
            </h3>
            <p
              v-if="subtitle"
              class="text-sm text-slate-500 dark:text-slate-400 break-words"
            >
              {{ subtitle }}
            </p>
            <div
              v-if="$slots.meta"
              class="flex flex-wrap items-center gap-2 pt-1"
            >
              <slot name="meta" />
            </div>
          </div>
          <UButton
            type="button"
            icon="i-lucide-x"
            color="neutral"
            variant="ghost"
            :disabled="busy"
            :aria-label="t('actionTakenReport.modal.close')"
            @click="open = false"
          />
        </div>

        <div class="flex-1 min-h-0 overflow-y-auto px-4 sm:px-6 py-5 space-y-5">
          <slot />
        </div>

        <div
          v-if="$slots.footer"
          class="shrink-0 border-t border-slate-100 dark:border-slate-800 px-4 sm:px-6 py-4"
        >
          <slot name="footer" />
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'

const props = withDefaults(defineProps<{
  title: string
  subtitle?: string
  size?: 'md' | 'lg'
  /** While a request runs the modal cannot be dismissed. */
  busy?: boolean
}>(), {
  subtitle: '',
  size: 'md',
  busy: false
})

const open = defineModel<boolean>('open', { default: false })
const { t } = useI18n()

const widthClass = computed(() => (props.size === 'lg' ? 'sm:max-w-4xl' : 'sm:max-w-2xl'))
</script>
