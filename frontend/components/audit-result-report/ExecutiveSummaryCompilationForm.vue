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

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
            <UFormField :label="t('executiveSummaryCompilation.reportYear')" required>
              <USelectMenu
                v-model="store.form.tahun"
                :items="[2026, 2025, 2024, 2023]"
                :placeholder="t('executiveSummaryForm.placeholders.year')"
                class="w-full font-semibold"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.quarter')" required>
              <USelectMenu
                v-model="reportingQuarter"
                :items="quarterOptions"
                :placeholder="t('executiveSummaryForm.placeholders.quarter')"
                class="w-full font-semibold"
                :disabled="isLocked"
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.formExtra.documentNumber')" required>
              <UInput
                v-model="store.form.nomorDokumen"
                :placeholder="t('executiveSummaryCompilation.formExtra.documentNumberExample')"
                class="w-full font-bold bg-gray-50 dark:bg-gray-800"
                disabled
              />
            </UFormField>
          </div>

          <!-- Kumpulan Semua Data LHA Sesuai Tahun dan Kuartal -->
          <div class="border border-gray-200 dark:border-gray-800 rounded-xl bg-gray-50/70 dark:bg-gray-800/40 p-5 space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-gray-200 dark:border-gray-700/80">
              <div class="flex items-center gap-3">
                <div class="p-2.5 bg-primary-100 dark:bg-primary-900/60 text-primary-600 dark:text-primary-400 rounded-lg">
                  <UIcon name="i-lucide-files" class="size-5" />
                </div>
                <div>
                  <div class="text-sm font-bold text-gray-900 dark:text-white flex items-center gap-2">
                    <span>{{ t('executiveSummaryCompilation.formExtra.sourcesTitle') }}</span>
                    <UBadge color="primary" variant="subtle" size="xs">
                      {{ reportingQuarter }} {{ store.form.tahun }}
                    </UBadge>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5"> {{ t('executiveSummaryCompilation.formExtra.sourcesDescription') }} </p>
                </div>
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <UBadge color="neutral" variant="outline" size="sm" class="font-mono"> {{ t('executiveSummaryCompilation.formExtra.total') }} {{ quarterlyLhas.length }} {{ t('executiveSummaryCompilation.formExtra.lha') }} </UBadge>
                <UButton
                  v-if="!isLocked"
                  color="primary"
                  variant="ghost"
                  icon="i-lucide-refresh-cw"
                  :label="t('executiveSummaryCompilation.formExtra.refresh')"
                  size="xs"
                  @click="syncFromIndividualLha"
                />
              </div>
            </div>

            <!-- List of LHAs (Table or Empty State) -->
            <div v-if="quarterlyLhas.length > 0" class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 shadow-xs">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800 text-xs">
                <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-600 dark:text-gray-300 font-bold uppercase tracking-wider">
                  <tr>
                    <th class="px-4 py-2.5 text-left">{{ t('executiveSummaryCompilation.matrix.columns.docNumber') }}</th>
                    <th class="px-4 py-2.5 text-left">{{ t('executiveSummaryCompilation.formExtra.assignmentLetter') }}</th>
                    <th class="px-4 py-2.5 text-left">{{ t('executiveSummaryCompilation.matrix.columns.title') }}</th>
                    <th class="px-4 py-2.5 text-left">{{ t('executiveSummaryForm.placeholders.division') }}</th>
                    <th class="px-4 py-2.5 text-left">{{ t('executiveSummaryCompilation.formExtra.reportDate') }}</th>
                    <th class="px-4 py-2.5 text-center">{{ t('executiveSummaryCompilation.findingsUnit') }}</th>
                    <th class="px-4 py-2.5 text-center">{{ t('strategicPlan.columns.status') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800 font-medium">
                  <tr v-for="lha in quarterlyLhas" :key="lha.id || lha.reportNumber" class="hover:bg-gray-50/60 dark:hover:bg-gray-800/50 transition-colors">
                    <td class="px-4 py-2.5 font-mono font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">
                      {{ lha.reportNumber || (lha as any).report_number }}
                    </td>
                    <td class="px-4 py-2.5 font-mono text-gray-600 dark:text-gray-400 whitespace-nowrap">
                      <span v-if="lha.assignmentLetterId" class="inline-flex items-center gap-1">
                        <UIcon name="i-lucide-file-signature" class="size-3 text-primary-500" />
                        {{ lha.assignmentLetterId }}
                      </span>
                      <span v-else class="text-gray-400 italic">-</span>
                    </td>
                    <td class="px-4 py-2.5 text-gray-800 dark:text-gray-200 max-w-xs truncate" :title="lha.reportTitle">
                      {{ lha.reportTitle }}
                    </td>
                    <td class="px-4 py-2.5 text-gray-600 dark:text-gray-400 whitespace-nowrap">
                      <UBadge color="neutral" variant="subtle" size="xs">
                        {{ lha.department || 'OP' }}
                      </UBadge>
                    </td>
                    <td class="px-4 py-2.5 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                      {{ lha.reportDate || '-' }}
                    </td>
                    <td class="px-4 py-2.5 text-center whitespace-nowrap">
                      <UBadge color="warning" variant="subtle" size="xs" class="font-bold">
                        {{ lha.findingsCount || (lha.findings ? lha.findings.length : 0) }} {{ t('executiveSummaryCompilation.findingsUnit') }}
                      </UBadge>
                    </td>
                    <td class="px-4 py-2.5 text-center whitespace-nowrap">
                      <UBadge :color="lha.status === 'Final' || (lha.status as string) === 'Approved' ? 'success' : 'neutral'" variant="soft" size="xs">
                        {{ lha.status }}
                      </UBadge>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Empty State when no LHAs for the selected quarter & year -->
            <div v-else class="text-center py-6 px-4 bg-white dark:bg-gray-900 rounded-lg border border-dashed border-gray-300 dark:border-gray-700 space-y-1">
              <UIcon name="i-lucide-file-x" class="size-8 text-gray-400 mx-auto" />
              <div class="text-xs font-semibold text-gray-700 dark:text-gray-300"> {{ t('executiveSummaryCompilation.formExtra.noDocumentsFor') }} {{ reportingQuarter }} {{ store.form.tahun }}
              </div>
              <p class="text-[11px] text-gray-400"> {{ t('executiveSummaryCompilation.formExtra.emptyDescription') }} </p>
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

        <!-- 3. Section Statistik Kompilasi -->
        <section id="sec-stats" v-show="activeSection === 'sec-stats'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryCompilation.stats.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryCompilation.stats.description') }}</p>
            </div>
            <UButton
              v-if="!isLocked"
              color="primary"
              variant="outline"
              icon="i-lucide-refresh-cw"
              :label="t('executiveSummaryCompilation.syncFromIndividual')"
              size="sm"
              :loading="sourcesLoading"
              @click="syncFromIndividualLha"
            />
          </div>

          <!-- Source status: loading / error with retry / no LHA in the quarter / empty sources / last sync -->
          <div v-if="sourcesLoading" class="flex items-center gap-2 text-sm text-gray-400">
            <UIcon name="i-lucide-loader-circle" class="size-4 animate-spin" />
            {{ t('executiveSummaryCompilation.sources.loading') }}
          </div>
          <div v-else-if="sourcesError" class="flex flex-wrap items-center gap-3 text-sm text-error-500">
            <span>{{ t('executiveSummaryCompilation.sources.error', { message: sourcesError }) }}</span>
            <UButton
              color="error"
              variant="outline"
              size="xs"
              icon="i-lucide-refresh-cw"
              :label="t('executiveSummaryCompilation.sources.retry')"
              @click="syncFromIndividualLha"
            />
          </div>
          <div v-else-if="displayedStats && displayedStats.jumlahLaporan === 0" class="text-sm text-gray-400">
            {{ t('executiveSummaryCompilation.sources.noLhaInQuarter') }}
          </div>
          <p v-if="lastSyncedLabel && !sourcesLoading && !sourcesError" class="text-xs text-gray-400">
            {{ t('executiveSummaryCompilation.sources.lastSynced', { time: lastSyncedLabel }) }}
          </p>
          <ul v-if="sourceEmptyNotes.length && !sourcesLoading && !sourcesError" class="text-sm text-gray-400 space-y-1">
            <li v-for="note in sourceEmptyNotes" :key="note">{{ note }}</li>
          </ul>
          <p v-if="displayedStats && displayedStats.risikoLainnya > 0 && !sourcesLoading" class="text-sm text-warning-600 dark:text-warning-400">
            {{ t('executiveSummaryCompilation.sources.otherRiskNote', { count: displayedStats.risikoLainnya }) }}
          </p>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-6" :class="statsBusyClass">
            <UFormField :label="t('executiveSummaryCompilation.reportCount')" :hint="t('executiveSummaryCompilation.sources.hintLha')">
              <UInput
                type="number"
                :model-value="store.form.jumlahLaporan"
                :placeholder="t('executiveSummaryCompilation.reportCountPlaceholder')"
                class="w-full font-bold"
                disabled
                readonly
              />
            </UFormField>

            <UFormField :label="t('executiveSummaryCompilation.recommendationCount')" :hint="t('executiveSummaryCompilation.sources.hintRecommendation')">
              <UInput
                type="number"
                :model-value="store.form.jumlahRekomendasi"
                :placeholder="t('executiveSummaryCompilation.recommendationCountPlaceholder')"
                class="w-full font-bold"
                disabled
                readonly
              />
            </UFormField>

            <!-- Auto calculated fields -->
            <UFormField :label="t('executiveSummaryCompilation.totalFindings')" :hint="t('executiveSummaryCompilation.totalFindingsHint')">
              <UInput
                :model-value="totalTemuanSummary"
                class="w-full font-bold bg-gray-50 dark:bg-gray-800"
                disabled
                readonly
              >
                <template #trailing>
                  <span class="text-md text-gray-400">{{ t('executiveSummaryCompilation.findingsUnit') }}</span>
                </template>
              </UInput>
            </UFormField>
          </div>

          <div class="p-6 bg-gray-50 dark:bg-gray-800/40 rounded-xl border border-gray-200 dark:border-gray-800 space-y-4" :class="statsBusyClass">
            <div class="flex justify-between items-center">
              <h4 class="text-md font-extrabold uppercase tracking-wider text-gray-400">{{ t('executiveSummaryCompilation.riskBreakdown') }}</h4>
              <span class="text-xs text-gray-400">{{ t('executiveSummaryCompilation.sources.hintRisk') }}</span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <UFormField
                v-for="bucket in riskBucketFields"
                :key="bucket.key"
                :label="t(bucket.label)"
              >
                <UInput
                  type="number"
                  :model-value="riskBuckets[bucket.key]"
                  class="w-full border-l-4 font-bold"
                  :class="bucket.borderClass"
                  disabled
                  readonly
                />
              </UFormField>
            </div>
          </div>
        </section>

        <!-- 4. Section Status Tindak Lanjut -->
        <section id="sec-followup" v-show="activeSection === 'sec-followup'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4">
            <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
              {{ t('executiveSummaryForm.followUp.title') }}
            </h2>
            <p class="text-sm text-gray-400">{{ t('executiveSummaryForm.followUp.description') }}</p>
            <p class="text-xs text-gray-400 mt-1">{{ t('executiveSummaryCompilation.sources.hintFollowUp') }}</p>
          </div>

          <!-- Counts from the ATR buckets of every finding in the quarter (read-only); frozen once approved -->
          <ExecutiveSummaryFollowUpTable
            :rows="store.form.followUpTable"
            :locked="isLocked"
            :loading="followUp.loading.value"
            :error="followUp.error.value"
            :approved-date-label="approvedDateLabel"
            :as-of-label="followUpAsOfLabel"
            :refreshable="!isLocked"
            :empty-hint="followUpEmptyHint"
            @refresh="refreshFollowUpStatus"
          />

          <!-- Live status next to the frozen snapshot, only for users who may view every ATR -->
          <ExecutiveSummaryCurrentFollowUp
            v-if="isLocked && canViewLiveFollowUpStatus"
            :query="followUpQuery"
            show-report-column
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
                  @click="syncFromIndividualLha"
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

        <!-- 6. Qualitative Analysis (Section V & VII) -->
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
              @click="syncFromIndividualLha"
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

        <!-- 7. Section Matriks Kompilasi LHA Individual -->
        <section id="sec-matriks" v-show="activeSection === 'sec-matriks'" class="space-y-6 scroll-mt-6">
          <div class="border-b border-gray-200 dark:border-gray-800 pb-4 flex justify-between items-center">
            <div>
              <h2 class="text-xl font-extrabold text-gray-900 dark:text-white flex items-center gap-2">
                {{ t('executiveSummaryCompilation.matrix.title') }}
              </h2>
              <p class="text-sm text-gray-400">{{ t('executiveSummaryCompilation.matrix.description') }}</p>
            </div>
            <UBadge color="primary" variant="subtle" size="md">
              {{ matrixTotalLabel }}
            </UBadge>
          </div>

          <p v-if="!isLocked && currentFollowUp?.findingsRestricted" class="text-sm text-gray-400 flex items-center gap-1.5">
            <UIcon name="i-lucide-eye-off" class="size-4" />
            {{ t('executiveSummaryCompilation.matrix.restricted') }}
          </p>

          <div v-if="store.form.matriksKompilasi.length > 0" class="border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-800">
              <thead class="bg-gray-50 dark:bg-gray-800/80">
                <tr>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.docNumber') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.unit') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.title') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.risk') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.recommendation') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.dueDate') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.pic') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.progress') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.evidence') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-bold text-gray-500 dark:text-white uppercase tracking-wider">{{ t('executiveSummaryCompilation.matrix.columns.followUpStatus') }}</th>
                </tr>
              </thead>
              <tbody class="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-800 text-sm">
                <tr v-for="(row, idx) in store.form.matriksKompilasi" :key="row.findingId || idx">
                  <td class="px-4 py-3 font-mono font-semibold text-primary-600 dark:text-primary-400 whitespace-nowrap">{{ row.nomor }}</td>
                  <td class="px-4 py-3 font-medium text-gray-800 dark:text-gray-200">{{ row.division ? store.formatDepartmentName(row.division) : '-' }}</td>
                  <td class="px-4 py-3 text-gray-700 dark:text-gray-300 max-w-xs truncate" :title="row.judulTemuan">{{ row.judulTemuan || '-' }}</td>
                  <td class="px-4 py-3">
                    <UBadge v-if="row.nilaiRisiko" :color="matrixRiskColor(row.nilaiRisiko)" variant="subtle" size="xs">
                      {{ matrixRiskLabel(row.nilaiRisiko) }}
                    </UBadge>
                    <span v-else class="text-gray-400">-</span>
                  </td>
                  <td class="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate" :title="row.rekomendasi">{{ row.rekomendasi || '-' }}</td>
                  <td class="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">{{ formatFollowUpDate(row.dueDate, locale) || row.dueDate || '-' }}</td>
                  <td class="px-4 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">{{ row.picUnit || '-' }}</td>
                  <td class="px-4 py-3 text-right tabular-nums text-gray-600 dark:text-gray-400">{{ Number(row.progres) || 0 }}%</td>
                  <td class="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-[12rem] truncate" :title="row.buktiTL">{{ row.buktiTL || '-' }}</td>
                  <td class="px-4 py-3">
                    <FollowUpStatusBadge :status="row.status || 'No Action'" :reason="row.noActionReason" />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else class="text-center py-8 bg-gray-50 dark:bg-gray-800/30 rounded-lg text-sm text-gray-400 border border-dashed">
            {{ t('executiveSummaryCompilation.matrix.empty') }}
            <button v-if="!isLocked" type="button" @click="syncFromIndividualLha" class="text-primary-500 font-bold ml-1 hover:underline">
              {{ t('executiveSummaryCompilation.matrix.emptyAction') }}
            </button>
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
            :loading="store.loading || sourcesLoading"
            :disabled="!canSaveDraft"
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
import { ref, computed, watch, onMounted } from 'vue'
import { useExecutiveSummaryStore, type SignificantFinding } from '~/stores/executive-summary'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useI18n } from '~/composables/useI18n'
import { useAppToast } from '~/composables/useAppToast'
import { useExecutiveSummarySources } from '~/composables/useExecutiveSummarySources'
import { useExecutiveSummaryFollowUp } from '~/composables/useExecutiveSummaryFollowUp'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import ExecutiveSummaryApproveModal from './ExecutiveSummaryApproveModal.vue'
import ExecutiveSummaryCurrentFollowUp from './ExecutiveSummaryCurrentFollowUp.vue'
import ExecutiveSummaryFollowUpTable from './ExecutiveSummaryFollowUpTable.vue'
import FollowUpStatusBadge from './FollowUpStatusBadge.vue'
import SignificantFindingCards from './SignificantFindingCards.vue'
import { scrollStepIntoView } from '~/utils/stepperScroll'
import {
  followUpRowsFromStats,
  formatSyncedAt,
  riskBreakdownToFields,
  riskBucketBreakdown,
  riskBucketTotal,
  riskFieldsToBreakdown,
  type CompilationStats,
  type RiskBucket
} from '~/utils/executiveSummaryCompilation'
import { formatQuarterPeriod, refreshNarrativePeriod } from '~/utils/executiveSummaryNarrative'
import {
  BUCKET_STATUS,
  applyFollowUpToTopFindings,
  formatFollowUpDate,
  matrixRowsFromFollowUp,
  type FollowUpStatusQuery,
  type FollowUpStatusResult,
  type LhaFindingInfo
} from '~/utils/executiveSummaryFollowUp'
const store = useExecutiveSummaryStore()
const authStore = useAuthStore()
const auditReportStore = useAuditResultReportStore()
const assignmentLetterStore = useAssignmentLetterStore()
const { canReviewExecutiveSummary, canViewLiveFollowUpStatus } = useRbac()
const { t, locale } = useI18n()
const appToast = useAppToast()
const { loading: sourcesLoading, error: sourcesError, fetchCompilationStats } = useExecutiveSummarySources()
// Follow-up status of every finding in the selected year and quarter, from the Action Taken Reports
const followUp = useExecutiveSummaryFollowUp()
const fileInput = ref<HTMLInputElement | null>(null)
const activeSection = ref('sec-upload')
const scrollContainer = ref<HTMLElement | null>(null)

// Last successful statistics response, and the 'year-quarter' key it was read for ('' after a failed sync)
const lastStats = ref<CompilationStats | null>(null)
const syncedStatsKey = ref('')
const currentStatsKey = computed(() => `${Number(store.form.tahun) || 2026}-${Number(store.form.quarter) || 1}`)
// Statistics are shown only when they belong to the selected year and quarter
const displayedStats = computed<CompilationStats | null>(() =>
  lastStats.value && syncedStatsKey.value === currentStatsKey.value ? lastStats.value : null
)
const lastSyncedLabel = computed(() => formatSyncedAt(displayedStats.value?.syncedAt, locale.value))

// Follow-up status: the response in followUp.data and the 'year-quarter' key it was read for
const followUpLoadedKey = ref('')
const currentFollowUp = computed<FollowUpStatusResult | null>(() =>
  followUpLoadedKey.value === currentStatsKey.value ? followUp.data.value : null
)
const followUpQuery = computed<FollowUpStatusQuery>(() => ({
  year: Number(store.form.tahun) || 2026,
  quarter: Number(store.form.quarter) || 1
}))
const followUpAsOfLabel = computed(() => formatSyncedAt(currentFollowUp.value?.asOf, locale.value))
const approvedDateLabel = computed(() =>
  isLocked.value ? formatFollowUpDate(store.form.approvedAt || store.form.signatureTanggal, locale.value) : ''
)
const followUpEmptyHint = computed(() =>
  !isLocked.value && currentFollowUp.value && currentFollowUp.value.total === 0 ? t('executiveSummaryCompilation.sources.emptyFollowUp') : ''
)
// Rows saved by the old per-LHA matrix carry no findingId and are counted as LHA
const matrixTotalLabel = computed(() => {
  const rows = store.form.matriksKompilasi
  if (rows.length > 0 && rows.every(r => !r.findingId)) return t('executiveSummaryCompilation.matrix.total', { count: rows.length })
  return t('executiveSummaryCompilation.matrix.totalFindings', { count: rows.length })
})

// Saving a draft is blocked while the statistics are loading or the last sync failed
const canSaveDraft = computed(() => !sourcesLoading.value && !sourcesError.value)
const statsBusyClass = computed(() => (sourcesLoading.value ? 'opacity-50 pointer-events-none' : ''))

// Display buckets of the risk breakdown (High / Medium / Low), read-only.
// The five audit-universe levels are mapped to buckets in utils (RISK_BUCKET_BY_LEVEL).
const riskBucketFields: { key: RiskBucket, label: string, borderClass: string }[] = [
  { key: 'high', label: 'executiveSummaryCompilation.riskBuckets.high', borderClass: 'border-error-500' },
  { key: 'medium', label: 'executiveSummaryCompilation.riskBuckets.medium', borderClass: 'border-warning-500' },
  { key: 'low', label: 'executiveSummaryCompilation.riskBuckets.low', borderClass: 'border-success-500' }
]
const riskBuckets = computed(() => riskBucketBreakdown(riskFieldsToBreakdown(store.form)))

// Empty-source hints shown under the statistics section after a successful sync
const sourceEmptyNotes = computed<string[]>(() => {
  const stats = displayedStats.value
  if (!stats) return []
  const notes: string[] = []
  if (stats.jumlahRekomendasi === 0) notes.push(t('executiveSummaryCompilation.sources.emptyRecommendation'))
  if (stats.totalTemuan === 0) notes.push(t('executiveSummaryCompilation.sources.emptyRisk'))
  const fu = currentFollowUp.value?.buckets ?? stats.followUp
  if (fu && fu.closed + fu.inProgress + fu.overdue + (fu.noAction ?? 0) === 0) {
    notes.push(t('executiveSummaryCompilation.sources.emptyFollowUp'))
  }
  return notes
})

if (!auditReportStore.loading && auditReportStore.reportList.length === 0) {
  auditReportStore.fetchReports()
}
if (!assignmentLetterStore.loading && assignmentLetterStore.assignmentLetterList.length === 0) {
  assignmentLetterStore.fetchAssignmentLetters()
}

// Navigation links
const sections = computed(() => [
  { id: 'sec-upload', index: '1', title: t('executiveSummaryCompilation.nav.upload') },
  { id: 'sec-narrative', index: '2', title: t('executiveSummaryCompilation.nav.narrative') },
  { id: 'sec-stats', index: '3', title: t('executiveSummaryCompilation.nav.stats') },
  { id: 'sec-followup', index: '4', title: t('executiveSummaryCompilation.nav.followUp') },
  { id: 'sec-topfindings', index: '5', title: t('executiveSummaryCompilation.nav.topFindings') },
  { id: 'sec-analysis', index: '6', title: t('executiveSummaryCompilation.nav.analysis') },
  { id: 'sec-matriks', index: '7', title: t('executiveSummaryCompilation.nav.matrix') },
  ...(store.isViewing ? [{ id: 'sec-notes', index: '8', title: t('executiveSummaryCompilation.nav.notes') }] : [])
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

const quarterOptions = ['Kuartal I', 'Kuartal II', 'Kuartal III', 'Kuartal IV']
const reportingQuarter = computed({
  get: (): string => quarterOptions[store.form.quarter - 1] ?? 'Kuartal I',
  set: (value: string) => {
    const quarterIndex = quarterOptions.indexOf(value)
    store.form.quarter = quarterIndex >= 0 ? quarterIndex + 1 : 1
    store.form.periodeBulan = quarterOptions[store.form.quarter - 1] ?? 'Kuartal I'
  }
})

// Role computed checks: Admin, CAE, and Audit Manager can write note & approve draft
const canWriteExecutiveNote = computed(() => Boolean(canReviewExecutiveSummary.value))
const canApprove = computed(() => Boolean(canReviewExecutiveSummary.value))
const isChiefAuditExecutive = computed(() => Boolean(canReviewExecutiveSummary.value))
const isDraft = computed(() => String(store.form.status || '').toLowerCase() === 'draft')

const isLocked = computed(() => {
  return store.form.status === 'Approved'
})

// Helper to generate compilation document number: ESC-[nomor urut]/[tahun]
const generateEscDocNumber = (targetYear: number = 2026): string => {
  if (typeof (store as any).generateEscDocNumber === 'function') {
    return (store as any).generateEscDocNumber(targetYear)
  }
  let maxSeq = 0
  const escRegex = new RegExp(`^ESC-(\\d+)/${targetYear}$`, 'i')
  for (const s of store.summaryList || []) {
    if (s.nomorDokumen) {
      const match = s.nomorDokumen.match(escRegex)
      if (match && match[1]) {
        const val = parseInt(match[1], 10)
        if (!isNaN(val) && val > maxSeq) maxSeq = val
      }
    }
  }
  const nextSeq = (maxSeq + 1).toString().padStart(3, '0')
  return `ESC-${nextSeq}/${targetYear}`
}

// Helper to determine year and quarter of an individual LHA
const getReportQuarterAndYear = (r: any, letters: any[] = []): { year: number; quarter: number } => {
  let year = 0
  let quarter = 0
  const dateStr = r.reportDate || (r as any).report_date || ''
  if (dateStr) {
    const parts = dateStr.split('T')[0].split('-')
    if (parts[0] && /^\d{4}$/.test(parts[0])) {
      year = parseInt(parts[0], 10)
    }
    if (parts[1]) {
      const month = parseInt(parts[1], 10)
      if (month >= 1 && month <= 12) {
        quarter = Math.ceil(month / 3)
      }
    }
  }

  // Linked Assignment Letter info
  if (r.assignmentLetterId) {
    const st = letters.find(s => s.letterNumber === r.assignmentLetterId)
    if (st) {
      if (!year && st.auditYear) {
        const yMatch = String(st.auditYear).match(/\b(20\d{2})\b/)
        if (yMatch) year = parseInt(yMatch[1] ?? '0', 10)
      }
      if (!quarter && st.startPeriod) {
        const parts = st.startPeriod.split('T')[0].split('-')
        if (parts[1]) {
          const m = parseInt(parts[1], 10)
          if (m >= 1 && m <= 12) quarter = Math.ceil(m / 3)
        }
      }
      if (!quarter && st.executionPeriod) {
        if (/Q1|Kuartal I\b|Triwulan I\b/i.test(st.executionPeriod)) quarter = 1
        else if (/Q2|Kuartal II\b|Triwulan II\b/i.test(st.executionPeriod)) quarter = 2
        else if (/Q3|Kuartal III\b|Triwulan III\b/i.test(st.executionPeriod)) quarter = 3
        else if (/Q4|Kuartal IV\b|Triwulan IV\b/i.test(st.executionPeriod)) quarter = 4
      }
    }
  }

  if (!year) {
    const numMatch = (r.reportNumber || (r as any).report_number || '').match(/\b(20\d{2})\b/)
    if (numMatch) {
      year = parseInt(numMatch[1], 10)
    } else {
      year = 2026
    }
  }
  if (!quarter) {
    quarter = 1
  }
  return { year, quarter }
}

// All LHAs corresponding to the selected Tahun Laporan and Periode Kuartal
const quarterlyLhas = computed(() => {
  const targetYear = Number(store.form.tahun) || 2026
  const targetQuarter = Number(store.form.quarter) || 1
  const validReports = auditReportStore.reportList.filter(r => {
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
  return validReports.filter(r => {
    const { year, quarter } = getReportQuarterAndYear(r, assignmentLetterStore.assignmentLetterList)
    return year === targetYear && quarter === targetQuarter
  })
})

// Section II total: totalTemuan from the backend after a sync; the stored risk fields otherwise (saved or approved summaries)
const totalTemuanSummary = computed(() => displayedStats.value?.totalTemuan ?? riskBucketTotal(riskBuckets.value))

// The narrative starts with the selected quarter and year: "Periode Kuartal 4 2026, SPI telah menerbitkan ..."
const narrativePeriod = computed(() => formatQuarterPeriod(t, Number(store.form.quarter) || 1, Number(store.form.tahun) || null))

const resetNarrativeToDefault = () => {
  store.form.narrative = store.defaultNarrativeTemplate(narrativePeriod.value)
}

// A new quarter or year updates the generated period prefix; free text written by the user is kept.
watch(narrativePeriod, (period) => {
  if (store.isViewing || store.isEditing || isLocked.value) return
  store.form.narrative = refreshNarrativePeriod(t, store.form.narrative, period)
})

// File Upload Handler (with corrupted file simulation & validation)
const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  // Size limit 10MB
  if (file.size > 10 * 1024 * 1024) {
    appToast.warning(t('executiveSummaryForm.messages.fileTooLarge'))
    return
  }

  // Type limit
  const ext = file.name.split('.').pop()?.toLowerCase()
  if (ext !== 'pdf' && ext !== 'docx' && ext !== 'doc') {
    appToast.warning(t('executiveSummaryForm.messages.unsupportedFormat'))
    return
  }

  // Simulated file corruption error handling
  if (file.name.toLowerCase().includes('corrupt') || file.name.toLowerCase().includes('damaged')) {
    appToast.warning(t('executiveSummaryForm.messages.corruptedFile'))
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

// Matrix risk badge: LHA significance levels, plus the Tinggi/Sedang/Rendah of older documents
const matrixRiskColor = (value: string) => {
  if (['Very Significant', 'Significant', 'Tinggi'].includes(value)) return 'error'
  if (['Quite Significant', 'Sedang'].includes(value)) return 'warning'
  return 'success'
}
const matrixRiskLabel = (value: string) => significanceLevelOptions.value.find(o => o.value === value)?.label || value

const normalizeSignificance = (value?: string): string => {
  if (value === 'Tinggi') return 'Significant'
  if (value === 'Sedang') return 'Quite Significant'
  if (value === 'Rendah') return 'Not Significant'
  return value || ''
}
const SEVERITY_RANK: Record<string, number> = { 'Very Significant': 4, Significant: 3, 'Quite Significant': 2, 'Not Significant': 1 }

// LHA details of a finding (significance, recommendation, audit object) by finding id, from the loaded LHAs
const lhaFindingInfo = computed(() => {
  const map = new Map<string, LhaFindingInfo>()
  for (const r of auditReportStore.reportList) {
    for (const f of (r.findings || []) as any[]) {
      if (!f?.id) continue
      map.set(f.id, {
        category: normalizeSignificance(f.category || f.significanceLevel || f.risiko),
        action: f.action || f.recommendation || '',
        reportTitle: r.reportTitle || '',
        department: r.department || ''
      })
    }
  }
  return map
})

/** Matrix: one row per finding of the quarter, with the status, progress, due date, PIC and evidence of its ATR. */
const buildMatrix = (data: FollowUpStatusResult) =>
  matrixRowsFromFollowUp(data, id => lhaFindingInfo.value.get(id))

/** Top 5: the most significant findings of the quarter, with the status of their ATR. */
const buildTopFindings = (data: FollowUpStatusResult): SignificantFinding[] => {
  const ranked = data.findings.map((f) => {
    const info = lhaFindingInfo.value.get(f.findingId)
    const risiko = info?.category || 'Significant'
    const division = f.divisi || info?.department || ''
    return {
      rank: SEVERITY_RANK[risiko] ?? 0,
      row: {
        unitDivision: division ? store.resolveDepartmentOption(division) : '',
        judulTemuan: f.title,
        risiko,
        statusTL: BUCKET_STATUS[f.bucket],
        usulan: info?.action || '',
        findingId: f.findingId,
        atrId: f.atrId,
        noActionReason: f.noActionReason
      } as SignificantFinding
    }
  })
  ranked.sort((a, b) => b.rank - a.rank)
  return ranked.slice(0, 5).map(r => r.row)
}

const notifyFollowUpError = () => {
  if (followUp.error.value) appToast.error(t('executiveSummaryForm.followUp.errorToast'), followUp.error.value)
}

/**
 * Refreshes only the ATR-driven parts of a draft (follow-up table, matrix, Top 5 statuses) without
 * touching the Top 5 selection or the text the user wrote. Used on open of a saved draft and by Refresh.
 */
const refreshFollowUpStatus = async () => {
  if (isLocked.value) return
  const key = currentStatsKey.value
  const data = await followUp.load(followUpQuery.value)
  if (!data) {
    notifyFollowUpError()
    return
  }
  followUpLoadedKey.value = key
  if (key !== currentStatsKey.value || isLocked.value) return
  store.form.followUpTable = followUpRowsFromStats(data.buckets, store.form.followUpTable)
  if (data.findingsRestricted) return
  store.form.matriksKompilasi = buildMatrix(data)
  applyFollowUpToTopFindings(store.form.topFindings, data.findings)
  // A finding whose ATR was cancelled is not counted, so it leaves the Top 5
  const ids = new Set(data.findings.map(f => f.findingId))
  store.form.topFindings = store.form.topFindings.filter(f => !f.findingId || ids.has(f.findingId))
}

// Automatic compilation synchronization for the selected year and quarter.
// The statistics come from the backend in one call and are written to the form only when the whole response is valid.
// Approved summaries keep their saved values and are not synchronized again.
const syncFromIndividualLha = async () => {
  const targetYear = Number(store.form.tahun) || 2026
  const targetQuarter = Number(store.form.quarter) || 1

  // Auto-set nomorDokumen if empty or follows legacy template
  if (!store.isEditing && !store.isViewing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(targetYear)
    }
  }

  if (isLocked.value) return

  // 1. Statistics (read-only) and the ATR follow-up status of the quarter, from the backend
  const key = `${targetYear}-${targetQuarter}`
  const [stats, followUpData] = await Promise.all([
    fetchCompilationStats({ year: targetYear, quarter: targetQuarter }),
    followUp.load({ year: targetYear, quarter: targetQuarter })
  ])
  if (stats) {
    lastStats.value = stats
    syncedStatsKey.value = key
    Object.assign(store.form, riskBreakdownToFields(stats.riskBreakdown))
    store.form.jumlahLaporan = stats.jumlahLaporan
    store.form.jumlahRekomendasi = stats.jumlahRekomendasi
  } else if (sourcesError.value) {
    // A failed sync drops the statistics, so saving stays blocked until a sync succeeds.
    lastStats.value = null
    syncedStatsKey.value = ''
  } else {
    // Superseded by a newer sync, which fills the rest of the form
    return
  }

  // 2. Follow-up table, matrix (one row per finding) and Top 5 from the ATR follow-up status.
  // On failure they keep their current values; the error is shown with a retry.
  if (followUpData) {
    followUpLoadedKey.value = key
    store.form.followUpTable = followUpRowsFromStats(followUpData.buckets, store.form.followUpTable)
    if (!followUpData.findingsRestricted) {
      store.form.matriksKompilasi = buildMatrix(followUpData)
      store.form.topFindings = buildTopFindings(followUpData)
    }
  } else if (followUp.error.value) {
    notifyFollowUpError()
    // Older follow-up counts from the statistics, when the backend sent a usable table
    if (stats?.followUp) store.form.followUpTable = followUpRowsFromStats(stats.followUp, store.form.followUpTable)
  }
  // Narrative, root cause and conclusion are written by the user; the sync does not generate text.
}

// Form submission & workflow helpers
const saveReportDraft = async () => {
  // Blocked while the statistics are loading or the last sync failed
  if (!canSaveDraft.value) return
  if (!store.form.nomorDokumen) {
    store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
  }
  if (!store.form.dokumenPath) {
    appToast.warning(t('executiveSummaryForm.messages.documentRequired'))
    return
  }

  // Set status Draft
  store.form.periodeBulan = reportingQuarter.value
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


// Watch quarter & tahun changes to automatically sync data
watch(() => [store.form.quarter, store.form.tahun], ([newQ, newY], [oldQ, oldY]) => {
  if (!store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || (oldY && store.form.nomorDokumen.endsWith(`/${oldY}`) && newY !== oldY)) {
      store.form.nomorDokumen = generateEscDocNumber(Number(newY) || 2026)
    }
    syncFromIndividualLha()
  }
})

// Watch modal open
watch(() => store.showModal, (isOpen) => {
  if (isOpen && !store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
    }
  }
})

// Watch reportList when reports are fetched asynchronously
watch(() => auditReportStore.reportList.length, (newLen) => {
  if (newLen > 0 && !store.isViewing && !store.isEditing) {
    if (store.form.jumlahLaporan === 0 || store.form.matriksKompilasi.length === 0) {
      syncFromIndividualLha()
      return
    }
  }
  // LHA details (significance, recommendation) complete the matrix rows once the LHAs are loaded
  const data = currentFollowUp.value
  if (newLen > 0 && data && !data.findingsRestricted && !isLocked.value) {
    store.form.matriksKompilasi = buildMatrix(data)
  }
})

onMounted(() => {
  if (!store.isViewing && !store.isEditing) {
    if (!store.form.nomorDokumen || store.form.nomorDokumen.startsWith('EXSUM-KOMP-Q') || store.form.nomorDokumen.startsWith('KOMPILASI-LHA-Q')) {
      store.form.nomorDokumen = generateEscDocNumber(Number(store.form.tahun) || 2026)
    }
    syncFromIndividualLha()
  } else if (!isLocked.value) {
    // A saved draft shows the current ATR status when it is opened
    refreshFollowUpStatus()
  }
})
</script>
