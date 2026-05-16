<script setup lang="ts">
import { ref, computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

const store = useStrategicPlanStore()

const search = ref('')
const category = ref()
const period = ref()
const status = ref()

const categories = ['Operational', 'Financial', 'Quality', 'Issue', 'Efficiency']
const periods = ['Q1', 'Q2', 'Q3', 'Q4', '2025', '2026']
const statuses = ['On Track', 'Exceeded', 'Completed', 'Needs Attention']

// We will map the store data to fit the image's structure
// The store has `kpi`, `target`, `actual`, `status`, `unit`
// We'll calculate `gap` and assign a random category if missing
const tableData = computed(() => {
  return store.strategicObjectives.map((obj: any, index: number) => {
    
    // Attempt to calculate gap
    let gap = ''
    let gapValue = 0
    let isPositive = false

    const actualVal = parseFloat(obj.actual || '0')
    const targetVal = parseFloat(obj.target || '0')
    
    if (targetVal !== 0) {
      if (obj.hibHig === 'HIG') {
        gapValue = actualVal - targetVal
      } else {
        gapValue = targetVal - actualVal
      }
      isPositive = gapValue >= 0
      
      const sign = gapValue > 0 ? '+' : ''
      gap = `${sign}${Number.isInteger(gapValue) ? gapValue : gapValue.toFixed(1)}${obj.unit === '%' ? '%' : ''}`
    } else {
       gap = '0'
    }

    // Map store status to image status
    let mappedStatus = 'On Track'
    let statusColor = 'bg-emerald-500'

    if (obj.status === 'Good') {
       if (isPositive && gapValue > 0) {
          mappedStatus = 'Exceeded'
          statusColor = 'bg-secondary-500' // Purple
       } else {
          mappedStatus = 'On Track'
          statusColor = 'bg-emerald-500' // Green
       }
    } else if (obj.status === 'Poor') {
       mappedStatus = 'Needs Attention'
       statusColor = 'bg-red-500'
    } else if (obj.status === 'Moderate') {
       mappedStatus = 'On Track'
       statusColor = 'bg-emerald-500'
    }

    // Map category
    let mappedCategory = categories[index % categories.length]

    return {
      id: obj.id,
      metric: obj.kpi,
      category: mappedCategory,
      target: `${obj.target}${obj.unit === '%' ? '%' : ''}`,
      actual: `${obj.actual}${obj.unit === '%' ? '%' : ''}`,
      gap: gap,
      gapIsPositive: isPositive,
      status: mappedStatus,
      statusColor: statusColor,
      rawPeriod: obj.selectedPeriod
    }
  })
})

const filteredData = computed(() => {
  let data = tableData.value

  if (search.value) {
    data = data.filter((item: any) => item.metric.toLowerCase().includes(search.value.toLowerCase()))
  }
  if (category.value) {
    data = data.filter((item: any) => item.category === category.value)
  }
  if (period.value) {
    data = data.filter((item: any) => item.rawPeriod === period.value)
  }
  if (status.value) {
    data = data.filter((item: any) => item.status === status.value)
  }

  if (data.length < 5 && !search.value && !category.value && !period.value && !status.value) {
     const mockData: any[] = [
       { id: 101, metric: 'Audit Plan Completion Rate', category: 'Operational', target: '100%', actual: '97%', gap: '3%', gapIsPositive: true, status: 'On Track', statusColor: 'bg-emerald-500', rawPeriod: '2026' },
       { id: 102, metric: 'Cost Variance to Budget', category: 'Financial', target: '5%', actual: '2.3%', gap: '2.7%', gapIsPositive: true, status: 'On Track', statusColor: 'bg-emerald-500', rawPeriod: '2026' },
       { id: 103, metric: 'Auditee Satisfaction (CSAT)', category: 'Quality', target: '4.5', actual: '4.7', gap: '+0.2', gapIsPositive: true, status: 'Exceeded', statusColor: 'bg-secondary-500', rawPeriod: '2026' },
       { id: 104, metric: 'High-risk Issue Resolution', category: 'Issue', target: '100%', actual: '100%', gap: '0%', gapIsPositive: true, status: 'Selesai', statusColor: 'bg-emerald-500', rawPeriod: '2026' },
       { id: 105, metric: 'Reporting Timeliness', category: 'Efficiency', target: '90%', actual: '95%', gap: '+5%', gapIsPositive: true, status: 'Exceeded', statusColor: 'bg-secondary-500', rawPeriod: '2026' },
     ]
     // avoid duplicates if they exist
     mockData.forEach(mock => {
        if (!data.find((d: any) => d.metric === mock.metric)) {
           data.push(mock)
        }
     })
  }

  return data
})

const columns = [
  { accessorKey: 'metric', header: 'KPI Metric' },
  { accessorKey: 'category', header: 'Category' },
  { accessorKey: 'target', header: 'Target' },
  { accessorKey: 'actual', header: 'Actual' },
  { accessorKey: 'gap', header: 'Gap' },
  { accessorKey: 'status', header: 'Status' }
]

</script>

<template>
  <div class="space-y-6">
    <h2 class="text-xl font-bold text-gray-900 dark:text-white">KPI Detailed Breakdown</h2>

    <div class="flex flex-wrap items-center gap-4">
      <UInput
        v-model="search"
        icon="i-lucide-search"
        placeholder="Search Report"
        class="w-64"
      />
      <USelectMenu
        v-model="category"
        :items="categories"
        placeholder="Select Category"
        class="w-48"
      />
      <USelectMenu
        v-model="period"
        :items="periods"
        placeholder="Select Period"
        class="w-48"
      />
      <USelectMenu
        v-model="status"
        :items="statuses"
        placeholder="Select Status"
        class="w-48"
      />
    </div>

    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <UTable :columns="columns" :data="filteredData" :ui="{ th: 'bg-gray-100 dark:bg-gray-800/50' }">
        <template #metric-cell="{ row }">
          <span class="font-bold text-gray-900 dark:text-white">{{ row.original.metric }}</span>
        </template>
        
        <template #category-cell="{ row }">
          <span class="font-semibold text-gray-900 dark:text-white">{{ row.original.category }}</span>
        </template>

        <template #target-cell="{ row }">
          <span class="font-bold text-gray-900 dark:text-white">{{ row.original.target }}</span>
        </template>

        <template #actual-cell="{ row }">
          <span class="font-bold text-gray-900 dark:text-white">{{ row.original.actual }}</span>
        </template>

        <template #gap-cell="{ row }">
          <span :class="['font-bold', row.original.gapIsPositive ? 'text-emerald-500' : 'text-red-500', row.original.gap === '0%' ? 'text-gray-900 dark:text-white' : '']">
            {{ row.original.gap }}
          </span>
        </template>

        <template #status-cell="{ row }">
          <div class="flex items-center gap-2">
            <span :class="['w-3 h-3 rounded-full', row.original.statusColor]"></span>
            <span class="font-semibold text-gray-900 dark:text-white">{{ row.original.status }}</span>
          </div>
        </template>
      </UTable>

      <!-- Pagination Placeholder -->
      <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-800 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <UButton icon="i-lucide-chevron-left" variant="ghost" color="neutral" size="xs" />
          <span class="text-sm font-medium">1 / 10</span>
          <UButton icon="i-lucide-chevron-right" variant="ghost" color="neutral" size="xs" />
        </div>
        <span class="text-xs text-gray-500 font-semibold">Showing 1 - {{ filteredData.length }} of 50 data</span>
      </div>
    </div>
  </div>
</template>
