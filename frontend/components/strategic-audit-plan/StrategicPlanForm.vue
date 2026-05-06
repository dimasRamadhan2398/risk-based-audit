<template>
    
    <!-- Add/Edit Modal -->
    <UModal
      v-model:open="store.isAddModalOpen"
      :title="store.isEditMode ? 'Edit Strategic Objective' : 'Add Strategic Objective'"
      :ui="{
        content: 'bg-secondary-100',
        title: 'text-highlighted font-semibold text-2xl ',
      }"
    >
      <template #body>
        <UForm :state="store.form" @submit.prevent="store.handleSubmit">
          <div class="space-y-4">
            <UFormField label="Number" required>
              <UInput
                v-model="store.form.number"
                type="number"
                placeholder="1"
                class="w-full"
              />
            </UFormField>

            <UFormField
              label="Corporate Strategic Objectives"
              required
              help="Main organizational objectives aligned with audit strategy"
            >
              <UTextarea
                v-model="store.form.objectives"
                placeholder="e.g. Increase Revenue, Improve Efficiency, etc."
                :rows="3"
                class="w-full"
              />
            </UFormField>

            <div class="grid grid-cols-2 gap-4">
              <UFormField label="KPI" required>
                <UInput
                  v-model="store.form.kpi"
                  placeholder="Key Performance Indicator"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Characteristic Data" required>
                <USelectMenu
                  v-model="store.form.characteristicData"
                  :items="store.characteristicDataOptions"
                  value-key="value"
                  option-key="label"
                  placeholder="Select data type"
                  class="w-full"
                />
              </UFormField>
            </div>

            <UFormField
              label="KPI Target for 5 Year Period"
              required
            >
              <div class="space-y-3">
                <div v-for="(target, index) in store.form.kpiTarget" :key="index" class="flex items-center gap-2">
                  <UInput 
                    v-model="target.year" 
                    type="number" 
                    placeholder="Year" 
                    class="w-32"
                  />
                  <div class="flex-1 flex items-center gap-2">
                    <UInput 
                      v-model="target.value" 
                      placeholder="Target Value" 
                      class="flex-1"
                    />
                    <span v-if="store.form.unit" class="text-sm text-gray-500 font-medium px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700 whitespace-nowrap min-w-[3rem] text-center">
                      {{ store.form.unit }}
                    </span>
                  </div>
                  <UButton 
                    icon="i-lucide-trash-2" 
                    color="error" 
                    variant="ghost" 
                    size="sm" 
                    @click="store.removeKpiTargetYear(index)"
                  />
                </div>
                
                <UButton
                  label="Tambah Tahun"
                  icon="i-lucide-plus"
                  variant="outline"
                  size="sm"
                  block
                  class="border-dashed"
                  @click="store.addKpiTargetYear"
                />
              </div>
            </UFormField>

            <UFormField
              label="Unit"
              required
            >
              <UInput
                v-model="store.form.unit"
                placeholder="Enter unit of KPI (e.g. %, Hours, Score)"
                class="w-full"
              />
            </UFormField>
          </div>
        </UForm>
      </template>
      <template #footer>
        <div class="flex flex-row justify-end gap-3">
          <UButton
            label="Cancel"
            variant="outline"
            color="neutral"
            @click="store.closeModal"
          />
          <UButton
            label="Save Objective"
            variant="solid"
            color="primary"
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