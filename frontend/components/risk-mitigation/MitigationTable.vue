<template>
  <div class="space-y-6">
    <!-- Title Card -->
    <div class="flex justify-between items-center bg-white dark:bg-gray-900 p-4 rounded-xl border border-gray-200 dark:border-gray-800 shadow-md">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-primary-50 dark:bg-primary-950 flex items-center justify-center">
          <UIcon name="i-heroicons-shield-check" class="text-primary-600 dark:text-primary-400 text-xl font-bold" />
        </div>
        <div>
          <h3 class="text-md font-extrabold text-gray-800 dark:text-gray-200">Risk Mitigation Plans & Controls</h3>
          <p class="text-md text-gray-400">Expand a row to monitor weekly/monthly checks and update realization notes.</p>
        </div>
      </div>
      <UButton 
        label="Add Mitigation Plan" 
        icon="i-heroicons-plus" 
        color="primary" 
        class="font-black shadow-md shadow-primary/20"
        @click="store.openForm()"
      />
    </div>

    <!-- Mitigation List Table -->
    <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-md">
          <thead>
            <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800 font-bold uppercase tracking-wider text-[10px] text-gray-400 text-center">
              <th class="py-4 px-4 text-left w-12">No</th>
              <th class="py-4 px-4 text-left min-w-[280px]">Rencana Mitigasi & Kontrol</th>
              <th class="py-4 px-4 text-left">Timeline</th>
              <th class="py-4 px-4 text-left">Unit in Charge</th>
              <th class="py-4 px-4 text-left">Person in Charge</th>
              <th class="py-4 px-4 text-left">Supervisi</th>
              <th class="py-4 px-4">Monitoring Progress</th>
              <th class="py-4 px-4">Actions</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="filteredMitigations.length > 0">
              <template v-for="(row, idx) in filteredMitigations" :key="row.id">
                <!-- Parent Row -->
                <tr 
                  class="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50/50 dark:hover:bg-gray-800/20 cursor-pointer transition-colors duration-200"
                  @click="toggleRow(row.id)"
                >
                  <!-- No -->
                  <td class="py-4 px-4 font-bold text-gray-400">{{ idx + 1 }}</td>
                  
                  <!-- Plan -->
                  <td class="py-4 px-4 font-bold text-gray-800 dark:text-gray-200 leading-relaxed max-w-[320px] whitespace-normal">
                    <div class="flex items-start gap-2">
                      <UIcon 
                        :name="expandedRowId === row.id ? 'i-heroicons-chevron-down' : 'i-heroicons-chevron-right'" 
                        class="text-gray-400 text-sm shrink-0 mt-1" 
                      />
                      <div>
                        <span class="inline-block px-2 py-0.5 bg-primary-50 dark:bg-primary-950 text-primary-700 dark:text-primary-300 font-bold rounded text-[10px] mb-1">
                          {{ row.riskControlId || ('CTL-RSK-' + (idx + 1)) }}
                        </span>
                        <p class="font-bold text-gray-800 dark:text-gray-200">{{ row.mitigationPlan }}</p>
                      </div>
                    </div>
                  </td>
                  
                  <!-- Timeline -->
                  <td class="py-4 px-4 whitespace-nowrap">
                    <div class="flex flex-col">
                      <span class="font-bold text-gray-700 dark:text-gray-300">
                        {{ formatDate(row.start_date) }} - {{ formatDate(row.end_date) }}
                      </span>
                      <span class="text-[9px] uppercase font-black text-primary-500 mt-0.5">
                        {{ getMonitoringUnit(row) }} Monitoring
                      </span>
                    </div>
                  </td>
                  
                  <!-- Unit -->
                  <td class="py-4 px-4 text-gray-600 dark:text-gray-400 font-medium">{{ row.unitInCharge }}</td>
                  
                  <!-- PIC -->
                  <td class="py-4 px-4 text-gray-600 dark:text-gray-400 font-medium">{{ row.pic }}</td>
                  
                  <!-- Supervisor -->
                  <td class="py-4 px-4 text-gray-600 dark:text-gray-400 font-medium">{{ row.supervisor }}</td>
                  
                  <!-- Progress Bar -->
                  <td class="py-4 px-4 text-center whitespace-nowrap" @click.stop>
                    <div class="flex flex-col items-center justify-center gap-1.5 min-w-[120px]">
                      <div class="flex justify-between items-center w-full text-[10px] font-bold text-gray-500">
                        <span>{{ getActualCount(row) }} / {{ getTargetCount(row) }} {{ getMonitoringUnit(row) }}</span>
                        <span class="text-primary-500">{{ getProgressPercent(row) }}%</span>
                      </div>
                      <div class="w-full h-1.5 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
                        <div 
                          class="h-full bg-gradient-to-r from-success-500 to-primary-500 transition-all duration-300"
                          :style="{ width: `${getProgressPercent(row)}%` }"
                        ></div>
                      </div>
                    </div>
                  </td>
                  
                  <!-- Actions -->
                  <td class="py-4 px-4 text-center" @click.stop>
                    <div class="flex gap-2 justify-center">
                      <UButton 
                        icon="i-heroicons-pencil-square" 
                        size="sm" 
                        color="warning" 
                        variant="ghost" 
                        @click="store.openForm(row)" 
                      />
                      <UButton 
                        icon="i-heroicons-trash" 
                        size="sm" 
                        color="error" 
                        variant="ghost" 
                        @click="store.deleteMitigation(row.id, props.currentRiskId)" 
                      />
                    </div>
                  </td>
                </tr>

                <!-- Expanded Collapsible Monitoring Panel -->
                <tr v-if="expandedRowId === row.id">
                  <td colspan="8" class="bg-gray-50/50 dark:bg-gray-900/30 p-6 border-t border-b border-gray-100 dark:border-gray-800/50">
                    <div class="space-y-4">
                      <!-- Collapsible Panel Header -->
                      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                        <div>
                          <h4 class="text-md font-black uppercase tracking-wider text-gray-600 dark:text-gray-400 flex items-center gap-1.5">
                            <UIcon name="i-heroicons-clipboard-document-check" />
                            Monitoring Control Checklist
                          </h4>
                          <p class="text-[10px] text-gray-400 mt-0.5">
                            Realization progress calculation is based on current date (today).
                          </p>
                        </div>
                        
                        <!-- Progress Summary Stats -->
                        <div class="flex items-center gap-4 bg-white dark:bg-gray-800 px-4 py-2 rounded-lg border border-gray-100 dark:border-gray-700 shadow-sm shrink-0">
                          <div class="text-[10px]">
                            <span class="block text-gray-400 font-black uppercase tracking-widest text-[8px] leading-none mb-1">Realization</span>
                            <span class="font-black text-md text-gray-700 dark:text-gray-200">
                              {{ getActualCount(row) }} / {{ getTargetCount(row) }} {{ getMonitoringUnit(row) }} Monitored
                            </span>
                          </div>
                          <div class="w-24 h-2 rounded-full bg-gray-100 dark:bg-gray-700 overflow-hidden">
                            <div 
                              class="h-full bg-gradient-to-r from-success-500 to-primary-500" 
                              :style="{ width: `${getProgressPercent(row)}%` }"
                            ></div>
                          </div>
                          <span class="font-black text-md text-primary-500">{{ getProgressPercent(row) }}%</span>
                        </div>
                      </div>

                      <!-- Checklist Grid -->
                      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pt-2">
                        <div 
                          v-for="check in row.monitoring" 
                          :key="check.id"
                          class="p-3 bg-white dark:bg-gray-800 rounded-xl border flex flex-col justify-between gap-3 shadow-2md hover:shadow-md transition-shadow duration-200"
                          :class="[
                            check.checked ? 'border-success-500/20 bg-success-500/5' : 'border-gray-200 dark:border-gray-700',
                            isOverdue(check) ? 'border-error-500/20 bg-error-500/5' : ''
                          ]"
                        >
                          <div class="flex items-start justify-between gap-3">
                            <UCheckbox 
                              v-model="check.checked" 
                              :label="check.label"
                              class="font-bold text-md text-gray-700 dark:text-gray-200"
                              color="success"
                            />
                            <UBadge 
                              :color="getCheckStatusColor(check)"
                              variant="subtle"
                              size="md"
                              class="font-black uppercase tracking-wider text-[8px]"
                            >
                              {{ getCheckStatus(check) }}
                            </UBadge>
                          </div>
                          
                          <!-- Notes input -->
                          <div class="space-y-1">
                            <UInput 
                              v-model="check.notes" 
                              placeholder="Tambah catatan monitoring..." 
                              size="sm" 
                              class="w-full text-md" 
                              icon="i-heroicons-chat-bubble-bottom-center-text"
                            />
                          </div>
                        </div>
                      </div>

                      <!-- Save button -->
                      <div class="flex justify-end gap-3 pt-2">
                        <UButton 
                          label="Save Monitoring Controls" 
                          icon="i-heroicons-check-circle"
                          color="success"
                          size="md"
                          class="font-black shadow-md shadow-success/20"
                          @click="saveMonitoring(row.id, row.monitoring || [])"
                        />
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
            <template v-else>
              <tr>
                <td colspan="8" class="py-12 text-center text-gray-400 dark:text-gray-600">
                  <div class="flex flex-col items-center justify-center gap-2">
                    <UIcon name="i-heroicons-inbox" class="text-4xl" />
                    <p class="font-bold">No risk mitigation plans found for this risk.</p>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watchEffect } from 'vue'
import { useMitigationStore } from '~/stores/mitigation-risk'

const store = useMitigationStore()
const toast = useToast()

const props = defineProps<{
  currentRiskId: string
}>()

watchEffect(() => {
  if (props.currentRiskId) {
    store.fetchMitigations(props.currentRiskId)
  }
})

const filteredMitigations = computed(() => {
  return store.getMitigationsByRiskId(props.currentRiskId)
})

// Expanded row state
const expandedRowId = ref<string | null>(null)

function toggleRow(id: string) {
  if (expandedRowId.value === id) {
    expandedRowId.value = null
  } else {
    expandedRowId.value = id
  }
}

// Helpers for dates & progress
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

const getMonitoringUnit = (row: any) => {
  const start = new Date(row.start_date)
  const end = new Date(row.end_date)
  const diffDays = Math.ceil(Math.abs(end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
  return diffDays < 60 ? 'Weeks' : 'Months'
}

const getTargetCount = (row: any) => {
  const now = new Date()
  if (!row.monitoring) return 0
  return row.monitoring.filter((check: any) => new Date(check.startDate) <= now).length
}

const getActualCount = (row: any) => {
  if (!row.monitoring) return 0
  return row.monitoring.filter((check: any) => check.checked).length
}

const getProgressPercent = (row: any) => {
  const target = getTargetCount(row)
  if (target === 0) return 0
  const actual = getActualCount(row)
  return Math.min(100, Math.round((actual / target) * 100))
}

const getCheckStatus = (check: any) => {
  const now = new Date()
  if (check.checked) return 'Completed'
  if (new Date(check.endDate) < now) return 'Overdue'
  return 'Scheduled'
}

const getCheckStatusColor = (check: any) => {
  const now = new Date()
  if (check.checked) return 'success'
  if (new Date(check.endDate) < now) return 'error'
  return 'primary'
}

const isOverdue = (check: any) => {
  const now = new Date()
  return !check.checked && new Date(check.endDate) < now
}

async function saveMonitoring(id: string, monitoring: any[]) {
  try {
    await store.updateMitigationMonitoring(id, monitoring, props.currentRiskId)
    toast.add({
      title: 'Success',
      description: 'Monitoring controls updated successfully.',
      color: 'success',
      icon: 'i-heroicons-check-circle'
    })
  } catch (err) {
    console.error('Failed to save monitoring:', err)
    toast.add({
      title: 'Error',
      description: 'Failed to update monitoring controls.',
      color: 'error',
      icon: 'i-heroicons-exclamation-triangle'
    })
  }
}
</script>