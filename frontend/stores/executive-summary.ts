import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { extractErrorMessage } from '~/utils/error'
import { getAuditServiceBaseUrl } from '~/composables/useApiUrl'
import { useAuditResultReportStore } from './audit-result-report'

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
  assignmentLetterId?: string
  quarter: number // 1, 2, 3, 4
  periodeBulan: string // e.g. "Maret"
  tahun: number // default 2026
  nomorDokumen: string
  dokumenPath: string
  status: 'Draft' | 'Approved' | 'Rejected'
  executiveNote?: string

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

export const ES_PERSISTENCE_KEY = 'risk_based_audit_executive_summaries_persisted_v2'

export const loadPersistedExecutiveSummaryOverrides = (): Record<string, Partial<ExecutiveSummary>> => {
  if (typeof window === 'undefined') return {}
  try {
    // Purge legacy contaminated v1 storage if present
    if (localStorage.getItem('risk_based_audit_executive_summaries_persisted_v1')) {
      localStorage.removeItem('risk_based_audit_executive_summaries_persisted_v1')
    }
    const raw = localStorage.getItem(ES_PERSISTENCE_KEY)
    const current = raw ? JSON.parse(raw) : {}
    // Delete dummy or non-standard legacy entries
    delete current['DOC-EXSUM-Q1-2026']
    delete current['020/LHA/01/KS IAD/2023']
    delete current['019/LHA/01/KS IAD/2025']
    return current
  } catch (e) {
    console.warn('Failed to read persisted executive summaries from localStorage:', e)
    return {}
  }
}

export const savePersistedExecutiveSummaryOverride = (keys: (string | undefined)[], data: Partial<ExecutiveSummary>) => {
  if (typeof window === 'undefined') return
  try {
    const current = loadPersistedExecutiveSummaryOverrides()
    // Extract only safe modifiable fields - NEVER overwrite nomorDokumen or ID across cards!
    const safeData: Partial<ExecutiveSummary> = {}
    if (data.status) safeData.status = data.status
    if (data.executiveNote !== undefined) safeData.executiveNote = data.executiveNote
    if (data.narrative) safeData.narrative = data.narrative
    if (data.jumlahRekomendasi !== undefined) safeData.jumlahRekomendasi = data.jumlahRekomendasi
    if (data.risikoTinggi !== undefined) safeData.risikoTinggi = data.risikoTinggi
    if (data.risikoSedang !== undefined) safeData.risikoSedang = data.risikoSedang
    if (data.risikoRendah !== undefined) safeData.risikoRendah = data.risikoRendah
    if (data.topFindings) safeData.topFindings = data.topFindings
    if (data.matriksKompilasi) safeData.matriksKompilasi = data.matriksKompilasi
    if (data.akarMasalah) safeData.akarMasalah = data.akarMasalah
    if (data.kesimpulan) safeData.kesimpulan = data.kesimpulan
    if (data.dokumenPath) safeData.dokumenPath = data.dokumenPath
    if (data.assignmentLetterId) safeData.assignmentLetterId = data.assignmentLetterId

    keys.filter((k): k is string => !!k && typeof k === 'string').forEach(k => {
      if (k === 'DOC-EXSUM-Q1-2026') return
      current[k] = { ...(current[k] || {}), ...safeData }
    })
    localStorage.setItem(ES_PERSISTENCE_KEY, JSON.stringify(current))
  } catch (e) {
    console.warn('Failed to save persisted executive summary override:', e)
  }
}

export const isValidUuid = (val?: string): boolean => {
  if (!val) return false
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(val)
}

export const useExecutiveSummaryStore = defineStore('executive-summary', () => {
  const summaryList = ref<ExecutiveSummary[]>([])
  const currentSummary = ref<ExecutiveSummary | null>(null)

  const showModal = ref(false)
  const isEditing = ref(false)
  const isViewing = ref(false)
  const loading = ref(false)
  const errorMsg = ref('')
  const toast = useToastNotification()

  // Default narrative template
  const defaultNarrativeTemplate = (bulan: string, tahun: number = 2026) => {
    return `Periode ${bulan} ${tahun}, SPI telah menerbitkan .... LHA. Total temuan …. dengan .... rekomendasi. Tingkat penyelesaian on-time 95%. Terdapat 5 temuan risiko tinggi overdue/terlambat terkait .... yang perlu arahan Direksi. Temuan berulang tertinggi pada proses ……………..`
  }

  // Active form data
  const form = reactive<Omit<ExecutiveSummary, 'id'>>({
    assignmentLetterId: '',
    quarter: 1,
    periodeBulan: 'Januari',
    tahun: 2026,
    nomorDokumen: '',
    dokumenPath: '',
    status: 'Draft',
    executiveNote: '',
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



  const mockSummaries: ExecutiveSummary[] = [
    {
      id: 'ES-LHA-021-2026',
      assignmentLetterId: 'ST-001/SKAI/2026',
      quarter: 1,
      periodeBulan: 'Januari - Maret',
      tahun: 2026,
      nomorDokumen: '021/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_021_LHA_2026.pdf',
      status: 'Approved',
      executiveNote: 'Mohon tindak lanjuti rekonsiliasi kas harian dan koordinasikan perbaikan dengan tim Keuangan.',
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
      assignmentLetterId: 'ST-002/SKAI/2026',
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
      assignmentLetterId: 'ST-003/SKAI/2026',
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
      assignmentLetterId: 'ST-004/SKAI/2026',
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
      id: 'ES-LHA-025-2026',
      assignmentLetterId: 'ST-005/SKAI/2026',
      quarter: 4,
      periodeBulan: 'September',
      tahun: 2026,
      nomorDokumen: '025/LHA/01/KS IAD/2026',
      dokumenPath: 'Executive_Summary_025_LHA_2026.pdf',
      status: 'Approved',
      executiveNote: 'Disetujui. Mohon koordinasikan dengan Direksi terkait temuan overhauling.',
      narrative: 'Executive Summary Individual untuk Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit (025/LHA/01/KS IAD/2026). Audit mengevaluasi keandalan instalasi K3LH, sertifikasi alat, dan fasilitas pemadam kebakaran.',
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
      signatureTanggal: '2026-09-15',
      signatureNamaKepala: 'Dewi Kusumawati',
      signatureNIK: 'NIK-100533'
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
        const backendItems = items
          .map(item => parseSummaryFromBackend(item))
          .filter(b => b.nomorDokumen !== 'DOC-EXSUM-Q1-2026' && b.nomorDokumen !== '020/LHA/01/KS IAD/2023' && b.nomorDokumen !== '019/LHA/01/KS IAD/2025')
        const docSet = new Set(backendItems.map(b => b.nomorDokumen))
        const remainingMocks = mockSummaries.filter(m => !docSet.has(m.nomorDokumen))
        summaryList.value = [...backendItems, ...remainingMocks]
      } else {
        summaryList.value = [...mockSummaries]
      }
    } catch (error) {
      console.error('Failed to fetch executive summaries, falling back to mock data:', error)
      errorMsg.value = extractErrorMessage(error, 'Failed to load executive summaries.')
      summaryList.value = [...mockSummaries]
    } finally {
      // Overlay persisted changes from localStorage (status approvals and notes)
      const overrides = loadPersistedExecutiveSummaryOverrides()
      summaryList.value = summaryList.value
        .map(item => {
          const override = overrides[item.nomorDokumen] || overrides[item.id]
          return override ? { ...item, ...override } : item
        })
        .filter(s => s.nomorDokumen !== 'DOC-EXSUM-Q1-2026' && s.nomorDokumen !== '020/LHA/01/KS IAD/2023' && s.nomorDokumen !== '019/LHA/01/KS IAD/2025')
      loading.value = false
    }
  }

  const parseSummaryFromBackend = (item: any): ExecutiveSummary => {
    return {
      ...item,
      assignmentLetterId: item.assignmentLetterId || item.assignment_letter_id || '',
      executiveNote: item.executiveNote || item.executive_note || '',
      quarter: Number(item.quarter),
      tahun: Number(item.tahun),
      jumlahLaporan: Number(item.jumlahLaporan),
      risikoTinggi: Number(item.risikoTinggi),
      risikoSedang: Number(item.risikoSedang),
      risikoRendah: Number(item.risikoRendah),
      jumlahRekomendasi: Number(item.jumlahRekomendasi),
      followUpTable: item.followUpTable ? (typeof item.followUpTable === 'string' ? JSON.parse(item.followUpTable) : item.followUpTable) : [],
      topFindings: item.topFindings ? (typeof item.topFindings === 'string' ? JSON.parse(item.topFindings) : item.topFindings) : [],
      matriksKompilasi: item.matriksKompilasi ? (typeof item.matriksKompilasi === 'string' ? JSON.parse(item.matriksKompilasi) : item.matriksKompilasi) : []
    }
  }

  const serializeSummaryForBackend = (data: Omit<ExecutiveSummary, 'id'> | ExecutiveSummary) => {
    const rawId = (data as any).id
    const payload: any = {
      ...data,
      assignmentLetterId: (data as any).assignmentLetterId || '',
      quarter: Number(data.quarter),
      tahun: Number(data.tahun),
      jumlahLaporan: Number(data.jumlahLaporan),
      risikoTinggi: Number(data.risikoTinggi),
      risikoSedang: Number(data.risikoSedang),
      risikoRendah: Number(data.risikoRendah),
      jumlahRekomendasi: Number(data.jumlahRekomendasi),
      followUpTable: typeof data.followUpTable === 'string' ? data.followUpTable : JSON.stringify(data.followUpTable || []),
      topFindings: typeof data.topFindings === 'string' ? data.topFindings : JSON.stringify(data.topFindings || []),
      matriksKompilasi: typeof data.matriksKompilasi === 'string' ? data.matriksKompilasi : JSON.stringify(data.matriksKompilasi || [])
    }
    if (!isValidUuid(rawId)) {
      delete payload.id
    }
    return payload
  }

  const openNewForm = (quarterNum: number = 1, assignmentLetterId: string = '') => {
    isEditing.value = false
    isViewing.value = false
    currentSummary.value = null

    // Set default month based on selected quarter
    let defaultMonth = 'Januari'
    if (quarterNum === 2) defaultMonth = 'April'
    if (quarterNum === 3) defaultMonth = 'Juli'
    if (quarterNum === 4) defaultMonth = 'Oktober'

    Object.assign(form, {
      assignmentLetterId: assignmentLetterId || 'ST-001/SKAI/2026',
      quarter: quarterNum || 1,
      periodeBulan: defaultMonth,
      tahun: 2026,
      nomorDokumen: '',
      dokumenPath: '',
      status: 'Draft',
      executiveNote: '',
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
    if (!form.assignmentLetterId && summary.assignmentLetterId) {
      form.assignmentLetterId = summary.assignmentLetterId
    }
    showModal.value = true
  }

  const openView = (summary: ExecutiveSummary) => {
    isEditing.value = false
    isViewing.value = true
    currentSummary.value = summary
    Object.assign(form, JSON.parse(JSON.stringify(summary)))
    if (!form.assignmentLetterId && summary.assignmentLetterId) {
      form.assignmentLetterId = summary.assignmentLetterId
    }
    showModal.value = true
  }

  const saveForm = async () => {
    loading.value = true
    errorMsg.value = ''
    try {
      const baseUrl = getAuditServiceBaseUrl()
      const payload = serializeSummaryForBackend(form)

      if (isEditing.value && currentSummary.value) {
        if (isValidUuid(currentSummary.value.id)) {
          await $fetch(`${baseUrl}/executive-summaries/${currentSummary.value.id}`, {
            method: 'PUT',
            body: payload
          })
        } else {
          const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
            method: 'POST',
            body: payload
          })
          if (res && res.data && res.data.id) {
            currentSummary.value.id = res.data.id
          }
        }
      } else {
        // Prevent duplicate creation: if summary with same nomorDokumen exists, update it!
        const existingItem = summaryList.value.find(s => form.nomorDokumen && s.nomorDokumen === form.nomorDokumen)
        if (existingItem && isValidUuid(existingItem.id)) {
          await $fetch(`${baseUrl}/executive-summaries/${existingItem.id}`, {
            method: 'PUT',
            body: payload
          })
          currentSummary.value = existingItem
        } else {
          const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
            method: 'POST',
            body: payload
          })
          if (res && res.data && res.data.id) {
            if (currentSummary.value) currentSummary.value.id = res.data.id
          }
        }
      }

      savePersistedExecutiveSummaryOverride(
        [form.nomorDokumen, currentSummary.value?.id],
        { ...form }
      )

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
      const detail = extractErrorMessage(error, 'Gagal menyimpan Executive Summary.')
      errorMsg.value = detail
      toast.showError('Gagal menyimpan Executive Summary.', detail)

      // Simulating save in state for offline capabilities
      const targetDoc = form.nomorDokumen
      const idx = summaryList.value.findIndex(s => 
        (currentSummary.value && s.id === currentSummary.value.id) || 
        (targetDoc && s.nomorDokumen === targetDoc)
      )
      if (idx !== -1) {
        summaryList.value[idx] = {
          ...summaryList.value[idx],
          ...JSON.parse(JSON.stringify(form))
        }
        currentSummary.value = summaryList.value[idx]
      } else {
        const newSummary: ExecutiveSummary = {
          id: `ES-Q${form.quarter}-2026-${Math.floor(100 + Math.random() * 900)}`,
          ...JSON.parse(JSON.stringify(form))
        }
        summaryList.value.push(newSummary)
        currentSummary.value = newSummary
      }

      savePersistedExecutiveSummaryOverride(
        [form.nomorDokumen, currentSummary.value?.id],
        { ...form }
      )

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
      toast.showSuccess('Executive Summary berhasil disimpan!')
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
      toast.showSuccess('Executive Summary berhasil dihapus!')
    } catch (error: any) {
      console.error('Failed to delete on backend, simulating local deletion:', error)
      const detail = extractErrorMessage(error, 'Gagal menghapus Executive Summary.')
      errorMsg.value = detail
      toast.showError('Gagal menghapus Executive Summary.', detail)
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

  const saveExecutiveNote = async () => {
    if (!currentSummary.value) return

    loading.value = true
    errorMsg.value = ''
    const note = form.executiveNote?.trim() || ''
    const updated: ExecutiveSummary = {
      ...currentSummary.value,
      ...JSON.parse(JSON.stringify(form)),
      executiveNote: note
    }

    // Persist immediately in localStorage so page refresh never loses the note
    savePersistedExecutiveSummaryOverride(
      [currentSummary.value.id, currentSummary.value.nomorDokumen, form.nomorDokumen],
      { executiveNote: note }
    )

    try {
      const baseUrl = getAuditServiceBaseUrl()
      if (isValidUuid(currentSummary.value.id)) {
        await $fetch(`${baseUrl}/executive-summaries/${currentSummary.value.id}`, {
          method: 'PUT',
          body: serializeSummaryForBackend(updated)
        })
      } else {
        const realItem = summaryList.value.find(s => (s.id === currentSummary.value!.id || s.nomorDokumen === updated.nomorDokumen) && isValidUuid(s.id))
        if (realItem) {
          await $fetch(`${baseUrl}/executive-summaries/${realItem.id}`, {
            method: 'PUT',
            body: serializeSummaryForBackend(updated)
          })
          updated.id = realItem.id
        } else {
          try {
            const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
              method: 'POST',
              body: serializeSummaryForBackend(updated)
            })
            if (res && res.data && res.data.id) {
              updated.id = res.data.id
            }
          } catch (postErr) {
            console.warn('POST executive summary note failed:', postErr)
          }
        }
      }
    } catch (error: any) {
      console.warn('Failed to save executive note on backend, saving to local store and storage:', error)
    } finally {
      const index = summaryList.value.findIndex(summary => 
        summary.id === currentSummary.value?.id || 
        (summary.nomorDokumen && summary.nomorDokumen === currentSummary.value?.nomorDokumen)
      )
      if (index !== -1) {
        summaryList.value[index] = { ...summaryList.value[index], ...updated }
      } else {
        summaryList.value.push(updated)
      }
      summaryList.value = [...summaryList.value]
      currentSummary.value = updated
      form.executiveNote = note

      // Sync note to matching report in AuditResultReport store for Auditor view
      try {
        const auditReportStore = useAuditResultReportStore()
        const targetDocNum = updated.nomorDokumen
        const matchedReport = auditReportStore.reportList.find(r => r.reportNumber === targetDocNum)
        if (matchedReport) {
          (matchedReport as any).executiveNote = note
        }
      } catch (e) {
        console.warn('Sync note to AuditResultReport failed:', e)
      }

      toast.showSuccess('Catatan Executive berhasil disimpan untuk Auditor.')
      loading.value = false
    }
  }

  const updateStatus = async (id: string, newStatus: 'Draft' | 'Approved' | 'Rejected') => {
    loading.value = true
    errorMsg.value = ''
    const item = summaryList.value.find(s => s.id === id)
      || (currentSummary.value?.id === id ? currentSummary.value : null)
    const updated: ExecutiveSummary = item
      ? { ...item, status: newStatus }
      : { ...(currentSummary.value || form), id, status: newStatus } as ExecutiveSummary

    // Persist immediately in localStorage so page refresh never loses the approved status
    savePersistedExecutiveSummaryOverride(
      [id, updated.nomorDokumen, item?.nomorDokumen, form.nomorDokumen],
      { status: newStatus }
    )

    try {
      const baseUrl = getAuditServiceBaseUrl()
      if (isValidUuid(id)) {
        const payload = serializeSummaryForBackend(updated)
        await $fetch(`${baseUrl}/executive-summaries/${id}`, {
          method: 'PUT',
          body: payload
        })
      } else {
        const realItem = summaryList.value.find(s => (s.id === id || s.nomorDokumen === updated.nomorDokumen) && isValidUuid(s.id))
        if (realItem) {
          await $fetch(`${baseUrl}/executive-summaries/${realItem.id}`, {
            method: 'PUT',
            body: serializeSummaryForBackend(updated)
          })
          updated.id = realItem.id
        } else {
          try {
            const res: any = await $fetch(`${baseUrl}/executive-summaries`, {
              method: 'POST',
              body: serializeSummaryForBackend(updated)
            })
            if (res && res.data && res.data.id) {
              updated.id = res.data.id
            }
          } catch (postErr) {
            console.warn('POST executive summary status failed:', postErr)
          }
        }
      }
    } catch (error: any) {
      console.warn('Failed to update status on backend, simulating local update:', error)
    } finally {
      const idx = summaryList.value.findIndex(s => 
        s.id === id || 
        (updated.nomorDokumen && s.nomorDokumen === updated.nomorDokumen)
      )
      if (idx !== -1) {
        summaryList.value[idx] = { ...summaryList.value[idx], ...updated }
      } else {
        summaryList.value.push(updated)
      }
      summaryList.value = [...summaryList.value]

      if (currentSummary.value) {
        currentSummary.value = { ...currentSummary.value, ...updated }
      }
      form.status = newStatus

      // Sync status to matching report in AuditResultReport store
      try {
        const auditReportStore = useAuditResultReportStore()
        const targetDocNum = updated.nomorDokumen
        const matchedReport = auditReportStore.reportList.find(r => r.reportNumber === targetDocNum)
        if (matchedReport) {
          matchedReport.status = newStatus === 'Approved' ? 'Final' : 'Draft'
        }
      } catch (e) {
        console.warn('Sync status to AuditResultReport failed:', e)
      }

      toast.showSuccess('Status Executive Summary berhasil diperbarui!')
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
    saveExecutiveNote,
    deleteSummary,
    updateStatus,
    fetchSummaries,
    defaultNarrativeTemplate
  }
})
