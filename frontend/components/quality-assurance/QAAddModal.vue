<template>
    <UModal v-model:open="store.isFormOpen" scrollable class="w-full sm:max-w-2xl bg-[var(--bg-main)] border-[var(--border-main)]">
      <template #content>
        <UCard :ui="{ header: 'px-6 py-4', body: 'px-6 py-6', footer: 'px-6 py-4' }">
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
                <UFormField label="Assessment Title" required>
                  <UInput v-model="store.newReport.assessmentTitle" placeholder="Ex: QAR - Audit 2026" class="w-full"/>
                </UFormField>
                <div class="grid grid-cols-2 gap-4">
                  <UFormField label="Execution Period">
                    <USelectMenu v-model="store.newReport.periodQuarter" :items="['Q1', 'Q2', 'Q3', 'Q4']" placeholder="Select Quarter" class="w-full"/>
                  </UFormField>
                  <UFormField label="&nbsp;">
                    <USelectMenu v-model="store.newReport.periodYear" :items="store.periods" placeholder="Select Year" class="w-full"/>
                  </UFormField>
                </div>
              </div>
            </div>

            <!-- Section 3 -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700">3. Results & Status</h4>
              <div class="grid grid-cols-2 gap-4">
                <UFormField label="Status" required>
                  <USelectMenu v-model="store.newReport.status" :items="store.qaStatuses" placeholder="Select Status" class="w-full"/>
                </UFormField>
                <UFormField label="Result/Score" required>
                  <USelectMenu 
                    v-if="store.newReport.type === QAType.IACM"
                    v-model="store.newReport.result" 
                    :items="['1', '2', '3', '4', '5']" 
                    placeholder="Select Score (1-5)" 
                    class="w-full"
                  />
                  <USelectMenu 
                    v-else-if="store.newReport.type === QAType.QAR"
                    v-model="store.newReport.result" 
                    :items="['Does not Conform', 'Partially Conform', 'Generally Conformed', 'Fully Conformance']" 
                    placeholder="Select Conformance" 
                    class="w-full"
                  />
                  <UInput 
                    v-else-if="store.newReport.type === QAType.REGULAR"
                    v-model="store.newReport.result" 
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
                  <UInput v-model="store.newReport.conductedBy" placeholder="Ex: PT Nama Perusahaan" class="w-full"/>
                </UFormField>
                <UFormField label="Internal Evaluator">
                  <UInput v-model="store.newReport.internalEvaluator" placeholder="Team Name / Lead Auditor..." class="w-full"/>
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
              <UButton :label="store.isEditing ? 'Update Report' : 'Save Report'" color="warning" class="px-8 font-bold" @click="store.saveReport" />
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
</template>

<script setup lang="ts">

import { useQualityAssuranceStore, QAType } from '~/stores/quality-assurance'

const store = useQualityAssuranceStore()

</script>