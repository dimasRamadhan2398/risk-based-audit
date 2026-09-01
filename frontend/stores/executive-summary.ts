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
    return config.public.auditServiceBaseUrl || 'http://localhost:8002/api/v1'
  }

  const mockSummaries: ExecutiveSummary[] = [
    {
      id: 'ES-LHA-021-2026',
      quarter: 2,
      periodeBulan: 'April',
      tahun: 2026,
      nomorDokumen: '021/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_021_LHA_2026.pdf',
      status: 'Approved',
      narrative: 'Executive Summary Individual untuk Laporan Hasil Audit Operasional Keuangan (021/LHA/01/KS IAD/2026). Audit dilakukan untuk mengevaluasi efektivitas ICOFR dan kepatuhan terhadap SOP pembayaran.',
      jumlahLaporan: 1,
      risikoTinggi: 3,
      risikoSedang: 2,
      risikoRendah: 0,
      jumlahRekomendasi: 5,
      followUpTable: [
        { status: 'Closed', jumlah: 3, persentase: 60.0, keterangan: 'Telah divalidasi' },
        { status: 'In Progress', jumlah: 2, persentase: 40.0, keterangan: 'On track' },
        { status: 'Overdue', jumlah: 0, persentase: 0.0, keterangan: '-' }
      ],
      topFindings: [
        { unitDivision: 'Finance', judulTemuan: 'Selisih pencatatan inventaris fisik vs buku besar', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Rekonsiliasi Harian' },
        { unitDivision: 'Finance', judulTemuan: 'Keterlambatan rekonsiliasi kas harian cabang utama', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Otomatisasi Sistem' }
      ],
      matriksKompilasi: [
        { nomor: '021/LHA/01', division: 'Finance', unitKerja: 'Departemen Keuangan', prosesBisnis: 'ICOFR', judulTemuan: 'Selisih pencatatan inventaris fisik vs buku besar', nilaiRisiko: 'Tinggi', rekomendasi: 'Lakukan rekonsiliasi harian dan alert SMTP', dueDate: '2026-05-15', picUnit: 'Manager Keuangan', progres: 60, status: 'In Progress', buktiTL: 'BA_Rekonsiliasi.pdf' }
      ],
      akarMasalah: 'Kurangnya otomatisasi alarm kegagalan backup data dan kelalaian non-aktifkan akses user kasir.',
      kesimpulan: 'Secara umum pengendalian internal departemen keuangan memadai dengan beberapa area peningkatan yang perlu segera ditindaklanjuti.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-04-15',
      signatureNamaKepala: 'Zeta Ramadhani',
      signatureNIK: 'NIK-100240'
    },
    {
      id: 'ES-LHA-022-2026',
      quarter: 2,
      periodeBulan: 'Mei',
      tahun: 2026,
      nomorDokumen: '022/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_022_LHA_2026.pdf',
      status: 'Approved',
      narrative: 'Executive Summary Individual untuk Audit Keamanan Sistem Informasi & ERP (022/LHA/01/KS IAD/2026). Audit mengevaluasi tata kelola akses pengguna, patch ERP, dan pengujian DRC.',
      jumlahLaporan: 1,
      risikoTinggi: 2,
      risikoSedang: 2,
      risikoRendah: 0,
      jumlahRekomendasi: 4,
      followUpTable: [
        { status: 'Closed', jumlah: 2, persentase: 50.0, keterangan: 'Telah diterapkan' },
        { status: 'In Progress', jumlah: 2, persentase: 50.0, keterangan: 'Pengadaan MFA' },
        { status: 'Overdue', jumlah: 0, persentase: 0.0, keterangan: '-' }
      ],
      topFindings: [
        { unitDivision: 'IT', judulTemuan: 'Akses Superadmin ERP belum menggunakan MFA', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Implementasi Mandatory MFA' }
      ],
      matriksKompilasi: [],
      akarMasalah: 'Keterlambatan rilis jadwal DRC drill tahunan.',
      kesimpulan: 'Keamanan TI berjalan cukup baik dengan rekomendasi pengetatan autentikasi superadmin.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-05-02',
      signatureNamaKepala: 'Andi Firmansyah',
      signatureNIK: 'NIK-100311'
    },
    {
      id: 'ES-LHA-023-2026',
      quarter: 3,
      periodeBulan: 'Agustus',
      tahun: 2026,
      nomorDokumen: '023/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_023_LHA_2026.pdf',
      status: 'Draft',
      narrative: 'Executive Summary Individual untuk Audit Operasional Gudang & Persediaan Logistik 2026 (023/LHA/01/KS IAD/2026).',
      jumlahLaporan: 1,
      risikoTinggi: 1,
      risikoSedang: 2,
      risikoRendah: 0,
      jumlahRekomendasi: 3,
      followUpTable: [],
      topFindings: [],
      matriksKompilasi: [],
      akarMasalah: 'Prosedur pemantauan fisik stok persediaan belum otomatis.',
      kesimpulan: 'Pengelolaan logistik gudang berjalan dengan baik.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-08-10',
      signatureNamaKepala: 'Rina Wulandari',
      signatureNIK: 'NIK-100422'
    },
    {
      id: 'ES-LHA-024-2026',
      quarter: 3,
      periodeBulan: 'Agustus',
      tahun: 2026,
      nomorDokumen: '024/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_024_LHA_2026.pdf',
      status: 'Approved',
      narrative: 'Executive Summary Individual untuk Laporan Hasil Audit Kepatuhan Procurement & SCM 2026 (024/LHA/01/KS IAD/2026).',
      jumlahLaporan: 1,
      risikoTinggi: 1,
      risikoSedang: 1,
      risikoRendah: 0,
      jumlahRekomendasi: 2,
      followUpTable: [],
      topFindings: [],
      matriksKompilasi: [],
      akarMasalah: 'Kelengkapan administrasi kertas kerja HPS vendor.',
      kesimpulan: 'Kepatuhan pengadaan memenuhi standar kepatuhan internal.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-08-18',
      signatureNamaKepala: 'Budi Santoso',
      signatureNIK: 'NIK-100155'
    },
    {
      id: 'DOC-EXSUM-Q1-2026',
      quarter: 1,
      periodeBulan: 'Maret',
      tahun: 2026,
      nomorDokumen: 'DOC-EXSUM-Q1-2026',
      dokumenPath: 'DOC-EXSUM-Q1-2026.pdf',
      status: 'Approved',
      narrative: 'Executive Summary Individual DOC-EXSUM-Q1-2026 untuk Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit (025/LHA/01/KS IAD/2026). Audit mengevaluasi keandalan instalasi K3LH, sertifikasi alat, dan fasilitas pemadam kebakaran.',
      jumlahLaporan: 1,
      risikoTinggi: 2,
      risikoSedang: 1,
      risikoRendah: 0,
      jumlahRekomendasi: 3,
      followUpTable: [
        { status: 'Closed', jumlah: 2, persentase: 66.7, keterangan: 'Selesai disertifikasi' },
        { status: 'In Progress', jumlah: 1, persentase: 33.3, keterangan: 'Progres pemeliharaan hidran' },
        { status: 'Overdue', jumlah: 0, persentase: 0.0, keterangan: '-' }
      ],
      topFindings: [
        { unitDivision: 'Maintenance', judulTemuan: 'Inspeksi berkala sistem pemadam kebakaran hidran belum 100% terlaksana', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Jadwal Pemeliharaan Hidran' }
      ],
      matriksKompilasi: [],
      akarMasalah: 'Keterlambatan pengadaan sparepart hidran dan pembaruan sertifikasi K3.',
      kesimpulan: 'Sistem pengendalian K3LH beroperasi secara aman dengan perbaikan pada fasilitas hidran.',
      signatureTempat: 'Jakarta',
      signatureTanggal: '2026-03-30',
      signatureNamaKepala: 'Dewi Kusumawati',
      signatureNIK: 'NIK-100533'
    },
    {
      id: 'ES-LHA-020-2023',
      quarter: 3,
      periodeBulan: 'September',
      tahun: 2023,
      nomorDokumen: '020/LHA/01/KS IAD/2023',
      dokumenPath: 'Executive_Summary_020_LHA_2023.pdf',
      status: 'Approved',
      narrative: 'Executive Summary Individual untuk Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau (020/LHA/01/KS IAD/2023). Audit mengevaluasi ketersediaan pembangkit (EAF/EFOR), K3LH, manajemen risiko, dan SCM.',
      jumlahLaporan: 1,
      risikoTinggi: 2,
      risikoSedang: 4,
      risikoRendah: 2,
      jumlahRekomendasi: 8,
      followUpTable: [
        { status: 'Closed', jumlah: 5, persentase: 62.5, keterangan: 'Telah ditindaklanjuti' },
        { status: 'In Progress', jumlah: 3, persentase: 37.5, keterangan: 'Progres rata-rata 80%' },
        { status: 'Overdue', jumlah: 0, persentase: 0.0, keterangan: '-' }
      ],
      topFindings: [
        { unitDivision: 'Operasi', judulTemuan: 'Pelaksanaan Overhaul ME+ PLTU TBK #1 Terjadi PE 6 Hari', risiko: 'Tinggi', statusTL: 'In Progress', usulan: 'Penyusunan DMR' }
      ],
      matriksKompilasi: [
        { nomor: '020/LHA/01', division: 'Operasi', unitKerja: 'UPDK Kepulauan Riau', prosesBisnis: 'O&M Pembangkit', judulTemuan: 'Pelaksanaan Overhaul ME+ PLTU TBK #1 Terjadi PE 6 Hari', nilaiRisiko: 'Tinggi', rekomendasi: 'Penyusunan DMR & Sertifikasi Pemeliharaan', dueDate: '2023-11-30', picUnit: 'Manager UPDK', progres: 80, status: 'In Progress', buktiTL: 'Laporan_Overhaul.pdf' }
      ],
      akarMasalah: 'Masa transisi holding sub-holding, keterbatasan sarana lab batubara, dan belum lengkapnya fitur Maximo WPC.',
      kesimpulan: 'Tata kelola dan pengendalian internal berjalan baik dengan 1 Risk Management AoI dan 8 Internal Control AoI.',
      signatureTempat: 'Tanjung Pinang',
      signatureTanggal: '2023-09-22',
      signatureNamaKepala: 'Tomy Afrilianto',
      signatureNIK: 'NIK-100188'
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
        items = response.data.items
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

      // 2-Way Sync: Update matching AuditResultReport item in Result Reports store
      try {
        const auditReportStore = useAuditResultReportStore()
        if (form.nomorDokumen && form.narrative) {
          await auditReportStore.syncExecutiveSummaryField(form.nomorDokumen, form.narrative, form.jumlahRekomendasi)
        }
      } catch (errSync) {
        console.warn('Sync to AuditResultReport failed:', errSync)
      }
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

      // 2-Way Sync: Update matching AuditResultReport item in Result Reports store
      try {
        const auditReportStore = useAuditResultReportStore()
        if (form.nomorDokumen && form.narrative) {
          await auditReportStore.syncExecutiveSummaryField(form.nomorDokumen, form.narrative, form.jumlahRekomendasi)
        }
      } catch (errSync) {
        console.warn('Sync to AuditResultReport failed:', errSync)
      }
    } finally {
      loading.value = false
    }
  }

  const deletedDocNumbers = ref<string[]>([])

  const deleteSummary = async (id: string, docNum?: string) => {
    if (!await useGlobalModalStore().confirmDelete({ description: 'Apakah Anda yakin ingin menghapus Laporan Eksekutif ini?' })) return
    loading.value = true
    const item = summaryList.value.find(s => s.id === id || s.nomorDokumen === docNum)
    const targetDocNum = docNum || (item ? item.nomorDokumen : '')

    if (targetDocNum && !deletedDocNumbers.value.includes(targetDocNum)) {
      deletedDocNumbers.value.push(targetDocNum)
    }

    try {
      const baseUrl = getAuditServiceBaseUrl()
      await $fetch(`${baseUrl}/executive-summaries/${id}`, { method: 'DELETE' })
      await fetchSummaries()
    } catch (error) {
      console.error('Failed to delete on backend, simulating local deletion:', error)
    } finally {
      summaryList.value = summaryList.value.filter(s => s.id !== id && s.nomorDokumen !== targetDocNum)
      if (targetDocNum) {
        try {
          const auditReportStore = useAuditResultReportStore()
          await auditReportStore.clearExecutiveSummaryField(targetDocNum)
        } catch (e) {
          console.warn('Failed to clear executive summary on report store:', e)
        }
      }
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
    deletedDocNumbers,
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
