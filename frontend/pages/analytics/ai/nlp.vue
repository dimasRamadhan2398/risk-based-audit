<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { Doughnut } from 'vue-chartjs'
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend
} from 'chart.js'
import { useAiAnalytics } from '~/composables/useAiAnalytics'
import { useI18n } from '~/composables/useI18n'
import AiInsightHeader from '~/components/analytics/AiInsightHeader.vue'
import { extractTextFromClientFile } from '~/utils/documentTextExtractor'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()
const {
  loading,
  nlpState,
  getRiskConfig,
  sentimentColor,
  riskCategoryColor,
  controlIndicatorColor,
  fetchAiAnalytics,
  analyzeDocumentOrText,
  addNlpTaxonomyItem,
  syncNlpMultiSources,
  deleteNlpDocument,
  clearAllCustomNlpDocuments
} = useAiAnalytics()

onMounted(() => {
  fetchAiAnalytics()
})

// Filters & Search
const selectedSourceType = ref<string>('ALL')
const selectedCategory = ref<string>('ALL')
const selectedIndicator = ref<string>('ALL')
const searchQuery = ref<string>('')
const chartMode = ref<'indicator' | 'sentiment'>('indicator')
const activeKpiFilter = ref<'ALL' | 'FRAUD' | 'WEAKNESS' | 'EXTERNAL' | null>(null)
const viewMode = ref<'table' | 'cards'>('table')

// Pagination
const currentPage = ref(1)
const itemsPerPage = ref(8)

// Modals & States
const showUploadModal = ref(false)
const showTaxonomyModal = ref(false)
const showDetailModal = ref(false)
const selectedDocDetail = ref<any>(null)
const isAnalyzingDoc = ref(false)
const isReadingFile = ref(false)
const isSyncingSources = ref(false)
const isDeletingDoc = ref(false)
const actionFeedback = ref<string>('')
const isDragging = ref(false)
const activeTaxonomyTab = ref<'categories' | 'indicators'>('categories')

const baselineDocIds = [
  'WP-2026-001', 'LHA-2026-002', 'EXT-DB-003', 'EXT-DB-004', 'NOTE-FW-005', 'PKPT-2026-006', 'QA-2026-007',
  'WP-001', 'LHA-002', 'PKPT-003', 'NOTE-004'
]

const hasManualImports = computed(() => {
  const docs = nlpState.value?.documents || []
  return docs.some((d: any) =>
    !baselineDocIds.includes(d.docId) && (d.docId?.startsWith('DOC-') || d.docId?.startsWith('CUSTOM-'))
  )
})

const handleDeleteDocument = async (docId: string) => {
  if (!docId) return
  isDeletingDoc.value = true
  try {
    await deleteNlpDocument(docId)
    actionFeedback.value = `Dokumen ${docId} berhasil dihapus!`
    showDetailModal.value = false
    setTimeout(() => { actionFeedback.value = '' }, 5000)
  } finally {
    isDeletingDoc.value = false
  }
}

const handleClearAllManualImports = async () => {
  isDeletingDoc.value = true
  try {
    await clearAllCustomNlpDocuments()
    actionFeedback.value = 'Semua data dokumen impor manual telah berhasil dihapus dan dibersihkan!'
    setTimeout(() => { actionFeedback.value = '' }, 6000)
  } finally {
    isDeletingDoc.value = false
  }
}

// Upload Form State
const uploadForm = ref({
  title: '',
  source: 'Working Paper',
  sourceType: 'Imported Document',
  riskCategory: '',
  text: '',
  fileName: '',
  fileType: 'PDF',
  fileSizeText: '',
  fileContentBase64: ''
})

// Dynamic Taxonomy Form State
const taxonomyForm = ref({
  itemType: 'risk_category' as 'risk_category' | 'control_indicator',
  name: '',
  keywordsText: '',
  description: '',
  severityBoost: 14
})

const sourceOptions = [
  { label: 'Working Paper (Kertas Kerja Audit)', value: 'Working Paper', type: 'Imported Document' },
  { label: 'Audit Result Report (Laporan Hasil Audit / LHA)', value: 'Audit Result Report', type: 'Imported Document' },
  { label: 'Annual Audit Plan (PKPT)', value: 'Annual Audit Plan', type: 'Imported Document' },
  { label: 'Executive Summary (Ringkasan Eksekutif)', value: 'Executive Summary', type: 'Imported Document' },
  { label: 'Assignment Letter (Surat Tugas)', value: 'Assignment Letter', type: 'Imported Document' },
  { label: 'Quality Assurance (Laporan QAR)', value: 'Quality Assurance', type: 'Imported Document' },
  { label: 'Consulting Service (Dokumen Konsultansi)', value: 'Consulting Service', type: 'Imported Document' },
  { label: 'Internal Note: Fieldwork Observation / Interview', value: 'Internal Note: Fieldwork Observation', type: 'Internal Note' },
  { label: 'Internal Note: Risk Control Matrix (RCM)', value: 'Internal Note: Risk Control Matrix', type: 'Internal Note' },
  { label: 'External DB: Core Banking / Compliance Incident', value: 'External DB: Compliance Incident Records', type: 'External Database' }
]

const availableCategories = computed<string[]>(() => {
  const fromTax = (nlpState.value?.taxonomy?.risk_categories || []).map((c: any) => c.name || c)
  const fromDist = Object.keys(nlpState.value?.categoryDistribution || {})
  const fromDocs = (nlpState.value?.documents || []).map((d: any) => d.autoCategory).filter(Boolean)
  const defaults = ['Financial', 'Operational', 'Compliance', 'Technology']
  return Array.from(new Set([...defaults, ...fromTax, ...fromDist, ...fromDocs]))
})

const availableIndicators = computed<string[]>(() => {
  const fromTax = (nlpState.value?.taxonomy?.control_indicators || []).map((c: any) => c.name || c)
  const fromDist = Object.keys(nlpState.value?.controlIndicatorDistribution || {})
  const fromDocs = (nlpState.value?.documents || []).map((d: any) => d.controlIndicator).filter(Boolean)
  const defaults = [
    'Fraud Indication',
    'Internal Control Weakness',
    'Segregation of Duties (SoD) Violation',
    'IT & Access Control Weakness',
    'Compliance Deviation',
    'Adequate Control'
  ]
  return Array.from(new Set([...defaults, ...fromTax, ...fromDist, ...fromDocs]))
})

const filteredDocuments = computed(() => {
  const docs = nlpState.value?.documents || []
  return docs.filter((doc: any) => {
    if (selectedSourceType.value !== 'ALL' && (doc.sourceType || 'Imported Document') !== selectedSourceType.value) {
      return false
    }
    if (selectedCategory.value !== 'ALL' && doc.autoCategory !== selectedCategory.value) {
      return false
    }
    if (selectedIndicator.value !== 'ALL' && (doc.controlIndicator || 'Internal Control Weakness') !== selectedIndicator.value) {
      return false
    }
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.toLowerCase()
      const hay = [
        doc.docId,
        doc.title,
        doc.source,
        doc.fileName,
        doc.autoCategory,
        doc.controlIndicator,
        doc.extractedFinding,
        doc.extractedRecommendation,
        doc.excerpt
      ].join(' ').toLowerCase()
      if (!hay.includes(q)) return false
    }
    return true
  })
})

// Pagination Computations
const totalPages = computed(() => {
  return Math.max(1, Math.ceil(filteredDocuments.value.length / itemsPerPage.value))
})

const paginatedDocuments = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredDocuments.value.slice(start, start + itemsPerPage.value)
})

watch([selectedSourceType, selectedCategory, selectedIndicator, searchQuery], () => {
  currentPage.value = 1
})

// KPI Metrics & Click-to-Filter Logic
const kpiStats = computed(() => {
  const docs = nlpState.value?.documents || []
  const total = docs.length
  const importedCount = docs.filter((d: any) => (d.sourceType || 'Imported Document') === 'Imported Document').length
  const noteCount = docs.filter((d: any) => d.sourceType === 'Internal Note').length
  const extDbCount = docs.filter((d: any) => d.sourceType === 'External Database').length
  const fraudCount = docs.filter((d: any) => String(d.controlIndicator || '').toLowerCase().includes('fraud')).length
  const controlWeaknessCount = docs.filter((d: any) => {
    const ind = String(d.controlIndicator || '').toLowerCase()
    return ind.includes('weakness') || ind.includes('violation') || ind.includes('deviation')
  }).length
  return { total, importedCount, noteCount, extDbCount, fraudCount, controlWeaknessCount }
})

const applyKpiFilter = (type: 'ALL' | 'FRAUD' | 'WEAKNESS' | 'EXTERNAL') => {
  if (activeKpiFilter.value === type) {
    resetAllFilters()
    return
  }
  activeKpiFilter.value = type
  currentPage.value = 1

  if (type === 'ALL') {
    selectedSourceType.value = 'ALL'
    selectedCategory.value = 'ALL'
    selectedIndicator.value = 'ALL'
    searchQuery.value = ''
  } else if (type === 'FRAUD') {
    selectedSourceType.value = 'ALL'
    selectedCategory.value = 'ALL'
    selectedIndicator.value = 'Fraud Indication'
  } else if (type === 'WEAKNESS') {
    selectedSourceType.value = 'ALL'
    selectedCategory.value = 'ALL'
    selectedIndicator.value = 'Internal Control Weakness'
  } else if (type === 'EXTERNAL') {
    selectedSourceType.value = 'External Database'
    selectedCategory.value = 'ALL'
    selectedIndicator.value = 'ALL'
  }
}

const resetAllFilters = () => {
  activeKpiFilter.value = null
  selectedSourceType.value = 'ALL'
  selectedCategory.value = 'ALL'
  selectedIndicator.value = 'ALL'
  searchQuery.value = ''
  currentPage.value = 1
}

const isFilterActive = computed(() => {
  return selectedSourceType.value !== 'ALL' ||
    selectedCategory.value !== 'ALL' ||
    selectedIndicator.value !== 'ALL' ||
    searchQuery.value.trim().length > 0 ||
    activeKpiFilter.value !== null
})

// Charts Data
const categoryPalette = [
  '#4F46E5', '#06B6D4', '#F59E0B', '#EF4444',
  '#8B5CF6', '#10B981', '#EC4899', '#14B8A6', '#F97316'
]

const categoryDoughnutData = computed(() => {
  const counts: Record<string, number> = {}
  availableCategories.value.forEach((cat) => { counts[cat] = 0 })
  ;(nlpState.value?.documents || []).forEach((d: any) => {
    const c = d.autoCategory || 'Operational'
    counts[c] = (counts[c] || 0) + 1
  })
  const labels = Object.keys(counts).filter(k => counts[k] > 0 || ['Financial', 'Operational', 'Compliance', 'Technology'].includes(k))
  return {
    labels,
    datasets: [{
      data: labels.map(l => counts[l] || 0),
      backgroundColor: labels.map((_, idx) => categoryPalette[idx % categoryPalette.length]),
      borderWidth: 2,
      borderColor: 'rgba(255,255,255,0.1)',
    }]
  }
})

const getSentimentCount = (key: string): number => {
  const dist = nlpState.value?.sentimentDistribution || {}
  const lower = key.toLowerCase()
  const upper = key.charAt(0).toUpperCase() + key.slice(1).toLowerCase()
  return (dist[lower] ?? dist[upper] ?? dist[key] ?? 0) as number
}

const sentimentOrIndicatorDoughnutData = computed(() => {
  if (chartMode.value === 'sentiment') {
    return {
      labels: ['Negative (Temuan / Risiko)', 'Neutral (Observasi)', 'Positive (Kontrol Memadai)'],
      datasets: [{
        data: [
          getSentimentCount('negative'),
          getSentimentCount('neutral'),
          getSentimentCount('positive')
        ],
        backgroundColor: ['#EF4444', '#F59E0B', '#10B981'],
        borderWidth: 2,
        borderColor: 'rgba(255,255,255,0.1)',
      }]
    }
  }

  const indCounts: Record<string, number> = {}
  ;(nlpState.value?.documents || []).forEach((d: any) => {
    const ind = d.controlIndicator || 'Internal Control Weakness'
    indCounts[ind] = (indCounts[ind] || 0) + 1
  })
  const labels = Object.keys(indCounts)
  const indColors = labels.map((l) => {
    if (l.includes('Fraud')) return '#DC2626'
    if (l.includes('SoD')) return '#EA580C'
    if (l.includes('IT')) return '#D97706'
    if (l.includes('Weakness')) return '#F59E0B'
    if (l.includes('Compliance')) return '#3B82F6'
    if (l.includes('Adequate')) return '#10B981'
    return '#8B5CF6'
  })
  return {
    labels,
    datasets: [{
      data: labels.map(l => indCounts[l]),
      backgroundColor: indColors,
      borderWidth: 2,
      borderColor: 'rgba(255,255,255,0.1)',
    }]
  }
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        boxWidth: 12,
        font: { size: 11 }
      }
    }
  }
}

// Source Option Changed
const onSourceOptionChange = () => {
  const found = sourceOptions.find(s => s.value === uploadForm.value.source)
  if (found) {
    uploadForm.value.sourceType = found.type
  }
}

// File Process
const processFile = async (file: File) => {
  uploadForm.value.fileName = file.name
  const ext = file.name.split('.').pop()?.toUpperCase() || 'PDF'
  uploadForm.value.fileType = ext
  const sizeKb = file.size / 1024
  uploadForm.value.fileSizeText = sizeKb > 1024 ? `${(sizeKb / 1024).toFixed(2)} MB` : `${Math.round(sizeKb)} KB`

  if (!uploadForm.value.title) {
    uploadForm.value.title = file.name.replace(/\.[^/.]+$/, '').replace(/[_-]/g, ' ')
  }

  isReadingFile.value = true

  // 1. Read Base64 for API transmission
  const reader = new FileReader()
  reader.onload = () => {
    const resStr = String(reader.result || '')
    const b64 = resStr.includes(';base64,') ? resStr.split(';base64,')[1] : resStr
    uploadForm.value.fileContentBase64 = b64 || ''
  }
  reader.readAsDataURL(file)

  // 2. Perform genuine full client-side text extraction (PDF / DOCX / XLSX / CSV / TXT)
  try {
    const extracted = await extractTextFromClientFile(file)
    if (extracted.text && extracted.text.trim()) {
      uploadForm.value.text = extracted.text
    }
  } catch (err) {
    console.warn('Client extraction error:', err)
  } finally {
    isReadingFile.value = false
  }
}

const handleFileUpload = (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) processFile(file)
  input.value = ''
}

const handleFileDrop = (e: DragEvent) => {
  isDragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) processFile(file)
}

const clearSelectedFile = () => {
  uploadForm.value.fileName = ''
  uploadForm.value.fileType = 'PDF'
  uploadForm.value.fileSizeText = ''
  uploadForm.value.fileContentBase64 = ''
  isReadingFile.value = false
}

// Submit Document
const submitDocumentForAnalysis = async () => {
  if (isReadingFile.value) return
  if (!uploadForm.value.fileContentBase64 && !uploadForm.value.fileName && !uploadForm.value.text.trim() && !uploadForm.value.title.trim()) {
    return
  }
  isAnalyzingDoc.value = true
  try {
    const cleanTitle = uploadForm.value.title.trim() || uploadForm.value.fileName.replace(/\.[^/.]+$/, '').replace(/[_-]/g, ' ') || 'Dokumen Audit Impor'
    const analyzed = await analyzeDocumentOrText({
      title: cleanTitle,
      text: uploadForm.value.text,
      source: uploadForm.value.source,
      sourceType: uploadForm.value.sourceType,
      fileName: uploadForm.value.fileName || `${cleanTitle.replace(/\s+/g, '_')}.${uploadForm.value.fileType.toLowerCase()}`,
      fileType: uploadForm.value.fileType,
      fileContentBase64: uploadForm.value.fileContentBase64,
      riskCategory: uploadForm.value.riskCategory || undefined
    })
    showUploadModal.value = false
    actionFeedback.value = `Berhasil mengekstrak teks/OCR & menganalisis "${analyzed.title}" → Kategori: ${analyzed.autoCategory} | Indikator: ${analyzed.controlIndicator}`
    uploadForm.value = {
      title: '',
      source: 'Working Paper',
      sourceType: 'Imported Document',
      riskCategory: '',
      text: '',
      fileName: '',
      fileType: 'PDF',
      fileSizeText: '',
      fileContentBase64: ''
    }
    setTimeout(() => { actionFeedback.value = '' }, 7000)
  } finally {
    isAnalyzingDoc.value = false
  }
}

// Submit Taxonomy Item
const submitNewTaxonomyItem = async () => {
  if (!taxonomyForm.value.name.trim()) return
  isSavingTaxonomy.value = true
  try {
    const keywords = taxonomyForm.value.keywordsText
      .split(',')
      .map(k => k.trim())
      .filter(Boolean)
    await addNlpTaxonomyItem({
      itemType: taxonomyForm.value.itemType,
      name: taxonomyForm.value.name.trim(),
      keywords,
      description: taxonomyForm.value.description.trim(),
      severityBoost: taxonomyForm.value.severityBoost
    })
    const labelType = taxonomyForm.value.itemType === 'risk_category' ? 'Risk Category' : 'Control Indicator'
    actionFeedback.value = `${labelType} baru "${taxonomyForm.value.name.trim()}" berhasil ditambahkan & model IndoBERT otomatis diperbarui!`
    taxonomyForm.value.name = ''
    taxonomyForm.value.keywordsText = ''
    taxonomyForm.value.description = ''
    setTimeout(() => { actionFeedback.value = '' }, 6000)
  } finally {
    isSavingTaxonomy.value = false
  }
}

// Sync Multi-Source (Zero CSV)
const handleSyncAllSources = async () => {
  isSyncingSources.value = true
  try {
    await syncNlpMultiSources()
    actionFeedback.value = 'Sinkronisasi Multi-Source (Dokumen Impor, Catatan Internal, & Database Eksternal) berhasil tanpa beban CSV. Model IndoBERT telah diperbarui secara live.'
    setTimeout(() => { actionFeedback.value = '' }, 7000)
  } finally {
    isSyncingSources.value = false
  }
}

const openDocumentDetail = (doc: any) => {
  selectedDocDetail.value = doc
  showDetailModal.value = true
}

const sourceTypeBadgeColor = (st?: string) => {
  if (st === 'External Database') return 'info'
  if (st === 'Internal Note') return 'warning'
  return 'primary'
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    actionFeedback.value = 'Teks berhasil disalin ke clipboard!'
    setTimeout(() => { actionFeedback.value = '' }, 3000)
  } catch (e) {
    console.warn(e)
  }
}
</script>

<template>
  <div class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8 space-y-6">
    <AiInsightHeader
      :title="t('navigation.nlpAnalysis')"
      :subtitle="t('analytics.subtitle')"
      icon="i-lucide-file-search"
      currentSubModule="nlp"
    />

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center h-96">
      <div class="flex flex-col items-center gap-4">
        <UIcon name="i-heroicons-cpu-chip" class="w-12 h-12 animate-pulse text-indigo-500" />
        <span class="text-sm font-semibold text-gray-500 animate-pulse">{{ t('analytics.loading') }}</span>
      </div>
    </div>

    <!-- Main Content -->
    <div v-else class="space-y-6">
      <!-- Top Alert & Action Toolbar -->
      <div class="p-5 sm:p-6 rounded-2xl bg-gradient-to-r from-emerald-50/90 via-teal-50/60 to-sky-50/80 dark:from-emerald-950/40 dark:via-gray-900 dark:to-sky-950/40 border border-emerald-200/80 dark:border-emerald-800/60 shadow-xs space-y-4 transition-all">
        <!-- Row 1: Header Title & Action Buttons -->
        <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 pb-4 border-b border-emerald-200/60 dark:border-emerald-800/50">
          <div class="flex items-start sm:items-center gap-3">
            <div class="p-2.5 rounded-xl bg-emerald-600 text-white shadow-xs shrink-0 mt-0.5 sm:mt-0">
              <UIcon name="i-heroicons-document-magnifying-glass" class="w-6 h-6" />
            </div>
            <div>
              <div class="flex items-center gap-2 flex-wrap">
                <h2 class="text-base sm:text-lg font-bold text-gray-900 dark:text-white leading-tight">
                  {{ t('analytics.nlp.alertTitle') }}
                </h2>
                <UBadge color="primary" variant="subtle" size="xs" class="font-bold">
                  IndoBERT NLP • Multi-Source
                </UBadge>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Ekstraksi Teks / OCR Otomatis & Klasifikasi Sentimen Temuan Audit
              </p>
            </div>
          </div>

          <!-- Action Buttons Toolbar -->
          <div class="flex flex-wrap items-center gap-2 sm:gap-2.5">
            <!-- <UButton
              icon="i-heroicons-arrow-up-tray"
              color="primary"
              size="sm"
              class="font-semibold shadow-xs"
              @click="showUploadModal = true"
            >
              {{ t('analytics.nlp.btnImportAnalyze') }}
            </UButton> -->
            <UButton
              icon="i-heroicons-tag"
              color="neutral"
              variant="soft"
              size="sm"
              class="font-medium"
              @click="showTaxonomyModal = true"
            >
              {{ t('analytics.nlp.btnManageTaxonomy') }}
            </UButton>
            <UButton
              icon="i-heroicons-arrow-path"
              color="info"
              variant="soft"
              size="sm"
              class="font-medium"
              :loading="isSyncingSources"
              @click="handleSyncAllSources"
            >
              {{ t('analytics.nlp.btnSyncSources') }}
            </UButton>
          </div>
        </div>

        <!-- Row 2: Full Width Description Paragraph -->
        <p class="text-xs sm:text-sm text-gray-700 dark:text-gray-300 leading-relaxed max-w-5xl">
          {{ t('analytics.nlp.alertDesc') }}
        </p>

        <!-- Row 3: 3-Step Pipeline Flow (Responsive 3-Column Grid) -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3 pt-1">
          <!-- Step 1 -->
          <div class="flex items-center gap-3 p-3 rounded-xl bg-white/90 dark:bg-gray-800/90 border border-emerald-200/70 dark:border-emerald-800/60 shadow-2xs">
            <div class="w-8 h-8 rounded-lg bg-emerald-500 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-xs">
              1
            </div>
            <div class="min-w-0">
              <p class="text-xs font-bold text-gray-900 dark:text-white truncate">Text Extraction & OCR</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">PDF, Word, Excel, Catatan & DB</p>
            </div>
          </div>

          <!-- Step 2 -->
          <div class="flex items-center gap-3 p-3 rounded-xl bg-white/90 dark:bg-gray-800/90 border border-emerald-200/70 dark:border-emerald-800/60 shadow-2xs">
            <div class="w-8 h-8 rounded-lg bg-indigo-500 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-xs">
              2
            </div>
            <div class="min-w-0">
              <p class="text-xs font-bold text-gray-900 dark:text-white truncate">IndoBERT NLP Model</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">Deteksi Multi-Task & Semantik</p>
            </div>
          </div>

          <!-- Step 3 -->
          <div class="flex items-center gap-3 p-3 rounded-xl bg-white/90 dark:bg-gray-800/90 border border-emerald-200/70 dark:border-emerald-800/60 shadow-2xs">
            <div class="w-8 h-8 rounded-lg bg-sky-500 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-xs">
              3
            </div>
            <div class="min-w-0">
              <p class="text-xs font-bold text-gray-900 dark:text-white truncate">Klasifikasi & Rekomendasi</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">Kategori Risiko, Fraud & Solusi</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Feedback Banner -->
      <UAlert
        v-if="actionFeedback"
        icon="i-heroicons-check-circle"
        color="primary"
        variant="subtle"
        :title="actionFeedback"
        :close-button="{ icon: 'i-heroicons-x-mark-20-solid', color: 'gray', variant: 'link', padded: false }"
        @close="actionFeedback = ''"
      />

      <!-- Interactive KPI Cards (Click to Filter) -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Card 1: Total Documents -->
        <div
          class="relative p-4 rounded-2xl bg-white dark:bg-gray-900 border transition-all cursor-pointer hover:shadow-md"
          :class="activeKpiFilter === 'ALL' || (!activeKpiFilter && isFilterActive)
            ? 'border-indigo-500 ring-2 ring-indigo-500/20'
            : 'border-gray-200 dark:border-gray-800 hover:border-indigo-300 dark:hover:border-indigo-700'"
          @click="applyKpiFilter('ALL')"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('analytics.nlp.kpiTotalDocs') }}
              </p>
              <p class="text-3xl font-extrabold text-gray-900 dark:text-white mt-1">{{ kpiStats.total }}</p>
              <div class="flex items-center gap-1.5 mt-2 text-[11px] text-gray-500 dark:text-gray-400">
                <span>{{ kpiStats.importedCount }} Impor</span>
                <span>•</span>
                <span>{{ kpiStats.noteCount }} Catatan</span>
                <span>•</span>
                <span>{{ kpiStats.extDbCount }} DB</span>
              </div>
            </div>
            <div class="p-3 rounded-xl bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
              <UIcon name="i-heroicons-document-text" class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-[11px] font-semibold text-indigo-600 dark:text-indigo-400 border-t border-gray-100 dark:border-gray-800 pt-2">
            <span>Tampilkan Semua Data</span>
            <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5" />
          </div>
        </div>

        <!-- Card 2: Fraud Detected -->
        <div
          class="relative p-4 rounded-2xl bg-white dark:bg-gray-900 border transition-all cursor-pointer hover:shadow-md"
          :class="activeKpiFilter === 'FRAUD'
            ? 'border-rose-500 ring-2 ring-rose-500/20'
            : 'border-gray-200 dark:border-gray-800 hover:border-rose-300 dark:hover:border-rose-700'"
          @click="applyKpiFilter('FRAUD')"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('analytics.nlp.kpiFraudDetected') }}
              </p>
              <p class="text-3xl font-extrabold text-rose-600 dark:text-rose-400 mt-1">{{ kpiStats.fraudCount }}</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-2">
                Fraud, transaksi fiktif & indikasi AML
              </p>
            </div>
            <div class="p-3 rounded-xl bg-rose-50 dark:bg-rose-950/60 text-rose-600 dark:text-rose-400">
              <UIcon name="i-heroicons-shield-exclamation" class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-[11px] font-semibold text-rose-600 dark:text-rose-400 border-t border-gray-100 dark:border-gray-800 pt-2">
            <span>Filter Indikasi Fraud</span>
            <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5" />
          </div>
        </div>

        <!-- Card 3: Internal Control Weakness -->
        <div
          class="relative p-4 rounded-2xl bg-white dark:bg-gray-900 border transition-all cursor-pointer hover:shadow-md"
          :class="activeKpiFilter === 'WEAKNESS'
            ? 'border-amber-500 ring-2 ring-amber-500/20'
            : 'border-gray-200 dark:border-gray-800 hover:border-amber-300 dark:hover:border-amber-700'"
          @click="applyKpiFilter('WEAKNESS')"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('analytics.nlp.kpiControlWeakness') }}
              </p>
              <p class="text-3xl font-extrabold text-amber-600 dark:text-amber-400 mt-1">{{ kpiStats.controlWeaknessCount }}</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-2">
                Kelemahan kontrol internal & SoD
              </p>
            </div>
            <div class="p-3 rounded-xl bg-amber-50 dark:bg-amber-950/60 text-amber-600 dark:text-amber-400">
              <UIcon name="i-heroicons-exclamation-triangle" class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-[11px] font-semibold text-amber-600 dark:text-amber-400 border-t border-gray-100 dark:border-gray-800 pt-2">
            <span>Filter Kelemahan Kontrol</span>
            <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5" />
          </div>
        </div>

        <!-- Card 4: External DB & Feature Notes -->
        <div
          class="relative p-4 rounded-2xl bg-white dark:bg-gray-900 border transition-all cursor-pointer hover:shadow-md"
          :class="activeKpiFilter === 'EXTERNAL'
            ? 'border-sky-500 ring-2 ring-sky-500/20'
            : 'border-gray-200 dark:border-gray-800 hover:border-sky-300 dark:hover:border-sky-700'"
          @click="applyKpiFilter('EXTERNAL')"
        >
          <div class="flex items-start justify-between">
            <div>
              <p class="text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('analytics.nlp.kpiExternalDb') }}
              </p>
              <p class="text-3xl font-extrabold text-sky-600 dark:text-sky-400 mt-1">{{ kpiStats.extDbCount + kpiStats.noteCount }}</p>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 mt-2">
                Database Eksternal & Catatan Fitur
              </p>
            </div>
            <div class="p-3 rounded-xl bg-sky-50 dark:bg-sky-950/60 text-sky-600 dark:text-sky-400">
              <UIcon name="i-heroicons-circle-stack" class="w-6 h-6" />
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-[11px] font-semibold text-sky-600 dark:text-sky-400 border-t border-gray-100 dark:border-gray-800 pt-2">
            <span>Filter Sumber Eksternal</span>
            <UIcon name="i-heroicons-chevron-right" class="w-3.5 h-3.5" />
          </div>
        </div>
      </div>

      <!-- Charts Section -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Chart 1: Risk Category Distribution -->
        <UCard class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-xs">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-chart-pie" class="w-5 h-5 text-indigo-500" />
                <h3 class="font-bold text-sm text-gray-900 dark:text-white">{{ t('analytics.nlp.chartCategoryTitle') }}</h3>
              </div>
              <UBadge color="neutral" variant="subtle" size="xs">
                {{ availableCategories.length }} Kategori Aktif
              </UBadge>
            </div>
          </template>
          <div class="h-64 sm:h-72">
            <Doughnut :data="categoryDoughnutData" :options="doughnutOptions" />
          </div>
        </UCard>

        <!-- Chart 2: Control Indicator / Sentiment Distribution -->
        <UCard class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-xs">
          <template #header>
            <div class="flex items-center justify-between gap-2 flex-wrap">
              <div class="flex items-center gap-2">
                <UIcon name="i-heroicons-presentation-chart-line" class="w-5 h-5 text-violet-500" />
                <h3 class="font-bold text-sm text-gray-900 dark:text-white">{{ t('analytics.nlp.chartSentimentTitle') }}</h3>
              </div>
              
              <!-- Mode Toggle -->
              <div class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 p-0.5 rounded-xl text-xs">
                <button
                  type="button"
                  class="px-2.5 py-1 rounded-lg font-semibold transition-all text-xs"
                  :class="chartMode === 'indicator'
                    ? 'bg-white dark:bg-gray-900 text-indigo-600 dark:text-indigo-400 shadow-xs'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'"
                  @click="chartMode = 'indicator'"
                >
                  Control Indicator
                </button>
                <button
                  type="button"
                  class="px-2.5 py-1 rounded-lg font-semibold transition-all text-xs"
                  :class="chartMode === 'sentiment'
                    ? 'bg-white dark:bg-gray-900 text-indigo-600 dark:text-indigo-400 shadow-xs'
                    : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'"
                  @click="chartMode = 'sentiment'"
                >
                  Sentiment
                </button>
              </div>
            </div>
          </template>
          <div class="h-64 sm:h-72">
            <Doughnut :data="sentimentOrIndicatorDoughnutData" :options="doughnutOptions" />
          </div>
        </UCard>
      </div>

      <!-- Main Data Exploration Card -->
      <UCard class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-xs overflow-hidden">
        <template #header>
          <div class="space-y-3">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <div class="flex items-center gap-2.5">
                <div class="p-1.5 rounded-lg bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
                  <UIcon name="i-heroicons-document-text" class="w-5 h-5" />
                </div>
                <div>
                  <h3 class="font-bold text-gray-900 dark:text-white text-base">{{ t('analytics.nlp.tableTitle') }}</h3>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    Menampilkan {{ filteredDocuments.length }} hasil analisis dari seluruh sumber data
                  </p>
                </div>
              </div>

              <!-- View Mode Toggle (Table / Card) on Desktop -->
              <div class="flex items-center gap-2">
                <UButton
                  v-if="hasManualImports"
                  icon="i-heroicons-trash"
                  color="rose"
                  variant="soft"
                  size="xs"
                  class="font-medium"
                  :loading="isDeletingDoc"
                  @click="handleClearAllManualImports"
                >
                  Bersihkan Impor Manual
                </UButton>

                <UButton
                  v-if="isFilterActive"
                  icon="i-heroicons-arrow-path"
                  color="neutral"
                  variant="ghost"
                  size="xs"
                  @click="resetAllFilters"
                >
                  Reset Filter
                </UButton>

                <div class="hidden md:flex items-center bg-gray-100 dark:bg-gray-800 p-0.5 rounded-xl">
                  <button
                    type="button"
                    class="p-1.5 rounded-lg transition-all"
                    :class="viewMode === 'table' ? 'bg-white dark:bg-gray-900 text-indigo-600 shadow-xs' : 'text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
                    title="Tampilan Tabel"
                    @click="viewMode = 'table'"
                  >
                    <UIcon name="i-heroicons-table-cells" class="w-4 h-4" />
                  </button>
                  <button
                    type="button"
                    class="p-1.5 rounded-lg transition-all"
                    :class="viewMode === 'cards' ? 'bg-white dark:bg-gray-900 text-indigo-600 shadow-xs' : 'text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
                    title="Tampilan Kartu"
                    @click="viewMode = 'cards'"
                  >
                    <UIcon name="i-heroicons-squares-2x2" class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Responsive Filter & Search Bar -->
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5 pt-1">
              <!-- Search Input -->
              <div class="relative">
                <UIcon name="i-heroicons-magnifying-glass" class="w-4 h-4 absolute left-3 top-2.5 text-gray-400" />
                <input
                  v-model="searchQuery"
                  type="text"
                  placeholder="Cari temuan, dokumen, OCR..."
                  class="w-full text-xs rounded-xl border border-gray-200 dark:border-gray-700 bg-gray-50/80 dark:bg-gray-800 pl-9 pr-8 py-2 focus:ring-2 focus:ring-indigo-500 focus:outline-none transition-all"
                />
                <button
                  v-if="searchQuery"
                  type="button"
                  class="absolute right-2.5 top-2.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                  @click="searchQuery = ''"
                >
                  <UIcon name="i-heroicons-x-mark" class="w-4 h-4" />
                </button>
              </div>

              <!-- Source Type Filter -->
              <select
                v-model="selectedSourceType"
                class="text-xs rounded-xl border border-gray-200 dark:border-gray-700 bg-gray-50/80 dark:bg-gray-800 px-3 py-2 font-medium focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              >
                <option value="ALL">Semua Sumber Data</option>
                <option value="Imported Document">Dokumen Impor (PDF/Word/Excel)</option>
                <option value="Internal Note">Kolom Catatan Fitur Internal</option>
                <option value="External Database">Database Eksternal (Data Sources)</option>
              </select>

              <!-- Category Filter -->
              <select
                v-model="selectedCategory"
                class="text-xs rounded-xl border border-gray-200 dark:border-gray-700 bg-gray-50/80 dark:bg-gray-800 px-3 py-2 font-medium focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              >
                <option value="ALL">Semua Risk Category</option>
                <option v-for="cat in availableCategories" :key="cat" :value="cat">{{ cat }}</option>
              </select>

              <!-- Indicator Filter -->
              <select
                v-model="selectedIndicator"
                class="text-xs rounded-xl border border-gray-200 dark:border-gray-700 bg-gray-50/80 dark:bg-gray-800 px-3 py-2 font-medium focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              >
                <option value="ALL">Semua Control Indicator</option>
                <option v-for="ind in availableIndicators" :key="ind" :value="ind">{{ ind }}</option>
              </select>
            </div>
          </div>
        </template>

        <!-- No Data State -->
        <div v-if="filteredDocuments.length === 0" class="py-16 text-center space-y-3">
          <div class="w-12 h-12 mx-auto rounded-full bg-gray-100 dark:bg-gray-800 flex items-center justify-center text-gray-400">
            <UIcon name="i-heroicons-document-magnifying-glass" class="w-6 h-6" />
          </div>
          <p class="font-bold text-sm text-gray-700 dark:text-gray-300">Tidak ada dokumen atau catatan yang sesuai filter</p>
          <p class="text-xs text-gray-400 max-w-sm mx-auto">
            Coba ubah kata kunci pencarian atau reset filter untuk menampilkan seluruh data.
          </p>
          <UButton color="neutral" variant="soft" size="xs" @click="resetAllFilters">Reset Semua Filter</UButton>
        </div>

        <!-- 1. DESKTOP TABLE VIEW (Visible on md+ screens when viewMode === 'table') -->
        <div v-else-if="viewMode === 'table'" class="hidden md:block overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50/80 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800 text-[11px] uppercase tracking-wider text-gray-500 dark:text-gray-400">
                <th class="text-left py-3.5 px-4 font-bold">{{ t('analytics.nlp.colDocID') }}</th>
                <th class="text-left py-3.5 px-4 font-bold min-w-[240px]">{{ t('analytics.nlp.colTitle') }}</th>
                <th class="text-left py-3.5 px-3 font-bold">{{ t('analytics.nlp.colSource') }}</th>
                <th class="text-center py-3.5 px-3 font-bold">{{ t('analytics.nlp.colCategory') }}</th>
                <th class="text-center py-3.5 px-3 font-bold min-w-[170px]">{{ t('analytics.nlp.colSentiment') }}</th>
                <th class="text-left py-3.5 px-4 font-bold min-w-[280px]">{{ t('analytics.nlp.colExtraction') }}</th>
                <th class="text-center py-3.5 px-3 font-bold">{{ t('analytics.nlp.colRiskLevel') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr
                v-for="doc in paginatedDocuments"
                :key="doc.docId"
                class="hover:bg-indigo-50/30 dark:hover:bg-indigo-950/20 transition-colors align-top cursor-pointer group"
                @click="openDocumentDetail(doc)"
              >
                <!-- Doc ID & Format -->
                <td class="py-3.5 px-4">
                  <div class="font-mono font-bold text-xs text-indigo-600 dark:text-indigo-400 group-hover:underline">
                    {{ doc.docId }}
                  </div>
                  <div class="mt-1 flex items-center gap-1.5">
                    <UBadge color="neutral" variant="soft" size="xs">{{ doc.fileType || 'PDF' }}</UBadge>
                  </div>
                  <div class="text-[10px] text-gray-400 mt-1">{{ doc.date }}</div>
                </td>

                <!-- Title & Excerpt -->
                <td class="py-3.5 px-4">
                  <div class="font-bold text-sm text-gray-900 dark:text-gray-100 leading-snug group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                    {{ doc.title }}
                  </div>
                  <div v-if="doc.fileName" class="text-[11px] text-indigo-500 font-mono truncate mt-0.5">
                    {{ doc.fileName }}
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 italic line-clamp-2 leading-relaxed">
                    "{{ doc.excerpt }}"
                  </p>
                  <div v-if="doc.extractionMethod" class="text-[10px] text-gray-400 mt-1 flex items-center gap-1">
                    <UIcon name="i-heroicons-cpu-chip" class="w-3 h-3 text-emerald-500 shrink-0" />
                    <span>{{ doc.extractionMethod }}</span>
                  </div>
                </td>

                <!-- Origin Source -->
                <td class="py-3.5 px-3">
                  <UBadge :color="sourceTypeBadgeColor(doc.sourceType)" variant="subtle" size="sm" class="font-medium whitespace-nowrap">
                    {{ doc.source }}
                  </UBadge>
                  <div class="text-[10px] text-gray-400 mt-1 font-medium">
                    {{ doc.sourceType || 'Imported Document' }}
                  </div>
                </td>

                <!-- Risk Category -->
                <td class="text-center py-3.5 px-3">
                  <UBadge :color="riskCategoryColor(doc.autoCategory)" variant="subtle" size="md" class="font-bold whitespace-nowrap">
                    {{ doc.autoCategory }}
                  </UBadge>
                  <div class="text-[10px] text-gray-400 mt-1 font-medium">
                    Akurasi: {{ Math.round((doc.confidence || 0.92) * 100) }}%
                  </div>
                </td>

                <!-- Sentiment, Control Indicator & Severity -->
                <td class="text-center py-3.5 px-3">
                  <div class="flex flex-col items-center gap-1.5">
                    <UBadge :color="controlIndicatorColor(doc.controlIndicator)" variant="solid" size="xs" class="font-bold max-w-[170px] truncate">
                      {{ doc.controlIndicator || 'Internal Control Weakness' }}
                    </UBadge>
                    <div class="flex items-center gap-1.5">
                      <UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="xs">
                        {{ doc.sentiment }}
                      </UBadge>
                      <span class="text-xs font-extrabold text-gray-700 dark:text-gray-300">
                        {{ doc.severityScore }}/100
                      </span>
                    </div>
                    <!-- Severity Progress Bar -->
                    <div class="w-full max-w-[120px] bg-gray-200 dark:bg-gray-700 h-1.5 rounded-full overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all"
                        :class="doc.severityScore >= 80 ? 'bg-rose-500' : (doc.severityScore >= 55 ? 'bg-amber-500' : 'bg-emerald-500')"
                        :style="{ width: `${Math.min(100, Math.max(5, doc.severityScore || 50))}%` }"
                      />
                    </div>
                  </div>
                </td>

                <!-- Extracted Finding & Recommendation -->
                <td class="py-3.5 px-4">
                  <div class="space-y-1.5 text-xs">
                    <div class="p-2 rounded-xl bg-rose-50/70 dark:bg-rose-950/30 border border-rose-200/60 dark:border-rose-800/50">
                      <div class="flex items-center gap-1 font-bold text-rose-700 dark:text-rose-300 uppercase text-[10px] tracking-wider">
                        <UIcon name="i-heroicons-exclamation-triangle" class="w-3.5 h-3.5" />
                        <span>{{ t('analytics.nlp.findingLabel') }}:</span>
                      </div>
                      <p class="text-gray-700 dark:text-gray-200 line-clamp-2 mt-0.5 leading-snug">
                        {{ doc.extractedFinding || doc.excerpt }}
                      </p>
                    </div>
                    <div class="p-2 rounded-xl bg-emerald-50/70 dark:bg-emerald-950/30 border border-emerald-200/60 dark:border-emerald-800/50">
                      <div class="flex items-center gap-1 font-bold text-emerald-700 dark:text-emerald-300 uppercase text-[10px] tracking-wider">
                        <UIcon name="i-heroicons-light-bulb" class="w-3.5 h-3.5" />
                        <span>{{ t('analytics.nlp.recommendationLabel') }}:</span>
                      </div>
                      <p class="text-gray-700 dark:text-gray-200 line-clamp-2 mt-0.5 leading-snug">
                        {{ doc.extractedRecommendation }}
                      </p>
                    </div>
                  </div>
                </td>

                <!-- Risk Level -->
                <td class="text-center py-3.5 px-3">
                  <UBadge
                    :class="getRiskConfig(doc.riskLevel).badgeClass"
                    variant="solid"
                    size="md"
                    class="font-bold"
                  >
                    {{ getRiskConfig(doc.riskLevel).label }}
                  </UBadge>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 2. RESPONSIVE CARD VIEW (Visible on mobile/tablet or when cards viewMode is selected) -->
        <div v-else class="p-4 grid grid-cols-1 md:grid-cols-2 gap-4">
          <div
            v-for="doc in paginatedDocuments"
            :key="doc.docId"
            class="p-4 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900/90 shadow-xs hover:border-indigo-400 dark:hover:border-indigo-600 transition-all cursor-pointer flex flex-col justify-between space-y-3"
            @click="openDocumentDetail(doc)"
          >
            <!-- Card Header -->
            <div class="space-y-2">
              <div class="flex items-center justify-between gap-2 flex-wrap">
                <div class="flex items-center gap-2">
                  <span class="font-mono font-bold text-xs text-indigo-600 dark:text-indigo-400">{{ doc.docId }}</span>
                  <UBadge color="neutral" variant="soft" size="xs">{{ doc.fileType || 'PDF' }}</UBadge>
                </div>
                <UBadge
                  :class="getRiskConfig(doc.riskLevel).badgeClass"
                  variant="solid"
                  size="xs"
                  class="font-bold"
                >
                  {{ getRiskConfig(doc.riskLevel).label }}
                </UBadge>
              </div>

              <!-- Title -->
              <h4 class="font-bold text-sm text-gray-900 dark:text-white leading-snug">
                {{ doc.title }}
              </h4>

              <!-- Source & Date -->
              <div class="flex items-center gap-2 text-[11px] text-gray-500">
                <UBadge :color="sourceTypeBadgeColor(doc.sourceType)" variant="subtle" size="xs">
                  {{ doc.source }}
                </UBadge>
                <span>•</span>
                <span>{{ doc.date }}</span>
              </div>
            </div>

            <!-- Category & Indicator Pills -->
            <div class="flex flex-wrap items-center gap-1.5 pt-1">
              <UBadge :color="riskCategoryColor(doc.autoCategory)" variant="subtle" size="xs" class="font-semibold">
                {{ doc.autoCategory }}
              </UBadge>
              <UBadge :color="controlIndicatorColor(doc.controlIndicator)" variant="solid" size="xs" class="font-bold">
                {{ doc.controlIndicator }}
              </UBadge>
              <UBadge :color="sentimentColor(doc.sentiment)" variant="subtle" size="xs">
                {{ doc.sentiment }} ({{ doc.severityScore }}/100)
              </UBadge>
            </div>

            <!-- Finding Snippet -->
            <div class="p-2.5 rounded-xl bg-rose-50/70 dark:bg-rose-950/30 border border-rose-200/60 dark:border-rose-800/40 text-xs">
              <div class="flex items-center gap-1 font-bold text-rose-700 dark:text-rose-300 text-[10px] uppercase">
                <UIcon name="i-heroicons-exclamation-triangle" class="w-3.5 h-3.5" />
                <span>{{ t('analytics.nlp.findingLabel') }}</span>
              </div>
              <p class="text-gray-700 dark:text-gray-300 line-clamp-2 mt-1 leading-snug">
                {{ doc.extractedFinding || doc.excerpt }}
              </p>
            </div>

            <!-- Footer Action -->
            <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800 text-xs text-indigo-600 dark:text-indigo-400 font-semibold">
              <span class="text-[11px] text-gray-400">Akurasi AI: {{ Math.round((doc.confidence || 0.92) * 100) }}%</span>
              <div class="flex items-center gap-1">
                <span>Lihat Detail</span>
                <UIcon name="i-heroicons-arrow-right" class="w-3.5 h-3.5" />
              </div>
            </div>
          </div>
        </div>

        <!-- Responsive Pagination Footer -->
        <template #footer>
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 text-xs text-gray-500 dark:text-gray-400 py-1">
            <div>
              Menampilkan
              <span class="font-bold text-gray-800 dark:text-gray-200">{{ Math.min((currentPage - 1) * itemsPerPage + 1, filteredDocuments.length) }}</span>
              -
              <span class="font-bold text-gray-800 dark:text-gray-200">{{ Math.min(currentPage * itemsPerPage, filteredDocuments.length) }}</span>
              dari
              <span class="font-bold text-gray-800 dark:text-gray-200">{{ filteredDocuments.length }}</span>
              dokumen
            </div>

            <div class="flex items-center gap-1.5 self-end sm:self-auto">
              <UButton
                icon="i-heroicons-chevron-left"
                color="neutral"
                variant="soft"
                size="xs"
                :disabled="currentPage <= 1"
                @click="currentPage--"
              >
                Sebelumnya
              </UButton>

              <span class="px-2 py-1 font-semibold text-gray-700 dark:text-gray-300">
                {{ currentPage }} / {{ totalPages }}
              </span>

              <UButton
                trailing-icon="i-heroicons-chevron-right"
                color="neutral"
                variant="soft"
                size="xs"
                :disabled="currentPage >= totalPages"
                @click="currentPage++"
              >
                Berikutnya
              </UButton>
            </div>
          </div>
        </template>
      </UCard>
    </div>

    <!-- MODAL 1: Impor & Analisa Dokumen (PDF, Word, Excel, Notes) -->
    <div
      v-if="showUploadModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
      @click.self="showUploadModal = false"
    >
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl max-w-xl w-full p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
              <UIcon name="i-heroicons-document-arrow-up" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="font-bold text-base text-gray-900 dark:text-white">{{ t('analytics.nlp.btnImportAnalyze') }}</h3>
              <p class="text-xs text-gray-500">Ekstraksi teks otomatis & klasifikasi IndoBERT</p>
            </div>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" size="xs" @click="showUploadModal = false" />
        </div>

        <div class="space-y-4 text-sm">
          <!-- Drag and Drop Area -->
          <div>
            <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
              Unggah File Dokumen (PDF, Word .docx/.doc, Excel .xlsx/.csv)
            </label>
            <div
              class="relative border-2 border-dashed rounded-2xl p-6 text-center cursor-pointer transition-all"
              :class="isDragging
                ? 'border-indigo-500 bg-indigo-50/70 dark:bg-indigo-950/40'
                : (uploadForm.fileName
                  ? 'border-emerald-400 bg-emerald-50/30 dark:bg-emerald-950/20'
                  : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50/50 dark:bg-gray-800/40')"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="handleFileDrop"
            >
              <input
                type="file"
                accept=".pdf,.doc,.docx,.xls,.xlsx,.csv,.txt"
                class="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                @change="handleFileUpload"
              />

              <div v-if="!uploadForm.fileName" class="space-y-2">
                <UIcon name="i-heroicons-arrow-up-tray" class="w-8 h-8 mx-auto text-indigo-500" />
                <p class="text-xs font-bold text-gray-700 dark:text-gray-200">
                  Tarik & Lepas file ke sini, atau <span class="text-indigo-600 dark:text-indigo-400 underline">pilih dari komputer</span>
                </p>
                <p class="text-[11px] text-gray-400">
                  Mendukung PDF (Teks & OCR Scanner), Word DOCX/DOC, Excel XLSX/CSV
                </p>
              </div>

              <!-- Selected File Card -->
              <div v-else class="flex items-center justify-between p-3 rounded-xl bg-white dark:bg-gray-800 border border-emerald-200 dark:border-emerald-800">
                <div class="flex items-center gap-3 text-left">
                  <div class="p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/60 text-emerald-600 dark:text-emerald-300">
                    <UIcon name="i-heroicons-document-check" class="w-6 h-6" />
                  </div>
                  <div>
                    <p class="font-bold text-xs text-gray-900 dark:text-white truncate max-w-xs">{{ uploadForm.fileName }}</p>
                    <p class="text-[10px] text-gray-500">
                      Format: {{ uploadForm.fileType }} • Ukuran: {{ uploadForm.fileSizeText }}
                    </p>
                  </div>
                </div>
                <button
                  type="button"
                  class="p-1 rounded-lg text-gray-400 hover:text-rose-500 transition-colors"
                  title="Hapus file"
                  @click.stop="clearSelectedFile"
                >
                  <UIcon name="i-heroicons-trash" class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>

          <!-- Metadata Fields -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
                Judul Dokumen / Catatan
              </label>
              <input
                v-model="uploadForm.title"
                type="text"
                placeholder="Contoh: LHA Pemeriksaan Kredit Cabang..."
                class="w-full text-xs rounded-xl border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              />
            </div>

            <div>
              <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
                Sumber Fitur / Database
              </label>
              <select
                v-model="uploadForm.source"
                class="w-full text-xs rounded-xl border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
                @change="onSourceOptionChange"
              >
                <option v-for="opt in sourceOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </div>
          </div>

          <!-- Direct Text Input / Extracted Content Preview -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block text-xs font-semibold text-gray-700 dark:text-gray-300">
                Isi Teks / Hasil Ekstraksi Otomatis Dokumen
              </label>
              <UBadge
                v-if="uploadForm.text.trim()"
                color="success"
                variant="subtle"
                size="xs"
                class="font-medium"
              >
                {{ uploadForm.text.length }} karakter terekstrak
              </UBadge>
              <span v-else-if="isReadingFile" class="text-[11px] text-indigo-500 animate-pulse font-medium">
                Mengekstrak teks dokumen...
              </span>
            </div>
            <textarea
              v-model="uploadForm.text"
              rows="5"
              placeholder="Teks akan otomatis diekstrak saat Anda memilih file (PDF, Word, Excel, CSV), atau Anda dapat menempelkan langsung teks audit di sini..."
              class="w-full text-xs rounded-xl border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2.5 focus:ring-2 focus:ring-indigo-500 focus:outline-none leading-relaxed font-mono"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-gray-200 dark:border-gray-800">
          <UButton color="neutral" variant="ghost" size="sm" @click="showUploadModal = false">Batal</UButton>
          <UButton
            color="primary"
            size="sm"
            icon="i-heroicons-sparkles"
            :loading="isAnalyzingDoc || isReadingFile"
            :disabled="isReadingFile || (!uploadForm.fileName && !uploadForm.fileContentBase64 && !uploadForm.text.trim() && !uploadForm.title.trim())"
            @click="submitDocumentForAnalysis"
          >
            {{ isReadingFile ? 'Membaca File...' : 'Ekstrak OCR & Analisis IndoBERT' }}
          </UButton>
        </div>
      </div>
    </div>

    <!-- MODAL 2: Kelola Dynamic Risk Categories & Control Indicators -->
    <div
      v-if="showTaxonomyModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
      @click.self="showTaxonomyModal = false"
    >
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl max-w-2xl w-full p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-xl bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400">
              <UIcon name="i-heroicons-tag" class="w-5 h-5" />
            </div>
            <div>
              <h3 class="font-bold text-base text-gray-900 dark:text-white">Kelola Dynamic Risk Category & Control Indicator</h3>
              <p class="text-xs text-gray-500">Dukungan taksonomi dinamis model IndoBERT NLP</p>
            </div>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" size="xs" @click="showTaxonomyModal = false" />
        </div>

        <!-- Taxonomy Navigation Tabs -->
        <div class="flex items-center gap-2 border-b border-gray-100 dark:border-gray-800 pb-2">
          <button
            type="button"
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all"
            :class="activeTaxonomyTab === 'categories'
              ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-500 hover:text-gray-800 dark:hover:text-gray-200'"
            @click="activeTaxonomyTab = 'categories'"
          >
            Risk Category ({{ availableCategories.length }})
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition-all"
            :class="activeTaxonomyTab === 'indicators'
              ? 'bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-500 hover:text-gray-800 dark:hover:text-gray-200'"
            @click="activeTaxonomyTab = 'indicators'"
          >
            Control Indicator ({{ availableIndicators.length }})
          </button>
        </div>

        <!-- Active Items List -->
        <div v-if="activeTaxonomyTab === 'categories'" class="space-y-2">
          <p class="text-xs font-bold text-gray-600 dark:text-gray-300">Daftar Risk Category Aktif:</p>
          <div class="flex flex-wrap gap-2">
            <UBadge
              v-for="cat in availableCategories"
              :key="cat"
              :color="riskCategoryColor(cat)"
              variant="subtle"
              size="md"
              class="font-semibold"
            >
              {{ cat }}
            </UBadge>
          </div>
        </div>

        <div v-else class="space-y-2">
          <p class="text-xs font-bold text-gray-600 dark:text-gray-300">Daftar Control Indicator Aktif:</p>
          <div class="flex flex-wrap gap-2">
            <UBadge
              v-for="ind in availableIndicators"
              :key="ind"
              :color="controlIndicatorColor(ind)"
              variant="subtle"
              size="md"
              class="font-semibold"
            >
              {{ ind }}
            </UBadge>
          </div>
        </div>

        <!-- Add Form -->
        <div class="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/60 border border-gray-200 dark:border-gray-700 space-y-3">
          <h4 class="font-bold text-xs uppercase tracking-wider text-indigo-600 dark:text-indigo-400">
            Tambah Taksonomi Baru
          </h4>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs">
            <div>
              <label class="block font-semibold text-gray-700 dark:text-gray-300 mb-1">Jenis Taksonomi</label>
              <select
                v-model="taxonomyForm.itemType"
                class="w-full text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2"
              >
                <option value="risk_category">Risk Category (misal: Strategic, ESG, Reputational)</option>
                <option value="control_indicator">Control Indicator (misal: Vendor SLA Breach, Data Privacy)</option>
              </select>
            </div>

            <div>
              <label class="block font-semibold text-gray-700 dark:text-gray-300 mb-1">Nama Kategori / Indikator</label>
              <input
                v-model="taxonomyForm.name"
                type="text"
                placeholder="Contoh: Cybersecurity, ESG, Data Privacy..."
                class="w-full text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2"
              />
            </div>
          </div>

          <div class="text-xs">
            <label class="block font-semibold text-gray-700 dark:text-gray-300 mb-1">
              Kata Kunci Pendeteksi IndoBERT (pisahkan dengan koma)
            </label>
            <input
              v-model="taxonomyForm.keywordsText"
              type="text"
              placeholder="Contoh: phising, kebocoran data, ransomware, hak akses pihak ketiga"
              class="w-full text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2"
            />
          </div>

          <div class="text-xs">
            <label class="block font-semibold text-gray-700 dark:text-gray-300 mb-1">Deskripsi Singkat</label>
            <input
              v-model="taxonomyForm.description"
              type="text"
              placeholder="Deskripsi untuk dokumentasi model..."
              class="w-full text-xs rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-800 px-3 py-2"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-gray-200 dark:border-gray-800">
          <UButton color="neutral" variant="ghost" size="sm" @click="showTaxonomyModal = false">Tutup</UButton>
          <UButton
            color="primary"
            size="sm"
            icon="i-heroicons-plus-circle"
            :loading="isSavingTaxonomy"
            @click="submitNewTaxonomyItem"
          >
            Simpan & Latih Model
          </UButton>
        </div>
      </div>
    </div>

    <!-- MODAL 3: Detail Analisis Lengkap Dokumen NLP -->
    <div
      v-if="showDetailModal && selectedDocDetail"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
      @click.self="showDetailModal = false"
    >
      <div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl max-w-2xl w-full p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <!-- Header -->
        <div class="flex items-start justify-between border-b border-gray-200 dark:border-gray-800 pb-3">
          <div>
            <div class="flex items-center gap-2 flex-wrap">
              <span class="font-mono text-xs font-bold text-indigo-600 dark:text-indigo-400">{{ selectedDocDetail.docId }}</span>
              <UBadge :color="sourceTypeBadgeColor(selectedDocDetail.sourceType)" variant="subtle" size="xs">
                {{ selectedDocDetail.source }}
              </UBadge>
              <UBadge color="neutral" variant="soft" size="xs">{{ selectedDocDetail.fileType || 'PDF' }}</UBadge>
            </div>
            <h3 class="font-bold text-base text-gray-900 dark:text-white mt-1.5">{{ selectedDocDetail.title }}</h3>
          </div>
          <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" size="xs" @click="showDetailModal = false" />
        </div>

        <!-- Metrics Grid -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/70 border border-gray-100 dark:border-gray-700">
            <span class="text-gray-400 block text-[10px] uppercase font-bold">Risk Category</span>
            <UBadge :color="riskCategoryColor(selectedDocDetail.autoCategory)" variant="subtle" size="sm" class="mt-1 font-bold">
              {{ selectedDocDetail.autoCategory }}
            </UBadge>
          </div>

          <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/70 border border-gray-100 dark:border-gray-700">
            <span class="text-gray-400 block text-[10px] uppercase font-bold">Control Indicator</span>
            <UBadge :color="controlIndicatorColor(selectedDocDetail.controlIndicator)" variant="solid" size="xs" class="mt-1 font-bold">
              {{ selectedDocDetail.controlIndicator }}
            </UBadge>
          </div>

          <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/70 border border-gray-100 dark:border-gray-700">
            <span class="text-gray-400 block text-[10px] uppercase font-bold">Severity Score</span>
            <div class="flex items-center gap-1.5 mt-1">
              <span class="font-bold text-sm text-gray-900 dark:text-white">{{ selectedDocDetail.severityScore }}/100</span>
              <UBadge :color="sentimentColor(selectedDocDetail.sentiment)" variant="subtle" size="xs">
                {{ selectedDocDetail.sentiment }}
              </UBadge>
            </div>
          </div>

          <div class="p-3 rounded-xl bg-gray-50 dark:bg-gray-800/70 border border-gray-100 dark:border-gray-700">
            <span class="text-gray-400 block text-[10px] uppercase font-bold">Engine & Akurasi</span>
            <p class="font-semibold text-gray-800 dark:text-gray-200 mt-1 truncate">
              {{ Math.round((selectedDocDetail.confidence || 0.92) * 100) }}% Confidence
            </p>
          </div>
        </div>

        <!-- Finding & Recommendation Sections -->
        <div class="space-y-3 text-xs">
          <!-- Finding -->
          <div class="p-4 rounded-xl bg-rose-50/80 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-800 space-y-1.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5 font-bold text-rose-700 dark:text-rose-300 uppercase tracking-wider text-[11px]">
                <UIcon name="i-heroicons-exclamation-triangle" class="w-4 h-4" />
                <span>Ekstraksi Temuan Audit (Extracted Finding)</span>
              </div>
              <button
                type="button"
                class="text-rose-600 dark:text-rose-400 hover:underline text-[11px] font-semibold"
                @click="copyToClipboard(selectedDocDetail.extractedFinding || selectedDocDetail.excerpt)"
              >
                Salin Temuan
              </button>
            </div>
            <p class="text-gray-800 dark:text-gray-200 leading-relaxed font-medium">
              {{ selectedDocDetail.extractedFinding || selectedDocDetail.excerpt }}
            </p>
          </div>

          <!-- Recommendation -->
          <div class="p-4 rounded-xl bg-emerald-50/80 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800 space-y-1.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5 font-bold text-emerald-700 dark:text-emerald-300 uppercase tracking-wider text-[11px]">
                <UIcon name="i-heroicons-light-bulb" class="w-4 h-4" />
                <span>Ekstraksi Rekomendasi Tindak Lanjut (Extracted Recommendation)</span>
              </div>
              <button
                type="button"
                class="text-emerald-600 dark:text-emerald-400 hover:underline text-[11px] font-semibold"
                @click="copyToClipboard(selectedDocDetail.extractedRecommendation)"
              >
                Salin Rekomendasi
              </button>
            </div>
            <p class="text-gray-800 dark:text-gray-200 leading-relaxed font-medium">
              {{ selectedDocDetail.extractedRecommendation }}
            </p>
          </div>

          <!-- Raw Excerpt / OCR -->
          <div class="p-3.5 rounded-xl bg-gray-50 dark:bg-gray-800/70 border border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between text-gray-500 font-bold uppercase tracking-wider text-[10px] mb-1">
              <span>Kutipan Teks Asli / Hasil OCR ({{ selectedDocDetail.fileName }})</span>
              <span v-if="selectedDocDetail.extractionMethod">{{ selectedDocDetail.extractionMethod }}</span>
            </div>
            <p class="text-gray-600 dark:text-gray-300 italic leading-relaxed text-xs">
              "{{ selectedDocDetail.excerpt }}"
            </p>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
          <UButton
            color="rose"
            variant="ghost"
            size="sm"
            icon="i-heroicons-trash"
            :loading="isDeletingDoc"
            @click="handleDeleteDocument(selectedDocDetail.docId)"
          >
            Hapus Dokumen Ini
          </UButton>
          <UButton color="neutral" variant="soft" size="sm" @click="showDetailModal = false">Tutup</UButton>
        </div>
      </div>
    </div>
  </div>
</template>
