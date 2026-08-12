<template>
    <!-- Add/Edit Modal -->
    <UModal
      v-model:open="store.isAddModalOpen"
      :title="t('strategicPlan.form.title')"
      :ui="{
        content: 'sm:max-w-2xl bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl',
        header: 'border-b border-gray-100 dark:border-gray-800 pb-4 text-gray-900 dark:text-white font-bold',
        body: 'p-6 space-y-4 bg-white dark:bg-gray-900 text-gray-900 dark:text-white',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #body>
        <UForm :state="store.form" @submit.prevent="store.handleSubmit">
          <div class="space-y-5">

            <!-- Link to Goal -->
            <div class="form-row" v-if="vmgStore.activeVmg?.goals?.length">
              <label class="form-label text-gray-700 dark:text-white">
                {{ t('strategicPlan.form.corporateGoal') }} <span class="text-orange-500">*</span>
              </label>
              <USelectMenu
                v-model="store.form.goalId"
                :items="goalOptions"
                value-key="value"
                :placeholder="t('strategicPlan.form.selectCorporateGoal')"
                class="w-full"
                required
              />
            </div>


            <!-- Strategic Objective -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">
                {{ t('strategicPlan.form.objective') }} <span class="text-orange-500">*</span>
              </label>
              <UTextarea
                v-model="store.form.strategicObjective"
                :placeholder="t('strategicPlan.form.objectivePlaceholder')"
                :rows="2"
                class="w-full"
              />
            </div>

            <!-- KPI Title -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">
                {{ t('strategicPlan.form.kpi') }} <span class="text-orange-500">*</span>
              </label>
              <UInput
                v-model="store.form.kpi"
                :placeholder="t('strategicPlan.form.kpiPlaceholder')"
                class="w-full"
              />
            </div>

            <!-- Unit -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">{{ t('strategicPlan.form.unit') }}</label>
              <USelectMenu
                v-model="store.form.unit"
                :items="store.unitOptions"
                value-key="value"
                :placeholder="t('strategicPlan.form.selectUnit')"
                class="w-full"
              />
            </div>

            <!-- HIB/HIG Radio -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">
                {{ t('strategicPlan.form.hibHig') }} <span class="text-orange-500">*</span>
              </label>
              <div class="flex flex-col gap-2">
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.hibHig"
                    value="HIG"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm font-medium text-gray-800 dark:text-white">{{ t('strategicPlan.form.higLabel') }}</span>
                </label>
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.hibHig"
                    value="HIB"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm font-medium text-gray-800 dark:text-white">{{ t('strategicPlan.form.hibLabel') }}</span>
                </label>
              </div>
            </div>

            <!-- Period Type Radio -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">
                {{ t('strategicPlan.form.periodType') }} <span class="text-orange-500">*</span>
              </label>
              <div class="flex flex-col gap-2">
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.periodType"
                    value="Quartal"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm font-medium text-gray-800 dark:text-white">{{ t('strategicPlan.form.quartal') }}</span>
                </label>
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.periodType"
                    value="Yearly"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm font-medium text-gray-800 dark:text-white">{{ t('strategicPlan.form.yearly') }}</span>
                </label>
              </div>
            </div>

            <!-- Yearly: Dari Tahun / Sampai Tahun -->
            <div v-if="store.form.periodType === 'Yearly'" class="form-row">
              <label class="form-label text-gray-700 dark:text-white">{{ t('strategicPlan.form.period') }}</label>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="text-sm font-bold text-gray-800 dark:text-white">{{ t('strategicPlan.form.from') }}</span>
                <USelectMenu
                  v-model="store.form.yearStart"
                  :items="store.yearOptions"
                  value-key="value"
                  placeholder="Year"
                  class="w-28"
                />
                <span class="text-sm font-bold text-gray-800 dark:text-white">{{ t('strategicPlan.form.to') }}</span>
                <USelectMenu
                  v-model="store.form.yearEnd"
                  :items="store.yearOptions"
                  value-key="value"
                  placeholder="Year"
                  class="w-28"
                />
              </div>
            </div>

            <!-- Period Tabs (Quarter or Year) -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white"></label>
              <div class="flex gap-0 border border-gray-300 dark:border-gray-600 rounded-md overflow-hidden w-fit">
                <button
                  v-for="period in store.availablePeriods"
                  :key="period"
                  type="button"
                  class="period-tab"
                  :class="{
                    'period-tab-active': store.form.selectedPeriod === period,
                    'period-tab-inactive': store.form.selectedPeriod !== period,
                  }"
                  @click="store.form.selectedPeriod = period"
                >
                  {{ period }}
                </button>
              </div>
            </div>

            <!-- Flexible Target & Realisasi KPI Matrix Component -->
            <TargetRealizationMatrix />

            <!-- Status (Read Only) -->
            <div class="form-row">
              <label class="form-label text-gray-700 dark:text-white">{{ t('strategicPlan.form.status') }}</label>
              <div class="readonly-field">
                {{ formatStatus(store.computedStatus) }}
              </div>
            </div>

          </div>
        </UForm>
      </template>
      <template #footer>
        <div class="flex flex-row justify-end gap-3">
          <UButton
            :label="t('strategicPlan.form.cancel')"
            variant="ghost"
            color="neutral"
            class="font-semibold text-gray-700 dark:text-white"
            @click="store.closeModal"
          />
          <UButton
            :label="store.isEditMode ? t('strategicPlan.form.update') : t('strategicPlan.form.submit')"
            variant="solid"
            color="primary"
            class="font-bold"
            @click="store.handleSubmit"
          />
        </div>
      </template>
    </UModal>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'
import { useVisionMissionGoalsStore } from '~/stores/vision-mission-goals'
import { useI18n } from '~/composables/useI18n'
import TargetRealizationMatrix from './TargetRealizationMatrix.vue'

const { t } = useI18n()
const store = useStrategicPlanStore()
const vmgStore = useVisionMissionGoalsStore()

const goalOptions = computed(() => {
  if (!vmgStore.activeVmg?.goals) return []
  return vmgStore.activeVmg.goals.map(g => ({
    label: `${g.goal_code} - ${g.goal_name}`,
    value: g.id || g.goal_code
  }))
})

const formatStatus = (status: string) => {
  if (!status) return '-'
  const lower = status.toLowerCase()
  if (['good', 'moderate', 'poor', 'pending', 'planned'].includes(lower)) {
    return t(`strategicPlan.status.${lower}`)
  }
  return status
}
</script>

<style scoped>
.form-row {
  display: grid;
  grid-template-columns: 140px 1fr;
  align-items: start;
  gap: 1rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 600;
  padding-top: 0.5rem;
}

html.dark .form-label,
.dark .form-label {
  color: #ffffff !important;
}

.readonly-field {
  background-color: #f3f4f6;
  border-radius: 0.375rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: #111827;
  font-weight: 600;
  min-height: 2.25rem;
  display: flex;
  align-items: center;
}

html.dark .readonly-field,
.dark .readonly-field {
  background-color: #1f2937 !important;
  color: #ffffff !important;
  border: 1px solid #374151;
}

.period-tab {
  padding: 0.375rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.15s ease;
  cursor: pointer;
  border: none;
  outline: none;
}

.period-tab-active {
  border: 2px solid #f97316;
  color: #f97316;
  background-color: white;
}

html.dark .period-tab-active,
.dark .period-tab-active {
  background-color: #1f2937 !important;
  color: #f97316 !important;
}

.period-tab-inactive {
  color: #6b7280;
  background-color: white;
}

html.dark .period-tab-inactive,
.dark .period-tab-inactive {
  color: #ffffff !important;
  background-color: #111827 !important;
}

.period-tab-inactive:hover {
  background-color: #f9fafb;
}

html.dark .period-tab-inactive:hover,
.dark .period-tab-inactive:hover {
  background-color: #1f2937 !important;
}
</style>