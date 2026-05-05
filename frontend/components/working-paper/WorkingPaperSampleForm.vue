<template>
    
      <UModal v-model:open="store.showModalF03" :dismissible="false" class="w-full sm:max-w-4xl">
        <div></div>

        <template #content>
        <UForm :state="store.sampleForm" @submit.prevent="store.handleSubmitF03">
        <div class="bg-[var(--bg-main)] rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto border border-[var(--border-main)] transition-colors duration-300">
        <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
            <h3 class="text-lg font-bold text-[var(--text-main)]">Samples</h3>
            <UIcon name="close" @click="store.closeModalF03" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer"></UIcon>
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
        
        
            <div v-for="(sampel, index) in store.sampleForm.samples" :key="sampel.id" class="border border-gray-200  rounded-xl p-6 relative">
            <div class="mb-4 flex justify-between items-center">
                <h3 class="text-lg font-bold">Sampel {{ index + 1 }}</h3>
                <UIcon name="i-heroicons-trash" color="error" variant="ghost" @click="store.removeSample(index)" />                
            </div>
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Nama Dokumen" class="font-semibold text-sm" />
                <UInput class="md:col-span-3" v-model="sampel.document" placeholder="Ex: PO-2026-001" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Langkah 1" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l1" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Langkah 2" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l2" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Langkah 3" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l3" :items="store.options.testResult" placeholder="Pilih Langkah" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center mt-2">
                <UFormField label="Status" class="font-semibold text-sm" />
                <div class="md:col-span-3 flex items-center gap-2">
                    <div class="w-4 h-4 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                    <span class="font-bold">{{ store.checkSampleStatus(sampel) ? 'Efektif' : 'Tidak Efektif' }}</span>
                </div>
                </div>
            </div>

            </div>
            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Sampel" @click="store.addSample" />
        

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full">
            <UFormField label="Kesimpulan" class="font-semibold text-sm" />
            <UTextarea class="md:col-span-3" v-model="store.sampleForm.conclusion" :rows="4" placeholder="Ketik kontrol pengamanan / SOP yang sedang dievaluasi di lapangan..." />
        </div>

        <div class="flex justify-end p-6 border-gray-100">
            <UButton 
                :label="store.isEditingF03 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF03" 
            />
        </div>
        </div>
        </div>
    </UForm>
    </template>
    </UModal>
    

</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>