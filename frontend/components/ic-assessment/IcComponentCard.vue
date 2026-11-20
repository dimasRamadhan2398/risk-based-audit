<template>
  <UCard
    class="border border-gray-200 dark:border-gray-800 dark:bg-gray-900 rounded-2xl"
    :ui="{ body: 'space-y-4' }"
  >
    <template #header>
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="text-xs font-mono text-gray-500 dark:text-gray-400">
            {{ component.code }}
          </p>
          <h3 class="text-base font-bold text-gray-900 dark:text-white">
            {{ componentLabel(t, component.code) }}
          </h3>
        </div>
        <UButton
          size="xs"
          color="neutral"
          variant="outline"
          icon="i-lucide-file-search"
          :label="t('icAssessment.actions.evidence')"
          :disabled="proposalRun < 1"
          @click="emit('openEvidence', component.code)"
        />
      </div>
    </template>

    <!-- System proposal -->
    <section class="rounded-xl bg-gray-50 dark:bg-gray-800/50 p-3 space-y-1">
      <div class="text-sm font-bold text-gray-900 dark:text-white">
        {{ t('icAssessment.component.proposed') }}
      </div>
      <template v-if="proposal.kind === 'notProposed'">
        <p class="text-sm text-gray-600 dark:text-gray-300">
          {{ t('icAssessment.component.notProposed') }}
        </p>
      </template>
      <template v-else-if="proposal.kind === 'insufficient'">
        <p class="text-base font-bold text-gray-800 dark:text-gray-100 flex items-center gap-1.5">
          <UIcon
            name="i-lucide-database-zap"
            class="size-4 shrink-0"
          />
          {{ t('icAssessment.component.insufficientData') }}
        </p>
        <p class="text-xs text-gray-600 dark:text-gray-300">
          {{ t('icAssessment.component.insufficientHelp') }}
        </p>
        <p
          v-if="proposal.coveragePct !== null"
          class="text-xs text-gray-600 dark:text-gray-300"
        >
          {{ t('icAssessment.component.coverage', { pct: formatPct(proposal.coveragePct) ?? '-' }) }}
        </p>
        <div
          v-if="proposal.reasons.length"
          class="mt-2 rounded-xl bg-info/10 ring ring-inset ring-info/25 p-3 space-y-2"
        >
          <p class="flex items-center gap-2 text-sm font-semibold text-gray-900 dark:text-white">
            <UIcon
              name="i-lucide-info"
              class="size-4 shrink-0 text-info"
            />
            {{ t('icAssessment.component.insufficientReasons') }}
          </p>
          <ul class="space-y-1.5">
            <li
              v-for="r in proposal.reasons"
              :key="r.signalKey"
              class="flex items-center justify-between gap-3 rounded-lg bg-white/70 dark:bg-gray-900/60 px-3 py-2 text-sm text-gray-800 dark:text-gray-100"
            >
              <span class="min-w-0">{{ signalLabel(t, r.signalKey) }}</span>
              <span class="shrink-0 rounded-md bg-gray-100 dark:bg-gray-800 px-2 py-0.5 text-xs text-gray-600 dark:text-gray-300">
                {{ t(`icAssessment.evidenceState.${r.state}`) }}
              </span>
            </li>
          </ul>
        </div>
      </template>
      <template v-else>
        <div class="flex items-baseline gap-3">
          <span class="text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ formatPct(proposal.pct) }}</span>
          <span class="text-sm text-gray-700 dark:text-gray-200">{{ ratingLabel(t, proposal.rating) }}</span>
        </div>
        <p
          v-if="proposal.coveragePct !== null"
          class="text-xs text-gray-600 dark:text-gray-300"
        >
          {{ t('icAssessment.component.coverage', { pct: formatPct(proposal.coveragePct) ?? '-' }) }}
        </p>
        <p
          v-if="proposal.partial"
          class="text-xs text-amber-800 dark:text-amber-300"
        >
          {{ t('icAssessment.component.partialProposal') }}
        </p>
      </template>
    </section>

    <!-- Derived from principles -->
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
      <span class="text-gray-500 dark:text-gray-400">{{ t('icAssessment.component.derived') }}:</span>
      <IcStatusText
        :status="derived.status"
        short
      />
      <span
        v-if="derived.severity !== 'NONE'"
        class="text-gray-700 dark:text-gray-200"
      >{{ t(`icAssessment.severity.${derived.severity}`) }}</span>
      <span
        v-if="derived.incomplete"
        class="text-xs text-gray-500 dark:text-gray-400 italic"
      >{{ t('icAssessment.component.derivedIncomplete') }}</span>
    </div>

    <!-- Final values -->
    <section class="space-y-3">
      <div class="text-sm font-bold text-gray-900 dark:text-white">
        {{ t('icAssessment.component.final') }}
      </div>
      <div
        v-if="editable"
        class="grid grid-cols-1 sm:grid-cols-3 gap-3"
      >
        <UFormField
          :label="t('icAssessment.component.finalRating')"
          :error="fieldError('rating')"
        >
          <USelect
            :model-value="component.finalRating ?? undefined"
            :items="ratingItems"
            :placeholder="t('icAssessment.component.select')"
            class="w-full"
            @update:model-value="(v: unknown) => update({ finalRating: v === undefined || v === null ? null : Number(v) })"
          />
        </UFormField>
        <UFormField
          :label="t('icAssessment.component.finalStatus')"
          :error="fieldError('status')"
        >
          <USelect
            :model-value="component.finalStatus ?? undefined"
            :items="statusItems"
            :placeholder="t('icAssessment.component.select')"
            class="w-full"
            @update:model-value="(v: unknown) => update({ finalStatus: (v as ComponentStatus) ?? null })"
          />
        </UFormField>
        <UFormField
          :label="t('icAssessment.component.finalSeverity')"
          :error="fieldError('severity')"
        >
          <USelect
            :model-value="component.finalSeverity ?? undefined"
            :items="severityItems"
            :placeholder="t('icAssessment.component.select')"
            class="w-full"
            @update:model-value="(v: unknown) => update({ finalSeverity: (v as Severity) ?? null })"
          />
        </UFormField>
      </div>
      <dl
        v-else
        class="grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm"
      >
        <div>
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('icAssessment.component.finalRating') }}
          </dt>
          <dd class="font-semibold text-gray-900 dark:text-white">
            {{ ratingLabel(t, component.finalRating) }}
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('icAssessment.component.finalStatus') }}
          </dt>
          <dd>
            <IcStatusText
              :status="component.finalStatus"
              short
            />
          </dd>
        </div>
        <div>
          <dt class="text-gray-500 dark:text-gray-400">
            {{ t('icAssessment.component.finalSeverity') }}
          </dt>
          <dd class="font-semibold text-gray-900 dark:text-white">
            {{ component.finalSeverity ? t(`icAssessment.severity.${component.finalSeverity}`) : '-' }}
          </dd>
        </div>
      </dl>

      <!-- Override reason: required and highlighted when the final rating differs from the proposal -->
      <div
        v-if="override || component.overrideReason"
        class="rounded-xl p-3 space-y-2"
        :class="override ? 'border-2 border-amber-400 dark:border-amber-500/70 bg-amber-50/60 dark:bg-amber-950/30' : 'border border-gray-200 dark:border-gray-800'"
      >
        <div class="flex items-center gap-1.5 text-sm font-semibold text-amber-900 dark:text-amber-200">
          <UIcon
            v-if="override"
            name="i-lucide-triangle-alert"
            class="size-4 shrink-0"
          />
          {{ t('icAssessment.component.overrideReason') }}
        </div>
        <p
          v-if="override && editable"
          class="text-xs text-amber-900 dark:text-amber-200"
        >
          {{ component.proposedRating === null
            ? t('icAssessment.component.overrideRequiredNoProposal', { min: OVERRIDE_REASON_MIN })
            : t('icAssessment.component.overrideRequired', { min: OVERRIDE_REASON_MIN }) }}
        </p>
        <template v-if="editable">
          <UTextarea
            :model-value="component.overrideReason"
            :rows="2"
            autoresize
            :maxlength="2000"
            class="w-full"
            :color="showErrors && reasonMissing ? 'error' : undefined"
            :highlight="showErrors && reasonMissing"
            @update:model-value="(v: unknown) => update({ overrideReason: String(v ?? '') })"
          />
          <p
            class="text-xs tabular-nums"
            :class="reasonMissing ? 'text-red-700 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'"
          >
            {{ t('icAssessment.component.overrideCount', { count: component.overrideReason.trim().length, min: OVERRIDE_REASON_MIN }) }}
          </p>
        </template>
        <p
          v-else
          class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-line"
        >
          {{ component.overrideReason || '-' }}
        </p>
      </div>

      <UFormField :label="t('icAssessment.component.conclusion')">
        <UTextarea
          v-if="editable"
          :model-value="component.conclusion"
          :rows="2"
          autoresize
          :maxlength="4000"
          class="w-full"
          @update:model-value="(v: unknown) => update({ conclusion: String(v ?? '') })"
        />
        <p
          v-else
          class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-line"
        >
          {{ component.conclusion || '-' }}
        </p>
      </UFormField>

      <ul
        v-if="showErrors && visibleIssues.length"
        class="text-xs text-red-700 dark:text-red-400 space-y-0.5"
      >
        <li
          v-for="i in visibleIssues"
          :key="i"
          class="flex items-start gap-1"
        >
          <UIcon
            name="i-lucide-circle-alert"
            class="size-3.5 mt-0.5 shrink-0"
          />
          {{ t(`icAssessment.blockers.component.${i}`, { component: component.code }) }}
        </li>
      </ul>
    </section>
  </UCard>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '~/composables/useI18n'
import IcStatusText from '~/components/ic-assessment/IcStatusText.vue'
import {
  COMPONENT_STATUSES,
  OVERRIDE_REASON_MIN,
  SEVERITIES,
  componentIssues,
  formatPct,
  isOverride,
  overrideReasonMissing,
  proposalView,
  type ComponentCode,
  type ComponentStatus,
  type DerivedComponent,
  type IcComponent,
  type IcEvidence,
  type Severity
} from '~/utils/icAssessment'
import { componentLabel, ratingLabel, signalLabel } from '~/utils/icAssessmentLabels'

const props = defineProps<{
  component: IcComponent
  derived: DerivedComponent
  proposalRun: number
  evidence: IcEvidence[]
  editable: boolean
  showErrors: boolean
}>()

const emit = defineEmits<{
  update: [patch: Partial<IcComponent>]
  openEvidence: [code: ComponentCode]
}>()

const { t } = useI18n()

const proposal = computed(() => proposalView(props.component, props.proposalRun, props.evidence))
const override = computed(() => isOverride(props.component))
const reasonMissing = computed(() => overrideReasonMissing(props.component))
const issues = computed(() => componentIssues(props.component, props.derived))
// finalMissing is shown on the fields themselves
const visibleIssues = computed(() => issues.value.filter(i => i !== 'finalMissing'))

const fieldError = (field: 'rating' | 'status' | 'severity'): string | undefined => {
  if (!props.showErrors) return undefined
  const c = props.component
  const missing = field === 'rating' ? c.finalRating === null : field === 'status' ? !c.finalStatus : !c.finalSeverity
  return missing ? t('icAssessment.component.select') : undefined
}

const ratingItems = computed(() => [5, 4, 3, 2, 1].map(r => ({ label: ratingLabel(t, r), value: r })))
const statusItems = computed(() => COMPONENT_STATUSES.map(s => ({ label: t(`icAssessment.componentStatus.${s}`), value: s })))
const severityItems = computed(() => SEVERITIES.map(s => ({ label: t(`icAssessment.severity.${s}`), value: s })))

const update = (patch: Partial<IcComponent>) => emit('update', patch)
</script>
