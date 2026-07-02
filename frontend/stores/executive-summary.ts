import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'

export interface FollowUpRow {
  status: 'Closed' | 'In Progress' | 'Overdue'
  jumlah: number
  persentase: number
  keterangan: string
}

export interface SignificantFinding {
  unitDivision: string
  judulTemuan: string
  risiko: 'Tinggi' | 'Sedang' | 'Rendah'
  statusTL: 'Closed' | 'In Progress' | 'Overdue'
  usulan: string
}

export interface MatriksRow {
  nomor: string
  division: string
  unitKerja: string
  prosesBisnis: string
  judulTemuan: string
  nilaiRisiko: 'Tinggi' | 'Sedang' | 'Rendah'
  rekomendasi: string
  dueDate: string
  picUnit: string
  progres: number
  status: 'Closed' | 'In Progress' | 'Overdue'
  buktiTL: string
}

export interface ExecutiveSummary {
  id: string
  quarter: number // 1, 2, 3, 4
  periodeBulan: string // e.g. "Maret"
  tahun: number // default 2026
  nomorDokumen: string
  dokumenPath: string
  status: 'Draft' | 'Approved' | 'Rejected'

  // Section I
  narrative: string

  // Section II
  jumlahLaporan: number
  risikoTinggi: number
  risikoSedang: number
  risikoRendah: number
  jumlahRekomendasi: number

  // Section III (JSON string stored in DB, array in store)
  followUpTable: FollowUpRow[]

  // Section IV (JSON string stored in DB, array in store)
  topFindings: SignificantFinding[]

  // Section VIII (JSON string stored in DB, array in store)
  matriksKompilasi: MatriksRow[]

  // Section V & VII
  akarMasalah: string
  kesimpulan: string

  // Signatures
  signatureTempat: string
  signatureTanggal: string
  signatureNamaKepala: string
  signatureNIK: string

  created_at?: string
  updated_at?: string
}

export const useExecutiveSummaryStore = defineStore('executive-summary', () => {
  const summaryList = ref<ExecutiveSummary[]>([])
  const currentSummary = ref<ExecutiveSummary | null>(null)

  const showModal = ref(false)
  const isEditing = ref(false)
  const isViewing = ref(false)
  const loading = ref(false)
  const errorMsg = ref('')

  // Default narrative template
  const defaultNarrativeTemplate = (bulan: string, tahun: number = 2026) => {
    return `Periode ${bulan} ${tahun}, SPI telah menerbitkan .... LHA. Total temuan …. dengan .... rekomendasi. Tingkat penyelesaian on-time 95%. Terdapat 5 temuan risiko tinggi overdue/terlambat terkait .... yang perlu arahan Direksi. Temuan berulang tertinggi pada proses ……………..`
  }

  // Active form data
  const form = reactive<Omit<ExecutiveSummary, 'id'>>({
    quarter: 1,
    periodeBulan: 'Januari',
    tahun: 2026,
    nomorDokumen: '',
    dokumenPath: '',
    status: 'Draft',
    narrative: '',
    jumlahLaporan: 0,
    risikoTinggi: 0,
    risikoSedang: 0,
    risikoRendah: 0,
    jumlahRekomendasi: 0,
    followUpTable: [
      { status: 'Closed', jumlah: 0, persentase: 0, keterangan: '' },
      { status: 'In Progress', jumlah: 0, persentase: 0, keterangan: '' },
      { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '' }
    ],
    topFindings: [],
    matriksKompilasi: [],
    akarMasalah: '',
    kesimpulan: '',
    signatureTempat: '',
    signatureTanggal: '',
    signatureNamaKepala: '',
    signatureNIK: ''
  })

  const getAuditServiceBaseUrl = () => {
    const config = useRuntimeConfig()
    return config.public.auditServiceBaseUrl || 'http://localhost:8080/api/v1'
  }

  const mockSummaries: ExecutiveSummary[] = [
    {
      id: 'ES-Q1-2026-001',
      quarter: 1,
      periodeBulan: 'Maret',
      tahun: 2026,
      nomorDokumen: 'LKA/01/SPI/KAI/2026',
      dokumenPath: 'executive_summary_q1_2026.pdf',
      status: 'Approved',
      narrative: 'Periode Q1 2026, SPI telah menerbitkan 12 LHA. Total temuan 32 dengan 48 rekomendasi. Tingkat penyelesaian on-time 95%. Terdapat 5 temuan risiko tinggi overdue/terlambat terkait IT Security yang perlu arahan Direksi. Temuan berulang tertinggi pada proses Procurement.',
      jumlahLaporan: 12,
      risikoTinggi: 10,
      risikoSedang: 15,
      risikoRendah: 7,
      jumlahRekomendasi: 48,
      followUpTable: [
        { status: 'Closed', jumlah: 96, persentase: 72.7, keterangan: 'Telah divalidasi SPI' },
        { status: 'In Progress', jumlah: 26, persentase: 19.7, keterangan: 'On track, rata2 progres 68%' },
        { status: 'Overdue', jumlah: 10, persentase: 7.6, keterangan: '3 Risiko Tinggi' }
      ],
      topFindings: [
        { unitDivision: 'OP', judulTemuan: 'Keterlambatan Kalibrasi Alat Berat', risiko: 'Tinggi', statusTL: 'Overdue', usulan: 'Eskalasi Direksi' },
        { unitDivision: 'KK/KSD', judulTemuan: 'Ketidakpatuhan Prosedur K3 Tambang', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Audit Khusus' }
      ],
      matriksKompilasi: [
        { nomor: '001/SPI/2026', division: 'OP', unitKerja: 'Division Operation Personnel', prosesBisnis: 'O&M', judulTemuan: 'Keterlambatan Kalibrasi Alat Berat', nilaiRisiko: 'Tinggi', rekomendasi: 'Segera lakukan kalibrasi ulang', dueDate: '2026-03-15', picUnit: 'Manager O&M', progres: 40, status: 'Overdue', buktiTL: 'BA Kalibrasi 1' },
        { nomor: '002/SPI/2026', division: 'KK/KSD', unitKerja: 'Division KSD', prosesBisnis: 'K3', judulTemuan: 'Ketidakpatuhan Prosedur K3 Tambang', nilaiRisiko: 'Tinggi', rekomendasi: 'Sediakan APD tambahan dan lakukan training harian', dueDate: '2026-04-10', picUnit: 'Manager K3', progres: 80, status: 'In Progress', buktiTL: '' }
      ],
      akarMasalah: 'Akar masalah utama adalah keterbatasan SDM tersertifikasi di unit operasi dan lambatnya respon vendor pengadaan.',
      kesimpulan: 'Secara umum, risiko operasional terkendali namun diperlukan perhatian serius direksi untuk masalah kalibrasi alat berat di site OP.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-03-31',
      signatureNamaKepala: 'Budi Santoso, CIA',
      signatureNIK: 'SPI-77621'
    }
  ]

  const fetchSummaries = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const response: any = await $fetch(`${baseUrl}/executive-summaries`, { method: 'GET' })
      let items: any[] = []
      if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && response.data && Array.isArray(response.data.items)) {
                items = response.data.items;
            } else if (response && Array.isArray(response.items)) {
        items = response.items
      } else if (Array.isArray(response)) {
        items = response
      }

      if (items.length > 0) {
        summaryList.value = items.map(item => parseSummaryFromBackend(item))
      } else {
        summaryList.value = [...mockSummaries]
      }
    } catch (error) {
      console.error('Failed to fetch executive summaries, falling back to mock data:', error)
      summaryList.value = [...mockSummaries]
    } finally {
      loading.value = false
    }
  }

  const parseSummaryFromBackend = (item: any): ExecutiveSummary => {
    return {
      ...item,
      quarter: Number(item.quarter),
      tahun: Number(item.tahun),
      jumlahLaporan: Number(item.jumlahLaporan),
      risikoTinggi: Number(item.risikoTinggi),
      risikoSedang: Number(item.risikoSedang),
      risikoRendah: Number(item.risikoRendah),
      jumlahRekomendasi: Number(item.jumlahRekomendasi),
      followUpTable: item.followUpTable ? JSON.parse(item.followUpTable) : [],
      topFindings: item.topFindings ? JSON.parse(item.topFindings) : [],
      matriksKompilasi: item.matriksKompilasi ? JSON.parse(item.matriksKompilasi) : []
    }
  }

  const serializeSummaryForBackend = (data: Omit<ExecutiveSummary, 'id'> | ExecutiveSummary) => {
    return {
      ...data,
      quarter: Number(data.quarter),
      tahun: Number(data.tahun),
      jumlahLaporan: Number(data.jumlahLaporan),
      risikoTinggi: Number(data.risikoTinggi),
      risikoSedang: Number(data.risikoSedang),
      risikoRendah: Number(data.risikoRendah),
      jumlahRekomendasi: Number(data.jumlahRekomendasi),
      followUpTable: JSON.stringify(data.followUpTable),
      topFindings: JSON.stringify(data.topFindings),
      matriksKompilasi: JSON.stringify(data.matriksKompilasi)
    }
  }

  const openNewForm = (quarterNum: number) => {
    isEditing.value = false
    isViewing.value = false
    currentSummary.value = null

    // Set default month based on selected quarter
    let defaultMonth = 'Januari'
    if (quarterNum === 2) defaultMonth = 'April'
    if (quarterNum === 3) defaultMonth = 'Juli'
    if (quarterNum === 4) defaultMonth = 'Oktober'

    Object.assign(form, {
      quarter: quarterNum,
      periodeBulan: defaultMonth,
      tahun: 2026,
      nomorDokumen: '',
      dokumenPath: '',
      status: 'Draft',
      narrative: defaultNarrativeTemplate(defaultMonth, 2026),
      jumlahLaporan: 0,
      risikoTinggi: 0,
      risikoSedang: 0,
      risikoRendah: 0,
      jumlahRekomendasi: 0,
      followUpTable: [
        { status: 'Closed', jumlah: 0, persentase: 0, keterangan: '' },
        { status: 'In Progress', jumlah: 0, persentase: 0, keterangan: '' },
        { status: 'Overdue', jumlah: 0, persentase: 0, keterangan: '' }
      ],
      topFindings: [],
      matriksKompilasi: [],
      akarMasalah: '',
      kesimpulan: '',
      signatureTempat: 'Jakarta',
      signatureTanggal: new Date().toISOString().split('T')[0],
      signatureNamaKepala: '',
      signatureNIK: ''
    })
    showModal.value = true
  }

  const openEditForm = (summary: ExecutiveSummary) => {
    isEditing.value = true
    isViewing.value = false
    currentSummary.value = summary
    Object.assign(form, JSON.parse(JSON.stringify(summary)))
    showModal.value = true
  }

  const openView = (summary: ExecutiveSummary) => {
    isEditing.value = false
    isViewing.value = true
    currentSummary.value = summary
    Object.assign(form, JSON.parse(JSON.stringify(summary)))
    showModal.value = true
  }

  const saveForm = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = serializeSummaryForBackend(form)

      if (isEditing.value && currentSummary.value) {
        await $fetch(`${baseUrl}/executive-summaries/${currentSummary.value.id}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        // Create random ID for fallback/mock
        const tempId = `ES-Q${form.quarter}-2026-${Math.floor(100 + Math.random() * 900)}`
        await $fetch(`${baseUrl}/executive-summaries`, {
          method: 'POST',
          body: { ...payload, id: tempId }
        })
      }
      showModal.value = false
      await fetchSummaries()
    } catch (error: any) {
      console.error('Failed to save summary to backend, simulating local save:', error)
      // Simulating save in state for offline capabilities
      if (isEditing.value && currentSummary.value) {
        const idx = summaryList.value.findIndex(s => s.id === currentSummary.value!.id)
        if (idx !== -1) {
          summaryList.value[idx] = {
            ...currentSummary.value,
            ...JSON.parse(JSON.stringify(form))
          }
        }
      } else {
        const newSummary: ExecutiveSummary = {
          id: `ES-Q${form.quarter}-2026-${Math.floor(100 + Math.random() * 900)}`,
          ...JSON.parse(JSON.stringify(form))
        }
        summaryList.value.push(newSummary)
      }
      showModal.value = false
    } finally {
      loading.value = false
    }
  }

  const deleteSummary = async (id: string) => {
    if (!confirm('Apakah Anda yakin ingin menghapus Laporan Eksekutif ini?')) return
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/executive-summaries/${id}`, { method: 'DELETE' })
      await fetchSummaries()
    } catch (error) {
      console.error('Failed to delete on backend, simulating local deletion:', error)
      summaryList.value = summaryList.value.filter(s => s.id !== id)
    } finally {
      loading.value = false
    }
  }

  const updateStatus = async (id: string, newStatus: 'Draft' | 'Approved' | 'Rejected') => {
    loading.value = true
    try {
      const baseUrl = getAuditServiceBaseUrl()
      // Fetch current item
      const item = summaryList.value.find(s => s.id === id)
      if (item) {
        const updated = { ...item, status: newStatus }
        const payload = serializeSummaryForBackend(updated)
        await $fetch(`${baseUrl}/executive-summaries/${id}`, {
          method: 'PUT',
          body: payload
        })
        await fetchSummaries()
      }
    } catch (error) {
      console.error('Failed to update status on backend, simulating local update:', error)
      const idx = summaryList.value.findIndex(s => s.id === id)
      if (idx !== -1 && summaryList.value[idx]) {
        summaryList.value[idx].status = newStatus
        if (currentSummary.value && currentSummary.value.id === id) {
          currentSummary.value.status = newStatus
          form.status = newStatus
        }
      }
    } finally {
      loading.value = false
    }
  }

  // Load initially
  fetchSummaries()

  return {
    summaryList,
    currentSummary,
    showModal,
    isEditing,
    isViewing,
    loading,
    errorMsg,
    form,
    openNewForm,
    openEditForm,
    openView,
    saveForm,
    deleteSummary,
    updateStatus,
    fetchSummaries,
    defaultNarrativeTemplate
  }
})
