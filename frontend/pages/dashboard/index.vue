<template>
  <div class="p-4 sm:p-6 min-h-screen overflow-x-hidden">
    <!-- Auditor Specific View (Khusus Pelaksana Audit) -->
    <div v-if="isAuditorView" key="auditor-dashboard-container" class="space-y-6 sm:space-y-8">
      <AuditorDashboard />
    </div>

    <!-- General / Admin / Macro View -->
    <div v-else key="macro-dashboard-container" class="space-y-6 sm:space-y-8">
      <!-- Header Section -->
      <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div class="min-w-0">
          <h1 class="text-2xl sm:text-3xl font-bold text-slate-900 tracking-tight dark:text-white">{{ t('dashboard.overview.dashboard') }}</h1>
          <p class="text-sm text-slate-500 mt-1.5 dark:text-slate-400">
{{ t('dashboard.overview.welcome', { name: authStore.getUser?.fullName || 'Andi' }) }}
          </p>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <UButton
            variant="outline"
            color="neutral"
            block
            class="cursor-pointer border border-slate-200 text-slate-700 shadow-sm font-medium sm:w-auto sm:inline-flex dark:border-slate-700 dark:text-slate-200"
            :loading="isSyncing"
            @click="handleSync"
          >
            <template #leading>
              <UIcon
                name="i-lucide-rotate-cw"
                :class="{ 'animate-spin': isSyncing }"
                class="size-4"
              />
            </template>
            Sync
          </UButton>
        </div>
      </div>

    <!-- Main Metrics Cards Section (First Row) -->
    <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
      <!-- Card 1: Total Risks -->
      <div
        class="p-5 sm:p-6 flex flex-col justify-between min-h-[135px] dashboard-card"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500 dark:text-slate-400">{{ t('riskAppetite.kpi.totalRisks') }}</p>
          <div class="rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center dark:bg-indigo-500/10">
            <UIcon name="i-lucide-shield" class="text-indigo-600 size-5 dark:text-indigo-400" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-2xl sm:text-3xl font-bold text-slate-900 tracking-tight dark:text-white">
            {{ totalRisks }}
          </h3>
        </div>
      </div>

      <!-- Card 2: High Risk -->
      <div
        class="p-5 sm:p-6 flex flex-col justify-between min-h-[135px] dashboard-card"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500 dark:text-slate-400">{{ t('auditActivityPlan.form.highRisk') }}</p>
          <div class="rounded-xl bg-red-50 p-2.5 flex items-center justify-center dark:bg-red-500/10">
            <UIcon name="i-lucide-alert-triangle" class="text-red-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-2xl sm:text-3xl font-bold text-red-600 tracking-tight">{{ highRisks }}</h3>
        </div>
      </div>

      <!-- Card 3: Audit Plans -->
      <div
        class="p-5 sm:p-6 flex flex-col justify-between min-h-[135px] dashboard-card"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500 dark:text-slate-400">{{ t('dashboard.overview.auditPlans') }}</p>
          <div class="rounded-xl bg-sky-50 p-2.5 flex items-center justify-center dark:bg-sky-500/10">
            <UIcon name="i-lucide-calendar" class="text-sky-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-2xl sm:text-3xl font-bold text-slate-900 tracking-tight dark:text-white">
            {{ auditPlansCount }}
          </h3>
          <div class="flex items-center gap-1 text-md font-semibold text-sky-500 mt-1">
            <span>{{ t('auditCharter.form.active') }}</span>
            <span class="text-slate-400 font-normal">{{ t('dashboard.overview.thisQuarter') }}</span>
          </div>
        </div>
      </div>

      <!-- Card 4: Completed -->
      <div
        class="p-5 sm:p-6 flex flex-col justify-between min-h-[135px] dashboard-card"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500 dark:text-slate-400">{{ t('auditExecution.phases.6.shortLabel') }}</p>
          <div class="rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center dark:bg-emerald-500/10">
            <UIcon name="i-lucide-check-circle" class="text-emerald-600 size-5 dark:text-emerald-400" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-2xl sm:text-3xl font-bold text-emerald-600 tracking-tight dark:text-emerald-400">
            {{ completedAuditsCount }}
          </h3>
          <div
            class="flex flex-wrap items-center gap-1 text-md font-semibold text-emerald-500 mt-1"
          >
            <span>{{ t('dashboard.overview.updated') }}</span>
            <span class="text-slate-400 font-normal">{{ t('dashboard.overview.thisMonth') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Audit Statistics Section (Second Row) -->
    <div class="p-4 sm:p-6 space-y-6 dashboard-card">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-bold text-slate-800 dark:text-white">{{ t('dashboard.overview.auditStatistics') }}</h3>
      </div>
      <div class="grid grid-cols-1 gap-4 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Planned Audits -->
        <div
          class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-subcard"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500 dark:text-slate-200">{{ t('dashboard.overview.plannedAudits') }}</span>
            <div class="rounded-lg bg-indigo-50 p-1.5 flex items-center justify-center dark:bg-indigo-500/10">
              <UIcon name="i-lucide-info" class="text-indigo-600 size-4 dark:text-indigo-400" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white">
              {{ auditMainStats.plannedAudit }}
            </h4>
            <p class="text-[11px] mt-1" :class="plannedAuditTrend.color">
              {{ plannedAuditTrend.icon }} {{ plannedAuditTrend.value }}
              <span class="text-slate-400 dark:text-slate-300 font-normal">{{ t('dashboard.overview.fromLastMonth') }}</span>
            </p>
          </div>
        </div>

        <!-- Open Findings -->
        <div
          class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-subcard"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500 dark:text-slate-200">{{ t('dashboard.overview.openFindings') }}</span>
            <div class="rounded-lg bg-amber-50 p-1.5 flex items-center justify-center dark:bg-amber-500/10">
              <UIcon name="i-lucide-alert-triangle" class="text-amber-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white">
              {{ auditMainStats.openFinding }}
            </h4>
            <p class="text-[11px] mt-1" :class="openFindingTrend.color">
              {{ openFindingTrend.icon }} {{ openFindingTrend.value }}
              <span class="text-slate-400 dark:text-slate-300 font-normal">{{ t('dashboard.overview.fromLastMonth') }}</span>
            </p>
          </div>
        </div>

        <!-- Execution Status -->
        <div
          class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-subcard"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500 dark:text-slate-200">{{ t('dashboard.overview.executionStatus') }}</span>
            <div class="rounded-lg bg-sky-50 p-1.5 flex items-center justify-center dark:bg-sky-500/10">
              <UIcon name="i-lucide-play" class="text-sky-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white">
              {{ (auditMainStats.executionStatus * 100).toFixed(0) }}%
            </h4>
            <p class="text-[11px] mt-1" :class="executionStatusTrend.color">
              {{ executionStatusTrend.icon }} {{ executionStatusTrend.value }}
              <span class="text-slate-400 dark:text-slate-300 font-normal">{{ t('dashboard.overview.fromLastMonth') }}</span>
            </p>
          </div>
        </div>

        <!-- ATR Compliance -->
        <div
          class="rounded-xl p-5 flex flex-col justify-between min-h-[120px] dashboard-subcard"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500 dark:text-slate-200">{{ t('dashboard.overview.atrCompliance') }}</span>
            <div class="rounded-lg bg-emerald-50 p-1.5 flex items-center justify-center dark:bg-emerald-500/10">
              <UIcon name="i-lucide-check-square" class="text-emerald-600 size-4 dark:text-emerald-400" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900 dark:text-white">
              {{ (auditMainStats.atrCompliance * 100).toFixed(0) }}%
            </h4>
            <p class="text-[11px] mt-1" :class="atrComplianceTrend.color">
              {{ atrComplianceTrend.icon }} {{ atrComplianceTrend.value }}
              <span class="text-slate-400 dark:text-slate-300 font-normal">{{ t('dashboard.overview.fromLastMonth') }}</span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Dedicated AI Section: KPI Forecasting & Anomaly Detection -->
    <div class="space-y-6">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-200 dark:border-slate-800 pb-4">
        <div>
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-sparkles" class="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
            <h2 class="text-xl font-bold text-slate-900 dark:text-white tracking-tight"> {{ t('dashboard.overview.predictiveTitle') }} </h2>
          </div>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1"> {{ t('dashboard.overview.predictiveDescription') }} </p>
        </div>

        <NuxtLink
          to="/analytics"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 flex items-center gap-1 self-start sm:self-auto"
        >
          <span>{{ t('dashboard.overview.openAnalytics') }}</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <!-- Grid of 2 Main Analytics Cards -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        <!-- CARD 1: KPI FORECASTING (PyTorch LSTM) -->
        <div class="lg:col-span-6 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 shadow-sm rounded-2xl p-4 sm:p-6 flex flex-col justify-between space-y-6 min-w-0">
          <div>
            <!-- Card Header -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-4">
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <h3 class="text-base sm:text-lg font-bold text-slate-900 dark:text-white"> {{ t('dashboard.overview.kpiForecasting') }} </h3>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5"> {{ t('dashboard.overview.forecastPeriod') }} </p>
              </div>

              <NuxtLink to="/analytics/kpi-forecast" class="text-xs font-semibold text-violet-600 hover:underline shrink-0 self-start sm:self-auto"> {{ t('dashboard.overview.details') }} </NuxtLink>
            </div>

            <!-- Line Chart Component -->
            <div class="h-[220px] sm:h-[280px] w-full relative">
              <Line :data="kpiChartData" :options="kpiChartOptions" />
            </div>

            <!-- Forecast Highlights Table / Cards -->
            <div class="mt-6 space-y-3">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2"> {{ t('dashboard.overview.atRiskKpi') }} </h4>

              <div
                v-for="kpi in (timeseriesState.kpiForecasts || []).slice(0, 3)"
                :key="kpi.code"
                class="p-3.5 rounded-xl border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
              >
                <div class="space-y-1 min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="font-mono text-xs font-bold text-violet-600 dark:text-violet-400">{{ kpi.code }}</span>
                    <span class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ kpi.kpiName }}</span>
                    <UBadge :color="getTrendBadgeColor(kpi.trend) as any" variant="subtle" size="sm" class="font-semibold text-[10px]">
                      {{ kpi.trend }}
                    </UBadge>
                  </div>
                  <p class="text-xs text-slate-500 dark:text-slate-400">
                    {{ kpi.recommendedAction }}
                  </p>
                </div>

                <div class="text-right shrink-0 flex sm:flex-col items-center sm:items-end justify-between gap-2 sm:gap-0">
                  <span class="text-xs text-slate-400">{{ t('dashboard.overview.projection') }}</span>
                  <span class="text-sm font-extrabold font-mono text-violet-600 dark:text-violet-400">
                    {{ kpi.forecastedValue }}{{ kpi.unit === "%" ? "%" : "" }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Footer Info -->
          <div class="pt-4 border-t border-slate-100 dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs text-slate-500">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-information-circle" class="w-4 h-4 text-violet-500" />
              {{ t('dashboard.overview.predictionInterval') }}
            </span>
            <span class="font-medium text-slate-700 dark:text-slate-300">{{ t('dashboard.overview.targetHorizon') }}</span>
          </div>
        </div>

        <!-- CARD 2: ANOMALY DETECTION (Isolation Forest) -->
        <div class="lg:col-span-6 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 shadow-sm rounded-2xl p-4 sm:p-6 flex flex-col justify-between space-y-6 min-w-0">
          <div>
            <!-- Card Header -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="text-base sm:text-lg font-bold text-slate-900 dark:text-white"> {{ t('dashboard.overview.transactionAnomalies') }} </h3>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5"> {{ t('dashboard.overview.transactionAnomaliesDescription') }} </p>
              </div>
            </div>

            <!-- Scatter Chart Component -->
            <div class="h-[220px] sm:h-[280px] w-full relative">
              <Scatter :data="anomalyScatterChartData" :options="anomalyScatterOptions" />
            </div>

            <!-- Top Detected Anomaly List -->
            <div class="mt-6 space-y-3">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2"> {{ t('dashboard.overview.criticalAnomalies') }} </h4>

              <div
                v-for="anm in filteredAnomalies.slice(0, 3)"
                :key="anm.id"
                class="p-3.5 rounded-xl border border-rose-100 dark:border-rose-900/40 bg-rose-50/40 dark:bg-rose-950/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
              >
                <div class="space-y-1 min-w-0">
                  <div class="flex flex-wrap items-center gap-2">
                    <UBadge :color="getSeverityColor(anm.severity) as any" variant="solid" size="sm" class="font-bold text-[10px]">
                      {{ anm.severity }}
                    </UBadge>
                    <span class="font-mono text-xs font-bold text-slate-900 dark:text-slate-100">{{ anm.id }}</span>
                    <span class="text-xs font-bold text-slate-600 dark:text-slate-400">· {{ anm.entity }}</span>
                  </div>
                  <p class="text-xs text-slate-600 dark:text-slate-300 font-medium">
                    {{ anm.description }}
                  </p>
                </div>

                <div class="text-right shrink-0 flex sm:flex-col items-center sm:items-end justify-between gap-2 sm:gap-0">
                  <span class="text-xs font-mono text-rose-600 dark:text-rose-400 font-bold"> {{ t('dashboard.overview.score') }} {{ anm.anomalyScore }}
                  </span>
                  <span v-if="anm.amount" class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    Rp {{ (anm.amount / 1000000).toLocaleString() }}M
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Footer Info -->
          <div class="pt-4 border-t border-slate-100 dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs text-slate-500">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-shield-exclamation" class="w-4 h-4 text-rose-500" /> {{ t('dashboard.overview.contaminationRate') }} {{ ((isolationState.summary?.contaminationRate || 0) * 100).toFixed(1) }}%
            </span>
            <NuxtLink to="/analytics" class="font-semibold text-rose-600 dark:text-rose-400 hover:underline"> {{ t('dashboard.overview.manageAll') }} {{ (isolationState.anomalies || []).length }} {{ t('dashboard.overview.anomalies') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>

    <!-- Internal Control Effectiveness Section (COSO 2013) -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

      <!-- Card 1: Yearly Internal Control Effectiveness (2 Cols) -->
      <div class="lg:col-span-2 p-4 sm:p-6 flex flex-col justify-between dashboard-card">
        <div>
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div class="flex items-center gap-2 min-w-0">
              <span class="text-sm font-semibold tracking-wider text-slate-400 dark:text-slate-400">{{ t('rcm.cards.effectiveness.badge') }}</span>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium text-slate-500 dark:text-slate-300 bg-slate-100 dark:bg-slate-800 px-2.5 py-1 rounded-md border border-transparent dark:border-slate-700"> {{ t('riskFactors.scoring.yearLabel') }} {{ rcmStore.selectedYear }} | {{ rcmStore.selectedDepartment }}
              </span>
              <UButton to="/risk-profile/risk-control-matrix" variant="outline" color="neutral" size="sm" class="font-medium">
                {{ t('dashboard.overview.riskControlMatrix') }}
              </UButton>
            </div>
          </div>
          <h2 class="text-xl font-bold text-slate-900 dark:text-white mt-1">{{ t('dashboard.overview.controlEffectiveness') }}</h2>
        </div>

        <div class="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4 bg-slate-50 dark:bg-slate-800/60 rounded-xl p-4 border border-slate-100 dark:border-slate-700/60">
          <!-- Big Score % -->
          <div class="flex flex-col justify-center border-b md:border-b-0 md:border-r border-slate-200 dark:border-slate-700 pb-3 md:pb-0 md:pr-4">
            <span class="text-sm text-slate-500 dark:text-slate-400 font-medium">{{ t('rcm.cards.effectiveness.realScore') }}</span>
            <div class="flex items-baseline gap-2 mt-1">
              <span class="text-2xl font-extrabold tracking-tight text-slate-900 dark:text-white">
                {{ rcmStore.internalControlEffectiveness }}%
              </span>
            </div>
            <div class="mt-2">
              <span :class="getRatingBadgeClass(rcmStore.effectivenessRating.rating)">
                {{ rcmStore.effectivenessRating.rating }}
              </span>
            </div>
          </div>

          <!-- Synchronized Risk Counts & Interpretation -->
          <div class="md:col-span-2 flex flex-col justify-center space-y-2 min-w-0">
            <div class="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 text-sm">
              <span class="text-slate-600 dark:text-slate-300 font-medium">{{ t('rcm.cards.effectiveness.inherentRiskLabel') }}</span>
              <span class="font-bold text-slate-900 dark:text-slate-100 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 whitespace-nowrap">
                {{ rcmStore.totalInherentRisk }} {{ t('settings.faq.categories.Risk') }} </span>
            </div>
            <div class="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 text-sm">
              <span class="text-slate-600 dark:text-slate-300 font-medium">{{ t('rcm.cards.effectiveness.residualRiskLabel') }}</span>
              <span class="font-bold text-red-600 dark:text-red-400 bg-white dark:bg-slate-900 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 whitespace-nowrap">
                {{ rcmStore.totalResidualRisk }} {{ t('settings.faq.categories.Risk') }} </span>
            </div>
            <UAlert
              :color="(rcmStore.effectivenessRating.alertColor as any) || 'info'"
              variant="subtle"
              icon="i-lucide-info"
              :title="'Interpretasi Hasil COSO (' + rcmStore.effectivenessRating.rating + '):'"
              :description="rcmStore.effectivenessRating.interpretation"
              class="rounded-xl shadow-xs border-none! bg-transparent! border-transparent!"
            />
          </div>
        </div>

        <div class="mt-3 text-sm text-slate-400 dark:text-slate-400 flex items-center gap-3 sm:gap-4">
          <UIcon name="i-lucide-check-circle-2" class="size-8 sm:size-10 shrink-0 text-emerald-500" />
          <span>{{ t('rcm.cards.effectiveness.syncNote') }}</span>
        </div>
      </div>

      <!-- Card 2: COSO 2013 5 Dimensions Summary (1 Col) -->
      <div class="p-4 sm:p-6 flex flex-col justify-between dashboard-card">
        <div>
          <div class="flex items-center justify-between gap-2">
            <h3 class="text-base font-bold text-slate-900 dark:text-white">{{ t('rcm.cards.coso.title') }}</h3>
            <span class="text-md font-bold text-primary-700 dark:text-primary-300 bg-primary-50 dark:bg-primary-950/60 px-2.5 py-1 rounded-full border border-primary-100 dark:border-primary-800 whitespace-nowrap shrink-0">
              {{ rcmStore.cosoAverages.totalWeighted }}%
            </span>
          </div>

          <div class="mt-4 space-y-3">
            <div v-for="dim in cosoDimensions" :key="dim.key" class="space-y-1">
              <div class="flex justify-between text-md">
                <span class="font-medium text-slate-700 dark:text-slate-300">{{ dim.shortLabel }}</span>
                <span class="font-bold text-slate-900 dark:text-white">{{ getDimAverage(dim.key) }}%</span>
              </div>
              <div class="w-full bg-slate-100 dark:bg-slate-800 h-2 rounded-full overflow-hidden">
                <div
                  class="bg-primary-600 dark:bg-primary-500 h-full rounded-full transition-all duration-300"
                  :style="{ width: `${getDimAverage(dim.key)}%` }"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-md text-slate-500 dark:text-slate-400">
          <span>{{ t('rcm.cards.coso.totalControls') }}</span>
          <span class="font-bold text-slate-900 dark:text-white">{{ rcmStore.filteredRCMList.length }} {{ t('dashboard.overview.item') }}</span>
        </div>
      </div>
    </div>

    <!-- Charts Section (Third Row) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Inherent vs Residual Risk by Department Chart -->
      <div
        class="lg:col-span-8 min-w-0 p-4 sm:p-6 flex flex-col justify-between dashboard-card"
      >
        <div>
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-base sm:text-lg font-bold text-slate-800 dark:text-white"> {{ t('dashboard.overview.departmentRisk') }} </h3>
          </div>
          <div class="w-full" :style="{ height: `${barChartHeight}px` }">
            <BarChart
              :data="mainRiskData"
              :categories="riskCategories"
              :x-formatter="xFormatter"
              :y-axis="['inherentRisk', 'residualRisk']"
              :radius="6"
              :height="barChartHeight"
              :hide-legend="true"
              :x-axis-config="{
                tickTextColor: '#64748b',
                tickTextFontSize: '11px',
              }"
              :y-axis-config="{
                tickTextColor: '#64748b',
                tickTextFontSize: '11px',
              }"
              :padding="{ top: 10, right: 10, bottom: 10, left: 10 }"
            />
          </div>
        </div>
        <!-- Custom Legend -->
        <div class="flex flex-wrap justify-center gap-x-6 gap-y-2 mt-4 border-t border-slate-50 pt-4 dark:border-slate-800">
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full bg-[#ff5c02]"></span>
            <span class="text-md font-semibold text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.inherentRisk') }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full bg-[#4d00ff]"></span>
            <span class="text-md font-semibold text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.residualRisk') }}</span>
          </div>
        </div>
      </div>

      <!-- Action Taken Report Donut Chart -->
      <div
        class="lg:col-span-4 min-w-0 p-4 sm:p-6 flex flex-col justify-between dashboard-card"
      >
        <div>
          <div class="flex items-center justify-between gap-2 mb-6">
            <h3 class="text-base sm:text-lg font-bold text-slate-800 dark:text-white">{{ t('actionTakenReport.title') }}</h3>
            <UButton to="/action-taken-report" variant="ghost" color="neutral" size="sm"
              >{{ t('dashboard.overview.details') }}</UButton
            >
          </div>
          <div class="w-full relative flex items-center justify-center" :style="{ height: `${donutChartHeight}px` }">
            <p
              v-if="atrStore.summaryLoading && atrStore.stats.total === 0"
              class="text-sm text-slate-500 dark:text-slate-400"
            >
              {{ t("common.loading") }}
            </p>
            <div
              v-else-if="atrStore.summaryError"
              class="flex flex-col items-center gap-1 text-center px-4"
            >
              <UIcon name="i-heroicons-exclamation-triangle" class="size-6 text-error-500" />
              <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">
                {{ t("actionTakenReport.dashboard.error") }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ atrStore.summaryError }}</p>
            </div>
            <div
              v-else-if="atrStore.stats.total === 0"
              class="flex flex-col items-center gap-1 text-center px-4"
            >
              <UIcon name="i-heroicons-circle-stack-20-solid" class="size-6 text-slate-400" />
              <p class="text-sm text-slate-500 dark:text-slate-400">
                {{ t("actionTakenReport.dashboard.empty") }}
              </p>
            </div>
            <DonutChart
              v-else
              :data="atrDonutData.map((item) => item.value)"
              :categories="atrCategories"
              :radius="8"
              :arc-width="26"
              :height="donutChartHeight"
              :hide-legend="true"
            >
              <div class="text-center flex flex-col items-center justify-center">
                <span class="text-xl font-bold text-slate-900 leading-tight dark:text-white"
                  >{{ atrStore.stats.donePercent }}%</span
                >
                <span class="text-[10px] text-slate-500 font-medium dark:text-slate-400">{{
                  t("actionTakenReport.dashboard.centerLabel")
                }}</span>
              </div>
            </DonutChart>
          </div>
        </div>
        <!-- Custom Legend -->
        <div
          v-if="atrStore.stats.total > 0"
          class="flex flex-wrap justify-center gap-x-4 gap-y-2 mt-4 px-2 border-t border-slate-50 pt-4 dark:border-slate-800"
        >
          <div v-for="slice in atrDonutData" :key="slice.key" class="flex items-center gap-1.5">
            <span
              class="w-2.5 h-2.5 rounded-full"
              :style="{ backgroundColor: atrSliceColors[slice.key] }"
            ></span>
            <span class="text-[11px] font-semibold text-slate-700 dark:text-white">{{
              t("actionTakenReport.dashboard.legend", { label: slice.name, percent: slice.value })
            }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Risk Heat Map & Registered Risk (Fourth Row) -->
    <div class="p-4 sm:p-6 space-y-6 dashboard-card">
      <div class="flex items-center justify-between gap-2">
        <h2 class="text-lg sm:text-xl font-bold text-slate-900 tracking-tight dark:text-white">{{ t('dashboard.overview.riskProfiles') }}</h2>
        <UButton to="/risk-profile" variant="ghost" color="neutral" size="sm"
          >{{ t('dashboard.planningExecution.configure') }}</UButton
        >
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 lg:gap-8 items-stretch">
        <!-- Heat Map Grid on the left (stretches to the Registered Risks card height) -->
        <div
          class="lg:col-span-5 min-w-0 h-full flex flex-col items-center rounded-2xl p-4 sm:p-6 dashboard-subcard"
        >
          <h4 class="text-sm font-semibold text-slate-700 mb-6 w-full text-left dark:text-slate-200"> {{ t('dashboard.overview.riskHeatMap') }} </h4>
          <div class="flex-1 flex flex-row gap-2 sm:gap-4 items-center w-full justify-center">
            <!-- Y-axis label (narrow box: the rotated text would otherwise reserve its unrotated width) -->
            <div class="flex flex-col items-center justify-center shrink-0 w-5">
              <span
                class="text-md font-semibold text-slate-500 uppercase tracking-wider origin-center -rotate-90 whitespace-nowrap dark:text-slate-400"
              > {{ t('dashboard.overview.probability') }} </span>
            </div>
            <!-- Heat map grid -->
            <div class="flex flex-col gap-2">
              <div class="grid grid-cols-5 gap-1 sm:gap-1.5">
                <template v-for="y in 5" :key="y">
                  <template v-for="x in 5" :key="`${x}-${y}`">
                    <div
                      class="w-8 h-8 sm:w-10 sm:h-10 flex items-center justify-center text-md font-semibold rounded shadow-sm hover:scale-105 transition-transform cursor-pointer"
                      :class="getHeatMapCellColor(x, y)"
                      :title="`Impact: ${x}, Probability: ${6 - y}, Risk: ${getRiskLevel(
                        x,
                        6 - y
                      )}`"
                    ></div>
                  </template>
                </template>
              </div>
              <div class="flex justify-center mt-1">
                <span
                  class="text-md font-semibold text-slate-500 uppercase tracking-wider dark:text-slate-400"
                  >{{ t('riskProfile.addModal.impact') }}</span
                >
              </div>
            </div>
          </div>
          <!-- Legend -->
          <div class="flex flex-wrap justify-center gap-x-3 gap-y-1.5 mt-6 max-w-[320px]">
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-green-400"></div>
              <span class="text-[11px] font-medium text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.low') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-green-600"></div>
              <span class="text-[11px] font-medium text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.lowModerate') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-yellow-500"></div>
              <span class="text-[11px] font-medium text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.moderate') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-orange-500"></div>
              <span class="text-[11px] font-medium text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.moderateHigh') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-red-500"></div>
              <span class="text-[11px] font-medium text-slate-600 dark:text-slate-300">{{ t('dashboard.overview.high') }}</span>
            </div>
          </div>
        </div>

        <!-- Registered Risks Table on the right -->
        <div class="lg:col-span-7 min-w-0 rounded-2xl p-4 sm:p-6 dashboard-subcard">
          <h4 class="text-sm font-semibold text-slate-700 mb-4 dark:text-slate-200">{{ t('dashboard.overview.registeredRisks') }}</h4>
          <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
            <TableEntities
              :data="registeredRiskHeatMap"
              :columns="registeredRiskColumns"
              :empty-state="{
                icon: 'i-heroicons-circle-stack-20-solid',
                label: 'Belum ada data yang dimasukkan ke Risk Heat Map.',
              }"
              class="w-full"
            />
          </UCard>
        </div>
      </div>
    </div>

    <!-- Fifth Row: Audit Planning & Execution (merged) -->
    <div
      class="min-w-0 p-4 sm:p-6 space-y-6 dashboard-card"
    >
      <h2 class="text-lg sm:text-xl font-bold text-slate-900 tracking-tight dark:text-white">
        {{ t("dashboard.planningExecution.title") }}
      </h2>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 lg:gap-0 lg:divide-x lg:divide-slate-100 dark:lg:divide-slate-800">
        <!-- Audit Planning Coverage -->
        <div class="min-w-0 flex flex-col justify-between lg:pr-8">
          <div>
            <div class="flex items-center justify-between gap-2 mb-6">
              <h3 class="text-base sm:text-lg font-bold text-slate-800 dark:text-white">{{ t("dashboard.planningExecution.coverageTitle") }}</h3>
              <UButton to="/annual-audit" variant="ghost" color="neutral" size="sm"
                >{{ t("dashboard.planningExecution.configure") }}</UButton
              >
            </div>
            <div class="space-y-4">
              <div class="flex flex-wrap justify-between items-center gap-x-3 gap-y-1">
                <span class="text-md font-semibold text-slate-500 dark:text-slate-400">{{ t('dashboard.overview.overallProgress') }}</span>
                <h5 class="text-base font-bold text-indigo-600 whitespace-nowrap dark:text-indigo-400">
                  {{ progressModel }}{{ t('dashboard.overview.completedPercent') }} </h5>
              </div>
              <UProgress
                :model-value="progressModel"
                color="secondary"
                class="h-2 rounded"
              />
            </div>
          </div>
          <div
            class="grid grid-cols-3 gap-2 border-t border-slate-50 pt-6 mt-6 text-center dark:border-slate-800"
          >
            <div>
              <p class="text-[11px] text-slate-400 font-semibold uppercase">{{ t('strategicPlan.status.planned') }}</p>
              <h4 class="text-lg font-bold text-slate-800 mt-1 dark:text-white">
                {{ auditCoverage.plannedAudits }}
              </h4>
            </div>
            <div>
              <p class="text-[11px] text-slate-400 font-semibold uppercase">{{ t('auditExecution.phases.6.shortLabel') }}</p>
              <h4 class="text-lg font-bold text-slate-800 mt-1 dark:text-white">
                {{ auditCoverage.completedAudits }}
              </h4>
            </div>
            <div>
              <p class="text-[11px] text-slate-400 font-semibold uppercase">{{ t('dashboard.overview.remaining') }}</p>
              <h4 class="text-lg font-bold text-slate-800 mt-1 dark:text-white">
                {{ auditCoverage.remainingAudits }}
              </h4>
            </div>
          </div>
        </div>

        <!-- Audit Execution Status -->
        <div class="min-w-0 space-y-6 border-t border-slate-100 pt-6 lg:border-t-0 lg:pt-0 lg:pl-8 dark:border-slate-800">
          <div class="flex items-center justify-between gap-2 mb-6">
            <h3 class="text-base sm:text-lg font-bold text-slate-800 dark:text-white">
              {{ t("dashboard.planningExecution.executionTitle") }}
            </h3>
            <UBadge color="primary" :label="executionPeriodLabel" variant="soft"></UBadge>
          </div>
          <p
            v-if="auditExecutionStore.loading && dashboardExecutionStatus.length === 0"
            class="text-sm text-slate-500 dark:text-slate-400 py-6 text-center"
          >
            {{ t("common.loading") }}
          </p>
          <div
            v-else-if="auditExecutionStore.error"
            class="flex flex-col items-center gap-1 py-6 text-center"
          >
            <UIcon name="i-heroicons-exclamation-triangle" class="size-6 text-error-500" />
            <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">
              {{ t("dashboard.planningExecution.executionError") }}
            </p>
            <p class="text-xs text-slate-500 dark:text-slate-400">{{ auditExecutionStore.error }}</p>
          </div>
          <div
            v-else-if="dashboardExecutionStatus.length === 0"
            class="flex flex-col items-center gap-1 py-6 text-center"
          >
            <UIcon name="i-heroicons-circle-stack-20-solid" class="size-6 text-slate-400" />
            <p class="text-sm text-slate-500 dark:text-slate-400">
              {{ t("dashboard.planningExecution.executionEmpty") }}
            </p>
          </div>
          <div v-else class="space-y-5 max-h-80 overflow-y-auto pr-1">
            <div
              v-for="(item, index) in dashboardExecutionStatus"
              :key="index"
              class="space-y-2"
            >
              <div class="flex justify-between items-center gap-3">
                <h3 class="text-sm font-semibold text-slate-700 min-w-0 truncate dark:text-slate-200">
                  {{ item.name }}
                </h3>
                <p class="text-md font-bold text-indigo-600 shrink-0 dark:text-indigo-400">{{ item.percentage }}%</p>
              </div>
              <UProgress
                :model-value="item.percentage"
                color="secondary"
                class="h-1.5 rounded"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Sixth Row: Action Taken Reports Table (full width) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Action Taken Reports Table -->
      <div
        class="lg:col-span-12 min-w-0 p-4 sm:p-6 dashboard-card"
      >
        <div class="flex flex-wrap items-center justify-between gap-2 mb-4">
          <h3 class="text-base sm:text-lg font-bold text-slate-800 dark:text-white">{{ t('dashboard.overview.actionTakenReports') }}</h3>
          <UButton to="/action-taken-report" variant="ghost" color="neutral" size="sm"
            >{{ t('dashboard.overview.viewFullReport') }}</UButton
          >
        </div>
        <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
          <TableEntities
            :data="atrTableData"
            :columns="tableColumns"
            :loading="atrStore.summaryLoading"
            :empty-state="{
              icon: 'i-heroicons-circle-stack-20-solid',
              label: t('actionTakenReport.dashboard.empty'),
            }"
            class="w-full"
          />
        </UCard>
      </div>
    </div>

    <!-- Seventh Row: Recent Finding Issues (full width) -->
    <div class="grid grid-cols-1 gap-6">
      <!-- Recent Finding Issues -->
      <div class="min-w-0 p-4 sm:p-6 dashboard-card">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-lg sm:text-xl font-bold text-slate-900 tracking-tight dark:text-white">
            {{ t("dashboard.recentFindings.title") }}
          </h2>
        </div>
        <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
          <TableEntities
            :data="recentFindingsData"
            :columns="auditTableColumns"
            :loading="auditResultStore.recentFindingsLoading"
            :empty-state="recentFindingsEmptyState"
            class="w-full"
          />
        </UCard>
      </div>
    </div>

    <!-- Eighth Row: Recent Risk Profiles & Upcoming Audits -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Risk Profiles -->
      <div class="min-w-0 p-4 sm:p-6 space-y-6 dashboard-card">
        <div class="flex items-center justify-between gap-2">
          <div class="min-w-0">
            <h3 class="text-lg font-bold text-slate-800 dark:text-white">{{ t('dashboard.overview.recentRiskProfiles') }}</h3>
            <p class="text-md text-slate-400 mt-1">{{ t('dashboard.overview.latestRiskAssessments') }}</p>
          </div>
          <UTooltip :text="t('dashboard.overview.viewAllRiskProfiles')">
            <UButton
              icon="i-lucide-chevron-right"
              variant="ghost"
              color="primary"
              to="/risk-profile"
              size="sm"
            />
          </UTooltip>
        </div>
        <div class="space-y-3">
          <div
            v-for="(risk, index) in riskProfileStore.risks.slice(0, 4)"
            :key="risk.id"
            class="flex items-center justify-between gap-3 p-3.5 rounded-xl dashboard-subcard"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="rounded-full bg-indigo-50 w-8 h-8 shrink-0 flex items-center justify-center dark:bg-indigo-500/10"
              >
                <span class="text-md font-bold text-indigo-700 dark:text-indigo-300">{{ index + 1 }}</span>
              </div>
              <div class="min-w-0">
                <p class="text-sm font-bold text-slate-800 truncate dark:text-white">
                  {{ risk.name }}
                </p>
                <p class="text-md text-slate-400 mt-0.5 truncate">{{ risk.category }}</p>
              </div>
            </div>
            <UBadge
              :color="risk.impact * risk.likelihood > 15 ? 'error' : 'warning'"
              variant="soft"
              size="sm"
              class="font-semibold shrink-0"
            >
              {{ risk.impact * risk.likelihood > 15 ? "High" : "Medium" }}
            </UBadge>
          </div>
        </div>
      </div>

      <!-- Upcoming Audits -->
      <div class="min-w-0 p-4 sm:p-6 space-y-6 dashboard-card">
        <div class="flex items-center justify-between gap-2">
          <div class="min-w-0">
            <h3 class="text-lg font-bold text-slate-800 dark:text-white">{{ t('dashboard.overview.upcomingAudits') }}</h3>
            <p class="text-md text-slate-400 mt-1">{{ t('dashboard.overview.scheduledActivities') }}</p>
          </div>
          <UTooltip :text="t('dashboard.overview.viewAllAnnualAudits')">
            <UButton
              icon="i-lucide-chevron-right"
              variant="ghost"
              color="primary"
              to="/annual-audit"
              size="sm"
            />
          </UTooltip>
        </div>
        <div class="space-y-3">
          <div
            v-for="plan in annualPlanStore.plans.slice(0, 4)"
            :key="plan.id"
            class="flex items-center justify-between gap-3 p-3.5 rounded-xl dashboard-subcard"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div class="rounded-lg bg-sky-50 w-8 h-8 shrink-0 flex items-center justify-center dark:bg-sky-500/10">
                <UIcon name="i-lucide-calendar" class="text-sky-600 size-4" />
              </div>
              <div class="min-w-0">
                <p class="text-sm font-bold text-slate-800 truncate dark:text-white">{{ plan.code }}</p>
                <p class="text-md text-slate-400 mt-0.5 truncate">{{ t('dashboard.overview.status') }} {{ plan.status }}</p>
              </div>
            </div>
            <UBadge color="info" variant="soft" size="sm" class="font-semibold shrink-0"
              >{{ t('riskMitigation.monitoring.scheduled') }}</UBadge
            >
          </div>
          <div
            v-if="annualPlanStore.plans.length === 0"
            class="text-center py-6 text-slate-400 text-sm"
          > {{ t('dashboard.overview.noUpcomingAudits') }} </div>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import AuditorDashboard from "~/components/dashboard/AuditorDashboard.vue";
import { useRbac } from "~/composables/useRbac";
import { Line, Scatter } from "vue-chartjs";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from "chart.js";
import { useAiAnalytics } from "~/composables/useAiAnalytics";
import {
  buildAnomalyScatterPoints,
  createAnomalyTypeConfigs,
  getAnomalyTypeConfig,
  listAnomalyTypes,
  resolveAnomalyType,
} from "~/composables/useAnomalyScatter";
import { useI18n } from "~/composables/useI18n";
import { useRiskProfileStore, riskLevelConfig } from "~/stores/risk-profile";
import { getControlEffectivenessColorClass, getRiskLevelColorClass } from "~/utils/riskLevelBadge";
import { useAnnualPlanStore } from "~/stores/annual-audit";
import { AnnualAuditPlanStatus } from "~/types/audit";
import { useActionTakenReportStore } from "~/stores/action-taken-report";
import { useAuditExecutionStore } from "~/stores/audit-execution";
import { getProgressValue } from "~/utils/executionSources";
import { useAuditResultReportStore } from "~/stores/audit-result-report";
import { useAuthStore } from "~/stores/auth";
import { useRCMStore, cosoDimensions } from "~/stores/rcm";
import { RiskLevel } from "~/types/risk";
import { ATR_SLICE_HEX, atrStatusI18nKey, formatAtrDate, type AtrSliceKey } from "~/utils/actionTakenReport";
import { UBadge } from "#components";
import OverflowTooltip from "~/components/shared/OverflowTooltip.vue";

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

definePageMeta({
  middleware: "auth",
});

const { t, locale } = useI18n();

const riskProfileStore = useRiskProfileStore();
const annualPlanStore = useAnnualPlanStore();
const atrStore = useActionTakenReportStore();
const auditExecutionStore = useAuditExecutionStore();
const auditResultStore = useAuditResultReportStore();
const authStore = useAuthStore();
const rcmStore = useRCMStore();
const { isAuditorView, setRoleView, isAdmin, canManageAudits } = useRbac();
const canToggleRoleView = computed(() => isAdmin.value || canManageAudits.value);

// ─── Internal Control Effectiveness (COSO 2013) ─────────────
const getDimAverage = (dimKey: string) => {
  const map: Record<string, number> = {
    design_effectiveness: rcmStore.cosoAverages.design,
    operating_effectiveness: rcmStore.cosoAverages.operating,
    coverage_completeness: rcmStore.cosoAverages.coverage,
    timeliness: rcmStore.cosoAverages.timeliness,
    automation_monitoring: rcmStore.cosoAverages.automation,
  };
  return map[dimKey] || 0;
};

const getRatingBadgeClass = (ratingLabel: string) =>
  `font-bold px-2.5 py-1 rounded-md text-sm shadow-md whitespace-nowrap inline-block ${getControlEffectivenessColorClass(ratingLabel)}`;

// ─── Responsive breakpoints & chart sizing ──────────────────
// SSR renders the desktop size; the media query re-evaluates on hydration.
const isMobile = useMediaQuery("(max-width: 639px)");
const barChartHeight = computed(() => (isMobile.value ? 280 : 350));
const donutChartHeight = computed(() => (isMobile.value ? 180 : 220));

// AI & Analytics Composables
// The KPI forecast chart shares its state with /analytics/ai/kpi-forecast so
// both charts always render the same series.
const { timeseriesState, isolationState, fetchAiAnalytics } = useAiAnalytics();

// The anomaly card mirrors the "Transaction" tab of /analytics/ai/anomaly-detection:
// same state, same category, same point builder.
const anomalyTypeConfigs = computed(() => createAnomalyTypeConfigs(t));

const availableAnomalyTypes = computed(() =>
  listAnomalyTypes(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    Object.keys(anomalyTypeConfigs.value)
  )
);

const dashboardAnomalyType = computed(() =>
  resolveAnomalyType(availableAnomalyTypes.value, "Transaction")
);

const anomalyConfig = computed(() =>
  getAnomalyTypeConfig(t, dashboardAnomalyType.value)
);

const filteredAnomalies = computed(() =>
  (isolationState.value.anomalies || []).filter(
    (a: any) => a.type === dashboardAnomalyType.value
  )
);

onMounted(() => {
  // If viewing Auditor role dashboard, operational data is handled by AuditorDashboard component
  if (isAuditorView.value) return;

  fetchAiAnalytics();
  // Load the real annual audit plans (all pages) for the Audit Planning Coverage card.
  annualPlanStore.fetchPlans();
  // Recent Finding Issues card (ARR + KKA + fieldwork, merged by the backend).
  auditResultStore.fetchRecentFindings(5);
  // RCM-based figures (control effectiveness etc.) come only from the risk-service API.
  rcmStore.fetchRCMList();
  // ATR donut, compliance, Open Findings and the ATR table: every ATR the user can see.
  atrStore.fetchSummary();
  // Audit Execution Status card and the Execution Status tile (same list as /audit-execution-status).
  auditExecutionStore.fetchAuditExecutions();
});

// ─── KPI Forecasting Line Chart Config ─────────────────────
const historicalKPI = computed<any[]>(
  () => timeseriesState.value.historicalKPI || []
);

const kpiChartData = computed(() => ({
  labels: historicalKPI.value.map((p) => p.period || ""),
  datasets: [
    {
      label: t("analytics.timeseries.labelActualKPI"),
      data: historicalKPI.value.map((p) => p.actual ?? null),
      borderColor: "#4f46e5",
      backgroundColor: "rgba(79, 70, 229, 0.15)",
      tension: 0.4,
      borderWidth: 3,
      pointRadius: 5,
      pointHoverRadius: 7,
      pointBackgroundColor: "#4f46e5",
      pointBorderColor: "#ffffff",
      pointBorderWidth: 2,
      spanGaps: false,
    },
    {
      label: t("analytics.timeseries.labelForecast"),
      data: historicalKPI.value.map((p) => p.forecast ?? null),
      borderColor: "#8b5cf6",
      backgroundColor: "rgba(139, 92, 246, 0.15)",
      borderDash: [6, 4],
      tension: 0.4,
      borderWidth: 3,
      pointRadius: 5,
      pointHoverRadius: 7,
      pointStyle: "rectRot",
      pointBackgroundColor: "#8b5cf6",
      pointBorderColor: "#ffffff",
      pointBorderWidth: 2,
      spanGaps: false,
    },
    {
      label: t("analytics.timeseries.labelUpperBound"),
      data: historicalKPI.value.map((p) => p.upperBound ?? null),
      borderColor: "transparent",
      backgroundColor: "rgba(139, 92, 246, 0.08)",
      fill: "+1",
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
    {
      label: t("analytics.timeseries.labelLowerBound"),
      data: historicalKPI.value.map((p) => p.lowerBound ?? null),
      borderColor: "transparent",
      backgroundColor: "rgba(139, 92, 246, 0.08)",
      fill: "-1",
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
  ],
}));

const kpiChartOptions = computed(() => {
  const boundLabels = [
    t("analytics.timeseries.labelUpperBound"),
    t("analytics.timeseries.labelLowerBound"),
  ];

  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: "top" as const,
        labels: {
          usePointStyle: true,
          font: { family: "Inter, sans-serif", size: 12, weight: "bold" as const },
          filter: (item: any) => !boundLabels.includes(item.text),
        },
      },
      tooltip: {
        backgroundColor: "rgba(15, 23, 42, 0.9)",
        titleFont: { size: 13, weight: "bold" as const },
        bodyFont: { size: 12 },
        padding: 12,
        cornerRadius: 10,
        callbacks: {
          label: (ctx: any) => {
            if (boundLabels.includes(ctx.dataset.label)) return "";
            return `${ctx.dataset.label}: ${ctx.parsed.y?.toFixed(1)}%`;
          },
        },
      },
    },
    scales: {
      x: {
        grid: { display: false },
        ticks: { font: { family: "Inter, sans-serif", size: 11 } },
      },
      // No fixed min/max: the forecast range comes from the model, so clamping
      // the axis would clip the confidence band.
      y: {
        grid: { color: "rgba(226, 232, 240, 0.6)" },
        ticks: {
          font: { family: "Inter, sans-serif", size: 11 },
          callback: (v: any) => `${v}%`,
        },
        title: {
          display: true,
          text: t("analytics.timeseries.axisKPIScore"),
          font: { size: 12, weight: "bold" as const },
        },
      },
    },
  };
});

// ─── Anomaly Detection Scatter Chart Config ────────────

const anomalyScatterChartData = computed(() => {
  const typeFilter = dashboardAnomalyType.value;
  const { normalPoints, anomalyPoints } = buildAnomalyScatterPoints(
    isolationState.value.anomalies,
    isolationState.value.scatterData,
    typeFilter
  );
  const colors = anomalyConfig.value.colors;

  return {
    datasets: [
      {
        label: t("analytics.isolation.labelNormalData", { type: typeFilter }),
        data: normalPoints,
        backgroundColor: "rgba(59,130,246,0.3)",
        borderColor: "rgba(59,130,246,0.5)",
        pointRadius: 4,
        pointHoverRadius: 6,
      },
      {
        label: t("analytics.isolation.labelAnomalies", { type: typeFilter }),
        data: anomalyPoints,
        backgroundColor: colors.bg,
        borderColor: colors.border,
        pointRadius: 8,
        pointHoverRadius: 10,
        pointStyle: colors.style,
      },
    ],
  };
});

const anomalyScatterOptions = computed(() => {
  const config = anomalyConfig.value;

  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: "top" as const,
        labels: {
          usePointStyle: true,
          font: { family: "Inter, sans-serif", size: 11, weight: "bold" as const },
        },
      },
      tooltip: {
        backgroundColor: "rgba(15, 23, 42, 0.9)",
        titleFont: { size: 13, weight: "bold" as const },
        bodyFont: { size: 12 },
        padding: 12,
        cornerRadius: 10,
        callbacks: {
          label: (ctx: any) => {
            const formattedX = config.formatX(ctx.parsed.x);
            const metricName = config.xAxisTitle.split(" (")[0];
            return `${metricName}: ${formattedX}, ${t("analytics.isolation.axisFrequency")}: ${ctx.parsed.y}`;
          },
        },
      },
    },
    scales: {
      x: {
        title: {
          display: true,
          text: config.xAxisTitle,
          font: { size: 11, weight: "bold" as const },
        },
        grid: { color: "rgba(226, 232, 240, 0.6)" },
        ticks: { font: { size: 11 } },
      },
      y: {
        title: {
          display: true,
          text: t("analytics.isolation.axisFrequency"),
          font: { size: 11, weight: "bold" as const },
        },
        grid: { display: false },
        ticks: { font: { size: 11 } },
      },
    },
  };
});

const getSeverityColor = (sev: string) => {
  switch (sev) {
    case "Critical":
      return "error";
    case "High":
      return "warning";
    case "Medium":
      return "info";
    default:
      return "neutral";
  }
};

const getTrendBadgeColor = (trend: string) => {
  switch (trend) {
    case "Improving":
      return "success";
    case "Declining":
    case "Deteriorating":
      return "error";
    default:
      return "warning";
  }
};

// Sync state and action
const isSyncing = ref(false);
const handleSync = async () => {
  isSyncing.value = true;
  try {
    await Promise.allSettled([
      riskProfileStore.fetchRisks(),
      annualPlanStore.fetchPlans(),
      auditExecutionStore.fetchAuditExecutions(),
      atrStore.fetchSummary()
    ]);
  } finally {
    isSyncing.value = false;
  }
};

// Stats for Header
const totalRisks = computed(() => riskProfileStore.risks.length);
const highRisks = computed(
  () =>
    riskProfileStore.risks.filter(
      (r) => riskProfileStore.getRiskLevel(r.likelihood, r.impact) === RiskLevel.HIGH
    ).length
);
const auditPlansCount = computed(() => annualPlanStore.plans.length);
const completedAuditsCount = computed(
  () => annualPlanStore.plans.filter((p) => p.status === AnnualAuditPlanStatus.APPROVED).length
);

// Audit Statistics (Section 2)
const plannedAuditCount = computed(
  () => annualPlanStore.plans.filter((p) => p.status !== AnnualAuditPlanStatus.APPROVED).length
);
// Open findings = ATR items not COMPLETED and not CANCELLED (see utils/actionTakenReport).
const openFindingsCount = computed(() => atrStore.stats.counts.open);
const executionStatusPercent = computed(() => {
  const executions = auditExecutionStore.auditExecutions;
  if (executions.length === 0) return 0;
  return executions.reduce((sum, e) => sum + getProgressValue(e.progress), 0) / (executions.length * 100);
});
// ATR compliance = completed / all items (same basis as the donut "Completed" slice).
const atrCompliancePercent = computed(() => atrStore.stats.compliance);

// For Trends
const auditMainStats = computed(() => ({
  plannedAudit: plannedAuditCount.value,
  openFinding: openFindingsCount.value,
  executionStatus: executionStatusPercent.value,
  atrCompliance: atrCompliancePercent.value,
}));

// Helper function to calculate trend
const calculateTrend = (
  current: number,
  previous: number
): { icon: string; value: string; color: string } => {
  if (previous === 0) {
    return { icon: "", value: "N/A", color: "text-slate-400 dark:text-slate-300" };
  }

  const difference = current - previous;
  const percentageChange = (difference / previous) * 100;
  const absoluteChange = Math.abs(percentageChange).toFixed(1);

  if (percentageChange > 0) {
    return {
      icon: "↑",
      value: `${absoluteChange}%`,
      color: "text-green-600 dark:text-green-400",
    };
  } else if (percentageChange < 0) {
    return {
      icon: "↓",
      value: `${absoluteChange}%`,
      color: "text-red-600 dark:text-red-400",
    };
  } else {
    return {
      icon: "-",
      value: "0%",
      color: "text-slate-400",
    };
  }
};

// Comparison base for trends (neutral 0 when no historical snapshot exists)
const auditMainStatsLastMonth = {
  plannedAudit: 0,
  openFinding: 0,
  executionStatus: 0,
  atrCompliance: 0,
};

// Computed properties for each metric's trend
const plannedAuditTrend = computed(() =>
  calculateTrend(plannedAuditCount.value, auditMainStatsLastMonth.plannedAudit)
);

const openFindingTrend = computed(() =>
  calculateTrend(openFindingsCount.value, auditMainStatsLastMonth.openFinding)
);

const executionStatusTrend = computed(() =>
  calculateTrend(
    executionStatusPercent.value * 100,
    auditMainStatsLastMonth.executionStatus * 100
  )
);

const atrComplianceTrend = computed(() =>
  calculateTrend(
    atrCompliancePercent.value * 100,
    auditMainStatsLastMonth.atrCompliance * 100
  )
);

// Risk graph categories
const riskCategories = {
  inherentRisk: { name: "Inherent Risk", color: "#ff5c02" },
  residualRisk: { name: "Residual Risk", color: "#4d00ff" },
};

// Inherent (Q1) vs Residual (Q4) exposure per department, straight from the
// Corporate Risk Profile — the same series the RCM summary reports.
const mainRiskData = computed(() => rcmStore.inherentVsResidualByDepartment);

const yearlyFilters = [2024, 2025, 2026];
const activeYear = ref(2026);
const xFormatter = (x: number): string => `${mainRiskData.value[x]?.name}`;

// ATR Data — non-overlapping slices from the ATR store (same numbers as the ATR page summary).
const atrSliceColors: Record<AtrSliceKey, string> = ATR_SLICE_HEX;

const atrDonutData = computed(() =>
  atrStore.stats.breakdown.map((slice) => ({
    key: slice.key,
    name: t(`actionTakenReport.status.${slice.key}`),
    value: slice.percent,
  }))
);

const atrCategories = computed(() =>
  Object.fromEntries(
    atrDonutData.value.map((slice) => [
      slice.key,
      { name: slice.name, color: atrSliceColors[slice.key] },
    ])
  )
);

const atrTableData = computed(() => {
  return atrStore.summaryItems
    .map((r) => ({
      id: r.report_number || "-",
      name: r.finding_title,
      owner: r.pic_name || "-",
      date: formatAtrDate(r.due_date, locale.value),
      status: [
        atrStatusI18nKey(r.status) ? t(atrStatusI18nKey(r.status)!) : r.status,
        r.is_overdue ? t("actionTakenReport.status.overdue") : "",
      ]
        .filter(Boolean)
        .join(" · "),
    }))
    .slice(0, 5);
});

// Explicit widths: Action Item takes the slack of the full-width card, the rest stay compact.
const tableColumns = [
  { accessorKey: "id", header: "Audit ID", class: "min-w-40" },
  { accessorKey: "name", header: "Action Item", class: "min-w-64 w-full" },
  { accessorKey: "owner", header: "Owner", class: "min-w-40" },
  { accessorKey: "date", header: "Due Date", class: "min-w-32" },
  { accessorKey: "status", header: "Status", class: "min-w-32" },
];

const registeredRiskHeatMap = computed(() => riskProfileStore.risks.slice(0, 5));

const registeredRiskColumns = [
  {
    accessorKey: "name",
    header: "Risk Name",
    // Fixed width so long names don't push Score/Level around; name wraps to 2 lines,
    // and clipped text shows a tooltip on hover.
    class: "w-64 min-w-64 max-w-64",
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("div", { class: "flex flex-col min-w-0" }, [
        h(OverflowTooltip, {
          text: rawObject.name,
          textClass: "font-bold line-clamp-2 break-words whitespace-normal",
        }),
        h(OverflowTooltip, {
          text: rawObject.category,
          textClass: "text-md text-slate-400 truncate block",
        }),
      ]);
    },
  },
  {
    accessorKey: "severity",
    header: "Score",
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("span", { class: "font-bold" }, rawObject.impact * rawObject.likelihood);
    },
  },
  {
    id: "level",
    // Same level, label and badge colour as the CRP risk list (RiskHeatMap.vue).
    header: () => t("dashboard.registeredRisks.columns.level"),
    cell: (row: any) => {
      const risk = row.row.original;
      const level = riskProfileStore.getRiskLevel(risk.likelihood, risk.impact);
      return h(
        "span",
        {
          class: `inline-block px-2.5 py-1 rounded text-[10px] font-black tracking-tight ${getRiskLevelColorClass(level)}`,
        },
        t(`riskProfile.riskLevelLabels.${level}`) || riskLevelConfig[level]?.label || level
      );
    },
  },
];

// Audit Coverage
// Planned = every annual audit plan, Completed = status "Approved",
// Remaining = everything else (Draft, Pending Review, Pending Approval, Revised).
const auditCoverage = computed(() => {
  const plannedAudits = auditPlansCount.value;
  const completedAudits = completedAuditsCount.value;
  return {
    plannedAudits,
    completedAudits,
    remainingAudits: plannedAudits - completedAudits,
  };
});
const progressModel = computed(() => {
  const total = auditCoverage.value.plannedAudits;
  if (total === 0) return 0;
  return Math.round((auditCoverage.value.completedAudits / total) * 100);
});

// Audit Execution
// Every execution the Audit Execution Status page lists (the card scrolls).
const dashboardExecutionStatus = computed(() => {
  return auditExecutionStore.auditExecutions.map((e) => ({
    name: e.name,
    percentage: getProgressValue(e.progress),
  }));
});
// Period of the card: the current quarter, as "Q4 2026".
const executionPeriodLabel = computed(() => {
  const now = new Date();
  return `Q${Math.floor(now.getMonth() / 3) + 1} ${now.getFullYear()}`;
});

// Recent Findings — GET /audit-result-reports/recent-findings: saved ARR findings
// merged with live KKA / fieldwork findings, newest first (fetched on mount).
const recentFindingsData = computed(() =>
  auditResultStore.recentFindings.map((f) => ({
    audit_finding: f.title,
    findings_category: f.category,
  }))
);

const recentFindingsEmptyState = computed(() =>
  auditResultStore.recentFindingsError
    ? {
        icon: "i-heroicons-exclamation-triangle",
        label: t("dashboard.recentFindings.error"),
        description: auditResultStore.recentFindingsError,
      }
    : {
        icon: "i-heroicons-circle-stack-20-solid",
        label: t("dashboard.recentFindings.empty"),
      }
);

const findingCategoryKeys: Record<string, string> = {
  "Very Significant": "verySignificant",
  Significant: "significant",
  "Quite Significant": "quiteSignificant",
  "Not Significant": "notSignificant",
};

const auditTableColumns = [
  {
    accessorKey: "audit_finding",
    header: () => t("dashboard.recentFindings.columns.finding"),
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("div", { class: "flex flex-col" }, [
        h("span", { class: "font-bold" }, rawObject.audit_finding),
      ]);
    },
  },
  {
    accessorKey: "findings_category",
    header: () => t("dashboard.recentFindings.columns.category"),
    cell: (row: any) => {
      const findings_category = row.getValue();
      const categoryKey = findingCategoryKeys[findings_category];
      return h(
        UBadge,
        {
          color: findings_category === "Very Significant" ? "error" : findings_category === "Significant" ? "error" : findings_category === "Quite Significant" ? "warning" : "success",
          variant: "soft",
        },
        () => (categoryKey ? t(`dashboard.recentFindings.categories.${categoryKey}`) : findings_category)
      );
    },
  },
];

const getHeatMapCellColor = (x: number, y: number): string => {
  const probability = 6 - y;
  const impact = x;
  const level = riskProfileStore.getRiskLevel(probability, impact);
  switch (level) {
    case RiskLevel.HIGH:
      return "bg-red-500";
    case RiskLevel.MODERATE_HIGH:
      return "bg-orange-500";
    case RiskLevel.MODERATE:
      return "bg-yellow-500";
    case RiskLevel.LOW_MODERATE:
      return "bg-green-600";
    case RiskLevel.LOW:
      return "bg-green-400";
    default:
      return "bg-gray-100";
  }
};

const getRiskLevel = (x: number, y: number) => riskProfileStore.getRiskLevel(y, x);
</script>
