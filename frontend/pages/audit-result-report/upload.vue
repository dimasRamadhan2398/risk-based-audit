<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 min-h-screen">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/audit-result-report" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Upload Laporan Hasil Audit (LHA)</h1>
        <p class="text-sm text-gray-500">Upload external Laporan Hasil Audit (Audit Result Report) documents</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex flex-col gap-10">
      <!-- Upload Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
              <UIcon name="i-lucide-upload" class="w-5 h-5 text-primary" />
              Upload LHA Document
            </h3>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-6">
            <UFormField label="Document Title" required>
              <UInput 
                v-model="form.title" 
                placeholder="Ex: LHA Audit Operasional 2026" 
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
              <label class="block text-sm font-medium text-gray-700">Upload Document File *</label>
              <div 
                @click="triggerFileSelect"
                @dragover.prevent="isDragging = true"
                @dragleave.prevent="isDragging = false"
                @drop.prevent="handleFileDrop"
                class="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors duration-200"
                :class="[
                  isDragging 
                    ? 'border-primary bg-blue-50/50' 
                    : form.fileName 
                      ? 'border-emerald-400 bg-emerald-50/30' 
                      : 'border-gray-300 hover:border-primary bg-white'
                ]"
              >
                <input 
                  type="file" 
                  ref="fileInput" 
                  class="hidden" 
                  @change="handleFileSelect"
                  accept=".pdf,.docx,.doc"
                />
                
                <div v-if="!form.fileName" class="space-y-3">
                  <UIcon name="i-lucide-file-up" class="w-10 h-10 mx-auto text-gray-400" />
                  <div>
                    <p class="text-sm text-gray-600 font-semibold">Click to upload or drag & drop</p>
                    <p class="text-md text-gray-400 mt-1">PDF, DOC, DOCX up to 10MB</p>
                  </div>
                </div>

                <div v-else class="space-y-3">
                  <UIcon name="i-lucide-file-check-2" class="w-10 h-10 mx-auto text-emerald-500" />
                  <div>
                    <p class="text-sm text-emerald-700 font-bold truncate max-w-[200px] mx-auto px-2">
                      {{ form.fileName }}
                    </p>
                    <p class="text-md text-emerald-600 mt-1">
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

            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 p-3 rounded-lg border border-red-200">
              {{ store.errorMsg }}
            </div>

            <UButton 
              type="submit"
              label="Upload Document" 
              color="primary" 
              class="w-full justify-center font-bold h-11 text-base" 
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            />
          </form>
        </UCard>
      </div>

      <!-- Uploaded Documents Table -->
      <div class="w-full">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200 h-full">
          <template #header>
            <div class="flex justify-between items-center">
              <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                <UIcon name="i-lucide-list" class="w-5 h-5 text-primary" />
                Uploaded LHA Documents
              </h3>
              <UBadge color="primary" variant="subtle">
                {{ store.uploadedDocuments.length }} Documents
              </UBadge>
            </div>
          </template>

          <div v-if="store.loading && store.uploadedDocuments.length === 0" class="py-12 text-center">
            <UIcon name="i-lucide-loader-2" class="w-8 h-8 text-primary animate-spin mx-auto mb-2" />
            <p class="text-gray-500 text-sm">Loading documents...</p>
          </div>

          <div v-else-if="store.uploadedDocuments.length === 0" class="py-16 text-center space-y-4 rounded-lg border-2 border-dashed border-gray-100 bg-gray-50/50">
            <div class="inline-flex p-4 rounded-full text-gray-300">
              <UIcon name="i-lucide-folder-open" class="w-12 h-12" />
            </div>
            <div class="max-w-md mx-auto">
              <h3 class="text-sm font-bold text-gray-900">No documents uploaded</h3>
              <p class="text-md text-gray-500 mt-1">Upload LHA documents to add them to your audit records.</p>
            </div>
          </div>

          <div v-else class="overflow-x-auto">
            <UTable :data="store.uploadedDocuments" :columns="columns">
              <template #title-cell="{ row }">
                <div>
                  <div class="font-bold text-gray-900 text-sm">{{ row.original.title }}</div>
                  <div class="text-[11px] text-gray-500 mt-0.5 line-clamp-1" v-if="row.original.description">
                    {{ row.original.description }}
                  </div>
                </div>
              </template>

              <template #fileName-cell="{ row }">
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-file-text" class="w-4 h-4 text-gray-400" />
                  <span class="text-md text-gray-700 truncate max-w-[150px] block">{{ row.original.fileName }}</span>
                </div>
              </template>

              <template #created_at-cell="{ row }">
                <span class="text-md text-gray-500">
                  {{ formatDate(row.original.created_at) }}
                </span>
              </template>

              <template #actions-cell="{ row }">
                <div class="flex items-center gap-1">
                  <UButton 
                    icon="i-lucide-download" 
                    color="primary" 
                    variant="ghost" 
                    size="sm" 
                    title="Download Document"
                    @click="store.downloadDocument(row.original.id, row.original.fileName)" 
                  />
                  <UButton 
                    icon="i-lucide-trash-2" 
                    color="error" 
                    variant="ghost" 
                    size="sm" 
                    title="Delete Document"
                    @click="handleDelete(row.original.id)" 
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
import { ref, onMounted } from 'vue'
import { useUploadAuditResultReportStore } from '~/stores/upload-audit-result-report'

const store = useUploadAuditResultReportStore()

onMounted(() => {
  store.fetchUploadedDocuments()
})

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const selectedFileLength = ref(0)

const form = ref({
  title: '',
  description: '',
  fileName: '',
  fileType: '',
  fileContent: ''
})

const columns = [
  { accessorKey: 'title', header: 'Document Title' },
  { accessorKey: 'fileName', header: 'File' },
  { accessorKey: 'created_at', header: 'Uploaded Date' },
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

const handleUpload = async () => {
  if (!form.value.title || !form.value.fileName) return

  try {
    await store.uploadDocument({
      title: form.value.title,
      description: form.value.description,
      fileName: form.value.fileName,
      fileType: form.value.fileType,
      fileContent: form.value.fileContent
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

const handleDelete = async (id: string) => {
  if (confirm('Are you sure you want to delete this uploaded document?')) {
    await store.deleteDocument(id)
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
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  }).format(date)
}
</script>
