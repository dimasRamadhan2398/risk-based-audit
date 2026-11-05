<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950">
    <!-- Stepper header (clickable) -->
    <div class="flex items-center gap-3 border-b border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 py-3 sm:px-6 sm:py-4 shrink-0">
      <!-- Horizontal scroll only: the header never grows or scrolls vertically -->
      <div
        ref="stepScrollRef"
        class="flex-1 min-w-0 overflow-x-auto overflow-y-hidden scrollbar-thin"
      >
        <div class="flex min-w-max items-center justify-center gap-3 px-1 py-1 sm:gap-4">
          <template
            v-for="(sec, idx) in sections"
            :key="sec.id"
          >
            <div
              class="group flex shrink-0 cursor-pointer select-none items-center gap-2 sm:gap-2.5"
              data-testid="form-step"
              @click="goToStep(idx)"
            >
              <div
                class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 sm:h-8 sm:w-8"
                :class="currentStepIndex === idx
                  ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105'
                  : currentStepIndex > idx
                    ? 'bg-secondary text-white shadow-sm'
                    : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
              >
                <UIcon
                  v-if="currentStepIndex > idx"
                  name="i-heroicons-check"
                  class="h-3.5 w-3.5 stroke-[3]"
                />
                <span v-else>{{ idx + 1 }}</span>
              </div>
              <span
                class="whitespace-nowrap text-xs font-semibold transition-colors"
                :class="currentStepIndex === idx
                  ? 'text-primary font-bold'
                  : currentStepIndex > idx
                    ? 'text-secondary font-semibold'
                    : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'"
              >
                {{ sec.title }}
              </span>
            </div>
            <div
              v-if="idx < sections.length - 1"
              class="h-0.5 w-5 shrink-0 rounded transition-all duration-300 sm:w-8"
              :class="currentStepIndex > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
            />
          </template>
        </div>
      </div>
      <UBadge
        :color="store.form.status === 'Approved' ? 'success' : 'warning'"
        variant="soft"
        class="hidden sm:inline-flex shrink-0 font-semibold"
        :title="store.form.status === 'Approved' ? t('executiveSummaryForm.lock.approvedDescription') : t('executiveSummaryForm.lock.draftDescription')"
      >
        <UIcon name="i-lucide-shield-alert" class="size-3.5 mr-1" />
        {{ store.form.status === 'Approved' ? t('executiveSummaryForm.lock.approved') : t('executiveSummaryForm.lock.draft') }}
      </UBadge>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 min-h-0 overflow-y-auto p-4 sm:p-6 lg:p-8 space-y-8 bg-white dark:bg-gray-900" ref="scrollContainer">
      
      <form @submit.prevent class="space-y-12">
        <!-- 1. Metadata & Document Upload -->
        <section id="sec-upload" v-show="activeSection === 'sec-upload'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.upload.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.upload.description') }}</p>
          </div>

          <div class="w-full">
            <UFormField :label="t('executiveSummaryForm.upload.lhaLabel')">
              <USelectMenu
                v-model="selectedLhaId"
                :items="lhaDropdownOptions"
                :placeholder="t('executiveSummaryForm.upload.lhaPlaceholder')"
                class="w-full font-semibold"
                :disabled="isLocked"
                @update:modelValue="onLhaSelect"
              />
            </UFormField>
          </div>

          <!-- Audit Result Report (LHA) Linked Context Card -->
          <div v-if="linkedLhaReport" class="bg-primary-50/60 dark:bg-primary-950/30 border border-primary-200 dark:border-primary-800/60 rounded-xl p-4 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div class="flex items-center gap-3">
              <div class="p-2.5 bg-primary-100 dark:bg-primary-900/60 text-primary-600 dark:text-primary-400 rounded-lg">
                <UIcon name="i-lucide-file-text" class="size-5" />
              </div>
              <div>
                <div class="text-xs font-bold uppercase tracking-wider text-primary-700 dark:text-primary-300 flex items-center gap-2">
                  <span>{{ t('executiveSummaryIndividual.formExtra.lhaNumber') }} {{ linkedLhaReport.reportNumber || (linkedLhaReport as any).report_number }}</span>
                  <UBadge :color="linkedLhaReport.status === 'Final' || (linkedLhaReport.status as any) === 'Approved' ? 'success' : 'warning'" variant="subtle" size="xs">
                    {{ linkedLhaReport.status || 'Draft' }}
                  </UBadge>
                  <UBadge v-if="linkedLhaReport.assignmentLetterId" color="neutral" variant="outline" size="xs" class="font-mono"> {{ t('executiveSummaryIndividual.formExtra.assignmentLetter') }} {{ linkedLhaReport.assignmentLetterId }}
                  </UBadge>
                </div>
                <div class="text-sm font-semibold text-gray-900 dark:text-white mt-0.5">
                  {{ linkedLhaReport.reportTitle }}
                </div>
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <div v-if="linkedLhaReport.department">
                <span class="text-gray-400">{{ t('executiveSummaryIndividual.formExtra.division') }}</span> <span class="font-medium text-gray-700 dark:text-gray-300">{{ store.formatDepartmentName(linkedLhaReport.department) }}</span>
              </div>
              <div v-if="linkedLhaReport.reportDate">
                <span class="text-gray-400">{{ t('auditActivityPlan.view.date') }}</span> <span class="font-medium text-gray-700 dark:text-gray-300">{{ linkedLhaReport.reportDate }}</span>
              </div>
              <div>
                <span class="text-gray-400">{{ t('executiveSummaryIndividual.formExtra.findings') }}</span> <span class="font-bold text-primary-600 dark:text-primary-400">{{ linkedLhaReport.findingsCount || (linkedLhaReport.findings?.length || 0) }}</span>
              </div>
            </div>
          </div>

          <!-- Document Upload Field -->
          <div class="bg-gray-50 dark:bg-gray-800/50 p-6 rounded-xl border border-dashed border-gray-300 dark:border-gray-700">
            <UFormField :label="t('executiveSummaryForm.upload.documentLabel')" required>
              <div class="space-y-4">
                <div class="flex items-center gap-4">
                  <!-- File Input simulation -->
                  <input
                    type="file"
                    ref="fileInput"
                    accept=".pdf,.docx"
                    class="hidden"
                    @change="handleFileUpload"
                    :disabled="isLocked"
                  />
                  <UButton
                    color="neutral"
                    variant="solid"
                    icon="i-lucide-upload"
                    :label="t('executiveSummaryForm.upload.chooseFile')"
                    :disabled="isLocked"
                    @click="fileInput?.click()"
                  />
                  <span v-if="store.form.dokumenPath" class="text-sm text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                    <UIcon name="i-lucide-file-check" class="text-success-500 size-5" />
                    {{ store.form.dokumenPath }}
                  </span>
                  <span v-else class="text-sm text-gray-400">{{ t('executiveSummaryForm.upload.noFile') }}</span>
                </div>
                <p class="text-md text-gray-400">
                  {{ t('executiveSummaryForm.upload.formats') }}
                </p>
              </div>
            </UFormField>
          </div>
        </section>

        <!-- 2. Section Narrative -->
        <section id="sec-narrative" v-show="activeSection === 'sec-narrative'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.narrative.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.narrative.description') }}</p>
          </div>

          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-md font-bold text-gray-400 uppercase">{{ t('executiveSummaryForm.narrative.label') }}</span>
              <UButton
                v-if="!isLocked"
                color="neutral"
                variant="ghost"
                icon="i-lucide-refresh-cw"
                :label="t('executiveSummaryForm.narrative.reset')"
                size="md"
                @click="resetNarrativeToDefault"
              />
            </div>
            <UTextarea
              v-model="store.form.narrative"
              :placeholder="t('executiveSummaryForm.narrative.placeholder')"
              :rows="6"
              class="w-full"
              :disabled="isLocked"
            />
          </div>
        </section>

        <!-- 4. Section Status Tindak Lanjut -->
        <section id="sec-followup" v-show="activeSection === 'sec-followup'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.followUp.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.followUp.description') }}</p>
          </div>

          <!-- Counts from the ATR buckets of the linked LHA's findings (read-only); frozen once approved -->
          <ExecutiveSummaryFollowUpTable
            :rows="store.form.followUpTable"
            :locked="isLocked"
            :loading="followUp.loading.value"
            :error="followUp.error.value"
            :approved-date-label="approvedDateLabel"
            :as-of-label="followUpAsOfLabel"
            :refreshable="!isLocked && !!followUpQuery"
            :empty-hint="followUpEmptyHint"
            @refresh="refreshFollowUp"
          />

          <!-- Live status next to the frozen snapshot, only for users who may view every ATR -->
          <ExecutiveSummaryCurrentFollowUp
            v-if="isLocked && canViewLiveFollowUpStatus && followUpQuery"
            :query="followUpQuery"
          />
        </section>

        <!-- 5. Section Top 5 Significant Findings -->
        <section id="sec-topfindings" v-show="activeSection === 'sec-topfindings'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <div class="flex justify-between items-center">
              <div>
                <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                  {{ t('executiveSummaryForm.topFindings.title') }}
                </h2>
                <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.topFindings.description') }}</p>
              </div>
              <div class="flex items-center gap-2">
                <UButton
                  v-if="!isLocked"
                  color="neutral"
                  variant="outline"
                  icon="i-lucide-refresh-cw"
                  :label="t('executiveSummaryForm.sync.findings')"
                  size="sm"
                  @click="requireLhaThen(syncFindingsFromLha)"
                />
                <UButton
                  v-if="!isLocked && store.form.topFindings.length < 5"
                  color="primary"
                  variant="soft"
                  icon="i-lucide-plus-circle"
                  :label="t('executiveSummaryForm.topFindings.add')"
                  size="sm"
                  @click="addTopFinding"
                />
              </div>
            </div>
          </div>

          <SignificantFindingCards
            ref="findingCardsRef"
            :findings="store.form.topFindings"
            :disabled="isLocked"
            :unit-options="store.unitDivisionOptions"
            :unit-loading="store.departmentsLoading"
            :unit-error="store.departmentsError"
            :risk-options="significanceLevelOptions"
            @remove="removeTopFinding"
          />
          <div v-if="store.form.topFindings.length === 0" class="text-center py-8 bg-gray-50 dark:bg-gray-800/30 rounded-lg text-sm text-gray-400 border border-dashed">
            {{ t('executiveSummaryForm.topFindings.empty') }}
            <button v-if="!isLocked" type="button" @click="addTopFinding" class="text-primary-500 font-bold ml-1 hover:underline">
              {{ t('executiveSummaryForm.topFindings.emptyAction') }}
            </button>
          </div>
        </section>

        <!-- 5. Recurring findings analysis & conclusion (report Sections V & VII) -->
        <section id="sec-analysis" v-show="activeSection === 'sec-analysis'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryForm.analysis.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.analysis.description') }}</p>
            </div>
            <UButton
              v-if="!isLocked"
              color="neutral"
              variant="outline"
              icon="i-lucide-refresh-cw"
              :label="t('executiveSummaryForm.sync.analysis')"
              size="sm"
              @click="requireLhaThen(syncFindingsFromLha)"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <UFormField :label="t('executiveSummaryForm.analysis.theme')">
              <UTextarea
                v-model="store.form.akarMasalah"
                :placeholder="t('executiveSummaryForm.analysis.rootCausePlaceholder')"
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryForm.analysis.conclusion')">
              <UTextarea
                v-model="store.form.kesimpulan"
                :placeholder="t('executiveSummaryForm.analysis.conclusionPlaceholder')"
                :rows="4"
                class="w-full"
                :disabled="isLocked"
              />
            </UFormField>
          </div>

          <!-- Electronic Signatures -->
          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4">
            <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryForm.signature.title') }}</h4>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
              <UFormField :label="t('executiveSummaryForm.signature.place')" required>
                <UInput
                  v-model="store.form.signatureTempat"
                  :placeholder="t('executiveSummaryForm.signature.placePlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signaturePlace'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureTempat ? store.form.signatureTempat.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.date')" required>
                <AppDatePicker
                  v-model="store.form.signatureTanggal"
                  :disabled="isLocked"
                />
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.headName')" required>
                <UInput
                  v-model="store.form.signatureNamaKepala"
                  :placeholder="t('executiveSummaryForm.signature.headNamePlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signatureName'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNamaKepala ? store.form.signatureNamaKepala.length : 0 }}/50
                </div>
              </UFormField>

              <UFormField :label="t('executiveSummaryForm.signature.headNik')" required>
                <UInput
                  v-model="store.form.signatureNIK"
                  :placeholder="t('executiveSummaryForm.signature.headNikPlaceholder')"
                  :disabled="isLocked"
                  maxlength="50"
                  @invalid="($event.target as any)?.setCustomValidity(t('executiveSummaryForm.validation.signatureNik'))"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.form.signatureNIK ? store.form.signatureNIK.length : 0 }}/50
                </div>
              </UFormField>
            </div>
          </div>
        </section>

        <!-- Catatan Executive untuk Auditor -->
        <section v-if="store.isViewing" id="sec-notes" v-show="activeSection === 'sec-notes'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-message-square-text" class="size-5 text-primary-500" />
              {{ t('executiveSummaryForm.notes.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.notes.description') }}</p>
          </div>

          <UFormField :label="t('executiveSummaryForm.notes.label')">
            <UTextarea
              v-model="store.form.executiveNote"
              :placeholder="canWriteExecutiveNote ? t('executiveSummaryForm.notes.placeholder') : (store.form.executiveNote ? '' : t('executiveSummaryForm.notes.empty'))"
              :rows="5"
              class="w-full"
              :disabled="!canWriteExecutiveNote"
            />
          </UFormField>
          <div v-if="canWriteExecutiveNote" class="flex justify-end">
            <UButton
              color="primary"
              icon="i-lucide-save"
              :label="t('executiveSummaryForm.notes.save')"
              :loading="store.loading"
              @click="store.saveExecutiveNote"
            />
          </div>
        </section>

      </form>
    </div>

    <!-- Footer navigation -->
    <div class="flex items-center justify-between gap-3 border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 px-4 py-3 sm:px-6 shrink-0">
      <UButton
        v-if="currentStepIndex > 0"
        color="neutral"
        variant="outline"
        icon="i-lucide-chevron-left"
        :label="t('executiveSummaryForm.stepper.previous')"
        @click="goToStep(currentStepIndex - 1)"
      />
      <span v-else />
      <UButton
        v-if="currentStepIndex < sections.length - 1"
        color="primary"
        trailing-icon="i-lucide-chevron-right"
        :label="t('executiveSummaryForm.stepper.next')"
        @click="goToStep(currentStepIndex + 1)"
      />
    </div>

    <!-- Sticky Bottom Footer Controls -->
    <div class="col-span-1 lg:col-span-4 bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex justify-between items-center sticky bottom-0 z-20">
      <div>
        <UButton
          :label="t('executiveSummaryForm.actions.close')"
          color="neutral"
          variant="outline"
          class="font-semibold"
          @click="() => {store.showModal = false}"
        />
      </div>

      <div class="flex items-center gap-3">
        <!-- Error alert message displays if necessary -->
        <span v-if="store.errorMsg" class="text-md text-error-500 font-semibold max-w-md truncate">{{ store.errorMsg }}</span>
        
        <!-- Workflow buttons -->
        <template v-if="!store.isViewing">
          <UButton
            v-if="!isLocked"
            color="primary"
            variant="solid"
            icon="i-lucide-save"
            :label="t('executiveSummaryForm.actions.saveDraft')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="saveReportDraft"
          />
        </template>
        
        <!-- Locked/Unlocked Overrides -->
        <template v-else>
          <UButton
            v-if="isDraft && canApprove"
            color="success"
            variant="solid"
            icon="i-lucide-check"
            :label="t('executiveSummaryForm.actions.approve')"
            class="font-bold px-6"
            :loading="store.loading"
            @click="openApproveSignature"
          />
        </template>
      </div>
    </div>

    <ExecutiveSummaryApproveModal
      v-model:open="showApproveModal"
      :report-number="store.form.nomorDokumen"
      :initial-place="store.form.signatureTempat"
      :initial-date="store.form.signatureTanggal"
      :initial-signer-name="store.form.signatureNamaKepala"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useExecutiveSummaryStore, isValidUuid, type SignificantFinding } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useExecutiveSummaryFollowUp } from '~/composables/useExecutiveSummaryFollowUp'
import ExecutiveSummaryApproveModal from './ExecutiveSummaryApproveModal.vue'
import ExecutiveSummaryCurrentFollowUp from './ExecutiveSummaryCurrentFollowUp.vue'
import ExecutiveSummaryFollowUpTable from './ExecutiveSummaryFollowUpTable.vue'
import SignificantFindingCards from './SignificantFindingCards.vue'
import { followUpRowsFromStats, formatSyncedAt } from '~/utils/executiveSummaryCompilation'
import {
  BUCKET_STATUS,
  applyFollowUpToTopFindings,
  formatFollowUpDate,
  type FollowUpStatusQuery,
  type FollowUpStatusResult
} from '~/utils/executiveSummaryFollowUp'
import { scrollStepIntoView } from '~/utils/stepperScroll'
import { UserRole } from '~/types/auth'
import { useI18n } from '~/composables/useI18n'
import { useAppToast } from '~/composables/useAppToast'
import {
  auditPeriodFromAssignmentLetter,
  formatAuditPeriod,
  prefixNarrative,
  refreshNarrativePeriod
} from '~/utils/executiveSummaryNarrative'
const store = useExecutiveSummaryStore()
const authStore = useAuthStore()
const { canReviewExecutiveSummary, canViewLiveFollowUpStatus } = useRbac()
const { t, locale } = useI18n()
const { warning, error: errorToast } = useAppToast()

const fileInput = ref<HTMLInputElement | null>(null)
const activeSection = ref('sec-upload')
const scrollContainer = ref<HTMLElement | null>(null)

// Navigation links
const sections = computed(() => [
  { id: 'sec-upload', index: '1', title: t('executiveSummaryIndividual.nav.upload') },
  { id: 'sec-narrative', index: '2', title: t('executiveSummaryIndividual.nav.narrative') },
  { id: 'sec-followup', index: '3', title: t('executiveSummaryIndividual.nav.followUp') },
  { id: 'sec-topfindings', index: '4', title: t('executiveSummaryIndividual.nav.topFindings') },
  { id: 'sec-analysis', index: '5', title: t('executiveSummaryIndividual.nav.analysis') },
  ...(store.isViewing ? [{ id: 'sec-notes', index: '6', title: t('executiveSummaryIndividual.nav.notes') }] : [])
])

// Dropdown options
const divisionOptions = computed(() => store.unitDivisionOptions)

// Significance Level options synchronized with Result Reports (LHA) Category / Significance Level
const significanceLevelOptions = computed(() => [
  { label: t('auditResultReport.dashboard.categories.verySignificant'), value: 'Very Significant' },
  { label: t('auditResultReport.dashboard.categories.significant'), value: 'Significant' },
  { label: t('auditResultReport.dashboard.categories.quiteSignificant'), value: 'Quite Significant' },
  { label: t('auditResultReport.dashboard.categories.notSignificant'), value: 'Not Significant' }
])

// Role computed checks: Admin, CAE, and Audit Manager can write note & approve draft
const canWriteExecutiveNote = computed(() => Boolean(canReviewExecutiveSummary.value))
const canApprove = computed(() => Boolean(canReviewExecutiveSummary.value))
const isChiefAuditExecutive = computed(() => Boolean(canReviewExecutiveSummary.value))
const isDraft = computed(() => String(store.form.status || '').toLowerCase() === 'draft')

const isLocked = computed(() => {
  return store.form.status === 'Approved'
})

const resetNarrativeToDefault = () => {
  store.form.narrative = store.defaultNarrativeTemplate(narrativePeriod.value)
}

// File Upload Handler (with corrupted file simulation & validation)
const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Size limit 10MB
  if (file.size > 10 * 1024 * 1024) {
    warning(t('executiveSummaryForm.messages.fileTooLarge'))
    return
  }

  // Type limit
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (ext !== 'pdf' && ext !== 'docx' && ext !== 'doc') {
    warning(t('executiveSummaryForm.messages.unsupportedFormat'))
    return
  }

  // Simulated file corruption error handling
  if (file.name.toLowerCase().includes('corrupt') || file.name.toLowerCase().includes('damaged')) {
    warning(t('executiveSummaryForm.messages.corruptedFile'))
    return
  }

  // Simulated success
  store.form.dokumenPath = file.name
}

// Section IV: Top 5 dynamic list methods
// Only the card of a newly added finding opens; existing cards keep their state
const findingCardsRef = ref<InstanceType<typeof SignificantFindingCards> | null>(null)
const addTopFinding = () => {
  if (store.form.topFindings.length >= 5) return
  const defaultDiv: string = store.unitDivisionOptions[0] || ''
  store.form.topFindings.push({
    unitDivision: defaultDiv,
    judulTemuan: '',
    risiko: 'Significant',
    statusTL: 'In Progress',
    usulan: ''
  })
  findingCardsRef.value?.expand(store.form.topFindings[store.form.topFindings.length - 1])
}

const removeTopFinding = (idx: number) => {
  store.form.topFindings.splice(idx, 1)
}

// Form submission & workflow helpers
const requireLhaThen = (sync: () => void) => {
  if (!selectedLhaId.value) {
    warning(t('executiveSummaryForm.messages.selectLhaFirst'))
    return
  }
  sync()
}

const saveReportDraft = async () => {
  if (!store.form.nomorDokumen) {
    warning(t('executiveSummaryForm.messages.docNumberRequired'))
    return
  }
  if (!store.form.dokumenPath) {
    warning(t('executiveSummaryForm.messages.documentRequired'))
    return
  }
  
  // Set status Draft
  store.form.status = 'Draft'
  await store.saveForm()
}

// Approval is signed: the signature box in ExecutiveSummaryApproveModal is required before the backend accepts it
const showApproveModal = ref(false)
const openApproveSignature = () => {
  if (store.currentSummary) {
    showApproveModal.value = true
  }
}

// Stepper navigation: one section is visible at a time
const currentStepIndex = computed(() => Math.max(0, sections.value.findIndex(sec => sec.id === activeSection.value)))

// Keeps the active step visible in the horizontal stepper strip (the header itself never scrolls vertically)
const stepScrollRef = ref<HTMLElement | null>(null)
const scrollActiveStepIntoView = () => {
  const container = stepScrollRef.value
  const step = container?.querySelectorAll<HTMLElement>('[data-testid="form-step"]')[currentStepIndex.value]
  if (container && step) scrollStepIntoView(container, step)
}
watch(currentStepIndex, () => scrollActiveStepIntoView(), { flush: 'post' })

const goToStep = (idx: number) => {
  const target = sections.value[idx]
  if (!target) return
  activeSection.value = target.id
  scrollContainer.value?.scrollTo({ top: 0 })
}

// Assignment Letter & LHA Stores Integration for 2-Way Sync
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()

if (!auditReportStore.loading && auditReportStore.reportList.length === 0) {
  auditReportStore.fetchReports()
}
// The audit period of the narrative comes from the assignment letter of the selected LHA
if (!assignmentLetterStore.loading && assignmentLetterStore.assignmentLetterList.length === 0) {
  assignmentLetterStore.fetchAssignmentLetters()
}
if (!store.departmentsLoaded && !store.departmentsLoading) {
  store.fetchDepartments().catch(() => {})
}

// Audit period of an LHA from its assignment letter dates: "Januari 2026", "Januari – Maret 2026"; '' when unknown
const auditPeriodForReport = (report?: { assignmentLetterId?: string } | null): string => {
  const letterRef = report?.assignmentLetterId
  if (!letterRef) return ''
  const letter = assignmentLetterStore.assignmentLetterList.find(
    (st: any) => st.letterNumber === letterRef || st.id === letterRef
  )
  return formatAuditPeriod(t, auditPeriodFromAssignmentLetter(letter))
}

const assignmentLetterDropdownOptions = computed(() => {
  const lettersFromStore = assignmentLetterStore.assignmentLetterList
    .filter((st: any) => st.letterNumber && st.letterNumber.startsWith('ST-') && st.letterNumber.includes('2026'))
    .map((st: any) => ({
      label: `${st.letterNumber} - ${st.auditTitle || ''}`,
      value: st.letterNumber,
      letter: st
    }))
  const uniqueNumbers = new Set(lettersFromStore.map(l => l.value))
  
  auditReportStore.reportList.forEach(r => {
    if (r.assignmentLetterId && r.assignmentLetterId.startsWith('ST-') && r.assignmentLetterId.includes('2026') && !uniqueNumbers.has(r.assignmentLetterId)) {
      uniqueNumbers.add(r.assignmentLetterId)
      lettersFromStore.push({
        label: `${r.assignmentLetterId} - ${r.reportTitle || ''}`,
        value: r.assignmentLetterId,
        letter: { letterNumber: r.assignmentLetterId, auditTitle: r.reportTitle, leader: 'Lead Auditor' } as any
      })
    }
  })

  // Ensure default known letters strictly format ST-XXX/SKAI/2026
  const defaults = ['ST-001/SKAI/2026', 'ST-002/SKAI/2026', 'ST-003/SKAI/2026', 'ST-004/SKAI/2026', 'ST-005/SKAI/2026']
  defaults.forEach(defNum => {
    if (!uniqueNumbers.has(defNum)) {
      uniqueNumbers.add(defNum)
      lettersFromStore.push({
        label: defNum,
        value: defNum,
        letter: { letterNumber: defNum, auditTitle: t('executiveSummaryForm.defaults.auditAssignment'), leader: 'Head of SKAI' } as any
      })
    }
  })

  return lettersFromStore
})

interface AssignmentLetterDropdownItem {
  label: string
  value: string
  letter: any
}

const selectedAssignmentLetter = ref<AssignmentLetterDropdownItem | undefined>(
  assignmentLetterDropdownOptions.value.find(o => o.value === (store.form.assignmentLetterId || 'ST-001/SKAI/2026')) ||
  assignmentLetterDropdownOptions.value[0]
)

const linkedAssignmentLetter = computed(() => {
  const stNum = selectedAssignmentLetter.value?.value || store.form.assignmentLetterId || 'ST-001/SKAI/2026'

  if (!stNum) return null
  return (assignmentLetterStore.assignmentLetterList.find(st => st.letterNumber === stNum) as any) || {
    letterNumber: stNum,
    auditTitle: store.form.narrative ? store.form.narrative.slice(0, 70) + '...' : t('executiveSummaryIndividual.defaults.auditTitle'),
    status: 'Published' as any,
    leader: 'Head of SKAI',
    executionPeriod: `${store.form.periodeBulan} ${store.form.tahun}`,
    auditYear: String(store.form.tahun || new Date().getFullYear())
  }
})

const linkedLhaReport = computed(() => {
  const targetDoc = typeof selectedLhaId.value === 'object'
    ? (selectedLhaId.value as any)?.value
    : (selectedLhaId.value || store.form.nomorDokumen)

  if (!targetDoc) return null

  const found = auditReportStore.reportList.find(r => 
    (r.reportNumber || (r as any).report_number) === targetDoc
  )
  if (found) return found

  if (selectedLhaId.value && typeof selectedLhaId.value === 'object' && (selectedLhaId.value as any).report) {
    return (selectedLhaId.value as any).report
  }

  if (store.form.nomorDokumen) {
    return {
      reportNumber: store.form.nomorDokumen,
      reportTitle: store.form.narrative ? store.form.narrative.slice(0, 70) + '...' : 'Laporan Hasil Audit',
      assignmentLetterId: store.form.assignmentLetterId || '',
      status: store.form.status === 'Approved' ? 'Final' : 'Draft',
      department: 'Operasional',
      reportDate: store.form.signatureTanggal || `${store.form.tahun}-03-31`,
      findingsCount: store.form.jumlahRekomendasi || store.form.topFindings.length || 0
    }
  }

  return null
})

// Period at the start of the narrative, from the linked LHA (assignment letters may load after the LHA is selected)
const narrativePeriod = computed(() => auditPeriodForReport(linkedLhaReport.value))

const lhaDropdownOptions: Ref<Array<{ label: string; value: string; report: any }>> = computed(() => {
  return auditReportStore.reportList
    .filter(r => {
      const num = r.reportNumber || (r as any).report_number
      return (
        num &&
        num !== 'DOC-EXSUM-Q1-2026' &&
        num !== '020/LHA/01/KS IAD/2023' &&
        num !== '019/LHA/01/KS IAD/2025' &&
        num !== 'LHA-020/SKAI/2023' &&
        num !== 'LHA-019/SKAI/2025'
      )
    })
    .map(r => ({
      label: `${r.reportNumber || (r as any).report_number} - ${r.reportTitle}`,
      value: r.reportNumber || (r as any).report_number,
      report: r
    }))
})

const selectedLhaId = ref<{ label: string; value: string; report: any } | undefined>(
  lhaDropdownOptions.value.find(o => o.value === store.form.nomorDokumen)
)

// Section III & Top 5 status: follow-up status of the linked LHA from the Action Taken Reports.
// Drafts load it when the LHA is selected or the form is opened, and on Refresh. Approved summaries keep
// the snapshot frozen at approval; their live status is shown read-only in ExecutiveSummaryCurrentFollowUp.
const followUp = useExecutiveSummaryFollowUp()
// Query key of the response in followUp.data, so a response is never applied to another LHA
const followUpLoadedKey = ref('')

const followUpQuery = computed<FollowUpStatusQuery | null>(() => {
  const report = linkedLhaReport.value as { id?: string, reportNumber?: string } | null
  if (report?.id && isValidUuid(report.id)) return { auditResultReportId: report.id }
  const reportNumber = report?.reportNumber || store.form.nomorDokumen
  return reportNumber ? { reportNumber } : null
})
const followUpQueryKey = computed(() => (followUpQuery.value ? JSON.stringify(followUpQuery.value) : ''))

const currentFollowUp = computed<FollowUpStatusResult | null>(() =>
  followUpLoadedKey.value && followUpLoadedKey.value === followUpQueryKey.value ? followUp.data.value : null
)

/** Follow-up findings by id for an LHA, when the loaded status belongs to it and is not restricted; null otherwise. */
function followUpFindingsFor(item: { id?: string, reportNumber?: string }) {
  const data = currentFollowUp.value
  if (!data || data.findingsRestricted) return null
  const belongs = data.findings.length === 0 || data.findings.some(f =>
    (item.id && f.auditResultReportId === item.id) || (item.reportNumber && f.reportNumber === item.reportNumber)
  )
  return belongs ? new Map(data.findings.map(f => [f.findingId, f])) : null
}

const applyFollowUpToForm = (data: FollowUpStatusResult) => {
  store.form.followUpTable = followUpRowsFromStats(data.buckets, store.form.followUpTable)
  if (data.findingsRestricted) return
  applyFollowUpToTopFindings(store.form.topFindings, data.findings)
  // A finding taken from the LHA whose ATR was cancelled is not counted, so it leaves the Top 5
  const ids = new Set(data.findings.map(f => f.findingId))
  store.form.topFindings = store.form.topFindings.filter(f => !f.findingId || ids.has(f.findingId))
}

const refreshFollowUp = async () => {
  if (isLocked.value) return
  const query = followUpQuery.value
  const key = followUpQueryKey.value
  if (!query) return
  const data = await followUp.load(query)
  if (data) {
    followUpLoadedKey.value = key
    // The LHA may have changed while the request was running
    if (key === followUpQueryKey.value && !isLocked.value) applyFollowUpToForm(data)
  } else if (followUp.error.value) {
    errorToast(t('executiveSummaryForm.followUp.errorToast'), followUp.error.value)
  }
}

watch(followUpQueryKey, (key) => {
  if (key && !isLocked.value) refreshFollowUp()
}, { immediate: true })

const followUpAsOfLabel = computed(() => formatSyncedAt(currentFollowUp.value?.asOf, locale.value))
const approvedDateLabel = computed(() =>
  isLocked.value ? formatFollowUpDate(store.form.approvedAt || store.form.signatureTanggal, locale.value) : ''
)
const followUpEmptyHint = computed(() => {
  if (isLocked.value) return ''
  if (!followUpQuery.value) return t('executiveSummaryForm.followUp.selectLha')
  if (currentFollowUp.value && currentFollowUp.value.total === 0) return t('executiveSummaryForm.followUp.empty')
  return ''
})

// A new audit period updates the generated period prefix; free text written by the user is kept.
// Registered after selectedLhaId: the watcher reads it (via linkedLhaReport) as soon as it is created.
watch(narrativePeriod, (period) => {
  if (store.isViewing || store.isEditing || isLocked.value) return
  store.form.narrative = refreshNarrativePeriod(t, store.form.narrative, period)
})

const onAssignmentLetterSelect = (val: any) => {
  if (!val) return
  const selectedNum = typeof val === 'object' ? val.value : val
  store.form.assignmentLetterId = selectedNum
  const found = assignmentLetterDropdownOptions.value.find(o => o.value === selectedNum)
  if (found) {
    selectedAssignmentLetter.value = found
  }

  // Auto-find and link corresponding LHA if exists
  const matchingLha = auditReportStore.reportList.find(r => r.assignmentLetterId === selectedNum)
  if (matchingLha) {
    onLhaSelect(matchingLha.reportNumber)
  }
}

const syncFindingsFromLha = (lhaItem?: any) => {
  const targetDoc = typeof selectedLhaId.value === 'object'
    ? (selectedLhaId.value as any)?.value
    : (selectedLhaId.value || store.form.nomorDokumen)

  const item = lhaItem || auditReportStore.reportList.find(r => 
    (r.reportNumber || (r as any).report_number) === targetDoc
  )
  if (!item) return

  store.form.nomorDokumen = item.reportNumber || (item as any).report_number
  if (item.assignmentLetterId) {
    store.form.assignmentLetterId = item.assignmentLetterId
    const found = assignmentLetterDropdownOptions.value.find(o => o.value === item.assignmentLetterId)
    if (found) selectedAssignmentLetter.value = found
  }
  if (item.executiveSummary) {
    // The narrative starts with the audit period: "Periode Januari 2026, ..."
    store.form.narrative = prefixNarrative(t, item.executiveSummary, auditPeriodForReport(item))
  }

  const findings = item.findings || []
  // Findings whose ATR was cancelled are not in the follow-up status, so they are left out of the Top 5
  const followUpById = followUpFindingsFor(item)
  const candidates = followUpById ? findings.filter((f: any) => !f.id || followUpById.has(f.id)) : findings
  if (findings.length > 0) {
    store.form.topFindings = candidates.slice(0, 5).map((f: any): SignificantFinding => {
      let rawDiv = f.unitDivision || f.unit_division || f.department || item.department || ''
      if (!rawDiv) {
        if (item.reportTitle?.includes('Keuangan')) rawDiv = 'FIN'
        else if (item.reportTitle?.includes('Sistem') || item.reportTitle?.includes('ERP') || item.reportTitle?.includes('TI')) rawDiv = 'IT'
        else if (item.reportTitle?.includes('SDM') || item.reportTitle?.includes('Payroll')) rawDiv = 'HR'
        else if (item.reportTitle?.includes('Procurement') || item.reportTitle?.includes('SCM')) rawDiv = 'KK/KSD'
        else rawDiv = 'OP'
      }

      const div = store.resolveDepartmentOption(rawDiv) || rawDiv

      let sigLevel = f.category || f.significanceLevel || f.risiko || 'Significant'
      if (sigLevel === 'Tinggi') sigLevel = 'Significant'
      else if (sigLevel === 'Sedang') sigLevel = 'Quite Significant'
      else if (sigLevel === 'Rendah') sigLevel = 'Not Significant'

      const atr = f.id ? followUpById?.get(f.id) : undefined
      return {
        unitDivision: div,
        judulTemuan: f.title || f.finding || '',
        risiko: sigLevel,
        // Status of the finding's ATR; until the follow-up status has loaded, the finding counts as not yet followed up
        statusTL: atr ? BUCKET_STATUS[atr.bucket] : 'No Action',
        usulan: f.action || f.recommendation || 'Perbaikan SOP dan Kontrol Pengendalian Internal',
        findingId: f.id || undefined,
        atrId: atr?.atrId ?? null,
        noActionReason: atr?.noActionReason ?? null
      }
    })

    // Finding categories only have three levels: same as the legacy Tinggi/Sedang/Rendah -> High/Moderate/Low
    store.form.risikoHigh = findings.filter((f: any) => ['Very Significant', 'Significant'].includes(f.category)).length
    store.form.risikoModerateHigh = 0
    store.form.risikoModerate = findings.filter((f: any) => f.category === 'Quite Significant').length
    store.form.risikoLowModerate = 0
    store.form.risikoLow = findings.filter((f: any) => f.category === 'Not Significant').length
  }

  if (item.findingsCount) {
    store.form.jumlahRekomendasi = item.findingsCount
  } else if (findings.length > 0) {
    store.form.jumlahRekomendasi = findings.length
  }

  // Section V & VII: Analisis Temuan Berulang & Kesimpulan
  if ((item as any).findingSummary || (item as any).finding_summary) {
    store.form.akarMasalah = (item as any).findingSummary || (item as any).finding_summary
  } else if (findings.length > 0) {
    store.form.akarMasalah = `Analisis temuan berulang mengidentifikasi kelemahan pada kepatuhan SOP dan monitoring: ${findings.slice(0, 2).map((f: any) => f.title).join('; ')}.`
  }

  if ((item as any).conclusion) {
    store.form.kesimpulan = (item as any).conclusion
  } else if (!store.form.kesimpulan && item.reportTitle) {
    store.form.kesimpulan = `Tata kelola dan pengendalian internal pada ${item.reportTitle} berjalan cukup memadai dengan prioritas penyelesaian rekomendasi temuan signifikan.`
  }

  selectedLhaId.value = lhaDropdownOptions.value.find(o => o.value === store.form.nomorDokumen)
}

const onLhaSelect = (val: any) => {
  if (!val) return
  const selectedNum = typeof val === 'object' ? val.value : val
  const item = auditReportStore.reportList.find(r => (r.reportNumber || (r as any).report_number) === selectedNum)
  if (item) {
    syncFindingsFromLha(item)
  }
}

// Watch initial state when opening form
watch(() => store.form.assignmentLetterId, (newVal) => {
  if (newVal && selectedAssignmentLetter.value?.value !== newVal) {
    const found = assignmentLetterDropdownOptions.value.find(o => o.value === newVal)
    if (found) selectedAssignmentLetter.value = found
  }
}, { immediate: true })

watch(() => store.form.nomorDokumen, (newDoc) => {
  if (newDoc) {
    selectedLhaId.value = lhaDropdownOptions.value.find(o => o.value === newDoc)
    if (!store.form.assignmentLetterId) {
      const match = auditReportStore.reportList.find(r => r.reportNumber === newDoc)
      if (match && match.assignmentLetterId) {
        store.form.assignmentLetterId = match.assignmentLetterId
        const found = assignmentLetterDropdownOptions.value.find(o => o.value === match.assignmentLetterId)
        if (found) selectedAssignmentLetter.value = found
      }
    }
    // If findings are empty, auto-sync from LHA
    if (store.form.topFindings.length === 0) {
      const match = auditReportStore.reportList.find(r => r.reportNumber === newDoc)
      if (match) {
        syncFindingsFromLha(match)
      }
    }
  }
}, { immediate: true })

</script>
