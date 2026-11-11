<template>
  <div
    class="rounded-xl border border-info-200 dark:border-info-900 bg-info-50/40 dark:bg-info-950/20 p-5 space-y-4"
    data-testid="es-current-follow-up"
  >
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h3 class="text-md font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon
            name="i-lucide-activity"
            class="size-4 text-info-500"
          />
          {{ t('executiveSummaryForm.followUp.current.title') }}
        </h3>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {{ t('executiveSummaryForm.followUp.current.description') }}
        </p>
        <p
          v-if="asOfLabel && !followUp.loading.value && !followUp.error.value"
          class="text-xs text-gray-400 mt-0.5"
        >
          {{ t('executiveSummaryForm.followUp.dataAsOf', { time: asOfLabel }) }}
        </p>
      </div>
      <UButton
        color="info"
        variant="outline"
        size="xs"
        icon="i-lucide-refresh-cw"
        :label="t('executiveSummaryForm.followUp.refresh')"
        :loading="followUp.loading.value"
        :disabled="!query"
        @click="refresh"
      />
    </div>

    <div
      v-if="followUp.loading.value && !result"
      class="flex items-center gap-2 text-sm text-gray-400"
    >
      <UIcon
        name="i-lucide-loader-circle"
        class="size-4 animate-spin"
      />
      {{ t('executiveSummaryForm.followUp.loading') }}
    </div>
    <div
      v-else-if="followUp.error.value"
      class="flex flex-wrap items-center gap-3 text-sm text-error-500"
    >
      <span>{{ t('executiveSummaryForm.followUp.error', { message: followUp.error.value }) }}</span>
      <UButton
        color="error"
        variant="outline"
        size="xs"
        icon="i-lucide-refresh-cw"
        :label="t('executiveSummaryForm.followUp.retry')"
        @click="refresh"
      />
    </div>

    <template v-else-if="result">
      <!-- Bucket counts -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div
          v-for="row in bucketRows"
          :key="row.status"
          class="rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 py-3"
        >
          <FollowUpStatusBadge :status="row.status" />
          <div class="mt-1.5 text-2xl font-bold tabular-nums text-gray-900 dark:text-white">
            {{ row.jumlah }}
          </div>
          <div class="text-xs text-gray-400 tabular-nums">
            {{ row.persentase.toFixed(1) }}%
          </div>
        </div>
      </div>

      <!-- Per-finding detail (hidden when the backend restricts it) -->
      <p
        v-if="result.findingsRestricted"
        class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1.5"
      >
        <UIcon
          name="i-lucide-eye-off"
          class="size-3.5"
        />
        {{ t('executiveSummaryForm.followUp.current.restricted') }}
      </p>
      <p
        v-else-if="result.findings.length === 0"
        class="text-sm text-gray-400"
      >
        {{ t('executiveSummaryForm.followUp.empty') }}
      </p>
      <div
        v-else
        class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900"
      >
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-xs">
          <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-600 dark:text-gray-300 font-bold uppercase tracking-wider">
            <tr>
              <th class="px-3 py-2 text-left">
                {{ t('executiveSummaryForm.followUp.current.columns.finding') }}
              </th>
              <th
                v-if="showReportColumn"
                class="px-3 py-2 text-left"
              >
                {{ t('executiveSummaryForm.followUp.current.columns.report') }}
              </th>
              <th class="px-3 py-2 text-left">
                {{ t('executiveSummaryForm.followUp.current.columns.pic') }}
              </th>
              <th class="px-3 py-2 text-left">
                {{ t('executiveSummaryForm.followUp.current.columns.dueDate') }}
              </th>
              <th class="px-3 py-2 text-right">
                {{ t('executiveSummaryForm.followUp.current.columns.progress') }}
              </th>
              <th class="px-3 py-2 text-left">
                {{ t('executiveSummaryForm.followUp.current.columns.status') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr
              v-for="f in result.findings"
              :key="f.findingId"
            >
              <td
                class="px-3 py-2 text-gray-800 dark:text-gray-200 max-w-xs truncate"
                :title="f.title"
              >
                {{ f.title || '-' }}
              </td>
              <td
                v-if="showReportColumn"
                class="px-3 py-2 font-mono text-gray-600 dark:text-gray-400 whitespace-nowrap"
              >
                {{ f.reportNumber || '-' }}
              </td>
              <td class="px-3 py-2 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                {{ f.picName || '-' }}
              </td>
              <td class="px-3 py-2 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                {{ formatFollowUpDate(f.dueDate, locale) || '-' }}
                <span
                  v-if="f.bucket === 'OVERDUE' && f.overdueDays > 0"
                  class="block text-error-500"
                >
                  {{ t('executiveSummaryForm.followUp.overdueDays', { days: f.overdueDays }) }}
                </span>
              </td>
              <td class="px-3 py-2 text-right tabular-nums text-gray-600 dark:text-gray-400">
                {{ f.progress === null ? '-' : `${f.progress}%` }}
              </td>
              <td class="px-3 py-2">
                <FollowUpStatusBadge
                  :status="BUCKET_STATUS[f.bucket]"
                  :reason="f.noActionReason"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useExecutiveSummaryFollowUp } from '~/composables/useExecutiveSummaryFollowUp'
import { followUpRowsFromStats, formatSyncedAt } from '~/utils/executiveSummaryCompilation'
import {
  BUCKET_STATUS,
  formatFollowUpDate,
  type FollowUpStatusQuery
} from '~/utils/executiveSummaryFollowUp'
import FollowUpStatusBadge from './FollowUpStatusBadge.vue'

// Read-only live ATR status shown next to the frozen figures of an approved summary.
// The parent renders it only for users who may view every ATR; the backend enforces the same.
const props = defineProps<{
  query: FollowUpStatusQuery | null
  /** Compilation summaries span several LHA, so each finding shows its LHA number */
  showReportColumn?: boolean
}>()

const { t, locale } = useI18n()
const followUp = useExecutiveSummaryFollowUp()

const result = computed(() => followUp.data.value)
const bucketRows = computed(() => (result.value ? followUpRowsFromStats(result.value.buckets) : []))
const asOfLabel = computed(() => formatSyncedAt(result.value?.asOf, locale.value))

const refresh = () => {
  if (props.query) followUp.load(props.query)
}

watch(() => JSON.stringify(props.query), () => {
  followUp.reset()
  refresh()
}, { immediate: true })
</script>
