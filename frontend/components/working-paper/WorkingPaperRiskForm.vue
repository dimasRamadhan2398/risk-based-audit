<template>
   
      <UModal 
        v-model:open="store.showModalF02" 
        :dismissible="false" 
        :ui="{
            content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
        }"
    >
        
        <template #content>
        <UForm :schema="riskSchema" :state="store.riskForm" @submit.prevent="store.handleSubmitF02">
        <div class="rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <div class="px-6 py-4 border-b border-secondary-200 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">{{ t('workingPaper.riskForm.title') }}</h3>
            <UIcon name="close" @click="store.closeModalF02" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>

        <div class="space-y-6 m-6">
        <UFormField 
            :label="t('workingPaper.riskForm.risk')" 
            name="risk" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <USelectMenu v-model="store.riskForm.risk" :items="store.options.risk" :placeholder="t('workingPaper.riskForm.riskPlaceholder')" class="w-full" />
        </UFormField>

        <UFormField 
            :label="t('workingPaper.riskForm.riskCategory')" 
            name="taxonomy" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UInput v-model="store.riskForm.taxonomy" disabled :placeholder="t('workingPaper.riskForm.autoFilledTaxonomy')" class="w-full" />
        </UFormField>

        <UFormField 
            :label="t('workingPaper.riskForm.riskLevel')" 
            name="riskLevel" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UInput v-model="store.riskForm.riskLevel" disabled :placeholder="t('workingPaper.riskForm.autoFilledRiskLevel')" class="w-full" />
        </UFormField>

        <UFormField 
            :label="t('workingPaper.riskForm.controlDescription')" 
            name="controlDescription" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UTextarea v-model="store.riskForm.controlDescription" :rows="4" :placeholder="t('workingPaper.riskForm.controlDescriptionPlaceholder')" class="w-full" />
        </UFormField>
    
        <div class="flex justify-end pt-10 border-gray-100">
            <UButton 
                type="submit"
                :label="store.isEditingF02 ? t('common.updateData') : t('common.submit')" 
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
import { useI18n } from '~/composables/useI18n'
import { useWorkingPaperStore, riskSchema } from '~/stores/working-paper'

const { t } = useI18n()
const store = useWorkingPaperStore()
</script>