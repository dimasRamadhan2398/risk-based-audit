<template>
    <!-- Add/Edit Modal -->
    <UModal
      v-model:open="store.isAddModalOpen"
      title="Strategic Plan Information"
      :ui="{
        content: 'bg-white sm:max-w-xl',
        title: 'text-gray-900 font-semibold text-2xl',
        body: 'text-gray-900',
      }"
    >
      <template #body>
        <UForm :state="store.form" @submit.prevent="store.handleSubmit">
          <div class="space-y-5">

            <!-- Objective ID -->
            <div class="form-row">
              <label class="form-label">
                Objective ID <span class="text-orange-500">*</span>
              </label>
              <UInput
                v-model="store.form.code"
                placeholder="Ex: SO-IA01"
                class="w-full"
              />
            </div>

            <!-- Strategic Objective -->
            <div class="form-row">
              <label class="form-label">
                Strategic Objective <span class="text-orange-500">*</span>
              </label>
              <UTextarea
                v-model="store.form.strategicObjective"
                placeholder="Ex: Enhance Operational Efficiency"
                :rows="2"
                class="w-full"
              />
            </div>

            <!-- KPI Title -->
            <div class="form-row">
              <label class="form-label">
                KPI <span class="text-orange-500">*</span>
              </label>
              <UInput
                v-model="store.form.kpi"
                placeholder="Ex: Revenue Operational Cost"
                class="w-full"
              />
            </div>

            <!-- Unit -->
            <div class="form-row">
              <label class="form-label">Unit</label>
              <USelectMenu
                v-model="store.form.unit"
                :items="store.unitOptions"
                value-key="value"
                placeholder="Select Unit"
                class="w-full"
              />
            </div>

            <!-- HIB/HIG Radio -->
            <div class="form-row">
              <label class="form-label">
                HIB/HIG <span class="text-orange-500">*</span>
              </label>
              <div class="flex flex-col gap-2">
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.hibHig"
                    value="HIG"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm text-gray-800">HIG (High is Good)</span>
                </label>
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.hibHig"
                    value="HIB"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm text-gray-800">HIB (High is Bad)</span>
                </label>
              </div>
            </div>

            <!-- Period Type Radio -->
            <div class="form-row">
              <label class="form-label">
                Period Type <span class="text-orange-500">*</span>
              </label>
              <div class="flex flex-col gap-2">
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.periodType"
                    value="Quartal"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm text-gray-800">Quartal</span>
                </label>
                <label class="inline-flex items-center gap-2 cursor-pointer">
                  <input
                    type="radio"
                    v-model="store.form.periodType"
                    value="Yearly"
                    class="accent-orange-500 w-4 h-4"
                  />
                  <span class="text-sm text-gray-800">Yearly</span>
                </label>
              </div>
            </div>

            <!-- Yearly: Dari Tahun / Sampai Tahun -->
            <div v-if="store.form.periodType === 'Yearly'" class="form-row">
              <label class="form-label">Period</label>
              <div class="flex items-center gap-3 flex-wrap">
                <span class="text-sm font-semibold text-gray-800">From</span>
                <USelectMenu
                  v-model="store.form.yearStart"
                  :items="store.yearOptions"
                  value-key="value"
                  placeholder="Year"
                  class="w-28"
                />
                <span class="text-sm font-semibold text-gray-800">To</span>
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
              <label class="form-label"></label>
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

            <!-- Actual -->
            <div class="form-row">
              <label class="form-label">
                Actual <span class="text-orange-500">*</span>
              </label>
              <UInput
                v-model="store.form.actual"
                placeholder="Ex: 100"
                class="w-full"
              />
            </div>

            <!-- Target -->
            <div class="form-row">
              <label class="form-label">
                Target <span class="text-orange-500">*</span>
              </label>
              <UInput
                v-model="store.form.target"
                placeholder="Ex: 300"
                class="w-full"
              />
            </div>

            <!-- Hitungan (Read Only) -->
            <div class="form-row">
              <label class="form-label">Calculation</label>
              <div class="readonly-field">
                {{ store.computedCalculation || '-' }}
              </div>
            </div>

            <!-- Status (Read Only) -->
            <div class="form-row">
              <label class="form-label">Status</label>
              <div class="readonly-field">
                {{ store.computedStatus || '-' }}
              </div>
            </div>

          </div>
        </UForm>
      </template>
      <template #footer>
        <div class="flex flex-row justify-end gap-3">
          <UButton
            label="Cancel"
            variant="ghost"
            color="neutral"
            @click="store.closeModal"
          />
          <UButton
            :label="store.isEditMode ? 'Update' : 'Submit'"
            variant="solid"
            class="bg-orange-500 hover:bg-orange-600 text-white"
            @click="store.handleSubmit"
          />
        </div>
      </template>
    </UModal>
</template>

<script setup lang="ts">
import { useStrategicPlanStore } from '~/stores/strategic-audit-plan'

const store = useStrategicPlanStore()
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
  color: #374151;
  padding-top: 0.5rem;
}

.readonly-field {
  background-color: #f3f4f6;
  border-radius: 0.375rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: #4b5563;
  min-height: 2.25rem;
  display: flex;
  align-items: center;
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

.period-tab-inactive {
  color: #6b7280;
  background-color: white;
}

.period-tab-inactive:hover {
  background-color: #f9fafb;
}
</style>