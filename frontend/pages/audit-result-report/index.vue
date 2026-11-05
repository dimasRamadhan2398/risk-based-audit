<template>
  <div class="p-4 sm:p-6 min-w-0">
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('auditResultReport.dashboard.title') }}</h1>
        <p class="text-gray-500 dark:text-gray-400">{{ t('auditResultReport.dashboard.subtitle') }}</p>
      </div>
      <div class="flex flex-wrap items-center gap-2 w-full sm:w-auto">
        <UButton
          v-if="canImportPlanDocs"
          color="neutral"
          variant="outline"
          icon="i-lucide-upload"
          :label="t('auditResultReport.dashboard.importDocument')"
          to="/audit-result-report/upload"
          class="w-full sm:w-auto font-bold shadow"
        />
      </div>
    </div>

    <!-- Assignment Letter Selector -->
    <UCard class="mb-6" :ui="{ body: 'p-4' }">
      <div class="flex flex-col md:flex-row md:items-center gap-4">
        <div class="flex-1">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            {{ t('auditResultReport.dashboard.assignment.label') }}
          </label>
          <USelectMenu
            v-model="store.selectedAssignmentLetter"
            :items="store.publishedAssignmentLetters"
            :placeholder="t('auditResultReport.dashboard.assignment.placeholder')"
            class="w-full sm:max-w-md"
            :disabled="store.publishedAssignmentLetters.length === 0"
          >
            <template #leading>
              <UIcon name="i-heroicons-document-text" class="size-5" />
            </template>
          </USelectMenu>
        </div>
        <div v-if="store.publishedAssignmentLetters.length === 0" class="text-sm text-amber-600">
          <UIcon name="i-heroicons-exclamation-triangle" class="size-4 inline mr-1" />
          {{ t('auditResultReport.dashboard.assignment.nonePublished') }}
        </div>
        <div v-else-if="store.selectedAssignmentLetter" class="text-sm text-green-600 flex items-center gap-2">
          <UIcon name="i-heroicons-check-circle" class="size-4 inline" />
          <span>{{ t('auditResultReport.dashboard.assignment.selected', { number: store.selectedAssignmentLetter }) }}</span>
        </div>
        <div v-else class="text-sm text-gray-600">
          <UIcon name="i-heroicons-document-text" class="size-4 inline mr-1" />
          {{ t('auditResultReport.dashboard.assignment.noneSelected') }}
        </div>
      </div>
    </UCard>

    <!-- Main Content -->
    <div v-if="store.hasSelectedAssignmentLetter">
      <!-- Loading State Skeleton -->
      <div v-if="store.loading && store.filteredReports.length === 0" class="space-y-4">
        <div class="hidden md:block p-12 text-center bg-transparent rounded-xl border border-gray-200 dark:border-gray-800">
          <UIcon name="i-lucide-loader-2" class="size-8 mx-auto animate-spin text-primary-500 mb-2" />
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</p>
        </div>
        <div class="block md:hidden space-y-3">
          <div
            v-for="i in 2"
            :key="i"
            class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] animate-pulse space-y-3"
          >
            <div class="flex justify-between items-center">
              <div class="h-5 w-32 bg-gray-200 dark:bg-gray-700 rounded"></div>
              <div class="h-5 w-16 bg-gray-200 dark:bg-gray-700 rounded-full"></div>
            </div>
            <div class="h-4 w-4/5 bg-gray-200 dark:bg-gray-700 rounded"></div>
            <div class="h-3 w-24 bg-gray-200 dark:bg-gray-700 rounded"></div>
            <div class="grid grid-cols-2 gap-2 pt-2">
              <div v-for="j in 4" :key="j" class="h-16 bg-gray-100 dark:bg-gray-800 rounded-lg"></div>
            </div>
            <div class="flex justify-between items-center pt-2">
              <div class="h-7 w-28 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
              <div class="flex gap-1">
                <div class="h-7 w-7 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
                <div class="h-7 w-7 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="store.filteredReports.length > 0" class="space-y-4">
        <!-- Desktop View (>= md): Full Table with Sticky Columns -->
        <div class="hidden md:block bg-transparent rounded-xl border border-gray-200 dark:border-gray-800 overflow-x-auto shadow-sm">
          <UTable :data="store.filteredReports" :columns="columns" :loading="store.loading" :ui="{ root: 'overflow-visible [&>table]:border-separate [&>table]:border-spacing-0', th: 'bg-gray-100 dark:bg-gray-800/50' }">
            <template #reportNumber-cell="{ row }">
              <span class="font-mono text-md font-semibold text-primary-600 dark:text-primary-400">
                {{ row.original.reportNumber || (row.original as any).report_number || '-' }}
              </span>
            </template>
            <template #reportDate-cell="{ row }">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ row.original.reportDate || (row.original as any).report_date?.split('T')[0] || '-' }}
              </span>
            </template>
            <template #status-cell="{ row }">
              <UBadge :color="row.original.status === 'Final' ? 'success' : 'neutral'" variant="soft">
                {{ getStatusLabel(row.original.status) }}
              </UBadge>
            </template>
            <template #findingsCount-cell="{ row }">
              <div class="flex flex-col gap-2 min-w-[30px]">
                <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center bg-gray-50 dark:bg-gray-800 rounded">
                  <span class="text-md font-semibold">{{ group.count }}</span>
                </div>
              </div>
            </template>
            <template #category-cell="{ row }">
              <div class="flex flex-col gap-2">
                <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center">
                  <UBadge :color="getRatingColor(group.category)" variant="subtle" size="sm" class="w-full justify-center">
                    {{ group.category }}
                  </UBadge>
                </div>
              </div>
            </template>
            <template #listOfFinding-cell="{ row }">
              <div class="flex flex-col gap-2">
                <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center">
                  <UPopover v-if="group.count > 0" mode="hover">
                    <UButton color="neutral" variant="soft" size="md" trailing-icon="i-heroicons-chevron-down" class="text-md">
                      {{ t('auditResultReport.dashboard.findings.viewCount', { count: group.count }) }}
                    </UButton>
                    <template #content>
                      <div class="p-3 max-w-sm max-h-60 overflow-y-auto">
                        <h4 class="text-md font-bold text-gray-900 dark:text-white mb-2 border-b pb-1">{{ t('auditResultReport.dashboard.findings.title', { category: getCategoryLabel(group.category) }) }}</h4>
                        <ul class="list-disc pl-4 text-md text-gray-700 dark:text-gray-300 space-y-1">
                          <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">
                            {{ item.title }}
                            <span v-if="item.source" class="text-[10px] text-gray-400 block italic">({{ item.source }})</span>
                          </li>
                        </ul>
                      </div>
                    </template>
                  </UPopover>
                  <span v-else class="text-md text-gray-400">-</span>
                </div>
              </div>
            </template>
            <template #action-cell="{ row }">
              <div class="flex flex-col gap-2">
                <div v-for="group in getGroupedFindings(row.original.findings)" :key="group.category" class="h-6 flex items-center justify-center">
                  <UPopover v-if="group.count > 0" mode="hover">
                    <UButton color="neutral" variant="soft" size="md" trailing-icon="i-heroicons-chevron-down" class="text-md">
                      {{ t('auditResultReport.dashboard.findings.viewActions') }}
                    </UButton>
                    <template #content>
                      <div class="p-3 max-w-sm max-h-60 overflow-y-auto">
                        <h4 class="text-md font-bold text-gray-900 dark:text-white mb-2 border-b pb-1">{{ t('auditResultReport.dashboard.findings.actionsTitle', { category: getCategoryLabel(group.category) }) }}</h4>
                        <ul class="list-disc pl-4 text-md text-gray-700 dark:text-gray-300 space-y-1">
                          <li v-for="(item, idx) in group.items" :key="idx" class="leading-relaxed">
                            <span class="font-semibold block mb-0.5 text-gray-800 dark:text-gray-200">{{ item.title }}:</span>
                            <span class="text-primary-700 dark:text-primary-400 block mb-1">{{ item.action || t('auditResultReport.dashboard.findings.noAction') }}</span>
                          </li>
                        </ul>
                      </div>
                    </template>
                  </UPopover>
                  <span v-else class="text-md text-gray-400">-</span>
                </div>
              </div>
            </template>
            <template #actions-cell="{ row }">
              <div class="flex gap-2 items-center">
                <!-- Approved LHA: its findings have ATRs (created by the backend on approval) -->
                <UTooltip
                  v-if="isLhaApproved(row.original.status)"
                  :text="t('auditResultReport.followUp.tooltip')"
                >
                  <UButton
                    color="primary"
                    variant="soft"
                    icon="i-lucide-list-checks"
                    size="sm"
                    :to="`/action-taken-report?lha=${encodeURIComponent((row.original as any).id)}`"
                    :label="followUpLabel((row.original as any).id)"
                  />
                </UTooltip>
                <UTooltip
                  v-if="!isLhaApproved(row.original.status)"
                  :text="t('auditResultReport.dashboard.actions.syncTooltip')"
                >
                  <UButton
                    color="primary"
                    variant="soft"
                    icon="i-lucide-refresh-cw"
                    size="sm"
                    :label="t('auditResultReport.dashboard.actions.sync')"
                    @click="syncReportFindings(row.original as any)"
                  />
                </UTooltip>
                <UTooltip v-if="canExport" :text="t('auditResultReport.dashboard.actions.generateDocx')">
                  <UButton
                    color="success"
                    variant="soft"
                    icon="i-heroicons-arrow-down-tray"
                    size="sm"
                    label="Docx"
                    @click="store.downloadDocx((row.original as any).id, (row.original as any).reportNumber)"
                  />
                </UTooltip>
                <!-- Edit button: only displayed when report status is draft -->
                <UTooltip
                  v-if="!isLhaApproved(row.original.status)"
                  :text="t('auditResultReport.editLha') || t('common.edit')"
                >
                  <UButton
                    color="warning"
                    variant="ghost"
                    icon="i-lucide-edit"
                    size="md"
                    @click="store.editReport(row.original as any)"
                  />
                </UTooltip>
                <!-- View button: eye icon (when final, form cannot be edited and only view icon is displayed) -->
                <UTooltip :text="t('auditResultReport.viewLhaDetail')">
                  <UButton
                    color="info"
                    variant="ghost"
                    icon="i-lucide-eye"
                    size="md"
                    @click="viewReportDetail(row.original as any)"
                  />
                </UTooltip>
              </div>
            </template>
          </UTable>
        </div>

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden space-y-3">
          <div
            v-for="report in store.filteredReports"
            :key="(report as any).id"
            class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
          >
            <!-- Card Header: Report Number & Status Badge -->
            <div class="flex items-center justify-between gap-2">
              <span class="inline-flex items-center gap-1.5 font-mono text-xs font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2.5 py-1 rounded-md border border-primary-200/50 dark:border-primary-800/40 tabular-nums">
                <UIcon name="i-lucide-file-text" class="size-3.5" />
                {{ report.reportNumber || (report as any).report_number || '-' }}
              </span>
              <UBadge
                :color="report.status === 'Final' ? 'success' : 'neutral'"
                variant="subtle"
                size="sm"
                class="font-semibold"
              >
                {{ getStatusLabel(report.status) }}
              </UBadge>
            </div>

            <!-- Card Body: Report Title & Date -->
            <div class="space-y-1.5">
              <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
                {{ report.reportTitle || (report as any).report_title || '-' }}
              </h4>
              <div class="flex items-center gap-1.5 text-xs text-[var(--text-muted)]">
                <UIcon name="i-lucide-calendar" class="size-3.5 text-gray-400 shrink-0" />
                <span class="font-medium tabular-nums">
                  {{ report.reportDate || (report as any).report_date?.split('T')[0] || '-' }}
                </span>
              </div>
            </div>

            <!-- Findings Summary (Vertical List) -->
            <div class="space-y-2 pt-2 border-t border-gray-100 dark:border-gray-800/80">
              <div class="flex items-center justify-between text-xs text-[var(--text-muted)]">
                <span class="font-semibold uppercase tracking-wider text-[11px]">{{ t('auditResultReport.dashboard.columns.findings') }}</span>
                <span class="text-[11px] font-medium text-gray-500 dark:text-gray-400 tabular-nums">
                  {{ (report.findings || []).length }} {{ t('auditResultReport.dashboard.columns.findings').toLowerCase() }}
                </span>
              </div>

              <div class="space-y-1.5">
                <div
                  v-for="group in getGroupedFindings(report.findings)"
                  :key="group.category"
                  class="flex items-center justify-between gap-2 p-2 rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40 text-xs transition-colors"
                >
                  <!-- Category Badge & Count -->
                  <div class="flex items-center gap-2 min-w-0">
                    <UBadge
                      :color="getRatingColor(group.category)"
                      variant="subtle"
                      size="md"
                      class="font-semibold whitespace-nowrap"
                    >
                      {{ getCategoryLabel(group.category) }}
                    </UBadge>
                    <span
                      class="font-bold tabular-nums text-sm"
                      :class="group.count > 0 ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500'"
                    >
                      {{ group.count }}
                    </span>
                  </div>

                  <!-- Popover Detail (Findings & Actions) if count > 0 -->
                  <div v-if="group.count > 0">
                    <UPopover mode="click">
                      <UButton
                        color="neutral"
                        variant="soft"
                        size="xs"
                        class="text-[11px] px-2.5 py-1 font-medium"
                      >
                        <span>{{ t('auditResultReport.dashboard.findings.viewCount', { count: group.count }) }}</span>
                        <UIcon name="i-heroicons-chevron-down" class="size-3 shrink-0 ml-1" />
                      </UButton>
                      <template #content>
                        <div class="p-3.5 max-w-xs sm:max-w-sm max-h-72 overflow-y-auto space-y-3">
                          <div class="flex items-center justify-between border-b pb-1.5 dark:border-gray-800">
                            <h4 class="text-xs font-bold text-gray-900 dark:text-white">
                              {{ t('auditResultReport.dashboard.findings.title', { category: getCategoryLabel(group.category) }) }}
                            </h4>
                            <UBadge :color="getRatingColor(group.category)" variant="subtle" size="md">
                              {{ group.count }}
                            </UBadge>
                          </div>
                          <ul class="space-y-2 text-xs">
                            <li
                              v-for="(item, idx) in group.items"
                              :key="idx"
                              class="p-2.5 rounded-lg bg-gray-50 dark:bg-gray-800/60 border border-gray-100 dark:border-gray-800 space-y-1.5"
                            >
                              <div class="font-semibold text-gray-900 dark:text-white leading-snug">
                                {{ item.title }}
                              </div>
                              <div v-if="item.action" class="text-primary-700 dark:text-primary-400 text-xs">
                                <span class="font-medium text-gray-500 dark:text-gray-400 block mb-0.5">{{ t('auditResultReport.dashboard.columns.action') }}:</span>
                                {{ item.action }}
                              </div>
                              <div v-else class="text-xs text-gray-400 italic">
                                {{ t('auditResultReport.dashboard.findings.noAction') }}
                              </div>
                              <div v-if="item.source" class="text-[10px] text-gray-400 italic">
                                ({{ item.source }})
                              </div>
                            </li>
                          </ul>
                        </div>
                      </template>
                    </UPopover>
                  </div>
                  <div v-else class="text-[11px] text-gray-400 px-2 font-medium">
                    -
                  </div>
                </div>
              </div>
            </div>

            <!-- Card Footer: Actions -->
            <div class="flex flex-wrap items-center justify-between gap-2 pt-3 border-t border-gray-100 dark:border-gray-800/80">
              <div class="flex flex-wrap items-center gap-1.5">
                <!-- Approved LHA: Follow up button -->
                <UButton
                  v-if="isLhaApproved(report.status)"
                  color="primary"
                  variant="soft"
                  icon="i-lucide-list-checks"
                  size="sm"
                  :to="`/action-taken-report?lha=${encodeURIComponent((report as any).id)}`"
                  :label="followUpLabel((report as any).id)"
                />
                <!-- Sync button if draft -->
                <UButton
                  v-if="!isLhaApproved(report.status)"
                  color="primary"
                  variant="soft"
                  icon="i-lucide-refresh-cw"
                  size="sm"
                  :label="t('auditResultReport.dashboard.actions.sync')"
                  @click="syncReportFindings(report as any)"
                />
                <!-- Docx export -->
                <UButton
                  v-if="canExport"
                  color="success"
                  variant="soft"
                  icon="i-heroicons-arrow-down-tray"
                  size="sm"
                  label="Docx"
                  @click="store.downloadDocx((report as any).id, (report as any).reportNumber)"
                />
              </div>

              <div class="flex items-center gap-1 ml-auto">
                <!-- Edit button if draft -->
                <UButton
                  v-if="!isLhaApproved(report.status)"
                  color="warning"
                  variant="ghost"
                  icon="i-lucide-edit"
                  size="md"
                  :title="t('auditResultReport.editLha') || t('common.edit')"
                  :aria-label="t('auditResultReport.editLha') || t('common.edit')"
                  @click="store.editReport(report as any)"
                />
                <!-- View button -->
                <UButton
                  color="info"
                  variant="ghost"
                  icon="i-lucide-eye"
                  size="md"
                  :title="t('auditResultReport.viewLhaDetail')"
                  :aria-label="t('auditResultReport.viewLhaDetail')"
                  @click="viewReportDetail(report as any)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="text-center py-16 bg-secondary-50/40 dark:bg-secondary-950/30 rounded-xl border-2 border-dashed border-secondary-200 dark:border-secondary-900/50">
        <h3 class="text-lg font-semibold text-gray-800 dark:text-gray-100">{{ t('auditResultReport.dashboard.empty.title') }}</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-2 max-w-md mx-auto mb-6 text-sm">
          {{ t('auditResultReport.dashboard.empty.description') }}
        </p>
        <div class="flex justify-center gap-3">
          <UButton v-if="canWrite"
            color="secondary"
            icon="i-lucide-plus"
            :label="t('auditResultReport.dashboard.empty.create')"
            size="lg"
            class="font-bold shadow"
            @click="store.openModal"
          />
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-16">
      <UIcon name="i-heroicons-document-magnifying-glass" class="size-20 text-gray-200 dark:text-gray-800 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-300 text-center">{{ t('auditResultReport.dashboard.selectAssignment.title') }}</h3>
      <p class="text-gray-500 mt-2 max-w-md mx-auto text-center">
        {{ t('auditResultReport.dashboard.selectAssignment.description') }}
      </p>
    </div>

    <!-- Report Form Modal -->
    <ResultReportForm />

    <!-- Detail LHA Modal (View only, without signature section) -->
    <ResultReportDetailModal
      v-model:open="showDetailModal"
      :report="selectedDetailReport"
      @print="handlePrintFromDetail"
    />

    <!-- Official LHA Document Print & Preview Modal 
    <ResultReportPrintModal
      v-model:open="showPrintModal"
      :report="selectedPrintReport"
    /> -->
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useI18n } from '~/composables/useI18n'
import { isLhaApproved } from '~/utils/actionTakenReport'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import ResultReportForm from '~/components/audit-result-report/ResultReportForm.vue'
import ResultReportDetailModal from '~/components/audit-result-report/ResultReportDetailModal.vue'
import ResultReportPrintModal from '~/components/audit-result-report/ResultReportPrintModal.vue'
import { useRbac } from '~/composables/useRbac'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'

const store = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()
const { canImportPlanDocs, canWrite, canExport } = useRbac()
const toast = useToastNotification()
const atrStore = useActionTakenReportStore()
const { t } = useI18n()

// ATR count per approved LHA for the "View follow-ups (N)" button (null = could not be loaded).
const atrCounts = ref<Record<string, number | null>>({})
watch(
  () => store.filteredReports.filter(r => isLhaApproved(r.status)).map(r => r.id),
  async (ids) => {
    for (const id of ids) {
      if (!id || id in atrCounts.value) continue
      atrCounts.value[id] = null
      atrCounts.value[id] = await atrStore.countForLha(id)
    }
  },
  { immediate: true }
)
const followUpLabel = (id: string) => {
  const count = atrCounts.value[id]
  return typeof count === 'number' ? t('auditResultReport.followUp.view', { count }) : t('auditResultReport.followUp.viewNoCount')
}

const showDetailModal = ref(false)
const selectedDetailReport = ref<any>(null)
const showPrintModal = ref(false)
const selectedPrintReport = ref<any>(null)

onMounted(() => {
  assignmentLetterStore.fetchAssignmentLetters()
  store.fetchReports()
})

const columns = computed(() => [
  // First two columns are frozen (sticky); fixed widths keep the second column's left offset exact.
  {
    accessorKey: 'reportNumber',
    header: t('auditResultReport.dashboard.columns.reportNumber'),
    meta: {
      class: {
        th: 'w-[170px] min-w-[170px] max-w-[170px] sticky left-0 z-20 !bg-gray-100 dark:!bg-gray-800',
        td: 'w-[170px] min-w-[170px] max-w-[170px] sticky left-0 z-10 bg-[var(--bg-main)]'
      }
    }
  },
  {
    accessorKey: 'reportTitle',
    header: t('auditResultReport.dashboard.columns.reportTitle'),
    meta: {
      class: {
        th: 'w-[280px] min-w-[280px] max-w-[280px] sticky left-[170px] z-20 !bg-gray-100 dark:!bg-gray-800 border-r border-r-[var(--border-main)] shadow-[2px_0_4px_-2px_rgba(0,0,0,0.15)]',
        td: 'w-[280px] min-w-[280px] max-w-[280px] whitespace-normal break-words sticky left-[170px] z-10 bg-[var(--bg-main)] border-r border-r-[var(--border-main)] shadow-[2px_0_4px_-2px_rgba(0,0,0,0.15)]'
      }
    }
  },
  { accessorKey: 'reportDate', header: t('auditResultReport.dashboard.columns.date') },
  { accessorKey: 'findingsCount', header: t('auditResultReport.dashboard.columns.findings') },
  { accessorKey: 'category', header: t('auditResultReport.dashboard.columns.category') },
  { accessorKey: 'listOfFinding', header: t('auditResultReport.dashboard.columns.findingList') },
  { accessorKey: 'action', header: t('auditResultReport.dashboard.columns.action') },
  { accessorKey: 'status', header: t('auditResultReport.dashboard.columns.status') },
  { accessorKey: 'actions', header: '' }
])

const categoryKeys: Record<string, string> = {
  'Very Significant': 'verySignificant',
  'Significant': 'significant',
  'Quite Significant': 'quiteSignificant',
  'Not Significant': 'notSignificant'
}

const getCategoryLabel = (category: string) => {
  const key = categoryKeys[category]
  return key ? t(`auditResultReport.dashboard.categories.${key}`) : category
}

const getStatusLabel = (status: string) => {
  const key = status.toLowerCase()
  return ['final', 'draft'].includes(key) ? t(`auditResultReport.dashboard.status.${key}`) : status
}

const getRatingColor = (rating: any) => {
  switch (rating) {
    case 'Very Significant': return 'error'
    case 'Significant': return 'warning'
    case 'Quite Significant': return 'info'
    case 'Not Significant': return 'success'
    default: return 'neutral'
  }
}

const CATEGORY_ORDER = ['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']

const getGroupedFindings = (findings: any[] | undefined) => {
  if (!findings || findings.length === 0) return CATEGORY_ORDER.map(c => ({ category: c, count: 0, items: [] }))
  
  return CATEGORY_ORDER.map(cat => {
    const items = findings.filter((f: any) => f.category === cat)
    return {
      category: cat,
      count: items.length,
      items
    }
  })
}

const syncReportFindings = async (report: any) => {
  store.editReport(report)
  await store.runAutoDetectFindings('merge')
}

const viewReportDetail = (report: any) => {
  selectedDetailReport.value = report
  showDetailModal.value = true
}

const handlePrintFromDetail = (report: any) => {
  showDetailModal.value = false
  printReport(report)
}

const printReport = (report: any) => {
  selectedPrintReport.value = report
  showPrintModal.value = true
}
</script>
