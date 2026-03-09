<template>
    <Teleport to="body">
      <div v-if="store.showModalF05" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-secondary-50 dark:bg-secondary-300 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <UForm @submit.prevent="store.handleSubmitF05">

        <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white pt-4">Rekomendasi & Tanggapan</h2>
        
        <div class="space-y-6 m-6">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Rekomendasi Auditor (Solusi) <span class="text-red-500">*</span></label>
            <UTextarea class="md:col-span-3" v-model="store.planForm.recommendation" :rows="3" placeholder="Ex: Tim IT perlu menambahkan fitur hard-block..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Tanggapan Audite (Managemen Response)</label>
            <UTextarea class="md:col-span-3" v-model="store.planForm.response" :rows="3" placeholder="Ex: Kami setuju, update akan dilakukan di Q3..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Detail Rencana Aksi (Action Plan)</h2>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">Deskripsi Action <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.planForm.actionDescription" placeholder="Ex: Staf lupa meminta TTD Manager" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">PIC (Penanggung Jawab) <span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.planForm.pic" icon="i-heroicons-magnifying-glass" :items="store.options.pic" placeholder="Cari Nama Karyawan Ex: Dimas - IT" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <label class="font-semibold text-sm">Periode Perencanaan <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" type="date" v-model="store.planForm.periodAction" icon="i-heroicons-calendar" />
            </div>
        

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton 
                :label="store.isEditingF05 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF05" 
            />
        </div>
        </div>
    </UForm>
    </div>
    </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>