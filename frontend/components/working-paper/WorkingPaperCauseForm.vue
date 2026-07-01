<template>
    
      <UModal v-model:open="store.showModalF04" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
        <div></div>
        <template #content>
        
        <UForm :state="store.causeForm" @submit.prevent="store.handleSubmitF04">
        <div class=" rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto  transition-colors duration-300">

        <div class="px-6 py-4 rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
            <h3 class="text-lg font-bold text-[var(--text-main)]">Analisis Akar Penyebab</h3>
            <UIcon name="close" @click="store.closeModalF04" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer"></UIcon>
        </div>
        
        <div class="space-y-6 m-6">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField label="Kondisi (Dampak di Lapangan)" class="font-semibold text-sm mt-2" />
            <UTextarea class="md:col-span-3" v-model="store.causeForm.condition" :rows="3" placeholder="Ex: Ditemukan dokumen PO tanpa tanda tangan Manager terkait pada tanggal..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField label="Unggah bukti temuan/foto" class="font-semibold text-sm mt-2" />
            
            <div class="md:col-span-3">
            <UFormField 
                @click="store.triggerUpload"
                class="block text-sm font-medium"
                size="lg"
            >
                <UInput 
                type="file" 
                ref="fileInput" 
                class="w-full" 
                icon="i-heroicons-paper-clip"
                accept="image/png, image/jpeg" 
                @change="store.onFileChange"
                />     
            </UFormField>
            </div>
        </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
            <UFormField label="Kriteria (Aturan/SOP Terkait)" class="font-semibold text-sm" />
            <UInput class="md:col-span-3" v-model="store.causeForm.criteria" placeholder="Cari Peraturan Internal (Contoh: SOP Pengadaan Bab IV - Otorisasi)" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField label="Dampak (Risk Impact)" class="font-semibold text-sm mt-2" />
            <UTextarea class="md:col-span-3" v-model="store.causeForm.impact" :rows="3" placeholder="Ex: Potensi fraud atau pembelian fiktif yang dapat merugikan keuangan..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800  mb-6">Analisa Akar Masalah</h2>
            <div v-for="(rca, index) in store.causeForm.rootCause" :key="rca.id" class="border border-gray-200  rounded-xl p-6 ">
            
            <div class="pb-6 flex justify-between">
                <h3 class="text-lg font-bold">Analisis {{ index + 1 }}</h3>
                <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeRootCause(index)" />
            </div>
                
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Kategori" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="rca.method" :items="store.options.rootCauseMethod" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Why 1" class="font-semibold text-sm" />
                <UInput class="md:col-span-3" v-model="rca.w1" placeholder="Ex: Staf lupa meminta TTD Manager" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Why 2" class="font-semibold text-sm" />
                <UInput class="md:col-span-3" v-model="rca.w2" placeholder="Ex: Karena staf terburu-buru mengejar kuota pengiriman" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField label="Why 3" class="font-semibold text-sm" />
                <UInput class="md:col-span-3" v-model="rca.w3" placeholder="-" />
                </div>
            </div>

            
            </div>

            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Analisa Akar Masalah" @click="store.addRootCause" />
        
        <div class="flex justify-end p-6 border-gray-100">
            <UButton 
                :label="store.isEditingF04 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF04" 
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

const store = useWorkingPaperStore()

</script>