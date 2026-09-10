<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useUploadPerformanceReportStore } from '~/stores/upload-performance-report'
import { useI18n } from '~/composables/useI18n'
import { useToast } from '#imports'
import TableEntities from '~/components/shared/TableEntities.vue'

const { t, locale } = useI18n()
const store = useUploadPerformanceReportStore()
const toast = useToast()

const selectedPeriodFilter = ref('Semua')
const selectedYearFilter = ref('2026')
const periodFilterOptions = computed(() => [
  { label: t('kpiPerformance.upload.filterAll'), value: 'Semua' },
  { label: 'Q1', value: 'Q1' },
  { label: 'Q2', value: 'Q2' },
  { label: 'Q3', value: 'Q3' },
  { label: 'Q4', value: 'Q4' },
  { label: t('kpiPerformance.upload.annual'), value: 'Tahunan' }
])
const yearOptions = [2024, 2025, 2026, 2027, 2028]

const form = ref({
  title: '',
  period: 'Q1',
  year: 2026,
  description: '',
  fileName: '',
  fileType: '',
  file: null as any
})

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const selectedFileLength = ref(0)

onMounted(() => {
  store.fetchUploadedReports(selectedPeriodFilter.value, parseInt(selectedYearFilter.value))
})

watch([selectedPeriodFilter, selectedYearFilter], () => {
  store.fetchUploadedReports(selectedPeriodFilter.value, parseInt(selectedYearFilter.value))
})

const triggerFileSelect = () => {
  fileInput.value?.click()
}

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    processFile(target.files[0])
  }
}

const handleFileDrop = (event: DragEvent) => {
  isDragging.value = false
  if (event.dataTransfer?.files && event.dataTransfer.files[0]) {
    processFile(event.dataTransfer.files[0])
  }
}

const processFile = (file: File) => {
  if (file.size > 10 * 1024 * 1024) {
    toast.add({
      title: t('kpiPerformance.upload.fileTooLarge'),
      description: t('kpiPerformance.upload.fileTooLargeDesc'),
      color: 'error'
    })
    return
  }

  selectedFileLength.value = file.size
  form.value.fileName = file.name
  form.value.fileType = file.type || getExtensionType(file.name)

  if (!form.value.title) {
    const periodName = form.value.period === 'Tahunan' ? t('kpiPerformance.upload.annual') : `Q${form.value.period.replace(/\D/g, '')}`
    form.value.title = `${periodName} ${form.value.year} - ${file.name.replace(/\.[^/.]+$/, '')}`
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    form.value.file = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const getExtensionType = (filename: string) => {
  const ext = filename.split('.').pop()?.toLowerCase()
  if (ext === 'pdf') return 'application/pdf'
  if (ext === 'docx' || ext === 'doc') return 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
  if (ext === 'xlsx' || ext === 'xls') return 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
  return 'application/octet-stream'
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

const formatBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i]
}

const handleUpload = async () => {
  if (!form.value.title || !form.value.fileName || !form.value.period || !form.value.year) {
    toast.add({
      title: t('kpiPerformance.upload.formIncomplete'),
      description: t('kpiPerformance.upload.formIncompleteDesc'),
      color: 'error'
    })
    return
  }

  try {
    await store.uploadReport({
      title: form.value.title,
      period: form.value.period,
      year: form.value.year,
      description: form.value.description,
      fileName: form.value.fileName,
      fileType: form.value.fileType,
      file: form.value.file
    })

    toast.add({
      title: t('kpiPerformance.upload.uploadSuccess'),
      description: t('kpiPerformance.upload.uploadSuccessDesc', {
        period: form.value.period === 'Tahunan' ? t('kpiPerformance.upload.annual') : form.value.period,
        year: form.value.year
      }),
      color: 'success'
    })

    // Reset form
    form.value.title = ''
    form.value.description = ''
    clearFile()
  } catch (err: any) {
    toast.add({
      title: t('kpiPerformance.upload.uploadFailed'),
      description: store.errorMsg || t('kpiPerformance.upload.uploadFailedDesc'),
      color: 'error'
    })
  }
}

const handleDelete = async (id: string, title: string) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('kpiPerformance.upload.deleteConfirm', { title }) })) {
    try {
      await store.deleteReport(id, selectedPeriodFilter.value, parseInt(selectedYearFilter.value))
      toast.add({
        title: t('kpiPerformance.upload.deleteSuccess'),
        description: t('kpiPerformance.upload.deleteSuccessDesc'),
        color: 'success'
      })
    } catch (err) {
      toast.add({
        title: t('kpiPerformance.upload.deleteFailed'),
        description: store.errorMsg || t('kpiPerformance.upload.deleteFailedDesc'),
        color: 'error'
      })
    }
  }
}

const handleDownload = async (id: string, fileName: string) => {
  await store.downloadReport(id, fileName)
}

const getPeriodBadgeColor = (period: string) => {
  switch (period) {
    case 'Q1': return 'info'
    case 'Q2': return 'success'
    case 'Q3': return 'warning'
    case 'Q4': return 'primary'
    case 'Tahunan': return 'neutral'
    default: return 'neutral'
  }
}

const columns = computed(() => [
  { accessorKey: 'title', header: t('kpiPerformance.upload.columns.title'), class: 'w-[38%]' },
  { accessorKey: 'period', header: t('kpiPerformance.upload.columns.period'), class: 'w-[9%]' },
  { accessorKey: 'year', header: t('kpiPerformance.upload.columns.year'), class: 'w-[8%]' },
  { accessorKey: 'fileName', header: t('kpiPerformance.upload.columns.file'), class: 'w-[20%]' },
  { accessorKey: 'created_at', header: t('kpiPerformance.upload.columns.date'), class: 'w-[15%]' },
  { accessorKey: 'actions', header: t('kpiPerformance.upload.columns.actions'), class: 'w-[10%]' }
])
</script>

<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen min-w-full">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/kpi-performance" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('kpiPerformance.upload.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('kpiPerformance.upload.subtitle') }}</p>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="flex flex-col gap-10">
      <!-- Upload Form Card (Left/Top) -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-upload-cloud" class="w-5 h-5 text-primary" />
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('kpiPerformance.upload.formTitle') }}</h3>
            </div>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-5">
            <!-- Period Selector -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-2">{{ t('kpiPerformance.upload.periodLabel') }}</label>
              <div class="grid grid-cols-5 gap-2">
                <button
                  v-for="p in ['Q1', 'Q2', 'Q3', 'Q4', 'Tahunan']"
                  :key="p"
                  type="button"
                  @click="form.period = p"
                  class="py-2.5 px-2 text-md font-bold rounded-lg border transition-all text-center"
                  :class="[
                    form.period === p
                      ? 'border-primary bg-primary/10 text-primary ring-2 ring-primary/30'
                      : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:border-primary/50'
                  ]"
                >
                  {{ p === 'Tahunan' ? t('kpiPerformance.upload.annual') : p }}
                </button>
              </div>
            </div>

            <!-- Year Selector -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">{{ t('kpiPerformance.upload.yearLabel') }}</label>
              <USelect
                v-model="form.year"
                :items="yearOptions"
                class="w-full"
              />
            </div>

            <!-- Document Title -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">{{ t('kpiPerformance.upload.documentTitle') }}</label>
              <UInput
                v-model="form.title"
                :placeholder="t('kpiPerformance.upload.documentTitlePlaceholder')"
                class="w-full"
                required
              />
            </div>

            <!-- Description -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">{{ t('kpiPerformance.upload.description') }}</label>
              <UTextarea
                v-model="form.description"
                :placeholder="t('kpiPerformance.upload.descriptionPlaceholder')"
                class="w-full"
                :rows="2"
              />
            </div>

            <!-- File Upload Drag Zone -->
            <div class="space-y-2 pt-1">
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('kpiPerformance.upload.fileLabel') }}</label>
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
                      ? 'border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20'
                      : 'border-gray-300 dark:border-gray-700 hover:border-primary bg-gray-50 dark:bg-gray-800/60'
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
                    <p class="text-sm font-semibold text-gray-700 dark:text-gray-300">{{ t('kpiPerformance.upload.dropzonePrompt') }}</p>
                    <p class="text-md text-gray-400 mt-1">{{ t('kpiPerformance.upload.dropzoneHint') }}</p>
                  </div>
                </div>

                <div v-else class="space-y-3">
                  <UIcon name="i-lucide-file-check-2" class="w-10 h-10 mx-auto text-emerald-500" />
                  <div>
                    <p class="text-sm font-bold text-emerald-700 dark:text-emerald-400 truncate max-w-[200px] mx-auto px-2">
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
                    {{ t('kpiPerformance.upload.changeFile') }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="store.errorMsg" class="text-md text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200">
              {{ store.errorMsg }}
            </div>

            <UButton
              type="submit"
              :label="t('kpiPerformance.upload.submitButton')"
              color="primary"
              class="w-full justify-center font-bold h-11 text-base"
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            />
          </form>
        </UCard>
      </div>

      <!-- Uploaded Documents Table (Right/Bottom) -->
      <div class="w-full">
        <UCard :ui="{ body: 'p-4' }" class="shadow-sm border border-gray-200 dark:border-gray-800 h-full">
          <template #header>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-file-text" class="w-5 h-5 text-primary" />
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">{{ t('kpiPerformance.upload.tableTitle') }}</h3>
              </div>
              <div class="flex items-center gap-2">
                <!-- Period filter tabs -->
                <USelect
                  v-model="selectedPeriodFilter"
                  :items="periodFilterOptions"
                  value-key="value"
                  label-key="label"
                  class="w-28 text-md"
                />
                <USelect
                  v-model="selectedYearFilter"
                  :items="yearOptions.map(String)"
                  class="w-24 text-md"
                />
              </div>
            </div>
          </template>

          <TableEntities
            :data="store.uploadedReports"
            :columns="columns"
            :loading="store.loading"
            :empty-state="{
              icon: 'i-lucide-folder-open',
              label: t('kpiPerformance.upload.emptyTitle'),
              description: t('kpiPerformance.upload.emptyDesc')
            }"
          >
            <template #title-cell="{ row }">
              <div>
                <div class="font-bold text-gray-900 dark:text-white">{{ row.original.title }}</div>
                <div v-if="row.original.description" class="text-[11px] text-gray-500 line-clamp-1 mt-0.5">{{ row.original.description }}</div>
              </div>
            </template>

            <template #period-cell="{ row }">
              <UBadge :color="getPeriodBadgeColor(row.original.period)" variant="subtle" size="md" class="font-bold">
                {{ row.original.period === 'Tahunan' ? t('kpiPerformance.upload.annual') : row.original.period }}
              </UBadge>
            </template>

            <template #year-cell="{ row }">
              <span class="font-medium text-gray-700 dark:text-gray-300">
                {{ row.original.year }}
              </span>
            </template>

            <template #fileName-cell="{ row }">
              <div>
                <div class="text-gray-700 dark:text-white font-medium truncate max-w-[140px]" :title="row.original.fileName">
                  {{ row.original.fileName }}
                </div>
                <div class="text-[10px] text-gray-400">{{ formatBytes(row.original.fileSize) }}</div>
              </div>
            </template>

            <template #created_at-cell="{ row }">
              <span class="text-gray-500">
                {{ new Intl.DateTimeFormat(locale.value === 'id' ? 'id-ID' : 'en-US', { day: '2-digit', month: 'short', year: 'numeric' }).format(new Date(row.original.created_at)) }}
              </span>
            </template>

            <template #actions-cell="{ row }">
              <div class="flex items-center gap-1">
                <UButton 
                  icon="i-lucide-eye" 
                  color="info" 
                  variant="ghost" 
                  size="sm" 
                  :title="t('kpiPerformance.upload.actions.view')"
                  @click="store.viewDocument(row.original.id, row.original.fileName)" 
                />
                <UButton 
                  icon="i-lucide-download"
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  @click="handleDownload(row.original.id, row.original.fileName)"
                  :title="t('kpiPerformance.upload.actions.download')"
                />
                <UButton
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="sm"
                  @click="handleDelete(row.original.id, row.original.title)"
                  :title="t('kpiPerformance.upload.actions.delete')"
                />
              </div>
            </template>
          </TableEntities>
        </UCard>
      </div>
    </div>
  </div>
</template>
