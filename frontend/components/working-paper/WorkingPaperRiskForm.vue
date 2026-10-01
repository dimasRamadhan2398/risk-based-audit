<template>
   
      <UModal 
        v-model:open="store.showModalF02" 
        :dismissible="false" 
        :ui="{
            content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-3xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'p-0',
            body: 'p-0 flex-1 min-h-0 flex flex-col overflow-hidden',
            footer: 'p-0',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
        }"
    >
        <template #content>
          <UForm :schema="riskSchema" :state="store.riskForm" class="flex flex-col h-full max-h-[90vh] overflow-hidden" @submit.prevent="store.handleSubmitF02">
            <!-- Pinned Header -->
            <div class="px-5 sm:px-6 py-4 border-b border-gray-100 dark:border-gray-800 flex justify-between items-center shrink-0 bg-white dark:bg-gray-900">
              <div class="flex items-center gap-3">
                <div class="w-9 h-9 rounded-xl bg-primary-50 dark:bg-primary-950/60 flex items-center justify-center text-primary-500">
                  <UIcon name="i-lucide-shield-alert" class="w-5 h-5" />
                </div>
                <h3 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white">{{ t('workingPaper.riskForm.title') }}</h3>
              </div>
              <UButton
                color="neutral"
                variant="ghost"
                icon="i-heroicons-x-mark-20-solid"
                class="-my-1"
                @click="store.closeModalF02"
              />
            </div>

            <!-- Scrollable Body -->
            <div class="p-4 sm:p-6 space-y-5 overflow-y-auto flex-1 min-h-0">
              <UFormField 
                :label="t('workingPaper.riskForm.risk')" 
                name="risk" 
                required 
                class="grid grid-cols-1 md:grid-cols-4 items-start gap-2 md:gap-4" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 dark:text-gray-300 md:mt-2' }"
              >
                <USelectMenu v-model="store.riskForm.risk" :items="store.options.risk" :placeholder="t('workingPaper.riskForm.riskPlaceholder')" class="w-full" />
              </UFormField>

              <UFormField 
                :label="t('workingPaper.riskForm.riskCategory')" 
                name="taxonomy" 
                required 
                class="grid grid-cols-1 md:grid-cols-4 items-start gap-2 md:gap-4" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 dark:text-gray-300 md:mt-2' }"
              >
                <UInput v-model="store.riskForm.taxonomy" disabled :placeholder="t('workingPaper.riskForm.autoFilledTaxonomy')" class="w-full bg-slate-50 dark:bg-slate-800" />
              </UFormField>

              <UFormField 
                :label="t('workingPaper.riskForm.riskLevel')" 
                name="riskLevel" 
                required 
                class="grid grid-cols-1 md:grid-cols-4 items-start gap-2 md:gap-4" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 dark:text-gray-300 md:mt-2' }"
              >
                <UInput v-model="store.riskForm.riskLevel" disabled :placeholder="t('workingPaper.riskForm.autoFilledRiskLevel')" class="w-full bg-slate-50 dark:bg-slate-800" />
              </UFormField>

              <UFormField 
                :label="t('workingPaper.riskForm.controlDescription')" 
                name="controlDescription" 
                required 
                class="grid grid-cols-1 md:grid-cols-4 items-start gap-2 md:gap-4" 
                :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 dark:text-gray-300 md:mt-2' }"
              >
                <UTextarea v-model="store.riskForm.controlDescription" :rows="4" :placeholder="t('workingPaper.riskForm.controlDescriptionPlaceholder')" class="w-full" />
              </UFormField>
            </div>

            <!-- Pinned Footer -->
            <div class="px-5 sm:px-6 py-3.5 border-t border-gray-100 dark:border-gray-800 flex justify-end items-center gap-3 shrink-0 bg-gray-50/70 dark:bg-gray-800/40">
              <UButton
                color="neutral"
                variant="ghost"
                :label="t('common.cancel')"
                @click="store.closeModalF02"
              />
              <UButton 
                type="submit"
                :label="store.isEditingF02 ? t('common.updateData') : t('common.submit')" 
                color="primary"
                class="font-semibold px-4"
              />
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