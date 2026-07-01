<template>
   
      <UModal v-model:open="store.showModalF02" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
        <div></div>

        <template #content>
        <UForm :state="store.riskForm" @submit.prevent="store.handleSubmitF02">
        <div class="rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <div class="px-6 py-4 border-b border-secondary-200 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">Risk</h3>
            <UIcon name="close" @click="store.closeModalF02" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>

        <div class="space-y-6 m-6">
        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Risk" class="font-semibold text-sm text-gray-700  mt-2" />
            <USelectMenu class="md:col-span-3" v-model="store.riskForm.risk" :items="store.options.risk" placeholder="Choose Risk" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Risk Category" class="font-semibold text-sm text-gray-700  mt-2" />
            <UInput class="md:col-span-3" v-model="store.riskForm.taxonomy" disabled placeholder="Automatically filled when selecting risk" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Risk Level" class="font-semibold text-sm text-gray-700  mt-2" />
            <UInput class="md:col-span-3" v-model="store.riskForm.riskLevel" disabled placeholder="Automatically filled when selecting risk" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Description of Tested Controls" class="font-semibold text-sm text-gray-700  mt-2" />
            <UTextarea class="md:col-span-3" v-model="store.riskForm.controlDescription" :rows="4" placeholder="Type the security control/SOP that is being evaluated in the field..." />
        </div>
    
        <div class="flex justify-end pt-10 border-gray-100">
            <UButton 
                :label="store.isEditingF02 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF02" 
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