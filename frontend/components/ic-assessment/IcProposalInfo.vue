<template>
  <UModal
    v-model:open="open"
    :title="t('icAssessment.proposalInfo.title')"
    :description="t('icAssessment.proposalInfo.intro')"
    :ui="{ content: 'sm:max-w-xl' }"
  >
    <UButton
      icon="i-lucide-info"
      color="neutral"
      variant="outline"
      :aria-label="t('icAssessment.proposalInfo.label')"
      :title="t('icAssessment.proposalInfo.label')"
      data-testid="ic-proposal-info"
    />

    <template #body>
      <div class="space-y-5 text-sm text-gray-700 dark:text-gray-300">
        <UAlert
          color="info"
          variant="soft"
          icon="i-lucide-bot-off"
          :description="t('icAssessment.proposalInfo.notAi')"
        />

        <section class="space-y-2">
          <h4 class="font-semibold text-gray-900 dark:text-white">
            {{ t('icAssessment.proposalInfo.signalsTitle') }}
          </h4>
          <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
            <table class="w-full text-left">
              <thead class="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">
                <tr>
                  <th class="px-3 py-2 font-semibold">
                    {{ t('icAssessment.proposalInfo.colSignal') }}
                  </th>
                  <th class="px-3 py-2 font-semibold text-right">
                    {{ t('icAssessment.proposalInfo.colWeight') }}
                  </th>
                </tr>
              </thead>
              <tbody
                v-for="code in COSO_COMPONENT_CODES"
                :key="code"
                class="divide-y divide-gray-100 dark:divide-gray-800 border-t border-gray-200 dark:border-gray-800"
              >
                <tr class="bg-gray-50/60 dark:bg-gray-800/30">
                  <th
                    colspan="2"
                    class="px-3 py-1.5 font-semibold text-gray-900 dark:text-white"
                  >
                    {{ componentLabel(t, code) }}
                  </th>
                </tr>
                <tr
                  v-for="sig in SIGNAL_WEIGHTS[code]"
                  :key="sig.key"
                >
                  <td class="px-3 py-1.5 pl-5">
                    {{ signalLabel(t, sig.key) }}
                  </td>
                  <td class="px-3 py-1.5 text-right tabular-nums">
                    {{ sig.weight }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="space-y-2">
          <h4 class="font-semibold text-gray-900 dark:text-white">
            {{ t('icAssessment.proposalInfo.calcTitle') }}
          </h4>
          <ul class="ml-4 list-disc space-y-1">
            <li>{{ t('icAssessment.proposalInfo.calcRatio') }}</li>
            <li>{{ t('icAssessment.proposalInfo.calcBands') }}</li>
            <li>{{ t('icAssessment.proposalInfo.calcExcluded') }}</li>
            <li>{{ t('icAssessment.proposalInfo.calcInsufficient', { pct: MIN_COVERAGE_PCT }) }}</li>
            <li>{{ t('icAssessment.proposalInfo.calcThresholds', { lha: DEFAULT_THRESHOLDS.lhaDueDays, mgmt: DEFAULT_THRESHOLDS.managementResponseDueDays }) }}</li>
          </ul>
        </section>

        <section class="space-y-2">
          <h4 class="font-semibold text-gray-900 dark:text-white">
            {{ t('icAssessment.proposalInfo.useTitle') }}
          </h4>
          <ul class="ml-4 list-disc space-y-1">
            <li>{{ t('icAssessment.proposalInfo.useStartingPoint') }}</li>
            <li>{{ t('icAssessment.proposalInfo.useOverride', { min: OVERRIDE_REASON_MIN }) }}</li>
            <li>{{ t('icAssessment.proposalInfo.useEvidence') }}</li>
          </ul>
        </section>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { COSO_COMPONENT_CODES, DEFAULT_THRESHOLDS, MIN_COVERAGE_PCT, OVERRIDE_REASON_MIN, SIGNAL_WEIGHTS } from '~/utils/icAssessment'
import { componentLabel, signalLabel } from '~/utils/icAssessmentLabels'

const { t } = useI18n()
const open = ref(false)
</script>
