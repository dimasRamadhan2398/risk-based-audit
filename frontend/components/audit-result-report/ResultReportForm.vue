<template>
  <UModal v-model:open="store.showModal" dismissible class="w-full sm:max-w-2xl">
    <template #content>
      <div class="flex flex-col h-full max-h-[90vh]">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 bg-gray-50 rounded-t-xl flex justify-between items-center">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-primary-100 rounded-lg">
              <UIcon name="i-heroicons-document-text" class="text-primary-600 size-6" />
            </div>
            <h3 class="text-lg font-bold text-gray-900">
              {{ store.isEditing ? 'Edit Audit Result Report' : 'Create New Audit Result Report' }}
            </h3>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark"
            @click="store.closeModal"
          />
        </div>

        <!-- Body -->
        <div class="p-6 overflow-y-auto flex-1">
          <UForm :state="store.reportForm" class="space-y-6" @submit="store.saveReport">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <UFormField label="Report Title" name="reportTitle" required class="md:col-span-2">
                <UInput
                  v-model="store.reportForm.reportTitle"
                  placeholder="e.g. Audit Report - Financial Operations 2026"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Report Date" name="reportDate" required>
                <UInput
                  v-model="store.reportForm.reportDate"
                  type="date"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Overall Rating" name="overallRating" required>
                <USelectMenu
                  v-model="store.reportForm.overallRating"
                  :items="['Satisfactory', 'Needs Improvement', 'Unsatisfactory']"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Findings Count" name="findingsCount">
                <UInput
                  v-model="store.reportForm.findingsCount"
                  type="number"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Status" name="status">
                <USelectMenu
                  v-model="store.reportForm.status"
                  :items="['Draft', 'Final']"
                  class="w-full"
                />
              </UFormField>

              <UFormField label="Executive Summary" name="executiveSummary" class="md:col-span-2">
                <UTextarea
                  v-model="store.reportForm.executiveSummary"
                  placeholder="Provide a high-level summary of the audit results..."
                  :rows="6"
                  class="w-full"
                />
              </UFormField>
            </div>

            <!-- Footer Actions -->
            <div class="flex justify-end gap-3 pt-6 border-t border-gray-100">
              <UButton
                label="Cancel"
                color="neutral"
                variant="ghost"
                @click="store.closeModal"
              />
              <UButton
                type="submit"
                :label="store.isEditing ? 'Update Report' : 'Save Report'"
                color="primary"
                icon="i-heroicons-check"
              />
            </div>
          </UForm>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { useAuditResultReportStore } from '~/stores/audit-result-report'

const store = useAuditResultReportStore()
</script>
