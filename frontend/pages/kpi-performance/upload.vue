<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useUploadPerformanceReportStore } from '~/stores/upload-performance-report'
import { useToast } from '#imports'

const store = useUploadPerformanceReportStore()
const toast = useToast()

const selectedPeriodFilter = ref('Semua')
const selectedYearFilter = ref('2026')
const periodFilterOptions = ['Semua', 'Q1', 'Q2', 'Q3', 'Q4', 'Tahunan']
const yearOptions = [2024, 2025, 2026, 2027, 2028]

const form = ref({
  title: '',
  period: 'Q1',
  year: 2026,
  description: '',
  fileName: '',
  fileType: '',
  fileContent: ''
})

const periodOptions = [
  { label: 'Q1 (Triwulan I)', value: 'Q1', description: 'Laporan Capaian Kinerja Q1 (Jan - Mar)' },
  { label: 'Q2 (Triwulan II)', value: 'Q2', description: 'Laporan Capaian Kinerja Q2 (Apr - Jun)' },
  { label: 'Q3 (Triwulan III)', value: 'Q3', description: 'Laporan Capaian Kinerja Q3 (Jul - Sep)' },
  { label: 'Q4 (Triwulan IV)', value: 'Q4', description: 'Laporan Capaian Kinerja Q4 (Okt - Des)' },
  { label: 'Kinerja Tahunan', value: 'Tahunan', description: 'Laporan Capaian Kinerja Konsolidasi Tahunan' }
]

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
      title: 'File Terlalu Besar',
      description: 'Ukuran file maksimum adalah 10MB.',
      color: 'error'
    })
    return
  }

  selectedFileLength.value = file.size
  form.value.fileName = file.name
  form.value.fileType = file.type || getExtensionType(file.name)

  if (!form.value.title) {
    const periodName = form.value.period === 'Tahunan' ? 'Kinerja Tahunan' : `Laporan Kinerja ${form.value.period}`
    form.value.title = `${periodName} ${form.value.year} - ${file.name.replace(/\.[^/.]+$/, '')}`
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    form.value.fileContent = e.target?.result as string
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
  form.value.fileContent = ''
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
  if (!form.value.title || !form.value.fileName || !form.value.fileContent) {
    toast.add({
      title: 'Formulir Belum Lengkap',
      description: 'Harap lengkapi Judul dan File Dokumen Laporan Kinerja.',
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
      fileContent: form.value.fileContent
    })

    toast.add({
      title: 'Berhasil Impor',
      description: `Dokumen Laporan Kinerja ${form.value.period} ${form.value.year} berhasil diunggah!`,
      color: 'success'
    })

    // Reset form
    form.value.title = ''
    form.value.description = ''
    clearFile()
  } catch (err: any) {
    toast.add({
      title: 'Gagal Mengunggah',
      description: store.errorMsg || 'Terjadi kesalahan saat mengunggah dokumen.',
      color: 'error'
    })
  }
}

const handleDelete = async (id: string, title: string) => {
  if (confirm(`Apakah Anda yakin ingin menghapus dokumen "${title}"?`)) {
    try {
      await store.deleteReport(id, selectedPeriodFilter.value, parseInt(selectedYearFilter.value))
      toast.add({
        title: 'Dokumen Dihapus',
        description: 'Dokumen Laporan Kinerja berhasil dihapus.',
        color: 'success'
      })
    } catch (err) {
      toast.add({
        title: 'Gagal Menghapus',
        description: store.errorMsg || 'Terjadi kesalahan saat menghapus dokumen.',
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

const columns = [
  { key: 'title', label: 'Judul Dokumen' },
  { key: 'period', label: 'Periode' },
  { key: 'year', label: 'Tahun' },
  { key: 'fileName', label: 'File' },
  { key: 'created_at', label: 'Tanggal Upload' },
  { key: 'actions', label: 'Aksi' }
]
</script>

<template>
  <div class="p-6 max-w-full mx-auto space-y-6 min-h-screen">
    <!-- Header -->
    <div class="flex items-center gap-4 mb-6">
      <UButton icon="i-lucide-arrow-left" color="neutral" variant="ghost" to="/kpi-performance" />
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Impor Dokumen Laporan Kinerja</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Unggah dan kelola dokumen Laporan Kinerja Q1, Q2, Q3, Q4 dan Laporan Kinerja Tahunan</p>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="flex flex-col gap-6">
      <!-- Upload Form Card (Left/Top) -->
      <div class="lg:col-span-5 space-y-6">
        <UCard class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <div class="flex items-center gap-2">
              <UIcon name="i-lucide-upload-cloud" class="w-5 h-5 text-primary" />
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">Formulir Impor Dokumen</h3>
            </div>
          </template>

          <form @submit.prevent="handleUpload" class="space-y-5">
            <!-- Period Selector -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-2">Periode Laporan Kinerja *</label>
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
                  {{ p === 'Tahunan' ? 'Tahunan' : p }}
                </button>
              </div>
            </div>

            <!-- Year Selector -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">Tahun Laporan *</label>
              <USelect
                v-model="form.year"
                :items="yearOptions"
                class="w-full"
              />
            </div>

            <!-- Document Title -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">Judul Dokumen *</label>
              <UInput
                v-model="form.title"
                placeholder="Contoh: Laporan Kinerja Q1 2026 Internal Audit"
                class="w-full"
                required
              />
            </div>

            <!-- Description -->
            <div>
              <label class="block text-sm font-semibold text-gray-800 dark:text-gray-200 mb-1">Deskripsi Ringkas</label>
              <UTextarea
                v-model="form.description"
                placeholder="Catatan atau ikhtisar dokumen laporan..."
                class="w-full"
                :rows="2"
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
                    <p class="text-md font-bold text-emerald-700 dark:text-emerald-400 truncate max-w-[220px] mx-auto">
                      {{ form.fileName }}
                    </p>
                    <p class="text-[11px] text-emerald-600 dark:text-emerald-500 mt-0.5">
                      {{ formatBytes(selectedFileLength) }}
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

            <div v-if="store.errorMsg" class="text-md text-red-600 font-semibold bg-red-50 dark:bg-red-950/30 p-3 rounded-lg border border-red-200">
              {{ store.errorMsg }}
            </div>

            <UButton
              type="submit"
              label="Impor Laporan Kinerja"
              color="primary"
              class="w-full justify-center font-bold h-10 text-sm"
              :loading="store.loading"
              icon="i-lucide-upload"
              :disabled="!form.title || !form.fileName"
            />
          </form>
        </UCard>
      </div>

      <!-- Uploaded Documents Table (Right/Bottom) -->
      <div class="lg:col-span-7 space-y-6">
        <UCard class="shadow-sm border border-gray-200 dark:border-gray-800">
          <template #header>
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div class="flex items-center gap-2">
                <UIcon name="i-lucide-file-text" class="w-5 h-5 text-primary" />
                <h3 class="text-lg font-bold text-gray-900 dark:text-white">Daftar Dokumen Laporan Kinerja</h3>
              </div>
              <div class="flex items-center gap-2">
                <!-- Period filter tabs -->
                <USelect
                  v-model="selectedPeriodFilter"
                  :items="periodFilterOptions"
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

          <div v-if="store.loading && store.uploadedReports.length === 0" class="py-12 text-center">
            <UIcon name="i-lucide-loader-2" class="w-8 h-8 text-primary animate-spin mx-auto mb-2" />
            <p class="text-gray-500 text-md">Memuat daftar dokumen...</p>
          </div>

          <div v-else-if="store.uploadedReports.length === 0" class="py-16 text-center space-y-3 rounded-lg border-2 border-dashed border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30">
            <UIcon name="i-lucide-folder-open" class="w-12 h-12 text-gray-300 mx-auto" />
            <div>
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">Belum Ada Dokumen Terimpor</h3>
              <p class="text-md text-gray-500 mt-1">Unggah dokumen Laporan Kinerja Q1, Q2, Q3, Q4, atau Tahunan melalui formulir di sebelah kiri.</p>
            </div>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-md">
              <thead class="bg-gray-50 dark:bg-gray-800/60 text-gray-600 dark:text-gray-400 font-semibold border-b border-gray-200 dark:border-gray-700">
                <tr>
                  <th class="py-3 px-4">Judul Dokumen</th>
                  <th class="py-3 px-3">Periode</th>
                  <th class="py-3 px-3">Tahun</th>
                  <th class="py-3 px-3">File</th>
                  <th class="py-3 px-3">Tanggal Upload</th>
                  <th class="py-3 px-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-gray-800">
                <tr v-for="doc in store.uploadedReports" :key="doc.id" class="hover:bg-gray-50/60 dark:hover:bg-gray-800/40">
                  <td class="py-3 px-4">
                    <div class="font-bold text-gray-900 dark:text-white">{{ doc.title }}</div>
                    <div v-if="doc.description" class="text-[11px] text-gray-500 line-clamp-1 mt-0.5">{{ doc.description }}</div>
                  </td>
                  <td class="py-3 px-3">
                    <UBadge :color="getPeriodBadgeColor(doc.period)" variant="subtle" size="md" class="font-bold">
                      {{ doc.period }}
                    </UBadge>
                  </td>
                  <td class="py-3 px-3 font-medium text-gray-700 dark:text-gray-300">
                    {{ doc.year }}
                  </td>
                  <td class="py-3 px-3">
                    <div class="text-gray-700 dark:text-white font-medium truncate max-w-[140px]" :title="doc.fileName">
                      {{ doc.fileName }}
                    </div>
                    <div class="text-[10px] text-gray-400">{{ formatBytes(doc.fileSize) }}</div>
                  </td>
                  <td class="py-3 px-3 text-gray-500">
                    {{ new Date(doc.created_at).toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' }) }}
                  </td>
                  <td class="py-3 px-3 text-right">
                    <div class="flex items-center justify-end gap-1">                  <UButton 
                    icon="i-lucide-eye" 
                    color="info" 
                    variant="ghost" 
                    size="sm" 
                    title="View Document"
                    @click="store.viewDocument(doc.id, doc.fileName)" 
                  />
                  <UButton 
                    icon="i-lucide-download"
                        color="neutral"
                        variant="ghost"
                        size="md"
                        @click="handleDownload(doc.id, doc.fileName)"
                        title="Unduh Dokumen"
                      />
                      <UButton
                        icon="i-lucide-trash-2"
                        color="error"
                        variant="ghost"
                        size="md"
                        @click="handleDelete(doc.id, doc.title)"
                        title="Hapus Dokumen"
                      />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </UCard>
      </div>
    </div>
  </div>
</template>
