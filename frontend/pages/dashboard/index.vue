<template>
  <div class="p-4 md:p-6 space-y-8 min-h-screen">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div>
        <h1 class="text-3xl font-bold text-slate-900 tracking-tight">Dashboard</h1>
        <p class="text-sm text-slate-500 mt-1.5">
          Welcome back, {{ authStore.getUser?.fullName || "Andi" }}. Here's your risk &
          audit performance overview.
        </p>
      </div>
      <div>
        <UButton
          variant="outline"
          color="neutral"
          class="  border border-slate-200 text-slate-700 hover: -50 shadow-sm font-medium"
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
        class="  border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between h-[135px]"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500">Total Risks</p>
          <div class="rounded-xl bg-indigo-50 p-2.5 flex items-center justify-center">
            <UIcon name="i-lucide-shield" class="text-indigo-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-3xl font-bold text-slate-900 tracking-tight">
            {{ totalRisks }}
          </h3>
          <div
            class="flex items-center gap-1 text-md font-semibold text-emerald-500 mt-1"
          >
            <span>↑ 12%</span>
            <span class="text-slate-400 font-normal">from last month</span>
          </div>
        </div>
      </div>

      <!-- Card 2: High Risk -->
      <div
        class="  border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between h-[135px]"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500">High Risk</p>
          <div class="rounded-xl bg-red-50 p-2.5 flex items-center justify-center">
            <UIcon name="i-lucide-alert-triangle" class="text-red-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-3xl font-bold text-red-600 tracking-tight">{{ highRisks }}</h3>
          <div class="flex items-center gap-1 text-md font-semibold text-red-500 mt-1">
            <span>↑ 3</span>
            <span class="text-slate-400 font-normal">requires attention</span>
          </div>
        </div>
      </div>

      <!-- Card 3: Audit Plans -->
      <div
        class="  border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between h-[135px]"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500">Audit Plans</p>
          <div class="rounded-xl bg-sky-50 p-2.5 flex items-center justify-center">
            <UIcon name="i-lucide-calendar" class="text-sky-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-3xl font-bold text-slate-900 tracking-tight">
            {{ auditPlansCount }}
          </h3>
          <div class="flex items-center gap-1 text-md font-semibold text-sky-500 mt-1">
            <span>Active</span>
            <span class="text-slate-400 font-normal">this quarter</span>
          </div>
        </div>
      </div>

      <!-- Card 4: Completed -->
      <div
        class="  border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between h-[135px]"
      >
        <div class="flex justify-between items-start">
          <p class="text-sm font-semibold text-slate-500">Completed</p>
          <div class="rounded-xl bg-emerald-50 p-2.5 flex items-center justify-center">
            <UIcon name="i-lucide-check-circle" class="text-emerald-600 size-5" />
          </div>
        </div>
        <div class="mt-2">
          <h3 class="text-3xl font-bold text-emerald-600 tracking-tight">
            {{ completedAuditsCount }}
          </h3>
          <div
            class="flex items-center gap-1 text-md font-semibold text-emerald-500 mt-1"
          >
            <span>Updated</span>
            <span class="text-slate-400 font-normal">this month</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Audit Statistics Section (Second Row) -->
    <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
      <div class="flex items-center justify-between">
        <h3 class="text-lg font-bold text-slate-800">Audit Statistics</h3>
      </div>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Planned Audits -->
        <div
          class=" border border-slate-100 rounded-xl p-5 flex flex-col justify-between h-[120px]"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500">Planned Audits</span>
            <div class="rounded-lg bg-indigo-50 p-1.5 flex items-center justify-center">
              <UIcon name="i-lucide-info" class="text-indigo-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900">
              {{ auditMainStats.plannedAudit }}
            </h4>
            <p class="text-[11px] mt-1" :class="plannedAuditTrend.color">
              {{ plannedAuditTrend.icon }} {{ plannedAuditTrend.value }}
              <span class="text-slate-400 font-normal">from last month</span>
            </p>
          </div>
        </div>

        <!-- Open Findings -->
        <div
          class=" -50 border border-slate-100 rounded-xl p-5 flex flex-col justify-between h-[120px]"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500">Open Findings</span>
            <div class="rounded-lg bg-amber-50 p-1.5 flex items-center justify-center">
              <UIcon name="i-lucide-alert-triangle" class="text-amber-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900">
              {{ auditMainStats.openFinding }}
            </h4>
            <p class="text-[11px] mt-1" :class="openFindingTrend.color">
              {{ openFindingTrend.icon }} {{ openFindingTrend.value }}
              <span class="text-slate-400 font-normal">from last month</span>
            </p>
          </div>
        </div>

        <!-- Execution Status -->
        <div
          class=" -50 border border-slate-100 rounded-xl p-5 flex flex-col justify-between h-[120px]"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500">Execution Status</span>
            <div class="rounded-lg bg-sky-50 p-1.5 flex items-center justify-center">
              <UIcon name="i-lucide-play" class="text-sky-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900">
              {{ (auditMainStats.executionStatus * 100).toFixed(0) }}%
            </h4>
            <p class="text-[11px] mt-1" :class="executionStatusTrend.color">
              {{ executionStatusTrend.icon }} {{ executionStatusTrend.value }}
              <span class="text-slate-400 font-normal">from last month</span>
            </p>
          </div>
        </div>

        <!-- ATR Compliance -->
        <div
          class=" -50 border border-slate-100 rounded-xl p-5 flex flex-col justify-between h-[120px]"
        >
          <div class="flex justify-between items-start">
            <span class="text-md font-semibold text-slate-500">ATR Compliance</span>
            <div class="rounded-lg bg-emerald-50 p-1.5 flex items-center justify-center">
              <UIcon name="i-lucide-check-square" class="text-emerald-600 size-4" />
            </div>
          </div>
          <div class="mt-1">
            <h4 class="text-2xl font-bold text-slate-900">
              {{ (auditMainStats.atrCompliance * 100).toFixed(0) }}%
            </h4>
            <p class="text-[11px] mt-1" :class="atrComplianceTrend.color">
              {{ atrComplianceTrend.icon }} {{ atrComplianceTrend.value }}
              <span class="text-slate-400 font-normal">from last month</span>
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
            <h2 class="text-xl font-bold text-slate-900 dark:text-white tracking-tight">
              Predictive Intelligence & Anomaly Analytics
            </h2>
          </div>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
            Analisis proyeksi masa depan KPI dan deteksi anomali finansial/operasional.
          </p>
        </div>

        <NuxtLink
          to="/analytics"
          class="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 flex items-center gap-1 self-start sm:self-auto"
        >
          <span>Buka Deep Analytics Center</span>
          <span>&rarr;</span>
        </NuxtLink>
      </div>

      <!-- Grid of 2 Main Analytics Cards -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
        <!-- CARD 1: KPI FORECASTING (PyTorch LSTM) -->
        <div class="lg:col-span-6 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 shadow-sm rounded-2xl p-6 flex flex-col justify-between space-y-6">
          <div>
            <!-- Card Header -->
            <div class="flex items-center justify-between mb-4">
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="text-lg font-bold text-slate-900 dark:text-white">
                    KPI Performance Forecasting
                  </h3>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Proyeksi tren historis vs estimasi Q3 2026 - Q1 2027 (Interval Kepercayaan 95%)
                </p>
              </div>

              <NuxtLink to="/analytics/kpi-forecast" class="text-xs font-semibold text-violet-600 hover:underline">
                Rincian
              </NuxtLink>
            </div>

            <!-- Line Chart Component -->
            <div class="h-[280px] w-full relative">
              <Line :data="kpiChartData" :options="kpiChartOptions" />
            </div>

            <!-- Forecast Highlights Table / Cards -->
            <div class="mt-6 space-y-3">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                At-Risk KPI Projections & Audit Recommendations
              </h4>

              <div
                v-for="kpi in timeseriesData.kpiForecasts.slice(0, 3)"
                :key="kpi.code"
                class="p-3.5 rounded-xl border border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
              >
                <div class="space-y-1">
                  <div class="flex items-center gap-2">
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
                  <span class="text-xs text-slate-400">Proyeksi</span>
                  <span class="text-sm font-extrabold font-mono text-violet-600 dark:text-violet-400">
                    {{ kpi.forecastedValue }}%
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Footer Info -->
          <div class="pt-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-information-circle" class="w-4 h-4 text-violet-500" />
              Interval Prediksi ±4.8% MAPE
            </span>
            <span class="font-medium text-slate-700 dark:text-slate-300">Target Horizon: Q3 2026</span>
          </div>
        </div>

        <!-- CARD 2: ANOMALY DETECTION (Isolation Forest) -->
        <div class="lg:col-span-6 bg-white dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 shadow-sm rounded-2xl p-6 flex flex-col justify-between space-y-6">
          <div>
            <!-- Card Header -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
              <div>
                <div class="flex items-center gap-2">
                  <h3 class="text-lg font-bold text-slate-900 dark:text-white">
                    Transaction Anomaly Detection
                  </h3>
                </div>
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
                  Deteksi pencilan transaksi & pola tak wajar berbasis machine learning yang tidak terawasi
                </p>
              </div>
            </div>

            <!-- Scatter Chart Component -->
            <div class="h-[280px] w-full relative">
              <Scatter :data="anomalyScatterChartData" :options="anomalyScatterOptions" />
            </div>

            <!-- Top Detected Anomaly List -->
            <div class="mt-6 space-y-3">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">
                Critical Anomaly Alerts Requiring Investigation
              </h4>

              <div
                v-for="anm in filteredAnomalies.slice(0, 3)"
                :key="anm.id"
                class="p-3.5 rounded-xl border border-rose-100 dark:border-rose-900/40 bg-rose-50/40 dark:bg-rose-950/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
              >
                <div class="space-y-1">
                  <div class="flex items-center gap-2">
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
                  <span class="text-xs font-mono text-rose-600 dark:text-rose-400 font-bold">
                    Score: {{ anm.anomalyScore }}
                  </span>
                  <span v-if="anm.amount" class="text-xs font-bold text-slate-800 dark:text-slate-200">
                    Rp {{ (anm.amount / 1000000).toLocaleString() }}M
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Footer Info -->
          <div class="pt-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500">
            <span class="flex items-center gap-1.5">
              <UIcon name="i-heroicons-shield-exclamation" class="w-4 h-4 text-rose-500" />
              Tingkat Kontaminasi: {{ (isolationData.summary.contaminationRate * 100).toFixed(1) }}%
            </span>
            <NuxtLink to="/analytics" class="font-semibold text-rose-600 dark:text-rose-400 hover:underline">
              Kelola Semua {{ isolationData.anomalies.length }} Anomali &rarr;
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>

    <!-- Internal Control Effectiveness Section (COSO 2013) -->
    <div class="bg-white border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="text-md font-semibold uppercase tracking-wider text-slate-400">Internal Audit & Risk Metric</span>
            <UBadge color="neutral" variant="subtle" size="md">COSO 2013</UBadge>
            <span class="text-md font-semibold text-primary-700 bg-primary-50 border border-primary-100 px-2.5 py-0.5 rounded-md ml-1">
              Tahun: {{ rcmStore.selectedYear }} | Dep: {{ rcmStore.selectedDepartment }}
            </span>
          </div>
          <h2 class="text-xl font-bold text-slate-900 tracking-tight mt-0.5">
            Internal Control Effectiveness
          </h2>
        </div>
        <UButton to="/risk-profile/risk-control-matrix" variant="outline" color="neutral" size="sm" class="font-medium">
          Risk Control Matrix &rarr;
        </UButton>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 items-stretch">
        <!-- Effectiveness Score Card -->
        <div class="bg-slate-50 border border-slate-100 rounded-xl p-5 flex flex-col justify-between">
          <div>
            <div class="flex justify-between items-start">
              <span class="text-md font-semibold text-slate-500 uppercase tracking-wider">Effectiveness Rating</span>
              <div class="rounded-lg bg-primary-100 p-1.5 flex items-center justify-center">
                <UIcon name="i-lucide-shield-check" class="text-primary-600 size-4" />
              </div>
            </div>
            <div class="mt-3 flex items-baseline gap-2">
              <h3 class="text-3xl font-extrabold text-slate-900 tracking-tight">
                {{ rcmStore.internalControlEffectiveness }}%
              </h3>
            </div>
            <div class="mt-2">
              <UBadge
                :color="rcmStore.effectivenessRating.badgeColor as any"
                variant="solid"
                class="font-bold px-2.5 py-1"
              >
                {{ rcmStore.effectivenessRating.rating }}
              </UBadge>
            </div>
          </div>

          <div class="mt-4 pt-3 border-t border-slate-200/60 text-md text-slate-600 space-y-1">
            <div class="flex justify-between">
              <span>Inherent Risk (Prioritas):</span>
              <span class="font-bold text-slate-900">{{ rcmStore.totalInherentRisk }} Risiko</span>
            </div>
            <div class="flex justify-between">
              <span>Residual Risk (Sisa):</span>
              <span class="font-bold text-red-600">{{ rcmStore.totalResidualRisk }} Risiko</span>
            </div>
          </div>
        </div>

        <!-- Interpretation Text -->
        <div class="lg:col-span-2 bg-slate-50 border border-slate-100 rounded-xl p-5 flex flex-col justify-between">
          <div>
            <span class="text-md font-semibold text-slate-500 uppercase tracking-wider">Interpretasi Hasil COSO 2013</span>
            <div class="mt-2 p-3 rounded-lg border text-md" :class="rcmStore.effectivenessRating.bgClass">
              <p class="font-bold flex items-center gap-1.5 mb-1">
                <UIcon name="i-lucide-check-circle-2" class="size-4" />
                {{ rcmStore.effectivenessRating.rating }} ({{ rcmStore.internalControlEffectiveness }}%)
              </p>
              <p class="leading-relaxed">{{ rcmStore.effectivenessRating.interpretation }}</p>
            </div>
          </div>

          <!-- COSO Dimensions Mini Progress in % -->
          <div class="mt-4 pt-3 border-t border-slate-200/60 grid grid-cols-5 gap-2 text-center text-md">
            <div>
              <span class="block text-slate-500 font-medium">Design</span>
              <span class="font-bold text-slate-800">{{ rcmStore.cosoAverages.design }}%</span>
            </div>
            <div>
              <span class="block text-slate-500 font-medium">Operating</span>
              <span class="font-bold text-slate-800">{{ rcmStore.cosoAverages.operating }}%</span>
            </div>
            <div>
              <span class="block text-slate-500 font-medium">Coverage</span>
              <span class="font-bold text-slate-800">{{ rcmStore.cosoAverages.coverage }}%</span>
            </div>
            <div>
              <span class="block text-slate-500 font-medium">Timeliness</span>
              <span class="font-bold text-slate-800">{{ rcmStore.cosoAverages.timeliness }}%</span>
            </div>
            <div>
              <span class="block text-slate-500 font-medium">Automation</span>
              <span class="font-bold text-slate-800">{{ rcmStore.cosoAverages.automation }}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Charts Section (Third Row) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Inherent vs Residual Risk by Department Chart -->
      <div
        class="lg:col-span-8   border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between"
      >
        <div>
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-bold text-slate-800">
              Inherent vs Residual Risk by Department
            </h3>
          </div>
          <div class="h-[350px] w-full">
            <BarChart
              :data="mainRiskData"
              :categories="riskCategories"
              :x-formatter="xFormatter"
              :y-axis="['inherentRisk', 'residualRisk']"
              :radius="6"
              :height="350"
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
        <div class="flex justify-center gap-6 mt-4 border-t border-slate-50 pt-4">
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full bg-[#ff5c02]"></span>
            <span class="text-md font-semibold text-slate-600">Inherent Risk</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full bg-[#4d00ff]"></span>
            <span class="text-md font-semibold text-slate-600">Residual Risk</span>
          </div>
        </div>
      </div>

      <!-- Action Taken Report Donut Chart -->
      <div
        class="lg:col-span-4   border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between"
      >
        <div>
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-bold text-slate-800">Action Taken Report</h3>
            <UButton to="/action-taken-report" variant="ghost" color="neutral" size="sm"
              >Details</UButton
            >
          </div>
          <div class="h-[220px] w-full relative flex items-center justify-center">
            <DonutChart
              :data="atrDonutData.map((item) => item.value)"
              :categories="atrCategories"
              :radius="8"
              :arc-width="26"
              :height="220"
              :hide-legend="true"
            >
              <div class="text-center flex flex-col items-center justify-center">
                <span class="text-xl font-bold text-slate-900 leading-tight"
                  >{{ atrStore.stats.donePercent }}%</span
                >
                <span class="text-[10px] text-slate-500 font-medium">Completed</span>
              </div>
            </DonutChart>
          </div>
        </div>
        <!-- Custom Legend -->
        <div
          class="flex flex-wrap justify-center gap-x-4 gap-y-2 mt-4 px-2 border-t border-slate-50 pt-4"
        >
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-full bg-[#4d00ff]"></span>
            <span class="text-[11px] font-semibold text-slate-600"
              >Completed ({{ atrStore.stats.donePercent }}%)</span
            >
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-full  -400"></span>
            <span class="text-[11px] font-semibold text-slate-600"
              >In Progress ({{ atrStore.stats.wipPercent }}%)</span
            >
          </div>
          <div class="flex items-center gap-1.5">
            <span class="w-2.5 h-2.5 rounded-full bg-[#ff5c02]"></span>
            <span class="text-[11px] font-semibold text-slate-600"
              >Overdue ({{ atrStore.stats.latePercent }}%)</span
            >
          </div>
        </div>
      </div>
    </div>

    <!-- Risk Heat Map & Registered Risk (Fourth Row) -->
    <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-bold text-slate-900 tracking-tight">Risk Profiles</h2>
        <UButton to="/risk-profile" variant="ghost" color="neutral" size="sm"
          >Configure</UButton
        >
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        <!-- Heat Map Grid on the left -->
        <div
          class="lg:col-span-5 flex flex-col items-center  -50/50 border border-slate-100 rounded-2xl p-6"
        >
          <h4 class="text-sm font-semibold text-slate-700 mb-6 w-full text-left">
            Risk Heat Map
          </h4>
          <div class="flex flex-row gap-4 items-center">
            <!-- Y-axis label -->
            <div class="flex flex-col items-center">
              <span
                class="text-md font-semibold text-slate-500 uppercase tracking-wider origin-center -rotate-90 whitespace-nowrap"
              >
                Probability
              </span>
            </div>
            <!-- Heat map grid -->
            <div class="flex flex-col gap-2">
              <div class="grid grid-cols-5 gap-1.5">
                <template v-for="y in 5" :key="y">
                  <template v-for="x in 5" :key="`${x}-${y}`">
                    <div
                      class="w-10 h-10 flex items-center justify-center text-md font-semibold rounded shadow-sm hover:scale-105 transition-transform cursor-pointer"
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
                  class="text-md font-semibold text-slate-500 uppercase tracking-wider"
                  >Impact</span
                >
              </div>
            </div>
          </div>
          <!-- Legend -->
          <div class="flex flex-wrap justify-center gap-x-3 gap-y-1.5 mt-6 max-w-[320px]">
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-green-400"></div>
              <span class="text-[11px] font-medium text-slate-600">Low</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-green-600"></div>
              <span class="text-[11px] font-medium text-slate-600">Low-Mod</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-yellow-500"></div>
              <span class="text-[11px] font-medium text-slate-600">Moderate</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-orange-500"></div>
              <span class="text-[11px] font-medium text-slate-600">Mod-High</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="w-3 h-3 rounded bg-red-500"></div>
              <span class="text-[11px] font-medium text-slate-600">High</span>
            </div>
          </div>
        </div>

        <!-- Registered Risks Table on the right -->
        <div class="lg:col-span-7   border border-slate-100 rounded-2xl p-6">
          <h4 class="text-sm font-semibold text-slate-700 mb-4">Registered Risks</h4>
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

    <!-- Fifth Row: Audit Coverage & Action Taken Reports Table -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Audit Planning Coverage -->
      <div
        class="lg:col-span-4   border border-slate-100 shadow-sm rounded-2xl p-6 flex flex-col justify-between"
      >
        <div>
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-bold text-slate-800">Audit Planning Coverage</h3>
            <UButton to="/annual-audit" variant="ghost" color="neutral" size="sm"
              >Configure</UButton
            >
          </div>
          <div class="space-y-4">
            <div class="flex justify-between items-center">
              <span class="text-md font-semibold text-slate-500">Overall Progress</span>
              <h5 class="text-base font-bold text-indigo-600">
                {{ progressModel }}% Completed
              </h5>
            </div>
            <UProgress
              :model-value="progressModel"
              color="secondary"
              class="h-2 rounded"
            />
          </div>
        </div>
        <div
          class="grid grid-cols-3 gap-2 border-t border-slate-50 pt-6 mt-6 text-center"
        >
          <div>
            <p class="text-[11px] text-slate-400 font-semibold uppercase">Planned</p>
            <h4 class="text-lg font-bold text-slate-800 mt-1">
              {{ auditCoverage.plannedAudits }}
            </h4>
          </div>
          <div>
            <p class="text-[11px] text-slate-400 font-semibold uppercase">Completed</p>
            <h4 class="text-lg font-bold text-slate-800 mt-1">
              {{ auditCoverage.completedAudits }}
            </h4>
          </div>
          <div>
            <p class="text-[11px] text-slate-400 font-semibold uppercase">Remaining</p>
            <h4 class="text-lg font-bold text-slate-800 mt-1">
              {{ auditCoverage.remainingAudits }}
            </h4>
          </div>
        </div>
      </div>

      <!-- Action Taken Reports Table -->
      <div
        class="lg:col-span-8   border border-slate-100 shadow-sm rounded-2xl p-6"
      >
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-bold text-slate-800">Action Taken Reports (ATR)</h3>
          <UButton to="/action-taken-report" variant="ghost" color="neutral" size="sm"
            >View Full Report</UButton
          >
        </div>
        <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
          <TableEntities
            :data="atrTableData"
            :columns="tableColumns"
            :empty-state="{
              icon: 'i-heroicons-circle-stack-20-solid',
              label: 'Belum ada data rencana audit.',
            }"
            class="w-full"
          />
        </UCard>
      </div>
    </div>

    <!-- Sixth Row: Audit Execution Status & Recent Finding Issues -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Audit Execution Status -->
      <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-bold text-slate-900 tracking-tight">
            Audit Execution Status
          </h2>
          <UBadge color="primary" label="Q4 2025" variant="soft"></UBadge>
        </div>
        <div class="space-y-5">
          <div
            v-for="(item, index) in dashboardExecutionStatus"
            :key="index"
            class="space-y-2"
          >
            <div class="flex justify-between items-center">
              <h3 class="text-sm font-semibold text-slate-700">
                {{ item.name }}
              </h3>
              <p class="text-md font-bold text-indigo-600">{{ item.percentage }}%</p>
            </div>
            <UProgress
              :model-value="item.percentage"
              color="secondary"
              class="h-1.5 rounded"
            />
          </div>
        </div>
      </div>

      <!-- Recent Finding Issues -->
      <div class="  border border-slate-100 shadow-sm rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-slate-900 tracking-tight">
            Recent Finding Issues
          </h2>
        </div>
        <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ body: 'p-0' }">
          <TableEntities
            :data="recentFindingsData"
            :columns="auditTableColumns"
            class="w-full"
          />
        </UCard>
      </div>
    </div>

    <!-- Seventh Row: Recent Risk Profiles & Upcoming Audits -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Risk Profiles -->
      <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-bold text-slate-800">Recent Risk Profiles</h3>
            <p class="text-md text-slate-400 mt-1">Latest risk assessments</p>
          </div>
          <UTooltip text="View all Risk Profiles">
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
            class="flex items-center justify-between p-3.5 rounded-xl border border-slate-200"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-full bg-indigo-50 w-8 h-8 flex items-center justify-center"
              >
                <span class="text-md font-bold text-indigo-700">{{ index + 1 }}</span>
              </div>
              <div>
                <p class="text-sm font-bold text-slate-800">
                  {{ risk.name }}
                </p>
                <p class="text-md text-slate-400 mt-0.5">{{ risk.category }}</p>
              </div>
            </div>
            <UBadge
              :color="risk.impact * risk.likelihood > 15 ? 'error' : 'warning'"
              variant="soft"
              size="sm"
              class="font-semibold"
            >
              {{ risk.impact * risk.likelihood > 15 ? "High" : "Medium" }}
            </UBadge>
          </div>
        </div>
      </div>

      <!-- Upcoming Audits -->
      <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-bold text-slate-800">Upcoming Audits</h3>
            <p class="text-md text-slate-400 mt-1">Scheduled audit activities</p>
          </div>
          <UTooltip text="View all Annual Audits">
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
            class="flex items-center justify-between p-3.5 rounded-xl border border-slate-200"
          >
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-sky-50 w-8 h-8 flex items-center justify-center">
                <UIcon name="i-lucide-calendar" class="text-sky-600 size-4" />
              </div>
              <div>
                <p class="text-sm font-bold text-slate-800">{{ plan.code }}</p>
                <p class="text-md text-slate-400 mt-0.5">Status: {{ plan.status }}</p>
              </div>
            </div>
            <UBadge color="info" variant="soft" size="sm" class="font-semibold"
              >Scheduled</UBadge
            >
          </div>
          <div
            v-if="annualPlanStore.plans.length === 0"
            class="text-center py-6 text-slate-400 text-sm"
          >
            No upcoming audits scheduled.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
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
import {
  useTimeSeriesData,
  useIsolationForestData
} from "~/composables/useAnalyticsData";
import { useI18n } from "~/composables/useI18n";
import { useRiskProfileStore } from "~/stores/risk-profile";
import { useAnnualPlanStore } from "~/stores/annual-audit";
import { useActionTakenReportStore } from "~/stores/action-taken-report";
import { useAuditExecutionStore } from "~/stores/audit-execution";
import { useAuditResultReportStore } from "~/stores/audit-result-report";
import { useAuthStore } from "~/stores/auth";
import { useRCMStore } from "~/stores/rcm";
import { RiskLevel } from "~/types/risk";
import { AuditStatus } from "~/types/audit";
import { UBadge } from "#components";

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

const { t } = useI18n();

const riskProfileStore = useRiskProfileStore();
const annualPlanStore = useAnnualPlanStore();
const atrStore = useActionTakenReportStore();
const auditExecutionStore = useAuditExecutionStore();
const auditResultStore = useAuditResultReportStore();
const authStore = useAuthStore();
const rcmStore = useRCMStore();

// AI & Analytics Composables
const timeseriesData = useTimeSeriesData();
const isolationData = useIsolationForestData();
const filteredAnomalies = computed(() => isolationData.anomalies);

// ─── KPI Forecasting Line Chart Config ─────────────────────
const kpiChartData = computed(() => ({
  labels: timeseriesData.historicalKPI.map((p) => p.period),
  datasets: [
    {
      label: "Actual KPI Score (%)",
      data: timeseriesData.historicalKPI.map((p) => p.actual),
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
      label: "KPI Forecast (%)",
      data: timeseriesData.historicalKPI.map((p) => p.forecast),
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
      label: "Upper Bound (95% CI)",
      data: timeseriesData.historicalKPI.map((p) => p.upperBound),
      borderColor: "transparent",
      backgroundColor: "rgba(139, 92, 246, 0.08)",
      fill: "+1",
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
    {
      label: "Lower Bound (95% CI)",
      data: timeseriesData.historicalKPI.map((p) => p.lowerBound),
      borderColor: "transparent",
      backgroundColor: "rgba(139, 92, 246, 0.08)",
      fill: "-1",
      tension: 0.4,
      pointRadius: 0,
      spanGaps: false,
    },
  ],
}));

const kpiChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: "top" as const,
      labels: {
        usePointStyle: true,
        font: { family: "Inter, sans-serif", size: 12, weight: "bold" as const },
        filter: (item: any) =>
          !["Upper Bound (95% CI)", "Lower Bound (95% CI)"].includes(item.text),
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
          if (
            ["Upper Bound (95% CI)", "Lower Bound (95% CI)"].includes(
              ctx.dataset.label
            )
          )
            return "";
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
    y: {
      min: 50,
      max: 85,
      grid: { color: "rgba(226, 232, 240, 0.6)" },
      ticks: {
        font: { family: "Inter, sans-serif", size: 11 },
        callback: (v: any) => `${v}%`,
      },
      title: {
        display: true,
        text: "KPI Index Score (%)",
        font: { size: 12, weight: "bold" as const },
      },
    },
  },
};

// ─── Anomaly Detection Scatter Chart Config ────────────

const anomalyScatterChartData = computed(() => {
  const normalPoints = isolationData.scatterData.filter((s) => !s.isAnomaly);
  const anomalyList = filteredAnomalies.value;

  const criticalPoints = anomalyList
    .filter((a) => a.severity === "Critical")
    .map((a, idx) => ({
      x: a.anomalyScore * -100,
      y: (idx + 1) * 12 + 15,
      label: a.id,
      title: a.entity,
    }));

  const highPoints = anomalyList
    .filter((a) => a.severity === "High")
    .map((a, idx) => ({
      x: a.anomalyScore * -100,
      y: (idx + 1) * 9 + 8,
      label: a.id,
      title: a.entity,
    }));

  const mediumPoints = anomalyList
    .filter((a) => a.severity === "Medium")
    .map((a, idx) => ({
      x: a.anomalyScore * -100,
      y: (idx + 1) * 7 + 5,
      label: a.id,
      title: a.entity,
    }));

  return {
    datasets: [
      {
        label: "Baseline Operational Data",
        data: normalPoints.map((p) => ({ x: p.x, y: p.y })),
        backgroundColor: "rgba(148, 163, 184, 0.35)",
        borderColor: "rgba(148, 163, 184, 0.6)",
        pointRadius: 4,
        pointHoverRadius: 6,
      },
      {
        label: "Critical Anomalies",
        data: criticalPoints,
        backgroundColor: "#ef4444",
        borderColor: "#b91c1c",
        pointRadius: 9,
        pointHoverRadius: 12,
        pointStyle: "triangle",
      },
      {
        label: "High Severity Outliers",
        data: highPoints,
        backgroundColor: "#f97316",
        borderColor: "#c2410c",
        pointRadius: 8,
        pointHoverRadius: 10,
        pointStyle: "rectRot",
      },
      {
        label: "Medium Risk Outliers",
        data: mediumPoints,
        backgroundColor: "#eab308",
        borderColor: "#a16207",
        pointRadius: 7,
        pointHoverRadius: 9,
        pointStyle: "circle",
      },
    ],
  };
});

const anomalyScatterOptions = {
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
          const pt = ctx.raw;
          if (pt.label) {
            return `[${pt.label}] ${pt.title}: Anomaly Score ${pt.x?.toFixed(1)}`;
          }
          return `Normal Baseline Point (Score: ${pt.x?.toFixed(1)})`;
        },
      },
    },
  },
  scales: {
    x: {
      title: {
        display: true,
        text: "Anomaly Deviation Index",
        font: { size: 11, weight: "bold" as const },
      },
      grid: { color: "rgba(226, 232, 240, 0.6)" },
      ticks: { font: { size: 11 } },
    },
    y: {
      title: {
        display: true,
        text: "Frequency / Transaction Spread",
        font: { size: 11, weight: "bold" as const },
      },
      grid: { display: false },
      ticks: { font: { size: 11 } },
    },
  },
};

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
      return "error";
    default:
      return "warning";
  }
};

// Sync state and action simulation
const isSyncing = ref(false);
const handleSync = async () => {
  isSyncing.value = true;
  await new Promise((resolve) => setTimeout(resolve, 1000));
  isSyncing.value = false;
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
  () => annualPlanStore.plans.filter((p) => p.status === "Done").length
);

// Audit Statistics (Section 2)
const plannedAuditCount = computed(
  () => annualPlanStore.plans.filter((p) => p.status !== "Done").length
);
const openFindingsCount = computed(
  () => atrStore.reportList.filter((r) => r.status !== AuditStatus.COMPLETED).length
);
const executionStatusPercent = computed(() => {
  const executions = auditExecutionStore.auditExecutions;
  if (executions.length === 0) return 0;
  return executions.reduce((sum, e) => sum + e.progress, 0) / (executions.length * 100);
});
const atrCompliancePercent = computed(() => atrStore.stats.donePercent / 100 || 0);

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
    return { icon: "", value: "N/A", color: "text-slate-400" };
  }

  const difference = current - previous;
  const percentageChange = (difference / previous) * 100;
  const absoluteChange = Math.abs(percentageChange).toFixed(1);

  if (percentageChange > 0) {
    return {
      icon: "↑",
      value: `${absoluteChange}%`,
      color: "text-green-600",
    };
  } else if (percentageChange < 0) {
    return {
      icon: "↓",
      value: `${absoluteChange}%`,
      color: "text-red-600",
    };
  } else {
    return {
      icon: "-",
      value: "0%",
      color: "text-slate-400",
    };
  }
};

// Mock last month for trends
const auditMainStatsLastMonth = {
  plannedAudit: 25,
  openFinding: 12,
  executionStatus: 0.7,
  atrCompliance: 0.8,
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

// Simplified risk data for bar chart based on real risk data
const mainRiskData = computed(() => {
  const departments = ["Finance", "IT", "Operations", "Legal", "HR"];
  return departments.map((dept) => {
    const deptRisks = riskProfileStore.risks.filter(
      (r) => r.category === dept || (dept === "IT" && r.category === "Technology")
    );
    return {
      name: dept,
      inherentRisk:
        deptRisks.reduce((sum, r) => sum + r.impact * r.likelihood, 0) /
        (deptRisks.length || 1),
      residualRisk:
        deptRisks.reduce((sum, r) => sum + r.impact * r.likelihood * 0.6, 0) /
        (deptRisks.length || 1), // Mocking residual as 60% of inherent
    };
  });
});

const yearlyFilters = [2024, 2025, 2026];
const activeYear = ref(2026);
const xFormatter = (x: number): string => `${mainRiskData.value[x]?.name}`;

// ATR Data
const atrDonutData = computed(() => [
  { name: "Completed", value: atrStore.stats.donePercent },
  { name: "In Progress", value: atrStore.stats.wipPercent },
  { name: "Overdue", value: atrStore.stats.latePercent },
]);

const atrCategories = {
  completed: { name: "Completed", color: "#4d00ff" },
  inProgress: { name: "In Progress", color: "#94a3b8" },
  overdue: { name: "Overdue", color: "#ff5c02" },
};

const atrTableData = computed(() => {
  return atrStore.reportList
    .map((r) => ({
      id: r.auditRef,
      name: r.title,
      owner: r.pic || "-",
      date: r.deadline,
      status: r.status,
    }))
    .slice(0, 5);
});

const tableColumns = [
  { accessorKey: "id", header: "Audit ID" },
  { accessorKey: "name", header: "Action Item" },
  { accessorKey: "owner", header: "Owner" },
  { accessorKey: "date", header: "Due Date" },
  { accessorKey: "status", header: "Status" },
];

const registeredRiskHeatMap = computed(() => riskProfileStore.risks.slice(0, 5));

const registeredRiskColumns = [
  {
    id: "id",
    header: "ID",
    cell: (row: any) => {
      const risk = row.row.original;
      const formattedId = riskProfileStore.getFormattedId(risk);
      return h("p", { class: "font-medium text-center font-mono" }, formattedId);
    },
  },
  {
    accessorKey: "name",
    header: "Risk Name",
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("div", { class: "flex flex-col" }, [
        h("span", { class: "font-bold" }, rawObject.name),
        h("span", { class: "text-md text-slate-400" }, rawObject.category),
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
];

// Audit Coverage
const auditCoverage = computed(() => ({
  plannedAudits: plannedAuditCount.value,
  completedAudits: completedAuditsCount.value,
  remainingAudits: annualPlanStore.plans.filter(
    (p) => p.status === "Not Available" || p.status === "Work In Progress"
  ).length,
}));
const progressModel = computed(() => {
  const total = annualPlanStore.plans.length;
  if (total === 0) return 0;
  return Math.round((completedAuditsCount.value / total) * 100);
});

// Audit Execution
const dashboardExecutionStatus = computed(() => {
  return auditExecutionStore.auditExecutions
    .map((e) => ({
      name: e.name,
      percentage: e.progress,
    }))
    .slice(0, 3);
});

// Recent Findings — flattened from Audit Result Report findings
const recentFindingsData = computed(() => {
  const rows: { audit_finding: string; findings_category: string }[] = []
  for (const report of auditResultStore.reportList) {
    if (report.findings && report.findings.length > 0) {
      for (const f of report.findings) {
        const findings_category = f.category || ''
        rows.push({
          audit_finding: f.title,
          findings_category
        })
      }
    }
  }
  return rows.slice(0, 5)
});

const auditTableColumns = [
  {
    accessorKey: "audit_finding",
    header: "Audit Finding",
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("div", { class: "flex flex-col" }, [
        h("span", { class: "font-bold" }, rawObject.audit_finding),
      ]);
    },
  },
  {
    accessorKey: "findings_category",
    header: "Findings Category",
    cell: (row: any) => {
      const findings_category = row.getValue();
      return h(
        UBadge,
        {
          color: findings_category === "Very Significant" ? "error" : findings_category === "Significant" ? "error" : findings_category === "Quite Significant" ? "warning" : "success",
          variant: "soft",
        },
        () => findings_category
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
