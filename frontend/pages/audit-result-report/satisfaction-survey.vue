<template>
  <div>
    <!-- Top Header -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Client Satisfaction Survey</h1>
        <p class="text-gray-500 dark:text-gray-400">Fill in the satisfaction questionnaire for published audit result reports</p>
      </div>
    </div>

    <!-- CSAT Summary Analytics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <!-- CSAT Average Card -->
      <UCard class="relative overflow-hidden bg-gradient-to-br from-primary-50 to-primary-100/50 dark:from-primary-950/20 dark:to-primary-900/10 border border-primary-200/50 dark:border-primary-800/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-semibold text-primary-700 dark:text-primary-400 uppercase tracking-wider">Average CSAT Score</p>
            <h3 class="text-4xl font-extrabold text-primary-900 dark:text-white mt-2">
              {{ averageCsat.toFixed(1) }} <span class="text-lg font-normal text-primary-600 dark:text-primary-400">/ 5.0</span>
            </h3>
            <p class="text-md text-primary-700 dark:text-primary-400 mt-2 flex items-center gap-1">
              <UIcon name="i-heroicons-sparkles" class="size-4 text-amber-500 animate-pulse" />
              Target CSAT is 4.5
            </p>
          </div>
          <div class="p-4 bg-primary-500/10 rounded-2xl border border-primary-500/20">
            <UIcon name="i-heroicons-face-smile" class="size-10 text-primary-600 dark:text-primary-400" />
          </div>
        </div>
      </UCard>

      <!-- Total Surveys Card -->
      <UCard class="relative overflow-hidden bg-gradient-to-br from-indigo-50 to-indigo-100/50 dark:from-indigo-950/20 dark:to-indigo-900/10 border border-indigo-200/50 dark:border-indigo-800/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-semibold text-indigo-700 dark:text-indigo-400 uppercase tracking-wider">Total Surveys Submitted</p>
            <h3 class="text-4xl font-extrabold text-indigo-900 dark:text-white mt-2">
              {{ surveysStore.surveys.length }}
            </h3>
            <p class="text-md text-indigo-700 dark:text-indigo-400 mt-2">
              From {{ publishedReports.length }} published reports
            </p>
          </div>
          <div class="p-4 bg-indigo-500/10 rounded-2xl border border-indigo-500/20">
            <UIcon name="i-heroicons-document-check" class="size-10 text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </UCard>

      <!-- Completion Rate Card -->
      <UCard class="relative overflow-hidden bg-gradient-to-br from-emerald-50 to-emerald-100/50 dark:from-emerald-950/20 dark:to-emerald-900/10 border border-emerald-200/50 dark:border-emerald-800/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-semibold text-emerald-700 dark:text-emerald-400 uppercase tracking-wider">Response Rate</p>
            <h3 class="text-4xl font-extrabold text-emerald-900 dark:text-white mt-2">
              {{ responseRate }}%
            </h3>
            <div class="mt-2 w-32">
              <UProgress :value="responseRate" color="success" size="sm" />
            </div>
          </div>
          <div class="p-4 bg-emerald-500/10 rounded-2xl border border-emerald-500/20">
            <UIcon name="i-heroicons-chart-bar-solid" class="size-10 text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </UCard>
    </div>

    <!-- Main Content Tabs / Tables -->
    <UCard class="shadow-sm border border-[var(--border-main)] overflow-hidden">
      <template #header>
        <div class="flex justify-between items-center py-1">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-heroicons-clipboard-document-list" class="text-primary-600" />
            Published Reports & Survey Status
          </h2>
          <UButton
            icon="i-heroicons-arrow-path"
            color="neutral"
            variant="ghost"
            size="sm"
            @click="refreshAllData"
            :loading="loadingData"
          >
            Refresh
          </UButton>
        </div>
      </template>

      <!-- Published Reports Table -->
      <div v-if="publishedReports.length > 0" class="overflow-x-auto">
        <UTable :data="publishedReports" :columns="columns" class="w-full">
          <!-- Report Number -->
          <template #reportNumber-cell="{ row }">
            <span class="font-mono text-md font-semibold text-primary-600 dark:text-primary-400">
              {{ row.original.reportNumber || (row.original as any).report_number || '-' }}
            </span>
          </template>

          <!-- Title -->
          <template #reportTitle-cell="{ row }">
            <div class="max-w-md truncate font-medium text-gray-800 dark:text-gray-200" :title="row.original.reportTitle">
              {{ row.original.reportTitle }}
            </div>
          </template>

          <!-- Department -->
          <template #department-cell="{ row }">
            <UBadge color="neutral" variant="soft">
              {{ row.original.department || 'General' }}
            </UBadge>
          </template>

          <!-- Report Date -->
          <template #reportDate-cell="{ row }">
            <span class="text-sm font-medium text-gray-600 dark:text-gray-400">
              {{ row.original.reportDate || (row.original as any).report_date?.split('T')[0] || '-' }}
            </span>
          </template>

          <!-- Status -->
          <template #status-cell="{ row }">
            <UBadge color="success" variant="soft" size="sm">
              {{ row.original.status }}
            </UBadge>
          </template>

          <!-- Survey Status -->
          <template #surveyStatus-cell="{ row }">
            <div class="flex items-center gap-2">
              <span v-if="getReportSurvey(row.original)" class="inline-flex items-center gap-1 text-sm font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950/30 px-2 py-1 rounded-md border border-emerald-100 dark:border-emerald-900/30">
                <UIcon name="i-heroicons-check-circle" class="size-4" />
                Submitted ({{ getReportSurvey(row.original)!.overall_score?.toFixed(1) }}/5)
              </span>
              <span v-else class="inline-flex items-center gap-1 text-sm font-semibold text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-2 py-1 rounded-md border border-amber-100 dark:border-amber-900/30">
                <UIcon name="i-heroicons-clock" class="size-4" />
                Pending Survey
              </span>
            </div>
          </template>

          <!-- Actions -->
          <template #actions-cell="{ row }">
            <div class="flex gap-2">
              <UButton
                v-if="!getReportSurvey(row.original)"
                color="primary"
                size="sm"
                icon="i-heroicons-pencil-square"
                label="Fill Survey"
                @click="openSurveyForm(row.original)"
              />
              <UButton
                v-else
                color="neutral"
                variant="outline"
                size="sm"
                icon="i-heroicons-eye"
                label="View Feedback"
                @click="viewSurveyFeedback(getReportSurvey(row.original)!, row.original)"
              />
            </div>
          </template>
        </UTable>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-16 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-dashed border-gray-200 dark:border-gray-800">
        <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-350">No Published Reports</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-2 max-w-md mx-auto">
          There are currently no published (Final) audit result reports available to evaluate.
        </p>
      </div>
    </UCard>

    <!-- Survey Form Modal -->
    <UModal v-model:open="showFormModal" dismissible :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
      <template #content>
        <div class="flex flex-col h-full max-h-[95vh]">
          <!-- Header -->
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-gray-50 dark:bg-gray-900/50">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-primary-100 dark:bg-primary-950/50 rounded-lg">
                <UIcon name="i-heroicons-chat-bubble-bottom-center-text" class="text-primary-600 size-6" />
              </div>
              <div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">Auditee Satisfaction Questionnaire</h3>
                <p class="text-md text-gray-500 dark:text-gray-400 mt-0.5">Please rate your satisfaction with our audit process</p>
              </div>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="showFormModal = false"
            />
          </div>

          <!-- Body -->
          <div class="p-6 overflow-y-auto flex-1 space-y-6">
            <!-- Report Meta Info Banner -->
            <div class="p-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800 rounded-xl space-y-2">
              <div class="grid grid-cols-2 gap-4 text-md">
                <div>
                  <span class="text-gray-500 block">Report Number</span>
                  <span class="font-mono font-semibold text-gray-800 dark:text-gray-200">
                    {{ selectedReport?.reportNumber || (selectedReport as any)?.report_number }}
                  </span>
                </div>
                <div>
                  <span class="text-gray-500 block">Department</span>
                  <span class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ selectedReport?.department || 'General' }}
                  </span>
                </div>
                <div class="col-span-2">
                  <span class="text-gray-500 block">Report Title</span>
                  <span class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ selectedReport?.reportTitle }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Survey Form -->
            <UForm :state="formState" class="space-y-6" @submit="submitSurvey">
              <!-- Name & Department info fields -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <UFormField label="Your Name (Auditee)" name="auditeeName" required>
                  <UInput
                    v-model="formState.auditeeName"
                    placeholder="Enter your name"
                    class="w-full"
                  />
                </UFormField>

                <UFormField label="Department / Unit" name="department" required>
                  <UInput
                    v-model="formState.department"
                    placeholder="Department"
                    disabled
                    class="w-full bg-gray-50 dark:bg-gray-900 cursor-not-allowed"
                  />
                </UFormField>
              </div>

              <!-- Rating 1: Clarity -->
              <div class="p-4 border border-gray-150 dark:border-gray-800 rounded-xl space-y-3">
                <div>
                  <h4 class="text-sm font-bold text-gray-800 dark:text-gray-250 flex justify-between">
                    <span>1. Clarity of Audit Findings & Recommendations</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingClarity }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">Kejelasan temuan audit dan rekomendasi perbaikan yang diberikan.</p>
                </div>
                <div class="flex items-center gap-2 pt-1">
                  <button
                    v-for="star in 5"
                    :key="star"
                    type="button"
                    @click="formState.ratingClarity = star"
                    class="focus:outline-none transition-transform hover:scale-110"
                  >
                    <UIcon
                      :name="star <= formState.ratingClarity ? 'i-heroicons-star-solid' : 'i-heroicons-star'"
                      :class="star <= formState.ratingClarity ? 'text-amber-500 w-8 h-8' : 'text-gray-300 dark:text-gray-700 w-8 h-8'"
                    />
                  </button>
                  <span class="ml-2 text-md font-semibold text-gray-600 dark:text-gray-400">
                    {{ getRatingText(formState.ratingClarity) }}
                  </span>
                </div>
              </div>

              <!-- Rating 2: Professionalism -->
              <div class="p-4 border border-gray-150 dark:border-gray-800 rounded-xl space-y-3">
                <div>
                  <h4 class="text-sm font-bold text-gray-800 dark:text-gray-250 flex justify-between">
                    <span>2. Professionalism of the Audit Team</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingProfessionalism }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">Profesionalitas, objektivitas, serta etika tim auditor internal selama proses audit.</p>
                </div>
                <div class="flex items-center gap-2 pt-1">
                  <button
                    v-for="star in 5"
                    :key="star"
                    type="button"
                    @click="formState.ratingProfessionalism = star"
                    class="focus:outline-none transition-transform hover:scale-110"
                  >
                    <UIcon
                      :name="star <= formState.ratingProfessionalism ? 'i-heroicons-star-solid' : 'i-heroicons-star'"
                      :class="star <= formState.ratingProfessionalism ? 'text-amber-500 w-8 h-8' : 'text-gray-300 dark:text-gray-700 w-8 h-8'"
                    />
                  </button>
                  <span class="ml-2 text-md font-semibold text-gray-600 dark:text-gray-400">
                    {{ getRatingText(formState.ratingProfessionalism) }}
                  </span>
                </div>
              </div>

              <!-- Rating 3: Timeliness -->
              <div class="p-4 border border-gray-150 dark:border-gray-800 rounded-xl space-y-3">
                <div>
                  <h4 class="text-sm font-bold text-gray-800 dark:text-gray-250 flex justify-between">
                    <span>3. Timeliness of Audit Report Delivery</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingTimeliness }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">Ketepatan waktu penyampaian laporan hasil audit (LHA) dari jadwal yang ditentukan.</p>
                </div>
                <div class="flex items-center gap-2 pt-1">
                  <button
                    v-for="star in 5"
                    :key="star"
                    type="button"
                    @click="formState.ratingTimeliness = star"
                    class="focus:outline-none transition-transform hover:scale-110"
                  >
                    <UIcon
                      :name="star <= formState.ratingTimeliness ? 'i-heroicons-star-solid' : 'i-heroicons-star'"
                      :class="star <= formState.ratingTimeliness ? 'text-amber-500 w-8 h-8' : 'text-gray-300 dark:text-gray-700 w-8 h-8'"
                    />
                  </button>
                  <span class="ml-2 text-md font-semibold text-gray-600 dark:text-gray-400">
                    {{ getRatingText(formState.ratingTimeliness) }}
                  </span>
                </div>
              </div>

              <!-- Computed Preview of Overall Score -->
              <div class="p-4 bg-gradient-to-r from-gray-50 to-slate-100 dark:from-slate-900 dark:to-slate-850 border border-gray-200 dark:border-gray-850 rounded-xl flex items-center justify-between">
                <div>
                  <span class="text-sm font-semibold text-gray-500 uppercase tracking-wider block">Estimated CSAT Score</span>
                  <span class="text-md text-gray-400 mt-0.5">Average: (Clarity + Professionalism + Timeliness) / 3</span>
                </div>
                <div class="text-right">
                  <span class="text-2xl font-extrabold text-primary-600 dark:text-primary-400">
                    {{ computedOverallScore.toFixed(2) }}
                  </span>
                  <span class="text-sm text-gray-500">/ 5.0</span>
                </div>
              </div>

              <!-- Comments -->
              <UFormField label="Additional Comments / Suggestions" name="comments">
                <UTextarea
                  v-model="formState.comments"
                  placeholder="Tell us what went well or what we can improve..."
                  :rows="4"
                  class="w-full animate-fade-in"
                />
              </UFormField>

              <!-- Footer Actions -->
              <div class="flex justify-end gap-3 pt-6 border-t border-gray-100 dark:border-gray-800">
                <UButton
                  label="Cancel"
                  color="neutral"
                  variant="ghost"
                  @click="showFormModal = false"
                />
                <UButton
                  type="submit"
                  label="Submit Survey"
                  color="primary"
                  icon="i-heroicons-check"
                  :loading="submitting"
                />
              </div>
            </UForm>
          </div>
        </div>
      </template>
    </UModal>

    <!-- View Feedback Modal -->
    <UModal v-model:open="showViewModal" dismissible :ui="{ content: 'sm:max-w-lg bg-[var(--bg-main)] border border-[var(--border-main)]' }">
      <template #content>
        <div class="flex flex-col h-full max-h-[90vh]">
          <!-- Header -->
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-gray-50 dark:bg-gray-900/50">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-emerald-100 dark:bg-emerald-950/50 rounded-lg">
                <UIcon name="i-heroicons-clipboard-document-check" class="text-emerald-600 size-6" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">Submitted Survey Details</h3>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="showViewModal = false"
            />
          </div>

          <!-- Body -->
          <div class="p-6 overflow-y-auto space-y-6">
            <div class="space-y-4">
              <!-- Report and Auditee Header Info -->
              <div class="p-4 bg-emerald-50/30 dark:bg-emerald-950/10 border border-emerald-100/50 dark:border-emerald-900/20 rounded-xl space-y-2">
                <div class="text-sm font-semibold text-emerald-800 dark:text-emerald-400">
                  {{ viewReportDetail?.reportTitle }}
                </div>
                <div class="text-md text-gray-500 dark:text-gray-400 grid grid-cols-2 gap-2 mt-1">
                  <div>
                    <span class="text-md text-gray-400 block">Submitted By</span>
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ viewSurveyDetail?.auditee_name }}</span>
                  </div>
                  <div>
                    <span class="text-md text-gray-400 block">Department</span>
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ viewSurveyDetail?.department }}</span>
                  </div>
                </div>
              </div>

              <!-- Star ratings display -->
              <div class="space-y-3">
                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">Clarity Rating</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_clarity || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_clarity }}/5)</span>
                  </div>
                </div>

                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">Professionalism Rating</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_professionalism || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_professionalism }}/5)</span>
                  </div>
                </div>

                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">Timeliness Rating</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_timeliness || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_timeliness }}/5)</span>
                  </div>
                </div>
              </div>

              <!-- Overall Score -->
              <div class="p-4 bg-gradient-to-r from-emerald-500/10 to-emerald-500/5 border border-emerald-500/20 rounded-xl flex items-center justify-between">
                <div>
                  <span class="text-sm font-semibold text-emerald-800 dark:text-emerald-400 uppercase tracking-wider block">CSAT Overall Score</span>
                  <span class="text-md text-gray-500 mt-0.5">Average calculated score</span>
                </div>
                <div class="text-right">
                  <span class="text-3xl font-extrabold text-emerald-600 dark:text-emerald-400">
                    {{ viewSurveyDetail?.overall_score?.toFixed(2) }}
                  </span>
                  <span class="text-sm text-gray-500">/ 5.0</span>
                </div>
              </div>

              <!-- Comments -->
              <div class="space-y-1">
                <span class="text-md text-gray-400 uppercase font-semibold">Comments / Feedback</span>
                <div class="p-4 bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800 rounded-xl text-md text-gray-700 dark:text-gray-300 whitespace-pre-line leading-relaxed italic">
                  "{{ viewSurveyDetail?.comments || 'No additional comments provided.' }}"
                </div>
              </div>
            </div>

            <!-- Footer Action -->
            <div class="flex justify-end pt-4 border-t border-gray-100 dark:border-gray-800">
              <UButton
                label="Close"
                color="neutral"
                variant="soft"
                @click="showViewModal = false"
              />
            </div>
          </div>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '~/stores/auth'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAuditExecutionStore } from '~/stores/audit-execution'
import { useAuditeeSurveyStore, type AuditeeSurvey } from '~/stores/auditee-survey'

const authStore = useAuthStore()
const reportStore = useAuditResultReportStore()
const executionStore = useAuditExecutionStore()
const surveysStore = useAuditeeSurveyStore()

const loadingData = ref(false)
const showFormModal = ref(false)
const showViewModal = ref(false)
const submitting = ref(false)

const selectedReport = ref<any>(null)
const viewSurveyDetail = ref<AuditeeSurvey | null>(null)
const viewReportDetail = ref<any>(null)

// Initial form state
const formState = ref({
  auditeeName: '',
  department: '',
  ratingClarity: 5,
  ratingProfessionalism: 5,
  ratingTimeliness: 5,
  comments: ''
})

const columns = [
  { accessorKey: 'reportNumber', header: 'No. LHA' },
  { accessorKey: 'reportTitle', header: 'Report Title' },
  { accessorKey: 'department', header: 'Department' },
  { accessorKey: 'reportDate', header: 'Publish Date' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'surveyStatus', header: 'Survey Status' },
  { accessorKey: 'actions', header: 'Action' }
]

// Computed list of only published (status Final) reports
const publishedReports = computed(() => {
  return reportStore.reportList.filter((r: any) => r.status === 'Final' || r.status === 'Published')
})

// Match a report with its corresponding AuditeeSurvey
const getReportSurvey = (report: any): AuditeeSurvey | undefined => {
  // Find matching audit execution
  const execution = executionStore.auditExecutions.find(
    (e: any) => e.ref === (report.assignmentLetterId || report.assignment_letter_id)
  )
  
  if (execution) {
    // If execution found, match survey by AuditExecutionID
    const survey = surveysStore.surveys.find((s: any) => s.audit_execution_id === execution.id)
    if (survey) return survey
  }

  // Fallback: match by department and year/month or name
  const repDateStr = report.reportDate || report.report_date
  const repDate = repDateStr ? new Date(repDateStr) : new Date()
  const repYear = repDate.getFullYear()
  const repMonth = repDate.getMonth() + 1 // 1-12

  return surveysStore.surveys.find((s: any) => 
    s.department === report.department && 
    s.year === repYear && 
    s.month === repMonth
  )
}

// Stats metrics computed properties
const averageCsat = computed(() => {
  if (surveysStore.surveys.length === 0) return 4.7 // fallback default baseline
  const sum = surveysStore.surveys.reduce((acc, curr) => acc + (curr.overall_score || 0), 0)
  return sum / surveysStore.surveys.length
})

const responseRate = computed(() => {
  if (publishedReports.value.length === 0) return 100
  const count = publishedReports.value.filter(r => getReportSurvey(r) !== undefined).length
  return Math.round((count / publishedReports.value.length) * 100)
})

const computedOverallScore = computed(() => {
  return (formState.value.ratingClarity + formState.value.ratingProfessionalism + formState.value.ratingTimeliness) / 3.0
})

const getRatingText = (rating: number) => {
  switch (rating) {
    case 1: return 'Very Dissatisfied'
    case 2: return 'Dissatisfied'
    case 3: return 'Neutral'
    case 4: return 'Satisfied'
    case 5: return 'Very Satisfied'
    default: return ''
  }
}

// Refresh all data from microservices
const refreshAllData = async () => {
  loadingData.value = true
  try {
    await Promise.all([
      reportStore.fetchReports(),
      executionStore.fetchAuditExecutions(),
      surveysStore.fetchSurveys()
    ])
  } catch (error) {
    console.error('Error refreshing survey data:', error)
  } finally {
    loadingData.value = false
  }
}

onMounted(async () => {
  await refreshAllData()
})

const openSurveyForm = (report: any) => {
  selectedReport.value = report
  
  // Set default values for the form
  formState.value = {
    auditeeName: authStore.user?.fullName || '',
    department: report.department || 'General',
    ratingClarity: 5,
    ratingProfessionalism: 5,
    ratingTimeliness: 5,
    comments: ''
  }
  
  showFormModal.value = true
}

const viewSurveyFeedback = (survey: AuditeeSurvey, report: any) => {
  viewSurveyDetail.value = survey
  viewReportDetail.value = report
  showViewModal.value = true
}

const submitSurvey = async () => {
  if (!selectedReport.value) return
  submitting.value = true
  
  try {
    // 1. Get corresponding Audit Execution ID
    const execution = executionStore.auditExecutions.find(
      (e: any) => e.ref === (selectedReport.value.assignmentLetterId || selectedReport.value.assignment_letter_id)
    )

    // Check if survey already submitted for this execution
    if (execution) {
      const existing = surveysStore.surveys.find((s: any) => s.audit_execution_id === execution.id)
      if (existing) {
        alert("A survey has already been submitted for this audit report!")
        submitting.value = false
        return
      }
    }
    
    // 2. Determine year and month from report date
    const repDateStr = selectedReport.value.reportDate || selectedReport.value.report_date
    const repDate = repDateStr ? new Date(repDateStr) : new Date()
    const repYear = repDate.getFullYear()
    const repMonth = repDate.getMonth() + 1

    const payload: AuditeeSurvey = {
      audit_execution_id: execution?.id,
      auditee_name: formState.value.auditeeName,
      department: formState.value.department,
      year: repYear,
      month: repMonth,
      rating_clarity: formState.value.ratingClarity,
      rating_professionalism: formState.value.ratingProfessionalism,
      rating_timeliness: formState.value.ratingTimeliness,
      comments: formState.value.comments
    }
    
    await surveysStore.createSurvey(payload)
    
    // Refresh page data
    await refreshAllData()
    
    showFormModal.value = false
  } catch (error: any) {
    console.error('Failed to submit survey:', error)
    alert(error.message || 'Failed to submit satisfaction survey. Please try again.')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
