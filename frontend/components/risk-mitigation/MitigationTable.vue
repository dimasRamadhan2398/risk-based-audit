<template>
  <div class="space-y-6">
    <!-- Title Card -->
    <div class="flex flex-col sm:flex-row justify-between items-center bg-white dark:bg-gray-900 p-4 rounded-xl border border-gray-200 dark:border-gray-800 shadow-md gap-4">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-lg bg-primary-50 dark:bg-primary-950 flex items-center justify-center">
          <UIcon name="i-heroicons-shield-check" class="text-primary-600 dark:text-primary-400 text-xl font-bold" />
        </div>
        <div>
          <h3 class="text-md font-extrabold text-gray-800 dark:text-gray-200">{{ t('riskMitigation.listTitle') }}</h3>
          <p class="text-md text-gray-400">{{ t('riskMitigation.listSubtitle') }}</p>
        </div>
      </div>
      <UButton
        :label="t('riskMitigation.add')"
        icon="i-heroicons-plus"
        color="primary"
        class="font-black shadow-md shadow-primary/20 w-full sm:w-auto justify-center text-center"
        @click="store.openForm()"
      />
    </div>

    <!-- Mitigation List Table -->
    <!-- Desktop View (>= md): Full Table -->
    <div class="hidden md:block bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-md">
          <thead>
            <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800 font-bold uppercase tracking-wider text-[10px] text-gray-400 text-center">
              <th class="py-4 px-4 text-left w-12">{{ t('riskMitigation.columns.number') }}</th>
              <th class="py-4 px-4 text-left min-w-[280px]">{{ t('riskMitigation.columns.plan') }}</th>
              <th class="py-4 px-4 text-left">{{ t('riskMitigation.columns.timeline') }}</th>
              <th class="py-4 px-4 text-left">{{ t('riskMitigation.columns.unit') }}</th>
              <th class="py-4 px-4 text-left">{{ t('riskMitigation.columns.pic') }}</th>
              <th class="py-4 px-4 text-left">{{ t('riskMitigation.columns.supervisor') }}</th>
              <th class="py-4 px-4">{{ t('riskMitigation.columns.progress') }}</th>
              <th class="py-4 px-4">{{ t('riskMitigation.columns.actions') }}</th>
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
                  <td class="py-4 px-4 w-[320px] min-w-[280px] max-w-[320px] align-top font-bold text-gray-800 dark:text-gray-200 leading-relaxed">
                    <div class="flex items-start gap-2 min-w-0">
                      <UIcon :name="expandedRowId === row.id ? 'i-heroicons-chevron-down' : 'i-heroicons-chevron-right'" class="text-gray-400 text-sm shrink-0 mt-1"/>

                      <div class="min-w-0 flex-1">
                        <span class="inline-block max-w-full px-2 py-0.5 mb-1 bg-primary-50 dark:bg-primary-950 text-primary-700 dark:text-primary-300 font-bold rounded text-[10px] wrap-anywhere">
                          {{ row.riskControlId || ('CTL-RSK-' + (idx + 1)) }}
                        </span>

                        <p class="font-bold text-gray-800 dark:text-gray-200 whitespace-normal wrap-anywhere">{{ row.mitigationPlan }}</p>
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
                        {{ t('riskMitigation.monitoring.label', { unit: getMonitoringUnit(row) }) }}
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
                            {{ t('riskMitigation.monitoring.checklistTitle') }}
                          </h4>
                          <p class="text-[10px] text-gray-400 mt-0.5">
                            {{ t('riskMitigation.monitoring.checklistDesc') }}
                          </p>
                        </div>

                        <!-- Progress Summary Stats -->
                        <div class="flex items-center gap-4 bg-white dark:bg-gray-800 px-4 py-2 rounded-lg border border-gray-100 dark:border-gray-700 shadow-sm shrink-0">
                          <div class="text-[10px]">
                            <span class="block text-gray-400 font-black uppercase tracking-widest text-[8px] leading-none mb-1">{{ t('riskMitigation.monitoring.realization') }}</span>
                            <span class="font-black text-md text-gray-700 dark:text-gray-200">
                              {{ t('riskMitigation.monitoring.monitored', { actual: getActualCount(row), target: getTargetCount(row), unit: getMonitoringUnit(row) }) }}
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
                            isOverdue(check) ? 'border-error-500/20 bg-error-500/5' : '',
                            !isCheckable(check) ? 'opacity-50 pointer-events-none' : ''
                          ]"
                        >
                          <div class="flex items-start justify-between gap-3">
                            <UCheckbox
                              v-model="check.checked"
                              :label="check.label"
                              class="font-bold text-md text-gray-700 dark:text-gray-200"
                              color="success"
                              :disabled="!isCheckable(check)"
                              @change="saveMonitoring(row.id, row.monitoring || [])"
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
                              :placeholder="t('riskMitigation.monitoring.notesPlaceholder')"
                              size="sm"
                              class="w-full text-md"
                              icon="i-heroicons-chat-bubble-bottom-center-text"
                              :disabled="!isCheckable(check)"
                              @change="saveMonitoring(row.id, row.monitoring || [])"
                            />
                          </div>
                        </div>
                      </div>

                      <!-- Save button removed for auto-save functionality -->
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
                    <p class="font-bold">{{ t('riskMitigation.empty') }}</p>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Mobile & Tablet View (< md): Stacked Cards List with Expandable Monitoring -->
    <div class="block md:hidden">
      <!-- Empty State -->
      <div
        v-if="!filteredMitigations.length"
        class="py-8 px-4 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-white dark:bg-gray-900"
      >
        <UIcon name="i-heroicons-inbox" class="size-8 mx-auto text-gray-400 mb-2" />
        <p class="text-xs font-semibold text-gray-500 dark:text-gray-400">{{ t('riskMitigation.empty') }}</p>
      </div>

      <!-- Cards List -->
      <div v-else class="space-y-3">
        <div
          v-for="(row, idx) in filteredMitigations"
          :key="row.id"
          class="rounded-xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 shadow-xs overflow-hidden transition-all duration-200"
          :class="{ 'ring-2 ring-primary-500/20 border-primary-500/40': expandedRowId === row.id }"
        >
          <!-- Card Header & Main Info -->
          <div class="p-4 space-y-3">
            <!-- Header Row: Index, Control ID, Actions -->
            <div class="flex items-center justify-between gap-2">
              <div class="flex items-center gap-1.5 flex-wrap">
                <span class="font-bold text-xs text-gray-400">#{{ idx + 1 }}</span>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 bg-primary-50 dark:bg-primary-950 text-primary-700 dark:text-primary-300 font-bold rounded text-[10px] border border-primary-200/50 dark:border-primary-800/40">
                  <UIcon name="i-heroicons-shield-check" class="size-3" />
                  {{ row.riskControlId || ('CTL-RSK-' + (idx + 1)) }}
                </span>
              </div>

              <!-- Action buttons -->
              <div class="flex items-center gap-1">
                <UButton
                  icon="i-heroicons-pencil-square"
                  size="xs"
                  color="warning"
                  variant="soft"
                  @click.stop="store.openForm(row)"
                />
                <UButton
                  icon="i-heroicons-trash"
                  size="xs"
                  color="error"
                  variant="soft"
                  @click.stop="store.deleteMitigation(row.id, props.currentRiskId)"
                />
              </div>
            </div>

            <!-- Mitigation Plan Description -->
            <div>
              <p class="text-xs text-gray-400 font-semibold uppercase tracking-wider mb-0.5">
                {{ t('riskMitigation.columns.plan') }}
              </p>
              <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
                {{ row.mitigationPlan }}
              </h4>
            </div>

            <!-- Timeline -->
            <div class="flex items-center justify-between gap-2 text-xs py-2 px-3 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-800">
              <div class="flex items-center gap-1.5 text-gray-700 dark:text-gray-300 font-semibold min-w-0">
                <UIcon name="i-heroicons-calendar" class="size-3.5 text-primary-500 shrink-0" />
                <span class="truncate">{{ formatDate(row.start_date) }} - {{ formatDate(row.end_date) }}</span>
              </div>
              <span class="text-[9px] uppercase font-black text-primary-600 dark:text-primary-400 shrink-0">
                {{ t('riskMitigation.monitoring.label', { unit: getMonitoringUnit(row) }) }}
              </span>
            </div>

            <!-- Metadata Details: Unit, PIC, Supervisor -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2 text-xs pt-1 text-gray-600 dark:text-gray-400 border-t border-gray-100 dark:border-gray-800">
              <div class="flex items-center gap-1.5 min-w-0">
                <UIcon name="i-heroicons-building-office" class="size-3.5 text-gray-400 shrink-0" />
                <span class="text-[11px] truncate"><strong class="font-bold text-gray-700 dark:text-gray-300">{{ t('riskMitigation.columns.unit') }}:</strong> {{ row.unitInCharge }}</span>
              </div>
              <div class="flex items-center gap-1.5 min-w-0">
                <UIcon name="i-heroicons-user" class="size-3.5 text-gray-400 shrink-0" />
                <span class="text-[11px] truncate"><strong class="font-bold text-gray-700 dark:text-gray-300">{{ t('riskMitigation.columns.pic') }}:</strong> {{ row.pic }}</span>
              </div>
              <div class="flex items-center gap-1.5 min-w-0">
                <UIcon name="i-heroicons-user-circle" class="size-3.5 text-gray-400 shrink-0" />
                <span class="text-[11px] truncate"><strong class="font-bold text-gray-700 dark:text-gray-300">{{ t('riskMitigation.columns.supervisor') }}:</strong> {{ row.supervisor }}</span>
              </div>
            </div>

            <!-- Progress Bar Section -->
            <div class="space-y-1.5 pt-2 border-t border-gray-100 dark:border-gray-800">
              <div class="flex justify-between items-center text-xs font-bold">
                <span class="text-gray-500 dark:text-gray-400 text-[11px]">
                  {{ getActualCount(row) }} / {{ getTargetCount(row) }} {{ getMonitoringUnit(row) }}
                </span>
                <span class="text-primary-500 font-black">{{ getProgressPercent(row) }}%</span>
              </div>
              <div class="w-full h-2 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-success-500 to-primary-500 transition-all duration-300"
                  :style="{ width: `${getProgressPercent(row)}%` }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Expand / Collapse Toggle Bar -->
          <button
            type="button"
            class="w-full py-2.5 px-4 bg-gray-50 dark:bg-gray-800/60 hover:bg-gray-100 dark:hover:bg-gray-800 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs font-bold text-gray-700 dark:text-gray-200 transition-colors cursor-pointer"
            @click="toggleRow(row.id)"
          >
            <span class="flex items-center gap-1.5 text-primary-600 dark:text-primary-400">
              <UIcon name="i-heroicons-clipboard-document-check" class="size-4" />
              {{ t('riskMitigation.monitoring.checklistTitle') }}
            </span>
            <div class="flex items-center gap-1 text-[11px] text-gray-500">
              <span>{{ expandedRowId === row.id ? (locale === 'id' ? 'Tutup Checklist' : 'Close Checklist') : (locale === 'id' ? 'Buka Checklist' : 'Open Checklist') }}</span>
              <UIcon :name="expandedRowId === row.id ? 'i-heroicons-chevron-up' : 'i-heroicons-chevron-down'" class="size-4 text-gray-500 transition-transform duration-200" />
            </div>
          </button>

          <!-- Collapsible Monitoring Checklist Panel -->
          <div v-if="expandedRowId === row.id" class="p-4 bg-gray-50/70 dark:bg-gray-950/40 border-t border-gray-100 dark:border-gray-800 space-y-4">
            <!-- Monitoring Header & Stats -->
            <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
              <div>
                <h4 class="text-xs font-black uppercase tracking-wider text-gray-700 dark:text-gray-300 flex items-center gap-1.5">
                  <UIcon name="i-heroicons-clipboard-document-check" class="text-primary-500" />
                  {{ t('riskMitigation.monitoring.checklistTitle') }}
                </h4>
                <p class="text-[10px] text-gray-400 mt-0.5">
                  {{ t('riskMitigation.monitoring.checklistDesc') }}
                </p>
              </div>

              <!-- Realization badge -->
              <div class="flex items-center gap-2.5 bg-white dark:bg-gray-800 px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-700 shadow-xs shrink-0 w-full sm:w-auto justify-between sm:justify-start">
                <span class="text-[10px] font-bold text-gray-600 dark:text-gray-300">
                  {{ t('riskMitigation.monitoring.monitored', { actual: getActualCount(row), target: getTargetCount(row), unit: getMonitoringUnit(row) }) }}
                </span>
                <span class="font-black text-xs text-primary-500">{{ getProgressPercent(row) }}%</span>
              </div>
            </div>

            <!-- Checklist Items -->
            <div class="space-y-2.5">
              <div
                v-for="check in row.monitoring"
                :key="check.id"
                class="p-3 bg-white dark:bg-gray-800 rounded-xl border flex flex-col gap-2.5 shadow-xs transition-shadow duration-200"
                :class="[
                  check.checked ? 'border-success-500/30 bg-success-500/5' : 'border-gray-200 dark:border-gray-700',
                  isOverdue(check) ? 'border-error-500/30 bg-error-500/5' : '',
                  !isCheckable(check) ? 'opacity-50 pointer-events-none' : ''
                ]"
              >
                <div class="flex items-start justify-between gap-3">
                  <UCheckbox
                    v-model="check.checked"
                    :label="check.label"
                    class="font-bold text-xs text-gray-700 dark:text-gray-200"
                    color="success"
                    :disabled="!isCheckable(check)"
                    @change="saveMonitoring(row.id, row.monitoring || [])"
                  />
                  <UBadge
                    :color="getCheckStatusColor(check)"
                    variant="subtle"
                    size="xs"
                    class="font-black uppercase tracking-wider text-[8px] shrink-0"
                  >
                    {{ getCheckStatus(check) }}
                  </UBadge>
                </div>

                <!-- Notes input -->
                <div class="space-y-1">
                  <UInput
                    v-model="check.notes"
                    :placeholder="t('riskMitigation.monitoring.notesPlaceholder')"
                    size="xs"
                    class="w-full text-xs"
                    icon="i-heroicons-chat-bubble-bottom-center-text"
                    :disabled="!isCheckable(check)"
                    @change="saveMonitoring(row.id, row.monitoring || [])"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watchEffect } from 'vue'
import { useMitigationStore } from '~/stores/mitigation-risk'

const store = useMitigationStore()
const toast = useToast()
const { t, locale } = useI18n()

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
  return date.toLocaleDateString(locale.value === 'id' ? 'id-ID' : 'en-US', { day: '2-digit', month: 'short', year: 'numeric' })
}

const getMonitoringUnit = (row: any) => {
  if (!row.monitoring || row.monitoring.length === 0) return t('riskMitigation.monitoring.weeks')
  return row.monitoring[0].id.startsWith('M') ? t('riskMitigation.monitoring.months') : t('riskMitigation.monitoring.weeks')
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
  if (check.checked) return t('riskMitigation.monitoring.completed')
  if (new Date(check.endDate) < now) return t('riskMitigation.monitoring.overdue')
  return t('riskMitigation.monitoring.scheduled')
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

const isCheckable = (check: any) => {
  const now = new Date()
  return new Date(check.startDate) <= now
}

async function saveMonitoring(id: string, monitoring: any[]) {
  try {
    await store.updateMitigationMonitoring(id, monitoring, props.currentRiskId)
    toast.add({
      title: t('riskMitigation.monitoring.updateSuccessTitle'),
      description: t('riskMitigation.monitoring.updateSuccessDesc'),
      color: 'success',
      icon: 'i-heroicons-check-circle'
    })
  } catch (err) {
    console.error('Failed to save monitoring:', err)
    toast.add({
      title: t('riskMitigation.monitoring.updateErrorTitle'),
      description: t('riskMitigation.monitoring.updateErrorDesc'),
      color: 'error',
      icon: 'i-heroicons-exclamation-triangle'
    })
  }
}
</script>
