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
            class="flex items-center gap-1 text-xs font-semibold text-emerald-500 mt-1"
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
          <div class="flex items-center gap-1 text-xs font-semibold text-red-500 mt-1">
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
          <div class="flex items-center gap-1 text-xs font-semibold text-sky-500 mt-1">
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
            class="flex items-center gap-1 text-xs font-semibold text-emerald-500 mt-1"
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
            <span class="text-xs font-semibold text-slate-500">Planned Audits</span>
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
            <span class="text-xs font-semibold text-slate-500">Open Findings</span>
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
            <span class="text-xs font-semibold text-slate-500">Execution Status</span>
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
            <span class="text-xs font-semibold text-slate-500">ATR Compliance</span>
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
            <USelect
              v-model="activeYear"
              :options="yearlyFilters"
              size="sm"
              class="w-24  "
            />
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
            <span class="text-xs font-semibold text-slate-600">Inherent Risk</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full bg-[#4d00ff]"></span>
            <span class="text-xs font-semibold text-slate-600">Residual Risk</span>
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
                class="text-xs font-semibold text-slate-500 uppercase tracking-wider origin-center -rotate-90 whitespace-nowrap"
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
                      class="w-10 h-10 flex items-center justify-center text-xs font-semibold rounded shadow-sm hover:scale-105 transition-transform cursor-pointer"
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
                  class="text-xs font-semibold text-slate-500 uppercase tracking-wider"
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
          <UTable
            :data="registeredRiskHeatMap"
            :columns="registeredRiskColumns"
            :empty-state="{
              icon: 'i-heroicons-circle-stack-20-solid',
              label: 'Belum ada data yang dimasukkan ke Risk Heat Map.',
            }"
            class="w-full text-sm border border-slate-100 rounded-xl overflow-hidden"
          />
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
              <span class="text-xs font-semibold text-slate-500">Overall Progress</span>
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
        <UTable
          :data="atrTableData"
          :columns="tableColumns"
          :empty-state="{
            icon: 'i-heroicons-circle-stack-20-solid',
            label: 'Belum ada data rencana audit.',
          }"
          class="w-full text-sm border border-slate-100 rounded-xl overflow-hidden"
        />
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
              <p class="text-xs font-bold text-indigo-600">{{ item.percentage }}%</p>
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
        <UTable
          :data="recentFindingsData"
          :columns="auditTableColumns"
          class="w-full text-sm border border-slate-100 rounded-xl overflow-hidden"
        />
      </div>
    </div>

    <!-- Seventh Row: Recent Risk Profiles & Upcoming Audits -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Risk Profiles -->
      <div class="  border border-slate-100 shadow-sm rounded-2xl p-6 space-y-6">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-lg font-bold text-slate-800">Recent Risk Profiles</h3>
            <p class="text-xs text-slate-400 mt-1">Latest risk assessments</p>
          </div>
          <UButton
            icon="i-lucide-chevron-right"
            variant="ghost"
            color="primary"
            to="/risk-profile"
            size="sm"
          />
        </div>
        <div class="space-y-3">
          <div
            v-for="(risk, index) in riskProfileStore.risks.slice(0, 4)"
            :key="risk.id"
            class="flex items-center justify-between p-3.5 rounded-xl border border-slate-100 hover:border-indigo-200 hover: -50/50 transition-all duration-200"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-full bg-indigo-50 w-8 h-8 flex items-center justify-center"
              >
                <span class="text-xs font-bold text-indigo-700">{{ index + 1 }}</span>
              </div>
              <div>
                <p class="text-sm font-bold text-slate-800">
                  {{ risk.name }}
                </p>
                <p class="text-xs text-slate-400 mt-0.5">{{ risk.category }}</p>
              </div>
            </div>
            <UBadge
              :color="risk.impact * risk.likelihood > 15 ? 'red' : 'warning'"
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
            <p class="text-xs text-slate-400 mt-1">Scheduled audit activities</p>
          </div>
          <UButton
            icon="i-lucide-chevron-right"
            variant="ghost"
            color="primary"
            to="/annual-audit"
            size="sm"
          />
        </div>
        <div class="space-y-3">
          <div
            v-for="plan in annualPlanStore.plans.slice(0, 4)"
            :key="plan.id"
            class="flex items-center justify-between p-3.5 rounded-xl border border-slate-100 hover:border-indigo-200 hover: -50/50 transition-all duration-200"
          >
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-sky-50 w-8 h-8 flex items-center justify-center">
                <UIcon name="i-lucide-calendar" class="text-sky-600 size-4" />
              </div>
              <div>
                <p class="text-sm font-bold text-slate-800">{{ plan.code }}</p>
                <p class="text-xs text-slate-400 mt-0.5">Status: {{ plan.status }}</p>
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
import { useI18n } from "~/composables/useI18n";
import { useRiskProfileStore } from "~/stores/risk-profile";
import { useAnnualPlanStore } from "~/stores/annual-audit";
import { useActionTakenReportStore } from "~/stores/action-taken-report";
import { useAuditExecutionStore } from "~/stores/audit-execution";
import { useAuditResultReportStore } from "~/stores/audit-result-report";
import { useAuthStore } from "~/stores/auth";
import { RiskLevel } from "~/types/risk";
import { AuditStatus } from "~/types/audit";

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

const UButton = resolveComponent("UButton");
const UBadge = resolveComponent("UBadge");

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
        h("span", { class: "text-xs text-slate-400" }, rawObject.category),
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

// Recent Findings
const recentFindingsData = computed(() => {
  return auditResultStore.reportList
    .map((r) => ({
      audit_finding: r.reportTitle,
      audit_category: r.overallRating,
      severity: r.findingsCount > 5 ? "High" : r.findingsCount > 2 ? "Medium" : "Low",
    }))
    .slice(0, 3);
});

const auditTableColumns = [
  {
    accessorKey: "audit_finding",
    header: "Audit Finding",
    cell: (row: any) => {
      const rawObject = row.row.original;
      return h("div", { class: "flex flex-col" }, [
        h("span", { class: "font-bold" }, rawObject.audit_finding),
        h("span", { class: "text-xs text-slate-400" }, rawObject.audit_category),
      ]);
    },
  },
  {
    accessorKey: "severity",
    header: "Severity",
    cell: (row: any) => {
      const severity = row.getValue();
      return h(
        UBadge,
        {
          color: severity === "High" ? "red" : severity === "Medium" ? "orange" : "green",
          variant: "soft",
        },
        () => severity
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
