<template>
  <UModal
    :open="store.isViewModalOpen"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-6xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeViewModal() }"
  >
    <template #content>
      <div class="flex max-h-[90vh] flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between gap-3 border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4">
          <div class="min-w-0">
            <h3 class="text-xl font-bold text-[var(--text-main)]">
              {{ t('auditActivityPlan.view.title') }}
            </h3>
            <p
              v-if="plan"
              class="truncate text-xs text-[var(--text-muted)]"
            >
              {{ plan.title }}
            </p>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            :aria-label="t('auditActivityPlan.view.close')"
            @click="store.closeViewModal"
          />
        </div>

        <div class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
          <!-- Loading -->
          <div
            v-if="store.viewLoading"
            class="flex h-32 items-center justify-center gap-2 text-[var(--text-muted)]"
          >
            <UIcon
              name="i-heroicons-arrow-path"
              class="h-5 w-5 animate-spin"
            />
            {{ t('auditActivityPlan.view.loading') }}
          </div>

          <!-- Error -->
          <UAlert
            v-else-if="store.viewError"
            color="error"
            variant="subtle"
            icon="i-heroicons-exclamation-triangle"
            :description="store.viewError"
            :actions="[{ label: t('auditActivityPlan.retry'), color: 'error', variant: 'outline', onClick: () => store.loadView() }]"
          />

          <template v-else-if="plan">
            <!-- Audit Activity Plan form -->
            <template v-if="!plan.isLegacy">
              <div class="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <p class="text-xs text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.header.annualPlan') }}
                  </p>
                  <p class="font-semibold">
                    {{ plan.annualPlan ? [plan.annualPlan.code, plan.annualPlan.title].filter(Boolean).join(' · ') : '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.table.status') }}
                  </p>
                  <p class="font-semibold">
                    {{ formatActivityStatus(plan.status) || '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.view.createdByLabel') }}
                  </p>
                  <p class="font-semibold">
                    {{ plan.createdBy || '-' }}
                  </p>
                </div>
                <div>
                  <p class="text-xs text-[var(--text-muted)]">
                    {{ t('auditActivityPlan.view.updatedAt') }}
                  </p>
                  <p class="font-semibold">
                    {{ plan.updatedAt ? formatDateLocale(plan.updatedAt, locale) : '-' }}
                  </p>
                </div>
              </div>

              <UAlert
                v-if="plan.annualPlan && plan.annualPlan.status !== AnnualAuditPlanStatus.APPROVED"
                color="warning"
                variant="subtle"
                icon="i-heroicons-information-circle"
                :description="t('auditActivityPlan.header.annualPlanNotApproved', { status: plan.annualPlan.status || '-' })"
              />
              <UAlert
                v-if="plan.rows.some(r => r.source === 'snapshot')"
                color="warning"
                variant="subtle"
                icon="i-heroicons-exclamation-triangle"
                :description="t('auditActivityPlan.rows.pkatMissing')"
                data-testid="snapshot-warning"
              />

              <ActivityPlanFormSheet :sheet="sheetFromDetail(plan)" />
            </template>

            <!-- Legacy plan: stored data, read-only -->
            <template v-else>
              <UAlert
                color="warning"
                variant="subtle"
                icon="i-heroicons-archive-box"
                :title="t('auditActivityPlan.legacy.badge')"
                :description="t('auditActivityPlan.legacy.notice')"
                data-testid="legacy-notice"
              />

              <UCard v-if="plan.legacy">
                <template #header>
                  <h4 class="text-lg font-medium">
                    {{ t('auditActivityPlan.legacy.section') }}
                  </h4>
                </template>
                <div class="space-y-6 text-sm">
                  <div class="space-y-2">
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planTitle') }}</strong>
                      <span>{{ plan.legacy.planTitle || plan.title || '-' }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planYear') }}</strong>
                      <span>{{ plan.legacy.planYear || '-' }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.planPeriod') }}</strong>
                      <span>{{ formatPeriod(plan.legacy.planPeriodStart, plan.legacy.planPeriodEnd, locale) || '-' }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.department') }}</strong>
                      <span>{{ plan.legacy.department || '-' }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.createdBy') }}</strong>
                      <span>{{ plan.createdBy || '-' }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.creationDate') }}</strong>
                      <span>{{ plan.legacy.creationDate ? formatDateLocale(plan.legacy.creationDate, locale) : '-' }}</span>
                    </div>
                  </div>

                  <div>
                    <h5 class="mb-2 font-semibold">
                      {{ t('auditActivityPlan.view.plannedActivities', { count: plan.legacy.plannedActivities.length }) }}
                    </h5>
                    <UTable
                      :data="plan.legacy.plannedActivities"
                      :columns="plannedActivitiesColumns"
                    />
                  </div>

                  <div>
                    <h5 class="mb-2 font-semibold">
                      {{ t('auditActivityPlan.view.resourceAuditors', { count: plan.legacy.resourceAuditors.length }) }}
                    </h5>
                    <UTable
                      :data="plan.legacy.resourceAuditors"
                      :columns="resourceAuditorsColumns"
                    />
                  </div>

                  <div class="space-y-2">
                    <h5 class="font-semibold">
                      {{ t('auditActivityPlan.view.budgetPlanning') }}
                    </h5>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.totalEstimatedCost') }}</strong>
                      <span>{{ formatRupiah(plan.legacy.budget.totalEstimatedCost) }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.totalAllocatedBudget') }}</strong>
                      <span>{{ formatRupiah(plan.legacy.budget.totalAllocatedBudget) }}</span>
                    </div>
                    <div class="flex">
                      <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.budgetNotes') }}</strong>
                      <span class="flex-1">{{ plan.legacy.budget.budgetNotes || '-' }}</span>
                    </div>
                  </div>

                  <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                    <div class="space-y-2">
                      <p class="mb-2 border-b pb-1 font-semibold">
                        {{ t('auditActivityPlan.view.creator') }}
                      </p>
                      <div class="flex">
                        <strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.name') }}</strong>
                        <span>{{ plan.legacy.review.creatorName || '-' }}</span>
                      </div>
                      <div class="flex">
                        <strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.position') }}</strong>
                        <span>{{ plan.legacy.review.creatorPosition || '-' }}</span>
                      </div>
                    </div>
                    <div class="space-y-2">
                      <p class="mb-2 border-b pb-1 font-semibold">
                        {{ t('auditActivityPlan.view.approver') }}
                      </p>
                      <div class="flex">
                        <strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.name') }}</strong>
                        <span>{{ plan.legacy.review.approverName || '-' }}</span>
                      </div>
                      <div class="flex">
                        <strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.position') }}</strong>
                        <span>{{ plan.legacy.review.approverPosition || '-' }}</span>
                      </div>
                      <div class="flex">
                        <strong class="w-24 shrink-0">{{ t('auditActivityPlan.view.date') }}</strong>
                        <span>{{ plan.legacy.review.approvalDate ? formatDateLocale(plan.legacy.review.approvalDate, locale) : '-' }}</span>
                      </div>
                    </div>
                    <div class="md:col-span-2">
                      <strong class="font-semibold">{{ t('auditActivityPlan.view.additionalNotes') }}</strong>
                      <p class="mt-1 text-gray-600 dark:text-gray-300">
                        {{ plan.legacy.review.additionalNotes || '-' }}
                      </p>
                    </div>
                  </div>
                </div>
              </UCard>
            </template>

            <!-- Attachments (both formats) -->
            <UCard v-if="plan.attachmentCategory || plan.attachments.length">
              <template #header>
                <h4 class="text-lg font-medium">
                  {{ t('auditActivityPlan.view.attachment') }}
                </h4>
              </template>
              <div class="space-y-3 text-sm">
                <div class="flex">
                  <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.attachmentCategory') }}</strong>
                  <span>{{ plan.attachmentCategory || '-' }}</span>
                </div>
                <div class="flex">
                  <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.uploadedBy') }}</strong>
                  <span>{{ plan.attachmentUploadedBy || '-' }}</span>
                </div>
                <div class="flex">
                  <strong class="w-48 shrink-0">{{ t('auditActivityPlan.view.uploadDate') }}</strong>
                  <span>{{ plan.attachmentUploadDate || '-' }}</span>
                </div>
                <div class="pt-2">
                  <strong class="mb-2 block">{{ t('auditActivityPlan.view.uploadedFiles') }}</strong>
                  <ul
                    v-if="plan.attachments.length"
                    class="space-y-2"
                  >
                    <li
                      v-for="(file, index) in plan.attachments"
                      :key="`${file.url}-${index}`"
                      class="flex items-center justify-between gap-2 rounded-md bg-gray-50 dark:bg-gray-800/50 p-2"
                    >
                      <div class="flex min-w-0 items-center gap-2">
                        <UIcon
                          name="i-heroicons-document-text"
                          class="shrink-0 text-gray-500"
                        />
                        <span class="truncate font-semibold">{{ file.name }}</span>
                        <UBadge
                          color="neutral"
                          variant="soft"
                          size="md"
                        >
                          {{ file.size }}
                        </UBadge>
                      </div>
                      <UButton v-if="canAccessFiles"
                        icon="i-heroicons-arrow-down-tray"
                        size="sm"
                        color="primary"
                        variant="link"
                        :label="t('auditActivityPlan.view.download')"
                        @click="store.openAttachment(file)"
                      />
                    </li>
                  </ul>
                  <p
                    v-else
                    class="mt-2 italic text-gray-500"
                  >
                    {{ t('auditActivityPlan.view.noFilesAttached') }}
                  </p>
                </div>
              </div>
            </UCard>
          </template>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4">
          <UButton
            :label="t('auditActivityPlan.view.close')"
            color="neutral"
            variant="ghost"
            @click="store.closeViewModal"
          />
          <UButton
            v-if="canWrite && (plan && !plan.isLegacy)"
            icon="i-lucide-edit"
            :label="t('auditActivityPlan.editPlan')"
            color="warning"
            variant="soft"
            data-testid="view-edit"
            @click="editFromView"
          />
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useRbac } from '~/composables/useRbac'
import { computed, h } from 'vue'
import { useActivityPlanStore } from '~/stores/activity-plan'
import { AnnualAuditPlanStatus, type LegacyPlannedAuditActivity } from '~/types/audit'
import { useI18n } from '~/composables/useI18n'
import { formatPeriod, formatDateLocale } from '~/utils/dateConverter'
import { formatActivityStatus, sheetFromDetail } from '~/utils/activityPlanForm'
import ActivityPlanFormSheet from '~/components/audit-activity-plan/ActivityPlanFormSheet.vue'

const { t, locale } = useI18n()
const { canWrite, canAccessFiles } = useRbac()
const store = useActivityPlanStore()

const plan = computed(() => store.selectedPlan)

const formatRupiah = (value: number) =>
  new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(Number(value) || 0)

const editFromView = () => {
  const current = plan.value
  if (!current) return
  store.closeViewModal()
  store.handleEdit(current)
}

const plannedActivitiesColumns = computed(() => [
  {
    accessorKey: 'activityCode',
    header: t('auditActivityPlan.view.columns.activityId'),
    cell: ({ row }: { row: { original: LegacyPlannedAuditActivity } }) => row.original.activityCode
      ? h('span', { class: 'font-mono text-xs font-semibold whitespace-nowrap text-gray-900 dark:text-gray-100' }, row.original.activityCode)
      : h('span', { class: 'text-gray-400 dark:text-gray-500' }, '-')
  },
  { accessorKey: 'auditName', header: t('auditActivityPlan.view.columns.auditName') },
  { accessorKey: 'auditee', header: t('auditActivityPlan.view.columns.auditee') },
  { accessorKey: 'category', header: t('auditActivityPlan.view.columns.category') },
  { accessorKey: 'riskName', header: t('auditActivityPlan.view.columns.riskName') },
  { accessorKey: 'riskLevel', header: t('auditActivityPlan.view.columns.riskLevel') },
  { accessorKey: 'duration', header: t('auditActivityPlan.view.columns.duration') },
  { accessorKey: 'priority', header: t('auditActivityPlan.view.columns.priority') }
])

const resourceAuditorsColumns = computed(() => [
  { accessorKey: 'name', header: t('auditActivityPlan.view.columns.name') },
  { accessorKey: 'position', header: t('auditActivityPlan.view.columns.position') },
  { accessorKey: 'competence', header: t('auditActivityPlan.view.columns.competence') },
  { accessorKey: 'availability', header: t('auditActivityPlan.view.columns.availability') }
])
</script>
