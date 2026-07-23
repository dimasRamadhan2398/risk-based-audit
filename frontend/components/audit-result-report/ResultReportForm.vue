<template>
  <UModal v-model:open="store.showModal" dismissible :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
    <template #content>
      <div class="flex flex-col h-full max-h-[90vh]">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 rounded-t-xl flex justify-between items-center">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg">
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

              <!-- Findings List Editor -->
              <div class="md:col-span-2 space-y-4 pt-4 border-t border-gray-100">
                <div class="flex justify-between items-center">
                  <h4 class="text-sm font-bold text-gray-900 uppercase tracking-wide flex items-center gap-2">
                    <UIcon name="i-heroicons-list-bullet" class="text-primary-600" />
                    Findings
                  </h4>
                  <UButton
                    label="Add Finding"
                    size="md"
                    color="primary"
                    variant="soft"
                    icon="i-heroicons-plus"
                    @click="() => { if (!store.reportForm.findings) store.reportForm.findings = []; store.reportForm.findings.push({ title: '', category: 'Significant', action: '' }) }"
                  />
                </div>
                
                <div class="space-y-3">
                  <div
                    v-for="(finding, idx) in store.reportForm.findings"
                    :key="idx"
                    class="flex items-start gap-3 bg-slate-50 dark:bg-slate-850 p-3 rounded-lg border border-slate-100 dark:border-slate-800"
                  >
                    <div class="flex-1 space-y-2">
                      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                        <UFormField label="Category" size="sm" class="md:col-span-1">
                          <USelectMenu
                            v-model="finding.category"
                            :items="['Very Significant', 'Significant', 'Quite Significant', 'Not Significant']"
                            class="w-full"
                          />
                        </UFormField>
                        <UFormField label="Finding Description / Title" size="sm" class="md:col-span-1">
                          <UInput
                            v-model="finding.title"
                            placeholder="e.g. Keterlambatan rekonsiliasi kas harian"
                            required
                            class="w-full"
                          />
                        </UFormField>
                        <UFormField label="Action / Tindak Lanjut" size="sm" class="md:col-span-2">
                          <UTextarea
                            v-model="finding.action"
                            placeholder="e.g. Evaluasi SOP dan pelatihan ulang"
                            :rows="2"
                            class="w-full"
                          />
                        </UFormField>
                      </div>
                    </div>
                    <UButton
                      color="error"
                      variant="ghost"
                      icon="i-heroicons-trash"
                      size="sm"
                      class="mt-6"
                      @click="() => { store.reportForm.findings.splice(idx, 1) }"
                    />
                  </div>
                  <div v-if="!store.reportForm.findings || store.reportForm.findings.length === 0" class="text-center py-4 bg-slate-50 dark:bg-slate-850/50 rounded-lg border border-dashed border-slate-200 dark:border-slate-800 text-md text-slate-400">
                    No findings added yet. Click "Add Finding" to list significant findings.
                  </div>
                </div>
              </div>
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
