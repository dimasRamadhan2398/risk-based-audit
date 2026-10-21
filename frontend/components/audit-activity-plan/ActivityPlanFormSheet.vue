<template>
  <div
    class="space-y-4 text-sm text-[var(--text-main)]"
    data-testid="activity-plan-sheet"
  >
    <h2 class="text-center text-base sm:text-lg font-bold uppercase tracking-wide">
      {{ t('auditActivityPlan.formTitle') }}
    </h2>

    <!-- Header block of the paper form -->
    <div class="grid grid-cols-1 gap-x-6 gap-y-2 rounded-lg border border-[var(--border-main)] p-3 sm:grid-cols-[12rem_1fr]">
      <span class="font-semibold">{{ t('auditActivityPlan.header.year') }}</span>
      <span data-testid="sheet-year">{{ sheet.year || '-' }}</span>

      <span class="font-semibold">{{ t('auditActivityPlan.header.activityType') }}</span>
      <div
        class="flex flex-wrap gap-x-4 gap-y-1"
        data-testid="sheet-activity-type"
      >
        <span
          v-for="type in ACTIVITY_PLAN_ACTIVITY_TYPES"
          :key="type"
          class="inline-flex items-center gap-1"
          :class="type === sheet.activityType ? 'font-semibold' : 'text-[var(--text-muted)]'"
          :data-checked="type === sheet.activityType ? 'true' : 'false'"
        >
          <UIcon
            :name="type === sheet.activityType ? 'i-heroicons-check-circle-solid' : 'i-heroicons-stop'"
            class="h-4 w-4"
            :class="type === sheet.activityType ? 'text-primary-500' : ''"
          />
          {{ t(ACTIVITY_TYPE_LABEL_KEYS[type]) }}
        </span>
      </div>

      <span class="font-semibold">{{ t('auditActivityPlan.header.title') }}</span>
      <span
        class="break-words"
        data-testid="sheet-title"
      >{{ sheet.title || '-' }}</span>
    </div>

    <!-- The 12 columns, in the order of the paper form -->
    <div class="overflow-x-auto rounded-lg border border-[var(--border-main)]">
      <table class="min-w-[1400px] w-full border-collapse text-xs">
        <thead class="bg-gray-100 dark:bg-gray-800">
          <tr>
            <th
              v-for="column in columns"
              :key="column.key"
              class="border border-[var(--border-main)] px-2 py-2 text-left font-semibold align-bottom"
              :class="column.class"
              data-testid="sheet-column"
            >
              {{ t(column.labelKey) }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!sheet.rows.length">
            <td
              :colspan="columns.length"
              class="border border-[var(--border-main)] px-2 py-6 text-center italic text-[var(--text-muted)]"
            >
              {{ t('auditActivityPlan.rows.empty') }}
            </td>
          </tr>
          <tr
            v-for="row in sheet.rows"
            :key="row.key"
            class="align-top"
            data-testid="sheet-row"
          >
            <td class="border border-[var(--border-main)] px-2 py-2 text-center">
              {{ row.no }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 font-mono">
              {{ row.assignmentCode || pendingCode }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 break-words">
              {{ row.auditObject || '-' }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2">
              {{ row.auditKind || '-' }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 break-words">
              <span>{{ row.auditUniverse?.name || '-' }}</span>
              <span
                v-if="row.auditUniverse?.riskLevel"
                class="ml-1 inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold"
                :class="getRiskLevelColorClass(row.auditUniverse?.riskLevel)"
              >{{ row.auditUniverse?.riskLevel }}</span>
            </td>
            <td
              v-for="field in listFields"
              :key="field"
              class="border border-[var(--border-main)] px-2 py-2"
            >
              <ol
                v-if="row[field].length"
                class="list-decimal space-y-0.5 pl-4"
                :data-testid="`sheet-list-${field}`"
              >
                <li
                  v-for="(item, i) in row[field]"
                  :key="i"
                  class="break-words"
                >
                  {{ item }}
                </li>
              </ol>
              <span v-else>-</span>
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2">
              {{ row.schedule || '-' }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 text-center">
              {{ row.teamSize }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 text-center">
              {{ row.totalMandays }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2">
              {{ row.teamLeaderName || '-' }}
            </td>
          </tr>
        </tbody>
        <tfoot
          v-if="sheet.rows.length"
          class="bg-gray-50 dark:bg-gray-900/60 font-semibold"
        >
          <tr>
            <td
              :colspan="9"
              class="border border-[var(--border-main)] px-2 py-2 text-right"
            >
              {{ t('auditActivityPlan.totals.rows') }}: {{ sheet.totals.rowCount }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 text-center">
              {{ sheet.totals.teamSize }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2 text-center">
              {{ sheet.totals.totalMandays }}
            </td>
            <td class="border border-[var(--border-main)] px-2 py-2" />
          </tr>
        </tfoot>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { ACTIVITY_PLAN_ACTIVITY_TYPES } from '~/types/audit'
import { ACTIVITY_TYPE_LABEL_KEYS, type ActivityPlanSheet } from '~/utils/activityPlanForm'
import { getRiskLevelColorClass } from '~/utils/riskLevelBadge'

/** Read-only render of the paper Formulir Rencana Kegiatan Audit. Used by the preview and the view modal. */
defineProps<{
  sheet: ActivityPlanSheet
}>()

const { t } = useI18n()

const listFields = ['justification', 'objectives', 'focus'] as const

/** Paper column order. */
const columns = [
  { key: 'no', labelKey: 'auditActivityPlan.columns.no', class: 'w-10 text-center' },
  { key: 'assignmentCode', labelKey: 'auditActivityPlan.columns.assignmentCode', class: 'w-28' },
  { key: 'auditObject', labelKey: 'auditActivityPlan.columns.auditObject', class: 'w-44' },
  { key: 'auditKind', labelKey: 'auditActivityPlan.columns.auditKind', class: 'w-28' },
  { key: 'auditUniverse', labelKey: 'auditActivityPlan.columns.auditUniverse', class: 'w-44' },
  { key: 'justification', labelKey: 'auditActivityPlan.columns.justification', class: 'w-56' },
  { key: 'objectives', labelKey: 'auditActivityPlan.columns.objectives', class: 'w-56' },
  { key: 'focus', labelKey: 'auditActivityPlan.columns.focus', class: 'w-56' },
  { key: 'schedule', labelKey: 'auditActivityPlan.columns.schedule', class: 'w-24' },
  { key: 'teamSize', labelKey: 'auditActivityPlan.columns.teamSize', class: 'w-20 text-center' },
  { key: 'totalMandays', labelKey: 'auditActivityPlan.columns.totalMandays', class: 'w-20 text-center' },
  { key: 'teamLeader', labelKey: 'auditActivityPlan.columns.teamLeader', class: 'w-36' }
] as const

const pendingCode = computed(() => t('auditActivityPlan.rows.assignmentCodePending'))
</script>
