<template>
  <div class="space-y-4 animate-fade-in">
    <!-- Header: Auditor Profile & Context Bar -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 px-4 py-3 bg-gradient-to-r from-emerald-900/10 via-primary-500/5 to-transparent border border-emerald-500/20 rounded-xl shadow-xs">
      <div class="flex items-start sm:items-center gap-3">
        <div class="size-10 rounded-xl bg-emerald-600/10 dark:bg-emerald-500/20 border border-emerald-500/30 flex items-center justify-center text-emerald-600 dark:text-emerald-400 shrink-0 shadow-xs">
          <UIcon
            name="i-lucide-clipboard-check"
            class="size-5"
          />
        </div>
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h1 class="text-lg sm:text-xl font-bold text-gray-900 dark:text-white tracking-tight">
              {{ t('dashboard.auditor.title') }}
            </h1>
            <!-- <UBadge
              color="success"
              variant="subtle"
              size="sm"
              class="font-bold uppercase tracking-wider"
            >
              {{ t('dashboard.auditor.roleBadge') }}
            </UBadge> -->
          </div>
          <p class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 mt-0.5 max-w-3xl">
            {{ t('dashboard.auditor.welcome', { name: currentUserFullName }) }}
          </p>
        </div>
      </div>

      <!-- Quick Action Controls & Role Switcher -->
      <div class="flex flex-wrap items-center gap-2 self-start lg:self-center shrink-0">
        <!-- Role Preview Pill (Visible for admins / reviewers to test/toggle) -->
        <div
          v-if="canToggleRoleView"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-xs font-medium"
        >
          <UIcon
            name="i-lucide-eye"
            class="size-4 text-emerald-500"
          />
          <span class="text-gray-600 dark:text-gray-300">{{ t('dashboard.auditor.previewMode') }}</span>
          <span class="font-bold text-emerald-600 dark:text-emerald-400">{{ t('dashboard.auditor.previewAuditor') }}</span>
          <button
            type="button"
            class="ml-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 underline text-[11px]"
            :title="t('dashboard.auditor.previewResetTitle')"
            @click="switchDashboardView(null)"
          >
            {{ t('dashboard.auditor.previewReset') }}
          </button>
        </div>

        <UButton
          variant="outline"
          color="neutral"
          size="sm"
          class="font-medium cursor-pointer"
          :loading="isSyncing"
          @click="manualSyncAuditorData"
        >
          <template #leading>
            <UIcon
              name="i-lucide-rotate-cw"
              :class="{ 'animate-spin': isSyncing }"
              class="size-4"
            />
          </template>
          {{ t('dashboard.auditor.sync') }}
        </UButton>
      </div>
    </div>

    <!-- Notice -->
    <div class="flex items-start gap-2 px-3 py-2 rounded-lg bg-secondary-500/10 border border-secondary-500/20 text-secondary-900 dark:text-secondary-200 text-xs sm:text-sm">
      <UIcon
        name="i-lucide-info"
        class="size-4 mt-0.5 text-secondary-600 dark:text-secondary-400 shrink-0"
      />
      <div>
        <span class="font-bold">{{ t('dashboard.auditor.notice.label') }}</span>
        <span class="ml-1 text-secondary-800 dark:text-secondary-300">{{ t('dashboard.auditor.notice.text') }}</span>
      </div>
    </div>

    <!-- 1. Audit statistics -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-stats-title"
    >
      <div>
        <h2
          id="auditor-stats-title"
          class="text-base font-bold text-gray-900 dark:text-white"
        >
          {{ t('dashboard.auditor.stats.title') }}
        </h2>
        <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
          {{ t('dashboard.auditor.stats.scope') }}
        </p>
      </div>
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-3">
        <div
          v-for="card in statCards"
          :key="card.key"
          class="rounded-lg p-3 flex flex-col justify-between border border-gray-200/80 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30"
        >
          <div class="flex items-start justify-between gap-2">
            <span class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ card.label }}</span>
            <div
              class="p-1.5 rounded-md shrink-0"
              :class="card.iconBg"
            >
              <UIcon
                :name="card.icon"
                class="size-4"
                :class="card.iconClass"
              />
            </div>
          </div>
          <AuditorWidgetState
            :state="card.state === 'empty' ? 'ready' : card.state"
            :error-text="t('dashboard.auditor.states.loadError')"
            :retry-label="t('dashboard.auditor.states.retry')"
            @retry="card.retry"
          >
            <div class="mt-1">
              <h4 class="text-xl font-extrabold leading-tight text-gray-900 dark:text-white">
                {{ card.value }}
              </h4>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5">
                {{ card.hint }}
              </p>
              <p
                v-if="card.state === 'empty'"
                class="text-[11px] text-gray-400 mt-1"
              >
                {{ t('dashboard.auditor.states.noData') }}
              </p>
            </div>
          </AuditorWidgetState>
        </div>
      </div>
    </section>

    <!-- 2. Audit planning & coverage -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-coverage-title"
    >
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2
            id="auditor-coverage-title"
            class="text-base font-bold text-gray-900 dark:text-white"
          >
            {{ t('dashboard.auditor.coverage.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditor.coverage.description') }}
          </p>
          <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
            {{ t('dashboard.auditor.coverage.scope') }}
          </p>
        </div>
        <UButton
          to="/annual-audit"
          variant="ghost"
          color="neutral"
          size="sm"
          class="self-start sm:self-auto"
        >
          {{ t('dashboard.auditor.coverage.openPlan') }}
        </UButton>
      </div>

      <AuditorWidgetState
        :state="planState === 'empty' ? 'ready' : planState"
        :error-text="t('dashboard.auditor.states.loadError')"
        :retry-label="t('dashboard.auditor.states.retry')"
        @retry="reloadPlans"
      >
        <p
          v-if="planState === 'empty'"
          class="text-xs text-gray-500 dark:text-gray-400"
        >
          {{ t('dashboard.auditor.coverage.empty') }}
        </p>
        <div class="space-y-2">
          <div class="flex flex-wrap justify-between items-center gap-x-3 gap-y-1">
            <span class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t('dashboard.auditor.coverage.overallProgress') }}</span>
            <h5 class="text-base font-bold text-indigo-600 dark:text-indigo-400 whitespace-nowrap">
              {{ coveragePercent }}%
            </h5>
          </div>
          <UProgress
            :model-value="coveragePercent"
            color="secondary"
            class="h-2 rounded"
          />
        </div>
        <div class="grid grid-cols-3 gap-2 border-t border-gray-100 dark:border-gray-800 pt-3 mt-3 text-center">
          <div>
            <p class="text-[11px] text-gray-400 font-semibold uppercase">
              {{ t('dashboard.auditor.coverage.planned') }}
            </p>
            <h4 class="text-base font-bold text-gray-800 dark:text-white">
              {{ auditCoverage.plannedAudits }}
            </h4>
          </div>
          <div>
            <p class="text-[11px] text-gray-400 font-semibold uppercase">
              {{ t('dashboard.auditor.coverage.approved') }}
            </p>
            <h4 class="text-base font-bold text-gray-800 dark:text-white">
              {{ auditCoverage.completedAudits }}
            </h4>
          </div>
          <div>
            <p class="text-[11px] text-gray-400 font-semibold uppercase">
              {{ t('dashboard.auditor.coverage.remaining') }}
            </p>
            <h4 class="text-base font-bold text-gray-800 dark:text-white">
              {{ auditCoverage.remainingAudits }}
            </h4>
          </div>
        </div>
      </AuditorWidgetState>
    </section>

    <!-- 3. Audit status and progress (active fieldwork) -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-status-title"
    >
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2
            id="auditor-status-title"
            class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <UIcon
              name="i-lucide-briefcase"
              class="size-5 text-emerald-600 dark:text-emerald-400"
            />
            {{ t('dashboard.auditor.status.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditor.status.description') }}
          </p>
          <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
            {{ t('dashboard.auditor.status.scope') }}
          </p>
        </div>

        <div class="flex items-center gap-2.5 w-full sm:w-auto">
          <UInput
            v-model="fieldworkSearch"
            icon="i-lucide-search"
            :placeholder="t('dashboard.auditor.status.searchPlaceholder')"
            size="sm"
            class="w-full sm:w-64"
          />
          <UButton
            to="/assignment-letter"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-external-link"
            class="shrink-0 font-medium"
          >
            {{ t('dashboard.auditor.status.allAssignments') }}
          </UButton>
        </div>
      </div>

      <AuditorWidgetState
        :state="statusState === 'empty' ? 'ready' : statusState"
        :error-text="t('dashboard.auditor.states.loadError')"
        :retry-label="t('dashboard.auditor.states.retry')"
        @retry="reloadLetters"
      >
        <div
          v-if="statusState === 'empty'"
          class="py-5 text-center text-gray-500 dark:text-gray-400"
        >
          <UIcon
            name="i-lucide-folder-open"
            class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2"
          />
          <p class="font-medium text-sm">
            {{ t('dashboard.auditor.status.empty') }}
          </p>
          <p class="text-xs text-gray-400 mt-1">
            {{ t('dashboard.auditor.status.emptyHint') }}
          </p>
        </div>

        <div
          v-else
          class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800"
        >
          <table class="w-full text-left text-sm">
            <thead class="bg-gray-50 dark:bg-gray-800/80 text-gray-700 dark:text-gray-200 border-b border-gray-200 dark:border-gray-700 text-xs uppercase font-bold tracking-wider">
              <tr>
                <th class="py-2 px-3 min-w-[200px]">
                  {{ t('dashboard.auditor.status.colLetter') }}
                </th>
                <th class="py-2 px-3 min-w-[130px]">
                  {{ t('dashboard.auditor.status.colUnit') }}
                </th>
                <th class="py-2 px-3 min-w-[170px]">
                  {{ t('dashboard.auditor.status.colPeriod') }}
                </th>
                <th class="py-2 px-3 min-w-[240px]">
                  {{ t('dashboard.auditor.status.colStages') }}
                </th>
                <th class="py-2 px-3 min-w-[120px] text-center">
                  {{ t('dashboard.auditor.status.colWorkingPapers') }}
                </th>
                <th class="py-2 px-3 min-w-[120px] text-center">
                  {{ t('dashboard.auditor.status.colStatus') }}
                </th>
                <th class="py-2 px-3 min-w-[160px] text-center">
                  {{ t('dashboard.auditor.status.colActions') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
              <tr
                v-for="item in filteredFieldworkList"
                :key="item.letterNumber"
                class="hover:bg-emerald-50/30 dark:hover:bg-emerald-950/10 transition-colors"
                :class="{ 'bg-emerald-50/20 dark:bg-emerald-950/20': selectedLetterNumber === item.letterNumber }"
              >
                <!-- Letter Number & Title -->
                <td class="py-2.5 px-3">
                  <div class="flex items-start gap-2.5">
                    <div class="size-7 rounded-md bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0 mt-0.5">
                      <UIcon
                        name="i-lucide-file-text"
                        class="size-4"
                      />
                    </div>
                    <div>
                      <span class="font-mono font-bold text-xs text-primary-600 dark:text-primary-400 block">
                        {{ item.letterNumber }}
                      </span>
                      <span class="font-semibold text-gray-900 dark:text-gray-100 text-xs sm:text-sm line-clamp-1">
                        {{ item.auditTitle }}
                      </span>
                      <span class="text-[11px] text-gray-500 dark:text-gray-400">
                        {{ t('dashboard.auditor.status.team') }}: {{ item.auditTeam || 'SKAI' }} &bull; {{ item.category || 'Assurance' }}
                      </span>
                    </div>
                  </div>
                </td>

                <!-- Auditee Unit -->
                <td class="py-2.5 px-3">
                  <div class="flex items-center gap-1.5 font-medium text-xs text-gray-800 dark:text-gray-200">
                    <UIcon
                      name="i-lucide-building"
                      class="size-3.5 text-gray-400 shrink-0"
                    />
                    <span>{{ item.workingUnit }}</span>
                  </div>
                </td>

                <!-- Execution Period -->
                <td class="py-2.5 px-3 text-xs">
                  <div class="space-y-1">
                    <div class="flex items-center gap-1 text-gray-700 dark:text-gray-300">
                      <UIcon
                        name="i-lucide-calendar"
                        class="size-3.5 text-gray-400"
                      />
                      <span>{{ formatPeriod(item.startPeriod, item.finishPeriod) }}</span>
                    </div>
                    <UBadge
                      :color="getDeadlineBadgeColor(item.finishPeriod)"
                      variant="soft"
                      size="sm"
                      class="text-[10px] font-semibold"
                    >
                      {{ getDeadlineLabel(item.finishPeriod) }}
                    </UBadge>
                  </div>
                </td>

                <!-- Fieldwork Milestones Stepper -->
                <td class="py-2.5 px-3">
                  <div class="space-y-1.5">
                    <div class="flex items-center justify-between text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                      <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 1 }">{{ t('dashboard.auditor.status.stepper.entry') }}</span>
                      <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 2 }">{{ t('dashboard.auditor.status.stepper.testing') }}</span>
                      <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 3 }">{{ t('dashboard.auditor.status.stepper.workingPapers') }}</span>
                      <span :class="{ 'text-emerald-600 dark:text-emerald-400 font-bold': item.stageStep >= 4 }">{{ t('dashboard.auditor.status.stepper.exit') }}</span>
                    </div>
                    <!-- Progress Tracker Bar -->
                    <div class="w-full h-1.5 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden flex">
                      <div
                        class="bg-emerald-500 h-full transition-all duration-500"
                        :style="{ width: `${(item.stageStep / 4) * 100}%` }"
                      />
                    </div>
                    <div class="text-[11px] text-gray-500 dark:text-gray-400">
                      {{ t('dashboard.auditor.status.currentStage') }}
                      <strong class="text-gray-800 dark:text-gray-200">{{ t(`dashboard.auditor.status.stage.${item.stageKey}`) }}</strong>
                    </div>
                  </div>
                </td>

                <!-- KKA Progress Bar -->
                <td class="py-2.5 px-3 text-center">
                  <div class="inline-flex flex-col items-center gap-1">
                    <span class="font-bold text-xs text-gray-900 dark:text-gray-100">
                      {{ t('dashboard.auditor.status.formsCount', { count: item.kkaCompleted }) }}
                    </span>
                    <div class="w-16 h-1.5 bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all"
                        :class="item.kkaCompleted === 5 ? 'bg-emerald-500' : 'bg-primary-500'"
                        :style="{ width: `${(item.kkaCompleted / 5) * 100}%` }"
                      />
                    </div>
                  </div>
                </td>

                <!-- Status Badge -->
                <td class="py-2.5 px-3 text-center">
                  <UBadge
                    :color="getFieldworkStatusColor(item.statusKey)"
                    variant="subtle"
                    size="sm"
                    class="font-semibold"
                  >
                    {{ t(`dashboard.auditor.status.statusLabel.${item.statusKey}`) }}
                  </UBadge>
                </td>

                <!-- Operational Actions -->
                <td class="py-2.5 px-3 text-center">
                  <div class="flex items-center justify-center gap-1.5">
                    <UButton
                      size="xs"
                      color="primary"
                      variant="solid"
                      icon="i-lucide-file-signature"
                      class="font-bold cursor-pointer"
                      @click="goToWorkingPaper(item.letterNumber)"
                    >
                      {{ t('dashboard.auditor.status.openWorkingPaper') }}
                    </UButton>
                    <UButton
                      size="xs"
                      color="neutral"
                      variant="outline"
                      icon="i-lucide-briefcase"
                      class="font-medium cursor-pointer"
                      @click="goToFieldwork(item.letterNumber)"
                    >
                      {{ t('dashboard.auditor.status.openFieldwork') }}
                    </UButton>
                    <UButton
                      size="xs"
                      color="neutral"
                      variant="ghost"
                      icon="i-lucide-file-text"
                      class="cursor-pointer"
                      :title="t('dashboard.auditor.status.openLha')"
                      @click="goToResultReport(item.letterNumber)"
                    />
                  </div>
                </td>
              </tr>

              <tr v-if="filteredFieldworkList.length === 0">
                <td
                  colspan="7"
                  class="py-5 text-center text-gray-500 dark:text-gray-400"
                >
                  <p class="font-medium">
                    {{ t('dashboard.auditor.status.emptyFiltered') }}
                  </p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </AuditorWidgetState>
    </section>

    <!-- 4. Task list with deadlines -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-tasks-title"
    >
      <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-2">
        <div>
          <h2
            id="auditor-tasks-title"
            class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <UIcon
              name="i-lucide-list-checks"
              class="size-5 text-primary-600 dark:text-primary-400"
            />
            {{ t('dashboard.auditor.tasks.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditor.tasks.description') }}
          </p>
          <p class="text-[11px] text-gray-400 dark:text-gray-500 mt-0.5">
            {{ t('dashboard.auditor.tasks.scope') }}
          </p>
        </div>
      </div>

      <AuditorWidgetState
        :state="tasksState"
        :error-text="t('dashboard.auditor.states.loadError')"
        :retry-label="t('dashboard.auditor.states.retry')"
        :empty-text="t('dashboard.auditor.tasks.empty')"
        @retry="reloadAtr"
      >
        <ul class="divide-y divide-gray-200 dark:divide-gray-800 rounded-xl border border-gray-200 dark:border-gray-800">
          <li
            v-for="item in taskRows"
            :key="item.id"
            class="flex flex-wrap items-center justify-between gap-3 px-3 py-2 text-xs"
          >
            <div class="min-w-0">
              <p class="font-semibold text-gray-900 dark:text-white line-clamp-1">
                {{ item.finding_title || item.report_title || item.report_number || '-' }}
              </p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5">
                {{ item.status === 'PENDING_REVIEW' ? t('dashboard.auditor.tasks.kindReview') : t('dashboard.auditor.tasks.kindFollowUp') }}
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-2 shrink-0">
              <UBadge
                :color="atrStatusBadgeColor(item.status)"
                variant="subtle"
                size="xs"
              >
                {{ atrStatusLabel(item.status) }}
              </UBadge>
              <UBadge
                v-if="item.is_overdue"
                color="error"
                variant="soft"
                size="xs"
              >
                {{ t('dashboard.auditor.tasks.overdueDays', { days: item.overdue_days }) }}
              </UBadge>
              <span
                v-else
                class="text-[11px] text-gray-500 dark:text-gray-400"
              >
                {{ item.due_date ? t('dashboard.auditor.tasks.dueOn', { date: formatAtrDate(item.due_date, locale) }) : t('dashboard.auditor.tasks.noDueDate') }}
              </span>
              <UButton
                to="/action-taken-report"
                size="xs"
                variant="ghost"
                color="neutral"
                icon="i-lucide-arrow-right"
              />
            </div>
          </li>
        </ul>
        <p class="mt-2 text-[11px] text-gray-400 dark:text-gray-500">
          {{ t('dashboard.auditor.tasks.placeholder') }}
        </p>
      </AuditorWidgetState>
    </section>

    <!-- 5. Calendar of important dates -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-calendar-title"
    >
      <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-2">
        <div>
          <h2
            id="auditor-calendar-title"
            class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <UIcon
              name="i-lucide-calendar-days"
              class="size-5 text-sky-500"
            />
            {{ t('dashboard.auditor.calendar.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditor.calendar.description') }}
          </p>
        </div>
      </div>

      <!-- TODO(B4): replace calendarEvents with GET /api/v1/me/calendar?from&to (audit-service), scoped to the user's team. -->
      <CalendarWidget
        :events="calendarEvents"
        :loading="!initialLoadDone"
        :error-text="calendarErrorText"
        :retry-label="t('dashboard.auditor.states.retry')"
        :note="employeeLinked ? t('dashboard.auditor.calendar.scope') : t('dashboard.auditor.calendar.notLinked')"
        @retry="reloadCalendarSources"
      />
    </section>

    <!-- 6. Notifications and reminders -->
    <section
      class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3"
      aria-labelledby="auditor-notifications-title"
    >
      <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-2">
        <div>
          <h2
            id="auditor-notifications-title"
            class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2"
          >
            <UIcon
              name="i-lucide-bell"
              class="size-5 text-amber-500"
            />
            {{ t('dashboard.auditor.notifications.title') }}
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {{ t('dashboard.auditor.notifications.description') }}
          </p>
        </div>
      </div>

      <AuditorWidgetState
        :state="notificationsState"
        :error-text="t('dashboard.auditor.states.loadError')"
        :retry-label="t('dashboard.auditor.states.retry')"
        :empty-text="t('dashboard.auditor.notifications.empty')"
        @retry="reloadAtr"
      >
        <ul class="divide-y divide-gray-200 dark:divide-gray-800 rounded-xl border border-gray-200 dark:border-gray-800">
          <li
            v-for="n in derivedNotifications"
            :key="`${n.kind}-${n.id}`"
            class="flex items-center justify-between gap-3 px-3 py-2 text-xs"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <UIcon
                :name="n.kind === 'overdue' ? 'i-lucide-alert-triangle' : (n.kind === 'pendingReview' ? 'i-lucide-eye' : 'i-lucide-clock')"
                class="size-4 shrink-0"
                :class="n.kind === 'overdue' ? 'text-red-500' : (n.kind === 'pendingReview' ? 'text-sky-500' : 'text-amber-500')"
              />
              <div class="min-w-0">
                <p class="font-semibold text-gray-900 dark:text-white">
                  {{ notificationText(n) }}
                </p>
                <p class="text-[11px] text-gray-500 dark:text-gray-400 line-clamp-1">
                  {{ n.item.finding_title || n.item.report_title || n.item.report_number || '-' }}
                </p>
              </div>
            </div>
            <UButton
              to="/action-taken-report"
              size="xs"
              variant="ghost"
              color="neutral"
              icon="i-lucide-arrow-right"
            />
          </li>
        </ul>
        <p class="mt-2 text-[11px] text-gray-400 dark:text-gray-500">
          {{ t('dashboard.auditor.notifications.placeholder') }}
        </p>
      </AuditorWidgetState>
    </section>

    <!-- KKA FORMS (F-01 to F-05) -->
    <div class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-file-signature"
              class="size-5 text-primary-600 dark:text-primary-400"
            />
            Kertas Kerja & Template
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            Akses cepat pembuatan atau kelanjutan formulir kertas kerja pengujian audit dan upload kertas kerja pendukung.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2 w-full sm:w-auto">
          <!-- Assignment Letter Selector for KKA -->
          <div class="w-full sm:w-64">
            <USelectMenu
              v-model="selectedLetterNumber"
              :items="assignmentLetterOptions"
              value-key="value"
              placeholder="Pilih Surat Tugas..."
              class="w-full"
            >
              <template #leading>
                <UIcon
                  name="i-lucide-file-text"
                  class="size-4 text-primary-500"
                />
              </template>
            </USelectMenu>
          </div>

          <!-- Button Upload KKA (Sesuai Matriks RBAC Row 8-3: UPLOAD) -->
          <UButton
            to="/working-paper/upload"
            variant="outline"
            color="neutral"
            size="sm"
            icon="i-lucide-upload"
            class="font-semibold cursor-pointer shrink-0"
          >
            Upload Kertas Kerja Pendukung
          </UButton>
        </div>
      </div>

      <!-- Selected Audit Context Banner -->
      <div class="px-3 py-2 rounded-lg bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700/80 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs">
        <div class="flex items-center gap-2">
          <UIcon
            name="i-lucide-target"
            class="size-4 text-primary-500 shrink-0"
          />
          <span class="text-gray-600 dark:text-gray-400">Fokus Kertas Kerja Saat Ini:</span>
          <strong class="font-mono text-primary-600 dark:text-primary-400 text-xs sm:text-sm">{{ activeAssignmentData.letterNumber }}</strong>
          <span class="text-gray-500 dark:text-gray-400 hidden sm:inline">&mdash; {{ activeAssignmentData.auditTitle }}</span>
        </div>
        <div class="text-gray-500 dark:text-gray-400 text-right">
          Unit: <strong class="text-gray-800 dark:text-gray-200">{{ activeAssignmentData.workingUnit }}</strong>
        </div>
      </div>

      <!-- 5 Interactive KKA Cards (F-01 s/d F-05) -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-3">
        <!-- F-01: Header & Latar Belakang -->
        <div class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-2.5 hover:border-primary-500 transition-all group">
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <UBadge
                :color="kkaFormStatus.f01 ? 'success' : 'neutral'"
                variant="soft"
                size="xs"
              >
                {{ kkaFormStatus.f01 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Header & Latar Belakang
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Proses bisnis, ruang lingkup, lokasi audit, serta tanggal mulai & selesai pengujian.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f01')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-02: Identifikasi Risiko & Pengendalian -->
        <div class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-2.5 hover:border-primary-500 transition-all group">
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <UBadge
                :color="kkaFormStatus.f02 ? 'success' : 'neutral'"
                variant="soft"
                size="xs"
              >
                {{ kkaFormStatus.f02 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Identifikasi Risiko & RCM
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Pemetaan taksonomi risiko, risk level, dan deskripsi sistem pengendalian internal terkait.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f02')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-03: Pengujian Sampel & Bukti -->
        <div class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-2.5 hover:border-primary-500 transition-all group">
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <UBadge
                :color="kkaFormStatus.f03 ? 'success' : 'neutral'"
                variant="soft"
                size="xs"
              >
                {{ kkaFormStatus.f03 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Uji Sampel & Bukti
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Populasi audit, penentuan sample size, pengujian substantif/kepatuhan, dan bukti dokumen.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f03')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-04: Analisis Sebab & Kondisi -->
        <div class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-2.5 hover:border-primary-500 transition-all group">
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <UBadge
                :color="kkaFormStatus.f04 ? 'success' : 'neutral'"
                variant="soft"
                size="xs"
              >
                {{ kkaFormStatus.f04 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Sebab & Kondisi (RCA)
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Kondisi riil vs kriteria standar, analisis akar masalah (5-Whys/Fishbone), dan dampak risiko.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f04')"
          >
            Buka Form &rarr;
          </UButton>
        </div>

        <!-- F-05: Rencana Aksi & Rekomendasi -->
        <div class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex flex-col justify-between space-y-2.5 hover:border-primary-500 transition-all group">
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <UBadge
                :color="kkaFormStatus.f05 ? 'success' : 'neutral'"
                variant="soft"
                size="xs"
              >
                {{ kkaFormStatus.f05 ? 'Lengkap' : 'Draft' }}
              </UBadge>
            </div>
            <h4 class="font-bold text-sm text-gray-900 dark:text-white group-hover:text-primary-600 transition-colors">
              Rekomendasi & Rencana Aksi
            </h4>
            <p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">
              Rekomendasi perbaikan auditor, respons auditee, penunjukan PIC, dan target penyelesaian.
            </p>
          </div>
          <UButton
            size="xs"
            color="primary"
            variant="solid"
            block
            class="font-bold cursor-pointer"
            @click="openKkaStep('f05')"
          >
            Buka Form &rarr;
          </UButton>
        </div>
      </div>
    </div>

    <!-- 2 COLUMN LAYOUT: DRAFT FINDINGS & PBC TRACKER (hidden in auditor view, see showDraftAndPbcWidgets) -->
    <div
      v-if="showDraftAndPbcWidgets"
      class="grid grid-cols-1 lg:grid-cols-12 gap-4"
    >
      <!-- Draft Findings & LHA (7 COLS) -->
      <div class="lg:col-span-7 p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon
                name="i-lucide-alert-triangle"
                class="size-5 text-amber-500"
              />
              Widget Draft Temuan & LHA
            </h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Ringkasan temuan sementara yang sedang disusun sebelum finalisasi laporan.
            </p>
          </div>

          <NuxtLink
            to="/audit-result-report"
            class="text-xs font-bold text-primary-600 dark:text-primary-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
          >
            <span>Buka Modul LHA (Draft)</span>
            <span>&rarr;</span>
          </NuxtLink>
        </div>

        <!-- Warning Callout: Wewenang Auditor (No Publish Button) -->
        <div class="px-3 py-2 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-900 dark:text-amber-200 text-xs flex items-start gap-2.5">
          <UIcon
            name="i-lucide-lock"
            class="size-4 text-amber-600 dark:text-amber-400 shrink-0 mt-0.5"
          />
          <div>
            <span class="font-bold">Otorisasi Laporan:</span>
            <span class="ml-1 text-amber-800 dark:text-amber-300">
              Hak akses Auditor adalah <strong>DRAFT</strong> (Row 9-1). Tombol finalisasi, otorisasi, dan publikasi LHA resmi merupakan wewenang Audit Manager & Chief Audit Executive (CAE).
            </span>
          </div>
        </div>

        <!-- Findings Category Pill Summary -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-xs">
          <div class="p-2.5 rounded-xl bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-900/40">
            <span class="block text-lg font-extrabold text-red-600 dark:text-red-400">{{ highFindingsCount }}</span>
            <span class="text-red-700 dark:text-red-300 text-[11px] font-semibold">Sangat Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40">
            <span class="block text-lg font-extrabold text-amber-600 dark:text-amber-400">{{ mediumFindingsCount }}</span>
            <span class="text-amber-700 dark:text-amber-300 text-[11px] font-semibold">Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-900/40">
            <span class="block text-lg font-extrabold text-blue-600 dark:text-blue-400">{{ lowFindingsCount }}</span>
            <span class="text-blue-700 dark:text-blue-300 text-[11px] font-semibold">Cukup Signifikan</span>
          </div>
          <div class="p-2.5 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
            <span class="block text-lg font-extrabold text-gray-700 dark:text-gray-300">{{ notSignificantCount }}</span>
            <span class="text-gray-600 dark:text-gray-400 text-[11px] font-semibold">Tidak Signifikan</span>
          </div>
        </div>

        <!-- Draft Findings Table -->
        <div class="overflow-x-auto rounded-xl border border-gray-200 dark:border-gray-800">
          <table class="w-full text-left text-xs">
            <thead class="bg-gray-50 dark:bg-gray-800 text-gray-700 dark:text-gray-300 uppercase font-bold text-[11px]">
              <tr>
                <th class="py-2.5 px-3">
                  Judul Temuan Sementara
                </th>
                <th class="py-2.5 px-3">
                  Kategori
                </th>
                <th class="py-2.5 px-3">
                  Sumber Kertas Kerja
                </th>
                <th class="py-2.5 px-3 text-center">
                  Status
                </th>
                <th class="py-2.5 px-3 text-center">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
              <tr
                v-for="(finding, idx) in draftFindingsList"
                :key="idx"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors"
              >
                <td class="py-2.5 px-3">
                  <span class="font-semibold text-gray-900 dark:text-white block line-clamp-1">
                    {{ finding.title }}
                  </span>
                  <span class="text-[10px] text-gray-500 dark:text-gray-400">
                    {{ finding.assignmentLetterId || activeAssignmentData.letterNumber }}
                  </span>
                </td>
                <td class="py-2.5 px-3">
                  <UBadge
                    :color="getFindingCategoryColor(finding.category)"
                    variant="subtle"
                    size="xs"
                  >
                    {{ finding.category }}
                  </UBadge>
                </td>
                <td class="py-2.5 px-3 text-gray-600 dark:text-gray-400">
                  {{ finding.source || 'Test Controls' }}
                </td>
                <td class="py-2.5 px-3 text-center">
                  <UBadge
                    color="neutral"
                    variant="soft"
                    size="xs"
                  >
                    Draft Auditor
                  </UBadge>
                </td>
                <td class="py-2.5 px-3 text-center">
                  <UButton
                    size="xs"
                    color="primary"
                    variant="ghost"
                    icon="i-lucide-pencil"
                    title="Edit Draft Temuan"
                    @click="goToResultReport(finding.assignmentLetterId || activeAssignmentData.letterNumber)"
                  />
                </td>
              </tr>

              <tr v-if="draftFindingsList.length === 0">
                <td
                  colspan="5"
                  class="py-6 text-center text-gray-500 dark:text-gray-400"
                >
                  Belum ada draft temuan untuk penugasan ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex items-center justify-between">
          <span class="text-xs text-gray-500 dark:text-gray-400">
            Draft siap diajukan ke review supervisor/manager saat Kertas Kerja selesai.
          </span>
          <UButton
            to="/audit-result-report"
            color="primary"
            variant="outline"
            size="sm"
            icon="i-lucide-file-text"
            class="font-semibold"
          >
            Lanjutkan Penyusunan LHA (Draft)
          </UButton>
        </div>
      </div>

      <!-- PBC TRACKER & DOKUMEN BUKTI (5 COLS) -->
      <div class="lg:col-span-5 p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon
                name="i-lucide-folder-check"
                class="size-5 text-sky-500"
              />
              Permintaan Bukti Klien (PBC)
            </h2>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Monitoring dokumen dan data yang diminta ke Auditee.
            </p>
          </div>

          <NuxtLink
            to="/audit-fieldwork"
            class="text-xs font-bold text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1"
          >
            <span>Fieldwork</span>
            <span>&rarr;</span>
          </NuxtLink>
        </div>

        <!-- PBC Summary Stats -->
        <div class="grid grid-cols-3 gap-2 text-center text-xs">
          <div class="p-2 rounded-xl bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700">
            <span class="block text-base font-bold text-gray-900 dark:text-white">{{ pbcStats.total }}</span>
            <span class="text-[10px] text-gray-500">Diminta</span>
          </div>
          <div class="p-2 rounded-xl bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-900/40">
            <span class="block text-base font-bold text-emerald-600 dark:text-emerald-400">{{ pbcStats.collected }}</span>
            <span class="text-[10px] text-emerald-700 dark:text-emerald-300 font-semibold">Diterima</span>
          </div>
          <div class="p-2 rounded-xl bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-900/40">
            <span class="block text-base font-bold text-amber-600 dark:text-amber-400">{{ pbcStats.pending }}</span>
            <span class="text-[10px] text-amber-700 dark:text-amber-300 font-semibold">Menunggu</span>
          </div>
        </div>

        <!-- PBC Item List -->
        <div class="space-y-2 max-h-[290px] overflow-y-auto pr-1">
          <div
            v-for="(pbc, idx) in pbcItemsList"
            :key="idx"
            class="px-3 py-2 rounded-lg border border-gray-100 dark:border-gray-800 bg-gray-50/60 dark:bg-gray-800/40 flex items-start justify-between gap-3 text-xs"
          >
            <div class="min-w-0">
              <span class="font-semibold text-gray-900 dark:text-white block line-clamp-1">
                {{ pbc.documentName }}
              </span>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1">
                {{ pbc.description || 'Tidak ada keterangan tambahan' }}
              </p>
              <span class="text-[10px] text-gray-400 mt-1 block">
                Deadline: {{ formatDisplayDate(pbc.requiredDate) }}
              </span>
            </div>

            <UBadge
              :color="pbc.status === 'Received' ? 'success' : 'warning'"
              variant="subtle"
              size="xs"
              class="shrink-0 font-semibold"
            >
              {{ pbc.status === 'Received' ? 'Diterima' : 'Menunggu' }}
            </UBadge>
          </div>

          <div
            v-if="pbcItemsList.length === 0"
            class="py-5 text-center text-gray-500 dark:text-gray-400"
          >
            <UIcon
              name="i-lucide-folder-open"
              class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2"
            />
            <p class="text-xs font-medium">
              Belum ada daftar permintaan bukti (PBC) untuk penugasan ini.
            </p>
            <p class="text-[11px] text-gray-400 mt-0.5">
              Kelola permintaan dokumen di modul Audit Fieldwork.
            </p>
          </div>
        </div>

        <!-- Quick Link to Fieldwork Documents Tab -->
        <UButton
          to="/audit-fieldwork"
          variant="outline"
          color="neutral"
          size="sm"
          block
          icon="i-lucide-arrow-right"
          class="font-semibold cursor-pointer"
        >
          Kelola Permintaan Bukti di Audit Fieldwork
        </UButton>
      </div>
    </div>

    <!-- RCM QUICK REFERENCE (READ-ONLY) -->
    <div class="p-4 bg-white dark:bg-gray-900 border border-gray-200/80 dark:border-gray-800 rounded-xl shadow-xs space-y-3">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2 class="text-base font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-book-open"
              class="size-5 text-indigo-500"
            />
            Referensi Cepat: Risk Control Matrix (RCM) & Kriteria Audit (Read-Only)
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            Akses rujukan risiko dan standar kontrol (Hak Akses Auditor: READ & FULL RCM) untuk penyusunan kertas kerja pengujian.
          </p>
        </div>
        <NuxtLink
          to="/risk-profile/risk-control-matrix"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Buka Modul RCM Lengkap</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div
          v-for="(rcm, idx) in rcmQuickReferences"
          :key="idx"
          class="p-3 rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 text-xs space-y-1"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono font-bold text-primary-600 dark:text-primary-400 text-[11px]">{{ rcm.code }}</span>
            <UBadge
              :color="rcm.riskLevel === 'High' ? 'error' : (rcm.riskLevel === 'Moderate' ? 'warning' : 'info')"
              variant="soft"
              size="xs"
            >
              {{ rcm.riskLevel }} Risk
            </UBadge>
          </div>
          <h4 class="font-bold text-gray-900 dark:text-white line-clamp-1">
            {{ rcm.process }}
          </h4>
          <p class="text-[11px] text-gray-600 dark:text-gray-300 line-clamp-2">
            <strong class="text-gray-700 dark:text-gray-200">Kontrol:</strong> {{ rcm.controlDescription }}
          </p>
        </div>

        <div
          v-if="rcmQuickReferences.length === 0"
          class="col-span-full py-4 text-center text-gray-500 dark:text-gray-400"
        >
          <UIcon
            name="i-lucide-book-open"
            class="size-8 mx-auto text-gray-300 dark:text-gray-600 mb-2"
          />
          <p class="text-xs font-medium">
            Belum ada data Risk Control Matrix (RCM) yang tercatat.
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '~/stores/auth'
import { useRbac } from '~/composables/useRbac'
import { useI18n } from '~/composables/useI18n'
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { useAuditFieldworkStore } from '~/stores/audit-fieldwork'
import { useWorkingPaperStore } from '~/stores/working-paper'
import { useAuditResultReportStore } from '~/stores/audit-result-report'
import { useRCMStore } from '~/stores/rcm'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useActionTakenReportStore } from '~/stores/action-taken-report'
import { useAuditExecutionStore } from '~/stores/audit-execution'
import { AnnualAuditPlanStatus, type ActionTakenReport } from '~/types/audit'
import { getProgressValue } from '~/utils/executionSources'
import {
  atrDateOnly,
  atrStatusBadgeColor,
  atrStatusI18nKey,
  formatAtrDate,
  isAtrClosed,
  jakartaToday
} from '~/utils/actionTakenReport'
import { makeCalendarEvent, type CalendarEvent } from '~/utils/calendarEvents'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import AuditorWidgetState from './AuditorWidgetState.vue'
import CalendarWidget from './CalendarWidget.vue'

type WidgetState = 'loading' | 'error' | 'empty' | 'ready'
type FieldworkStatusKey = 'preparation' | 'fieldwork' | 'reporting'
type StageKey = 'preparation' | 'testing' | 'workingPapers' | 'reporting'

const DAY_MS = 24 * 60 * 60 * 1000
const DUE_SOON_DAYS = 3
const TASK_LIMIT = 10
const NOTIFICATION_LIMIT = 10
// Stage 1..4 maps to these keys (index = stageStep - 1).
const STAGE_KEYS: StageKey[] = ['preparation', 'testing', 'workingPapers', 'reporting']

const router = useRouter()
const toast = useToastNotification()
const { t, locale } = useI18n()
const authStore = useAuthStore()
const { isAdmin, canManageAudits, setRoleView } = useRbac()

const assignmentLetterStore = useAssignmentLetterStore()
const fieldworkStore = useAuditFieldworkStore()
const workingPaperStore = useWorkingPaperStore()
const auditResultStore = useAuditResultReportStore()
const rcmStore = useRCMStore()
const annualPlanStore = useAnnualPlanStore()
const atrStore = useActionTakenReportStore()
const auditExecutionStore = useAuditExecutionStore()

const isSyncing = ref(false)
const initialLoadDone = ref(false)
const fieldworkSearch = ref('')
const selectedLetterNumber = ref<string>('')

// Current User Info
const currentUserFullName = computed(() => {
  return authStore.user?.fullName || authStore.user?.username || 'Auditor SKAI'
})

// Linked employee (spec B1). Until /auth/me returns it, team-based widgets stay in their placeholder state.
const employeeLinked = computed(() => !!authStore.user?.employeeId)

// Draft findings / LHA and PBC widgets are hidden in the auditor view for now.
// Their data sources (draftFindingsList, pbcItemsList, ...) stay in place so the block can be re-enabled by flipping this flag.
const showDraftAndPbcWidgets = false

// Can the current user toggle role views (Admin or Manager testing)
const canToggleRoleView = computed(() => {
  return isAdmin.value || canManageAudits.value
})

const switchDashboardView = (view: string | null) => {
  setRoleView(view)
  toast.info(view ? `Beralih ke mode pratinjau: ${view}` : 'Tampilan dashboard dikembalikan ke default')
}

// Maps a store's loading/error/count to a widget state.
// Before the first load finishes an empty source counts as loading, so the page does not flash "no data".
const sourceState = (loading: boolean, error: string | null | undefined, count: number): WidgetState => {
  if (count === 0 && (loading || !initialLoadDone.value)) return 'loading'
  if (count === 0 && error) return 'error'
  return count === 0 ? 'empty' : 'ready'
}

// Reload handlers used by the per-widget retry buttons
const reloadPlans = () => annualPlanStore.fetchPlans()
const reloadAtr = () => atrStore.fetchSummary()
const reloadExecutions = () => auditExecutionStore.fetchAuditExecutions()
const reloadLetters = () => assignmentLetterStore.fetchAssignmentLetters()

// Helper to count how many KKA forms are completed for a specific assignment letter
const getCompletedKkaFormsCount = (letterNumber: string, letterId?: string) => {
  const isMatch = (itemLetter: any) => {
    return itemLetter === letterNumber || (letterId && itemLetter === letterId)
  }

  const hasF01 = workingPaperStore.dataF01.some((d: any) => isMatch(d.assignmentLetterId))
  const hasF02 = workingPaperStore.dataF02.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF03 = workingPaperStore.dataF03.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF04 = workingPaperStore.dataF04.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))
  const hasF05 = workingPaperStore.dataF05.some((d: any) => isMatch(d.workingPaperId || d.assignmentLetterId))

  return (hasF01 ? 1 : 0) + (hasF02 ? 1 : 0) + (hasF03 ? 1 : 0) + (hasF04 ? 1 : 0) + (hasF05 ? 1 : 0)
}

// ─── Assignment letters (engagements) ──────────────────────────────────────
// Interim scope: every published letter. Team-scoped list comes from GET /api/v1/me/engagements (B3).
const combinedFieldworkList = computed(() => {
  return assignmentLetterStore.assignmentLetterList
    .filter((st: any) => st.status === 'Published')
    .map((st: any) => {
      const kkaCompleted = getCompletedKkaFormsCount(st.letterNumber, st.id)

      // Determine the real audit stage from the data lifecycle:
      // 1. Has draft/final LHA report -> Stage 4 (Exit / Reporting)
      // 2. Has Working Paper forms -> Stage 3 (Working papers)
      // 3. Has Fieldwork observations/interviews/samples/documents -> Stage 2 (Testing)
      // 4. Default -> Stage 1 (Entry / Preparation)
      const hasReport = auditResultStore.reportList.some(
        (r: any) => r.assignmentLetterId === st.letterNumber || (st.id && r.assignmentLetterId === st.id)
      )

      const fwData = fieldworkStore.fieldworkData?.[st.letterNumber]
      const hasFieldworkData = !!(
        fwData?.interviews?.length
        || fwData?.observations?.length
        || fwData?.documents?.length
        || fwData?.samples?.length
        || fwData?.testControls?.length
      )

      let stageStep = 1
      if (hasReport) {
        stageStep = 4
      } else if (kkaCompleted > 0) {
        stageStep = 3
      } else if (hasFieldworkData) {
        stageStep = 2
      }

      const statusKey: FieldworkStatusKey = stageStep === 4 ? 'reporting' : (stageStep >= 2 ? 'fieldwork' : 'preparation')

      return {
        letterNumber: st.letterNumber,
        auditTitle: st.auditTitle,
        workingUnit: st.workingUnit || t('dashboard.auditor.status.unitFallback'),
        auditTeam: st.auditTeam || 'SKAI',
        category: st.category || 'Assurance',
        startPeriod: st.startPeriod || '',
        finishPeriod: st.finishPeriod || '',
        statusKey,
        stageStep,
        stageKey: STAGE_KEYS[stageStep - 1] as StageKey,
        kkaCompleted
      }
    })
})

// Filtered fieldwork list by search query
const filteredFieldworkList = computed(() => {
  const q = fieldworkSearch.value.toLowerCase().trim()
  if (!q) return combinedFieldworkList.value
  return combinedFieldworkList.value.filter((item) => {
    return (
      (item.letterNumber || '').toLowerCase().includes(q)
      || (item.auditTitle || '').toLowerCase().includes(q)
      || (item.workingUnit || '').toLowerCase().includes(q)
    )
  })
})

// Assignment letter dropdown options
const assignmentLetterOptions = computed(() => {
  return combinedFieldworkList.value.map(item => ({
    label: `${item.letterNumber || '-'} - ${item.auditTitle || 'Penugasan Audit'}`,
    value: item.letterNumber
  }))
})

// Currently selected assignment letter details
const activeAssignmentData = computed(() => {
  const target = selectedLetterNumber.value || combinedFieldworkList.value[0]?.letterNumber
  const found = combinedFieldworkList.value.find(i => i.letterNumber === target)
  return found || combinedFieldworkList.value[0] || {
    letterNumber: '-',
    auditTitle: 'Belum ada surat tugas yang dipilih',
    workingUnit: '-',
    auditTeam: '-',
    category: '-',
    startPeriod: '',
    finishPeriod: '',
    statusKey: 'preparation' as FieldworkStatusKey,
    stageStep: 0,
    stageKey: 'preparation' as StageKey,
    kkaCompleted: 0
  }
})

const statusState = computed<WidgetState>(() => {
  const base = sourceState(
    assignmentLetterStore.loading,
    assignmentLetterStore.errorMsg,
    assignmentLetterStore.assignmentLetterList.length
  )
  if (base !== 'ready') return base
  return combinedFieldworkList.value.length === 0 ? 'empty' : 'ready'
})

// ─── Audit statistics (scope: all data the user can access, interim) ─────
const plannedAuditCount = computed(
  () => annualPlanStore.plans.filter(p => p.status !== AnnualAuditPlanStatus.APPROVED).length
)
// Open findings = ATR items not COMPLETED and not CANCELLED (see utils/actionTakenReport).
const openFindingsCount = computed(() => atrStore.stats.counts.open)
const executionPercent = computed(() => {
  const executions = auditExecutionStore.auditExecutions
  if (executions.length === 0) return 0
  return Math.round(executions.reduce((sum, e) => sum + getProgressValue(e.progress), 0) / executions.length)
})
// ATR compliance = completed / all items (same basis as the admin donut).
const atrCompliancePercent = computed(() => Math.round((atrStore.stats.compliance || 0) * 100))

const planState = computed<WidgetState>(() =>
  sourceState(annualPlanStore.loading, annualPlanStore.errorMsg, annualPlanStore.plans.length)
)
const atrState = computed<WidgetState>(() =>
  sourceState(atrStore.summaryLoading, atrStore.summaryError, atrStore.summaryItems.length)
)
const executionState = computed<WidgetState>(() =>
  sourceState(auditExecutionStore.loading, auditExecutionStore.error, auditExecutionStore.auditExecutions.length)
)

const statCards = computed(() => [
  {
    key: 'planned',
    label: t('dashboard.auditor.stats.plannedAudits'),
    hint: t('dashboard.auditor.stats.plannedAuditsHint'),
    value: String(plannedAuditCount.value),
    state: planState.value,
    retry: reloadPlans,
    icon: 'i-lucide-calendar-check',
    iconClass: 'text-indigo-600 dark:text-indigo-400',
    iconBg: 'bg-indigo-50 dark:bg-indigo-500/10'
  },
  {
    key: 'open',
    label: t('dashboard.auditor.stats.openFindings'),
    hint: t('dashboard.auditor.stats.openFindingsHint'),
    value: String(openFindingsCount.value),
    state: atrState.value,
    retry: reloadAtr,
    icon: 'i-lucide-alert-triangle',
    iconClass: 'text-amber-600 dark:text-amber-400',
    iconBg: 'bg-amber-50 dark:bg-amber-500/10'
  },
  {
    key: 'execution',
    label: t('dashboard.auditor.stats.executionProgress'),
    hint: t('dashboard.auditor.stats.executionProgressHint'),
    value: `${executionPercent.value}%`,
    state: executionState.value,
    retry: reloadExecutions,
    icon: 'i-lucide-activity',
    iconClass: 'text-emerald-600 dark:text-emerald-400',
    iconBg: 'bg-emerald-50 dark:bg-emerald-500/10'
  },
  {
    key: 'compliance',
    label: t('dashboard.auditor.stats.atrCompliance'),
    hint: t('dashboard.auditor.stats.atrComplianceHint'),
    value: `${atrCompliancePercent.value}%`,
    state: atrState.value,
    retry: reloadAtr,
    icon: 'i-lucide-check-circle',
    iconClass: 'text-sky-600 dark:text-sky-400',
    iconBg: 'bg-sky-50 dark:bg-sky-500/10'
  }
])

// ─── Audit planning & coverage (PKAT) ─────────────────────────────────────
// Planned = every annual audit plan, Approved = status APPROVED, Remaining = everything else.
const auditCoverage = computed(() => {
  const plannedAudits = annualPlanStore.plans.length
  const completedAudits = annualPlanStore.plans.filter(p => p.status === AnnualAuditPlanStatus.APPROVED).length
  return {
    plannedAudits,
    completedAudits,
    remainingAudits: plannedAudits - completedAudits
  }
})
const coveragePercent = computed(() => {
  const total = auditCoverage.value.plannedAudits
  if (total === 0) return 0
  return Math.round((auditCoverage.value.completedAudits / total) * 100)
})

// ─── ATR-based tasks and reminders (interim sources) ──────────────────────
// TODO(B5): replace with GET /api/v1/me/tasks (audit-service). Interim: open ATR items the user can access.
const compareAtrByUrgency = (a: ActionTakenReport, b: ActionTakenReport) => {
  if (a.is_overdue !== b.is_overdue) return a.is_overdue ? -1 : 1
  const da = atrDateOnly(a.due_date)
  const db = atrDateOnly(b.due_date)
  if (da && db) return da.localeCompare(db)
  if (da) return -1
  if (db) return 1
  return 0
}

const openAtrItems = computed(() =>
  atrStore.summaryItems
    .filter(item => !isAtrClosed(item.status))
    .sort(compareAtrByUrgency)
)

const tasksState = computed<WidgetState>(() => {
  if (atrState.value !== 'ready') return atrState.value
  return openAtrItems.value.length === 0 ? 'empty' : 'ready'
})

const taskRows = computed(() => openAtrItems.value.slice(0, TASK_LIMIT))

const atrStatusLabel = (status: unknown) => {
  const key = atrStatusI18nKey(status)
  return key ? t(key) : String(status ?? '-')
}

// Whole days from today (Asia/Jakarta) to a due date. NaN when there is no due date.
const daysUntil = (value: unknown): number => {
  const due = atrDateOnly(value)
  if (!due) return Number.NaN
  return Math.round((Date.parse(`${due}T00:00:00Z`) - Date.parse(`${jakartaToday()}T00:00:00Z`)) / DAY_MS)
}

// ─── Calendar of important dates (interim sources) ────────────────────────
// TODO(B4): replace with GET /api/v1/me/calendar. Interim: published Surat Tugas periods, the fieldwork dates of the
// letters already loaded on this page, and open ATR due dates. Unparseable dates are skipped.
const calendarEvents = computed<CalendarEvent[]>(() => {
  const events: CalendarEvent[] = []
  const push = (event: CalendarEvent | null) => {
    if (event) events.push(event)
  }

  for (const st of combinedFieldworkList.value) {
    push(makeCalendarEvent(
      { id: `letter:${st.letterNumber}`, kind: 'engagement', title: st.auditTitle || st.letterNumber, subtitle: st.letterNumber, to: '/assignment-letter' },
      st.startPeriod,
      st.finishPeriod
    ))
  }

  const fieldworkData = fieldworkStore.fieldworkData || {}
  for (const [letter, fw] of Object.entries(fieldworkData)) {
    if (!fw) continue
    const base = (kind: CalendarEvent['kind'], id: unknown, title: string) =>
      ({ id: `${kind}:${letter}:${id}`, kind, title: title || letter, subtitle: letter, to: '/audit-fieldwork' })
    fw.interviews?.forEach(i => push(makeCalendarEvent(base('interview', i.id, i.topic || i.interviewee), i.date)))
    fw.observations?.forEach(o => push(makeCalendarEvent(base('observation', o.id, o.activity), o.date)))
    fw.samples?.forEach(s => push(makeCalendarEvent(base('sample', s.id, s.documentName), s.date)))
    fw.documents?.forEach(d => push(makeCalendarEvent(base('pbcDue', d.id, d.documentName), d.requiredDate)))
    fw.testControls?.forEach(c => push(makeCalendarEvent(base('testControlDue', c.id, c.controlName), c.dueDate)))
  }

  for (const item of openAtrItems.value) {
    push(makeCalendarEvent(
      {
        id: `atr:${item.id}`,
        kind: 'atrDue',
        title: item.finding_title || item.report_number || '-',
        subtitle: item.report_number || undefined,
        overdue: item.is_overdue,
        to: '/action-taken-report'
      },
      item.due_date
    ))
  }
  return events
})

// Only an error when every source failed; one failed source still leaves a usable calendar.
const calendarErrorText = computed(() =>
  assignmentLetterStore.errorMsg && atrStore.summaryError
    ? t('dashboard.calendarWidget.loadError')
    : ''
)

const reloadCalendarSources = () => Promise.allSettled([reloadLetters(), reloadAtr()])

// TODO(B9): replace with GET /api/v1/notifications?unread=true (audit-service). Interim: derived from ATR data.
// Only ATR-based types are derived here; PBC requests, comments and revisions are not available yet.
type DerivedNotification = {
  id: string
  kind: 'overdue' | 'pendingReview' | 'dueSoon'
  days: number
  item: ActionTakenReport
}

const derivedNotifications = computed<DerivedNotification[]>(() => {
  const list: DerivedNotification[] = []
  for (const item of openAtrItems.value) {
    if (item.is_overdue) {
      list.push({ id: item.id, kind: 'overdue', days: item.overdue_days, item })
    } else if (item.status === 'PENDING_REVIEW') {
      list.push({ id: item.id, kind: 'pendingReview', days: 0, item })
    } else {
      const days = daysUntil(item.due_date)
      if (days >= 0 && days <= DUE_SOON_DAYS) {
        list.push({ id: item.id, kind: 'dueSoon', days, item })
      }
    }
  }
  return list.slice(0, NOTIFICATION_LIMIT)
})

const notificationsState = computed<WidgetState>(() => {
  if (atrState.value !== 'ready') return atrState.value
  return derivedNotifications.value.length === 0 ? 'empty' : 'ready'
})

const notificationText = (n: DerivedNotification) => {
  if (n.kind === 'overdue') return t('dashboard.auditor.notifications.typeOverdue', { days: n.days })
  if (n.kind === 'pendingReview') return t('dashboard.auditor.notifications.typePendingReview')
  return t('dashboard.auditor.notifications.typeDueSoon', { days: n.days })
}

// Real Draft Findings (Auditor Scope: Live Working Paper Causes + Ineffective Controls + Draft LHA)
const draftFindingsList = computed(() => {
  const targetLetter = selectedLetterNumber.value
  const list: Array<{
    title: string
    category: 'Very Significant' | 'Significant' | 'Quite Significant' | 'Not Significant'
    assignmentLetterId: string
    source: string
    status: string
  }> = []
  const seen = new Set<string>()

  const addFinding = (title: string, category: any, letterId: string, source: string) => {
    const cleanTitle = (title || '').trim()
    if (!cleanTitle) return
    const key = `${letterId}__${cleanTitle.toLowerCase()}`
    if (seen.has(key)) return
    seen.add(key)
    list.push({
      title: cleanTitle,
      category: (category as any) || 'Significant',
      assignmentLetterId: letterId,
      source: source || 'Kertas Kerja Audit',
      status: 'Draft'
    })
  }

  // 1. Findings from saved reports (Draft status only)
  auditResultStore.reportList
    .filter(r => r.status === 'Draft' && (!targetLetter || r.assignmentLetterId === targetLetter))
    .forEach((report) => {
      if (Array.isArray(report.findings)) {
        report.findings.forEach((f) => {
          addFinding(f.title, f.category, report.assignmentLetterId, 'Draft LHA')
        })
      }
    })

  // 2. Real-time auto-detected findings from KKA & Fieldwork for selected letter
  if (targetLetter) {
    if (typeof auditResultStore.getLocalAutoFindings === 'function') {
      try {
        const autoFindings = auditResultStore.getLocalAutoFindings(targetLetter)
        if (Array.isArray(autoFindings)) {
          autoFindings.forEach((f) => {
            addFinding(f.title, f.category, targetLetter, f.source || 'Digital Working Paper (KKA)')
          })
        }
      } catch (err) {
        console.warn('Error reading auto findings for letter:', targetLetter, err)
      }
    }
  } else {
    combinedFieldworkList.value.forEach((item) => {
      if (typeof auditResultStore.getLocalAutoFindings === 'function') {
        try {
          const autoFindings = auditResultStore.getLocalAutoFindings(item.letterNumber)
          if (Array.isArray(autoFindings)) {
            autoFindings.forEach((f) => {
              addFinding(f.title, f.category, item.letterNumber, f.source || 'Digital Working Paper (KKA)')
            })
          }
        } catch (err) {
          console.warn('Error reading auto findings for letter:', item.letterNumber, err)
        }
      }
    })
  }

  // 3. Recent findings from backend API matching target letter
  if (Array.isArray(auditResultStore.recentFindings)) {
    auditResultStore.recentFindings
      .filter(rf => !targetLetter || rf.assignmentLetterId === targetLetter)
      .forEach((rf) => {
        addFinding(rf.title, rf.category, rf.assignmentLetterId || targetLetter || '-', rf.source || 'Temuan Lapangan')
      })
  }

  return list
})

const highFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Very Significant').length
})

const mediumFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Significant').length
})

const lowFindingsCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Quite Significant').length
})

const notSignificantCount = computed(() => {
  return draftFindingsList.value.filter(f => f.category === 'Not Significant').length
})

// Real PBC Document Tracker from Fieldwork Store
const pbcItemsList = computed(() => {
  const targetLetter = selectedLetterNumber.value
  if (!targetLetter) {
    return []
  }

  const docList = fieldworkStore.fieldworkData?.[targetLetter]?.documents
    || (fieldworkStore.selectedAssignmentLetter === targetLetter ? fieldworkStore.documents : [])

  if (!Array.isArray(docList)) return []

  return docList.map((doc: any) => {
    const isCollected = !!(doc.file || doc.fileName || doc.filePath || doc.fileUrl)
    return {
      id: doc.id,
      documentName: doc.documentName || 'Dokumen Permintaan Bukti',
      description: doc.description || '',
      requiredDate: doc.requiredDate || '',
      status: isCollected ? 'Received' : 'Pending'
    }
  })
})

const pbcStats = computed(() => {
  const total = pbcItemsList.value.length
  const collected = pbcItemsList.value.filter(item => item.status === 'Received').length
  const pending = total - collected
  return {
    total,
    collected,
    pending
  }
})

// Real KKA Module Completion Status (F-01 s/d F-05) for Selected Assignment Letter
const kkaFormStatus = computed(() => {
  const letter = selectedLetterNumber.value
  if (!letter) {
    return {
      f01: false,
      f02: false,
      f03: false,
      f04: false,
      f05: false
    }
  }

  const hasF01 = workingPaperStore.dataF01.some((d: any) => d.assignmentLetterId === letter)
  const hasF02 = workingPaperStore.dataF02.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF03 = workingPaperStore.dataF03.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF04 = workingPaperStore.dataF04.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)
  const hasF05 = workingPaperStore.dataF05.some((d: any) => (d.workingPaperId || d.assignmentLetterId) === letter)

  return {
    f01: hasF01,
    f02: hasF02,
    f03: hasF03,
    f04: hasF04,
    f05: hasF05
  }
})

// Real RCM Quick References from RCM Store
const rcmQuickReferences = computed(() => {
  if (!rcmStore.rcmList || rcmStore.rcmList.length === 0) return []
  return rcmStore.rcmList.slice(0, 3).map((item: any) => {
    const riskLevelScore = item.inherent_risk || item.total_weighted_score || 0
    let riskLevel = 'Moderate'
    if (riskLevelScore >= 15 || item.design_effectiveness_rating === 1) {
      riskLevel = 'High'
    } else if (riskLevelScore < 8 && item.operating_effectiveness_rating >= 4) {
      riskLevel = 'Low'
    }

    return {
      code: item.control_code || item.risk_code || 'RCM-REF',
      process: item.risk_event || item.department || 'Proses Operasional',
      riskLevel,
      controlDescription: item.control_description || 'Pengendalian internal standar operasional.'
    }
  })
})

// Navigation helpers
const goToWorkingPaper = (letterNumber: string) => {
  fieldworkStore.selectedAssignmentLetter = letterNumber
  workingPaperStore.selectedAssignmentLetter = letterNumber
  router.push('/working-paper')
}

const openKkaStep = (stepSlot: string) => {
  if (selectedLetterNumber.value) {
    fieldworkStore.selectedAssignmentLetter = selectedLetterNumber.value
    workingPaperStore.selectedAssignmentLetter = selectedLetterNumber.value
  }
  router.push(`/working-paper?step=${stepSlot}`)
}

const goToFieldwork = (letterNumber: string) => {
  fieldworkStore.selectedAssignmentLetter = letterNumber
  router.push('/audit-fieldwork')
}

const goToResultReport = (letterNumber: string) => {
  auditResultStore.selectedAssignmentLetter = letterNumber
  router.push('/audit-result-report')
}

// Helpers
const formatPeriod = (start?: string, end?: string) => {
  if (!start && !end) return '-'
  if (start && end) return `${start} ${t('dashboard.auditor.status.periodTo')} ${end}`
  return start || end || '-'
}

const getDeadlineLabel = (finishPeriod?: string) => {
  if (!finishPeriod) return t('dashboard.auditor.status.deadline.none')
  const end = new Date(finishPeriod)
  if (isNaN(end.getTime())) return '-'
  const now = new Date()
  const diffDays = Math.ceil((end.getTime() - now.getTime()) / DAY_MS)
  if (diffDays < 0) return t('dashboard.auditor.status.deadline.overdue', { days: Math.abs(diffDays) })
  if (diffDays === 0) return t('dashboard.auditor.status.deadline.today')
  return t('dashboard.auditor.status.deadline.remaining', { days: diffDays })
}

const getDeadlineBadgeColor = (finishPeriod?: string): 'error' | 'warning' | 'success' | 'neutral' => {
  if (!finishPeriod) return 'neutral'
  const end = new Date(finishPeriod)
  if (isNaN(end.getTime())) return 'neutral'
  const now = new Date()
  const diffDays = Math.ceil((end.getTime() - now.getTime()) / DAY_MS)
  if (diffDays < 0) return 'error'
  if (diffDays <= 3) return 'error'
  if (diffDays <= 7) return 'warning'
  return 'success'
}

const formatDisplayDate = (d?: string) => {
  if (!d) return 'Tidak ada tenggat'
  try {
    const date = new Date(d)
    if (isNaN(date.getTime())) return d
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
  } catch {
    return d
  }
}

const getFieldworkStatusColor = (statusKey?: FieldworkStatusKey): 'success' | 'info' | 'warning' | 'neutral' => {
  if (statusKey === 'fieldwork') return 'success'
  if (statusKey === 'reporting') return 'info'
  if (statusKey === 'preparation') return 'warning'
  return 'neutral'
}

const getFindingCategoryColor = (cat?: string): 'error' | 'warning' | 'info' | 'neutral' => {
  if (cat === 'Very Significant') return 'error'
  if (cat === 'Significant') return 'warning'
  if (cat === 'Quite Significant') return 'info'
  return 'neutral'
}

let isMounted = false

onBeforeUnmount(() => {
  isMounted = false
})

// Silent initial load for dashboard on mount (no toast notification)
const loadInitialAuditorData = async () => {
  try {
    await Promise.allSettled([
      assignmentLetterStore.fetchAssignmentLetters(),
      workingPaperStore.fetchAllData(),
      auditResultStore.fetchReports(),
      auditResultStore.fetchRecentFindings(10),
      rcmStore.fetchRCMList(),
      annualPlanStore.fetchPlans(),
      atrStore.fetchSummary(),
      auditExecutionStore.fetchAuditExecutions()
    ])
    if (!isMounted) return
    if (!selectedLetterNumber.value && combinedFieldworkList.value[0]) {
      selectedLetterNumber.value = combinedFieldworkList.value[0].letterNumber
    }
    if (selectedLetterNumber.value) {
      fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value).catch(() => {})
    }
  } catch (err) {
    console.warn('Initial auditor dashboard data load warning:', err)
  } finally {
    initialLoadDone.value = true
  }
}

// Manual Sync handler - triggered on explicit button click
const manualSyncAuditorData = async () => {
  isSyncing.value = true
  try {
    await Promise.allSettled([
      assignmentLetterStore.fetchAssignmentLetters(),
      workingPaperStore.fetchAllData(),
      auditResultStore.fetchReports(),
      auditResultStore.fetchRecentFindings(10),
      rcmStore.fetchRCMList(),
      annualPlanStore.fetchPlans(),
      atrStore.fetchSummary(),
      auditExecutionStore.fetchAuditExecutions()
    ])
    if (!isMounted) return
    if (selectedLetterNumber.value) {
      await fieldworkStore.fetchAllFieldworkData(selectedLetterNumber.value, true)
    }
    toast.success(t('dashboard.auditor.syncSuccess'))
  } catch {
    if (isMounted) {
      toast.error(t('dashboard.auditor.syncError'))
    }
  } finally {
    if (isMounted) {
      isSyncing.value = false
    }
  }
}

// Watch selected letter change to safely synchronize stores
watch(
  () => selectedLetterNumber.value,
  (rawVal) => {
    if (!isMounted) return
    const newVal = typeof rawVal === 'string' ? rawVal : (rawVal as any)?.value || ''
    if (newVal) {
      if (fieldworkStore.selectedAssignmentLetter !== newVal) {
        fieldworkStore.selectedAssignmentLetter = newVal
      }
      if (workingPaperStore.selectedAssignmentLetter !== newVal) {
        workingPaperStore.selectedAssignmentLetter = newVal
      }
    }
  }
)

// Auto-select first letter when list is loaded
watch(
  () => combinedFieldworkList.value,
  (list) => {
    if (!isMounted) return
    if (!selectedLetterNumber.value && list && list[0]) {
      selectedLetterNumber.value = list[0].letterNumber
    }
  },
  { immediate: true }
)

onMounted(async () => {
  isMounted = true
  await loadInitialAuditorData()
})
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(6px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}  
</style>
