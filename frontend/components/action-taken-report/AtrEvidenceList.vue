<template>
  <div>
    <ul
      v-if="evidence.length"
      class="divide-y divide-gray-100 dark:divide-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
    >
      <li
        v-for="file in evidence"
        :key="file.id"
        class="flex items-center justify-between gap-3 px-3 py-2"
      >
        <div class="flex items-center gap-2 min-w-0">
          <UIcon
            name="i-lucide-file-text"
            class="w-4 h-4 shrink-0 text-gray-400"
          />
          <div class="min-w-0">
            <p
              class="text-sm font-medium truncate"
              :title="file.file_name"
            >
              {{ file.file_name || t('actionTakenReport.evidence.unnamed') }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('actionTakenReport.evidence.uploadedMeta', { date: formatAtrDateTime(file.uploaded_at, locale), by: file.uploaded_by || '-' }) }}
            </p>
          </div>
        </div>
        <UTooltip :text="t('actionTakenReport.evidence.download')">
          <UButton
            icon="i-lucide-download"
            color="neutral"
            variant="ghost"
            size="sm"
            :loading="store.downloadingEvidenceId === file.id"
            :aria-label="t('actionTakenReport.evidence.downloadNamed', { name: file.file_name })"
            @click="store.downloadEvidence(atrId, file)"
          />
        </UTooltip>
      </li>
    </ul>
    <p
      v-else
      class="text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('actionTakenReport.evidence.empty') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { AtrEvidence } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { formatAtrDateTime } from '~/utils/actionTakenReport'

defineProps<{ atrId: string, evidence: AtrEvidence[] }>()

const { t, locale } = useI18n()
const store = useActionTakenReportStore()
</script>
