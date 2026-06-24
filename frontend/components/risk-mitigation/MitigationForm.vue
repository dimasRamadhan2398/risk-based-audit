<template>
  <UModal v-model:open="store.isFormOpen" :dismissible="false" :ui="{ content: 'sm:max-w-2xl' }">
    
    <div></div>

    <template #content>
      <UForm :state="store.form" @submit.prevent="() => store.handleSubmit(props.currentRiskId)">
        <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          
          <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center">
            <h3 class="text-lg font-bold text-[var(--text-main)]">
              {{ store.isEditing ? 'Edit Mitigation & Control Plan' : 'Tambah Mitigation & Control Plan' }}
            </h3>
            <UIcon name="close" @click="store.closeForm" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer">&times;</UIcon>
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            

            <UFormField label="Mitigation Plan & Controls" required>
              <UInput v-model="store.form.mitigationPlan" placeholder="Explain the mitigation plan and controls in detail..." class="w-full" required />
            </UFormField>

            <UFormField label="Supervisor" required>
              <USelectMenu 
              :items="store.supervisorOptions" 
              v-model="store.form.supervisor" 
              placeholder="Enter supervisor's name..." 
              class="w-full" 
              required />
            </UFormField>
            
            <UFormField label="PIC" required>
              <USelectMenu 
              :items="store.picOptions" 
              v-model="store.form.pic" 
              placeholder="Enter PIC's name..." 
              class="w-full" 
              required />
            </UFormField>

            <UFormField label="Unit In Charge" required>
              <USelectMenu 
              :items="store.unitInChargeOptions" 
              v-model="store.form.unitInCharge" 
              placeholder="Enter unit in charge..." 
              class="w-full" 
              required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Start Date" required>
                    <UInput v-model="store.form.start_date" type="date" icon="i-heroicons-calendar" required />
                </UFormField>

                <UFormField label="End Date" required>
                    <UInput v-model="store.form.end_date" type="date" icon="i-heroicons-calendar-days" required />
                </UFormField>
            </div>

            <UFormField label="Additional Notes">
              <UTextarea v-model="store.form.notes" :rows="2" placeholder="Optional notes..." class="w-full" />
            </UFormField>

          </div>
      
          <div class="px-6 py-4 border-t border-gray-200 bg-gray-50 rounded-b-xl flex justify-end gap-3">
            <UButton label="Cancel" color="neutral" variant="outline" @click="store.closeForm" />
            <UButton :label="store.isEditing ? 'Save Changes' : 'Save Mitigation'" color="primary" type="submit" />
          </div>
      
        </div>
      </UForm>  
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useMitigationStore } from '~/stores/mitigation-risk'

// Menerima Risk ID dari halaman induk agar mitigasi tersambung dengan risiko yang tepat
const props = defineProps<{
  currentRiskId: string
}>()

const store = useMitigationStore()
</script>