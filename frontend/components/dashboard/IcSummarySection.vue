<template>
  <section
    v-if="caps.canRead"
    class="p-5 sm:p-6 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-2xl shadow-xs space-y-4"
    data-testid="ic-summary-section"
  >
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div>
        <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <UIcon
            name="i-lucide-shield-check"
            class="size-5 text-sky-500"
          />
          {{ t('icAssessment.dashboard.title') }}
        </h2>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
          {{ t('icAssessment.dashboard.subtitle', { period }) }}
        </p>
      </div>
      <NuxtLink
        to="/internal-control-assessment"
        class="text-xs font-bold text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
      >
        <span>{{ t('icAssessment.dashboard.open') }}</span>
        <span aria-hidden="true">&rarr;</span>
      </NuxtLink>
    </div>

    <p
      v-if="store.summaryLoading && !store.summary"
      class="text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('icAssessment.dashboard.loading') }}
    </p>

    <div
      v-else-if="store.summaryFailed"
      class="py-6 text-center text-sm text-amber-800 dark:text-amber-300 flex items-center justify-center gap-2"
    >
      <UIcon
        name="i-lucide-triangle-alert"
        class="size-4 shrink-0"
        aria-hidden="true"
      />
      <span>{{ t('icAssessment.dashboard.loadError') }}</span>
    </div>

    <div
      v-else-if="!summary || summary.total === 0"
      class="py-6 text-center text-sm text-gray-500 dark:text-gray-400"
      data-testid="ic-summary-empty"
    >
      <p class="font-medium">
        {{ t('icAssessment.dashboard.empty') }}
      </p>
      <NuxtLink
        to="/internal-control-assessment"
        class="text-sky-600 dark:text-sky-400 hover:underline text-xs font-semibold"
      >
        {{ t('icAssessment.dashboard.emptyLink') }}
      </NuxtLink>
    </div>

    <template v-else>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
        <div
          v-for="code in COSO_COMPONENT_CODES"
          :key="code"
          class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 space-y-2"
        >
          <p class="text-sm font-semibold text-gray-800 dark:text-gray-100">
            {{ t(`coso.components.${code}`) }}
          </p>
          <ul class="space-y-1 text-sm">
            <li
              v-for="status in STATUSES"
              :key="status"
              class="flex items-center justify-between gap-2"
            >
              <IcStatusText
                :status="status"
                short
              />
              <span class="font-bold tabular-nums text-gray-900 dark:text-white">{{ summary.byComponent[code][status] }}</span>
            </li>
          </ul>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-x-6 gap-y-1 text-sm text-gray-600 dark:text-gray-300 border-t border-gray-100 dark:border-gray-800 pt-3">
        <span>{{ t('icAssessment.dashboard.departments', { count: summary.total }) }}</span>
        <span>{{ t('icAssessment.conclusion.EFFECTIVE') }}: <strong class="tabular-nums">{{ summary.byConclusion.EFFECTIVE }}</strong></span>
        <span>{{ t('icAssessment.conclusion.NOT_EFFECTIVE') }}: <strong class="tabular-nums">{{ summary.byConclusion.NOT_EFFECTIVE }}</strong></span>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useFiscalYear } from '~/composables/useFiscalYear'
import { useIcAssessmentStore } from '~/stores/ic-assessment'
import IcStatusText from '~/components/ic-assessment/IcStatusText.vue'
import { COSO_COMPONENT_CODES } from '~/utils/icAssessment'
import { periodLabel } from '~/utils/icAssessmentLabels'

// COSO 2013 component results of the approved internal control assessments (GET /ic-assessments/summary).
// Shows the annual assessment (quarter 0) of the selected fiscal year.
const STATUSES = ['YES', 'PARTIAL', 'NO'] as const

const { t } = useI18n()
const { internalControlCaps } = useRbac()
const caps = computed(() => internalControlCaps.value)
const store = useIcAssessmentStore()
const { selectedFiscalYear } = useFiscalYear()

const summary = computed(() => store.summary)
const period = computed(() => periodLabel(t, selectedFiscalYear.value, 0))

onMounted(() => {
  if (caps.value.canRead) store.fetchSummary(selectedFiscalYear.value, 0)
})
</script>
