<template>
  <UModal 
    v-model:open="store.showModal" 
    dismissible 
    :ui="{
      content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
      body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
      footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
  >
    <template #content>
      <div class="flex flex-col h-full max-h-[90vh]">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 rounded-t-xl flex justify-between items-center">
          <div class="flex items-center gap-3">
            <div class="p-2 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600">
              <UIcon name="i-heroicons-document-text" class="size-6" />
            </div>
            <div>
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">
                {{ store.isEditing ? 'Edit Audit Result Report' : 'Create New Audit Result Report' }}
              </h3>
              <p class="text-xs text-gray-500">
                Surat Tugas: <span class="font-semibold text-primary-600">{{ store.reportForm.assignmentLetterId || store.selectedAssignmentLetter || '-' }}</span>
              </p>
            </div>
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
              <UFormField label="Assignment Letter (Surat Tugas)" name="assignmentLetterId" required class="md:col-span-2">
                <USelectMenu
                  v-model="store.reportForm.assignmentLetterId"
                  :items="store.publishedAssignmentLetters"
                  placeholder="Select Assignment Letter"
                  class="w-full"
                  :disabled="store.isEditing"
                  @update:model-value="onLetterChange"
                />
              </UFormField>

              <UFormField label="Report Title" name="reportTitle" required class="md:col-span-2">
                <UInput
                  v-model="store.reportForm.reportTitle"
                  placeholder="e.g. Audit Report - Financial Operations 2026"
                  class="w-full"
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Report Title maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                  {{ store.reportForm.reportTitle ? store.reportForm.reportTitle.length : 0 }}/100
                </div>
              </UFormField>

              <UFormField label="Report Date" name="reportDate" required>
                <AppDatePicker
                  v-model="store.reportForm.reportDate"
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

              <UFormField label="Findings Count" name="findingsCount">
                <div class="flex items-center gap-2">
                  <UInput
                    v-model="store.reportForm.findingsCount"
                    type="number"
                    class="w-full"
                  />
                  <span class="text-xs text-gray-500 whitespace-nowrap">
                    ({{ store.reportForm.findings?.length || 0 }} list)
                  </span>
                </div>
              </UFormField>

              <!-- Findings List Editor -->
              <div class="md:col-span-2 space-y-4 pt-4 border-t border-gray-100 dark:border-gray-800">
                <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
                  <div class="flex items-center gap-2">
                    <h4 class="text-sm font-bold text-gray-900 dark:text-white uppercase tracking-wide flex items-center gap-2">
                      <UIcon name="i-heroicons-list-bullet" class="text-primary-600" />
                      Findings / Temuan Audit
                    </h4>
                    <UBadge color="primary" variant="subtle" size="sm">
                      {{ store.reportForm.findings?.length || 0 }}
                    </UBadge>
                  </div>
                  
                  <div class="flex items-center gap-2 w-full sm:w-auto justify-end">
                    <UTooltip text="Tarik temuan dari modul Digital Working Paper (KKA) dan Fieldwork Test Controls">
                      <UButton
                        label="Tarik Temuan Otomatis"
                        size="sm"
                        color="primary"
                        variant="solid"
                        icon="i-heroicons-sparkles"
                        :loading="store.isAutoDetecting"
                        @click="store.runAutoDetectFindings('replace')"
                      />
                    </UTooltip>
                    <UButton
                      label="Add Manual"
                      size="sm"
                      color="neutral"
                      variant="outline"
                      icon="i-heroicons-plus"
                      @click="addManualFinding"
                    />
                  </div>
                </div>

                <!-- Findings Category Summary Badge Bar -->
                <div v-if="store.reportForm.findings && store.reportForm.findings.length > 0" class="flex flex-wrap items-center gap-2 p-3 bg-slate-50 dark:bg-slate-800/60 rounded-lg border border-slate-200 dark:border-slate-700 text-xs">
                  <span class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-1 mr-1">
                    <UIcon name="i-heroicons-chart-pie" class="size-4 text-primary-600" />
                    Distribusi Kategori:
                  </span>
                  <UBadge color="error" variant="subtle" size="xs">
                    Very Significant: {{ countCategory('Very Significant') }}
                  </UBadge>
                  <UBadge color="warning" variant="subtle" size="xs">
                    Significant: {{ countCategory('Significant') }}
                  </UBadge>
                  <UBadge color="info" variant="subtle" size="xs">
                    Quite Significant: {{ countCategory('Quite Significant') }}
                  </UBadge>
                  <UBadge color="success" variant="subtle" size="xs">
                    Not Significant: {{ countCategory('Not Significant') }}
                  </UBadge>
                </div>
                
                <!-- Findings Cards -->
                <div class="space-y-3">
                  <div
                    v-for="(finding, idx) in store.reportForm.findings"
                    :key="idx"
                    class="flex items-start gap-3 bg-slate-50 dark:bg-slate-850 p-4 rounded-xl border border-slate-200 dark:border-slate-800 transition shadow-sm hover:border-primary-200 dark:hover:border-primary-800"
                  >
                    <div class="flex-1 space-y-3">
                      <!-- Card Header (Index & Source Badge) -->
                      <div class="flex justify-between items-center border-b border-slate-100 dark:border-slate-800 pb-2">
                        <span class="text-xs font-bold text-gray-500 uppercase tracking-wider">
                          Temuan #{{ idx + 1 }}
                        </span>
                        <UBadge
                          :color="getSourceColor(finding.source)"
                          variant="subtle"
                          size="xs"
                          class="flex items-center gap-1 font-medium"
                        >
                          <UIcon :name="getSourceIcon(finding.source)" class="size-3" />
                          {{ finding.source || 'Manual' }}
                        </UBadge>
                      </div>

                      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                        <UFormField label="Category / Tingkat Signifikansi" size="sm" class="md:col-span-1">
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
                            maxlength="255"
                          />
                          <div class="text-xs text-gray-500 mt-1 text-right">
                            {{ finding.title ? finding.title.length : 0 }}/255
                          </div>
                        </UFormField>

                        <UFormField label="Action / Tindak Lanjut Rekomendasi" size="sm" class="md:col-span-2">
                          <UTextarea
                            v-model="finding.action"
                            placeholder="e.g. Evaluasi SOP dan pelatihan ulang tim kasir"
                            :rows="2"
                            class="w-full"
                          />
                        </UFormField>
                      </div>
                    </div>

                    <UTooltip text="Hapus Temuan">
                      <UButton
                        color="error"
                        variant="ghost"
                        icon="i-heroicons-trash"
                        size="sm"
                        class="mt-1"
                        @click="removeFinding(idx)"
                      />
                    </UTooltip>
                  </div>

                  <!-- Empty State in Findings List -->
                  <div v-if="!store.reportForm.findings || store.reportForm.findings.length === 0" class="text-center py-8 bg-slate-50 dark:bg-slate-850/50 rounded-xl border border-dashed border-slate-300 dark:border-slate-800 space-y-2">
                    <div class="p-3 bg-primary-50 dark:bg-primary-950/40 rounded-full w-fit mx-auto text-primary-600">
                      <UIcon name="i-heroicons-sparkles" class="size-6" />
                    </div>
                    <p class="text-sm font-semibold text-slate-700 dark:text-slate-200">
                      Belum ada temuan audit di dalam laporan ini
                    </p>
                    <p class="text-xs text-slate-500 max-w-md mx-auto">
                      Klik <strong>"Tarik Temuan Otomatis"</strong> untuk mengambil temuan dari Digital Working Paper (KKA) dan Fieldwork Test Controls penugasan ini.
                    </p>
                    <div class="flex justify-center gap-2 pt-2">
                      <UButton
                        label="Tarik Temuan Otomatis"
                        size="sm"
                        color="primary"
                        icon="i-heroicons-sparkles"
                        :loading="store.isAutoDetecting"
                        @click="store.runAutoDetectFindings('replace')"
                      />
                      <UButton
                        label="Tambah Manual"
                        size="sm"
                        color="neutral"
                        variant="outline"
                        icon="i-heroicons-plus"
                        @click="addManualFinding"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Footer Actions -->
            <div class="flex justify-end gap-3 pt-6 border-t border-gray-100 dark:border-gray-800">
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
                :loading="store.loading"
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

const countCategory = (cat: string) => {
  if (!store.reportForm.findings) return 0
  return store.reportForm.findings.filter((f: any) => f.category === cat).length
}

const getSourceColor = (source?: string) => {
  if (!source) return 'neutral'
  if (source.includes('KKA') || source.includes('Working Paper')) return 'primary'
  if (source.includes('Fieldwork')) return 'info'
  if (source.includes('ATR') || source.includes('Action Taken')) return 'warning'
  return 'neutral'
}

const getSourceIcon = (source?: string) => {
  if (!source) return 'i-heroicons-pencil'
  if (source.includes('KKA') || source.includes('Working Paper')) return 'i-heroicons-document-text'
  if (source.includes('Fieldwork')) return 'i-heroicons-shield-check'
  if (source.includes('ATR') || source.includes('Action Taken')) return 'i-heroicons-arrow-path'
  return 'i-heroicons-pencil'
}

const addManualFinding = () => {
  if (!store.reportForm.findings) store.reportForm.findings = []
  store.reportForm.findings.push({
    title: '',
    category: 'Significant',
    action: '',
    source: 'Manual'
  })
  store.reportForm.findingsCount = store.reportForm.findings.length
}

const removeFinding = (idx: number) => {
  if (store.reportForm.findings) {
    store.reportForm.findings.splice(idx, 1)
    store.reportForm.findingsCount = store.reportForm.findings.length
  }
}

const onLetterChange = async (newVal: string) => {
  if (!store.isEditing && newVal) {
    await store.autoPopulateFindings(newVal)
  }
}
</script>

