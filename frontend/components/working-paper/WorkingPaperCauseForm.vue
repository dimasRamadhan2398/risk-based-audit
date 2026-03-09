<template>
    <Teleport to="body">
      <div v-if="store.showModalF04" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-secondary-50 dark:bg-secondary-300 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <UForm @submit.prevent="store.handleSubmitF04">

        <div class="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 dark:text-white">Analisis Akar Penyebab</h3>
            <UIcon name="close" @click="store.closeModalF04" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>
        
        <div class="space-y-6 m-6">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Kondisi (Dampak di Lapangan)</label>
            <UTextarea class="md:col-span-3" v-model="store.causeForm.condition" :rows="3" placeholder="Ex: Ditemukan dokumen PO tanpa tanda tangan Manager terkait pada tanggal..." />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Unggah bukti temuan/foto</label>
            
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
            <label class="font-semibold text-sm">Kriteria (Aturan/SOP Terkait) <span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.causeForm.criteria" placeholder="Cari Peraturan Internal (Contoh: SOP Pengadaan Bab IV - Otorisasi)" />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <label class="font-semibold text-sm mt-2">Dampak (Risk Impact)</label>
            <UTextarea class="md:col-span-3" v-model="store.causeForm.impact" :rows="3" placeholder="Ex: Potensi fraud atau pembelian fiktif yang dapat merugikan keuangan..." />
            </div>

            <h2 class="text-xl text-center font-bold text-gray-800 dark:text-white mb-6">Analisa Akar Masalah</h2>
            <div v-for="(rca, index) in store.causeForm.rootCause" :key="rca.id" class="border border-gray-200 dark:border-gray-700 rounded-xl p-6 bg-white dark:bg-gray-800">
            <h3 class="text-lg font-bold mb-4">Analisis {{ index + 1 }}</h3>
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Kategori <span class="text-red-500">*</span></label>
                <USelectMenu class="md:col-span-3" v-model="rca.method" :items="store.options.rootCauseMethod" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 1 <span class="text-red-500">*</span></label>
                <UInput class="md:col-span-3" v-model="rca.w1" placeholder="Ex: Staf lupa meminta TTD Manager" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 2</label>
                <UInput class="md:col-span-3" v-model="rca.w2" placeholder="Ex: Karena staf terburu-buru mengejar kuota pengiriman" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <label class="font-semibold text-sm">Why 3</label>
                <UInput class="md:col-span-3" v-model="rca.w3" placeholder="-" />
                </div>
            </div>

            <div class="mt-4 flex justify-end">
                <UButton label="Hapus Analisa Akar Masalah" color="error" variant="ghost" class="font-bold" @click="store.removeRootCause(index)" />
            </div>
            </div>

            <UButton color="primary" icon="i-heroicons-plus" variant="soft" label="Tambah Analisa Akar Masalah" @click="store.addRootCause" />
        
        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton 
                :label="store.isEditingF04 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF04" 
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

const store = useWorkingPaperStore()

</script>