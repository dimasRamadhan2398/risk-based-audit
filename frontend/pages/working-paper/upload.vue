<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen min-w-full">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/working-paper" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('workingPaper.upload.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('workingPaper.upload.subtitle') }}</p>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="flex flex-col gap-10">
      <!-- Import Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-upload" class="w-5 h-5 text-primary" />
              {{ t('workingPaper.upload.formTitle') }}
            </h3>
          </template>

          <form @submit.prevent="handleImport" class="space-y-4">
            <UFormField :label="t('workingPaper.upload.documentTitle')" required>
              <UInput 
                v-model="form.title" 
                :placeholder="t('workingPaper.upload.documentTitlePlaceholder')" 
                class="w-full" 
                required 
              />
            </UFormField>

            <UFormField :label="t('workingPaper.upload.description')">
              <UTextarea 
                v-model="form.description" 
                :placeholder="t('workingPaper.upload.descriptionPlaceholder')" 
                :rows="3" 
                class="w-full" 
              />
            </UFormField>

            <!-- File Upload Box -->
            <UFormField :label="t('workingPaper.upload.fileLabel')" required>
              <div 
                @click="triggerFileSelect"
                @dragover.prevent="isDragging = true"
                @dragleave.prevent="isDragging = false"
                @drop.prevent="handleFileDrop"
                class="border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors duration-200"
                :class="[
                  isDragging 
                    ? 'border-primary bg-blue-50/50 dark:bg-primary-950/30' 
                    : form.fileName 
                      ? 'border-emerald-400 bg-emerald-50/30 dark:border-emerald-500 dark:bg-emerald-950/20' 
                      : 'border-gray-300 dark:border-gray-700 hover:border-primary bg-gray-50 dark:bg-gray-800/60'
                ]"
              >
                <input 
                  type="file" 
                  ref="fileInput" 
                  class="hidden" 
                  @change="handleFileSelect"
                  accept=".pdf,.doc,.docx,.xls,.xlsx"
                />
                
                <div v-if="!form.fileName" class="space-y-3">
                  <UIcon name="i-lucide-file-up" class="w-10 h-10 mx-auto text-gray-400" />
                  <div>
                    <p class="text-sm text-gray-600 dark:text-gray-300 font-semibold">{{ t('workingPaper.upload.dropzonePrompt') }}</p>
                    <p class="text-md text-gray-400 mt-1">{{ t('workingPaper.upload.dropzoneHint') }}</p>
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
                    {{ t('workingPaper.upload.removeFile') }}
                  </button>
                </div>
              </div>
            </UFormField>

            <!-- Error message display -->
            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200 dark:border-red-900">
              {{ store.errorMsg }}
            </div>

            <!-- Submit Button -->
            <UButton 
              type="submit" 
              color="primary" 
              class="w-full justify-center font-bold h-11 text-base" 
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            >
              {{ t('workingPaper.upload.submitButton') }}
            </UButton>
          </form>
        </UCard>
      </div>

      <!-- Imported Files Table -->
      <div class="w-full">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200 dark:border-gray-800 h-full">
          <template #header>
            <div class="flex justify-between items-center">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <UIcon name="i-lucide-list" class="w-5 h-5 text-primary" />
                {{ t('workingPaper.upload.tableTitle') }}
              </h3>
              <UBadge color="primary" variant="subtle">
                {{ t('workingPaper.upload.documentsCount', { count: store.importedPapers.length }) }}
              </UBadge>
            </div>
          </template>

          <TableEntities
            :data="store.importedPapers"
            :columns="columns"
            :loading="store.loading"
            :empty-state="{
              icon: 'i-lucide-folder-open',
              label: t('workingPaper.upload.emptyTitle'),
              description: t('workingPaper.upload.emptyDesc')
            }"
          >
            <template #title-cell="{ row }">
              <div>
                <div class="font-bold text-gray-900 dark:text-white text-sm">{{ row.original.title }}</div>
                <div class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5 line-clamp-1">
                  {{ row.original.description || '-' }}
                </div>
              </div>
            </template>

            <template #fileName-cell="{ row }">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-file-text" class="w-4 h-4 text-gray-400" />
                <span class="text-sm text-gray-700 dark:text-gray-300 truncate max-w-[200px] block" :title="row.original.fileName">{{ row.original.fileName }}</span>
              </div>
            </template>

            <template #fileSize-cell="{ row }">
              <span class="text-sm text-gray-500 dark:text-gray-400 font-medium">
                {{ formatBytes(row.original.fileSize) }}
              </span>
            </template>

            <template #created_at-cell="{ row }">
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ formatDate(row.original.created_at) }}
              </span>
            </template>

            <template #actions-cell="{ row }">
              <div class="flex items-center gap-1">
                <UButton 
                  icon="i-lucide-eye" 
                  color="info" 
                  variant="ghost" 
                  size="sm" 
                  :title="t('workingPaper.upload.actions.view')" 
                  @click="store.viewDocument(row.original.id, row.original.fileName)" 
                />
                <UButton 
                  icon="i-lucide-download" 
                  color="primary" 
                  variant="ghost" 
                  size="sm" 
                  :title="t('workingPaper.upload.actions.download')" 
                  @click="store.downloadImportedPaper(row.original.id, row.original.fileName)" 
                />
                <UButton 
                  icon="i-lucide-trash-2" 
                  color="error" 
                  variant="ghost" 
                  size="sm" 
                  :title="t('workingPaper.upload.actions.delete')" 
                  @click="handleDelete(row.original.id)" 
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
import { ref, computed, onMounted } from 'vue'
import { useImportWorkingPaperStore } from '~/stores/import-working-paper'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'

const { t } = useI18n()
const store = useImportWorkingPaperStore()
const toast = useToastNotification()

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

const columns = computed(() => [
  { accessorKey: 'title', header: t('workingPaper.upload.columns.title'), class: 'w-[42%]' },
  { accessorKey: 'fileName', header: t('workingPaper.upload.columns.file'), class: 'w-[22%]' },
  { accessorKey: 'fileSize', header: t('workingPaper.upload.columns.size'), class: 'w-[13%]' },
  { accessorKey: 'created_at', header: t('workingPaper.upload.columns.date'), class: 'w-[13%]' },
  { accessorKey: 'actions', header: t('workingPaper.upload.columns.actions'), class: 'w-[10%]' }
])

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
    toast.showError(t('workingPaper.upload.fileSizeLimit'))
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
    
    toast.showSuccess('Document imported successfully!')
    form.value.title = ''
    form.value.description = ''
    clearFile()
  } catch (error: any) {
    toast.showError(error?.data?.message || 'Failed to import document.')
  }
}

const handleDelete = async (id: string) => {
  if (confirm(t('workingPaper.upload.deleteConfirm'))) {
    try {
      await store.deleteImportedPaper(id)
      toast.showSuccess('Document deleted successfully!')
    } catch (err) {
      toast.showError('Failed to delete file.')
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
</script>
