<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen min-w-full">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/quality-assurance" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('qualityAssurance.importIacm.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('qualityAssurance.importIacm.subtitle') }}</p>
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
              {{ t('qualityAssurance.importIacm.formTitle') }}
            </h3>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-6">
            <UFormField :label="t('qualityAssurance.importIacm.documentTitle')" required>
              <UInput 
                v-model="form.title" 
                :placeholder="t('qualityAssurance.importIacm.documentTitlePlaceholder')" 
                class="w-full"
                required
              />
            </UFormField>

            <UFormField :label="t('qualityAssurance.importIacm.description')">
              <UTextarea 
                v-model="form.description" 
                :placeholder="t('qualityAssurance.importIacm.descriptionPlaceholder')" 
                class="w-full"
              />
            </UFormField>

            <div class="space-y-2 pt-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('qualityAssurance.importIacm.fileLabel') }}</label>
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
                    <p class="text-sm text-gray-600 dark:text-gray-300 font-semibold">{{ t('qualityAssurance.importIacm.dropzonePrompt') }}</p>
                    <p class="text-md text-gray-400 mt-1">{{ t('qualityAssurance.importIacm.dropzoneHint') }}</p>
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
                    {{ t('qualityAssurance.importIacm.removeFile') }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200 dark:border-red-900">
              {{ store.errorMsg }}
            </div>

            <UButton 
              type="submit" 
              :label="t('qualityAssurance.importIacm.submitButton')" 
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
                {{ t('qualityAssurance.importIacm.tableTitle') }}
              </h3>
              <UBadge color="warning" variant="subtle">
                {{ t('qualityAssurance.importIacm.documentsCount', { count: store.iacmImportedReports.length }) }}
              </UBadge>
            </div>
          </template>

          <TableEntities
            :data="store.iacmImportedReports"
            :columns="columns"
            :loading="store.loading"
            :empty-state="{
              icon: 'i-lucide-folder-open',
              label: t('qualityAssurance.importIacm.emptyTitle'),
              description: t('qualityAssurance.importIacm.emptyDesc')
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
                  {{ row.original.attachment ? row.original.attachment.name : 'IACM_Document.pdf' }}
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
                  :title="t('qualityAssurance.importIacm.actions.view')" 
                  @click="store.viewDocument(row.original.id, row.original.fileName)" 
                />
                <UButton 
                  icon="i-lucide-download" 
                  color="success" 
                  variant="ghost" 
                  size="md" 
                  :title="t('qualityAssurance.importIacm.actions.download')" 
                  @click="store.downloadAttachment(row.original.id, row.original.attachment ? row.original.attachment.name : 'document.pdf')" 
                />
                <UButton 
                  icon="i-lucide-trash-2" 
                  color="error" 
                  variant="ghost" 
                  size="md" 
                  :title="t('qualityAssurance.importIacm.actions.delete')" 
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
import { ref, computed, onMounted } from 'vue'
import { useQualityAssuranceStore } from '~/stores/quality-assurance'
import { QAType, QAStatus, type QAReport } from '~/types/quality-assurance'
import { useI18n } from '~/composables/useI18n'
import TableEntities from '~/components/shared/TableEntities.vue'

const { t } = useI18n()
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

const columns = computed(() => [
  { accessorKey: 'title', header: t('qualityAssurance.importIacm.columns.title'), class: 'w-[50%]' },
  { accessorKey: 'fileName', header: t('qualityAssurance.importIacm.columns.file'), class: 'w-[24%]' },
  { accessorKey: 'created_at', header: t('qualityAssurance.importIacm.columns.date'), class: 'w-[16%]' },
  { accessorKey: 'actions', header: t('qualityAssurance.importIacm.columns.actions'), class: 'w-[10%]' }
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
    alert(t('qualityAssurance.importIacm.fileSizeLimit'))
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
      type: QAType.IACM,
      periodQuarter: 'Yearly',
      periodYear: '2026',
      result: 'Level 4 (Managed)',
      status: QAStatus.COMPLETED,
      conductedBy: form.value.description || 'BPKP / Kementerian BUMN',
      validator: 'BPKP Assessor',
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
  if (await useGlobalModalStore().confirmDelete({ description: t('qualityAssurance.importIacm.deleteConfirm') })) {
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
