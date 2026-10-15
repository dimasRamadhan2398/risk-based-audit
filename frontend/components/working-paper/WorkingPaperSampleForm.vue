<template>

      <UModal
        v-model:open="store.showModalF03"
        :dismissible="false"
        :ui="{
            content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-4 sm:p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-4 sm:p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
        }"
    >
        <template #content>
        <UForm :schema="sampleSchema" :state="store.sampleForm" @submit.prevent="store.handleSubmitF03">
        <div class="bg-[var(--bg-main)] rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto border border-[var(--border-main)] transition-colors duration-300">
        <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
            <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
            <h3 class="text-lg font-bold text-[var(--text-main)]">{{ t('workingPaper.sampleForm.title') }}</h3>
            <UIcon name="close" @click="store.closeModalF03" class="text-[var(--text-muted)] hover:text-[var(--text-main)] text-2xl cursor-pointer"></UIcon>
        </div>
        <div class="space-y-6 m-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-full">
            <UFormField :label="t('workingPaper.sampleForm.population')" required name="population">
            <UInput type="number" v-model="store.sampleForm.population" placeholder="Ex: 100" class="w-full"/>
            </UFormField>
            <UFormField :label="t('workingPaper.sampleForm.sampleSize')" required name="sampleSize">
            <UInput type="number" v-model="store.sampleForm.sampleSize" placeholder="Ex: 10" class="w-full"/>
            </UFormField>
        </div>


            <div v-for="(sampel, index) in store.sampleForm.samples" :key="sampel.id" class="border border-gray-200  rounded-xl p-6 relative">
            <div class="mb-4 flex justify-between items-center">
                <h3 class="text-lg font-bold">{{ t('workingPaper.sampleForm.sample', { number: index + 1 }) }}</h3>
                <UIcon name="i-heroicons-trash" color="error" variant="ghost" @click="store.removeSample(index)" />
            </div>
            <div class="space-y-4 max-w-full">
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.sampleForm.document')" class="font-semibold text-sm" required/>
                <UInput
                    class="md:col-span-3"
                    v-model="sampel.document"
                    placeholder="Ex: PO-2026-001"
                    required
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity(t('workingPaper.sampleForm.documentValidation'))"
                    @input="($event.target as any)?.setCustomValidity('')"
                />

                </div>
                <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ sampel.document ? sampel.document.length : 0 }}/100
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.sampleForm.step', { number: 1 })" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l1" :items="store.options.testResult" :placeholder="t('workingPaper.sampleForm.stepPlaceholder')" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.sampleForm.step', { number: 2 })" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l2" :items="store.options.testResult" :placeholder="t('workingPaper.sampleForm.stepPlaceholder')" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
                <UFormField :label="t('workingPaper.sampleForm.step', { number: 3 })" class="font-semibold text-sm" />
                <USelectMenu class="md:col-span-3" v-model="sampel.l3" :items="store.options.testResult" :placeholder="t('workingPaper.sampleForm.stepPlaceholder')" />
                </div>
                <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center mt-2">
                <UFormField :label="t('workingPaper.sampleForm.status')" class="font-semibold text-sm" />
                <div class="md:col-span-3 flex items-center gap-2">
                    <div class="w-4 h-4 rounded-full" :class="store.checkSampleStatus(sampel) ? 'bg-green-500' : 'bg-red-500'"></div>
                    <span class="font-bold">{{ store.checkSampleStatus(sampel) ? t('workingPaper.common.effective') : t('workingPaper.common.ineffective') }}</span>
                </div>
                </div>
            </div>

            </div>
            <UButton color="primary" icon="i-heroicons-plus" variant="soft" :label="t('workingPaper.sampleForm.addSample')" @click="store.addSample" />


        <UFormField
            :label="t('workingPaper.sampleForm.conclusion')"
            name="conclusion"
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full"
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm' }"
        >
            <UTextarea v-model="store.sampleForm.conclusion" :rows="4" :placeholder="t('workingPaper.sampleForm.conclusionPlaceholder')" class="w-full" />
        </UFormField>

        <div class="flex justify-end p-6 border-gray-100">
            <UButton
                type="submit"
                :label="store.isEditingF03 ? t('workingPaper.sampleForm.update') : t('workingPaper.sampleForm.submit')"
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
import { useWorkingPaperStore, sampleSchema } from '~/stores/working-paper'

const store = useWorkingPaperStore()
const { t } = useI18n()
</script>
