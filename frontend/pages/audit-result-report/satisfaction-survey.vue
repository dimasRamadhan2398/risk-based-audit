<template>
  <div>
    <!-- Top Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('satisfactionSurvey.title') }}</h1>
        <p class="text-gray-500 dark:text-gray-400">{{ t('satisfactionSurvey.subtitle') }}</p>
      </div>
      <UButton
        :icon="showExplanationGuide ? 'i-heroicons-eye-slash' : 'i-heroicons-information-circle'"
        :label="showExplanationGuide ? t('satisfactionSurvey.hideExplanation') : t('satisfactionSurvey.showExplanation')"
        color="neutral"
        variant="outline"
        size="sm"
        class="font-semibold shadow-xs"
        @click="showExplanationGuide = !showExplanationGuide"
      />
    </div>

    <!-- CSAT Summary Analytics Cards (Design System Primary, Secondary, Success Fill Color) -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
      <!-- 1. CSAT Average Card (Primary Fill Color) -->
      <UCard
        :ui="{ body: 'p-6 text-white dark:text-white', background: 'bg-transparent' }"
        class="relative overflow-hidden bg-primary-600 dark:bg-primary-600 text-white dark:text-white shadow-lg rounded-2xl border-0"
      >
        <div class="relative z-10 flex flex-col justify-between h-full">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-bold uppercase tracking-wider text-white">
                  {{ t('satisfactionSurvey.cards.avgCsat.title') }}
                </span>
                <UTooltip :text="t('satisfactionSurvey.cards.avgCsat.tooltip')">
                  <UIcon name="i-heroicons-information-circle" class="size-4 text-white hover:text-white/80 cursor-pointer transition-colors" />
                </UTooltip>
              </div>
              <h3 class="text-4xl font-black text-white mt-2 tracking-tight">
                {{ averageCsat.toFixed(1) }} <span class="text-lg font-medium text-white">/ 5.0</span>
              </h3>
            </div>
            <div class="p-3.5 bg-white/20 rounded-2xl border border-white/25 backdrop-blur-sm shadow-inner shrink-0 text-white">
              <UIcon name="i-heroicons-face-smile" class="size-8 text-white" />
            </div>
          </div>

          <div class="mt-4 pt-3 border-t border-white/20 flex flex-wrap items-center justify-between gap-2">
            <span class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-white/20 backdrop-blur-sm text-white">
              <UIcon name="i-heroicons-sparkles" class="size-3.5 text-white" />
              {{ t('satisfactionSurvey.cards.avgCsat.target') }}
            </span>
            <span class="text-xs text-white font-medium">
              {{ averageCsat >= 4.5 ? t('satisfactionSurvey.cards.avgCsat.targetMet') : t('satisfactionSurvey.cards.avgCsat.targetNotMet') }}
            </span>
          </div>
          <p class="text-[11px] !text-white mt-2 leading-relaxed">
            {{ t('satisfactionSurvey.cards.avgCsat.description') }}
          </p>
        </div>

        <!-- Decorative background glow -->
        <div class="absolute -bottom-8 -right-8 w-36 h-36 bg-white/10 rounded-full blur-2xl pointer-events-none"></div>
      </UCard>

      <!-- 2. Total Surveys Card (Secondary Fill Color) -->
      <UCard
        :ui="{ body: 'p-6 text-white dark:text-white', background: 'bg-transparent' }"
        class="relative overflow-hidden bg-secondary-600 dark:bg-secondary-600 text-white dark:text-white shadow-lg rounded-2xl border-0"
      >
        <div class="relative z-10 flex flex-col justify-between h-full">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-bold uppercase tracking-wider text-white">
                  {{ t('satisfactionSurvey.cards.totalSurveys.title') }}
                </span>
                <UTooltip :text="t('satisfactionSurvey.cards.totalSurveys.tooltip')">
                  <UIcon name="i-heroicons-information-circle" class="size-4 text-white hover:text-white/80 cursor-pointer transition-colors" />
                </UTooltip>
              </div>
              <h3 class="text-4xl font-black text-white mt-2 tracking-tight">
                {{ surveysStore.surveys.length }}
                <span class="text-lg font-normal text-white">{{ t('satisfactionSurvey.cards.totalSurveys.unit') }}</span>
              </h3>
            </div>
            <div class="p-3.5 bg-white/20 rounded-2xl border border-white/25 backdrop-blur-sm shadow-inner shrink-0 text-white">
              <UIcon name="i-heroicons-document-check" class="size-8 text-white" />
            </div>
          </div>

          <div class="mt-4 pt-3 border-t border-white/20 flex flex-wrap items-center justify-between gap-2">
            <span class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-white/20 backdrop-blur-sm text-white">
              <UIcon name="i-heroicons-document-text" class="size-3.5 text-white" />
              {{ t('satisfactionSurvey.cards.totalSurveys.publishedCount', { count: publishedReports.length }) }}
            </span>
            <span class="text-xs text-white font-medium">
              {{ t('satisfactionSurvey.cards.totalSurveys.pendingCount', { count: Math.max(0, publishedReports.length - surveysStore.surveys.length) }) }}
            </span>
          </div>
          <p class="text-[11px] !text-white mt-2 leading-relaxed">
            {{ t('satisfactionSurvey.cards.totalSurveys.description') }}
          </p>
        </div>

        <!-- Decorative background glow -->
        <div class="absolute -bottom-8 -right-8 w-36 h-36 bg-white/10 rounded-full blur-2xl pointer-events-none"></div>
      </UCard>

      <!-- 3. Response Rate Card (Success Fill Color) -->
      <UCard
        :ui="{ body: 'p-6 text-white dark:text-white', background: 'bg-transparent' }"
        class="relative overflow-hidden bg-success-600 dark:bg-success-600 text-white dark:text-white shadow-lg rounded-2xl border-0"
      >
        <div class="relative z-10 flex flex-col justify-between h-full">
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-bold uppercase tracking-wider text-white">
                  {{ t('satisfactionSurvey.cards.responseRate.title') }}
                </span>
                <UTooltip :text="t('satisfactionSurvey.cards.responseRate.tooltip')">
                  <UIcon name="i-heroicons-information-circle" class="size-4 text-white hover:text-white/80 cursor-pointer transition-colors" />
                </UTooltip>
              </div>
              <h3 class="text-4xl font-black text-white mt-2 tracking-tight">
                {{ responseRate }}%
              </h3>
            </div>
            <div class="p-3.5 bg-white/20 rounded-2xl border border-white/25 backdrop-blur-sm shadow-inner shrink-0 text-white">
              <UIcon name="i-heroicons-chart-bar-solid" class="size-8 text-white" />
            </div>
          </div>

          <!-- White Progress Bar -->
          <div class="mt-3">
            <div class="w-full bg-white/25 rounded-full h-2.5 overflow-hidden backdrop-blur-sm">
              <div
                class="bg-white h-full rounded-full transition-all duration-500 shadow-sm"
                :style="{ width: `${Math.min(100, Math.max(0, responseRate))}%` }"
              ></div>
            </div>
          </div>

          <div class="mt-4 pt-3 border-t border-white/20 flex flex-wrap items-center justify-between gap-2">
            <span class="inline-flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-white/20 backdrop-blur-sm text-white">
              <UIcon name="i-heroicons-flag" class="size-3.5 text-white" />
              {{ t('satisfactionSurvey.cards.responseRate.target') }}
            </span>
            <span class="text-xs text-white font-medium">
              {{ responseRate >= 80 ? t('satisfactionSurvey.cards.responseRate.statusHigh') : t('satisfactionSurvey.cards.responseRate.statusLow') }}
            </span>
          </div>
          <p class="text-[11px] !text-white mt-2 leading-relaxed">
            {{ t('satisfactionSurvey.cards.responseRate.description') }}
          </p>
        </div>

        <!-- Decorative background glow -->
        <div class="absolute -bottom-8 -right-8 w-36 h-36 bg-white/10 rounded-full blur-2xl pointer-events-none"></div>
      </UCard>
    </div>

    <!-- CSAT Metrics Explanation Guide Banner -->
    <UCard
      v-if="showExplanationGuide"
      class="mb-8 border border-primary-200 dark:border-primary-900/40 bg-primary-50/50 dark:bg-primary-950/20 shadow-xs rounded-2xl overflow-hidden"
    >
      <div class="flex items-start justify-between gap-4 pb-4 border-b border-primary-100 dark:border-primary-900/30">
        <div class="flex items-center gap-3">
          <div class="p-2.5 bg-primary-600 text-white rounded-xl shadow-xs">
            <UIcon name="i-heroicons-academic-cap" class="size-5" />
          </div>
          <div>
            <h3 class="text-base font-bold text-gray-900 dark:text-white">
              {{ t('satisfactionSurvey.guide.title') }}
            </h3>
          </div>
        </div>
        <UButton
          icon="i-heroicons-x-mark"
          color="neutral"
          variant="ghost"
          size="xs"
          @click="showExplanationGuide = false"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 pt-4 text-xs">
        <!-- Metric 1 Explanation (Primary) -->
        <div class="p-4 bg-white dark:bg-gray-900 rounded-xl border border-gray-200/80 dark:border-gray-800 space-y-2">
          <div class="flex items-center gap-2 text-primary-700 dark:text-primary-400 font-bold text-sm">
            <UIcon name="i-heroicons-face-smile" class="size-4 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.csat.title') }}</span>
          </div>
          <p class="text-gray-600 dark:text-gray-300 leading-relaxed">
            <strong>{{ t('satisfactionSurvey.guide.csat.definition').split(':')[0] }}</strong>
          </p>
          <div class="space-y-1 text-gray-500 dark:text-gray-400">
            <p><strong>{{ t('satisfactionSurvey.guide.csat.dimensionsTitle') }}</strong></p>
            <ul class="list-disc pl-4 space-y-0.5">
              <li><strong>Clarity:</strong> {{ t('satisfactionSurvey.guide.csat.clarity') }}</li>
              <li><strong>Professionalism:</strong> {{ t('satisfactionSurvey.guide.csat.professionalism') }}</li>
              <li><strong>Timeliness:</strong> {{ t('satisfactionSurvey.guide.csat.timeliness') }}</li>
            </ul>
          </div>
          <div class="pt-2 text-[11px] text-primary-600 dark:text-primary-400 font-semibold border-t border-gray-100 dark:border-gray-800 flex items-center gap-1.5">
            <UIcon name="i-lucide-target" class="size-3.5 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.csat.target') }}</span>
          </div>
        </div>

        <!-- Metric 2 Explanation (Secondary) -->
        <div class="p-4 bg-white dark:bg-gray-900 rounded-xl border border-gray-200/80 dark:border-gray-800 space-y-2">
          <div class="flex items-center gap-2 text-secondary-700 dark:text-secondary-400 font-bold text-sm">
            <UIcon name="i-heroicons-document-check" class="size-4 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.surveys.title') }}</span>
          </div>
          <p class="text-gray-600 dark:text-gray-300 leading-relaxed">
            <strong>{{ t('satisfactionSurvey.guide.surveys.definition').split(':')[0] }}</strong>
          </p>
          <div class="space-y-1 text-gray-500 dark:text-gray-400">
            <p><strong>{{ t('satisfactionSurvey.guide.surveys.flowTitle') }}</strong></p>
            <ul class="list-disc pl-4 space-y-0.5">
              <li>{{ t('satisfactionSurvey.guide.surveys.flow1') }}</li>
              <li>{{ t('satisfactionSurvey.guide.surveys.flow2') }}</li>
              <li>{{ t('satisfactionSurvey.guide.surveys.flow3') }}</li>
            </ul>
          </div>
          <div class="pt-2 text-[11px] text-secondary-600 dark:text-secondary-400 font-semibold border-t border-gray-100 dark:border-gray-800 flex items-center gap-1.5">
            <UIcon name="i-lucide-bar-chart-2" class="size-3.5 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.surveys.note') }}</span>
          </div>
        </div>

        <!-- Metric 3 Explanation (Success) -->
        <div class="p-4 bg-white dark:bg-gray-900 rounded-xl border border-gray-200/80 dark:border-gray-800 space-y-2">
          <div class="flex items-center gap-2 text-success-700 dark:text-success-400 font-bold text-sm">
            <UIcon name="i-heroicons-chart-bar-solid" class="size-4 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.rate.title') }}</span>
          </div>
          <p class="text-gray-600 dark:text-gray-300 leading-relaxed">
            <strong>{{ t('satisfactionSurvey.guide.rate.definition').split(':')[0] }}</strong>
          </p>
          <div class="space-y-1 text-gray-500 dark:text-gray-400">
            <p><strong>{{ t('satisfactionSurvey.guide.rate.formulaTitle') }}</strong></p>
            <div class="p-2 bg-gray-50 dark:bg-gray-800/60 rounded font-mono text-[11px] text-gray-700 dark:text-gray-300">
              {{ t('satisfactionSurvey.guide.rate.formula') }}
            </div>
            <p class="mt-1">{{ t('satisfactionSurvey.guide.rate.desc') }}</p>
          </div>
          <div class="pt-2 text-[11px] text-success-600 dark:text-success-400 font-semibold border-t border-gray-100 dark:border-gray-800 flex items-center gap-1.5">
            <UIcon name="i-lucide-target" class="size-3.5 shrink-0" />
            <span>{{ t('satisfactionSurvey.guide.rate.target') }}</span>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Main Content Tabs / Tables -->
    <UCard class="shadow-sm border border-[var(--border-main)] overflow-hidden">
      <template #header>
        <div class="flex justify-between items-center py-1">
          <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon name="i-heroicons-clipboard-document-list" class="text-primary-600" />
            {{ t('satisfactionSurvey.table.title') }}
          </h2>
          <UButton
            icon="i-heroicons-arrow-path"
            color="neutral"
            variant="ghost"
            size="sm"
            @click="refreshAllData"
            :loading="loadingData"
          >
            {{ t('satisfactionSurvey.table.refresh') }}
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
                {{ t('satisfactionSurvey.table.surveyStatus.submitted', { score: getReportSurvey(row.original)!.overall_score?.toFixed(1) }) }}
              </span>
              <span v-else class="inline-flex items-center gap-1 text-sm font-semibold text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950/30 px-2 py-1 rounded-md border border-amber-100 dark:border-amber-900/30">
                <UIcon name="i-heroicons-clock" class="size-4" />
                {{ t('satisfactionSurvey.table.surveyStatus.pending') }}
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
                :label="t('satisfactionSurvey.table.actions.fill')"
                @click="openSurveyForm(row.original)"
              />
              <UButton
                v-else
                color="neutral"
                variant="outline"
                size="sm"
                icon="i-heroicons-eye"
                :label="t('satisfactionSurvey.table.actions.view')"
                @click="viewSurveyFeedback(getReportSurvey(row.original)!, row.original)"
              />
            </div>
          </template>
        </UTable>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-16 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-dashed border-gray-200 dark:border-gray-800">
        <UIcon name="i-heroicons-document-text" class="size-16 text-gray-300 mx-auto mb-4" />
        <h3 class="text-lg font-semibold text-gray-700 dark:text-gray-350">{{ t('satisfactionSurvey.table.empty.title') }}</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-2 max-w-md mx-auto">
          {{ t('satisfactionSurvey.table.empty.desc') }}
        </p>
      </div>
    </UCard>

    <!-- Survey Form Modal -->
    <UModal 
      v-model:open="showFormModal" 
      dismissible 
      :ui="{
        content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <div class="flex flex-col h-full max-h-[95vh]">
          <!-- Header -->
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-gray-50 dark:bg-gray-900/50">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-primary-100 dark:bg-primary-950/50 rounded-lg">
                <UIcon name="i-heroicons-chat-bubble-bottom-center-text" class="text-primary-600 size-6" />
              </div>
              <div>
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('satisfactionSurvey.formModal.title') }}</h3>
                <p class="text-md text-gray-500 dark:text-gray-400 mt-0.5">{{ t('satisfactionSurvey.formModal.subtitle') }}</p>
              </div>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="() => { showFormModal = false }"
            />
          </div>

          <!-- Body -->
          <div class="p-6 overflow-y-auto flex-1 space-y-6">
            <!-- Report Meta Info Banner -->
            <div class="p-4 bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800 rounded-xl space-y-2">
              <div class="grid grid-cols-2 gap-4 text-md">
                <div>
                  <span class="text-gray-500 block">{{ t('satisfactionSurvey.formModal.reportNumber') }}</span>
                  <span class="font-mono font-semibold text-gray-800 dark:text-gray-200">
                    {{ selectedReport?.reportNumber || (selectedReport as any)?.report_number }}
                  </span>
                </div>
                <div>
                  <span class="text-gray-500 block">{{ t('satisfactionSurvey.formModal.department') }}</span>
                  <span class="font-semibold text-gray-800 dark:text-gray-200">
                    {{ selectedReport?.department || 'General' }}
                  </span>
                </div>
                <div class="col-span-2">
                  <span class="text-gray-500 block">{{ t('satisfactionSurvey.formModal.reportTitle') }}</span>
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
                <UFormField :label="t('satisfactionSurvey.formModal.auditeeName')" name="auditeeName" required>
                  <UInput
                    v-model="formState.auditeeName"
                    :placeholder="t('satisfactionSurvey.formModal.auditeeNamePlaceholder')"
                    class="w-full"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('satisfactionSurvey.formModal.auditeeNameValidation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ formState.auditeeName ? formState.auditeeName.length : 0 }}/100
                  </div>
                </UFormField>

                <UFormField :label="t('satisfactionSurvey.formModal.departmentLabel')" name="department" required>
                  <UInput
                    v-model="formState.department"
                    :placeholder="t('satisfactionSurvey.formModal.departmentLabel')"
                    disabled
                    class="w-full bg-gray-50 dark:bg-gray-900 cursor-not-allowed"
                  />
                </UFormField>
              </div>

              <!-- Rating 1: Clarity -->
              <div class="p-4 border border-gray-150 dark:border-gray-800 rounded-xl space-y-3">
                <div>
                  <h4 class="text-sm font-bold text-gray-800 dark:text-gray-250 flex justify-between">
                    <span>{{ t('satisfactionSurvey.formModal.q1Title') }}</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingClarity }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">{{ t('satisfactionSurvey.formModal.q1Desc') }}</p>
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
                    <span>{{ t('satisfactionSurvey.formModal.q2Title') }}</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingProfessionalism }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">{{ t('satisfactionSurvey.formModal.q2Desc') }}</p>
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
                    <span>{{ t('satisfactionSurvey.formModal.q3Title') }}</span>
                    <span class="text-primary-600 font-extrabold text-md">{{ formState.ratingTimeliness }} / 5</span>
                  </h4>
                  <p class="text-md text-gray-500 mt-1">{{ t('satisfactionSurvey.formModal.q3Desc') }}</p>
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
              <UCard
                :ui="{ body: 'p-4 text-white dark:text-white flex items-center justify-between', background: 'bg-transparent' }"
                class="bg-primary-600 text-white dark:text-white shadow-md rounded-xl border-0"
              >
                <div>
                  <span class="text-sm font-bold uppercase tracking-wider block text-white">{{ t('satisfactionSurvey.formModal.estimatedScore') }}</span>
                  <span class="text-xs text-white/90 mt-0.5 block">{{ t('satisfactionSurvey.formModal.estimatedScoreFormula') }}</span>
                </div>
                <div class="text-right">
                  <span class="text-2xl font-extrabold text-white">
                    {{ computedOverallScore.toFixed(2) }}
                  </span>
                  <span class="text-sm text-white/90 font-medium"> / 5.0</span>
                </div>
              </UCard>

              <!-- Comments -->
              <UFormField :label="t('satisfactionSurvey.formModal.comments')" name="comments">
                <UTextarea
                  v-model="formState.comments"
                  :placeholder="t('satisfactionSurvey.formModal.commentsPlaceholder')"
                  :rows="4"
                  class="w-full animate-fade-in"
                />
              </UFormField>

              <!-- Footer Actions -->
              <div class="flex justify-end gap-3 pt-6 border-t border-gray-100 dark:border-gray-800">
                <UButton
                  :label="t('satisfactionSurvey.formModal.cancel')"
                  color="neutral"
                  variant="ghost"
                  @click="() => { showFormModal = false }"
                />
                <UButton
                  type="submit"
                  :label="t('satisfactionSurvey.formModal.submit')"
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
    <UModal 
      v-model:open="showViewModal" 
      dismissible 
      :ui="{
        content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <div class="flex flex-col h-full max-h-[90vh]">
          <!-- Header -->
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center bg-gray-50 dark:bg-gray-900/50">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-emerald-100 dark:bg-emerald-950/50 rounded-lg">
                <UIcon name="i-heroicons-clipboard-document-check" class="text-emerald-600 size-6" />
              </div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('satisfactionSurvey.viewModal.title') }}</h3>
            </div>
            <UButton
              color="neutral"
              variant="ghost"
              icon="i-heroicons-x-mark"
              @click="() => { showViewModal = false }"
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
                    <span class="text-md text-gray-400 block">{{ t('satisfactionSurvey.viewModal.submittedBy') }}</span>
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ viewSurveyDetail?.auditee_name }}</span>
                  </div>
                  <div>
                    <span class="text-md text-gray-400 block">{{ t('satisfactionSurvey.viewModal.department') }}</span>
                    <span class="font-medium text-gray-700 dark:text-gray-300">{{ viewSurveyDetail?.department }}</span>
                  </div>
                </div>
              </div>

              <!-- Star ratings display -->
              <div class="space-y-3">
                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">{{ t('satisfactionSurvey.viewModal.clarity') }}</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_clarity || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_clarity }}/5)</span>
                  </div>
                </div>

                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">{{ t('satisfactionSurvey.viewModal.professionalism') }}</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_professionalism || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_professionalism }}/5)</span>
                  </div>
                </div>

                <div class="flex justify-between items-center p-3 bg-gray-50 dark:bg-slate-900/50 rounded-lg">
                  <span class="text-md text-gray-600 dark:text-gray-400 font-medium">{{ t('satisfactionSurvey.viewModal.timeliness') }}</span>
                  <div class="flex items-center gap-1">
                    <UIcon v-for="star in 5" :key="star" name="i-heroicons-star-solid" :class="star <= (viewSurveyDetail?.rating_timeliness || 0) ? 'text-amber-500 w-4 h-4' : 'text-gray-250 dark:text-gray-700 w-4 h-4'" />
                    <span class="ml-2 font-mono font-bold text-gray-800 dark:text-gray-200">({{ viewSurveyDetail?.rating_timeliness }}/5)</span>
                  </div>
                </div>
              </div>

              <!-- Overall Score -->
              <div class="p-4 bg-gradient-to-r from-emerald-500/10 to-emerald-500/5 border border-emerald-500/20 rounded-xl flex items-center justify-between">
                <div>
                  <span class="text-sm font-semibold text-emerald-800 dark:text-emerald-400 uppercase tracking-wider block">{{ t('satisfactionSurvey.viewModal.overallScore') }}</span>
                  <span class="text-md text-gray-500 mt-0.5">{{ t('satisfactionSurvey.viewModal.overallScoreSubtitle') }}</span>
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
                <span class="text-md text-gray-400 uppercase font-semibold">{{ t('satisfactionSurvey.viewModal.comments') }}</span>
                <div class="p-4 bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800 rounded-xl text-md text-gray-700 dark:text-gray-300 whitespace-pre-line leading-relaxed italic">
                  "{{ viewSurveyDetail?.comments || t('satisfactionSurvey.viewModal.noComments') }}"
                </div>
              </div>
            </div>

            <!-- Footer Action -->
            <div class="flex justify-end pt-4 border-t border-gray-100 dark:border-gray-800">
              <UButton
                :label="t('satisfactionSurvey.viewModal.close')"
                color="neutral"
                variant="soft"
                @click="() => { showViewModal = false }"
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
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useI18n } from '~/composables/useI18n'

const { t } = useI18n()
const authStore = useAuthStore()
const reportStore = useAuditResultReportStore()
const executionStore = useAuditExecutionStore()
const surveysStore = useAuditeeSurveyStore()

const loadingData = ref(false)
const showFormModal = ref(false)
const showViewModal = ref(false)
const submitting = ref(false)
const showExplanationGuide = ref(true)

const selectedReport = ref<any>(null)
const viewSurveyDetail = ref<AuditeeSurvey | null>(null)
const viewReportDetail = ref<any>(null)
const toast = useToastNotification()

// Initial form state
const formState = ref({
  auditeeName: '',
  department: '',
  ratingClarity: 5,
  ratingProfessionalism: 5,
  ratingTimeliness: 5,
  comments: ''
})

const columns = computed(() => [
  { accessorKey: 'reportNumber', header: t('satisfactionSurvey.table.columns.reportNumber') },
  { accessorKey: 'reportTitle', header: t('satisfactionSurvey.table.columns.reportTitle') },
  { accessorKey: 'department', header: t('satisfactionSurvey.table.columns.department') },
  { accessorKey: 'reportDate', header: t('satisfactionSurvey.table.columns.publishDate') },
  { accessorKey: 'status', header: t('satisfactionSurvey.table.columns.status') },
  { accessorKey: 'surveyStatus', header: t('satisfactionSurvey.table.columns.surveyStatus') },
  { accessorKey: 'actions', header: t('satisfactionSurvey.table.columns.action') }
])

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
    case 1: return t('satisfactionSurvey.formModal.ratingScale.1')
    case 2: return t('satisfactionSurvey.formModal.ratingScale.2')
    case 3: return t('satisfactionSurvey.formModal.ratingScale.3')
    case 4: return t('satisfactionSurvey.formModal.ratingScale.4')
    case 5: return t('satisfactionSurvey.formModal.ratingScale.5')
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
        alert(t('satisfactionSurvey.formModal.alreadySubmitted'))
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
    toast.showSuccess(t('satisfactionSurvey.formModal.success'))
  } catch (error: any) {
    console.error('Failed to submit survey:', error)
    toast.showError(error.message || t('satisfactionSurvey.formModal.error'))
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
