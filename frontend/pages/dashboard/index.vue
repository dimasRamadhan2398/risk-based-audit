<template>
  <div class="space-y-6">
    <!-- Header Section -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col gap-2">
          <h1 class="text-3xl font-bold text-gray-900">
            {{ t("dashboard.title") }}
          </h1>
          <p class="text-sm text-gray-600">
            Enterprise Risk Management and Internal Audit Platform
          </p>
        </div>
      </template>
    </UCard>

    <UCard>
      <template #header>
              <h4 class="text-black font-semibold">Risk Management Statistics</h4>
      </template>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        <UCard variant="outline" class="relative overflow-hidden">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Total Risks</p>
                <p class="text-3xl font-bold text-gray-900 mt-1">24</p>
              </div>
              <div class="rounded-lg bg-primary-100 p-3">
                <UIcon name="info" class="text-primary-600 size-6" />
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
                <p class="text-3xl font-bold text-red-600 mt-1">5</p>
              </div>
              <div class="rounded-lg bg-red-100 p-3">
                <UIcon name="warning" class="text-red-600 size-6" />
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
                <p class="text-3xl font-bold text-gray-900 mt-1">8</p>
              </div>
              <div class="rounded-lg bg-blue-100 p-3">
                <UIcon name="plan" class="text-blue-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-blue-600">2 active</span>
            <span class="text-gray-500">this quarter</span>
          </div>
        </UCard>

        <UCard variant="outline">
          <template #header>
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-600">Completed</p>
                <p class="text-3xl font-bold text-green-600 mt-1">15</p>
              </div>
              <div class="rounded-lg bg-green-100 p-3">
                <UIcon name="check" class="text-green-600 size-6" />
              </div>
            </div>
          </template>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-green-600">↑ 8</span>
            <span class="text-gray-500">this month</span>
          </div>
        </UCard>
      </div>
    </UCard>

    <!-- Statistics Cards -->
    <UCard>
      <template #header>
        <h4 class="text-black font-semibold">Audit Statistics</h4>
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
                <UIcon name="info" class="text-primary-600 size-6" />
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
                <UIcon name="warning" class="text-orange-600 size-6" />
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
                <UIcon name="plan" class="text-blue-600 size-6" />
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
                <UIcon name="check" class="text-green-600 size-6" />
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
              class="flex items-center justify-between px-6 border-b border-black py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Risk Heat Map
              </h2>
              <h6 class="text-xl font-semibold text-gray-900">
                Configure
              </h6>
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
              class="flex items-center justify-between px-6 border-b border-black py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Registered Risk
              </h2>
              <h6 class="text-xl font-semibold text-gray-900">
                Configure
              </h6>
            </div>
          </template> 
          <UTable :data="registeredRiskHeatMap" :columns="registeredRiskColumns" :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data yang dimasukkan ke Risk Heat Map.' }" class="w-full text-sm text-left" />
        </UCard>
      </div>
      <div class="basis-9/12 w-full min-w-fit gap-8 flex flex-col h-full">
        <UCard variant="soft" :ui="{ header: 'px-0 pt-0 pb-4' }">
          <template #header>
            <div
              class="flex items-center justify-between px-6 border-b border-black py-4 -mx-6"
            >
              <h2 class="text-xl font-semibold text-gray-900">
                Audit Planning Coverage
              </h2>
              <h6 class="text-xl font-semibold text-gray-900">
                View Full Plan
              </h6>
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
              <div class="flex flex-row text-black font-semibold gap-32 flex-wrap px-4 justify-around">
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
            <div class="flex flex-row justify-between min-h-full text-black">
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
              <h6>View Details</h6>
            </div>
          </div>
        </template>
        <DonutChart :data="atrDonut.map((item) => item.value)" :height="400" :categories="atrCategories" :radius="10" :hide-legend="false" :type="DonutType.Full" :padding="{ top: 0, left: 0, right: 0, bottom: 0}" :arc-width="20" :legend-position="LegendPosition.TopRight" :legend-style=legendStyle :pad-angle="0">
          <div class="text-center">
            <div class="font-semibold text-black">
              <span class="text-xl">{{ atrDonut[0]!.value }}%</span>  Completed
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
               <h6>View Full Report</h6>
             </div>
           </div>
          </template>
        <UTable :data="tableData" :columns="tableColumns" :empty-state="{ icon: 'i-heroicons-circle-stack-20-solid', label: 'Belum ada data rencana audit.' }" class="w-full text-sm text-left" />
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
      <div v-for="(item, index) in auditExecutionStatus" :key="index">
        <div class="gap-4 flex flex-col pb-6">
          <div class="flex flex-row justify-between">
            <h3 class="text-lg font-semibold text-gray-900">
              {{ item.name }}
            </h3>
            <p class="text-sm text-gray-600">
              {{ doubleToPercentage(item.percentage) }}
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
      <UTable :data="auditFindingData" :columns="auditTableColumns"></UTable>
    </UCard>
  </div>

    <!-- Quick Actions Menu -->
    <UCard variant="soft">
      <template #header>
        <div class="flex flex-col gap-1">
          <h2 class="text-xl font-semibold text-gray-900">Quick Actions</h2>
          <p class="text-sm text-gray-600">Access key modules and features</p>
        </div>
      </template>
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="(item, index) in menuList" :key="index">
          <NuxtLink :to="item.href">
            <UCard
              variant="outline"
              class="hover:shadow-lg transition-all duration-200 hover:border-primary-400 cursor-pointer group"
            >
              <div class="flex items-center gap-4">
                <div
                  class="rounded-lg bg-primary-100 p-3 group-hover:bg-primary-200 transition-colors"
                >
                  <UIcon :name="item.icon" class="text-primary-600 size-6" />
                </div>
                <div class="flex-1 min-w-0">
                  <h3 class="font-semibold text-gray-900 truncate">
                    {{ item.name }}
                  </h3>
                  <p class="text-sm text-gray-600 truncate">
                    {{ item.description }}
                  </p>
                </div>
                <UIcon
                  name="chevron-right"
                  class="text-gray-400 group-hover:text-primary-600 transition-colors"
                />
              </div>
            </UCard>
          </NuxtLink>
        </div>
      </div>
    </UCard>

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
              icon="chevron-right"
              variant="ghost"
              color="primary"
              to="/risk-profile"
            />
          </div>
        </template>
        <div class="space-y-3">
          <div
            v-for="i in 4"
            :key="i"
            class="flex items-center justify-between p-3 rounded-lg bg-white border border-gray-200 hover:border-primary-300 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-full bg-primary-100 w-10 h-10 flex items-center justify-center"
              >
                <span class="text-sm font-semibold text-primary-700">{{
                  i
                }}</span>
              </div>
              <div>
                <p class="font-medium text-gray-900">
                  Operational Risk {{ i }}
                </p>
                <p class="text-sm text-gray-500">Updated 2 days ago</p>
              </div>
            </div>
            <UBadge color="warning" variant="soft">Medium</UBadge>
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
              icon="chevron-right"
              variant="ghost"
              color="primary"
              to="/strategic-audit-plan"
            />
          </div>
        </template>
        <div class="space-y-3">
          <div
            v-for="i in 4"
            :key="i"
            class="flex items-center justify-between p-3 rounded-lg bg-white border border-gray-200 hover:border-primary-300 transition-colors"
          >
            <div class="flex items-center gap-3">
              <div
                class="rounded-lg bg-blue-100 w-10 h-10 flex items-center justify-center"
              >
                <UIcon name="plan" class="text-blue-600 size-5" />
              </div>
              <div>
                <p class="font-medium text-gray-900">Q{{ i }} Audit Plan</p>
                <p class="text-sm text-gray-500">
                  Starting in {{ i * 7 }} days
                </p>
              </div>
            </div>
            <UBadge color="info" variant="soft">Scheduled</UBadge>
          </div>
        </div>
      </UCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "~/composables/useI18n";
import { MenuList } from "./menu";

definePageMeta({
  middleware: "auth",
});

const { t } = useI18n();

const riskDataStore = useRiskProfileStore();
const auditDataStore = useAuditDataStore();
const menuList = ref(MenuList);
const progressModel = ref(50);

const auditMainStats = computed(() => auditDataStore.getAuditMainStats);
const auditMainStatsLastMonth = computed(
  () => auditDataStore.getAuditMainStatsLastMonth,
);
const auditCoverage = computed(() => auditDataStore.getAuditCoverage);
const UButton = resolveComponent("UButton");
const UBadge = resolveComponent("UBadge");

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

// Computed properties for each metric's trend
const plannedAuditTrend = computed(() =>
  calculateTrend(
    auditMainStats.value.plannedAudit,
    auditMainStatsLastMonth.value.plannedAudit,
  ),
);

const openFindingTrend = computed(() =>
  calculateTrend(
    auditMainStats.value.openFinding,
    auditMainStatsLastMonth.value.openFinding,
  ),
);

const executionStatusTrend = computed(() =>
  calculateTrend(
    auditMainStats.value.executionStatus * 100,
    auditMainStatsLastMonth.value.executionStatus * 100,
  ),
);

const atrComplianceTrend = computed(() =>
  calculateTrend(
    auditMainStats.value.atrCompliance * 100,
    auditMainStatsLastMonth.value.atrCompliance * 100,
  ),
);

// Risk graph
const riskCategories = {
  inherentRisk: { name: "Inherent Risk", color: "#ff5c02" }, // red-500
  residualRisk: { name: "Residual Risk", color: "#4d00ff" }, // green-500
};

const mainRiskData = computed(() => auditDataStore.getRiskData);

const yearlyFilters = computed(() => auditDataStore.getDropdownYear);
const activeYear = ref(new Date().getFullYear());
const xFormatter = (x: number): string => `${mainRiskData.value[x]?.name}`
const tooltipTitleFormatter = (x: any): string => `${x.name}`

const label = {
  xLabel: "Department",
  yLabel: "Risk Score",
}

onMounted(() => {
  auditDataStore.fetchAuditData();
});

//Action Taken Report
const atrDonut = [
  {
    name: "Completed",
    value: 90,
  },
  {
    name: "In Progress",
    value: 5,
  },
  {
    name: "Overdue",
    value: 5,
  },
]

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

const tableData = computed(() => {
  return auditDataStore.getAtrReports
})

const auditExecutionStatus = computed(() => auditDataStore.getAuditExecutionStatus)

const tableColumns = [
  {
    accessorKey: 'id',
    header: 'Audit ID',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'name',
    header: 'Action Item',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'owner',
    header: 'Owner',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'date',
    header: 'Due Date',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  }
]

const registeredRiskHeatMap = computed(() => riskDataStore.getRegisteredRisks)

const registeredRiskColumns = [
  {
    id: 'id',
    header: 'ID',

    cell: (row: any) => {
      return h('p', { class: 'text-black font-medium text-center' }, row.index + 1);
    },
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'risk_name',
    header: 'Risk Name',

    cell: (row : any) => {
      const rawObject = toRaw(row.row.original);
      return h('div', {
        class: 'flex items-start gap-2'
      }, [
          h('div', undefined, [
          h('h5', { class: 'font-medium text-highlighted' }, `${rawObject.risk_name}`), 
          h('p', { class: 'text-left' }, `${rawObject.risk_category}`),
        ])
      ])
    },
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-left",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'latest_residual_risk',
    header: 'Latest Residual Risk',
    cell: (row : any) => {
      const rawObject = toRaw(row.row.original);
      return h('p', {
        class: 'flex items-start gap-2'
      }, [
          h('div', undefined, [
          h('h5', { class: 'font-medium text-highlighted' }, `${rawObject.audit_finding}`), 
          h('p', { class: 'text-left' }, `${rawObject.audit_category}`),
        ])
      ])
    },
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-left",
        td: "text-black font-medium text-center",
      },
    },
  },
]

const auditFindingData = computed(() => auditDataStore.getRecentFindings)

const auditTableColumns = [
  {
    accessorKey: 'audit_finding',
    header: 'Audit Finding',

    cell: (row : any) => {
      const rawObject = toRaw(row.row.original);
      return h('div', {
        class: 'flex items-start gap-2'
      }, [
          h('div', undefined, [
          h('h5', { class: 'font-medium text-highlighted' }, `${rawObject.audit_finding}`), 
          h('p', { class: 'text-left' }, `${rawObject.audit_category}`),
        ])
      ])
    },
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-left",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    accessorKey: 'severity',
    header: 'Severity',
        cell: (info: { getValue: () => any; }) => info.getValue(),
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center w-[120%]",
        td: "text-black font-medium text-center",
      },
    },
  },
  {
    id: 'actions',
    header: 'Action',
    meta: {
      class: {
        th: "font-semibold bg-primary text-secondary-900 text-center",
        td: "text-black font-medium text-center",
      },
    },
    cell: (row : any ) => {
       return h(UButton, {
            icon: 'chevronRight',
            color: 'primary',
            size: 'xl',
            variant: 'ghost',
            'aria-label': 'Actions dropdown'
          })
    }
  }
]

const getRiskLevel = (impact: number, probability: number): string => {
  const score = impact * probability;
  if (score <= 3) return "Low";
  if (score <= 6) return "Low to Moderate";
  if (score <= 9) return "Moderate";
  if (score <= 12) return "Moderate to High";
  return "High";
};

const getHeatMapCellColor = (x: number, y: number): string => {
  // y is inverted (5 at top = probability 5, 1 at bottom = probability 1)
  const probability = 6 - y;
  const impact = x;
  const riskLevel = getRiskLevel(impact, probability);

  const colorMap: Record<string, string> = {
    "Low": "bg-success-400 hover:bg-success-500 cursor-pointer transition-colors",
    "Low to Moderate": "bg-success-600 hover:bg-success-700 cursor-pointer transition-colors",
    "Moderate": "bg-warning-500 hover:bg-warning-600 cursor-pointer transition-colors",
    "Moderate to High": "bg-primary-400 hover:bg-primary-500 cursor-pointer transition-colors",
    "High": "bg-error-500 hover:bg-error-600 cursor-pointer transition-colors",
  };

  return colorMap[riskLevel] || "bg-gray-200";
};
</script>
