<template>
  <USlideover
    :open="open"
    :title="t('icAssessment.evidence.title', { component: componentName })"
    :description="t('icAssessment.evidence.description')"
    :ui="{ content: 'max-w-2xl bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100' }"
    @update:open="emit('update:open', $event)"
  >
    <template #body>
      <div class="space-y-5">
        <UFormField v-if="proposalRun > 1" :label="t('icAssessment.evidence.run')">
          <USelect v-model="run" :items="runOptions" class="w-48" />
        </UFormField>

        <!-- Thresholds used (frozen with the run); relevant to the IC timeliness signals -->
        <div v-if="componentCode === 'IC'" class="rounded-xl border border-gray-200 dark:border-gray-800 p-3 space-y-1">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('icAssessment.evidence.thresholdsUsed') }}</p>
          <template v-if="thresholds">
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ t('icAssessment.evidence.lhaDueDays', { days: thresholds.lhaDueDays }) }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ t('icAssessment.evidence.mrDueDays', { days: thresholds.managementResponseDueDays }) }}</p>
          </template>
          <p v-else class="text-sm text-gray-500 dark:text-gray-400">{{ t('icAssessment.evidence.noThresholds') }}</p>
        </div>

        <div v-if="loading" class="space-y-3">
          <USkeleton v-for="i in 3" :key="i" class="h-24 w-full rounded-xl" />
        </div>
        <p v-else-if="loadFailed" class="text-sm text-red-700 dark:text-red-400">{{ t('icAssessment.evidence.loadError') }}</p>
        <p v-else-if="!rows.length" class="text-sm text-gray-500 dark:text-gray-400">{{ t('icAssessment.evidence.empty') }}</p>

        <ul v-else class="space-y-3">
          <li
            v-for="e in rows"
            :key="e.id || e.signalKey"
            class="rounded-xl border border-gray-200 dark:border-gray-800 p-4 space-y-2"
          >
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ signalLabel(t, e.signalKey) }}</p>
                <p class="text-xs text-gray-500 dark:text-gray-400 font-mono">{{ e.signalKey }}</p>
              </div>
              <div class="text-right shrink-0">
                <p class="text-lg font-bold tabular-nums text-gray-900 dark:text-white">
                  {{ formatPct(e.valuePct) ?? '-' }}
                </p>
                <p v-if="e.numerator !== null && e.denominator !== null" class="text-xs text-gray-500 dark:text-gray-400 tabular-nums">
                  {{ t('icAssessment.evidence.ratio', { num: e.numerator, den: e.denominator }) }}
                </p>
              </div>
            </div>

            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs">
              <span class="inline-flex items-center gap-1 font-medium" :class="EVIDENCE_STATE_VISUAL[e.state].cls">
                <UIcon :name="EVIDENCE_STATE_VISUAL[e.state].icon" class="size-3.5" />
                {{ t(`icAssessment.evidenceState.${e.state}`) }}
              </span>
              <span class="text-gray-600 dark:text-gray-300">{{ t('icAssessment.evidence.weight', { weight: e.weight }) }}</span>
              <span v-if="e.principleHint" class="text-gray-600 dark:text-gray-300">{{ t('icAssessment.evidence.principleHint', { code: e.principleHint }) }}</span>
              <span v-if="e.collectedAt" class="text-gray-500 dark:text-gray-400">{{ t('icAssessment.evidence.collectedAt', { date: formatDateTime(e.collectedAt) }) }}</span>
            </div>

            <p v-if="e.scopeNote" class="text-xs text-gray-600 dark:text-gray-300">
              <span class="font-semibold">{{ t('icAssessment.evidence.scope') }}:</span> {{ e.scopeNote }}
            </p>

            <dl v-if="detailEntries(e).length" class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-0.5 text-xs">
              <template v-for="[k, v] in detailEntries(e)" :key="k">
                <dt class="text-gray-500 dark:text-gray-400 font-mono truncate">{{ k }}</dt>
                <dd class="text-gray-800 dark:text-gray-200 tabular-nums break-words">{{ v }}</dd>
              </template>
            </dl>

            <div v-if="evidenceSourceLink(e, departmentId)" class="pt-1">
              <UButton
                size="xs"
                variant="link"
                color="primary"
                icon="i-lucide-external-link"
                :label="t('icAssessment.evidence.openSource')"
                :to="evidenceSourceLink(e, departmentId)!"
              />
            </div>
          </li>
        </ul>
      </div>
    </template>
  </USlideover>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useIcAssessmentStore } from '~/stores/ic-assessment'
import { formatPct, type ComponentCode, type IcEvidence, type IcThresholds } from '~/utils/icAssessment'
import { componentLabel, EVIDENCE_STATE_VISUAL, evidenceSourceLink, signalLabel } from '~/utils/icAssessmentLabels'

const props = defineProps<{
  open: boolean
  assessmentId: string
  departmentId: string
  componentCode: ComponentCode | null
  proposalRun: number
}>()

const emit = defineEmits<{ 'update:open': [value: boolean] }>()

const { t, locale } = useI18n()
const store = useIcAssessmentStore()

const run = ref(props.proposalRun)
const items = ref<IcEvidence[]>([])
const thresholds = ref<IcThresholds | null>(null)
const loading = ref(false)
const loadFailed = ref(false)

const componentName = computed(() => (props.componentCode ? componentLabel(t, props.componentCode) : ''))

const runOptions = computed(() =>
  Array.from({ length: props.proposalRun }, (_, i) => props.proposalRun - i).map(r => ({
    label: r === props.proposalRun ? `${t('icAssessment.evidence.runN', { run: r })} (${t('icAssessment.evidence.latest')})` : t('icAssessment.evidence.runN', { run: r }),
    value: r
  }))
)

const rows = computed(() => items.value.filter(e => e.componentCode === props.componentCode))

const load = async () => {
  if (!props.open || !props.assessmentId || run.value < 1) return
  loading.value = true
  loadFailed.value = false
  const res = await store.fetchEvidence(props.assessmentId, run.value)
  loading.value = false
  if (!res) {
    loadFailed.value = true
    items.value = []
    thresholds.value = null
    return
  }
  items.value = res.items
  thresholds.value = res.thresholds ?? thresholdsFromDetail(res.items)
}

// Fallback: thresholds recorded per signal in `detail` (threshold_days)
const thresholdsFromDetail = (list: IcEvidence[]): IcThresholds | null => {
  const days = (key: string) => {
    const v = list.find(e => e.signalKey === key)?.detail?.threshold_days
    return typeof v === 'number' ? v : null
  }
  const lha = days('ic.lha_timeliness')
  const mr = days('ic.management_response')
  return lha === null && mr === null ? null : { lhaDueDays: lha ?? 0, managementResponseDueDays: mr ?? 0 }
}

watch(() => [props.open, props.proposalRun] as const, ([open]) => {
  if (open) {
    run.value = props.proposalRun
    load()
  }
}, { immediate: true })
watch(run, () => load())

/** Scalar detail values; id lists are shown as their count. */
const detailEntries = (e: IcEvidence): [string, string][] =>
  Object.entries(e.detail || {})
    .filter(([, v]) => Array.isArray(v) || (v !== null && v !== undefined && typeof v !== 'object'))
    .map(([k, v]) => [k, Array.isArray(v) ? String(v.length) : String(v)] as [string, string])
    .slice(0, 20)

const formatDateTime = (iso: string) => {
  const d = new Date(iso)
  return isNaN(d.getTime()) ? iso : d.toLocaleString(locale.value === 'en' ? 'en-US' : 'id-ID', { dateStyle: 'medium', timeStyle: 'short' })
}
</script>
