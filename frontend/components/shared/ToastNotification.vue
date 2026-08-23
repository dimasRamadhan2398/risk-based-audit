<template>
  <!-- Reusable Toast Notification trigger / provider component if needed in template scope -->
  <slot :show-success="showSuccess" :show-error="showError" :show-warning="showWarning" :show-info="showInfo" />
</template>

<script setup lang="ts">
import { useAppToast } from '~/composables/useAppToast'

const appToast = useAppToast()

const props = defineProps<{
  title?: string
  description?: string
  type?: 'success' | 'error' | 'warning' | 'info'
}>()

const showSuccess = (title?: string, description?: string) => {
  appToast.success(title || props.title || 'Berhasil', description || props.description)
}

const showError = (title?: string, description?: string) => {
  appToast.error(title || props.title || 'Gagal', description || props.description)
}

const showWarning = (title?: string, description?: string) => {
  appToast.warning(title || props.title || 'Peringatan', description || props.description)
}

const showInfo = (title?: string, description?: string) => {
  appToast.info(title || props.title || 'Informasi', description || props.description)
}

defineExpose({
  showSuccess,
  showError,
  showWarning,
  showInfo,
})
</script>
