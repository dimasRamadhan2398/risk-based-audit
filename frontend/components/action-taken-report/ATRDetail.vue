<template>
  <AtrModalShell
    v-model:open="store.showDetail"
    size="lg"
    :title="t('actionTakenReport.detail.title')"
    :subtitle="item?.report_number ? t('actionTakenReport.detail.subtitle', { number: item.report_number }) : ''"
  >
    <template
      v-if="item"
      #meta
    >
      <AtrStatusBadge
        :status="item.status"
        :is-overdue="item.is_overdue"
        :overdue-days="item.overdue_days"
        size="md"
      />
      <UIcon
        v-if="store.detailLoading"
        name="i-lucide-loader-2"
        class="w-4 h-4 animate-spin text-primary-500"
        :aria-label="t('actionTakenReport.table.loading')"
      />
    </template>

    <div
      v-if="!item"
      class="flex items-center justify-center gap-2 py-10 text-sm text-gray-500"
    >
      <UIcon
        name="i-lucide-loader-2"
        class="w-5 h-5 animate-spin"
      />
      {{ t('actionTakenReport.table.loading') }}
    </div>

    <template v-else>
      <AtrFindingSummary :item="item">
        <template #lha-actions>
          <UButton
            v-if="item.audit_result_report_id"
            :label="t('actionTakenReport.detail.showLhaFollowUps')"
            icon="i-lucide-list-filter"
            color="primary"
            variant="link"
            size="sm"
            class="px-0"
            @click="emit('filter-lha', item.audit_result_report_id)"
          />
        </template>
      </AtrFindingSummary>

      <!-- Assignment and progress -->
      <section class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4">
        <h4 class="text-base font-bold">
          {{ t('actionTakenReport.detail.followUpSection') }}
        </h4>
        <dl class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-sm">
          <div class="space-y-1">
            <dt class="font-semibold text-gray-600 dark:text-gray-300">
              {{ t('actionTakenReport.detail.pic') }}
            </dt>
            <dd>{{ item.pic_name || t('actionTakenReport.table.unassigned') }}</dd>
          </div>
          <div class="space-y-1">
            <dt class="font-semibold text-gray-600 dark:text-gray-300">
              {{ t('actionTakenReport.detail.dueDate') }}
            </dt>
            <dd :class="item.is_overdue ? 'text-error-600 dark:text-error-400 font-semibold' : ''">
              {{ formatAtrDate(item.due_date, locale) }}
            </dd>
          </div>
          <div class="space-y-1">
            <dt class="font-semibold text-gray-600 dark:text-gray-300">
              {{ t('actionTakenReport.detail.progress') }}
            </dt>
            <dd class="flex items-center gap-2">
              <UProgress
                :model-value="item.progress"
                :max="100"
                size="sm"
                class="flex-1"
              />
              <span class="tabular-nums font-semibold">{{ item.progress }}%</span>
            </dd>
          </div>
        </dl>
        <div class="space-y-1 text-sm">
          <p class="font-semibold text-gray-600 dark:text-gray-300">
            {{ t('actionTakenReport.detail.actionPlan') }}
          </p>
          <p
            v-if="item.action_plan"
            class="whitespace-pre-line break-words"
          >
            {{ item.action_plan }}
          </p>
          <p
            v-else
            class="text-gray-500 dark:text-gray-400"
          >
            {{ t('actionTakenReport.detail.noActionPlan') }}
          </p>
        </div>
      </section>

      <!-- Review note (latest decision) -->
      <section
        v-if="item.review_note || item.reviewed_at"
        class="p-4 sm:p-5 rounded-xl space-y-2 border"
        :class="reviewTone"
      >
        <h4 class="text-base font-bold">
          {{ t('actionTakenReport.detail.reviewSection') }}
        </h4>
        <p class="text-xs text-gray-600 dark:text-gray-300">
          {{ t('actionTakenReport.detail.reviewedMeta', { by: item.reviewed_by || '-', date: formatAtrDateTime(item.reviewed_at, locale) }) }}
        </p>
        <p class="text-sm whitespace-pre-line break-words">
          {{ item.review_note || '-' }}
        </p>
      </section>

      <!-- Status timeline -->
      <section class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-4">
        <h4 class="text-base font-bold">
          {{ t('actionTakenReport.timeline.title') }}
        </h4>
        <ol class="space-y-3">
          <li
            v-for="step in timeline"
            :key="step.key"
            class="flex items-start gap-3"
          >
            <span
              class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-full border"
              :class="step.state === 'done'
                ? 'bg-primary-500 border-primary-500 text-white'
                : step.state === 'current'
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400 ring-2 ring-primary-500/20'
                  : 'border-gray-300 dark:border-gray-600 text-gray-400'"
            >
              <UIcon
                :name="step.state === 'done' ? 'i-lucide-check' : step.icon"
                class="w-3.5 h-3.5"
              />
            </span>
            <div class="min-w-0">
              <p
                class="text-sm font-semibold"
                :class="step.state === 'todo' ? 'text-gray-400 dark:text-gray-500' : ''"
              >
                {{ step.label }}
              </p>
              <p
                v-if="step.detail"
                class="text-xs text-gray-500 dark:text-gray-400 break-words"
              >
                {{ step.detail }}
              </p>
            </div>
          </li>
        </ol>
      </section>

      <!-- Evidence -->
      <section class="p-4 sm:p-5 border border-gray-200 dark:border-gray-700 rounded-xl space-y-3">
        <h4 class="text-base font-bold">
          {{ t('actionTakenReport.evidence.title') }}
        </h4>
        <AtrEvidenceList
          :atr-id="item.id"
          :evidence="item.evidence"
        />
      </section>
    </template>

    <template #footer>
      <div class="flex flex-col-reverse sm:flex-row sm:flex-wrap sm:justify-end gap-2">
        <UButton
          :label="t('actionTakenReport.modal.close')"
          color="neutral"
          variant="ghost"
          class="justify-center"
          @click="store.closeModal()"
        />
        <template v-if="item">
          <UButton
            v-if="actions.cancel"
            :label="t('actionTakenReport.actions.cancel')"
            icon="i-lucide-ban"
            color="error"
            variant="outline"
            class="justify-center"
            @click="store.openCancel(item)"
          />
          <UButton
            v-if="actions.assign"
            :label="item.pic_user_id ? t('actionTakenReport.actions.reassign') : t('actionTakenReport.actions.assign')"
            icon="i-lucide-user-plus"
            color="neutral"
            variant="outline"
            class="justify-center"
            @click="store.openAssignment(item)"
          />
          <UButton
            v-if="actions.review"
            :label="t('actionTakenReport.actions.review')"
            icon="i-lucide-clipboard-check"
            color="primary"
            class="justify-center"
            @click="store.openReview(item)"
          />
          <UButton
            v-if="actions.actionPlan"
            :label="t('actionTakenReport.actions.fillActionPlan')"
            icon="i-lucide-pencil-line"
            color="primary"
            class="justify-center"
            @click="store.openActionPlan(item)"
          />
        </template>
      </div>
    </template>
  </AtrModalShell>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { useAtrActions } from '~/composables/useAtrActions'
import { formatAtrDate, formatAtrDateTime, normalizeAtrStatus } from '~/utils/actionTakenReport'
import AtrModalShell from '~/components/action-taken-report/AtrModalShell.vue'
import AtrStatusBadge from '~/components/action-taken-report/AtrStatusBadge.vue'
import AtrFindingSummary from '~/components/action-taken-report/AtrFindingSummary.vue'
import AtrEvidenceList from '~/components/action-taken-report/AtrEvidenceList.vue'

const emit = defineEmits<{ (e: 'filter-lha', lhaId: string): void }>()

const store = useActionTakenReportStore()
const { t, locale } = useI18n()
const { availableActions } = useAtrActions()

const item = computed(() => store.current)
const actions = computed(() => availableActions(item.value))

const reviewTone = computed(() => normalizeAtrStatus(item.value?.status) === 'COMPLETED'
  ? 'border-success-200 dark:border-success-900/50 bg-success-50/60 dark:bg-success-950/20'
  : 'border-warning-200 dark:border-warning-900/50 bg-warning-50/60 dark:bg-warning-950/20')

type StepState = 'done' | 'current' | 'todo'
interface TimelineStep { key: string, label: string, detail: string, icon: string, state: StepState }

// No history endpoint exists yet: the timeline is derived from the current status and the item's dates.
const timeline = computed<TimelineStep[]>(() => {
  const it = item.value
  if (!it) return []
  const status = normalizeAtrStatus(it.status)
  const order = ['PLANNED', 'IN_PROGRESS', 'PENDING_REVIEW', 'COMPLETED']
  const reached = (target: string) => order.indexOf(status) >= order.indexOf(target)
  const stateFor = (target: string): StepState => {
    if (status === target) return target === 'COMPLETED' ? 'done' : 'current'
    return reached(target) ? 'done' : 'todo'
  }
  const rejected = status === 'IN_PROGRESS' && !!it.review_note

  const steps: TimelineStep[] = [
    {
      key: 'created',
      label: t('actionTakenReport.timeline.created'),
      detail: formatAtrDateTime(it.created_at, locale.value),
      icon: 'i-lucide-file-plus',
      state: 'done'
    },
    {
      key: 'assigned',
      label: t('actionTakenReport.timeline.assigned'),
      detail: it.pic_user_id
        ? t('actionTakenReport.timeline.assignedDetail', { pic: it.pic_name || '-', date: formatAtrDate(it.due_date, locale.value) })
        : t('actionTakenReport.timeline.notAssigned'),
      icon: 'i-lucide-user-plus',
      state: it.pic_user_id ? 'done' : (status === 'CANCELLED' ? 'todo' : 'current')
    },
    {
      key: 'inProgress',
      label: t('actionTakenReport.status.inProgress'),
      detail: rejected ? t('actionTakenReport.timeline.returned') : '',
      icon: 'i-lucide-loader',
      state: status === 'CANCELLED' ? (it.action_plan ? 'done' : 'todo') : stateFor('IN_PROGRESS')
    },
    {
      key: 'pendingReview',
      label: t('actionTakenReport.status.pendingReview'),
      detail: '',
      icon: 'i-lucide-hourglass',
      state: status === 'CANCELLED' ? 'todo' : stateFor('PENDING_REVIEW')
    }
  ]

  if (status === 'CANCELLED') {
    steps.push({
      key: 'cancelled',
      label: t('actionTakenReport.status.cancelled'),
      detail: formatAtrDateTime(it.reviewed_at || it.updated_at, locale.value),
      icon: 'i-lucide-ban',
      state: 'done'
    })
  } else {
    steps.push({
      key: 'completed',
      label: t('actionTakenReport.status.completed'),
      detail: status === 'COMPLETED' ? formatAtrDateTime(it.reviewed_at || it.updated_at, locale.value) : '',
      icon: 'i-lucide-flag',
      state: stateFor('COMPLETED')
    })
  }

  // PLANNED with a PIC: "In Progress" is the next step to reach.
  if (status === 'PLANNED' && it.pic_user_id) {
    const step = steps.find(s => s.key === 'inProgress')
    if (step) step.state = 'current'
  }
  return steps
})
</script>
