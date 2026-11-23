<template>
  <div class="space-y-6 min-w-0">
    <button
      type="button"
      class="inline-flex items-center gap-1 text-sm text-primary-600 dark:text-primary-400 hover:underline"
      @click="goBack"
    >
      <UIcon
        name="i-lucide-arrow-left"
        class="size-4"
        aria-hidden="true"
      />
      {{ t('icAssessment.detail.back') }}
    </button>

    <!-- Loading / error states -->
    <div
      v-if="store.detailLoading && !draft"
      class="py-16 text-center text-sm text-gray-500 dark:text-gray-400"
    >
      <UIcon
        name="i-lucide-loader-circle"
        class="size-6 animate-spin mx-auto mb-2"
        aria-hidden="true"
      />
      {{ t('icAssessment.dashboard.loading') }}
    </div>
    <UAlert
      v-else-if="!draft"
      color="warning"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      :title="notFound ? t('icAssessment.detail.notFound') : t('icAssessment.detail.loadError')"
    >
      <template #actions>
        <UButton
          size="xs"
          color="neutral"
          variant="outline"
          :label="t('icAssessment.actions.reload')"
          @click="load"
        />
      </template>
    </UAlert>

    <template v-else>
      <!-- Header -->
      <div class="flex flex-col lg:flex-row lg:items-start justify-between gap-4">
        <div class="min-w-0 space-y-1">
          <h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white flex flex-wrap items-center gap-2">
            <UIcon
              name="i-lucide-shield-check"
              class="size-6 sm:size-7 text-primary-500 shrink-0"
              aria-hidden="true"
            />
            <span>{{ draft.departmentName }}</span>
            <UBadge
              :color="ASSESSMENT_STATUS_COLOR[draft.status] ?? 'neutral'"
              variant="subtle"
            >
              {{ t(`icAssessment.status.${draft.status}`) }}
            </UBadge>
          </h1>
          <p class="text-sm text-gray-600 dark:text-gray-300">
            {{ t('icAssessment.detail.period') }}: {{ periodLabel(t, draft.year, draft.quarter) }}
            <span
              class="mx-1"
              aria-hidden="true"
            >|</span>
            {{ t('icAssessment.detail.branch') }}: {{ draft.locationName || t('icAssessment.detail.departmentWide') }}
            <span
              class="mx-1"
              aria-hidden="true"
            >|</span>
            {{ t('icAssessment.detail.version', { version: draft.version }) }}
          </p>
          <p
            v-if="draft.asOfDate"
            class="text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('icAssessment.detail.asOf', { date: formatDate(draft.asOfDate) }) }}
          </p>
          <p
            v-if="draft.createdBy"
            class="text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('icAssessment.detail.createdBy', { name: createdByName }) }}
          </p>
          <p
            v-if="draft.submittedBy"
            class="text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('icAssessment.detail.submittedBy', { name: draft.submittedBy.name, date: formatDate(draft.submittedAt) }) }}
          </p>
          <p
            v-if="draft.approvedBy"
            class="text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('icAssessment.detail.approvedBy', { name: draft.approvedBy.name, date: formatDate(draft.approvedAt) }) }}
          </p>
          <p
            v-if="draft.reopenCount > 0"
            class="text-xs text-amber-800 dark:text-amber-300"
          >
            {{ t('icAssessment.detail.reopened', { count: draft.reopenCount, reason: draft.lastReopenReason || '-' }) }}
          </p>
        </div>

        <!-- Stepper: text and icon, never colour alone -->
        <ol
          class="flex items-center gap-2 text-sm shrink-0"
          :aria-label="t('icAssessment.stepper.DRAFT')"
        >
          <li
            v-for="(step, idx) in STEPS"
            :key="step"
            class="flex items-center gap-2"
          >
            <span
              class="inline-flex items-center gap-1.5 rounded-full border px-3 py-1 font-medium"
              :class="idx === stepIndex
                ? 'border-primary-500 text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/50'
                : idx < stepIndex
                  ? 'border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-200'
                  : 'border-gray-200 dark:border-gray-800 text-gray-500 dark:text-gray-400'"
              :aria-current="idx === stepIndex ? 'step' : undefined"
            >
              <UIcon
                :name="idx < stepIndex ? 'i-lucide-check' : idx === stepIndex ? 'i-lucide-circle-dot' : 'i-lucide-circle'"
                class="size-4"
                aria-hidden="true"
              />
              {{ t(`icAssessment.stepper.${step}`) }}
            </span>
            <UIcon
              v-if="idx < STEPS.length - 1"
              name="i-lucide-chevron-right"
              class="size-4 text-gray-400"
              aria-hidden="true"
            />
          </li>
        </ol>
      </div>

      <!-- Notices -->
      <UAlert
        v-if="store.conflict"
        color="warning"
        variant="subtle"
        icon="i-lucide-triangle-alert"
        :title="t('icAssessment.toast.conflictTitle')"
        :description="t('icAssessment.toast.conflictBody')"
      >
        <template #actions>
          <UButton
            size="xs"
            color="neutral"
            variant="outline"
            :label="t('icAssessment.actions.reload')"
            @click="load"
          />
        </template>
      </UAlert>
      <UAlert
        v-if="locked"
        color="neutral"
        variant="subtle"
        icon="i-lucide-lock"
        :title="t('icAssessment.period.lockedBanner', { period: periodLabel(t, draft.year, draft.quarter) })"
      />
      <p
        v-if="readOnlyNotice"
        class="text-sm text-gray-600 dark:text-gray-300 flex items-center gap-1.5"
      >
        <UIcon
          name="i-lucide-info"
          class="size-4 shrink-0"
          aria-hidden="true"
        />
        {{ readOnlyNotice }}
      </p>

      <!-- Drift indicator (approved assessments, readers with full access) -->
      <div
        v-if="draft.status === 'APPROVED' && caps.fullRead"
        class="rounded-xl border border-gray-200 dark:border-gray-800 p-3 text-sm flex flex-wrap items-center gap-3"
        data-testid="ic-drift"
      >
        <template v-if="store.driftLoading">
          <UIcon
            name="i-lucide-loader-circle"
            class="size-4 animate-spin"
            aria-hidden="true"
          />
          <span>{{ t('icAssessment.dashboard.loading') }}</span>
        </template>
        <template v-else-if="store.driftUnavailable || !store.drift">
          <UIcon
            name="i-lucide-circle-slash"
            class="size-4 text-gray-600 dark:text-gray-300"
            aria-hidden="true"
          />
          <span>{{ t('icAssessment.drift.unavailable') }}</span>
        </template>
        <template v-else>
          <UIcon
            :name="store.drift.changed ? 'i-lucide-triangle-alert' : 'i-lucide-check'"
            class="size-4"
            :class="store.drift.changed ? 'text-amber-700 dark:text-amber-400' : 'text-emerald-700 dark:text-emerald-400'"
            aria-hidden="true"
          />
          <div class="min-w-0">
            <p class="font-medium">
              {{ store.drift.changed ? t('icAssessment.drift.changed') : t('icAssessment.drift.unchanged') }}
            </p>
            <p
              v-if="store.drift.changed"
              class="text-xs text-gray-600 dark:text-gray-300"
            >
              {{ t('icAssessment.drift.changedHelp') }}
            </p>
            <ul
              v-if="store.drift.changed"
              class="text-xs text-gray-600 dark:text-gray-300 list-disc list-inside"
            >
              <li
                v-for="s in changedDriftSignals"
                :key="s.signalKey"
              >
                {{ t('icAssessment.drift.signal', { signal: signalLabel(t, s.signalKey), from: formatPct(s.approvedValuePct) ?? '-', to: formatPct(s.currentValuePct) ?? '-' }) }}
              </li>
            </ul>
          </div>
        </template>
        <UButton
          size="xs"
          color="neutral"
          variant="outline"
          class="ml-auto"
          :label="t('icAssessment.drift.check')"
          :loading="store.driftLoading"
          @click="store.checkDrift(draft.id)"
        />
      </div>

      <!-- Action bar -->
      <!-- Sticky so Save / Discard / Submit stay reachable while scrolling (the panel body is the scroll container) -->
      <div
        class="sticky top-0 z-20 -mx-4 sm:-mx-6 px-4 sm:px-6 py-2 flex flex-wrap items-center gap-x-3 gap-y-2 bg-[var(--bg-main)] border-b border-[var(--border-main)]"
        data-testid="ic-actions"
      >
        <div class="mr-auto min-w-0 flex items-center gap-2">
          <span class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ draft.departmentName }}</span>
          <UBadge
            v-if="dirty"
            color="warning"
            variant="subtle"
            size="sm"
            class="shrink-0"
          >
            {{ t('icAssessment.detail.unsaved') }}
          </UBadge>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <UFieldGroup>
            <UButton
              v-if="actions.propose"
              :icon="draft.proposalRun > 0 ? 'i-lucide-refresh-cw' : 'i-lucide-wand-sparkles'"
              color="neutral"
              variant="outline"
              :loading="busy === 'propose'"
              :disabled="store.saving"
              :label="draft.proposalRun > 0 ? t('icAssessment.actions.repropose') : t('icAssessment.actions.propose')"
              @click="doPropose"
            />
            <IcProposalInfo />
          </UFieldGroup>
          <UButton
            v-if="actions.edit"
            icon="i-lucide-save"
            color="primary"
            :loading="busy === 'save'"
            :disabled="store.saving || !dirty"
            :label="t('icAssessment.actions.save')"
            @click="doSave"
          />
          <UButton
            v-if="actions.edit && dirty"
            icon="i-lucide-rotate-ccw"
            color="neutral"
            variant="ghost"
            :label="t('icAssessment.actions.discard')"
            :aria-label="t('icAssessment.actions.discard')"
            :ui="{ label: 'hidden sm:inline' }"
            @click="doDiscard"
          />
          <UButton
            v-if="actions.submit"
            icon="i-lucide-send"
            color="primary"
            variant="solid"
            :loading="busy === 'submit'"
            :disabled="store.saving"
            :label="t('icAssessment.actions.submit')"
            @click="doSubmit"
          />
          <UButton
            v-if="actions.approve"
            icon="i-lucide-check-check"
            color="success"
            :disabled="store.saving"
            :label="t('icAssessment.actions.approve')"
            @click="approveOpen = true"
          />
          <UButton
            v-if="actions.returnToDraft"
            icon="i-lucide-undo-2"
            color="warning"
            variant="outline"
            :disabled="store.saving"
            :label="t('icAssessment.actions.return')"
            :aria-label="t('icAssessment.actions.return')"
            :ui="{ label: 'hidden sm:inline' }"
            @click="reason = 'return'"
          />
          <UButton
            v-if="actions.reopen"
            icon="i-lucide-lock-open"
            color="warning"
            variant="outline"
            :disabled="store.saving"
            :label="t('icAssessment.actions.reopen')"
            :aria-label="t('icAssessment.actions.reopen')"
            :ui="{ label: 'hidden sm:inline' }"
            @click="reason = 'reopen'"
          />
          <UButton
            v-if="actions.delete"
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            :disabled="store.saving"
            :label="t('icAssessment.actions.delete')"
            :aria-label="t('icAssessment.actions.delete')"
            :ui="{ label: 'hidden sm:inline' }"
            @click="doDelete"
          />
        </div>
        <p
          v-if="actions.approveBlockedBySoD"
          class="text-sm text-gray-600 dark:text-gray-300"
        >
          {{ t('icAssessment.approve.sod') }}
        </p>
      </div>

      <!-- Blockers before submit -->
      <UAlert
        v-if="showErrors && blockers.length"
        color="warning"
        variant="subtle"
        icon="i-lucide-list-checks"
        :title="t('icAssessment.blockers.title')"
      >
        <template #description>
          <ul class="list-disc list-inside text-sm">
            <li
              v-for="b in blockers"
              :key="b.key"
            >
              {{ t(`icAssessment.blockers.${b.key}`, b.params) }}
            </li>
          </ul>
        </template>
      </UAlert>

      <UTabs
        v-model="tab"
        :items="tabItems"
        :content="false"
        class="w-full"
      />

      <!-- Assessment tab -->
      <div
        v-show="tab === 'assessment'"
        class="space-y-6"
      >
        <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">
          <IcComponentCard
            v-for="c in draft.components"
            :key="c.code"
            :component="c"
            :derived="derived[c.code]"
            :proposal-run="draft.proposalRun"
            :evidence="draft.evidence"
            :editable="actions.edit"
            :show-errors="showErrors"
            @update="(patch) => Object.assign(c, patch)"
            @open-evidence="openEvidence"
          />
        </div>

        <section class="space-y-2">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white">
            {{ t('icAssessment.principles.title') }}
          </h2>
          <IcPrinciplesTable
            :principles="draft.principles"
            :department-id="draft.departmentId"
            :editable="actions.edit"
            :show-errors="showErrors"
            @update="onPrincipleUpdate"
          />
        </section>

        <!-- Overall conclusion -->
        <UCard
          class="border border-gray-200 dark:border-gray-800 dark:bg-gray-900 rounded-2xl"
          :ui="{ body: 'space-y-4' }"
        >
          <template #header>
            <h2 class="text-base font-bold text-gray-900 dark:text-white">
              {{ t('icAssessment.conclusionSection.title') }}
            </h2>
          </template>

          <p class="text-sm text-gray-700 dark:text-gray-200">
            {{ t('icAssessment.conclusionSection.derived') }}:
            <strong>{{ t(`icAssessment.conclusion.${derivedConclusion}`) }}</strong>
          </p>

          <UFormField
            :label="t('icAssessment.conclusionSection.final')"
            :error="conclusionError('conclusionMissing') || conclusionError('effectiveNotAllowed')"
          >
            <USelect
              v-if="actions.edit"
              :model-value="draft.finalConclusion ?? undefined"
              :items="conclusionItems"
              :placeholder="t('icAssessment.conclusionSection.finalPlaceholder')"
              class="w-full sm:w-72"
              @update:model-value="(v: unknown) => draft && (draft.finalConclusion = (v as FinalConclusion) ?? null)"
            />
            <p
              v-else
              class="text-sm font-medium text-gray-900 dark:text-white"
            >
              {{ draft.finalConclusion ? t(`icAssessment.conclusion.${draft.finalConclusion}`) : t('icAssessment.conclusion.none') }}
            </p>
          </UFormField>

          <UFormField :label="t('icAssessment.conclusionSection.text')">
            <UTextarea
              v-if="actions.edit"
              v-model="draft.conclusionText"
              :rows="3"
              :placeholder="t('icAssessment.conclusionSection.textPlaceholder')"
              class="w-full"
            />
            <p
              v-else
              class="text-sm text-gray-700 dark:text-gray-200 whitespace-pre-line"
            >
              {{ draft.conclusionText || '-' }}
            </p>
          </UFormField>

          <UFormField
            :label="t('icAssessment.conclusionSection.justification')"
            :error="conclusionError('justificationRequired')"
          >
            <UTextarea
              v-if="actions.edit"
              v-model="draft.conclusionJustification"
              :rows="3"
              :placeholder="t('icAssessment.conclusionSection.justificationPlaceholder')"
              class="w-full"
            />
            <p
              v-else
              class="text-sm text-gray-700 dark:text-gray-200 whitespace-pre-line"
            >
              {{ draft.conclusionJustification || '-' }}
            </p>
          </UFormField>

          <p
            class="text-sm text-gray-700 dark:text-gray-200"
            data-testid="ic-operate-together"
          >
            {{ t('icAssessment.conclusionSection.operateTogether') }}:
            <strong>{{ operateTogetherText }}</strong>
            <span
              v-if="draft.operateTogetherNote"
              class="block text-xs text-gray-600 dark:text-gray-300 mt-0.5"
            >{{ draft.operateTogetherNote }}</span>
          </p>
        </UCard>
      </div>

      <!-- History tab -->
      <div v-show="tab === 'history'">
        <IcHistory
          :events="store.events"
          :loading="store.eventsLoading"
        />
      </div>
    </template>

    <IcEvidenceDrawer
      v-if="draft"
      v-model:open="evidenceOpen"
      :assessment-id="draft.id"
      :department-id="draft.departmentId"
      :component-code="evidenceCode"
      :proposal-run="draft.proposalRun"
    />

    <IcApproveModal
      v-model:open="approveOpen"
      :subtitle="draft ? `${draft.departmentName}, ${periodLabel(t, draft.year, draft.quarter)}` : ''"
      :signer-name="signerName"
      :loading="store.saving"
      @confirm="doApprove"
    />

    <IcReasonModal
      :open="reason !== null"
      :title="reason === 'reopen' ? t('icAssessment.reason.reopenTitle') : t('icAssessment.reason.returnTitle')"
      :description="reason === 'reopen' ? t('icAssessment.reason.reopenDescription') : t('icAssessment.reason.returnDescription')"
      :label="t('icAssessment.reason.label')"
      required
      confirm-color="warning"
      :loading="store.saving"
      @update:open="(o: boolean) => { if (!o) reason = null }"
      @confirm="doReason"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { useI18n } from '~/composables/useI18n'
import { useRbac } from '~/composables/useRbac'
import { useAppToast } from '~/composables/useAppToast'
import { useAuthStore } from '~/stores/auth'
import { useGlobalModalStore } from '~/stores/global-modal'
import { useIcAssessmentStore } from '~/stores/ic-assessment'
import IcComponentCard from '~/components/ic-assessment/IcComponentCard.vue'
import IcPrinciplesTable from '~/components/ic-assessment/IcPrinciplesTable.vue'
import IcEvidenceDrawer from '~/components/ic-assessment/IcEvidenceDrawer.vue'
import IcHistory from '~/components/ic-assessment/IcHistory.vue'
import IcProposalInfo from '~/components/ic-assessment/IcProposalInfo.vue'
import IcApproveModal from '~/components/ic-assessment/IcApproveModal.vue'
import IcReasonModal from '~/components/ic-assessment/IcReasonModal.vue'
import {
  FINAL_CONCLUSIONS,
  conclusionIssues,
  deriveAllComponents,
  deriveConclusion,
  formatPct,
  icActions,
  isLockedFor,
  submitBlockers,
  type ComponentCode,
  type ConclusionIssue,
  type FinalConclusion,
  type IcAssessment,
  type IcPrinciple
} from '~/utils/icAssessment'
import { buildUpdatePayload } from '~/utils/icAssessmentParse'
import { ASSESSMENT_STATUS_COLOR, periodLabel, signalLabel } from '~/utils/icAssessmentLabels'
import { getUserErrorMessage } from '~/utils/error'

definePageMeta({ middleware: 'auth' })

const STEPS = ['DRAFT', 'SUBMITTED', 'APPROVED'] as const

const { t, locale } = useI18n()
const { internalControlCaps } = useRbac()
const caps = computed(() => internalControlCaps.value)
const store = useIcAssessmentStore()
const authStore = useAuthStore()
const modal = useGlobalModalStore()
const toast = useAppToast()
const route = useRoute()
const router = useRouter()

const id = computed(() => String(route.params.id))

// The server stores the creator's username; show the full name when it is the signed-in user
const createdByName = computed(() => {
  const c = draft.value?.createdBy
  if (!c) return ''
  const me = authStore.user
  if (me?.fullName && ((c.id && c.id === String(me.id)) || (c.name && c.name === me.username))) return me.fullName
  return c.name || c.id
})

// `draft` is the working copy; `store.current` is what the server last returned.
const draft = ref<IcAssessment | null>(null)
const notFound = ref(false)
const tab = ref<'assessment' | 'history'>('assessment')
const showErrors = ref(false)
const busy = ref<'' | 'save' | 'propose' | 'submit'>('')
const evidenceOpen = ref(false)
const evidenceCode = ref<ComponentCode | null>(null)
const approveOpen = ref(false)
const reason = ref<'return' | 'reopen' | null>(null)

const clone = (a: IcAssessment): IcAssessment => JSON.parse(JSON.stringify(a))

const tabItems = computed(() => [
  { label: t('icAssessment.tabs.assessment'), value: 'assessment' },
  { label: t('icAssessment.tabs.history'), value: 'history' }
])

const original = computed(() => store.current)
const dirty = computed(() => !!original.value && !!draft.value && buildUpdatePayload(original.value, draft.value) !== null)

const locked = computed(() =>
  !!draft.value && (isLockedFor(store.periods, draft.value.year, draft.value.quarter) || draft.value.periodLocked === true)
)
const actions = computed(() => icActions(caps.value, String(authStore.user?.id ?? ''), draft.value, locked.value))
const signerName = computed(() => String(authStore.user?.fullName || authStore.user?.username || ''))

const stepIndex = computed(() => {
  const s = draft.value?.status
  return s === 'APPROVED' ? 2 : s === 'SUBMITTED' ? 1 : 0
})

const readOnlyNotice = computed(() => {
  const a = draft.value
  if (!a || actions.value.edit) return ''
  if (a.status === 'APPROVED') return t('icAssessment.detail.readOnlyApproved')
  if (a.status === 'SUBMITTED') return t('icAssessment.detail.readOnlySubmitted')
  return caps.value.canAssess ? '' : t('icAssessment.detail.readOnlyViewer')
})

const derived = computed(() => deriveAllComponents(draft.value?.principles ?? []))
const derivedConclusion = computed(() => (draft.value ? deriveConclusion(draft.value.components, draft.value.principles) : 'INCOMPLETE'))
const conclusionItems = computed(() => FINAL_CONCLUSIONS.map(c => ({ label: t(`icAssessment.conclusion.${c}`), value: c })))
const blockers = computed(() => (draft.value ? submitBlockers(draft.value) : []))

const conclusionError = (issue: ConclusionIssue): string | undefined => {
  if (!showErrors.value || !draft.value || !conclusionIssues(draft.value).includes(issue)) return undefined
  return issue === 'conclusionMissing'
    ? t('icAssessment.blockers.conclusion.conclusionMissing')
    : issue === 'effectiveNotAllowed'
      ? t('icAssessment.conclusionSection.effectiveNotAllowed')
      : t('icAssessment.conclusionSection.justificationRequired')
}

const operateTogetherText = computed(() => {
  const v = draft.value?.operateTogether
  if (v === true) return t('icAssessment.conclusionSection.operateTogetherYes')
  if (v === false) return t('icAssessment.conclusionSection.operateTogetherNo')
  return t('icAssessment.conclusionSection.operateTogetherPending')
})

const changedDriftSignals = computed(() => (store.drift?.signals ?? []).filter(s => s.changed === true))

const formatDate = (iso: string | null) => {
  if (!iso) return '-'
  const d = new Date(iso)
  return isNaN(d.getTime()) ? iso : d.toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { dateStyle: 'medium' })
}

const fail = () => {
  toast.error(t('icAssessment.toast.actionFailed'), getUserErrorMessage(store.lastError, t))
}

// ---- loading ----

const load = async () => {
  notFound.value = false
  showErrors.value = false
  const a = await store.fetchDetail(id.value)
  if (!a) {
    draft.value = null
    notFound.value = (store.lastError as { statusCode?: number } | null)?.statusCode === 404
    return
  }
  draft.value = clone(a)
  store.fetchPeriods(a.year)
  if (caps.value.fullRead) store.fetchEvents(a.id)
  if (a.status === 'APPROVED' && caps.value.fullRead) store.checkDrift(a.id)
}

// A write returned a new server state: the working copy follows it.
const adopt = (a: IcAssessment | null) => {
  if (!a) return false
  draft.value = clone(a)
  if (caps.value.fullRead) store.fetchEvents(a.id)
  if (a.status === 'APPROVED' && caps.value.fullRead) store.checkDrift(a.id)
  return true
}

onMounted(load)
watch(id, load)

// ---- editing ----

const onPrincipleUpdate = (code: string, patch: Partial<IcPrinciple>) => {
  const p = draft.value?.principles.find(x => x.code === code)
  if (p) Object.assign(p, patch)
}

const openEvidence = (code: ComponentCode) => {
  evidenceCode.value = code
  evidenceOpen.value = true
}

const doDiscard = async () => {
  if (!(await modal.confirmSubmit({ description: t('icAssessment.confirm.discard'), confirmLabel: t('icAssessment.actions.discard') }))) return
  if (original.value) draft.value = clone(original.value)
}

/** Saves pending edits. true = nothing to save or saved. */
const saveIfDirty = async (): Promise<boolean> => {
  if (!dirty.value || !original.value || !draft.value) return true
  const saved = await store.save(original.value, draft.value)
  if (!saved) {
    fail()
    return false
  }
  return adopt(saved)
}

const doSave = async () => {
  busy.value = 'save'
  const ok = await saveIfDirty()
  busy.value = ''
  if (ok) toast.success(t('icAssessment.toast.saved'))
}

const doPropose = async () => {
  if (!draft.value) return
  const ok = await modal.confirmSubmit({
    description: t('icAssessment.confirm.propose'),
    body: [t('icAssessment.confirm.proposeBody'), ...(dirty.value ? [t('icAssessment.confirm.saveFirstBody')] : [])],
    confirmLabel: draft.value.proposalRun > 0 ? t('icAssessment.actions.repropose') : t('icAssessment.actions.propose')
  })
  if (!ok) return
  busy.value = 'propose'
  try {
    if (!(await saveIfDirty()) || !store.current) return
    const res = await store.propose(store.current)
    if (!res) return fail()
    adopt(res)
    toast.success(t('icAssessment.toast.proposed'))
  } finally {
    busy.value = ''
  }
}

const doSubmit = async () => {
  if (!draft.value) return
  showErrors.value = true
  if (blockers.value.length) {
    toast.error(t('icAssessment.toast.fixFirst'))
    return
  }
  const ok = await modal.confirmSubmit({
    description: t('icAssessment.confirm.submit'),
    body: [t('icAssessment.confirm.submitBody')],
    confirmLabel: t('icAssessment.actions.submit')
  })
  if (!ok) return
  busy.value = 'submit'
  try {
    if (!(await saveIfDirty()) || !store.current) return
    const res = await store.submit(store.current)
    if (!res) return fail()
    adopt(res)
    showErrors.value = false
    toast.success(t('icAssessment.toast.submitted'))
  } finally {
    busy.value = ''
  }
}

const doApprove = async (value: Parameters<typeof store.approve>[1]) => {
  if (!store.current) return
  const res = await store.approve(store.current, value)
  if (!res) return fail()
  approveOpen.value = false
  adopt(res)
  toast.success(t('icAssessment.toast.approved'))
}

const doReason = async (text: string) => {
  if (!store.current || !reason.value) return
  const kind = reason.value
  const res = kind === 'reopen' ? await store.reopen(store.current, text) : await store.returnToDraft(store.current, text)
  if (!res) return fail()
  reason.value = null
  adopt(res)
  toast.success(kind === 'reopen' ? t('icAssessment.toast.reopened') : t('icAssessment.toast.returned'))
}

const doDelete = async () => {
  if (!draft.value) return
  if (!(await modal.confirmDelete({ description: t('icAssessment.confirm.delete') }))) return
  const ok = await store.remove(draft.value.id)
  if (!ok) return fail()
  toast.success(t('icAssessment.toast.deleted'))
  draft.value = null
  router.push('/internal-control-assessment')
}

// ---- unsaved-change guard ----

const confirmLeave = async () => {
  if (!dirty.value) return true
  return modal.confirmSubmit({
    description: t('icAssessment.confirm.leave'),
    confirmLabel: t('icAssessment.confirm.leaveConfirm')
  })
}

const goBack = async () => {
  if (await confirmLeave()) {
    // The route guard asks again unless the draft is dropped first.
    if (original.value) draft.value = clone(original.value)
    router.push('/internal-control-assessment')
  }
}

onBeforeRouteLeave(async () => (await confirmLeave()) || false)

const beforeUnload = (e: BeforeUnloadEvent) => {
  if (!dirty.value) return
  e.preventDefault()
  e.returnValue = ''
}
onMounted(() => window.addEventListener('beforeunload', beforeUnload))
onBeforeUnmount(() => window.removeEventListener('beforeunload', beforeUnload))
</script>
