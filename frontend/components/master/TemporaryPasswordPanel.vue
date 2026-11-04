<template>
  <div class="space-y-3">
    <div class="rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 text-sm flex justify-between">
      <span class="text-[var(--text-muted)]">{{ t('masterData.employee.account.username') }}</span>
      <span class="font-medium text-[var(--text-main)]">{{ username }}</span>
    </div>
    <div>
      <p class="text-xs font-semibold text-[var(--text-muted)] mb-1">
        {{ t('masterData.employee.account.temporaryPassword') }}
      </p>
      <div class="flex items-center gap-2">
        <code class="flex-1 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 font-mono text-lg tracking-wider text-[var(--text-main)] select-all break-all">{{ temporaryPassword }}</code>
        <UButton
          :icon="copied ? 'i-heroicons-check' : 'i-heroicons-clipboard-document'"
          color="neutral"
          variant="subtle"
          size="lg"
          :aria-label="copied ? t('masterData.employee.account.copied') : t('masterData.employee.account.copy')"
          @click="copy"
        />
      </div>
    </div>
    <UAlert
      color="warning"
      variant="soft"
      icon="i-heroicons-exclamation-triangle"
      :description="t('masterData.employee.account.shownOnceNote')"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '~/composables/useI18n'
import { useAppToast } from '~/composables/useAppToast'

// One-time credentials panel shared by the employee create flow and the
// reset-password dialog. The password lives only in the parent's state.
const props = defineProps<{ username: string, temporaryPassword: string }>()

const { t } = useI18n()
const copied = ref(false)

const copy = async () => {
  try {
    await navigator.clipboard.writeText(props.temporaryPassword)
    copied.value = true
  } catch {
    useAppToast().error(t('errors.unknown'))
  }
}

// A different password means a different one-time value; reset the copied tick
watch(() => props.temporaryPassword, () => {
  copied.value = false
})
</script>
