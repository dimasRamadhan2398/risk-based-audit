<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 min-h-screen">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/quality-assurance" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Import QAR Report</h1>
        <p class="text-sm text-gray-500">Upload External Quality Assessment Review (QAR) Reports</p>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="flex flex-col gap-10">
      <!-- Import Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
              <UIcon name="i-lucide-upload" class="w-5 h-5 text-warning" />
              Import Assessment Document
            </h3>
          </template>

          <form @submit.prevent="handleImport" class="space-y-6">
            <!-- General Info -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700 border-b pb-2">1. General Information</h4>
              <div class="space-y-4 pt-2">
                <UFormField label="Assessment Title" required>
                  <UInput 
                    v-model="form.assessmentTitle" 
                    placeholder="Ex: External QAR Report 2026" 
                    class="w-full"
                    required
                  />
                </UFormField>

                <div class="grid grid-cols-2 gap-4">
                  <UFormField label="Execution Period">
                    <USelectMenu 
                      v-model="form.periodQuarter" 
                      :items="['Q1', 'Q2', 'Q3', 'Q4']" 
                      placeholder="Select Quarter" 
                      class="w-full"
                    />
                  </UFormField>
                  <UFormField label="&nbsp;">
                    <USelectMenu 
                      v-model="form.periodYear" 
                      :items="store.periods" 
                      placeholder="Select Year" 
                      class="w-full"
                    />
                  </UFormField>
                </div>
              </div>
            </div>

            <!-- Results & Consultant -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700 border-b pb-2">2. Results & Validator</h4>
              <div class="space-y-4 pt-2">
                <UFormField label="Result/Score" required>
                  <USelectMenu 
                    v-model="form.result" 
                    :items="['Does not Conform', 'Partially Conform', 'Generally Conformed', 'Fully Conformance']" 
                    placeholder="Select Conclusion" 
                    class="w-full"
                  />
                </UFormField>
                <UFormField label="Professional Consultant (Validator)" required>
                  <UInput 
                    v-model="form.validator" 
                    placeholder="Ex: Ernst & Young, PwC" 
                    class="w-full"
                    required
                  />
                </UFormField>
                <UFormField label="Conducted By" required>
                  <UInput 
                    v-model="form.conductedBy" 
                    placeholder="Ex: PT Nama Perusahaan" 
                    class="w-full"
                    required
                  />
                </UFormField>
              </div>
            </div>

            <!-- Supporting Documents / File Upload -->
            <div class="space-y-4">
              <h4 class="font-bold text-gray-700 border-b pb-2">3. Upload Assessment File</h4>
              <div class="pt-2">
                <div 
                  @click="triggerFileSelect"
                  @dragover.prevent="isDragging = true"
                  @dragleave.prevent="isDragging = false"
                  @drop.prevent="handleFileDrop"
                  class="border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors duration-200"
                  :class="[
                    isDragging 
                      ? 'border-warning bg-orange-50/50' 
                      : form.fileName 
                        ? 'border-emerald-400 bg-emerald-50/30' 
                        : 'border-gray-300 hover:border-warning bg-white'
                  ]"
                >
                  <input 
                    type="file" 
                    ref="fileInput" 
                    class="hidden" 
                    @change="handleFileSelect"
                    accept=".pdf,.docx,.doc"
                  />
                  
                  <div v-if="!form.fileName" class="space-y-2">
                    <UIcon name="i-lucide-file-up" class="w-8 h-8 mx-auto text-gray-400" />
                    <p class="text-md text-gray-600 font-semibold">Click to upload or drag & drop</p>
                    <p class="text-[10px] text-gray-400">PDF, DOC, DOCX up to 10MB</p>
                  </div>

                  <div v-else class="space-y-2">
                    <UIcon name="i-lucide-file-check-2" class="w-8 h-8 mx-auto text-emerald-500" />
                    <p class="text-md text-emerald-700 font-bold truncate max-w-full px-2">
                      {{ form.fileName }}
                    </p>
                    <p class="text-[10px] text-emerald-600">
                      {{ formatBytes(selectedFileLength) }}
                    </p>
                    <button 
                      type="button" 
                      @click.stop="clearFile" 
                      class="text-[10px] text-red-500 hover:underline font-bold mt-1 block mx-auto"
                    >
                      Remove File
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 p-3 rounded-lg border border-red-200">
              {{ store.errorMsg }}
            </div>

            <UButton 
              type="submit"
              label="Import Report" 
              color="warning" 
              class="w-full justify-center font-bold" 
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.assessmentTitle || !form.fileName || !form.validator || !form.result || !form.conductedBy"
            />
          </form>
        </UCard>
      </div>

      <!-- Imported QAR Reports Table -->
      <div class="lg:col-span-7">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200 h-full">
          <template #header>
            <div class="flex justify-between items-center">
              <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                <UIcon name="i-lucide-list" class="w-5 h-5 text-warning" />
                Imported QAR Reports
              </h3>
              <UBadge color="warning" variant="subtle">
                {{ store.importedReports.length }} Reports
              </UBadge>
            </div>
          </template>

          <!-- Loading State -->
          <div v-if="store.loading && store.importedReports.length === 0" class="py-12 text-center">
            <UIcon name="i-lucide-loader-2" class="w-8 h-8 text-warning animate-spin mx-auto mb-2" />
            <p class="text-gray-500 text-sm">Loading reports...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="store.importedReports.length === 0" class="py-16 text-center space-y-4 rounded-lg border-2 border-dashed border-gray-100 bg-gray-50/50">
            <div class="inline-flex p-4 rounded-full text-gray-300">
              <UIcon name="i-lucide-folder-open" class="w-12 h-12" />
            </div>
            <div class="max-w-md mx-auto">
              <h3 class="text-sm font-bold text-gray-900">No imported reports</h3>
              <p class="text-md text-gray-500 mt-1">Upload QAR reports on the left to add them to your quality assurance records.</p>
            </div>
          </div>

          <!-- Table -->
          <div v-else class="overflow-x-auto">
            <UTable :data="store.importedReports" :columns="importColumns">
              <template #assessmentTitle-cell="{ row }">
                <div>
                  <div class="font-bold text-gray-900 text-sm">{{ row.original.assessmentTitle }}</div>
                  <div class="text-[11px] text-gray-500 mt-0.5 line-clamp-1">
                    Validator: {{ row.original.validator || '-' }}
                  </div>
                </div>
              </template>

              <template #period-cell="{ row }">
                <UBadge color="neutral" variant="subtle" size="sm">
                  {{ row.original.period }}
                </UBadge>
              </template>
              
              <template #result-cell="{ row }">
                <UBadge 
                  :color="row.original.result === 'Does not Conform' ? 'error' : row.original.result === 'Partially Conform' ? 'warning' : 'success'" 
                  variant="subtle" 
                  size="sm"
                >
                  {{ row.original.result }}
                </UBadge>
              </template>

              <template #conductedBy-cell="{ row }">
                <span class="text-md font-medium text-gray-700">
                  {{ row.original.conductedBy }}
                </span>
              </template>

              <template #actions-cell="{ row }">
                <div class="flex items-center gap-1">
                  <UButton 
                    v-if="row.original.attachment"
                    icon="i-lucide-download" 
                    color="primary" 
                    variant="ghost" 
                    size="sm" 
                    title="Download Report"
                    @click="store.downloadAttachment(row.original.id, row.original.attachment.name)" 
                  />
                  <UButton 
                    icon="i-lucide-trash-2" 
                    color="error" 
                    variant="ghost" 
                    size="sm" 
                    title="Delete Report"
                    @click="handleDelete(row.original)" 
                  />
                </div>
              </template>
            </UTable>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType, QAStatus, type QAReport } from '~/types/quality-assurance'

const store = useQualityAssuranceStore()

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const selectedFileLength = ref(0)

const form = ref({
  assessmentTitle: '',
  periodQuarter: 'Q1',
  periodYear: '2025',
  result: 'Generally Conformed',
  validator: '',
  conductedBy: '',
  fileName: '',
  fileType: '',
  fileContent: ''
})

const importColumns = [
  { accessorKey: 'assessmentTitle', header: 'Assessment Title' },
  { accessorKey: 'period', header: 'Period' },
  { accessorKey: 'result', header: 'Result/Score' },
  { accessorKey: 'conductedBy', header: 'Conducted By' },
  { accessorKey: 'actions', header: 'Actions' }
]

const triggerFileSelect = () => {
  fileInput.value?.click()
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    processFile(file)
  }
}

const handleFileDrop = (event: DragEvent) => {
  isDragging.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    processFile(file)
  }
}

const processFile = (file: File) => {
  if (file.size > 10 * 1024 * 1024) {
    alert('File size exceeds the 10MB limit.')
    return
  }

  form.value.fileName = file.name
  form.value.fileType = file.type || 'application/octet-stream'
  selectedFileLength.value = file.size

  const reader = new FileReader()
  reader.onload = () => {
    form.value.fileContent = reader.result as string
  }
  reader.readAsDataURL(file)
}

const clearFile = () => {
  form.value.fileName = ''
  form.value.fileType = ''
  form.value.fileContent = ''
  selectedFileLength.value = 0
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const handleImport = async () => {
  if (!form.value.assessmentTitle || !form.value.fileName || !form.value.validator || !form.value.conductedBy) return

  try {
    await store.importQARReport({
      assessmentTitle: form.value.assessmentTitle,
      type: QAType.QAR,
      periodQuarter: form.value.periodQuarter,
      periodYear: form.value.periodYear,
      result: form.value.result,
      status: QAStatus.COMPLETED,
      conductedBy: form.value.conductedBy,
      validator: form.value.validator,
      fileName: form.value.fileName,
      fileType: form.value.fileType,
      fileContent: form.value.fileContent
    })
    
    if (!store.errorMsg) {
      // Reset form instead of redirecting
      form.value.assessmentTitle = ''
      form.value.validator = ''
      form.value.conductedBy = ''
      clearFile()
    }
  } catch (error) {
    // Error is stored in store
  }
}

const handleDelete = async (report: QAReport) => {
  if (confirm('Are you sure you want to delete this imported report?')) {
    store.selectedReport = report
    await store.deleteReport()
  }
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>
