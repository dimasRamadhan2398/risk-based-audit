<template>
  <!-- Reusable Toast Notification trigger / provider component if needed in template scope -->
  <slot :show-success="showSuccess" :show-error="showError" :show-warning="showWarning" :show-info="showInfo" />
</template>

<script lang="ts">
import { useAppToast } from '~/composables/useAppToast'
import { useI18n } from '~/composables/useI18n'

export function useToastNotification() {
  const appToast = useAppToast()
  const { t } = useI18n()
  return {
    showSuccess: (title?: string, description?: string) => appToast.success(title || t('common.success'), description),
    showError: (title?: string, description?: string) => appToast.error(title || t('common.error'), description),
    showWarning: (title?: string, description?: string) => appToast.warning(title || t('common.warning'), description),
    showInfo: (title?: string, description?: string) => appToast.info(title || t('common.info'), description),
    ...appToast
  }
}
</script>

<script setup lang="ts">
const appToast = useAppToast()
const { t } = useI18n()

const props = defineProps<{
  title?: string
  description?: string
  type?: 'success' | 'error' | 'warning' | 'info'
}>()

const showSuccess = (title?: string, description?: string) => {
  appToast.success(title || props.title || t('common.success'), description || props.description)
}

const showError = (title?: string, description?: string) => {
  appToast.error(title || props.title || t('common.error'), description || props.description)
}

const showWarning = (title?: string, description?: string) => {
  appToast.warning(title || props.title || t('common.warning'), description || props.description)
}

const showInfo = (title?: string, description?: string) => {
  appToast.info(title || props.title || t('common.info'), description || props.description)
}

defineExpose({
  showSuccess,
  showError,
  showWarning,
  showInfo,
})
</script>
