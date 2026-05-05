<template>
    
      <UModal v-model:open="store.showModalF05" :dismissible="false" class="w-full sm:max-w-4xl">
        <div></div>

        <template #content>
        <UForm :state="store.planForm" @submit.prevent="store.handleSubmitF05">
        <div class="bg-secondary-50  rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">

        <div class="px-6 py-4 border-b border-secondary-200  bg-secondary-50  rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">Rekomendasi & Tanggapan</h3>
            <UIcon name="close" @click="store.closeModalF05" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>
        
        <div class="space-y-6 m-6">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField label="Rekomendasi Auditor (Solusi)" class="font-semibold text-sm mt-2" />
            <UTextarea class="md:col-span-3" v-model="store.planForm.recommendation" :rows="3" placeholder="Ex: Tim IT perlu menambahkan fitur hard-block..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField label="Tanggapan Audite (Managemen Response)" class="font-semibold text-sm mt-2" />
            <UTextarea class="md:col-span-3" v-model="store.planForm.response" :rows="3" placeholder="Ex: Kami setuju, update akan dilakukan di Q3..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800  mb-6">Detail Rencana Aksi (Action Plan)</h2>
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <UFormField label="Deskripsi Action" class="font-semibold text-sm" />
            <UInput class="md:col-span-3" v-model="store.planForm.actionDescription" placeholder="Ex: Staf lupa meminta TTD Manager" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <UFormField label="PIC (Penanggung Jawab)" class="font-semibold text-sm" />
            <USelectMenu class="md:col-span-3" v-model="store.planForm.pic" icon="i-heroicons-magnifying-glass" :items="store.options.pic" placeholder="Cari Nama Karyawan Ex: Dimas - IT" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <UFormField label="Periode Perencanaan" class="font-semibold text-sm" />
            <UInput class="md:col-span-3" type="date" v-model="store.planForm.periodAction" icon="i-heroicons-calendar" />
            </div>
        

        <div class="flex justify-end pt-10 border-gray-100">
            <UButton 
                :label="store.isEditingF05 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF05" 
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