<template>
  <div class="p-4 sm:p-6 max-w-full mx-auto space-y-6 min-h-screen w-full min-w-0">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/executive-summary-compilation" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('executiveSummary.upload.title') }}</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('executiveSummary.upload.subtitle') }}</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex flex-col gap-10">
      <!-- Import Form Card -->
      <div class="w-full space-y-6">
        <UCard :ui="{ body: 'p-6' }" class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <h3 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <UIcon name="i-lucide-upload" class="w-5 h-5 text-primary" />
              {{ t('executiveSummary.upload.formTitle') }}
            </h3>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-6">
            <UFormField :label="t('executiveSummary.upload.documentTitle')" required>
              <UInput 
                v-model="form.title" 
                :placeholder="t('executiveSummary.upload.documentTitlePlaceholder')" 
                class="w-full"
                required
                maxlength="100"
                @invalid="($event.target as any)?.setCustomValidity('Judul maksimal 100 karakter dan wajib diisi')"
                @input="($event.target as any)?.setCustomValidity('')"
              />
              <div class="text-xs text-gray-500 mt-1 text-right">
                {{ form.title ? form.title.length : 0 }}/100
              </div>
            </UFormField>

            <UFormField :label="t('executiveSummary.upload.description')">
              <UTextarea 
                v-model="form.description" 
                :placeholder="t('executiveSummary.upload.descriptionPlaceholder')" 
                class="w-full"
              />
            </UFormField>

            <div class="space-y-2 pt-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('executiveSummary.upload.fileLabel') }}</label>
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
                  accept=".pdf,.docx,.doc"
                />
                
                <div v-if="!form.fileName" class="space-y-3">
                  <UIcon name="i-lucide-file-up" class="w-10 h-10 mx-auto text-gray-400" />
                  <div>
                    <p class="text-sm text-gray-600 dark:text-gray-300 font-semibold">{{ t('executiveSummary.upload.dropzonePrompt') }}</p>
                    <p class="text-md text-gray-400 mt-1">{{ t('executiveSummary.upload.dropzoneHint') }}</p>
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
                    {{ t('executiveSummary.upload.removeFile') }}
                  </button>
                </div>
              </div>
            </div>

            <div v-if="store.errorMsg" class="text-sm text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200 dark:border-red-900">
              {{ store.errorMsg }}
            </div>

            <UButton 
              type="submit" 
              :label="t('executiveSummary.upload.submitButton')" 
              color="primary" 
              class="w-full justify-center font-bold h-11 text-base" 
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            />
          </form>
        </UCard>
      </div>

      <!-- Imported Documents Table -->
      <div class="w-full min-w-0 p-4 sm:p-6 dashboard-card">
        <div class="flex items-center justify-between gap-2 mb-4">
          <h3 class="text-base font-bold text-slate-900 tracking-tight dark:text-white flex items-center gap-2">
            <UIcon name="i-lucide-list" class="size-5 text-primary-500" />
            {{ t('executiveSummary.upload.tableTitle') }}
          </h3>
          <UBadge color="primary" variant="subtle" size="sm" class="font-semibold tabular-nums">
            {{ t('executiveSummary.upload.documentsCount', { count: store.uploadedDocuments.length }) }}
          </UBadge>
        </div>

        <!-- Desktop View (>= md): Full Table Entities -->
        <div class="hidden md:block">
          <UCard class="overflow-hidden border border-gray-200 dark:border-gray-800" :ui="{ root: 'bg-transparent!', body: 'p-0 bg-transparent!' }">
            <TableEntities
              :data="store.uploadedDocuments"
              :columns="columns"
              :loading="store.loading"
              :empty-state="{
                icon: 'i-lucide-folder-open',
                label: t('executiveSummary.upload.emptyTitle'),
                description: t('executiveSummary.upload.emptyDesc')
              }"
              class="w-full"
            >
              <template #title-cell="{ row }">
                <div class="py-1 min-w-0">
                  <div class="font-semibold text-gray-900 dark:text-white text-sm line-clamp-1">
                    {{ row.original.title }}
                  </div>
                  <div class="text-xs text-[var(--text-muted)] mt-0.5 line-clamp-1" v-if="row.original.description">
                    {{ row.original.description }}
                  </div>
                </div>
              </template>

              <template #fileName-cell="{ row }">
                <div class="flex items-center gap-2 py-1 min-w-0">
                  <span class="inline-flex items-center justify-center size-7 rounded-lg bg-primary-50 dark:bg-primary-950/40 text-primary-600 dark:text-primary-400 border border-primary-200/50 dark:border-primary-800/40 shrink-0">
                    <UIcon name="i-lucide-file-text" class="size-3.5" />
                  </span>
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300 truncate max-w-[200px] block" :title="row.original.fileName">
                    {{ row.original.fileName }}
                  </span>
                </div>
              </template>

              <template #created_at-cell="{ row }">
                <span class="text-xs font-medium text-[var(--text-muted)] tabular-nums">
                  {{ formatDate(row.original.created_at) }}
                </span>
              </template>

              <template #actions-cell="{ row }">
                <div class="flex items-center gap-1">
                  <UTooltip :text="t('executiveSummary.upload.actions.view')">
                    <UButton 
                      icon="i-lucide-eye" 
                      color="info" 
                      variant="ghost" 
                      size="xs" 
                      class="hover:bg-info-50 dark:hover:bg-info-950/30"
                      :title="t('executiveSummary.upload.actions.view')" 
                      @click="store.viewDocument(row.original.id, row.original.fileName)" 
                    />
                  </UTooltip>
                  <UTooltip :text="t('executiveSummary.upload.actions.download')">
                    <UButton 
                      icon="i-lucide-download" 
                      color="primary" 
                      variant="ghost" 
                      size="xs" 
                      class="hover:bg-primary-50 dark:hover:bg-primary-950/30"
                      :title="t('executiveSummary.upload.actions.download')" 
                      @click="store.downloadDocument(row.original.id, row.original.fileName)" 
                    />
                  </UTooltip>
                  <UTooltip :text="t('executiveSummary.upload.actions.delete')">
                    <UButton 
                      icon="i-lucide-trash-2" 
                      color="error" 
                      variant="ghost" 
                      size="xs" 
                      class="hover:bg-error-50 dark:hover:bg-error-950/30"
                      :title="t('executiveSummary.upload.actions.delete')" 
                      @click="handleDelete(row.original.id)" 
                    />
                  </UTooltip>
                </div>
              </template>
            </TableEntities>
          </UCard>
        </div>

        <!-- Mobile & Tablet View (< md): Stacked Cards List -->
        <div class="block md:hidden">
          <!-- Loading State Skeleton -->
          <div v-if="store.loading" class="space-y-3">
            <div v-for="i in 3" :key="i" class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] animate-pulse space-y-2.5">
              <div class="flex justify-between items-center">
                <div class="h-4 w-28 bg-gray-200 dark:bg-gray-700 rounded"></div>
                <div class="h-4 w-20 bg-gray-200 dark:bg-gray-700 rounded"></div>
              </div>
              <div class="h-4 w-3/4 bg-gray-200 dark:bg-gray-700 rounded"></div>
              <div class="h-8 bg-gray-100 dark:bg-gray-800 rounded-lg"></div>
            </div>
          </div>

          <!-- Empty State -->
          <div
            v-else-if="!store.uploadedDocuments.length"
            class="py-8 text-center border border-dashed border-gray-200 dark:border-gray-800 rounded-xl bg-[var(--bg-surface)]"
          >
            <UIcon name="i-lucide-folder-open" class="size-8 mx-auto text-gray-400 mb-2" />
            <p class="text-xs font-semibold text-[var(--text-main)]">{{ t('executiveSummary.upload.emptyTitle') }}</p>
            <p class="text-xs text-[var(--text-muted)] mt-1">{{ t('executiveSummary.upload.emptyDesc') }}</p>
          </div>

          <!-- Cards List -->
          <div v-else class="space-y-3">
            <div
              v-for="(doc, idx) in store.uploadedDocuments"
              :key="doc.id"
              class="p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-[var(--bg-surface)] hover:border-primary-500/40 dark:hover:border-primary-500/40 transition-all shadow-xs space-y-3"
            >
              <!-- Card Header: Index Badge & Date Badge -->
              <div class="flex items-center justify-between gap-2">
                <span class="inline-flex items-center gap-1 font-mono text-[11px] font-bold text-primary-600 dark:text-primary-400 bg-primary-50 dark:bg-primary-950/40 px-2 py-0.5 rounded-md border border-primary-200/50 dark:border-primary-800/40 tabular-nums">
                  <UIcon name="i-lucide-file-text" class="size-3" />
                  #{{ idx + 1 }}
                </span>
                <span class="inline-flex items-center gap-1 text-[11px] font-medium text-[var(--text-muted)] tabular-nums">
                  <UIcon name="i-lucide-calendar" class="size-3 text-gray-400" />
                  {{ formatDate(doc.created_at) }}
                </span>
              </div>

              <!-- Card Body: Title and Description -->
              <div>
                <h4 class="text-sm font-bold text-gray-900 dark:text-white leading-snug">
                  {{ doc.title }}
                </h4>
                <p v-if="doc.description" class="text-xs text-[var(--text-muted)] mt-1 line-clamp-2">
                  {{ doc.description }}
                </p>
              </div>

              <!-- Attachment Info Chip -->
              <div class="flex items-center gap-2 p-2.5 rounded-lg bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-800 text-xs">
                <UIcon name="i-lucide-paperclip" class="w-3.5 h-3.5 text-primary-500 shrink-0" />
                <span class="truncate font-medium text-gray-700 dark:text-gray-300 flex-1">{{ doc.fileName }}</span>
              </div>

              <!-- Card Footer: Actions -->
              <div class="flex items-center justify-end gap-1 pt-2 border-t border-gray-100 dark:border-gray-800/80">
                <UButton
                  icon="i-lucide-eye"
                  color="info"
                  variant="ghost"
                  size="md"
                  :title="t('executiveSummary.upload.actions.view')"
                  :aria-label="t('executiveSummary.upload.actions.view')"
                  @click="store.viewDocument(doc.id, doc.fileName)"
                />
                <UButton
                  icon="i-lucide-download"
                  color="primary"
                  variant="ghost"
                  size="md"
                  :title="t('executiveSummary.upload.actions.download')"
                  :aria-label="t('executiveSummary.upload.actions.download')"
                  @click="store.downloadDocument(doc.id, doc.fileName)"
                />
                <UButton
                  icon="i-lucide-trash-2"
                  color="error"
                  variant="ghost"
                  size="md"
                  :title="t('executiveSummary.upload.actions.delete')"
                  :aria-label="t('executiveSummary.upload.actions.delete')"
                  @click="handleDelete(doc.id)"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUploadExecutiveSummaryReportStore } from '~/stores/upload-executive-summary-report'
import { useI18n } from '~/composables/useI18n'
import { useAppToast } from '~/composables/useAppToast'
import TableEntities from '~/components/shared/TableEntities.vue'
import { useGlobalModalStore } from '~/stores/global-modal'

const { t, locale } = useI18n()
const appToast = useAppToast()
const store = useUploadExecutiveSummaryReportStore()

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
  file: null as any
})

const columns = computed(() => [
  { accessorKey: 'title', header: t('executiveSummary.upload.columns.title'), class: 'min-w-[200px] w-[45%]' },
  { accessorKey: 'fileName', header: t('executiveSummary.upload.columns.file'), class: 'min-w-[150px] w-[25%]' },
  { accessorKey: 'created_at', header: t('executiveSummary.upload.columns.date'), class: 'min-w-[120px] w-[18%]' },
  { accessorKey: 'actions', header: t('executiveSummary.upload.columns.actions'), class: 'min-w-[100px] w-[12%]' }
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
    appToast.warning(t('executiveSummary.upload.fileSizeLimit'))
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
    await store.uploadDocument({
      title: form.value.title,
      description: form.value.description,
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

const handleDelete = async (id: string) => {
  if (await useGlobalModalStore().confirmDelete({ description: t('executiveSummary.upload.deleteConfirm') })) {
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
  return new Intl.DateTimeFormat(locale.value === 'id' ? 'id-ID' : 'en-US', {
    day: '2-digit',
    month: 'short',
    year: 'numeric'
  }).format(date)
}
</script>
