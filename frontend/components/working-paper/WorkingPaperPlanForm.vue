<template>
    
      <UModal v-model:open="store.showModalF05" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
        <div></div>

        <template #content>
        <UForm :schema="planSchema" :state="store.planForm" @submit.prevent="store.handleSubmitF05">
        <div class="rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">

        <div class="px-6 py-4 border-b border-secondary-200 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">Rekomendasi & Tanggapan</h3>
            <UIcon name="close" @click="store.closeModalF05" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>
        
        <div class="space-y-6 m-6">
            <UFormField 
                label="Rekomendasi Auditor (Solusi)" 
                name="recommendation" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm mt-2' }"
            >
                <UTextarea v-model="store.planForm.recommendation" :rows="3" placeholder="Ex: Tim IT perlu menambahkan fitur hard-block..." class="w-full" />
            </UFormField>

            <UFormField 
                label="Tanggapan Audite (Managemen Response)" 
                name="response" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm mt-2' }"
            >
                <UTextarea v-model="store.planForm.response" :rows="3" placeholder="Ex: Kami setuju, update akan dilakukan di Q3..." class="w-full" />
            </UFormField>

            <h2 class="text-xl text-center font-bold text-gray-800  mb-6">Detail Rencana Aksi (Action Plan)</h2>
            <UFormField 
                label="Deskripsi Action" 
                name="actionDescription" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm' }"
            >
                <UInput v-model="store.planForm.actionDescription" placeholder="Ex: Staf lupa meminta TTD Manager" class="w-full" />
            </UFormField>

            <UFormField 
                label="PIC (Penanggung Jawab)" 
                name="pic" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm' }"
            >
                <USelectMenu v-model="store.planForm.pic" icon="i-heroicons-magnifying-glass" :items="store.options.pic" placeholder="Cari Nama Karyawan Ex: Dimas - IT" class="w-full" />
            </UFormField>

            <UFormField 
                label="Periode Perencanaan" 
                name="periodAction" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm' }"
            >
                <UInput type="date" v-model="store.planForm.periodAction" icon="i-heroicons-calendar" class="w-full" />
            </UFormField>
        

        <div class="flex justify-end pt-10 border-gray-100">
            <UButton 
                type="submit"
                :label="store.isEditingF05 ? 'Update Data' : 'Submit'" 
                color="primary"
            />
        </div>
        </div>
        </div>
    </UForm>
    
    </template>
    </UModal>
    
</template>

<script setup lang="ts">
import { useWorkingPaperStore, planSchema } from '~/stores/working-paper'

const store = useWorkingPaperStore()
</script>