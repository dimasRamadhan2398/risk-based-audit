<template>
    <UModal 
      v-model:open="store.isFormOpen" 
      scrollable 
      :ui="{
        content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
        header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
        body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
        footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
        overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
      }"
    >
      <template #content>
        <UCard :ui="{ header: 'sticky top-0 z-20 px-6 py-4 bg-[var(--bg-main)] border-b border-[var(--border-main)]', body: 'px-6 py-6 overflow-y-auto max-h-[60vh]', footer: 'px-6 py-4'}">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-xl font-bold">{{ store.isEditing ? 'Edit Assessment' : 'Add New Assessment' }}</h3>
              <UButton color="neutral" variant="ghost" icon="i-lucide-x" @click="store.closeForm" />
            </div>
          </template>

          <div class="space-y-8">
            <!-- Section 1 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">1. Select Assessment Type</h4>
              <div class="space-y-3">
                <div 
                  v-for="type in store.qaTypes" 
                  :key="type"
                  class="flex items-center p-4 border rounded-xl cursor-pointer transition-all"
                  :class="store.newReport.type === type ? 'border-orange-500 bg-orange-50/50' : 'border-gray-200'"
                  @click="store.newReport.type = type"
                >
                  <URadio :model-value="store.newReport.type === type" class="mr-4" />
                  <div class="space-y-0.5">
                    <p class="font-bold text-sm">{{ type }}</p>
                    <p class="text-md text-gray-500" v-if="type === QAType.REGULAR">[Description: Efficiency & effectiveness focus]</p>
                    <p class="text-md text-gray-500" v-if="type === QAType.SAIV">[Description: GIAS Compliance + External Validation]</p>
                    <p class="text-md text-gray-500" v-if="type === QAType.QAR">[Description: Professional Consultant - IPPF 2027 & GIAS 2024]</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Section 2 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">2. General Information</h4>
              <div class="space-y-4">
                <UFormField label="Assessment Title" required :error="errors.assessmentTitle ? 'Title is required' : ''">
                  <UInput 
                    v-model="store.newReport.assessmentTitle" 
                    placeholder="Ex: QAR - Audit 2026" 
                    class="w-full"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                    @input="($event.target as any)?.setCustomValidity('')" 
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.newReport.assessmentTitle ? store.newReport.assessmentTitle.length : 0 }}/100
                  </div>
                </UFormField>
                <UFormField label="Execution Period">
                  <USelectMenu 
                    v-model="store.newReport.periodYear" 
                    :items="store.periods" 
                    placeholder="Select Year" 
                    class="w-full"
                  />
                </UFormField>
              </div>
            </div>

            <!-- Section 3 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">3. Results & Status</h4>
              <div class="grid grid-cols-2 gap-4">
                <UFormField label="Status" required :error="errors.status ? 'Status is required' : ''">
                  <USelectMenu v-model="store.newReport.status" :items="store.qaStatuses" placeholder="Select Status" class="w-full" />
                </UFormField>
                <UFormField label="Result/Score" required :error="errors.result ? 'Result is required' : ''">
                  <USelectMenu 
                    v-if="store.newReport.type === QAType.IACM"
                    v-model="store.newReport.result" 
                    :items="['1', '2', '3', '4', '5']" 
                    placeholder="Select Score (1-5)" 
                    class="w-full"
                  />
                  <USelectMenu 
                    v-else-if="store.newReport.type === QAType.QAR || store.newReport.type === QAType.SAIV"
                    v-model="store.newReport.result" 
                    :items="['Does not Conform', 'Partially Conform', 'Generally Conformed', 'Fully Conformance']" 
                    placeholder="Select Conformance" 
                    class="w-full"
                  />
                  <UInput 
                    v-else-if="store.newReport.type === QAType.REGULAR"
                    v-model="store.newReport.result"
                    type="number" 
                    step="0.1" 
                    min="0" 
                    max="10" 
                    placeholder="Ex: 8.5/10" 
                    class="w-full"
                  />
                  <UInput 
                    v-else
                    v-model="store.newReport.result" 
                    placeholder="Ex: 92%" 
                    class="w-full"
                  />
                </UFormField>
              </div>
            </div>

            <!-- Section 4 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">4. Special Details</h4>
              <div class="grid grid-cols-2 gap-4">
                <UFormField label="Conducted By">
                  <USelectMenu 
                    v-model="store.newReport.conductedBy" 
                    :items="['PT BAI', 'External']" 
                    placeholder="Select Conducted By" 
                    class="w-full"
                  />
                </UFormField>
                <UFormField label="Internal Evaluator">
                  <UInput 
                    v-model="store.newReport.internalEvaluator" 
                    placeholder="Team Name / Lead Auditor..." 
                    class="w-full"
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.newReport.internalEvaluator ? store.newReport.internalEvaluator.length : 0 }}/100
                  </div>
                </UFormField>
              </div>
            </div>

            <!-- Section 5 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">5. Supporting Documents</h4>
              <UFileUpload
                v-model="store.newReport.attachment"
                label="Click to upload or drag and drop"
                description="PDF, DOCX up to 10MB"
                accept=".pdf,.docx,.doc"
                :max-size="10 * 1024 * 1024"
                :icon="'i-lucide-file-up'"
                :file-icon="'i-lucide-file-text'"
                :file-delete="{ color: 'neutral', variant: 'link' }"
                class="w-full"
              />
            </div>
          </div>

          <template #footer>
            <div class="flex justify-end gap-3">
              <UButton label="Cancel" variant="ghost" color="neutral" @click="store.closeForm" />
              <UButton :label="store.isEditing ? 'Update Report' : 'Save Report'" color="primary" class="px-8 font-bold" @click="validateAndSave" />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
</template>

<script setup lang="ts">

import { ref } from 'vue'
import { useQualityAssuranceStore, QAType } from '~/stores/quality-assurance'
const toast = useToast()
const store = useQualityAssuranceStore()

const errors = ref({
  assessmentTitle: false,
  status: false,
  result: false
})

const validateAndSave = () => {
  errors.value.assessmentTitle = !store.newReport.assessmentTitle
  errors.value.status = !store.newReport.status
  errors.value.result = !store.newReport.result

  if (errors.value.assessmentTitle || errors.value.status || errors.value.result) {
    toast.add({
      title: 'Validation Error',
      description: 'Mohon isi semua kolom yang wajib diisi (berwarna merah).',
      color: 'error',
      icon: 'i-heroicons-exclamation-circle'
    })
    return
  }
  
  store.saveReport()
}

</script>