<script setup lang="ts">
import { ref } from 'vue'
import { useToast } from '#imports'

definePageMeta({
  middleware: 'auth'
})

const toast = useToast()

const form = ref({
  title: '',
  description: '',
  fileName: '',
  file: null as any
})

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const selectedFileSize = ref(0)

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
      title: 'File Terlalu Besar',
      description: 'Ukuran file maksimum adalah 10MB.',
      color: 'error'
    })
    return
  }

  selectedFileSize.value = file.size
  form.value.fileName = file.name

  if (!form.value.title) {
    form.value.title = `Dokumen Strategic Audit Plan - ${file.name.replace(/\.[^/.]+$/, '')}`
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    form.value.fileContent = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const clearFile = () => {
  form.value.fileName = ''
  form.value.file = null as any
  selectedFileSize.value = 0
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

const handleUpload = () => {
  if (!form.value.title || !form.value.fileName) {
    toast.add({
      title: 'Formulir Belum Lengkap',
      description: 'Harap lengkapi Judul dan File Dokumen.',
      color: 'error'
    })
    return
  }

  toast.add({
    title: 'Berhasil Impor',
    description: `Dokumen "${form.value.title}" berhasil diunggah!`,
    color: 'success'
  })

  form.value.title = ''
  form.value.description = ''
  clearFile()
}
</script>

<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/strategic-audit-plan" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Import Strategic Audit Plan</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Unggah dan kelola dokumen Strategic Audit Plan (Peta Strategis Internal Audit)</p>
      </div>
    </div>

    <!-- Main Content -->
    <div class="max-w-3xl space-y-6">
      <UCard class="shadow-sm border border-gray-200 dark:border-gray-800">
        <template #header>
          <div class="flex items-center gap-2">
            <UIcon name="i-lucide-upload-cloud" class="w-5 h-5 text-primary" />
            <h3 class="text-lg font-bold text-gray-900 dark:text-white">Formulir Impor Dokumen Strategic Audit Plan</h3>
          </div>
        </template>

        <form @submit.prevent="handleUpload" class="space-y-5">
          <div>
            <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">Judul Dokumen *</label>
            <UInput
              v-model="form.title"
              placeholder="Contoh: Dokumen Strategic Audit Plan 2026-2030"
              class="w-full"
              required
            />
          </div>

          <div>
            <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">Deskripsi Ringkas</label>
            <UTextarea
              v-model="form.description"
              placeholder="Catatan atau ringkasan dokumen..."
              class="w-full"
              :rows="3"
            />
          </div>

          <!-- File Upload Drag Zone -->
          <div class="space-y-2 pt-1">
            <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200">Upload File Dokumen *</label>
            <div
              @click="triggerFileSelect"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="handleFileDrop"
              class="border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors duration-200"
              :class="[
                isDragging
                  ? 'border-primary bg-primary/5'
                  : form.fileName
                    ? 'border-emerald-500 bg-emerald-50/40 dark:bg-emerald-950/20'
                    : 'border-gray-300 dark:border-gray-700 hover:border-primary bg-white dark:bg-gray-800'
              ]"
            >
              <input
                type="file"
                ref="fileInput"
                class="hidden"
                @change="handleFileSelect"
                accept=".pdf,.docx,.doc,.xls,.xlsx"
              />

              <div v-if="!form.fileName" class="space-y-2">
                <UIcon name="i-lucide-file-up" class="w-9 h-9 mx-auto text-gray-400" />
                <div>
                  <p class="text-md font-semibold text-gray-700 dark:text-gray-300">Klik untuk upload atau drag & drop file</p>
                  <p class="text-[11px] text-gray-400 mt-0.5">PDF, DOCX, XLSX hingga 10MB</p>
                </div>
              </div>

              <div v-else class="space-y-2">
                <UIcon name="i-lucide-file-check-2" class="w-9 h-9 mx-auto text-emerald-500" />
                <div>
                  <p class="text-md font-bold text-emerald-700 dark:text-emerald-400 truncate max-w-[260px] mx-auto">
                    {{ form.fileName }}
                  </p>
                  <p class="text-[11px] text-emerald-600 dark:text-emerald-500 mt-0.5">
                    {{ formatBytes(selectedFileSize) }}
                  </p>
                </div>
                <button
                  type="button"
                  @click.stop="clearFile"
                  class="text-md text-red-500 hover:underline font-bold mt-1 inline-block"
                >
                  Ganti File
                </button>
              </div>
            </div>
          </div>

          <UButton
            type="submit"
            label="Impor Strategic Audit Plan"
            color="primary"
            class="w-full justify-center font-bold h-10 text-sm"
            icon="i-lucide-upload"
            :disabled="!form.title || !form.fileName"
          />
        </form>
      </UCard>
    </div>
  </div>
</template>
