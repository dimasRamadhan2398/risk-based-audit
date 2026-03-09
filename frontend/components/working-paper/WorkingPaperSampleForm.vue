<template>
    <Teleport to="body">
      <div v-if="store.showModalF03" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-secondary-50 dark:bg-secondary-300 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <UForm @submit.prevent="store.handleSubmitF03">

        <div class="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 dark:text-white">Samples</h3>
            <UIcon name="close" @click="store.closeModalF03" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>
        <div class="space-y-6 m-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-full">
            <UFormField label="Total Populasi" required>
            <UInput type="number" v-model="store.sampleForm.population" placeholder="Ex: 100" class="w-full"/>
            </UFormField>
            <UFormField label="Jumlah Sampel yang Diuji" required>
            <UInput type="number" v-model="store.sampleForm.sampleSize" placeholder="Ex: 10" class="w-full"/>
            </UFormField>
        </div>
        
        
            <div v-for="(sampel, index) in store.sampleForm.samples" :key="sampel.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 relative bg-white dark:bg-gray-800">
            <h3 class="text-lg font-bold mb-4">Sampel {{ index + 1 }}</h3>
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Nama Dokumen <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="sampel.document" placeholder="Ex: PO-2026-001" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 1 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l1" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 2 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l2" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Langkah 3 <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="sampel.l3" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center mt-2">
                <label class="font-semibold text-sm">Status</label>
                <div class="md:col-span-3 flex items-center gap-2">
                    <div class="w-4 h-4 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                    <span class="font-bold">{{ store.checkSampleStatus(sampel) ? 'Efektif' : 'Tidak Efektif' }}</span>
                </div>
                </div>
            </div>

            <div class="mt-4 flex justify-end">
                <UButton label="Hapus Sampel" color="error" variant="ghost" class="font-bold" @click="store.removeSample(index)" />
            </div>
            </div>
            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Sampel" @click="store.addSample" />
        

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full">
            <label class="font-semibold text-sm mt-2">Kesimpulan</label>
            <UTextarea class="md:col-span-3" v-model="store.sampleForm.conclusion" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
        </div>

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton 
                :label="store.isEditingF03 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF03" 
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