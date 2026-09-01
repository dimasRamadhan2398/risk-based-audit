<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen min-w-full">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/quality-assurance" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Import SAIV</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Import external Self Assessment with Independent Validation (SAIV) documents</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex flex-col gap-10">
      <!-- Import Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-upload" class="w-5 h-5 text-warning" />
              Import SAIV Document
            </h3>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-6">
            <UFormField label="Document Title" required>
              <UInput 
                v-model="form.title" 
                placeholder="Ex: Laporan SAIV Document 2026" 
                class="w-full"
                required
              />
            </UFormField>

            <UFormField label="Description">
              <UTextarea 
                v-model="form.description" 
                placeholder="Brief description of the document..." 
                class="w-full"
              />
            </UFormField>

            <div class="space-y-2 pt-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Import Document File *</label>
              <div 
                @click="triggerFileSelect"
                @dragover.prevent="isDragging = true"
                @dragleave.prevent="isDragging = false"
                @drop.prevent="handleFileDrop"
                class="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors duration-200"
                :class="[
                  isDragging 
                    ? 'border-warning bg-orange-50/50 dark:bg-orange-950/30' 
                    : form.fileName 
                      ? 'border-emerald-400 bg-emerald-50/30 dark:border-emerald-500 dark:bg-emerald-950/20' 
                      : 'border-gray-300 dark:border-gray-700 hover:border-warning bg-gray-50 dark:bg-gray-800/60'
                ]"
              >
                <input 
                  type="file" 
                  ref="fileInput" 
                  class="hidden" 
                  @change="handleFileSelect"
                  accept=".pdf,.docx,.doc,.xls,.xlsx"
                />
                
                <div v-if="!form.fileName" class="space-y-3">
                  <UIcon name="i-lucide-file-up" class="w-10 h-10 mx-auto text-gray-400" />
                  <div>
                    <p class="text-sm text-gray-600 dark:text-gray-300 font-semibold">Click to upload or drag & drop</p>
                    <p class="text-md text-gray-400 mt-1">PDF, DOC, DOCX, XLSX up to 10MB</p>
                  </div>
                </div>

                <div v-else class="space-y-3">
                  <UIcon name="i-lucide-file-check-2" class="w-10 h-10 mx-auto text-emerald-500" />
                  <div>
                    <p class="text-sm text-emerald-700 dark:text-emerald-400 font-bold truncate max-w-[200px] mx-auto px-2">
                      {{ form.fileName }}
                    </p>
                    <p class="text-md text-emerald-600 dark:text-emerald-500 mt-1">
                      {{ formatBytes(selectedFileLength) }}
                    </p>
                  </div>
                  <button 
                    type="button" 
                    @click.stop="clearFile" 
                    class="text-md text-red-500 hover:underline font-bold mt-2 block mx-auto"
                  >
                    Remove File
                  </button>
                </div>
              </div>
            </div>

            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200 dark:border-red-900">
              {{ store.errorMsg }}
            </div>

            <UButton 
              type="submit" 
              label="Import Document" 
              color="warning" 
              class="w-full justify-center font-bold h-11 text-base" 
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            />
          </form>
        </UCard>
      </div>

      <!-- Imported Documents Table -->
      <div class="w-full">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200 dark:border-gray-800 h-full">
          <template #header>
            <div class="flex justify-between items-center">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <UIcon name="i-lucide-list" class="w-5 h-5 text-warning" />
                Imported SAIV Documents
              </h3>
              <UBadge color="warning" variant="subtle">
                {{ store.saivImportedReports.length }} Documents
              </UBadge>
            </div>
          </template>

          <TableEntities
            :data="store.saivImportedReports"
            :columns="columns"
            :loading="store.loading"
            :empty-state="{
              icon: 'i-lucide-folder-open',
              label: 'No documents imported',
              description: 'Import SAIV documents to add them to your records.'
            }"
          >
            <template #title-cell="{ row }">
              <div>
                <div class="font-bold text-gray-900 dark:text-white text-sm">{{ row.original.assessmentTitle }}</div>
                <div class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1" v-if="row.original.conductedBy">
                  {{ row.original.conductedBy }}
                </div>
              </div>
            </template>

            <template #fileName-cell="{ row }">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-file-text" class="w-4 h-4 text-gray-400" />
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate max-w-[200px] block">
                  {{ row.original.attachment ? row.original.attachment.name : 'SAIV_Document.pdf' }}
                </span>
              </div>
            </template>

            <template #created_at-cell="{ row }">
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ formatDate(row.original.created_at || row.original.period) }}
              </span>
            </template>

            <template #actions-cell="{ row }">
              <div class="flex items-center gap-1">
                <UButton 
                  icon="i-lucide-eye" 
                  color="neutral" 
                  variant="ghost" 
                  size="md" 
                  title="View Document"
                  @click="store.viewDocument(row.original.id, row.original.fileName)" 
                />
                <UButton 
                  icon="i-lucide-download" 
                  color="success" 
                  variant="ghost" 
                  size="md" 
                  title="Download Document"
                  @click="store.downloadAttachment(row.original.id, row.original.attachment ? row.original.attachment.name : 'document.pdf')" 
                />
                <UButton 
                  icon="i-lucide-trash-2" 
                  color="error" 
                  variant="ghost" 
                  size="md" 
                  title="Delete Document"
                  @click="handleDelete(row.original)" 
                />
              </div>
            </template>
          </TableEntities>
        </UCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType, QAStatus, type QAReport } from '~/types/quality-assurance'
import TableEntities from '~/components/shared/TableEntities.vue'

const store = useQualityAssuranceStore()

onMounted(() => {
  store.fetchReports()
})

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const selectedFileLength = ref(0)

const form = ref({
  title: '',
  description: '',
  fileName: '',
  fileType: '',
  file: null as any
})

const columns = [
  { accessorKey: 'title', header: 'Document Title', class: 'w-[50%]' },
  { accessorKey: 'fileName', header: 'File', class: 'w-[24%]' },
  { accessorKey: 'created_at', header: 'Imported Date', class: 'w-[16%]' },
  { accessorKey: 'actions', header: 'Actions', class: 'w-[10%]' }
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

  form.value.file = file
}

const clearFile = () => {
  form.value.fileName = ''
  form.value.fileType = ''
  form.value.file = null as any
  selectedFileLength.value = 0
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

const handleUpload = async () => {
  if (!form.value.title || !form.value.fileName) return

  try {
    await store.importQARReport({
      assessmentTitle: form.value.title,
      type: QAType.SAIV,
      periodQuarter: 'Q1',
      periodYear: '2026',
      result: 'Generally Conformed',
      status: QAStatus.COMPLETED,
      conductedBy: form.value.description || 'Independent Validator',
      validator: 'Independent Validator',
      fileName: form.value.fileName,
      fileType: form.value.fileType,
      file: form.value.file
    })
    
    if (!store.errorMsg) {
      form.value.title = ''
      form.value.description = ''
      clearFile()
    }
  } catch (error) {
    // Error handled in store
  }
}

const handleDelete = async (report: QAReport) => {
  if (await useGlobalModalStore().confirmDelete({ description: 'Are you sure you want to delete this imported document?' })) {
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

const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  if (isNaN(date.getTime())) return dateString
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  }).format(date)
}
</script>
