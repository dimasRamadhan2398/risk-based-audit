<template>
  <UModal v-model:open="store.isFormOpen" :dismissible="false" :ui="{ content: 'sm:max-w-2xl' }">
    
    <div></div>

    <template #content>
      <UForm :state="store.form" @submit.prevent="() => store.handleSubmit(props.currentRiskId)">
        <div class="relative bg-white dark:bg-gray-900 rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 rounded-t-xl flex justify-between items-center">
            <h3 class="text-lg font-bold text-gray-800 dark:text-white">
              {{ store.isEditing ? 'Edit Rencana Mitigasi' : 'Tambah Rencana Mitigasi' }}
            </h3>
            <UIcon name="close" @click="store.closeForm" class="text-gray-400 hover:text-gray-600 text-2xl cursor-pointer">&times;</UIcon>
          </div>

          <div class="p-6 overflow-y-auto space-y-5">
            
            <UFormField label="Aktivitas Mitigasi" required>
              <UInput v-model="store.form.actionPlan" placeholder="Jelaskan langkah mitigasi secara detail..." class="w-full" required />
            </UFormField>

            <UFormField label="Supervisor" required>
              <UInput v-model="store.form.supervisor" placeholder="Masukkan nama supervisor..." class="w-full" required />
            </UFormField>
            
            <UFormField label="PIC" required>
              <UInput v-model="store.form.pic" placeholder="Masukkan nama PIC..." class="w-full" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <UFormField label="Tanggal Mulai" required>
                    <UInput v-model="store.form.start_date" type="date" icon="i-heroicons-calendar" required />
                </UFormField>

                <UFormField label="Tanggal Berakhir" required>
                    <UInput v-model="store.form.end_date" type="date" icon="i-heroicons-calendar-days" required />
                </UFormField>
            </div>

            <!-- <UFormField label="Status" required>
              <USelectMenu v-model="store.form.status" :items="store.statusOptions" placeholder="Pilih Status" class="input-field bg-white w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500" required />
            </UFormField> -->

            <UFormField label="Keterangan / Catatan Tambahan">
              <UTextarea v-model="store.form.notes" :rows="2" placeholder="Catatan opsional..." class="w-full" />
            </UFormField>

          </div>
      
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 rounded-b-xl flex justify-end gap-3">
            <UButton label="Batal" color="neutral" variant="outline" @click="store.closeForm" />
            <UButton :label="store.isEditing ? 'Simpan Perubahan' : 'Simpan Mitigasi'" color="primary" type="submit" />
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