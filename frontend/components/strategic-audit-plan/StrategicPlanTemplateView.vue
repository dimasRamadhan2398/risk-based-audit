<template>
  <div
    v-if="soRows.length"
    class="space-y-6"
  >
    <!-- Sasaran Strategis -->
    <section class="space-y-2">
      <h2 class="text-lg font-semibold text-[var(--text-main)]">
        {{ t('strategicPlan.template.objectivesTitle') }}
      </h2>
      <div class="overflow-x-auto rounded-xl border border-[var(--border-main)]">
        <table class="min-w-full text-sm">
          <thead class="bg-[var(--bg-surface)] text-left text-[var(--text-muted)]">
            <tr>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.soNo') }}
              </th>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.strategicObjective') }}
              </th>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.kpiNo') }}
              </th>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.kpi') }}
              </th>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.dataType') }}
              </th>
              <th
                v-for="(y, i) in years"
                :key="y"
                class="px-3 py-2 text-center whitespace-nowrap"
              >
                {{ t('strategicPlan.template.year', { n: i + 1 }) }}
                <span class="block text-xs font-normal">{{ y }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, idx) in soRows"
              :key="row.id"
              class="border-t border-[var(--border-main)] align-top"
            >
              <td
                v-if="row.span > 0"
                :rowspan="row.span"
                class="px-3 py-2 font-medium whitespace-nowrap"
              >
                {{ row.code || '-' }}
              </td>
              <td
                v-if="row.span > 0"
                :rowspan="row.span"
                class="px-3 py-2 whitespace-normal break-words min-w-56"
              >
                {{ row.objective || '-' }}
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                {{ kpiNumber(idx) }}
              </td>
              <td class="px-3 py-2 whitespace-normal break-words min-w-56">
                {{ row.kpi || '-' }}
              </td>
              <td class="px-3 py-2 text-center">
                {{ row.hibHig || '-' }}
              </td>
              <td
                v-for="(cell, i) in row.cells"
                :key="i"
                class="px-3 py-2 text-center tabular-nums whitespace-nowrap"
              >
                {{ cell }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Program/Kegiatan Utama -->
    <section
      v-if="initiativeRows.length"
      class="space-y-2"
    >
      <h2 class="text-lg font-semibold text-[var(--text-main)]">
        {{ t('strategicPlan.template.initiativesTitle') }}
      </h2>
      <div class="overflow-x-auto rounded-xl border border-[var(--border-main)]">
        <table class="min-w-full text-sm">
          <thead class="bg-[var(--bg-surface)] text-left text-[var(--text-muted)]">
            <tr>
              <th class="px-3 py-2">
                {{ t('strategicPlan.template.initiative') }}
              </th>
              <th
                v-for="(y, i) in years"
                :key="y"
                class="px-3 py-2 text-center whitespace-nowrap"
              >
                {{ t('strategicPlan.template.year', { n: i + 1 }) }}
                <span class="block text-xs font-normal">{{ y }}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in initiativeRows"
              :key="row.id"
              class="border-t border-[var(--border-main)]"
            >
              <td class="px-3 py-2 whitespace-normal break-words min-w-72">
                {{ row.initiative }}
              </td>
              <td
                v-for="(cell, i) in row.cells"
                :key="i"
                class="px-3 py-2 text-center"
              >
                {{ cell }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useI18n } from '~/composables/useI18n'
import { buildInitiativeRows, buildSoRows, templateYears } from '~/utils/strategicPlanTemplate'

const { t, locale } = useI18n()
const store = useStrategicPlanStore()

const plans = computed(() => store.strategicObjectives)
const years = computed(() => templateYears(plans.value))
const soRows = computed(() => buildSoRows(plans.value, years.value, locale.value))
const initiativeRows = computed(() => buildInitiativeRows(plans.value, years.value))

// "1.1", "1.2"...: SO sequence, then the KPI's position inside its SO group.
const kpiNumber = (idx: number): string => {
  let head = idx
  while (head > 0 && soRows.value[head]!.span === 0) head--
  const soIndex = soRows.value.slice(0, head + 1).filter(r => r.span > 0).length
  return `${soIndex}.${idx - head + 1}`
}
</script>
