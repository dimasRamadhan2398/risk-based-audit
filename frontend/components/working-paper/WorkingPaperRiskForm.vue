<template>
   
      <UModal v-model:open="store.showModalF02" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
        <div></div>

        <template #content>
        <UForm :schema="riskSchema" :state="store.riskForm" @submit.prevent="store.handleSubmitF02">
        <div class="rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <div class="px-6 py-4 border-b border-secondary-200 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">Risk</h3>
            <UIcon name="close" @click="store.closeModalF02" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>

        <div class="space-y-6 m-6">
        <UFormField 
            label="Risk" 
            name="risk" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <USelectMenu v-model="store.riskForm.risk" :items="store.options.risk" placeholder="Choose Risk" class="w-full" />
        </UFormField>

        <UFormField 
            label="Risk Category" 
            name="taxonomy" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UInput v-model="store.riskForm.taxonomy" disabled placeholder="Automatically filled when selecting risk" class="w-full" />
        </UFormField>

        <UFormField 
            label="Risk Level" 
            name="riskLevel" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UInput v-model="store.riskForm.riskLevel" disabled placeholder="Automatically filled when selecting risk" class="w-full" />
        </UFormField>

        <UFormField 
            label="Description of Tested Controls" 
            name="controlDescription" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UTextarea v-model="store.riskForm.controlDescription" :rows="4" placeholder="Type the security control/SOP that is being evaluated in the field..." class="w-full" />
        </UFormField>
    
        <div class="flex justify-end pt-10 border-gray-100">
            <UButton 
                type="submit"
                :label="store.isEditingF02 ? 'Update Data' : 'Submit'" 
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
import { useWorkingPaperStore, riskSchema } from '~/stores/working-paper'

const store = useWorkingPaperStore()
</script>