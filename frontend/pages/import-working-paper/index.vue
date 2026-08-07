<template>
  <div class="p-6 max-w-7xl mx-auto space-y-6 min-h-screen">
    <!-- Header -->
    <div class="flex justify-between items-center mb-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Import Working Paper</h1>
        <p class="text-sm text-gray-500">Import reference working papers for auditor guidelines</p>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="flex flex-col gap-10">
      <!-- Import Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
              <UIcon name="i-heroicons-arrow-up-tray" class="w-5 h-5 text-primary" />
              Import Working Paper
            </h3>
          </template>

          <form @submit.prevent="handleImport" class="space-y-4">
            <UFormField label="Document Title" required>
              <UInput 
                v-model="form.title" 
                placeholder="e.g. IT Audit Guidelines 2026" 
                class="w-full" 
                required 
              />
            </UFormField>

            <UFormField label="Description">
              <UTextarea 
                v-model="form.description" 
                placeholder="Optional description of the document..." 
                :rows="3" 
                class="w-full" 
              />
            </UFormField>

            <!-- File Import Box -->
            <UFormField label="File Reference" required>
              <div 
                @click="triggerFileSelect"
                @dragover.prevent="isDragging = true"
                @dragleave.prevent="isDragging = false"
                @drop.prevent="handleFileDrop"
                class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors duration-200"
                :class="[
                  isDragging 
                    ? 'border-primary bg-primary-50' 
                    : form.fileName 
                      ? 'border-green-400 bg-green-50' 
                      : 'border-gray-300 hover:border-primary'
                ]"
              >
                <input 
                  type="file" 
                  ref="fileInput" 
                  class="hidden" 
                  @change="handleFileSelect"
                  accept=".pdf,.doc,.docx,.xls,.xlsx"
                />
                
                <div v-if="!form.fileName" class="space-y-2">
                  <UIcon name="i-heroicons-document-arrow-up" class="w-8 h-8 mx-auto text-gray-400" />
                  <p class="text-md text-gray-500 font-medium">Click to select or drag & drop</p>
                  <p class="text-[10px] text-gray-400">PDF, DOCX, XLSX up to 10MB</p>
                </div>

                <div v-else class="space-y-2">
                  <UIcon name="i-heroicons-document-check" class="w-8 h-8 mx-auto text-green-500" />
                  <p class="text-md text-green-700 font-bold truncate max-w-full px-2">
                    {{ form.fileName }}
                  </p>
                  <p class="text-[10px] text-green-600">
                    {{ formatBytes(selectedFileLength) }}
                  </p>
                  <button 
                    type="button" 
                    @click.stop="clearFile" 
                    class="text-md text-red-500 hover:underline font-semibold block mx-auto"
                  >
                    Remove File
                  </button>
                </div>
              </div>
            </UFormField>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 p-3 rounded-lg border border-red-200">
              {{ store.errorMsg }}
            </div>

            <!-- Submit Button -->
            <UButton 
              type="submit" 
              color="primary" 
              class="w-full justify-center font-bold" 
              :loading="store.loading"
              icon="i-heroicons-arrow-up-tray"
              :disabled="!form.title || !form.fileName"
            >
              Import Document
            </UButton>
          </form>
        </UCard>
      </div>

      <!-- Imported Files Table -->
      <div class="lg:col-span-2">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200">
          <template #header>
            <div class="flex justify-between items-center">
              <h3 class="text-lg font-bold text-gray-900 flex items-center gap-2">
                <UIcon name="i-heroicons-table-cells" class="w-5 h-5 text-primary" />
                Reference Guidelines Documents
              </h3>
              <UBadge color="primary" variant="subtle">
                {{ store.importedPapers.length }} Documents
              </UBadge>
            </div>
          </template>

          <!-- Loading State -->
          <div v-if="store.loading && store.importedPapers.length === 0" class="py-12 text-center">
            <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 text-primary animate-spin mx-auto mb-2" />
            <p class="text-gray-500 text-sm">Loading guidelines...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="store.importedPapers.length === 0" class="py-16 text-center space-y-4 rounded-lg">
            <div class="inline-flex p-4 rounded-full text-gray-400">
              <UIcon name="i-heroicons-folder-open" class="w-12 h-12" />
            </div>
            <div class="max-w-md mx-auto">
              <h3 class="text-sm font-bold text-gray-900">No guidelines imported</h3>
              <p class="text-md text-gray-500 mt-1">Import files on the left to make reference documents available for auditors.</p>
            </div>
          </div>

          <!-- Table -->
          <div v-else class="overflow-x-auto">
            <UTable :data="store.importedPapers" :columns="columns">
              <template #title-cell="{ row }">
                <div>
                  <div class="font-bold text-gray-900 text-sm">{{ row.original.title }}</div>
                  <div class="text-[11px] text-gray-500 mt-0.5 line-clamp-1">
                    {{ row.original.description || 'No description' }}
                  </div>
                </div>
              </template>

              <template #fileName-cell="{ row }">
                <div class="flex items-center gap-2 max-w-[200px]">
                  <UIcon :name="getFileIcon(row.original.fileName)" class="w-5 h-5 flex-shrink-0" :class="getFileIconColor(row.original.fileName)" />
                  <span class="text-md text-gray-600 truncate" :title="row.original.fileName">{{ row.original.fileName }}</span>
                </div>
              </template>

              <template #fileSize-cell="{ row }">
                <span class="text-md text-gray-500 font-medium">
                  {{ formatBytes(row.original.fileSize) }}
                </span>
              </template>

              <template #created_at-cell="{ row }">
                <span class="text-md text-gray-500">
                  {{ formatDate(row.original.created_at) }}
                </span>
              </template>

              <template #actions-cell="{ row }">
                <div class="flex items-center gap-1">
                  <UButton 
                    icon="i-heroicons-arrow-down-tray" 
                    color="primary" 
                    variant="ghost" 
                    size="sm" 
                    title="Download File"
                    @click="store.downloadImportedPaper(row.original.id, row.original.fileName)" 
                  />
                  <UButton 
                    icon="i-heroicons-trash" 
                    color="error" 
                    variant="ghost" 
                    size="sm" 
                    title="Delete File"
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
import { useImportWorkingPaperStore } from '~/stores/import-working-paper'

const store = useImportWorkingPaperStore()

onMounted(() => {
  store.fetchImportedPapers()
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
  { accessorKey: 'fileName', header: 'File Reference' },
  { accessorKey: 'fileSize', header: 'Size' },
  { accessorKey: 'created_at', header: 'Imported At' },
  { accessorKey: 'actions', header: 'Actions' }
]

onMounted(async () => {
  await store.fetchImportedPapers()
})

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
  try {
    await store.importWorkingPaper({
      title: form.value.title,
      description: form.value.description,
      fileName: form.value.fileName,
      fileType: form.value.fileType,
      fileContent: form.value.fileContent
    })
    
    // Reset Form
    form.value.title = ''
    form.value.description = ''
    clearFile()
  } catch (error) {
    // Error is handled in the store
  }
}

const handleDelete = async (id: string) => {
  if (confirm('Are you sure you want to delete this guidelines document?')) {
    try {
      await store.deleteImportedPaper(id)
    } catch (err) {
      alert('Failed to delete file.')
    }
  }
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (e) {
    return dateStr
  }
}

const getFileIcon = (name: string) => {
  const ext = name.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'pdf':
      return 'i-heroicons-document-text'
    case 'xls':
    case 'xlsx':
      return 'i-heroicons-table-cells'
    case 'doc':
    case 'docx':
      return 'i-heroicons-document-text'
    default:
      return 'i-heroicons-document'
  }
}

const getFileIconColor = (name: string) => {
  const ext = name.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'pdf':
      return 'text-red-500'
    case 'xls':
    case 'xlsx':
      return 'text-green-600'
    case 'doc':
    case 'docx':
      return 'text-blue-500'
    default:
      return 'text-gray-400'
  }
}
</script>
