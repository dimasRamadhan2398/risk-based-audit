<template>
    
      <UModal 
        v-model:open="store.showModalF04" 
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
        
        <UForm :schema="causeSchema" :state="store.causeForm" @submit.prevent="store.handleSubmitF04">
        <div class=" rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto  transition-colors duration-300">

        <div class="px-6 py-4 rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
            <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('workingPaper.causeForm.title') }}</h3>
            <UIcon name="close" @click="store.closeModalF04" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer"></UIcon>
        </div>
        
        <div class="space-y-6 m-6">
            <UFormField 
                :label="t('workingPaper.causeForm.condition')" 
                name="condition" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm mt-2' }"
            >
                <UTextarea 
                    v-model="store.causeForm.condition" 
                    :rows="3" 
                    :placeholder="t('workingPaper.causeForm.conditionPlaceholder')" 
                    class="w-full"
                    maxlength="200"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.causeForm.condition ? store.causeForm.condition.length : 0 }}/200
                </div>
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
            <UFormField :label="t('workingPaper.causeForm.uploadEvidence')" class="font-semibold text-sm mt-2" />
            
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

            <UFormField 
                :label="t('workingPaper.causeForm.criteria')" 
                name="criteria" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm' }"
            >
                <UInput v-model="store.causeForm.criteria" :placeholder="t('workingPaper.causeForm.criteriaPlaceholder')" class="w-full" />
            </UFormField>

            <UFormField 
                :label="t('workingPaper.causeForm.impact')" 
                name="impact" 
                class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm mt-2' }"
            >
                <UTextarea v-model="store.causeForm.impact" :rows="3" :placeholder="t('workingPaper.causeForm.impactPlaceholder')" class="w-full" />
            </UFormField>

            <h2 class="text-xl text-center font-bold text-gray-800  mb-6">{{ t('workingPaper.causeForm.rootCauseAnalysis') }}</h2>
            <div v-for="(rca, index) in store.causeForm.rootCause" :key="rca.id" class="border border-gray-200  rounded-xl p-6 ">
            
            <div class="pb-6 flex justify-between">
                <h3 class="text-lg font-bold">{{ t('workingPaper.causeForm.analysisIndex', { index: index + 1 }) }}</h3>
                <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeRootCause(index)" />
            </div>
                
            
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.causeForm.category')" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="rca.method" :items="store.options.rootCauseMethod" :placeholder="t('workingPaper.causeForm.categoryPlaceholder')" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.causeForm.why1')" class="font-semibold text-sm" />
                <UInput 
                    class="md:col-span-3" 
                    v-model="rca.w1" 
                    :placeholder="t('workingPaper.causeForm.why1Placeholder')"
                    required
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('workingPaper.causeForm.why1Validation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                />
                </div>
            </div>

            
            </div>

            <UButton color="primary" icon="i-heroicons-plus" variant="soft" :label="t('workingPaper.causeForm.addRootCause')" @click="store.addRootCause" />
        
        <div class="flex justify-end p-6 border-gray-100">
            <UButton 
                type="submit"
                :label="store.isEditingF04 ? t('common.updateData') : t('common.submit')" 
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
import { useWorkingPaperStore, causeSchema } from '~/stores/working-paper'

const { t } = useI18n()
const store = useWorkingPaperStore()

</script>