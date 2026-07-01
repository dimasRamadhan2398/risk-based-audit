import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'

export interface TopFinding {
  unit: string
  judul: string
  risiko: 'Tinggi' | 'Sedang' | 'Rendah'
  status: 'Closed' | 'In Progress' | 'Overdue'
  usulan: string
}

export interface MatriksFinding {
  nomor: string
  division: string
  unitKerja: string
  prosesBisnis: string
  judulTemuan: string
  nilaiRisiko: 'Tinggi' | 'Sedang' | 'Rendah'
  rekomendasi: string
  dueDate: string
  picUnit: string
  progres: number // 0-100
  status: 'Closed' | 'In Progress' | 'Overdue'
  buktiTl: string
}

export interface ExecutiveSummaryReport {
  id: string
  triwulan: 'Q1' | 'Q2' | 'Q3' | 'Q4'
  bulan: string
  tahun: number
  nomorDokumen: string
  fileName: string
  isApproved: boolean
  
  // Section I
  executiveSummaryNarrative: string

  // Section II
  jumlahLaporan: number
  risikoTinggi: number
  risikoSedang: number
  risikoRendah: number
  jumlahRekomendasi: number

  // Section III
  statusClosedCount: number
  statusInProgressCount: number
  statusOverdueCount: number
  keteranganClosed: string
  keteranganInProgress: string
  keteranganOverdue: string

  // Section IV
  topFindings: TopFinding[]

  // Section V & VII
  temaBerulang: string
  kesimpulan: string
  ttdTempat: string
  ttdTanggal: string
  ttdNama: string
  ttdNik: string

  // Section VIII / Lampiran
  matriksFindings: MatriksFinding[]
}

export const useExecutiveSummaryStore = defineStore('executive-summary', () => {
  // Mock Data
  const reports = ref<ExecutiveSummaryReport[]>([
    {
      id: 'ES-001',
      triwulan: 'Q1',
      bulan: 'Maret',
      tahun: 2026,
      nomorDokumen: 'LKA/B/001/SPI/KAI/I/2026',
      fileName: 'Executive_Summary_Q1_2026.pdf',
      isApproved: true,
      executiveSummaryNarrative: 'Periode Q1 2026, SPI telah menerbitkan 5 LHA. Total temuan 15 dengan 20 rekomendasi. Tingkat penyelesaian on-time 95%. Terdapat 2 temuan risiko tinggi overdue/terlambat terkait IT Security yang perlu arahan Direksi. Temuan berulang tertinggi pada proses pengadaan barang.',
      jumlahLaporan: 5,
      risikoTinggi: 3,
      risikoSedang: 7,
      risikoRendah: 5,
      jumlahRekomendasi: 20,
      statusClosedCount: 10,
      statusInProgressCount: 3,
      statusOverdueCount: 2,
      keteranganClosed: 'Telah divalidasi SPI',
      keteranganInProgress: 'On track, rata2 progres 75%',
      keteranganOverdue: 'Terkendala pengadaan perangkat',
      topFindings: [
        { unit: 'IT', judul: 'Kerentanan Keamanan Server Utama', risiko: 'Tinggi', status: 'Overdue', usulan: 'Eskalasi Direksi' },
        { unit: 'Finance', judul: 'Keterlambatan Rekonsiliasi Bank', risiko: 'Sedang', status: 'In Progress', usulan: 'Percepat SOP Baru' }
      ],
      temaBerulang: 'Proses pengadaan barang mengalami kendala administrasi berulang.',
      kesimpulan: 'Secara umum pengendalian intern memadai, namun perlu perhatian khusus pada sektor keamanan siber.',
      ttdTempat: 'Jakarta',
      ttdTanggal: '2026-04-05',
      ttdNama: 'Heri Gunawan',
      ttdNik: 'NIK-983011',
      matriksFindings: [
        { nomor: '001/SPI/2026', division: 'IT', unitKerja: 'IT Infrastructure Division', prosesBisnis: 'Keamanan Siber', judulTemuan: 'Kerentanan Keamanan Server Utama', nilaiRisiko: 'Tinggi', rekomendasi: 'Lakukan security patching berkala', dueDate: '2026-03-31', picUnit: 'Manager IT', progres: 40, status: 'Overdue', buktiTl: '' },
        { nomor: '002/SPI/2026', division: 'FIN', unitKerja: 'Corporate Finance Unit', prosesBisnis: 'Rekonsiliasi', judulTemuan: 'Keterlambatan Rekonsiliasi Bank', nilaiRisiko: 'Sedang', rekomendasi: 'Optimalkan sistem otomasi bank', dueDate: '2026-05-15', picUnit: 'Head of Finance', progres: 70, status: 'In Progress', buktiTl: '' },
        { nomor: '003/SPI/2026', division: 'HR', unitKerja: 'Human Resources Division', prosesBisnis: 'Rekrutmen', judulTemuan: 'Berkas Lamaran Kurang Lengkap', nilaiRisiko: 'Rendah', rekomendasi: 'Lengkapi berkas administrasi', dueDate: '2026-03-10', picUnit: 'HR Specialist', progres: 100, status: 'Closed', buktiTl: 'BA Kelayakan 1' }
      ]
    }
  ])

  // Active form state for creation or editing
  const currentTriwulan = ref<'Q1' | 'Q2' | 'Q3' | 'Q4'>('Q1')
  const showForm = ref(false)
  const isEditing = ref(false)
  const editingId = ref<string | null>(null)

  const activeReport = reactive<Omit<ExecutiveSummaryReport, 'id' | 'triwulan' | 'isApproved'>>({
    bulan: '',
    tahun: 2026,
    nomorDokumen: '',
    fileName: '',
    executiveSummaryNarrative: '',
    jumlahLaporan: 0,
    risikoTinggi: 0,
    risikoSedang: 0,
    risikoRendah: 0,
    jumlahRekomendasi: 0,
    statusClosedCount: 0,
    statusInProgressCount: 0,
    statusOverdueCount: 0,
    keteranganClosed: 'Telah divalidasi SPI',
    keteranganInProgress: 'On track',
    keteranganOverdue: 'Overdue',
    topFindings: [],
    temaBerulang: '',
    kesimpulan: '',
    ttdTempat: 'Jakarta',
    ttdTanggal: new Date().toISOString().split('T')[0],
    ttdNama: 'Kepala SPI',
    ttdNik: 'NIK-100200',
    matriksFindings: []
  })

  // Dropdown options
  const bulanOptions = computed(() => {
    switch (currentTriwulan.value) {
      case 'Q1': return ['Januari', 'Februari', 'Maret']
      case 'Q2': return ['April', 'Mei', 'Juni']
      case 'Q3': return ['Juli', 'Agustus', 'September']
      case 'Q4': return ['Oktober', 'November', 'Desember']
    }
  })

  const tahunOptions = [2024, 2025, 2026, 2027, 2028]

  // Calculated fields for current active report
  const totalFindings = computed(() => {
    return Number(activeReport.risikoTinggi) + Number(activeReport.risikoSedang) + Number(activeReport.risikoRendah)
  })

  const totalStatusCount = computed(() => {
    return Number(activeReport.statusClosedCount) + Number(activeReport.statusInProgressCount) + Number(activeReport.statusOverdueCount)
  })

  const percentClosed = computed(() => {
    const total = totalStatusCount.value
    if (total === 0) return 0
    return Math.round((Number(activeReport.statusClosedCount) / total) * 1000) / 10
  })

  const percentInProgress = computed(() => {
    const total = totalStatusCount.value
    if (total === 0) return 0
    return Math.round((Number(activeReport.statusInProgressCount) / total) * 1000) / 10
  })

  const percentOverdue = computed(() => {
    const total = totalStatusCount.value
    if (total === 0) return 0
    return Math.round((Number(activeReport.statusOverdueCount) / total) * 1000) / 10
  })

  // Rule 2 Sync validation check
  const isSyncWarning = computed(() => {
    const countTinggi = activeReport.matriksFindings.filter(f => f.nilaiRisiko === 'Tinggi').length
    const countSedang = activeReport.matriksFindings.filter(f => f.nilaiRisiko === 'Sedang').length
    const countRendah = activeReport.matriksFindings.filter(f => f.nilaiRisiko === 'Rendah').length

    return (
      countTinggi !== Number(activeReport.risikoTinggi) ||
      countSedang !== Number(activeReport.risikoSedang) ||
      countRendah !== Number(activeReport.risikoRendah)
    )
  })

  // Actions
  const initForm = (triwulan: 'Q1' | 'Q2' | 'Q3' | 'Q4') => {
    currentTriwulan.value = triwulan
    isEditing.value = false
    editingId.value = null
    
    // Default placeholders
    const defaultNarrative = `Periode ${triwulan} 2026, SPI telah menerbitkan .... LHA. Total temuan …. dengan .... rekomendasi. Tingkat penyelesaian on-time 95%. Terdapat 5 temuan risiko tinggi overdue/terlambat terkait .... yang perlu arahan Direksi. Temuan berulang tertinggi pada proses ……………...`
    
    Object.assign(activeReport, {
      bulan: triwulan === 'Q1' ? 'Januari' : triwulan === 'Q2' ? 'April' : triwulan === 'Q3' ? 'Juli' : 'Oktober',
      tahun: 2026,
      nomorDokumen: '',
      fileName: '',
      executiveSummaryNarrative: defaultNarrative,
      jumlahLaporan: 0,
      risikoTinggi: 0,
      risikoSedang: 0,
      risikoRendah: 0,
      jumlahRekomendasi: 0,
      statusClosedCount: 0,
      statusInProgressCount: 0,
      statusOverdueCount: 0,
      keteranganClosed: 'Telah divalidasi SPI',
      keteranganInProgress: 'On track',
      keteranganOverdue: 'Overdue',
      topFindings: [
        { unit: 'OP', judul: 'SOP Operasional Belum Terupdate', risiko: 'Sedang', status: 'In Progress', usulan: 'Revisi SOP' }
      ],
      temaBerulang: '',
      kesimpulan: '',
      ttdTempat: 'Jakarta',
      ttdTanggal: new Date().toISOString().split('T')[0],
      ttdNama: 'Heri Gunawan',
      ttdNik: 'NIK-983011',
      matriksFindings: []
    })

    showForm.value = true
  }

  const editReport = (report: ExecutiveSummaryReport) => {
    currentTriwulan.value = report.triwulan
    isEditing.value = true
    editingId.value = report.id
    
    // Copy all data
    Object.assign(activeReport, JSON.parse(JSON.stringify(report)))
    showForm.value = true
  }

  const saveActiveReport = () => {
    if (!activeReport.nomorDokumen) {
      alert('Nomor Dokumen wajib diisi.')
      return
    }
    if (!activeReport.fileName && !isEditing.value) {
      alert('Upload Dokumen Resmi wajib diisi.')
      return
    }

    if (isEditing.value && editingId.value) {
      const idx = reports.value.findIndex(r => r.id === editingId.value)
      if (idx !== -1) {
        reports.value[idx] = {
          ...reports.value[idx],
          ...JSON.parse(JSON.stringify(activeReport))
        }
      }
    } else {
      const newReport: ExecutiveSummaryReport = {
        id: 'ES-' + Date.now(),
        triwulan: currentTriwulan.value,
        isApproved: false,
        ...JSON.parse(JSON.stringify(activeReport))
      }
      reports.value.push(newReport)
    }

    showForm.value = false
  }

  const deleteReport = (id: string) => {
    if (confirm('Apakah Anda yakin ingin menghapus laporan kompilasi ini?')) {
      reports.value = reports.value.filter(r => r.id !== id)
    }
  }

  const toggleApproval = (id: string) => {
    const report = reports.value.find(r => r.id === id)
    if (report) {
      report.isApproved = !report.isApproved
    }
  }

  // Matriks list actions
  const addMatriksRow = () => {
    activeReport.matriksFindings.push({
      nomor: `00${activeReport.matriksFindings.length + 1}/SPI/2026`,
      division: 'OP',
      unitKerja: '',
      prosesBisnis: '',
      judulTemuan: '',
      nilaiRisiko: 'Sedang',
      rekomendasi: '',
      dueDate: new Date().toISOString().split('T')[0],
      picUnit: '',
      progres: 0,
      status: 'In Progress',
      buktiTl: ''
    })
  }

  const removeMatriksRow = (index: number) => {
    activeReport.matriksFindings.splice(index, 1)
  }

  return {
    reports,
    currentTriwulan,
    showForm,
    isEditing,
    editingId,
    activeReport,
    bulanOptions,
    tahunOptions,
    totalFindings,
    totalStatusCount,
    percentClosed,
    percentInProgress,
    percentOverdue,
    isSyncWarning,
    initForm,
    editReport,
    saveActiveReport,
    deleteReport,
    toggleApproval,
    addMatriksRow,
    removeMatriksRow
  }
})
