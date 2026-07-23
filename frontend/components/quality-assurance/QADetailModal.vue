<template>
    <UModal v-model:open="store.isDetailOpen" scrollable class="w-full sm:max-w-2xl bg-[var(--bg-main)] border-[var(--border-main)]">
      <template #content>
        <UCard :ui="{ header: 'px-6 py-4', body: 'px-6 py-6', footer: 'px-6 py-4' }">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4">
                <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" @click="store.closeDetail" />
                <h3 class="text-xl font-bold">Detail: {{ store.selectedReport?.assessmentTitle }}</h3>
              </div>
              <div class="flex items-center space-x-2">
                <UButton 
                  icon="i-lucide-edit" 
                  label="Edit" 
                  color="neutral" 
                  variant="ghost" 
                  class="font-bold" 
                  @click="store.editReport"
                />
                <UButton 
                  icon="i-lucide-trash" 
                  label="Delete" 
                  color="error" 
                  variant="ghost" 
                  class="font-bold" 
                  @click="store.deleteReport"
                />
              </div>
            </div>
          </template>

          <div class="space-y-10">
             <!-- Section 1 -->
             <div class="space-y-4">
              <h4 class="text-lg font-bold text-gray-900 dark:text-white">1. General Information</h4>
              <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
                <div class="grid grid-cols-3 gap-4">
                  <p class="font-bold text-gray-700">Assessment Title</p>
                  <p class="col-span-2 font-medium">{{ store.selectedReport?.assessmentTitle }}</p>
                </div>
                <div class="grid grid-cols-3 gap-4">
                  <p class="font-bold text-gray-700">Assessment Type</p>
                  <div class="col-span-2 flex items-center space-x-2">
                    <div :class="['w-4 h-4 rounded-full', store.getTypeIconColor(store.selectedReport?.type!)]"></div>
                    <p class="font-medium">{{ store.selectedReport?.type }}</p>
                  </div>
                </div>
                <div class="grid grid-cols-3 gap-4">
                  <p class="font-bold text-gray-700">Period</p>
                  <p class="col-span-2 font-medium">{{ store.selectedReport?.period }}</p>
                </div>
              </div>
            </div>

            <!-- Section 2 -->
            <div class="space-y-4">
              <h4 class="text-lg font-bold text-gray-900 dark:text-white">2. Results</h4>
              <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
                <div class="grid grid-cols-3 gap-4">
                  <p class="font-bold text-gray-700">Result/Score</p>
                  <p class="col-span-2 text-xl font-bold">
                    {{ store.selectedReport?.type === QAType.QAR ? formatOverallConclusion(store.selectedReport?.result!) : store.selectedReport?.result }}
                  </p>
                </div>
                <div class="grid grid-cols-3 gap-4">
                  <p class="font-bold text-gray-700">Status</p>
                  <div class="col-span-2 flex items-center space-x-2">
                    <div :class="['w-4 h-4 rounded-full', store.getStatusColor(store.selectedReport?.status!)]"></div>
                    <p class="font-bold">{{ store.selectedReport?.status }} (Final Project)</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Section 3 -->
            <div class="space-y-4" v-if="store.selectedReport?.validator || store.selectedReport?.internalEvaluator || store.selectedReport?.conductedBy">
              <h4 class="text-lg font-bold text-gray-900 dark:text-white">3. Special Details</h4>
              <div class="p-6 border border-gray-100 dark:border-gray-800 rounded-xl space-y-6">
                <div class="grid grid-cols-3 gap-4" v-if="store.selectedReport?.conductedBy">
                  <p class="font-bold text-gray-700">Conducted By:</p>
                  <p class="col-span-2 font-bold text-orange-500">{{ store.selectedReport?.conductedBy }}</p>
                </div>
                <div class="grid grid-cols-3 gap-4" v-if="store.selectedReport?.validator || store.selectedReport?.internalEvaluator">
                  <p class="font-bold text-gray-700">{{ store.selectedReport?.validator ? 'Professional Consultant (KAP) / Validator:' : 'Internal Evaluator:' }}</p>
                  <p class="col-span-2 font-bold">{{ store.selectedReport?.validator || store.selectedReport?.internalEvaluator }}</p>
                </div>
              </div>
            </div>

            <!-- Section 4 -->
            <div class="space-y-4">
              <h4 class="text-lg font-bold text-gray-900 dark:text-white">4. Supporting Documents</h4>
              <div class="p-4 border border-gray-100 dark:border-gray-800 rounded-xl flex items-center justify-between bg-gray-50/50 dark:bg-gray-900/50">
                 <div class="flex items-center space-x-3" v-if="store.selectedReport?.attachment">
                    <UIcon name="i-lucide-file-text" class="size-8 text-gray-400" />
                    <div class="space-y-0.5">
                      <p class="font-bold text-sm">{{ store.selectedReport.attachment.name }}</p>
                      <p class="text-md text-gray-500">{{ store.selectedReport.attachment.size }} • Uploaded</p>
                    </div>
                 </div>
                 <div class="flex items-center space-x-3" v-else>
                    <UIcon name="i-lucide-file-text" class="size-8 text-gray-400" />
                    <div class="space-y-0.5">
                      <p class="font-bold text-sm text-gray-400">Tidak ada lampiran</p>
                      <p class="text-md text-gray-400">No document attached</p>
                    </div>
                 </div>
                 <div class="flex items-center space-x-2" v-if="store.selectedReport?.attachment">
                    <UButton 
                      v-if="store.selectedReport.attachment.filePath"
                      icon="i-lucide-download" 
                      label="Unduh" 
                      color="neutral" 
                      variant="ghost" 
                      size="sm" 
                      class="font-bold" 
                      @click="store.downloadAttachment(store.selectedReport.id, store.selectedReport.attachment.name)"
                    />
                 </div>
              </div>
            </div>
          </div>
        </UCard>
      </template>
    </UModal>
</template>

<script setup lang="ts">
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType } from '~/types/quality-assurance'

const store = useQualityAssuranceStore()

const formatOverallConclusion = (result: string) => {
  if (!result) return '-'
  const res = result.trim().toLowerCase()
  if (res === 'g/c*' || res === 'gc' || res.includes('generally')) {
    return 'Generally Conformed'
  }
  if (res === 'fc' || res.includes('fully') || res.includes('conformance')) {
    return 'Fully Conformance'
  }
  if (res === 'pc' || res.includes('partially')) {
    return 'Partially Conform'
  }
  if (res === 'dnc' || res.includes('does not') || res.includes('doesnot')) {
    return 'Does not Conform'
  }
  return result
}
</script>