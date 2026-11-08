<template>
  <div class="space-y-3">
    <!-- Source line: frozen snapshot of an approved summary, or the live ATR data of a draft -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="space-y-0.5">
        <p
          v-if="approvedDateLabel"
          class="text-sm font-semibold text-gray-700 dark:text-gray-200 flex items-center gap-1.5"
        >
          <UIcon
            name="i-lucide-lock"
            class="size-4 text-success-500"
          />
          {{ t('executiveSummaryForm.followUp.statusAsOf', { date: approvedDateLabel }) }}
        </p>
        <p
          v-if="approvedDateLabel"
          class="text-xs text-gray-400"
        >
          {{ t('executiveSummaryForm.followUp.frozenHint') }}
        </p>
        <template v-else>
          <p class="text-xs text-gray-400">
            {{ t('executiveSummaryForm.followUp.readOnlyHint') }}
          </p>
          <p
            v-if="asOfLabel && !loading && !error"
            class="text-xs text-gray-400"
          >
            {{ t('executiveSummaryForm.followUp.dataAsOf', { time: asOfLabel }) }}
          </p>
        </template>
      </div>
      <UButton
        v-if="refreshable"
        color="primary"
        variant="outline"
        size="sm"
        icon="i-lucide-refresh-cw"
        :label="t('executiveSummaryForm.followUp.refresh')"
        :loading="loading"
        @click="emit('refresh')"
      />
    </div>

    <div
      v-if="loading"
      class="flex items-center gap-2 text-sm text-gray-400"
    >
      <UIcon
        name="i-lucide-loader-circle"
        class="size-4 animate-spin"
      />
      {{ t('executiveSummaryForm.followUp.loading') }}
    </div>
    <div
      v-else-if="error"
      class="flex flex-wrap items-center gap-3 text-sm text-error-500"
    >
      <span>{{ t('executiveSummaryForm.followUp.error', { message: error }) }}</span>
      <UButton
        v-if="refreshable"
        color="error"
        variant="outline"
        size="xs"
        icon="i-lucide-refresh-cw"
        :label="t('executiveSummaryForm.followUp.retry')"
        @click="emit('refresh')"
      />
    </div>
    <p
      v-else-if="emptyHint"
      class="text-sm text-gray-400"
    >
      {{ emptyHint }}
    </p>

    <div
      class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm"
      :class="loading ? 'opacity-50 pointer-events-none' : ''"
    >
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
        <thead class="bg-gray-50 dark:bg-gray-800/80">
          <tr>
            <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">
              {{ t('executiveSummaryForm.table.status') }}
            </th>
            <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">
              {{ t('executiveSummaryForm.table.count') }}
            </th>
            <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider w-36">
              {{ t('executiveSummaryForm.table.percent') }}
            </th>
            <th class="px-6 py-3.5 text-left text-md font-bold text-gray-500 dark:text-white uppercase tracking-wider">
              {{ t('executiveSummaryForm.table.notes') }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800 font-medium">
          <tr
            v-for="row in rows"
            :key="row.status"
          >
            <td class="px-6 py-4 text-sm font-bold">
              <FollowUpStatusBadge
                :status="row.status"
                size="md"
              />
            </td>
            <td class="px-6 py-4 text-sm font-semibold tabular-nums text-gray-900 dark:text-white">
              {{ row.jumlah || 0 }}
            </td>
            <td class="px-6 py-4 text-sm font-semibold text-gray-600 dark:text-white tabular-nums">
              {{ formatPercent(row.persentase) }}%
            </td>
            <td class="px-6 py-2">
              <UTextarea
                v-model="row.keterangan"
                :placeholder="t('executiveSummaryForm.followUp.notePlaceholder')"
                class="w-full"
                :disabled="locked"
              />
            </td>
          </tr>
          <!-- Total Row -->
          <tr class="bg-gray-50 dark:bg-gray-800/40 font-bold border-t-2">
            <td class="px-6 py-4 text-sm text-gray-900 dark:text-white uppercase tracking-wider">
              {{ t('executiveSummaryForm.table.total') }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-900 dark:text-white tabular-nums font-extrabold">
              {{ total }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-900 dark:text-white tabular-nums font-extrabold">
              {{ total > 0 ? '100.0' : '0.0' }}%
            </td>
            <td class="px-6 py-4 text-md text-gray-400 italic">
              {{ t('executiveSummaryForm.table.autoCalculated') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import type { FollowUpRow } from '~/stores/executive-summary'
import FollowUpStatusBadge from './FollowUpStatusBadge.vue'

// Section III follow-up table. Counts are read-only (from the ATR buckets); only the notes are typed.
// Old documents saved with three rows render as they are.
const props = withDefaults(defineProps<{
  rows: FollowUpRow[]
  locked?: boolean
  loading?: boolean
  error?: string
  /** Approval date label of an approved summary: shows "Status as of <date>" */
  approvedDateLabel?: string
  /** Time of the last ATR read of a draft */
  asOfLabel?: string
  refreshable?: boolean
  emptyHint?: string
}>(), {
  locked: false,
  loading: false,
  error: '',
  approvedDateLabel: '',
  asOfLabel: '',
  refreshable: false,
  emptyHint: ''
})

const emit = defineEmits<{
  refresh: []
}>()

const { t } = useI18n()

const total = computed(() => props.rows.reduce((acc, row) => acc + (Number(row.jumlah) || 0), 0))

const formatPercent = (val: number) => (val ? Number(val).toFixed(1) : '0.0')
</script>
