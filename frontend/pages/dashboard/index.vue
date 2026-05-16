<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col gap-2">
          <h1 class="text-3xl font-bold text-[var(--text-main)]">
            Dashboard
          </h1>
          <p class="text-sm text-[var(--text-muted)]">
            Enterprise Risk Management and Internal Audit Platform
          </p>
        </div>
      </template>
    </UCard>

    <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
                <p class="text-sm font-medium text-gray-600">Total Risks</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">{{ totalRisks }}</p>
              </div>
              <div class="rounded-lg bg-primary-100 p-3">
                <UIcon name="i-lucide-shield" class="text-primary-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-green-600">↑ 12%</span>
            <span class="text-gray-500">from last month</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">High Risk</p>
                <p class="text-3xl font-bold text-red-600 mt-1">{{ highRisks }}</p>
              </div>
              <div class="rounded-lg bg-red-100 p-3">
                <UIcon name="i-lucide-alert-triangle" class="text-red-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-red-600">↑ 3</span>
            <span class="text-gray-500">requires attention</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Audit Plans</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">{{ auditPlansCount }}</p>
              </div>
              <div class="rounded-lg bg-blue-100 p-3">
                <UIcon name="i-lucide-calendar" class="text-blue-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-blue-600">Active</span>
            <span class="text-gray-500">this quarter</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Completed</p>
                <p class="text-3xl font-bold text-green-600 mt-1">{{ completedAuditsCount }}</p>
              </div>
              <div class="rounded-lg bg-green-100 p-3">
                <UIcon name="i-lucide-check-circle" class="text-green-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-green-600">Updated</span>
            <span class="text-gray-500">this month</span>
          </div>
        </UCard>
      </div>

    <!-- Statistics Cards -->
    <UCard>
      <template #header>
        <h4 class=" font-semibold">Audit Statistics</h4>
      </template>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <UCard variant="outline" class="relative overflow-hidden">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Planned Audits</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">
                  {{ auditMainStats.plannedAudit }}
                </p>
              </div>
              <div class="rounded-lg bg-primary-100 p-1 center">
                <UIcon name="i-lucide-info" class="text-primary-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span :class="plannedAuditTrend.color">
              {{ plannedAuditTrend.icon }} {{ plannedAuditTrend.value }}
            </span>
            <span class="text-gray-500">from last month</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Open Findings</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">
                  {{ auditMainStats.openFinding }}
                </p>
              </div>
              <div class="rounded-lg bg-orange-100 p-3">
                <UIcon name="i-lucide-alert-triangle" class="text-orange-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span :class="openFindingTrend.color">
              {{ openFindingTrend.icon }} {{ openFindingTrend.value }}
            </span>
            <span class="text-gray-500">from last month</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">
                  Execution Status
                </p>
                <p class="text-3xl font-bold text-gray-900 mt-1">
                  {{ (auditMainStats.executionStatus * 100).toFixed(0) }}%
                </p>
              </div>
              <div class="rounded-lg bg-blue-100 p-3">
                <UIcon name="i-lucide-play" class="text-blue-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span :class="executionStatusTrend.color">
              {{ executionStatusTrend.icon }} {{ executionStatusTrend.value }}
            </span>
            <span class="text-gray-500">from last month</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">ATR Compliance</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">
                  {{ (auditMainStats.atrCompliance * 100).toFixed(0) }}%
                </p>
              </div>
              <div class="rounded-lg bg-green-100 p-3">
                <UIcon name="i-lucide-check-square" class="text-green-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span :class="atrComplianceTrend.color">
              {{ atrComplianceTrend.icon }} {{ atrComplianceTrend.value }}
            </span>
            <span class="text-gray-500">from last month</span>
          </div>
        </UCard>
      </div>
    </UCard>

    <div class="flex flex-row gap-8 flex-nowrap items-stretch">
      <div class="basis-3/12 w-full min-w-fit min-h-full gap-2 flex flex-col h-full">
        <UCard variant="soft" class="flex-1" :ui="{ header: 'px-0 pt-0 pb-4' }">
          <template #header>
            <div
              class="flex items-center justify-between px-6 border-b  py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Risk Heat Map
              </h2>
              <UButton to="/risk-profile" variant="ghost" color="neutral">Configure</UButton>
            </div>
          </template> 
          <div class="flex flex-col gap-2">
            <!-- Y-axis label -->
            <!-- Heat map grid -->
            <div class="flex flex-row gap-4 items-center mr-24">
              <div class="flex flex-col items-center">
                <span class="text-xs font-medium text-gray-600 mb-1" style="transform: rotate(270deg);">
                Probability
                </span>
              </div>
              <div class="min-w-full flex flex-col gap-4">
                <div class="grid grid-cols-5 gap-1 flex-1">
                  <template v-for="y in 5" :key="y">
                    <template v-for="x in 5" :key="`${x}-${y}`">
                      <div
                        class="aspect-square flex items-center justify-center text-xs font-semibold rounded"
                        :class="getHeatMapCellColor(x, y)"
                        :title="`Impact: ${x}, Probability: ${6-y}, Risk: ${getRiskLevel(x, 6-y)}`"
                      >
                      </div>
                    </template>
                  </template>
                </div>
                <div class="flex justify-center">
                  <span class="text-xs font-medium text-gray-600">Impact</span>
                </div>
              </div>
            </div>
            <!-- X-axis label -->
            <!-- Legend -->
            <div class="flex flex-wrap justify-center gap-2 mt-2">
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded bg-success-400"></div>
                <span class="text-xs text-gray-600">Low</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded bg-success-600"></div>
                <span class="text-xs text-gray-600">Low-Mod</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded bg-warning-500"></div>
                <span class="text-xs text-gray-600">Moderate</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded bg-primary-400"></div>
                <span class="text-xs text-gray-600">Mod-High</span>
              </div>
              <div class="flex items-center gap-1">
                <div class="w-3 h-3 rounded bg-error-500"></div>
                <span class="text-xs text-gray-600">High</span>
              </div>
            </div>
          </div>
        </UCard>
        <UCard variant="soft" class="flex-1 min-h-[27rem]" :ui="{ header: 'px-0 pt-0 pb-4', body: 'p-0' }">
          <template #header>
            <div
              class="flex items-center justify-between px-6 border-b  py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Registered Risk
              </h2>
              <UButton to="/risk-profile" variant="ghost" color="neutral">Configure</UButton>
            </div>
          </template> 
          <UTable :data="registeredRiskHeatMap" :columns="registeredRiskColumns" :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data yang dimasukkan ke Risk Heat Map.' }" class="w-full text-sm text-left" />
        </UCard>
      </div>
      <div class="basis-9/12 w-full min-w-fit gap-8 flex flex-col h-full">
        <UCard variant="soft" :ui="{ header: 'px-0 pt-0 pb-4' }">
          <template #header>
            <div
              class="flex items-center justify-between px-6 border-b  py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Audit Planning Coverage
              </h2>
              <UButton to="/annual-audit" variant="ghost" color="neutral">Configure</UButton>
            </div>
          </template>
          <template #default>
            <div class="h-fit flex flex-col gap-8">
              <div class="gap-2">
                <div class="self-end text-end">
                <h5 class="text-lg font-semibold text-secondary-600! mb-2">
                  {{ progressModel }}% Completed
                </h5>
              </div>
              <UProgress v-model="progressModel" color="secondary" />
              </div>
              <div class="flex flex-row  font-semibold gap-32 flex-wrap px-4 justify-around">
                <div class="flex flex-row gap-16">
                  <p>Planned Audits</p>
                  <p>{{ auditCoverage.plannedAudits }}</p>
                </div class="flex flex-row">
                <div class="flex flex-row gap-16">
                  <p>Completed Audits</p>
                  <p>{{ auditCoverage.completedAudits }}</p>
                </div class="flex flex-row">
                <div class="flex flex-row gap-16">
                  <p>Remaining</p>
                  <p>{{ auditCoverage.remainingAudits }}</p>
                </div class="flex flex-row">
              </div>
            </div>
          </template>
        </UCard>
        <UCard variant="soft">
          <template #header>
            <div class="flex flex-row justify-between min-h-full ">
              <h2>Inherent vs Residual Risk by Department</h2>
              <USelect v-model="activeYear" :options="yearlyFilters" />
            </div>
          </template>
          <BarChart
            :data="mainRiskData"
            :categories="riskCategories"
            :x-formatter="xFormatter"
            :y-axis="['inherentRisk', 'residualRisk']"
            :radius="4"
            :height="500"
            :padding="{
              top: 0,
              left: 0,
              right: 0,
              bottom: 4
            }"
            :tooltip-title-formatter="tooltipTitleFormatter"
            :x-label=label.xLabel
            :y-label=label.yLabel
          />
        </UCard>
      </div>
    </div>
    <div class="flex flex-row gap-8 flex-nowrap items-stretch flex-2">
    <div class="flex-1/2">
      <UCard>
        <template #header>
          <div class="flex flex-row gap-8 justify-between items-center">
            <div>
              <h5>Action Taken Report</h5>
            </div>
            <div>
              <UButton to="/action-taken-report" variant="ghost" color="neutral">View Details</UButton>
            </div>
          </div>
        </template>
        <DonutChart :data="atrDonutData.map((item) => item.value)" :height="400" :categories="atrCategories" :radius="10" :hide-legend="false" :type="(DonutType.Full as any)" :padding="{ top: 0, left: 0, right: 0, bottom: 0}" :arc-width="20" :legend-position="(LegendPosition.TopRight as any)" :legend-style=legendStyle :pad-angle="0">
          <div class="text-center">
            <div class="font-semibold ">
              <span class="text-xl">{{ atrDonutData[0]!.value }}%</span>  Completed
            </div>
          </div>
        </DonutChart>
      </UCard>
    </div>
    <div class="flex-1/2 min-h-96">
      <UCard class="flex-1/2 min-h-full">
        <template #header>
           <div class="flex flex-row gap-8 justify-between items-center">
             <div>
               <h5>Action Taken Reports (ATR)</h5>
             </div>
             <div>
               <UButton to="/action-taken-report" variant="ghost" color="neutral">View Full Report</UButton>
             </div>
           </div>
          </template>
        <UTable :data="atrTableData" :columns="tableColumns" :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }" class="w-full text-sm text-left" />
      </UCard>
    </div>
  </div>

  <div class="flex flex-row gap-8 flex-nowrap">
    <UCard class="basis-1/2">
      <template #header>
        <div class="flex flex-row gap-1 justify-between">
          <h2 class="text-xl font-semibold text-gray-900">Audit Execution Status</h2>
          <UBadge color="primary" label="Q4 2025" variant="soft"></UBadge>
        </div>
      </template>
      <div v-for="(item, index) in dashboardExecutionStatus" :key="index">
        <div class="gap-4 flex flex-col pb-6">
          <div class="flex flex-row justify-between">
            <h3 class="text-lg font-semibold text-gray-900">
              {{ item.name }}
            </h3>
            <p class="text-sm text-gray-600">
              {{ item.percentage }}%
            </p>
          </div>
          <UProgress v-model="item.percentage" color="secondary" />
        </div>
      </div>
    </UCard>
    <UCard class="basis-1/2">
      <template #header>
        <div class="flex flex-row gap-1 justify-between">
          <h2 class="text-xl font-semibold text-gray-900">Recent Finding Issues</h2>
        </div>
      </template>
      <UTable :data="recentFindingsData" :columns="auditTableColumns"></UTable>
    </UCard>
  </div>

    <!-- Recent Activity & Overview -->
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <!-- Recent Risk Profiles -->
      <UCard variant="soft">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">
                Recent Risk Profiles
              </h3>
              <p class="text-sm text-gray-600">Latest risk assessments</p>
            </div>
            <UButton
              icon="i-lucide-chevron-right"
              variant="ghost"
              color="primary"
              to="/risk-profile"
            />
          </div>
        </template>
        <div class="space-y-3">
          <div
            v-for="(risk, index) in riskProfileStore.risks.slice(0, 4)"
            :key="risk.id"
            class="flex items-center justify-between p-3 rounded-lg  border border-gray-200 hover:border-primary-300 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-full bg-primary-100 w-10 h-10 flex items-center justify-center"
              >
                <span class="text-sm font-semibold text-primary-700">{{
                  index + 1
                }}</span>
              </div>
              <div>
                <p class="font-medium text-gray-900">
                  {{ risk.name }}
                </p>
                <p class="text-sm text-gray-500">{{ risk.category }}</p>
              </div>
            </div>
            <UBadge :color="risk.impact * risk.likelihood > 15 ? 'red' : 'warning'" variant="soft">
               {{ risk.impact * risk.likelihood > 15 ? 'High' : 'Medium' }}
            </UBadge>
          </div>
        </div>
      </UCard>

      <!-- Upcoming Audits -->
      <UCard variant="soft">
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h3 class="text-lg font-semibold text-gray-900">
                Upcoming Audits
              </h3>
              <p class="text-sm text-gray-600">Scheduled audit activities</p>
            </div>
            <UButton
              icon="i-lucide-chevron-right"
              variant="ghost"
              color="primary"
              to="/annual-audit"
            />
          </div>
        </template>
        <div class="space-y-3">
          <div
            v-for="plan in annualPlanStore.plans.slice(0, 4)"
            :key="plan.id"
            class="flex items-center justify-between p-3 rounded-lg  border border-gray-200 hover:border-primary-300 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-lg bg-blue-100 w-10 h-10 flex items-center justify-center"
              >
                <UIcon name="i-lucide-calendar" class="text-blue-600 size-5" />
              </div>
              <div>
                <p class="font-medium text-gray-900">{{ plan.code }}</p>
                <p class="text-sm text-gray-500">
                  Status: {{ plan.status }}
                </p>
              </div>
            </div>
            <UBadge color="info" variant="soft">Scheduled</UBadge>
          </div>
          <div v-if="annualPlanStore.plans.length === 0" class="text-center py-4 text-gray-500 text-sm">
             No upcoming audits scheduled.
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import { useRiskProfileStore } from "~/stores/risk-profile";
import { useAnnualPlanStore } from "~/stores/annual-audit";
import { useActionTakenReportStore } from "~/stores/action-taken-report";
import { useAuditExecutionStore } from "~/stores/audit-execution";
import { useAuditResultReportStore } from "~/stores/audit-result-report";
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

const UButton = resolveComponent("UButton");
const UBadge = resolveComponent("UBadge");

// Constants for Charts
const DonutType = {
  Full: 'full',
  Half: 'half'
};

const LegendPosition = {
  Top: 'top',
  Bottom: 'bottom',
  Left: 'left',
  Right: 'right',
  TopRight: 'top-right'
};

// Stats for Header
const totalRisks = computed(() => riskProfileStore.risks.length);
const highRisks = computed(() => 
  riskProfileStore.risks.filter(r => 
    riskProfileStore.getRiskLevel(r.likelihood, r.impact) === RiskLevel.HIGH
  ).length
);
const auditPlansCount = computed(() => annualPlanStore.plans.length);
const completedAuditsCount = computed(() => 
  annualPlanStore.plans.filter(p => p.status === 'Done').length
);

// Audit Statistics (Section 2)
const plannedAuditCount = computed(() => 
  annualPlanStore.plans.filter(p => p.status !== 'Done').length
);
const openFindingsCount = computed(() => 
  atrStore.reportList.filter(r => r.status !== AuditStatus.COMPLETED).length
);
const executionStatusPercent = computed(() => {
  const executions = auditExecutionStore.auditExecutions;
  if (executions.length === 0) return 0;
  return executions.reduce((sum, e) => sum + e.progress, 0) / (executions.length * 100);
});
const atrCompliancePercent = computed(() => (atrStore.stats.donePercent / 100) || 0);

// For Trends (keep mockup logic for now as we don't have historical data easily accessible)
const auditMainStats = computed(() => ({
  plannedAudit: plannedAuditCount.value,
  openFinding: openFindingsCount.value,
  executionStatus: executionStatusPercent.value,
  atrCompliance: atrCompliancePercent.value,
}));

// Helper function to calculate trend
const calculateTrend = (
  current: number,
  previous: number,
): { icon: string; value: string; color: string } => {
  if (previous === 0) {
    return { icon: "", value: "N/A", color: "text-gray-500" };
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
      color: "text-gray-500",
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
  calculateTrend(
    plannedAuditCount.value,
    auditMainStatsLastMonth.plannedAudit,
  ),
);

const openFindingTrend = computed(() =>
  calculateTrend(
    openFindingsCount.value,
    auditMainStatsLastMonth.openFinding,
  ),
);

const executionStatusTrend = computed(() =>
  calculateTrend(
    executionStatusPercent.value * 100,
    auditMainStatsLastMonth.executionStatus * 100,
  ),
);

const atrComplianceTrend = computed(() =>
  calculateTrend(
    atrCompliancePercent.value * 100,
    auditMainStatsLastMonth.atrCompliance * 100,
  ),
);

// Risk graph
const riskCategories = {
  inherentRisk: { name: "Inherent Risk", color: "#ff5c02" },
  residualRisk: { name: "Residual Risk", color: "#4d00ff" },
};

// Simplified risk data for bar chart based on real risk data
const mainRiskData = computed(() => {
  const departments = ['Finance', 'IT', 'Operations', 'Legal', 'HR'];
  return departments.map(dept => {
    const deptRisks = riskProfileStore.risks.filter(r => r.category === dept || (dept === 'IT' && r.category === 'Technology'));
    return {
      name: dept,
      inherentRisk: deptRisks.reduce((sum, r) => sum + (r.impact * r.likelihood), 0) / (deptRisks.length || 1),
      residualRisk: deptRisks.reduce((sum, r) => sum + (r.impact * r.likelihood * 0.6), 0) / (deptRisks.length || 1), // Mocking residual as 60% of inherent
    }
  });
});

const yearlyFilters = [2024, 2025, 2026];
const activeYear = ref(2026);
const xFormatter = (x: number): string => `${mainRiskData.value[x]?.name}`
const tooltipTitleFormatter = (x: any): string => `${x.name}`

const label = {
  xLabel: "Department",
  yLabel: "Risk Score",
}

// ATR Data
const atrDonutData = computed(() => [
  { name: "Completed", value: atrStore.stats.donePercent },
  { name: "In Progress", value: atrStore.stats.wipPercent },
  { name: "Overdue", value: atrStore.stats.latePercent },
]);

const atrCategories = {
  primary: {
    name: 'Completed',
    color: 'var(--color-secondary-500)'
  },
  secondary: {
    name: 'In Progress',
    color: 'var(--color-neutral-500)'
  },
  tertiary: {
    name: 'Overdue',
    color: 'var(--color-primary-500)'
  },
}

const legendStyle : Record<string, string> = 
  {
      fontSize: '14px',
      fontWeight: '600',
      color: '#374151',
      marginTop: '0px',
      padding: '10px',
      marginBottom: '24px',
      backgroundColor: '#f9fafb',
      borderRadius: '8px'
  }

const atrTableData = computed(() => {
  return atrStore.reportList.map(r => ({
    id: r.auditRef,
    name: r.title,
    owner: r.pic || '-',
    date: r.deadline,
    status: r.status
  })).slice(0, 5);
});

const tableColumns = [
  { accessorKey: 'id', header: 'Audit ID' },
  { accessorKey: 'name', header: 'Action Item' },
  { accessorKey: 'owner', header: 'Owner' },
  { accessorKey: 'date', header: 'Due Date' },
  { accessorKey: 'status', header: 'Status' }
]

const registeredRiskHeatMap = computed(() => riskProfileStore.risks.slice(0, 5))

const registeredRiskColumns = [
  {
    id: 'id',
    header: 'ID',
    cell: (row: any) => {
      const risk = row.row.original
      const formattedId = riskProfileStore.getFormattedId(risk)
      // Menggunakan font-mono agar konsisten dengan tampilan di RiskHeatMap
      return h('p', { class: 'font-medium text-center font-mono' }, formattedId)
    },
  },
  {
    accessorKey: 'name',
    header: 'Risk Name',
    cell: (row : any) => {
      const rawObject = row.row.original;
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'font-bold' }, rawObject.name),
        h('span', { class: 'text-xs text-gray-500' }, rawObject.category)
      ])
    }
  },
  {
    accessorKey: 'severity',
    header: 'Score',
    cell: (row : any) => {
      const rawObject = row.row.original;
      return h('span', { class: 'font-bold' }, rawObject.impact * rawObject.likelihood)
    }
  }
]

// Audit Coverage
const auditCoverage = computed(() => ({
  plannedAudits: plannedAuditCount.value,
  completedAudits: completedAuditsCount.value,
  remainingAudits: annualPlanStore.plans.filter(p => p.status === 'Not Available' || p.status === 'Work In Progress').length
}));
const progressModel = computed(() => {
  const total = annualPlanStore.plans.length;
  if (total === 0) return 0;
  return Math.round((completedAuditsCount.value / total) * 100);
});

// Audit Execution
const dashboardExecutionStatus = computed(() => {
  return auditExecutionStore.auditExecutions.map(e => ({
    name: e.name,
    percentage: e.progress
  })).slice(0, 3);
});

// Recent Findings
const recentFindingsData = computed(() => {
  return auditResultStore.reportList.map(r => ({
    audit_finding: r.reportTitle,
    audit_category: r.overallRating,
    severity: r.findingsCount > 5 ? 'High' : r.findingsCount > 2 ? 'Medium' : 'Low'
  })).slice(0, 3);
});

const auditTableColumns = [
  {
    accessorKey: 'audit_finding',
    header: 'Audit Finding',
    cell: (row : any) => {
      const rawObject = row.row.original;
      return h('div', { class: 'flex flex-col' }, [
        h('span', { class: 'font-bold' }, rawObject.audit_finding),
        h('span', { class: 'text-xs text-gray-500' }, rawObject.audit_category)
      ])
    }
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
    cell: (row: any) => {
      const severity = row.getValue();
      return h(UBadge, {
        color: severity === 'High' ? 'red' : severity === 'Medium' ? 'orange' : 'green',
        variant: 'soft'
      }, () => severity)
    }
  }
]

const getHeatMapCellColor = (x: number, y: number): string => {
  const probability = 6 - y;
  const impact = x;
  const level = riskProfileStore.getRiskLevel(probability, impact);
  switch (level) {
    case RiskLevel.HIGH: return 'bg-red-500';
    case RiskLevel.MODERATE_HIGH: return 'bg-orange-500';
    case RiskLevel.MODERATE: return 'bg-yellow-500';
    case RiskLevel.LOW_MODERATE: return 'bg-green-600';
    case RiskLevel.LOW: return 'bg-green-400';
    default: return 'bg-gray-100';
  }
};

const getRiskLevel = (x: number, y: number) => riskProfileStore.getRiskLevel(y, x);
</script>
